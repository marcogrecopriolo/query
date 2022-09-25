//  Copyright 2022-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

//go:build enterprise

package tenant

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/couchbase/cbauth/service"
	atomic "github.com/couchbase/go-couchbase/platform"
	"github.com/couchbase/query/distributed"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/logging"
	"github.com/couchbase/query/util"
	"github.com/couchbase/regulator"
	"github.com/couchbase/regulator/factory"
	"github.com/couchbase/regulator/metering"
	"github.com/gorilla/mux"
)

var isServerless bool
var resourceManagers []ResourceManager
var throttleTimes map[string]util.Time = make(map[string]util.Time, _MAX_TENANTS)
var throttleLock sync.RWMutex

type Unit atomic.AlignedUint64
type Service int
type Services [_SIZER]Unit
type ResourceManager func(string)

type Context regulator.UserCtx
type Endpoint interface {
	Mux() *mux.Router
	Authorize(req *http.Request) errors.Error
	WriteError(err errors.Error, w http.ResponseWriter, req *http.Request)
}

const (
	QUERY_CU = Service(iota)
	JS_CU
	GSI_RU
	FTS_RU
	KV_RU
	KV_WU
	_SIZER
)

const _THROTTLE_DELAY = 500 * time.Millisecond

var toReg = [_SIZER]struct {
	service  regulator.Service
	unit     regulator.UnitType
	billable bool
}{
	{regulator.Query, regulator.Compute, false}, // query, not billable
	{regulator.Query, regulator.Compute, true},  // js, billable
	{regulator.Index, regulator.Read, true},     // gsi, billable
	{regulator.Search, regulator.Read, true},    // fts, billable
	{regulator.Data, regulator.Read, true},      // kv ru, billable
	{regulator.Data, regulator.Write, true},     // kv wu, billable
}

func Init(serverless bool) {
	isServerless = serverless
}

func Start(endpoint Endpoint, nodeid string, regulatorsettingsfile string) {
	if !isServerless {
		return
	}
	handle := factory.InitRegulator(regulator.InitSettings{NodeID: service.NodeID(nodeid),
		SettingsFile: regulatorsettingsfile, Service: regulator.Query,
		ServiceCheckMask: regulator.Index | regulator.Search})
	mux := endpoint.Mux()
	tenantHandler := func(w http.ResponseWriter, req *http.Request) {
		err := endpoint.Authorize(req)
		if err != nil {
			endpoint.WriteError(err, w, req)
			return
		}
		handle.WriteMetrics(w)
	}
	mux.HandleFunc(regulator.MeteringEndpoint, tenantHandler).Methods("GET")
	mux.HandleFunc("/_prometheusMetricsHigh", tenantHandler).Methods("GET")
}

func RegisterResourceManager(m ResourceManager) {
	if !isServerless {
		return
	}
	resourceManagers = append(resourceManagers, m)
}

func IsServerless() bool {
	return isServerless
}

func AddUnit(dest *Unit, u Unit) {
	atomic.AddUint64((*atomic.AlignedUint64)(dest), uint64(u))
}

func (this Unit) String() string {
	return strconv.FormatUint(uint64(this), 10)
}

func (this Unit) NonZero() bool {
	return this > 0
}

func Throttle(isAdmin bool, user, bucket string, buckets []string, timeout time.Duration) (Context, errors.Error) {

	if isAdmin {
		return regulator.NewUserCtx(bucket, user), nil
	}
	tenant := bucket
	if tenant == "" {
		return nil, errors.NewServiceTenantMissingError()
	} else {
		found := false
		for _, b := range buckets {
			if b == tenant {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.NewServiceTenantNotAuthorizedError(bucket)
		}
	}

	ctx := regulator.NewUserCtx(tenant, user)
	r, d, e := regulator.CheckQuota(ctx, &regulator.CheckQuotaOpts{
		MaxThrottle:       timeout,
		NoThrottle:        false,
		NoReject:          false,
		EstimatedDuration: time.Duration(0),
		EstimatedUnits:    []regulator.Units{},
	})
	switch r {
	case regulator.CheckResultNormal:

		// if KV is throttling this tenant slow it down before the request starts in order
		// to use the would be KV throttling to service other less active tenants
		// the query throttling will limit KV requests which in turn will lessen KV need
		// for throttling
		throttleLock.RLock()
		lastTime, ok := throttleTimes[bucket]
		throttleLock.RUnlock()
		if ok {
			t := util.Now()
			d := t.Sub(lastTime)

			if d > time.Duration(0) {
				// TODO record throttles
				logging.Debuga(func() string { return fmt.Sprintf("bucket %v throttled for %v by query", bucket, d) })
				time.Sleep(d)
			}

			// remove delay hint to minimise cost
			throttleLock.Lock()
			currLastTime, ook := throttleTimes[bucket]
			if ook && currLastTime == lastTime {
				delete(throttleTimes, bucket)
			}
			throttleLock.Unlock()
		}
		return ctx, nil
	case regulator.CheckResultThrottle:
		time.Sleep(d)
		return ctx, nil
	case regulator.CheckResultReject:
		return nil, errors.NewServiceTenantRejectedError(d)
	default:
		return ctx, errors.NewServiceTenantThrottledError(e)
	}
}

func Bucket(ctx Context) string {
	if ctx != nil {
		return ctx.Bucket()
	}
	return ""
}
func User(ctx Context) string {
	if ctx != nil {
		return ctx.User()
	}
	return ""
}

// TODO define units for query and js-evaluator
func RecordCU(ctx Context, d time.Duration, m uint64) Unit {
	units, _ := metering.QueryEvalComputeToCU(d, m)
	regulator.RecordUnits(ctx, units)
	return Unit(units.Whole())
}

func RecordJsCU(ctx Context, d time.Duration, m uint64) Unit {
	units, _ := metering.QueryUDFComputeToCU(d, m)
	regulator.RecordUnits(ctx, units)
	return Unit(units.Whole())
}

func RefundUnits(ctx Context, units Services) error {

	// no refund needed for full admin
	if ctx.Bucket() == "" {
		return nil
	}
	for s, u := range units {
		if u.NonZero() && toReg[s].billable {
			ru, err := regulator.NewUnits(toReg[s].service, toReg[s].unit, uint64(u))
			if err != nil {
				return err
			}
			err = regulator.RefundUnits(ctx, ru)
			if err != nil {
				return nil
			}
		}
	}
	return nil
}

func Units2Map(serv Services) map[string]interface{} {
	var out []regulator.Units

	for s, u := range serv {
		if u.NonZero() && toReg[s].billable {
			ru, err := regulator.NewUnits(toReg[s].service, toReg[s].unit, uint64(u))
			if err != nil {
				continue
			}
			out = append(out, ru)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return regulator.UnitsToMap(false, out...)
}

func EncodeNodeName(name string) string {
	if isServerless {
		return distributed.RemoteAccess().NodeUUID(name)
	} else {
		return name
	}
}

func DecodeNodeName(name string) string {
	if isServerless {
		return distributed.RemoteAccess().UUIDToHost(name)
	} else {
		return name
	}
}

func Suspend(bucket string, delay time.Duration) {
	t := util.Now().Add(delay)
	throttleLock.Lock()
	oldT, ok := throttleTimes[bucket]
	if !ok || t.Sub(oldT) > time.Duration(0) {

		// TODO record throttles
		logging.Debuga(func() string { return fmt.Sprintf("bucket %v throttled to %v by KV", bucket, t) })
		throttleTimes[bucket] = t
	}
	throttleLock.Unlock()
}
