//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package algebra

import (
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/value"
)

/*
Represents a subquery statement. It inherits from
ExpressionBase since the result representation of
the subquery is an expression and contains a field
that refers to the select statement to represent
the subquery.
*/
type Subquery struct {
	expression.ExpressionBase
	query *Select
}

/*
The function NewSubquery returns a pointer to the
Subquery struct by assigning the input attributes
to the fields of the struct.
*/
func NewSubquery(query *Select) *Subquery {
	rv := &Subquery{
		query: query,
	}

	rv.SetExpr(rv)
	return rv
}

/*
Representation as a N1QL string.
*/
func (this *Subquery) String() string {
	var s string
	if this.IsCorrelated() || this.query.subresult.IsCorrelated() {
		s += "correlated "
	}

	return s + "(" + this.query.String() + ")"
}

/*
Visitor pattern.
*/
func (this *Subquery) Accept(visitor expression.Visitor) (interface{}, error) {
	return visitor.VisitSubquery(this)
}

/*
Subqueries return a value of type ARRAY.
*/
func (this *Subquery) Type() value.Type { return value.ARRAY }

func (this *Subquery) Evaluate(item value.Value, context expression.Context) (value.Value, error) {
	return context.(Context).EvaluateSubquery(this.query, item)
}

/*
Return false. Subquery cannot be used as a secondary
index key.
*/
func (this *Subquery) Indexable() bool {
	return false
}

/*
Return false.
*/
func (this *Subquery) IndexAggregatable() bool {
	return false
}

/*
Return false.
*/
func (this *Subquery) EquivalentTo(other expression.Expression) bool {
	return false
}

/*
Return false.
*/
func (this *Subquery) SubsetOf(other expression.Expression) bool {
	return false
}

/*
Return inner query's Expressions.
*/
func (this *Subquery) Children() expression.Expressions {
	return this.query.Expressions()
}

/*
Map inner query's Expressions.
*/
func (this *Subquery) MapChildren(mapper expression.Mapper) error {
	return this.query.MapExpressions(mapper)
}

/*
Return this subquery expression.
*/
func (this *Subquery) Copy() expression.Expression {
	return this
}

/*
TODO: This is overly broad. Ideally, we would allow:

SELECT g, (SELECT d2.* FROM d2 USE KEYS d.g) AS d2
FROM d
GROUP BY g;

but not allow:

SELECT g, (SELECT d2.* FROM d2 USE KEYS d.foo) AS d2
FROM d
GROUP BY g;
*/

func (this *Subquery) SurvivesGrouping(groupKeys expression.Expressions, allowed *value.ScopeValue) (
	bool, expression.Expression) {

	if !this.query.IsCorrelated() {
		return true, nil
	}

	// If the subquery is correlated only with the Group As alias, then it can survive grouping
	for _, v := range this.GetCorrelation() {
		if (v & expression.IDENT_IS_GROUP_AS) == 0 {
			return false, nil
		}
	}

	return true, nil
}

/*
This method calls FormalizeSubquery to qualify all the children
of the query, and returns an error if any.
*/
func (this *Subquery) Formalize(parent *expression.Formalizer) error {
	if parent != nil && parent.IsCheckCorrelation() {
		// when checking correlation, do not go into subqueries,
		// since the subqueries themselves are marked by CORRELATED
		return nil
	}

	err := this.query.FormalizeSubquery(parent, true)
	if err != nil {
		return err
	}

	// if the subquery is correlated, add the correlation reference to
	// the parent formalizer such that any nested correlation can be detected
	// at the next level
	if this.query.IsCorrelated() {
		err = parent.AddCorrelatedIdentifiers(this.query.GetCorrelation())
	}

	return err
}

/*
Returns the subquery select statement, namely the input
query.
*/
func (this *Subquery) Select() *Select {
	return this.query
}

func (this *Subquery) IsCorrelated() bool {
	return this.query.IsCorrelated()
}

func (this *Subquery) GetCorrelation() map[string]uint32 {
	return this.query.GetCorrelation()
}

func (this *Subquery) SetInFunction(hasVariables bool) {
	this.query.inlineFunc = true
	if hasVariables {
		this.query.hasVariables = true
	}
}

func (this *Subquery) CoveredBy(keyspace string, exprs expression.Expressions, options expression.CoveredOptions) expression.Covered {
	// Check if COVER_IN_SUBQUERY flag is set already
	// To prevent erroneous unset of the flag - only set & later unset flag if not already set
	set := options.InSubqueryTraversal()

	if !set {
		options.SetInSubqueryFlag()
	}

	rv := expression.CoveredSkip

	// Only consider the subquery for the covering check
	// if it is correlated with the keyspace
	if this.IsCorrelated() {
		correlated := this.query.GetCorrelation()
		if c, ok := correlated[keyspace]; ok {
			if c&expression.IDENT_IS_KEYSPACE > 0 {
				rv = this.ExprBase().CoveredBy(keyspace, exprs, options)
			}
		}
	}

	if !set {
		options.UnsetInSubqueryFlag()
	}

	return rv
}
