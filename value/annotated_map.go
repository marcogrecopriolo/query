//  Copyright 2022-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package value

import (
	"bufio"
	"compress/zlib"
	"container/heap"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/logging"
	"github.com/couchbase/query/util"
)

const _INITIAL_MAP_SIZE = 1024

type annotatedMapEntry struct {
	key string
	val AnnotatedValue
}

type mapSpillFile struct {
	f       *os.File
	reader  io.Reader
	current *annotatedMapEntry
	sz      int64

	read time.Duration

	compress bool
}

func (this *mapSpillFile) rewind(trackMemory func(int64), afterRead func(AnnotatedValue)) error {
	_, err := this.f.Seek(0, os.SEEK_SET)
	if err != nil {
		return errors.NewValueError(errors.E_VALUE_SPILL_READ, err)
	}
	this.reader = bufio.NewReaderSize(this.f, 64*util.KiB)
	if this.compress {
		this.reader, _ = zlib.NewReader(this.reader)
	}
	err = this.nextValue(trackMemory, afterRead)
	if err != nil {
		if _, ok := err.(errors.Error); ok {
			return err
		}
		return errors.NewValueError(errors.E_VALUE_SPILL_READ, err)
	}
	return nil
}

func (this *mapSpillFile) Read(b []byte) (int, error) {
	return io.ReadFull(this.reader, b)
}

func (this *mapSpillFile) nextValue(trackMemory func(int64), afterRead func(AnnotatedValue)) error {
	this.current = nil
	s := time.Now()
	k, err := readSpillValue(this, nil)
	this.read += time.Now().Sub(s)
	if err != nil {
		if err == io.EOF {
			return err
		}
		return errors.NewValueError(errors.E_VALUE_RECONSTRUCT, err)
	}
	key, ok := k.(string)
	if !ok {
		return errors.NewValueError(errors.E_VALUE_INVALID)
	}
	s = time.Now()
	v, err := readSpillValue(this, nil)
	this.read += time.Now().Sub(s)
	if err != nil {
		if err == io.EOF {
			return err
		}
		return errors.NewValueError(errors.E_VALUE_RECONSTRUCT, err)
	}
	av, ok := v.(AnnotatedValue)
	if !ok {
		return errors.NewValueError(errors.E_VALUE_INVALID)
	}
	if afterRead != nil {
		afterRead(av)
	}
	this.current = &annotatedMapEntry{key: key, val: av}
	if trackMemory != nil {
		trackMemory(int64(this.current.val.Size()))
	}
	return nil
}

func (this *mapSpillFile) release() {
	if this.f != nil {
		util.ReleaseTemp(this.f.Name(), this.sz)
		this.f.Close()
	}
	this.f = nil
	if this.current != nil {
		this.current.val = nil
	}
	this.current = nil
}

type mapSpillFileHeap []*mapSpillFile

func (this *mapSpillFileHeap) Len() int { return len(*this) }
func (this *mapSpillFileHeap) Less(i, j int) bool {
	return (*this)[i].current.key < (*this)[j].current.key
}
func (this *mapSpillFileHeap) Swap(i, j int)      { (*this)[i], (*this)[j] = (*this)[j], (*this)[i] }
func (this *mapSpillFileHeap) Push(x interface{}) { *this = append(*this, x.(*mapSpillFile)) }
func (this *mapSpillFileHeap) Pop() interface{} {
	i := len(*this) - 1
	last := (*this)[i]
	*this = (*this)[:i]
	return last
}

type AnnotatedMap struct {
	sync.RWMutex
	shouldSpill func(uint64, uint64) bool
	trackMemory func(int64)
	merge       func(AnnotatedValue, AnnotatedValue) AnnotatedValue
	beforeSpill func(AnnotatedValue)
	afterRead   func(AnnotatedValue)

	inMem   map[string]AnnotatedValue
	memSize uint64
	spill   mapSpillFileHeap

	accumSpillTime time.Duration

	compress bool
}

func NewAnnotatedMap(
	shouldSpill func(uint64, uint64) bool,
	trackMemory func(int64),
	merge func(AnnotatedValue, AnnotatedValue) AnnotatedValue,
	beforeSpill func(AnnotatedValue),
	afterRead func(AnnotatedValue),
	compressSpill bool) *AnnotatedMap {

	rv := &AnnotatedMap{
		shouldSpill: shouldSpill,
		trackMemory: trackMemory,
		merge:       merge,
		beforeSpill: beforeSpill,
		afterRead:   afterRead,
		inMem:       make(map[string]AnnotatedValue, _INITIAL_MAP_SIZE),
		compress:    compressSpill,
	}

	return rv
}

func (this *AnnotatedMap) Copy() *AnnotatedMap {
	rv := &AnnotatedMap{
		shouldSpill: this.shouldSpill,
		trackMemory: this.trackMemory,
		merge:       this.merge,
		beforeSpill: this.beforeSpill,
		afterRead:   this.afterRead,
		inMem:       make(map[string]AnnotatedValue, _INITIAL_MAP_SIZE),
		compress:    this.compress,
	}
	return rv
}

func (this *AnnotatedMap) Load(key string) AnnotatedValue {
	this.RLock()
	rv, ok := this.inMem[key]
	this.RUnlock()
	if !ok {
		return nil
	}
	return rv
}

func (this *AnnotatedMap) Store(key string, av AnnotatedValue) errors.Error {
	var err errors.Error
	this.Lock()
	existing, ok := this.inMem[key]
	if ok {
		this.memSize -= existing.Size()
		err = this.setLOCKED(key, 0, av)
	} else {
		err = this.setLOCKED(key, uint64(len(key)), av)
	}
	this.Unlock()
	return err
}

func (this *AnnotatedMap) LoadOrStore(key string, av AnnotatedValue) (AnnotatedValue, bool, errors.Error) {
	this.Lock()
	existing, ok := this.inMem[key]
	if ok {
		this.Unlock()
		return existing, false, nil
	}
	err := this.setLOCKED(key, uint64(len(key)), av)
	this.Unlock()
	return av, true, err
}

func (this *AnnotatedMap) setLOCKED(key string, keySize uint64, av AnnotatedValue) errors.Error {

	if this.wouldSpill(keySize + av.Size()) {
		if err := this.spillToDisk(); err != nil {
			return err
		}
	}
	this.inMem[key] = av
	this.memSize += keySize + av.Size()
	return nil
}

func (this *AnnotatedMap) spillToDisk() errors.Error {
	var err error
	if this.memSize == 0 || len(this.inMem) == 0 {
		// nothing to spill
		return nil
	}
	start := time.Now()
	spill := &mapSpillFile{compress: this.compress}
	spill.f, err = util.CreateTemp(_SPILL_FILE_PATTERN, true)
	if err != nil {
		return errors.NewValueError(errors.E_VALUE_SPILL_CREATE, err)
	}
	var writer writerFlusher
	if this.compress {
		writer = zlib.NewWriter(spill.f)
	} else {
		writer = bufio.NewWriter(spill.f)
	}
	buf := _SPILL_POOL.Get()
	imsz := this.memSize
	// this is a notable compromise: to keep non-spilling perfomance we use a map but this must then be sorted in order
	// to facilitate efficient merging of spilled maps.  The overhead of allocating space to duplicate the keys for sorting
	// could be notable which is contrary to the reducing memory use, which is the point of spilling.  Since they are strings,
	// the values need not be duplicated themselves so this should be minor in most cases
	keys := make([]string, 0, len(this.inMem))
	for k := range this.inMem {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lastSize := int64(0)
	spaceCheckInterval := int(len(keys) / 10)
	if spaceCheckInterval < 1 {
		spaceCheckInterval = 1
	}
	for n, k := range keys {
		me := this.inMem[k]
		err = writeSpillValue(writer, k, buf)
		if err != nil {
			_SPILL_POOL.Put(buf)
			return errors.NewValueError(errors.E_VALUE_SPILL_WRITE, err)
		}
		if this.beforeSpill != nil {
			this.beforeSpill(me)
		}
		err = writeSpillValue(writer, me, buf)
		if err != nil {
			_SPILL_POOL.Put(buf)
			return errors.NewValueError(errors.E_VALUE_SPILL_WRITE, err)
		}
		if this.trackMemory != nil {
			this.trackMemory(-int64(me.Size()))
		}
		this.memSize -= (uint64(len(k)) + me.Size())
		me.Recycle()
		delete(this.inMem, k)

		if n%spaceCheckInterval == 0 {
			spill.sz, err = spill.f.Seek(0, os.SEEK_CUR)
			if err != nil {
				return errors.NewValueError(errors.E_VALUE_SPILL_SIZE, err)
			}
			if !util.UseTemp(spill.f.Name(), spill.sz-lastSize) {
				return errors.NewTempFileQuotaExceededError()
			}
			lastSize = spill.sz
		}
	}
	_SPILL_POOL.Put(buf)
	writer.Flush()
	spill.sz, err = spill.f.Seek(0, os.SEEK_CUR)
	if err != nil {
		return errors.NewValueError(errors.E_VALUE_SPILL_SIZE, err)
	}
	if !util.UseTemp(spill.f.Name(), spill.sz-lastSize) {
		return errors.NewTempFileQuotaExceededError()
	}
	this.spill = append(this.spill, spill)
	d := time.Now().Sub(start)
	this.accumSpillTime += d
	logging.Debuga(func() string {
		return fmt.Sprintf("[%p,%p] %v mem: %v -> temp: %v (%.3f %%)", this, spill, d, imsz, spill.sz,
			float64(spill.sz)/float64(imsz)*100.0)
	})
	return nil
}

func (this *AnnotatedMap) Foreach(f func(string, AnnotatedValue) bool) errors.Error {
	this.Lock()
	returned := 0
	if this.spill != nil {
		for i := range this.spill {
			err := this.spill[i].rewind(this.trackMemory, this.afterRead)
			if err != nil {
				this.Unlock()
				logging.Debugf("[%p] rewind failed on [%d] %s: %v", this, i, this.spill[i].f.Name(), err)
				return errors.NewValueError(errors.E_VALUE_SPILL_READ, err)
			}
		}
		sort.Sort(&this.spill)
		err := this.mergeKeys()
		if err != nil {
			this.Unlock()
			return errors.NewValueError(errors.E_VALUE_SPILL_READ, err)
		}
		// in-memory map needs to be accessed in sorted key order for merging with spilled maps
		keys := make([]string, 0, len(this.inMem))
		for k := range this.inMem {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		n := 0
		for n < len(keys) || len(this.spill) > 0 {
			if n < len(keys) && (len(this.spill) == 0 || keys[n] < this.spill[0].current.key) {
				if !f(keys[n], this.inMem[keys[n]]) {
					this.Unlock()
					return nil
				}
				delete(this.inMem, keys[n])
				n++
			} else if len(this.spill) > 0 && (n >= len(keys) || keys[n] > this.spill[0].current.key) {
				if !f(this.spill[0].current.key, this.spill[0].current.val) {
					this.Unlock()
					return nil
				}
				err := this.spill[0].nextValue(this.trackMemory, this.afterRead)
				if err != io.EOF && err != nil {
					this.Unlock()
					return errors.NewValueError(errors.E_VALUE_SPILL_READ, err)
				} else if err == io.EOF {
					d := heap.Pop(&this.spill).(*mapSpillFile)
					d.release()
					this.accumSpillTime += d.read
				} else {
					heap.Fix(&this.spill, 0)
				}
				err = this.mergeKeys()
				if err != nil {
					this.Unlock()
					return errors.NewValueError(errors.E_VALUE_SPILL_READ, err)
				}
			} else {
				if this.trackMemory != nil {
					this.trackMemory(-int64(this.inMem[keys[n]].Size()))
					this.trackMemory(-int64(this.spill[0].current.val.Size()))
				}
				merged := this.merge(this.inMem[keys[n]], this.spill[0].current.val)
				if this.trackMemory != nil {
					this.trackMemory(int64(merged.Size()))
				}
				if !f(keys[n], merged) {
					this.Unlock()
					return nil
				}
				// recycle as these will never flow out of this container if they aren't the merged result
				if merged != this.inMem[keys[n]] {
					this.inMem[keys[n]].Recycle()
				}
				if merged != this.spill[0].current.val {
					this.spill[0].current.val.Recycle()
				}
				delete(this.inMem, keys[n])
				n++
				err := this.spill[0].nextValue(this.trackMemory, this.afterRead)
				if err != io.EOF && err != nil {
					this.Unlock()
					return errors.NewValueError(errors.E_VALUE_SPILL_READ, err)
				} else if err == io.EOF {
					d := heap.Pop(&this.spill).(*mapSpillFile)
					d.release()
					this.accumSpillTime += d.read
				} else {
					heap.Fix(&this.spill, 0)
				}
				err = this.mergeKeys()
				if err != nil {
					this.Unlock()
					return errors.NewValueError(errors.E_VALUE_SPILL_READ, err)
				}
			}
			returned++
		}
	} else {
		for k, v := range this.inMem {
			if !f(k, v) {
				break
			}
			returned++
		}
	}
	this.Release() // foreach is a one-off...
	this.Unlock()
	logging.Debugf("[%p] items: %v, accumSpillTime: %v", this, returned, this.accumSpillTime)
	return nil
}

func (this *AnnotatedMap) mergeKeys() error {
	for len(this.spill) > 1 {
		// a heap doesn't maintain an entirely sorted order but does guarantee the smallest will be popped
		// popping and pushing (if need be) appears on par with sorting the array, especially since we'd typically expect the
		// spill array to be small (<100 items) we're not going to see huge gains with full sorting
		// compare smallest two
		a := heap.Pop(&this.spill).(*mapSpillFile)
		b := heap.Pop(&this.spill).(*mapSpillFile)
		if a.current.key == b.current.key {
			if this.trackMemory != nil {
				this.trackMemory(-int64(a.current.val.Size()))
				this.trackMemory(-int64(b.current.val.Size()))
			}
			merged := this.merge(a.current.val, b.current.val)
			if this.trackMemory != nil {
				this.trackMemory(int64(merged.Size()))
			}
			if merged != a.current.val {
				a.current.val.Recycle() // because it will never flow out of this container
				a.current.val = merged
			}
			if merged != b.current.val {
				b.current.val.Recycle() // because it will never flow out of this container
			}
			heap.Push(&this.spill, a)
			err := b.nextValue(this.trackMemory, this.afterRead)
			if err != nil && err != io.EOF {
				return err
			} else if err == io.EOF {
				b.release()
				this.accumSpillTime += b.read
			} else {
				heap.Push(&this.spill, b)
			}
		} else {
			heap.Push(&this.spill, b)
			heap.Push(&this.spill, a)
			break
		}
	}
	return nil
}

func (this *AnnotatedMap) Release() {
	for i := range this.spill {
		this.spill[i].release()
		this.accumSpillTime += this.spill[i].read
	}
	this.spill = nil
	if len(this.inMem) > 0 {
		if len(this.inMem) <= _INITIAL_MAP_SIZE {
			for k, _ := range this.inMem {
				delete(this.inMem, k)
			}
		} else {
			this.inMem = make(map[string]AnnotatedValue, _INITIAL_MAP_SIZE)
		}
	}
	this.memSize = 0
}

// Returns if the size to be added would result in spilling
func (this *AnnotatedMap) wouldSpill(size uint64) bool {
	return this.shouldSpill != nil && this.memSize > 0 && this.shouldSpill(this.memSize, size)
}

// Adjust size of map
// if size added causes spilling - spill map to disk
func (this *AnnotatedMap) AdjustSize(size uint64) errors.Error {
	this.Lock()
	this.memSize += size

	// if added size would cause spilling - spill the map
	if this.wouldSpill(0) {
		if err := this.spillToDisk(); err != nil {
			this.Unlock()
			return err
		}
	}

	this.Unlock()
	return nil
}
