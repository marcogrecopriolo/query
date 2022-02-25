//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package planner

import (
	"github.com/couchbase/query/expression"
	base "github.com/couchbase/query/plannerbase"
)

func SargableFor(pred expression.Expression, keys expression.Expressions, missing, gsi bool,
	context *PrepareContext, aliases map[string]bool) (min, max, sum int, skeys []bool) {

	skeys = make([]bool, len(keys))
	if pred == nil {
		return
	}

	if or, ok := pred.(*expression.Or); ok {
		return sargableForOr(or, keys, missing, gsi, context, aliases)
	}

	skiped := false

	for i := 0; i < len(keys); i++ {
		// Terminate on statically-valued expression
		if keys[i].Value() != nil {
			return
		}

		s := &sargable{keys[i], missing, gsi, context, aliases}

		r, err := pred.Accept(s)

		if err != nil {
			return
		}

		if r.(bool) {
			max = i + 1
			skeys[i] = true
			sum++
		} else {
			if !gsi {
				return
			}
			skiped = true
		}

		if !skiped {
			min = max
		}

		if gsi {
			missing = true
		}
	}

	return
}

func sargableForOr(or *expression.Or, keys expression.Expressions, missing, gsi bool,
	context *PrepareContext, aliases map[string]bool) (min, max, sum int, skeys []bool) {

	for _, c := range or.Operands() {
		cmin, cmax, csum, cskeys := SargableFor(c, keys, missing, gsi, context, aliases)
		if (cmin == 0 && !missing) || cmax == 0 || csum < cmin {
			skeys = make([]bool, len(keys))
			return 0, 0, 0, skeys
		}

		if min == 0 || min > cmin {
			min = cmin
			skeys = cskeys
		}

		if max == 0 || max < cmax {
			max = cmax
			skeys = cskeys
		}

		sum += csum
	}

	return
}

type sargable struct {
	key     expression.Expression
	missing bool
	gsi     bool
	context *PrepareContext
	aliases map[string]bool
}

// Arithmetic

func (this *sargable) VisitAdd(pred *expression.Add) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitDiv(pred *expression.Div) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitMod(pred *expression.Mod) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitMult(pred *expression.Mult) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitNeg(pred *expression.Neg) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitSub(pred *expression.Sub) (interface{}, error) {
	return this.visitDefault(pred)
}

// Case

func (this *sargable) VisitSearchedCase(pred *expression.SearchedCase) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitSimpleCase(pred *expression.SimpleCase) (interface{}, error) {
	return this.visitDefault(pred)
}

// Collection

func (this *sargable) VisitArray(pred *expression.Array) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitEvery(pred *expression.Every) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitExists(pred *expression.Exists) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitFirst(pred *expression.First) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitObject(pred *expression.Object) (interface{}, error) {
	return this.visitDefault(pred)
}

// Comparison

func (this *sargable) VisitBetween(pred *expression.Between) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitEq(pred *expression.Eq) (interface{}, error) {
	return this.visitBinary(pred)
}

func (this *sargable) VisitLE(pred *expression.LE) (interface{}, error) {
	return this.visitBinary(pred)
}

func (this *sargable) VisitLike(pred *expression.Like) (interface{}, error) {
	return this.visitLike(pred)
}

func (this *sargable) VisitLT(pred *expression.LT) (interface{}, error) {
	return this.visitBinary(pred)
}

func (this *sargable) VisitIsMissing(pred *expression.IsMissing) (interface{}, error) {
	if this.missing && pred.Operand().EquivalentTo(this.key) {
		return true, nil
	}

	return this.visitDefault(pred)
}

func (this *sargable) VisitIsNotMissing(pred *expression.IsNotMissing) (interface{}, error) {
	return this.visitUnary(pred)
}

func (this *sargable) VisitIsNotNull(pred *expression.IsNotNull) (interface{}, error) {
	return this.visitUnary(pred)
}

func (this *sargable) VisitIsNotValued(pred *expression.IsNotValued) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitIsNull(pred *expression.IsNull) (interface{}, error) {
	return this.visitUnary(pred)
}

func (this *sargable) VisitIsValued(pred *expression.IsValued) (interface{}, error) {
	return this.visitUnary(pred)
}

// Concat
func (this *sargable) VisitConcat(pred *expression.Concat) (interface{}, error) {
	return this.visitDefault(pred)
}

// Constant
func (this *sargable) VisitConstant(pred *expression.Constant) (interface{}, error) {
	return this.visitDefault(pred)
}

// Identifier
func (this *sargable) VisitIdentifier(pred *expression.Identifier) (interface{}, error) {
	return this.visitDefault(pred)
}

// Construction

func (this *sargable) VisitArrayConstruct(pred *expression.ArrayConstruct) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitObjectConstruct(pred *expression.ObjectConstruct) (interface{}, error) {
	return this.visitDefault(pred)
}

// Logic

func (this *sargable) VisitNot(pred *expression.Not) (interface{}, error) {
	return this.visitUnary(pred)
}

// Navigation

func (this *sargable) VisitElement(pred *expression.Element) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitField(pred *expression.Field) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitFieldName(pred *expression.FieldName) (interface{}, error) {
	return this.visitDefault(pred)
}

func (this *sargable) VisitSlice(pred *expression.Slice) (interface{}, error) {
	return this.visitDefault(pred)
}

// Self
func (this *sargable) VisitSelf(pred *expression.Self) (interface{}, error) {
	return this.visitDefault(pred)
}

// Function
func (this *sargable) VisitFunction(pred expression.Function) (interface{}, error) {
	switch pred := pred.(type) {
	case *expression.RegexpLike:
		return this.visitLike(pred)
	}

	return this.visitDefault(pred)
}

// Subquery
func (this *sargable) VisitSubquery(pred expression.Subquery) (interface{}, error) {
	return this.visitDefault(pred)
}

// NamedParameter
func (this *sargable) VisitNamedParameter(pred expression.NamedParameter) (interface{}, error) {
	return this.visitDefault(pred)
}

// PositionalParameter
func (this *sargable) VisitPositionalParameter(pred expression.PositionalParameter) (interface{}, error) {
	return this.visitDefault(pred)
}

// Cover
func (this *sargable) VisitCover(pred *expression.Cover) (interface{}, error) {
	return pred.Covered().Accept(this)
}

// All
func (this *sargable) VisitAll(pred *expression.All) (interface{}, error) {
	return pred.Array().Accept(this)
}

func (this *sargable) visitDefault(pred expression.Expression) (bool, error) {
	return this.defaultSargable(pred), nil
}

func (this *sargable) defaultSargable(pred expression.Expression) bool {
	return base.SubsetOf(pred, this.key) ||
		((pred.PropagatesMissing() || pred.PropagatesNull()) &&
			pred.DependsOn(this.key))
}
