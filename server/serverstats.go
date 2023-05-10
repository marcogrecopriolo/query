//  Copyright 2022-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package server

import (
	"runtime"
	"sync/atomic"
	"time"

	json "github.com/couchbase/go_json"
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/ffdc"
	"github.com/couchbase/query/logging"
	"github.com/couchbase/query/memory"
	"github.com/couchbase/query/tenant"
	"github.com/couchbase/query/util"
	"github.com/couchbase/query/value"
)

const (
	_STATS_INTRVL   = 30 * time.Second // load factor interval
	_LOG_INTRVL     = 10               // log interval 5min
	_MOVING_WINDOW  = 30               // 15min, load factor moving average of 15min i.e 30 values
	DEF_LOAD_FACTOR = 35               // default load factor above 30% so that at start no nodes will be added

	_FFDC_MEM_THRESH = 90                                     // FFDC memory use threshold
	_FFDC_MEM_RATE   = 10                                     // FFDC memory increase rate threshold
	_SAMPLES_2_HOURS = int((time.Hour * 2) / _STATS_INTRVL)   // number of samples for determining average memory use
	_SAMPLES_MIN     = int((time.Minute * 5) / _STATS_INTRVL) // minimum number of samples for average memory use
)

var qsLoadFactor uint32 // Query Service moving average Load Factor

//////////////////////////////////////////////////////////////
// Concrete Type/Struct
//////////////////////////////////////////////////////////////

type statsCollector struct {
	server           *Server
	loadFactors      [_MOVING_WINDOW]int
	sumOfLoadFactors int
	nLoadFactors     int
}

//////////////////////////////////////////////////////////////
// Cpu/Memory Collector
//////////////////////////////////////////////////////////////

// Start Stats collection
func (this *Server) StartStatsCollector() (err error) {

	collector := &statsCollector{server: this}

	for i := 0; i < len(collector.loadFactors); i++ {
		collector.loadFactors[i] = DEF_LOAD_FACTOR
		collector.sumOfLoadFactors += collector.loadFactors[i]
		collector.nLoadFactors += 1
	}

	updateQsLoadFactor(int(collector.sumOfLoadFactors / collector.nLoadFactors))
	newUtime, newStime := util.CpuTimes()
	this.lastTotalTime = newUtime + newStime
	this.lastNow = time.Now()
	this.lastCpuPercent = 1.0

	// start stats collection
	go collector.runCollectStats()

	return nil
}

type runningAverage struct {
	total   int64
	samples []int64
	index   int
}

func newRunningAverage(samples int) *runningAverage {
	rv := &runningAverage{samples: make([]int64, samples)}
	return rv
}

func (this *runningAverage) value() int64 {
	return this.total / int64(this.count())
}

func (this *runningAverage) count() int {
	if this.index >= len(this.samples) {
		return len(this.samples)
	}
	return this.index
}

func (this *runningAverage) record(v int64) {
	i := this.index & len(this.samples)
	if this.index >= len(this.samples) {
		this.total -= this.samples[i]
	}
	this.samples[i] = v
	this.total += v
	this.index++
}

// Gather Cpu/Memory
func (c *statsCollector) runCollectStats() {
	var lastGC uint64
	ticker := time.NewTicker(_STATS_INTRVL)
	defer func() {
		ticker.Stop()
		// cannot panic and die
		e := recover()
		logging.Debugf("System stats collector failed with: %v.  Restarting.", e)
		go c.runCollectStats()
	}()

	index := 0
	unhealthyCount := 0
	averageMemoryUsage := newRunningAverage(_SAMPLES_2_HOURS)

	oldStats := make(map[string]interface{}, 6)
	newStats := make(map[string]interface{}, 6)
	c.server.AccountingStore().ExternalVitals(oldStats)
	tickerFunc := func() {
		if c.server.IsHealthy() {
			unhealthyCount = 0
			newStats["healthy"] = true
		} else {
			unhealthyCount++
			newStats["healthy"] = unhealthyCount > 1
		}
		loadFactor := c.server.loadFactor(true)
		c.sumOfLoadFactors += (loadFactor - c.loadFactors[index])
		c.loadFactors[index] = loadFactor
		updateQsLoadFactor(int(c.sumOfLoadFactors / c.nLoadFactors))

		newStats["loadfactor"] = getQsLoadFactor()
		newStats["load"] = c.server.Load()
		newStats["process.service.usage"] = c.server.ServicerUsage()
		newStats["process.percore.cpupercent"] = c.server.CpuUsage(false)
		newStats["process.memory.usage"], lastGC = c.server.MemoryUsage(false)
		newStats["request.queued.count"] = c.server.QueuedRequests()
		newStats["node.allocated.values"] = value.AllocatedValuesCount()
		m := memory.AllocatedMemory()
		if m > 0 {
			newStats["node.memory.usage"] = m
		} else {
			delete(newStats, "node.memory.usage")
		}
		if tenant.IsServerless() {
			tenants := make(map[string]interface{})
			tenant.Foreach(func(n string, m memory.MemoryManager) {
				tenants[n] = m.AllocatedMemory()
			})
			newStats["tenant.memory.usage"] = tenants
		}

		// get per bucket stats
		var bstats map[string]interface{}
		store, ok := datastore.GetDatastore().(datastore.Datastore2)
		if ok {
			store.ForeachBucket(func(b datastore.ExtendedBucket) {
				stats := b.GetIOStats(false, false, false, tenant.IsServerless())
				if len(stats) != 0 {
					if bstats == nil {
						bstats = make(map[string]interface{})
					}
					bstats[b.Name()] = stats
				}
			})
		}
		if bstats != nil {
			newStats["bucket.IO.stats"] = bstats
		}

		if ac, ok := c.server.AccountingStore().(interface{ CompletedRequests() int64 }); ok {
			newStats["request.completed.count"] = ac.CompletedRequests()
		}
		oldStats = c.server.AccountingStore().ExternalVitals(newStats)

		// FFDC triggers
		ncc, _ := newStats["request.completed.count"].(int64)
		occ, _ := oldStats["request.completed.count"].(int64)
		mu, _ := newStats["process.memory.usage"].(int64)

		if ncc == occ { // no progress in last interval
			ratio := c.server.QueuedRequests() / (c.server.Servicers() + c.server.PlusServicers())
			if ratio >= 3 {
				logging.Warnf("No processed requests with queue of %v", c.server.QueuedRequests())
				ffdc.Capture(ffdc.StalledQueue)
			} else {
				ffdc.Reset(ffdc.StalledQueue)
			}
		} else {
			ffdc.Reset(ffdc.StalledQueue)
		}

		newStats = oldStats

		trigger := true
		if mu >= _FFDC_MEM_THRESH {
			logging.Warnf("Memory threshold exceeded: %v%% > %v%%", mu, _FFDC_MEM_THRESH)
			if trigger {
				ffdc.Capture(ffdc.MemoryThreshold)
				trigger = false
			}
		} else {
			ffdc.Reset(ffdc.MemoryThreshold)
		}

		averageMemoryUsage.record(mu)
		delta := mu - averageMemoryUsage.value()
		if delta > _FFDC_MEM_RATE && averageMemoryUsage.count() > _SAMPLES_MIN {
			logging.Warnf("Memory growth rate threshold exceeded: %v%% > 10%%", delta, _FFDC_MEM_RATE)
			if trigger {
				ffdc.Capture(ffdc.MemoryRate)
				trigger = false
			}
		} else {
			ffdc.Reset(ffdc.MemoryRate)
		}

		clean := false
		if (index % _LOG_INTRVL) == 0 {
			mstats, _ := c.server.AccountingStore().Vitals()
			prss := mstats["process.rss"]
			if prss != nil {
				rss, ok := prss.(uint64)
				nodeQuota := memory.NodeQuota() * util.MiB
				clean = ok && rss > nodeQuota && nodeQuota > 0
			}

			if buf, e := json.Marshal(mstats); e == nil {
				logging.Infof("Query Engine Stats %v", string(buf))
			}
		}
		index++
		index %= c.nLoadFactors

		// expire tenants if required
		if clean && tenant.IsServerless() {
			tenant.Foreach(func(n string, m memory.MemoryManager) {
				m.Expire()
			})
		}

		util.ResyncTime()
		if util.Now().UnixNano()-int64(lastGC) > int64(_STATS_INTRVL) {
			logging.Debugf("Running GC")
			runtime.GC()
		}

	}

	tickerFunc()
	index--
	for range ticker.C {
		tickerFunc()
	}
}

func updateQsLoadFactor(loadFactor int) {
	atomic.StoreUint32(&qsLoadFactor, uint32(loadFactor))
}

func getQsLoadFactor() int {
	return int(atomic.LoadUint32(&qsLoadFactor))
}
