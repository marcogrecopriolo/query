//  Copyright 2023-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package migration

// TODO much like functions/metakv this package is ns_server specific

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/couchbase/cbauth/metakv"
	"github.com/couchbase/query/distributed"
	"github.com/couchbase/query/logging"
)

const _MIGRATION_PATH = "/query/migration/"
const _MIGRATION_STATE = "/state"
const (
	_RESERVED = "reserved"
	_STARTED  = "started"
	_MIGRATED = "migrated"
)

const _GRACE = 30 * time.Minute

type waiters struct {
	lock     sync.Mutex
	cond     *sync.Cond
	released bool
}

type migrationDescriptor struct {
	Node  string    `json:"node"`
	State string    `json:"state"`
	When  time.Time `json:"when"`
}

var mapLock sync.Mutex
var waitersMap map[string]*waiters

func init() {
	waitersMap = make(map[string]*waiters)
	go metakv.RunObserveChildren(_MIGRATION_PATH, callback, make(chan struct{}))
}

func setState(state string) []byte {
	desc := &migrationDescriptor{distributed.RemoteAccess().WhoAmI(), state, time.Now()}
	out, _ := json.Marshal(desc)
	return out
}

func state(val []byte) string {
	var in migrationDescriptor
	json.Unmarshal(val, &in)
	return in.State
}

// register ownership of the migration process
func Register(what string) bool {
	return metakv.Add(_MIGRATION_PATH+what+_MIGRATION_STATE, setState(_RESERVED)) == nil
}

// this is expected to be called *only* by the node that does the migration
func Start(what string) {
	metakv.Set(_MIGRATION_PATH+what+_MIGRATION_STATE, setState(_STARTED), nil)
}

// to be called before determining if a migration is needed to restart a failed
// migration if necessary
// TODO currently we employ a simple grace period technique, find a better approach
func Resume(what string) bool {
	var in migrationDescriptor

	val, rev, err := metakv.Get(_MIGRATION_PATH + what + _MIGRATION_STATE)
	if err != nil {
		return false
	}
	if json.Unmarshal(val, &in) != nil {
		return false
	}
	if in.State == _MIGRATED || time.Since(in.When) < _GRACE {
		return false
	}

	// we are in migration, have been for too long, we need to take evasive action
	if in.Node != "" {

		// it was us
		if in.Node == distributed.RemoteAccess().WhoAmI() {
			if in.State == _STARTED {
				return true
			}

			// since we hadn't started, just reset the state
			metakv.Delete(_MIGRATION_PATH+what+_MIGRATION_STATE, rev)
			return false
		}

		found := false
		for _, n := range distributed.RemoteAccess().GetNodeNames() {
			if in.Node == n {
				found = true
				break
			}
		}

		// the node is there, but stuff isn't happening
		if found {

			// TODO is there a better strategy here?
			// we cannot take evasive action because the node is still operating
			// it either completes, or it'll have to restart to clear after itself
			logging.Infof("Node %v is migrating %v, but not operating - please restart the node", in.Node, what)
			return false
		}

	} else {
		logging.Infof("A node outside of the cluster is running migration %v - attempting to correct", what)
	}

	// if the node is not found but was migrating, we'll try to start migration
	// immediately
	// if it wasn't started, we'll reset the state
	if in.State == _STARTED {
		return metakv.Set(_MIGRATION_PATH+what+_MIGRATION_STATE, setState(in.State), rev) == nil
	}

	// if it hadn't started, reset the state and we'll likely pick up the tab
	metakv.Delete(_MIGRATION_PATH+what+_MIGRATION_STATE, rev)
	return false
}

// this is expected to be called *only* by the node that does the migration
func Complete(what string) {
	err := metakv.Set(_MIGRATION_PATH+what+_MIGRATION_STATE, setState(_MIGRATED), nil)
	if err != nil {
		logging.Warnf("Migration of %v - cannot switch to completed, err %v: please restart node",
			what, err)
	}

	// We rely on the metakv callback to wake up the waiters
}

// used to move to a migrated state without actually doing any migration
// for when a new cluster is detected
func TryComplete(what string) bool {
	e1 := metakv.Add(_MIGRATION_PATH+what+_MIGRATION_STATE, setState(_MIGRATED))
	if e1 == nil {
		return true
	}
	val, _, e2 := metakv.Get(_MIGRATION_PATH + what + _MIGRATION_STATE)
	return e2 == nil && state(val) == _MIGRATED
}

// determine if migration ca be skipped
func IsComplete(what string) bool {
	val, _, err := metakv.Get(_MIGRATION_PATH + what + _MIGRATION_STATE)
	return err == nil && state(val) == _MIGRATED
}

// checking for migration to complete and waiting is it hasn't
func Await(what string) {
	val, _, err := metakv.Get(_MIGRATION_PATH + what + _MIGRATION_STATE)
	if err == nil && state(val) == _MIGRATED {
		return
	}

	// no dice
	mapLock.Lock()
	w := waitersMap[what]
	if w != nil {
		mapLock.Unlock()
		w.cond.L.Lock()
		if w.released {
			w.cond.L.Unlock()
			return
		}
		w.cond.Wait()

		w.cond.L.Unlock()
		return
	}

	// add migration
	w = &waiters{}
	w.cond = sync.NewCond(&w.lock)
	w.cond.L.Lock()
	waitersMap[what] = w
	mapLock.Unlock()
	w.cond.Wait()

	// wait leaves the lock locked on exit
	w.cond.L.Unlock()
}

// migration callback
func callback(kve metakv.KVEntry) error {
	path := string(kve.Path)
	if !strings.HasPrefix(path, _MIGRATION_PATH) ||
		!strings.HasSuffix(path, _MIGRATION_STATE) {
		return nil
	}

	// this is a good place to hook in a migration callback if we want
	// to offer the option of reacting to changing migration states

	if state(kve.Value) != _MIGRATED {
		return nil
	}
	what := path[len(_MIGRATION_PATH):]
	what = what[:len(what)-len(_MIGRATION_STATE)]
	mapLock.Lock()
	w := waitersMap[what]
	if w != nil {
		mapLock.Unlock()
		logging.Infof("Migration releasing waiters for %v", what)
		w.cond.L.Lock()
		w.released = true
		w.cond.L.Unlock()
		w.cond.Broadcast()
	} else {

		// no waiters found, but record state for posterity
		// just in case somebody tries to wait on a migrated topic
		w = &waiters{}
		w.cond = sync.NewCond(&w.lock)
		w.cond.L.Lock()
		waitersMap[what] = w
		mapLock.Unlock()
		logging.Infof("Migration complete with no waiters for %v", what)
	}
	return nil
}
