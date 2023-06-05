//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package value

import (
	"io"
	"sync"

	json "github.com/couchbase/go_json"
	"github.com/couchbase/query/util"
)

// A Value with delayed parsing.
type marshalledValue struct {
	raw        interface{}
	marshalled Value
	sync.RWMutex
}

func NewMarshalledValue(raw interface{}) Value {

	rv := &marshalledValue{}
	rv.raw = raw
	return rv
}

func (this *marshalledValue) String() string {
	return this.unwrap().String()
}

func (this *marshalledValue) ToString() string {
	return this.unwrap().String()
}

func (this *marshalledValue) MarshalJSON() ([]byte, error) {
	return this.unwrap().MarshalJSON()
}

func (this *marshalledValue) WriteXML(order []string, w io.Writer, prefix, indent string, fast bool) error {
	return this.unwrap().WriteXML(order, w, prefix, indent, fast)
}

func (this *marshalledValue) WriteJSON(order []string, w io.Writer, prefix, indent string, fast bool) error {
	return this.unwrap().WriteJSON(order, w, prefix, indent, fast)
}

func (this *marshalledValue) WriteSpill(w io.Writer, buf []byte) error {
	return this.unwrap().WriteSpill(w, buf)
}

func (this *marshalledValue) ReadSpill(r io.Reader, buf []byte) error {
	return this.unwrap().ReadSpill(r, buf)
}

func (this *marshalledValue) Type() Type {
	return this.unwrap().Type()
}

func (this *marshalledValue) Actual() interface{} {
	return this.unwrap().Actual()
}

func (this *marshalledValue) ActualForIndex() interface{} {
	return this.unwrap().ActualForIndex()
}

func (this *marshalledValue) Equals(other Value) Value {
	return this.unwrap().Equals(other)
}

func (this *marshalledValue) EquivalentTo(other Value) bool {
	return this.unwrap().EquivalentTo(other)
}

func (this *marshalledValue) Collate(other Value) int {
	return this.unwrap().Collate(other)
}

func (this *marshalledValue) Compare(other Value) Value {
	return this.unwrap().Compare(other)
}

func (this *marshalledValue) Truth() bool {
	return this.unwrap().Truth()
}

func (this *marshalledValue) Copy() Value {
	return this.unwrap().Copy()
}

func (this *marshalledValue) CopyForUpdate() Value {
	return this.unwrap().CopyForUpdate()
}

// Delayed parsing
func (this *marshalledValue) Field(field string) (Value, bool) {
	return this.unwrap().Field(field)
}

func (this *marshalledValue) SetField(field string, val interface{}) error {
	return this.unwrap().SetField(field, val)
}

func (this *marshalledValue) UnsetField(field string) error {
	return this.unwrap().UnsetField(field)
}

func (this *marshalledValue) Index(index int) (Value, bool) {
	return this.unwrap().Index(index)
}

func (this *marshalledValue) SetIndex(index int, val interface{}) error {
	return this.unwrap().SetIndex(index, val)
}

func (this *marshalledValue) Slice(start, end int) (Value, bool) {
	return this.unwrap().Slice(start, end)
}

func (this *marshalledValue) SliceTail(start int) (Value, bool) {
	return this.unwrap().SliceTail(start)
}

func (this *marshalledValue) Descendants(buffer []interface{}) []interface{} {
	return this.unwrap().Descendants(buffer)
}

func (this *marshalledValue) ParsedFields(min, max string, re interface{}) []interface{} {
	parsed, ok := this.unwrap().(interface {
		ParsedFields(string, string, interface{}) []interface{}
	})
	if ok {
		return parsed.ParsedFields(min, max, re)
	}
	return nil
}

func (this *marshalledValue) Fields() map[string]interface{} {
	return this.unwrap().Fields()
}

func (this *marshalledValue) FieldNames(buffer []string) []string {
	return this.unwrap().FieldNames(buffer)
}

func (this *marshalledValue) DescendantPairs(buffer []util.IPair) []util.IPair {
	return this.unwrap().DescendantPairs(buffer)
}

func (this *marshalledValue) Successor() Value {
	return this.unwrap().Successor()
}

func (this *marshalledValue) Track() {
	this.unwrap().Track()
}

func (this *marshalledValue) Recycle() {
	this.Lock()
	this.raw = nil
	if this.marshalled != nil {
		this.marshalled.Recycle()
		this.marshalled = nil
	}
	this.Unlock()
}

func (this *marshalledValue) Tokens(set *Set, options Value) *Set {
	return this.unwrap().Tokens(set, options)
}

func (this *marshalledValue) ContainsToken(token, options Value) bool {
	return this.unwrap().ContainsToken(token, options)
}

func (this *marshalledValue) ContainsMatchingToken(matcher MatchFunc, options Value) bool {
	return this.unwrap().ContainsMatchingToken(matcher, options)
}

func (this *marshalledValue) Size() uint64 {
	return this.unwrap().Size()
}

// Delayed parse.
func (this *marshalledValue) unwrap() Value {
	if this.raw != nil {
		this.Lock()
		if this.raw != nil {
			marshalled, err := json.Marshal(this.raw)
			if err != nil {
				this.marshalled = NewNullValue()
			} else {
				this.marshalled = NewParsedValue(marshalled, true)
			}
			this.raw = nil
		}
		this.Unlock()
	}
	return this.marshalled
}
