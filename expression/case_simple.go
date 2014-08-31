//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package expression

import (
	"github.com/couchbaselabs/query/value"
)

type WhenTerms []*WhenTerm

type WhenTerm struct {
	When Expression
	Then Expression
}

type SimpleCase struct {
	ExpressionBase
	searchTerm Expression
	whenTerms  WhenTerms
	elseTerm   Expression
}

func NewSimpleCase(searchTerm Expression, whenTerms WhenTerms, elseTerm Expression) Expression {
	return &SimpleCase{
		searchTerm: searchTerm,
		whenTerms:  whenTerms,
		elseTerm:   elseTerm,
	}
}

func (this *SimpleCase) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitSimpleCase(this)
}

func (this *SimpleCase) Type() value.Type {
	t := value.NULL

	if this.elseTerm != nil {
		t = this.elseTerm.Type()
	}

	for _, w := range this.whenTerms {
		tt := w.Then.Type()
		if t > value.NULL && tt > value.NULL && tt != t {
			return value.JSON
		} else {
			t = tt
		}
	}

	return t
}

func (this *SimpleCase) Evaluate(item value.Value, context Context) (value.Value, error) {
	s, err := this.searchTerm.Evaluate(item, context)
	if err != nil {
		return nil, err
	}

	if s.Type() <= value.NULL {
		return s, nil
	}

	for _, w := range this.whenTerms {
		wv, err := w.When.Evaluate(item, context)
		if err != nil {
			return nil, err
		}

		if s.Equals(wv) {
			tv, err := w.Then.Evaluate(item, context)
			if err != nil {
				return nil, err
			}

			return tv, nil
		}
	}

	if this.elseTerm == nil {
		return value.NULL_VALUE, nil
	}

	ev, err := this.elseTerm.Evaluate(item, context)
	if err != nil {
		return nil, err
	}

	return ev, nil
}

func (this *SimpleCase) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *SimpleCase) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *SimpleCase) Children() Expressions {
	rv := make(Expressions, 0, 2+(len(this.whenTerms)<<1))

	rv = append(rv, this.searchTerm)
	for _, w := range this.whenTerms {
		rv = append(rv, w.When)
		rv = append(rv, w.Then)
	}

	if this.elseTerm != nil {
		rv = append(rv, this.elseTerm)
	}

	return rv
}

func (this *SimpleCase) MapChildren(mapper Mapper) (err error) {
	this.searchTerm, err = mapper.Map(this.searchTerm)
	if err != nil {
		return
	}

	for _, w := range this.whenTerms {
		w.When, err = mapper.Map(w.When)
		if err != nil {
			return
		}

		w.Then, err = mapper.Map(w.Then)
		if err != nil {
			return
		}
	}

	if this.elseTerm != nil {
		this.elseTerm, err = mapper.Map(this.elseTerm)
		if err != nil {
			return
		}
	}

	return
}
