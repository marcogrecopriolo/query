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
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/plan"
	base "github.com/couchbase/query/plannerbase"
)

func SargFor(pred, vpred expression.Expression, entry *indexEntry, keys datastore.IndexKeys,
	isMissing bool, isArrays []bool, max int, isJoin, doSelec bool, baseKeyspace *base.BaseKeyspace,
	keyspaceNames map[string]string, advisorValidate bool, aliases map[string]bool,
	context *PrepareContext) (SargSpans, bool, error) {

	// Optimize top-level OR predicate
	if or, ok := pred.(*expression.Or); ok {
		return sargForOr(or, vpred, entry, keys, isMissing, isArrays, max, isJoin, doSelec,
			baseKeyspace, keyspaceNames, advisorValidate, aliases, context)
	}

	sargKeys := keys[0:max]

	// Get sarg spans for index sarg keys. The sarg spans are
	// truncated when they exceed the limit.
	sargSpans, exactSpan, err := getSargSpans(pred, vpred, entry, sargKeys, isMissing, isArrays,
		isJoin, doSelec, baseKeyspace, keyspaceNames, advisorValidate, aliases, context)
	if sargSpans == nil || err != nil {
		return nil, exactSpan, err
	}

	return composeSargSpan(sargSpans, exactSpan)
}

func sargForOr(or *expression.Or, vpred expression.Expression, entry *indexEntry, keys datastore.IndexKeys,
	isMissing bool, isArrays []bool, max int, isJoin, doSelec bool, baseKeyspace *base.BaseKeyspace,
	keyspaceNames map[string]string, advisorValidate bool, aliases map[string]bool,
	context *PrepareContext) (SargSpans, bool, error) {

	exact := true
	hasVector := false
	spans := make([]SargSpans, len(or.Operands()))
	for i, c := range or.Operands() {
		// Variable length sarging
		_, max1, _, _ := SargableFor(c, vpred, entry.index, keys, isMissing, true, isArrays, context, aliases)
		if max1 == 0 {
			max1 = 1
		}
		s, ex, err := SargFor(c, vpred, entry, keys, isMissing, isArrays, max1, isJoin, doSelec,
			baseKeyspace, keyspaceNames, advisorValidate, aliases, context)
		if err != nil {
			return nil, false, err
		}

		spans[i] = s
		exact = exact && ex
		hasVector = hasVector || s.HasVector()

		if exact && !entry.HasFlag(IE_OR_NON_SARG_EXPR) {
			setFlag := false
			if max1 < max {
				// check for non-sargable key in predicate
				exprs, _, err := indexCoverExpressions(entry, keys[:max1], c, nil, baseKeyspace.Name(), context)
				if err != nil {
					return nil, false, err
				}
				implicitAny := implicitAnyCover(entry, true, context.FeatureControls())
				if !expression.IsCovered(c, baseKeyspace.Name(), exprs, implicitAny) {
					entry.SetFlags(IE_OR_NON_SARG_EXPR, true)
					setFlag = true
				}

			}
			if !setFlag && hasConstSubExpr(c, keyspaceNames) {
				entry.SetFlags(IE_OR_NON_SARG_EXPR, true)
			}
		}
	}

	var rv SargSpans = NewUnionSpans(spans...)
	rv = rv.Streamline()
	if hasVector {
		if _, ok := rv.(*TermSpans); !ok {
			return nil, false, errors.NewPlanInternalError("sargForOr: unexpected OR predicate for vector index key: " +
				or.String())
		}
	}
	return rv, exact, nil
}

func sargFor(pred expression.Expression, index datastore.Index, key expression.Expression, isJoin, doSelec bool,
	baseKeyspace *base.BaseKeyspace, keyspaceNames map[string]string, advisorValidate, isMissing, isArray, isVector bool,
	keyPos int, aliases map[string]bool, context *PrepareContext) (SargSpans, bool, error) {

	s := newSarg(key, index, baseKeyspace, keyspaceNames, isJoin, doSelec, advisorValidate, isMissing, isArray,
		isVector, keyPos, aliases, context)

	r, err := pred.Accept(s)
	if err != nil {
		return nil, false, err
	}
	if r == nil {
		exact := true
		if s.constPred {
			exact = false
		} else if pred.DependsOn(key) {
			exact = false
		}
		return nil, exact, nil
	}

	rs := r.(SargSpans)
	return rs, rs.Exact(), nil
}

func SargForFilters(filters base.Filters, vpred expression.Expression, entry *indexEntry, keys datastore.IndexKeys,
	isMissing bool, isArrays []bool, max int, underHash, doSelec bool, baseKeyspace *base.BaseKeyspace,
	keyspaceNames map[string]string, advisorValidate bool, aliases map[string]bool,
	exactFilters map[*base.Filter]bool, context *PrepareContext) (SargSpans, bool, error) {

	sargSpans := make([]SargSpans, max)
	exactSpan := true
	arrayKeySpans := make(map[int][]SargSpans)

	sargKeys := keys[0:max]

	for _, fl := range filters {
		if fl.IsOnclause() {
			if baseKeyspace.IsOuter() && fl.NotPushable() {
				continue
			}
		} else {
			if baseKeyspace.OnclauseOnly() || baseKeyspace.IsOuter() {
				// only ON-clause filter should be used
				continue
			}
		}

		fltrExpr := fl.FltrExpr()
		isJoin := fl.IsJoin() && !underHash
		flSargSpans, flExactSpan, err := getSargSpans(fl.FltrExpr(), vpred, entry, sargKeys,
			isMissing, isArrays, isJoin, doSelec, baseKeyspace, keyspaceNames,
			advisorValidate, aliases, context)
		if err != nil {
			return nil, flExactSpan, err
		}

		if flExactSpan && exactFilters != nil {
			valid := false
			for _, rs := range flSargSpans {
				if rs != nil && rs.Size() > 0 {
					valid = true
					break
				}
			}
			if valid {
				exactFilters[fl] = true
			}
		}

		exactSpan = exactSpan && flExactSpan

		for pos, sargKey := range sargKeys {
			isArray, _, _ := sargKey.Expr.IsArrayIndexKey()
			isVector := sargKey.HasAttribute(datastore.IK_VECTOR)
			if flSargSpans[pos] == nil || flSargSpans[pos].Size() == 0 {
				if exactSpan && !isArray && !isVector && fltrExpr.DependsOn(sargKey.Expr) {
					exactSpan = false
				}
				continue
			} else if !isArray && !isVector && flSargSpans[pos] == _EMPTY_SPANS {
				return _EMPTY_SPANS, true, nil
			}

			if isArray {
				if _, ok := arrayKeySpans[pos]; !ok {
					arrayKeySpans[pos] = make([]SargSpans, 0, len(filters))
				}
				arrayKeySpans[pos] = append(arrayKeySpans[pos], flSargSpans[pos])
			} else {
				if sargSpans[pos] == nil || sargSpans[pos].Size() == 0 {
					sargSpans[pos] = flSargSpans[pos]
				} else {
					sargSpans[pos] = sargSpans[pos].Constrain(flSargSpans[pos])
					if sargSpans[pos] == _EMPTY_SPANS {
						return _EMPTY_SPANS, true, nil
					}
				}
				if isVector {
					// no need to regenerate span for vector index key
					vpred = nil
				}
			}
		}
	}

	for pos, arrayKeySpan := range arrayKeySpans {
		sargSpans[pos] = addArrayKeys(arrayKeySpan)
	}

	var hasSpan bool
	for _, s := range sargSpans {
		if s != nil {
			hasSpan = true
			break
		}
	}

	if !hasSpan && len(filters) != 0 {
		return nil, false, nil
	}

	return composeSargSpan(sargSpans, exactSpan)
}

/*
Compose SargSpan for a composite index
*/
func composeSargSpan(sargSpans []SargSpans, exactSpan bool) (SargSpans, bool, error) {
	// Truncate sarg spans when they exceed the limit
	size := 1
	n := 0
	tooMany := -1
	for i, spans := range sargSpans {
		sz := 1
		if spans != nil {
			sz = spans.Size()
		}

		if sz == 0 ||
			(sz > 1 && size >= 1 && sz*size > plan.FULL_SPAN_FANOUT) {
			exactSpan = false
			tooMany = i
			break
		}

		size *= sz
		n++
	}

	var ns SargSpans
	if n == 0 && tooMany == 0 {
		// too many spans on the first index key
		ns = _WHOLE_SPANS.Copy()
		ns.SetExact(false)
		return ns, ns.Exact(), nil
	}

	// Sarg composite indexes right to left
	for i := n - 1; i >= 0; i-- {
		rs := sargSpans[i]

		if rs == nil {
			rs = _WHOLE_SPANS.Copy()
		}
		if rs.Size() == 0 { // Reset
			ns = nil
			continue
		}

		// Start
		if ns == nil {
			ns = rs
			continue
		}

		ns = ns.Copy()
		ns = ns.Compose(rs)
		ns = ns.Streamline()

		if ns == _EMPTY_SPANS {
			return _EMPTY_SPANS, true, nil
		}
	}

	if ns == nil || ns.Size() == 0 {
		return _EMPTY_SPANS, true, nil
	}

	if ns.Exact() && !exactSpan {
		ns.SetExact(exactSpan)
	}

	return ns, ns.Exact(), nil
}

/*
Get sarg spans for index sarg keys.
*/
func getSargSpans(pred, vpred expression.Expression, entry *indexEntry, sargKeys datastore.IndexKeys,
	isMissing bool, isArrays []bool, isJoin, doSelec bool, baseKeyspace *base.BaseKeyspace,
	keyspaceNames map[string]string, advisorValidate bool, aliases map[string]bool,
	context *PrepareContext) ([]SargSpans, bool, error) {

	if pred == nil && vpred == nil {
		return nil, false, errors.NewPlanInternalError("getSargSpans: no predicates")
	}

	// is the predicate simple?
	simple := true
	switch pred.(type) {
	case *expression.And, *expression.Or, *expression.Not:
		simple = false
	}

	n := len(sargKeys)

	exactSpan := true
	sargSpans := make([]SargSpans, n)

	// Sarg composite indexes right to left
	for i := n - 1; i >= 0; i-- {
		isVector := sargKeys[i].HasAttribute(datastore.IK_VECTOR)

		spred := pred
		if isVector {
			spred = vpred
		}
		if spred == nil {
			continue
		}

		s := newSarg(sargKeys[i].Expr, entry.index, baseKeyspace, keyspaceNames, isJoin, doSelec,
			advisorValidate, (isMissing || i > 0), (i < len(isArrays) && isArrays[i]),
			isVector, i, aliases, context)
		r, err := spred.Accept(s)
		if err != nil {
			return nil, false, err
		}

		if r != nil {
			rs := r.(SargSpans)
			rs = rs.Streamline()

			sargSpans[i] = rs

			if rs.Size() == 0 {
				exactSpan = false
				continue
			}

			// If one key span is EMPTY then whole index span can be EMPTY
			if rs == _EMPTY_SPANS {
				// make sure the returned slice is of length n since
				// the caller assumes that (it'll be streamlined later)
				for j := n - 1; j >= 0; j-- {
					sargSpans[j] = _EMPTY_SPANS
				}
				return sargSpans, true, nil
			}

			if simple && !isVector {
				// if the same simple predicate can be used to sarg multiple
				// index keys, we can safely just use the exactSpan information
				// from this key and disregard that of the previous keys since
				// the index keys are walked backwards.
				// Specifically, if it generate an exact span for one of the
				// index key, we can set exactSpan to true even if it is not exact
				// for a different key (which appears after this index key)
				exactSpan = rs.Exact()
				if exactSpan {
					// if there is a _VALUED_SPANS from the same simple
					// predicate (generated when predicate depends on the key),
					// we can make that exact span as well
					// (array index cover depends on exact)
					for j := i + 1; j < n; j++ {
						os := sargSpans[j]
						if os != nil && os.Size() > 0 && !os.Exact() &&
							isSpecialSargSpan(os, plan.RANGE_VALUED_SPAN) {
							os = os.Copy()
							os.SetExact(exactSpan)
							sargSpans[j] = os
						}
					}
				}
			} else {
				exactSpan = exactSpan && rs.Exact()
			}
		} else if exactSpan {
			// if a constant or query parameters is used as a (boolean) predicate
			// then it'll not be used to generate spans, and it won't be caught
			// by covering checks later on; set exactSpan to be false in this case
			// to be safe (since this may introduce false positives from index scan)
			if s.constPred {
				exactSpan = false
			} else if spred.DependsOn(sargKeys[i].Expr) {
				exactSpan = false
			}
		}
	}

	return sargSpans, exactSpan, nil
}

// this function is called for subterms of an OR clause, it assumes the original OR expression
// is already flattened, thus it does not expect an OR expression in here
func hasConstSubExpr(pred expression.Expression, keyspaceNames map[string]string) bool {
	if pred == nil || len(keyspaceNames) == 0 {
		return false
	}

	if and, ok := pred.(*expression.And); ok {
		for _, op := range and.Operands() {
			if hasConstSubExpr(op, keyspaceNames) {
				return true
			}
		}
		return false
	}

	keyspaces, err := expression.CountKeySpaces(pred, keyspaceNames)
	if err == nil && len(keyspaces) == 0 {
		return true
	}

	return false
}
