//  Copieright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package value

import (
	"io"

	"github.com/couchbase/query/util"
)

/*
Type Empty struct
*/
type nullValue struct {
}

/*
Initialized as a pointer to an empty nullValue.
*/
var NULL_VALUE Value = &nullValue{}

/*
Returns a NULL_VALUE.
*/
func NewNullValue() Value {
	return NULL_VALUE
}

var _NULL_BYTES = []byte("null")

func (this *nullValue) String() string {
	return "null"
}

func (this *nullValue) MarshalJSON() ([]byte, error) {
	return _NULL_BYTES, nil
}

func (this *nullValue) WriteJSON(w io.Writer, prefix, indent string) error {
	_, err := w.Write(_NULL_BYTES)
	return err
}

/*
Type NULL
*/
func (this *nullValue) Type() Type {
	return NULL
}

/*
Returns nil.
*/
func (this *nullValue) Actual() interface{} {
	return nil
}

/*
Returns MISSING or NULL.
*/
func (this *nullValue) Equals(other Value) Value {
	other = other.unwrap()
	switch other.Type() {
	case MISSING:
		return other
	default:
		return this
	}
}

func (this *nullValue) EquivalentTo(other Value) bool {
	return other.Type() == NULL
}

/*
Returns the relative position of null wrt other.
*/
func (this *nullValue) Collate(other Value) int {
	return int(NULL - other.Type())
}

func (this *nullValue) Compare(other Value) Value {
	other = other.unwrap()
	switch other := other.(type) {
	case missingValue:
		return other
	default:
		return this
	}
}

/*
Returns false.
*/
func (this *nullValue) Truth() bool {
	return false
}

/*
Return receiver.
*/
func (this *nullValue) Copy() Value {
	return this
}

/*
Return receiver.
*/
func (this *nullValue) CopyForUpdate() Value {
	return this
}

/*
Calls missingField.
*/
func (this *nullValue) Field(field string) (Value, bool) {
	return missingField(field), false
}

/*
Not valid for NULL.
*/
func (this *nullValue) SetField(field string, val interface{}) error {
	return Unsettable(field)
}

/*
Not valid for NULL.
*/
func (this *nullValue) UnsetField(field string) error {
	return Unsettable(field)
}

/*
Calls missingIndex.
*/
func (this *nullValue) Index(index int) (Value, bool) {
	return missingIndex(index), false
}

/*
Not valid for NULL.
*/
func (this *nullValue) SetIndex(index int, val interface{}) error {
	return Unsettable(index)
}

/*
Returns NULL_VALUE
*/
func (this *nullValue) Slice(start, end int) (Value, bool) {
	return NULL_VALUE, false
}

/*
Returns NULL_VALUE
*/
func (this *nullValue) SliceTail(start int) (Value, bool) {
	return NULL_VALUE, false
}

/*
Returns the input buffer as is.
*/
func (this *nullValue) Descendants(buffer []interface{}) []interface{} {
	return buffer
}

/*
Null has no fields to list. Hence return nil.
*/
func (this *nullValue) Fields() map[string]interface{} {
	return nil
}

func (this *nullValue) FieldNames(buffer []string) []string {
	return nil
}

/*
Returns the input buffer as is.
*/
func (this nullValue) DescendantPairs(buffer []util.IPair) []util.IPair {
	return buffer
}

/*
NULL is succeeded by FALSE.
*/
func (this *nullValue) Successor() Value {
	return FALSE_VALUE
}

func (this nullValue) Recycle() {
}

func (this *nullValue) unwrap() Value {
	return this
}
