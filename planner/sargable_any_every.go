//  Copyright 2016-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package planner

import (
	"github.com/couchbase/query/expression"
)

func (this *sargable) VisitAnyEvery(pred *expression.AnyEvery) (interface{}, error) {
	if this.defaultSargable(pred) {
		return true, nil
	}

	all, ok := this.key.(*expression.All)
	if !ok {
		return false, nil
	}

	var satisfies, mapping expression.Expression
	bindings := pred.Bindings()
	array, ok := all.Array().(*expression.Array)
	if !ok {
		if len(bindings) != 1 || bindings[0].Descend() ||
			!bindings[0].Expression().EquivalentTo(all.Array()) {
			return false, nil
		}
		bindVar := expression.NewIdentifier(bindings[0].Variable())
		bindVar.SetBindingVariable(true)
		mapping = bindVar
		satisfies = pred.Satisfies()
	} else {
		if !bindings.SubsetOf(array.Bindings()) {
			return false, nil
		}
		mapping = array.ValueMapping()

		var err error
		satisfies, err = getSatisfies(pred, this.key, array, this.aliases)
		if err != nil {
			return false, err
		}

		if array.When() != nil && !checkSubset(satisfies, array.When(), this.context) {
			return false, nil
		}
	}

	mappings := expression.Expressions{mapping}
	min, _, _, _ := SargableFor(satisfies, mappings, this.missing, this.gsi, []bool{true}, this.context, this.aliases)
	return min > 0, nil
}
