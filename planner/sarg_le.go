//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package planner

import (
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/plan"
	base "github.com/couchbase/query/plannerbase"
)

func (this *sarg) VisitLE(pred *expression.LE) (interface{}, error) {
	key := this.key.Expr
	if base.SubsetOf(pred, key) {
		if expression.Equivalent(pred, key) {
			return _EXACT_SELF_SPANS, nil
		}
		return _SELF_SPANS, nil
	}

	var expr expression.Expression
	range2 := &plan.Range2{}

	selec := this.getSelec(pred)

	if pred.First().EquivalentTo(key) {
		expr = this.getSarg(pred.Second())
		range2.Low = expression.NULL_EXPR
		range2.High = expr
		range2.Inclusion = datastore.HIGH
		range2.Selec1 = OPT_SELEC_NOT_AVAIL
		range2.Selec2 = selec
	} else if pred.Second().EquivalentTo(key) {
		expr = this.getSarg(pred.First())
		range2.Low = expr
		range2.Inclusion = datastore.LOW
		range2.Selec1 = selec
		range2.Selec2 = OPT_SELEC_NOT_AVAIL
		if pred.HasExprFlag(expression.EXPR_DERIVED_FROM_LIKE) {
			range2.SetFlag(plan.RANGE_DERIVED_FROM_LIKE)
		}
	} else if pred.DependsOn(key) {
		return _VALUED_SPANS, nil
	} else {
		return nil, nil
	}

	if expr == nil {
		return _VALUED_SPANS, nil
	}

	span := plan.NewSpan2(nil, plan.Ranges2{range2}, true)
	return NewTermSpans(span), nil
}
