//  Copyright 2017-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package planner

import (
	"github.com/couchbase/query/algebra"
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/logging"
	"github.com/couchbase/query/plan"
	"github.com/couchbase/query/util"
	"github.com/couchbase/query/value"
)

func (this *builder) indexPushDownProperty(entry *indexEntry, keys,
	unnestFiletrs expression.Expressions, pred, origPred expression.Expression,
	alias string, unnestAliases []string, unnest, covering, allKeyspaces, implicitAny bool) (
	pushDownProperty PushDownProperties) {

	// Check all predicates are part of spans, exact and no false positives possible
	exact := allKeyspaces && !this.hasBuilderFlag(BUILDER_HAS_EXTRA_FLTR) &&
		this.checkExactSpans(entry, pred, origPred, alias, unnestAliases, unnestFiletrs, implicitAny)
	if exact {
		pushDownProperty |= _PUSHDOWN_EXACTSPANS
	}

	// Covering index check for other pushdowns
	if covering && exact {
		pushDownProperty |= this.indexCoveringPushDownProperty(entry, keys, alias,
			unnestAliases, unnest, implicitAny, pushDownProperty)
	}

	vector := entry.HasFlag(IE_VECTOR_KEY_SARGABLE)
	idxKeys := entry.idxKeys

	// Check Query Order By matches with index key order.
	exactLimitOffset := exact
	if this.order != nil {
		if this.group == nil || isPushDownProperty(pushDownProperty, _PUSHDOWN_FULLGROUPAGGS) {
			if exact && vector {
				idxKeys, _ = replaceVectorKey(idxKeys, entry)
			}
			ok, _, partSortCount := this.useIndexOrder(entry, idxKeys)
			logging.Debugf("indexPushDownProperty: ok: %v, count: %v", ok, partSortCount)
			if ok {
				pushDownProperty |= _PUSHDOWN_ORDER
			} else {
				exactLimitOffset = false
				if partSortCount > 0 && partSortCount < len(this.order.Terms()) && !indexHasFlattenKeys(entry.index) {
					entry.partialSortTermCount = partSortCount
					pushDownProperty |= _PUSHDOWN_PARTIAL_ORDER
				}
			}
		} else {
			exactLimitOffset = false
		}
	} else if this.group != nil && !isPushDownProperty(pushDownProperty, _PUSHDOWN_FULLGROUPAGGS) {
		exactLimitOffset = false
	}

	// Check all predicates are part of spans, exact and no false positives possible
	if isPushDownProperty(pushDownProperty, _PUSHDOWN_EXACTSPANS) {

		// LIMIT Pushdown is possible when
		//        *  Query Order By not present
		//        *  Query Order By matches with Index key order
		//        *  LIMIT is hint to indexer

		if this.limit != nil && exactLimitOffset {
			pushDownProperty |= _PUSHDOWN_LIMIT
		} else if vector {
			// if determined above that ORDER can be pushed down but LIMIT cannot,
			// need to unset ORDER pushdown
			order := isPushDownProperty(pushDownProperty, _PUSHDOWN_ORDER)
			partOrder := isPushDownProperty(pushDownProperty, _PUSHDOWN_PARTIAL_ORDER)
			if order || partOrder {
				if order {
					pushDownProperty &^= _PUSHDOWN_ORDER
				}
				for i := 0; i < len(idxKeys); i++ {
					if _, ok := idxKeys[i].Expr.(*expression.Ann); ok || idxKeys[i].HasAttribute(datastore.IK_VECTOR) {
						// vector key
						if order {
							if i > 0 {
								pushDownProperty |= _PUSHDOWN_PARTIAL_ORDER
								entry.partialSortTermCount = i
							}
						} else {
							entry.partialSortTermCount = i
							if i == 0 {
								pushDownProperty &^= _PUSHDOWN_PARTIAL_ORDER
							}
						}
						break
					}
				}
			}
		}

		// OFFSET Pushdown is possible when
		//        *  Index API2
		//        *  Offset can be pushed based on spans becasue OFFSET needs to be exact NOT hint to indexer
		//        *  Query Order By not present
		//        *  Query Order By matches with Index key order

		if this.offset != nil && exactLimitOffset &&
			useIndex2API(entry.index, this.context.IndexApiVersion()) &&
			entry.spans.CanPushDownOffset(entry.index, overlapSpans(pred),
				!unnest && indexHasArrayIndexKey(entry.index)) {
			pushDownProperty |= _PUSHDOWN_OFFSET
		}
	}
	if this.indexAdvisor {
		this.collectPushdownProperty(entry.index, alias, pushDownProperty)
	}
	return pushDownProperty
}

func (this *builder) indexCoveringPushDownProperty(entry *indexEntry,
	indexKeys expression.Expressions, alias string, unnestAliases []string,
	unnest, implicitAny bool, pushDownProperty PushDownProperties) PushDownProperties {

	// spans needs to be exact
	if !isPushDownProperty(pushDownProperty, _PUSHDOWN_EXACTSPANS) {
		return pushDownProperty
	}

	// Check aggregate pushdowns using API3
	aggProperty, tryApi2 := this.indexAggPushDownProperty(entry, indexKeys, alias,
		unnestAliases, unnest, implicitAny, pushDownProperty)
	pushDownProperty |= aggProperty

	// Exploiting IndexScan for aggregates using API2/API1.
	//          * COUNT(), COUNT(DISTINCT op ), MIN(), MAX(),
	//          * Requires single aggregate in projection
	//          * NO Group By

	if !isPushDownProperty(pushDownProperty, _PUSHDOWN_GROUPAGGS) && !this.joinEnum() &&
		this.oldAggregates && tryApi2 && len(this.aggs) == 1 && len(this.group.By()) == 0 &&
		!indexHasArrayIndexKey(entry.index) {
		for _, ag := range this.aggs {
			switch agg := ag.(type) {

			case *algebra.Count:
				if this.canPushDownCount(entry, agg.Operands()[0], indexKeys, agg.Distinct()) {
					pushDownProperty |= _PUSHDOWN_GROUPAGGS
				}

			case *algebra.Min:
				if this.canPushDownMinMax(entry, agg.Operands()[0], indexKeys, false) {
					pushDownProperty |= _PUSHDOWN_GROUPAGGS
				}

			case *algebra.Max:
				if this.canPushDownMinMax(entry, agg.Operands()[0], indexKeys, true) {
					pushDownProperty |= _PUSHDOWN_GROUPAGGS
				}
			}
		}
	}

	// Check Projection Distinct can be pushdown
	if !isPushDownProperty(pushDownProperty, _PUSHDOWN_DISTINCT) &&
		(this.group == nil || isPushDownProperty(pushDownProperty, _PUSHDOWN_GROUPAGGS)) &&
		this.canPushDownProjectionDistinct(entry, this.projection, entry.keys) {
		pushDownProperty |= _PUSHDOWN_DISTINCT
	}

	return pushDownProperty
}

func (this *builder) indexAggPushDownProperty(entry *indexEntry, indexKeys expression.Expressions,
	alias string, unnestAliases []string, unnest, implicitAny bool,
	pushDownProperty PushDownProperties) (PushDownProperties, bool) {

	if this.group == nil || !useIndex3API(entry.index, this.context.IndexApiVersion()) ||
		!isPushDownProperty(pushDownProperty, _PUSHDOWN_EXACTSPANS) ||
		isPushDownProperty(pushDownProperty, _PUSHDOWN_GROUPAGGS) ||
		!util.IsFeatureEnabled(this.context.FeatureControls(), util.N1QL_GROUPAGG_PUSHDOWN) {
		return pushDownProperty, !implicitAny
	}

	// Group keys needs to be covered by index keys (including document key)
	for _, gexpr := range this.group.By() {
		if !isImplicitCovered(gexpr, indexKeys, alias, unnestAliases, implicitAny, entry.arrayKey) {
			return pushDownProperty, false
		}
	}

	groupMatch, maxKeyPos := this.indexGroupLeadingIndexKeysMatch(entry, indexKeys)

	arrayIndexIsOK := isValidAggregateArrayIndex(entry, unnest)

nextagg:
	for _, agg := range this.aggs {
		op := agg.Operands()[0]

		constOp := (op == nil || op.Value() != nil)

		// aggregate expression needs to be covered by index keys (including document key)
		if !constOp {
			if !isImplicitCovered(op, indexKeys, alias, unnestAliases, implicitAny, entry.arrayKey) {
				return pushDownProperty, false
			}
		}

		switch agg.(type) {
		case *algebra.Min, *algebra.Max:
			continue nextagg
		default:
			// Distinct aggregates argument can be any key in the matched leading keys + 0|1
			// 0 for partition index and 1 for non partition index
			if agg.Distinct() {
				if groupMatch {
					if constOp {
						continue nextagg
					}

					for i, key := range indexKeys {
						if i <= maxKeyPos && op.EquivalentTo(key) {
							continue nextagg
						}
					}
				}

				// Distinct key not matched leading keys + 0+1
				return pushDownProperty, false
			} else {
				// non Unnest IndexScan can only use DISTINCT array index with eqaulity preidicate
				// Unnest IndexScan can only use DISTINCT array with Distinct aggregates

				if !arrayIndexIsOK {
					return pushDownProperty, false
				}
			}
		}
	}

	pushDownProperty |= _PUSHDOWN_GROUPAGGS
	if groupMatch {
		pushDownProperty |= _PUSHDOWN_FULLGROUPAGGS
	}
	return pushDownProperty, false
}

func isImplicitCovered(expr expression.Expression, indexKeys expression.Expressions, alias string,
	unnestAliases []string, implicitAny bool, arrayKey *expression.All) bool {
	if !expression.IsCovered(expr, alias, indexKeys, implicitAny) {
		return false
	}
	for _, a := range unnestAliases {
		if !expression.IsCovered(expr, a, indexKeys, implicitAny) {
			return false
		}
	}

	// Check Any clause in the expression.
	// EVERY, ANY AND EVERY will not implicitly cover
	if implicitAny {
		mapAnys, err := expression.GatherAny(expression.Expressions{expr}, arrayKey, false)
		if err != nil || len(mapAnys) > 0 {
			return false
		}
	}

	return true

}

/*
 For array Index to use for  non-MIN,non-MAX aggregates
    For non Unnest Scan
       * Equality predicate on array Index required
       * DISTINCT on the array Index key required
       * ALL is ok if it is leading key. Indexer will consider one row for META().id
 For Unnest Scan ALL index or DISTINCT index with Distinct aggregates can use it
*/

func isValidAggregateArrayIndex(entry *indexEntry, unnest bool) bool {
	if entry.arrayKey == nil || entry.arrayKeyPos < 0 {
		return true
	}
	pos := entry.arrayKeyPos
	all := entry.arrayKey
	noDistinct := entry.arrayKey.NoDistinct()
	noAll := entry.arrayKey.NoAll()
	if unnest {
		return noDistinct
	} else if all.Flatten() && all.FlattenSize() > 1 {
		eq := true
		for i := 0; i < all.FlattenSize(); i++ {
			eq, _ = entry.spans.EquivalenceRangeAt(i + pos)
			if !eq {
				return false
			}
		}
		return eq && (noAll || pos == 0)
	} else {
		eq, _ := entry.spans.EquivalenceRangeAt(pos)
		return eq && (noAll || pos == 0)
	}
	return true
}

func (this *builder) indexGroupLeadingIndexKeysMatch(entry *indexEntry, indexKeys expression.Expressions) (bool, int) {

	// generate unique group keys.
	groupkeys := make(map[string]bool, len(this.group.By())+1)
	for _, gexpr := range this.group.By() {
		// ignore constants
		if gexpr.Value() != nil {
			continue
		}
		groupkeys[gexpr.String()] = true
	}

	// For Partition index the partition keys needs to be in group keys to use Distinct aggregates
	if len(groupkeys) > 0 {
		for _, pexpr := range entry.partitionKeys {
			if _, ok := groupkeys[pexpr.String()]; !ok {
				return false, 0
			}
		}
	} else if len(entry.partitionKeys) > 0 {
		// no group keys present, every partition key present in the index keys and equivalent span then
		// it resolves single partition. Let apply non-partition index rules.

		idxKeys := make(map[string]int, len(indexKeys))
		for i, iexpr := range indexKeys {
			idxKeys[iexpr.String()] = i
		}
		for _, pexpr := range entry.partitionKeys {
			i, ok := idxKeys[pexpr.String()]
			if !ok {
				return false, 0
			}
			if eq, _ := entry.spans.EquivalenceRangeAt(i); !eq {
				return false, 0
			}
		}
	}

	// Check group keys matching leading keys. If equality predicate that index key can be skipped in group
	nMatched := 0
	nGroupMatched := 0
	for nMatched < len(indexKeys) {
		if _, ok := groupkeys[indexKeys[nMatched].String()]; ok {
			// index key matched with group key, check duplicate index keys
			duplicate := false
			for k := 0; !duplicate && k <= nMatched-1; k++ {
				if indexKeys[nMatched].EquivalentTo(indexKeys[k]) {
					duplicate = true
				}
			}
			nMatched++
			if !duplicate {
				nGroupMatched++
			}
		} else if eq, _ := entry.spans.EquivalenceRangeAt(nMatched); eq {
			// index key is equality predicate, skip it
			nMatched++
		} else {
			// not matched break it
			break
		}
	}

	// Check all group keys matched with leading index keys
	return (nGroupMatched == len(groupkeys)), nMatched
}

func (this *builder) checkExactSpans(entry *indexEntry, pred, origPred expression.Expression, alias string,
	unnestAliases []string, unnestFiletrs expression.Expressions, implicitAny bool) bool {
	// spans are not exact
	if !entry.exactSpans || hasUnknownsInSargableArrayKey(entry) || entry.HasFlag(IE_OR_NON_SARG_EXPR) {
		return false
	}
	if pred == nil {
		return this.where == nil
	}

	// check for non sargable key is in predicate
	exprs, _, err := indexCoverExpressions(entry, entry.idxSargKeys, pred, nil, alias, this.context)
	if err != nil {
		return false
	}

	if this.aggConstraint != nil {
		exprs = append(exprs, this.aggConstraint)
	}

	if len(unnestFiletrs) > 0 {
		exprs = append(exprs, unnestFiletrs...)
	}

	if !expression.IsCovered(pred, alias, exprs, implicitAny) {
		if origPred == nil || !expression.IsCovered(origPred, alias, exprs, implicitAny) {
			return false
		}
	}

	for _, a := range unnestAliases {
		if !expression.IsCovered(pred, a, exprs, implicitAny) {
			return false
		}
	}

	// all predicates are part of spans, exact and no false positives possible
	return true
}

func (this *builder) canPushDownCount(entry *indexEntry, op expression.Expression,
	keys expression.Expressions, distinct bool) bool {

	// COUNT( DISTINCT op) is supported in API2
	if distinct && !useIndex2API(entry.index, this.context.IndexApiVersion()) {
		return false
	}

	// no operand can be push down
	if op == nil {
		return !distinct
	}

	// constant non-MISSING, non-NULL operand can be pushdown
	val := op.Value()
	if val != nil {
		return val.Type() > value.NULL
	}

	// operand needs to be leading key
	if len(keys) == 0 || !op.EquivalentTo(keys[0]) {
		return false
	}

	// Scan should not include NULL or MISSING
	return entry.spans.SkipsLeadingNulls()
}

func (this *builder) canPushDownMinMax(entry *indexEntry, op expression.Expression, keys expression.Expressions,
	max bool) bool {
	// aggregate operand is constant
	if op.Value() != nil {
		return true
	}

	// MAX() pushdown is supported in API2 only
	if max && !useIndex2API(entry.index, this.context.IndexApiVersion()) {
		return false
	}

	// aggregate operand needs to be leading key
	if len(keys) == 0 || !op.EquivalentTo(keys[0]) {
		return false
	}

	// get the index collation of the leading key
	descCollation := indexKeyIsDescCollation(0, entry.idxKeys)
	if max {
		// MAX() can be pushdown when leading index key is DESC collation
		return descCollation && entry.spans.CanUseIndexOrder(false)
	} else {
		// MIN() can be pushdown when leading index key is ASC collation and NULLS are not included
		return !descCollation && entry.spans.CanUseIndexOrder(false) && entry.spans.SkipsLeadingNulls()
	}
}

func (this *builder) canPushDownProjectionDistinct(entry *indexEntry, projection *algebra.Projection,
	indexKeys expression.Expressions) bool {

	// Only supported in API2
	if projection == nil || !useIndex2API(entry.index, this.context.IndexApiVersion()) || !projection.Distinct() {
		return false
	}

	// Disable distinct pushdown for HASH partition. API3
	if useIndex3API(entry.index, this.context.IndexApiVersion()) {
		partition, err := entry.index.(datastore.Index3).PartitionKeys()
		if err != nil || (partition != nil && partition.Strategy != datastore.NO_PARTITION) {
			return false
		}
	}
	if indexHasVector(entry.index) {
		return false
	}

	hash := _STRING_BOOL_POOL.Get()
	defer _STRING_BOOL_POOL.Put(hash)

	for _, key := range indexKeys {
		hash[key.String()] = true
	}

	// all projections needs to be part of the index keys
	for _, expr := range projection.Expressions() {
		if expr.Value() == nil {
			if _, ok := hash[expr.String()]; !ok {
				return false
			}
		}
	}

	return true
}

func (this *builder) useIndexOrder(entry *indexEntry, keys datastore.IndexKeys) (bool, plan.IndexKeyOrders, int) {

	// Force the use of sorts on indexes that we know not to be ordered
	// (for now system indexes)
	// for now - if they are of a non descript type, then they aren't sorted
	// when GSI starts implementing other types of indexes (eg bitmap)
	// we will revisit this approach

	if entry.index.Type() == datastore.SYSTEM || /*entry.index.Type() == datastore.SEQ_SCAN ||*/
		!entry.spans.CanUseIndexOrder(useIndex3API(entry.index, this.context.IndexApiVersion())) {
		return false, nil, 0
	}

	logging.Debugf("useIndexOrder: entry: %v, order: %v", entry, this.order.Terms())

	vector := entry.HasFlag(IE_VECTOR_KEY_SARGABLE)
	if vector && entry.IsPushDownProperty(_PUSHDOWN_ORDER) {
		// for vector index key, only do key replacement if we've previously determined that
		// order can be pushed down, otherwise key replacement needs to be performed by caller
		var err error
		keys, err = replaceVectorKey(keys, entry)
		if err != nil {
			logging.Debugf("useIndexOrder: replaceVectorKey returns error %v", err)
			return false, nil, 0
		}
	}

	var filters map[string]value.Value
	if entry.cond != nil {
		filters = _FILTER_COVERS_POOL.Get()
		defer _FILTER_COVERS_POOL.Put(filters)
		filters = entry.cond.FilterCovers(filters)
		filters = entry.origCond.FilterCovers(filters)
	}

	var hashProj map[string]expression.Expression

	if this.projection != nil {
		hashProj = make(map[string]expression.Expression, len(this.projection.Terms()))
		for _, term := range this.projection.Terms() {
			hashProj[term.Alias()] = term.Expression()
		}
	}

	i := 0
	indexOrder := make(plan.IndexKeyOrders, 0, len(keys))
	partSortTermCount := 0
	vectorOrder := false
	if vector {
		if _, ok := entry.spans.(*TermSpans); ok {
			vectorOrder = true
		}
	}
outer:
	for _, orderTerm := range this.order.Terms() {

		// if sort order or nulls order are named/positional parameters or function parameters i.e non constants
		// then we can't use index order
		if (orderTerm.DescendingExpr() != nil &&
			(orderTerm.DescendingExpr().Indexable() == false || orderTerm.DescendingExpr().Value() == nil)) ||
			(orderTerm.NullsPosExpr() != nil &&
				(orderTerm.NullsPosExpr().Indexable() == false || orderTerm.NullsPosExpr().Value() == nil)) {
			return false, indexOrder, partSortTermCount
		}

		// orderTerm is constant
		if orderTerm.Expression().Static() != nil {
			partSortTermCount++
			continue
		}

		// non-constant orderTerms are more than index keys
		if i >= len(keys) {
			// match with condition EQ terms
			if equalConditionFilter(filters, orderTerm.Expression().String()) {
				partSortTermCount++
				continue outer
			}
			// index order gives us partial sorting
			return false, indexOrder, partSortTermCount
		}

		d := orderTerm.Descending(nil, nil)
		nl := orderTerm.NullsLast(nil, nil)
		naturalOrder := false
		if orderTerm.IsVectorTerm() {
			naturalOrder = !d && nl
		} else if d && nl {
			naturalOrder = true
		} else if !d && !nl {
			naturalOrder = true
		}
		for {
			projexpr, projalias := hashProj[orderTerm.Expression().Alias()]
			if indexKeyIsDescCollation(i, keys) == d &&
				(naturalOrder || !entry.spans.CanProduceUnknowns(i)) &&
				(orderTerm.Expression().EquivalentTo(keys[i].Expr) ||
					(projalias && expression.Equivalent(projexpr, keys[i].Expr))) {
				// orderTerm matched with index key
				if vector {
					if _, ok := keys[i].Expr.(*expression.Ann); ok {
						// vector key
						if !vectorOrder {
							return false, indexOrder, partSortTermCount
						}
						// once we pass the vector key, no need to check for
						// eq range anymore
						vectorOrder = false
					} else if vectorOrder {
						// non-vector key
						// can only push order for vector key if all previous
						// keys are EQ only
						if eq, _ := entry.spans.EquivalenceRangeAt(i); !eq {
							vectorOrder = false
						}
					}
				}
				indexOrder = append(indexOrder, plan.NewIndexKeyOrders(i, d))
				i++
				partSortTermCount++
				continue outer
			} else if equalConditionFilter(filters, orderTerm.Expression().String()) {
				// orderTerm matched with Condition EQ
				partSortTermCount++
				continue outer
			} else if eq, _ := entry.spans.EquivalenceRangeAt(i); eq {
				// orderTerm not yet matched, but can skip equivalence range key, don't add to indexOrder
				i++
				if i >= len(keys) {
					return false, indexOrder, partSortTermCount
				}
			} else {
				return false, indexOrder, partSortTermCount
			}
		}
	}

	return true, indexOrder, 0 // complete sorting via index
}

func equalConditionFilter(filters map[string]value.Value, str string) bool {
	if filters == nil {
		return false
	}

	v, ok := filters[str]
	return ok && v != nil
}

func indexKeyIsDescCollation(keypos int, indexKeys datastore.IndexKeys) bool {
	return len(indexKeys) > 0 && keypos < len(indexKeys) && indexKeys[keypos].HasAttribute(datastore.IK_DESC)
}
