//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included in
//  the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
//  file, in accordance with the Business Source License, use of this software will
//  be governed by the Apache License, Version 2.0, included in the file
//  licenses/APL.txt.

package planner

import (
	"github.com/couchbase/query/expression"
)

func (this *sargable) VisitAny(pred *expression.Any) (interface{}, error) {
	if this.defaultSargable(pred) {
		return true, nil
	}

	all, ok := this.key.(*expression.All)
	if !ok {
		return false, nil
	}

	array, ok := all.Array().(*expression.Array)
	if !ok {
		bindings := pred.Bindings()
		return len(bindings) == 1 &&
				!bindings[0].Descend() &&
				bindings[0].Expression().EquivalentTo(all.Array()),
			nil
	}

	if !pred.Bindings().SubsetOf(array.Bindings()) {
		return false, nil
	}

	satisfies, err := getSatisfies(pred, this.key, array, this.aliases)
	if err != nil {
		return false, nil
	}

	if array.When() != nil && !checkSubset(satisfies, array.When(), this.context) {
		return false, nil
	}

	mappings := expression.Expressions{array.ValueMapping()}
	min, _, _, _ := SargableFor(satisfies, mappings, this.missing, this.gsi, this.context, this.aliases)
	return min > 0, nil
}

func getSatisfies(pred, key expression.Expression, array *expression.Array, aliases map[string]bool) (
	satisfies expression.Expression, err error) {
	var pBindings expression.Bindings
	switch p := pred.(type) {
	case *expression.Any:
		satisfies = p.Satisfies()
		pBindings = p.Bindings()
	case *expression.AnyEvery:
		satisfies = p.Satisfies()
		pBindings = p.Bindings()
	}
	if expression.HasRenameableBindings(pred, key, aliases) == expression.BINDING_VARS_DIFFER {
		renamer := expression.NewRenamer(pBindings, array.Bindings())
		return renamer.Map(satisfies.Copy())
	}
	return satisfies, nil
}
