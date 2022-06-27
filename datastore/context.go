//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package datastore

import (
	"time"

	"github.com/couchbase/query/auth"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/tenant"
	"github.com/couchbase/query/value"
)

var NULL_CONTEXT Context = &contextImpl{}

var NULL_QUERY_CONTEXT QueryContext = &queryContextImpl{}

type Context interface {
	GetScanCap() int64
	MaxParallelism() int
	Fatal(errors.Error)
	Error(errors.Error)
	Warning(errors.Error)
	GetReqDeadline() time.Time
	RecordFtsRU(ru tenant.Unit)
	RecordGsiRU(ru tenant.Unit)
	RecordKvRU(ru tenant.Unit)
	RecordKvWU(wu tenant.Unit)
}

type contextImpl struct {
}

func (ci *contextImpl) GetScanCap() int64 {
	return GetScanCap()
}

func (ci *contextImpl) MaxParallelism() int {
	return 1
}

func (ci *contextImpl) Fatal(err errors.Error) {
}

func (ci *contextImpl) Error(err errors.Error) {
}

func (ci *contextImpl) Warning(err errors.Error) {
}

func (ci *contextImpl) GetReqDeadline() time.Time {
	return time.Time{}
}

func (ci *contextImpl) FeatureControl() uint64 {
	return 0
}

func (ci *contextImpl) RecordFtsRU(ru tenant.Unit) {
}

func (ci *contextImpl) RecordGsiRU(ru tenant.Unit) {
}

func (ci *contextImpl) RecordKvRU(ru tenant.Unit) {
}

func (ci *contextImpl) RecordKvWU(wu tenant.Unit) {
}

// A subset of execution.Context that is useful at the datastore level.
type QueryContext interface {
	GetReqDeadline() time.Time
	UseReplica() bool
	Credentials() *auth.Credentials
	Warning(errors.Error)
	Error(errors.Error)
	GetTxContext() interface{}
	SetTxContext(tc interface{})
	Datastore() Datastore
	TxDataVal() value.Value
	DurabilityLevel() DurabilityLevel
	KvTimeout() time.Duration
	PreserveExpiry() bool
	RecordFtsRU(ru tenant.Unit)
	RecordGsiRU(ru tenant.Unit)
	RecordKvRU(ru tenant.Unit)
	RecordKvWU(wu tenant.Unit)
}

type queryContextImpl struct {
}

func (ci *queryContextImpl) Credentials() *auth.Credentials {
	return auth.NewCredentials()
}

func (ci *queryContextImpl) Warning(err errors.Error) {
}

func (ci *queryContextImpl) Error(err errors.Error) {
}

func (ci *queryContextImpl) GetReqDeadline() time.Time {
	return time.Time{}
}

func (ci *queryContextImpl) UseReplica() bool {
	return false
}

func (ci *queryContextImpl) GetTxContext() interface{} {
	return nil
}

func (ci *queryContextImpl) Datastore() Datastore {
	return GetDatastore()
}

func (ci *queryContextImpl) SetTxContext(tc interface{}) {
}

func (ci *queryContextImpl) TxDataVal() value.Value {
	return nil
}

func (ci *queryContextImpl) DurabilityLevel() DurabilityLevel {
	return DL_NONE
}

func (ci *queryContextImpl) KvTimeout() time.Duration {
	return DEF_KVTIMEOUT
}

func (ci *queryContextImpl) PreserveExpiry() bool {
	return false
}

func (ci *queryContextImpl) RecordFtsRU(ru tenant.Unit) {
}

func (ci *queryContextImpl) RecordGsiRU(ru tenant.Unit) {
}

func (ci *queryContextImpl) RecordKvRU(ru tenant.Unit) {
}

func (ci *queryContextImpl) RecordKvWU(wu tenant.Unit) {
}
