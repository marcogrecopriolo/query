//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package expression

import (
	"fmt"

	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/value"
)

/*
Bit flags for formalizer flags
*/
const (
	FORM_MAP_SELF     = 1 << iota // Map SELF to keyspace: used in sarging index
	FORM_MAP_KEYSPACE             // Map keyspace to SELF: used in creating index
	FORM_INDEX_SCOPE              // formalizing index key or index condition
	FORM_IN_FUNCTION              // We are setting variables for function invocation
)

const (
	DEF_OUTNAME = "out"
)

type WithInfo struct {
	permanent   bool
	correlated  bool
	correlation map[string]uint32
}

func newWithInfo(correlated bool, correlation map[string]uint32) *WithInfo {
	return &WithInfo{
		correlated:  correlated,
		correlation: correlation,
	}
}

func newPermanentWithInfo() *WithInfo {
	return &WithInfo{
		permanent: true,
	}
}

func (this *WithInfo) Copy() *WithInfo {
	rv := &WithInfo{
		permanent:  this.permanent,
		correlated: this.correlated,
	}
	if this.correlated {
		rv.correlation = make(map[string]uint32, len(this.correlation))
		for k, v := range this.correlation {
			rv.correlation[k] = v
		}
	}
	return rv
}

func (this *WithInfo) IsCorrelated() bool {
	return this.correlated
}

func (this *WithInfo) GetCorrelation() map[string]uint32 {
	return this.correlation
}

/*
Convert expressions to full form qualified by keyspace aliases.
*/
type Formalizer struct {
	MapperBase

	keyspace    string
	withs       map[string]*WithInfo
	allowed     *value.ScopeValue
	identifiers *value.ScopeValue
	aliases     *value.ScopeValue
	flags       uint32
	correlation map[string]uint32
}

func NewFormalizer(keyspace string, parent *Formalizer) *Formalizer {
	return newFormalizer(keyspace, parent, false, false)
}

func NewSelfFormalizer(keyspace string, parent *Formalizer) *Formalizer {
	return newFormalizer(keyspace, parent, true, false)
}

func NewKeyspaceFormalizer(keyspace string, parent *Formalizer) *Formalizer {
	return newFormalizer(keyspace, parent, false, true)
}

func NewFunctionFormalizer(keyspace string, parent *Formalizer) *Formalizer {
	rv := newFormalizer(keyspace, parent, false, false)
	rv.flags |= FORM_IN_FUNCTION
	return rv
}

func newFormalizer(keyspace string, parent *Formalizer, mapSelf, mapKeyspace bool) *Formalizer {
	var pv, av value.Value
	var withs map[string]*WithInfo
	var correlation map[string]uint32

	flags := uint32(0)
	if parent != nil {
		pv = parent.allowed
		av = parent.aliases
		mapSelf = mapSelf || parent.mapSelf()
		mapKeyspace = mapKeyspace || parent.mapKeyspace()
		if len(parent.withs) > 0 {
			withs = make(map[string]*WithInfo, len(parent.withs))
			for k, v := range parent.withs {
				withs[k] = v.Copy()
			}
		}
		if len(parent.correlation) > 0 {
			correlation = make(map[string]uint32, len(parent.correlation))
			for k, v := range parent.correlation {
				correlation[k] = v
			}
		}
	}

	if mapSelf {
		flags |= FORM_MAP_SELF
	}
	if mapKeyspace {
		flags |= FORM_MAP_KEYSPACE
	}

	rv := &Formalizer{
		keyspace:    keyspace,
		withs:       withs,
		allowed:     value.NewScopeValue(make(map[string]interface{}), pv),
		identifiers: value.NewScopeValue(make(map[string]interface{}, 64), nil),
		aliases:     value.NewScopeValue(make(map[string]interface{}), av),
		flags:       flags,
		correlation: correlation,
	}

	if !mapKeyspace && keyspace != "" {
		rv.SetAllowedAlias(keyspace, true)
	}

	rv.mapper = rv
	return rv
}

func (this *Formalizer) mapSelf() bool {
	return (this.flags & FORM_MAP_SELF) != 0
}

func (this *Formalizer) mapKeyspace() bool {
	return (this.flags & FORM_MAP_KEYSPACE) != 0
}

func (this *Formalizer) InFunction() bool {
	return (this.flags & FORM_IN_FUNCTION) != 0
}

func (this *Formalizer) indexScope() bool {
	return (this.flags & FORM_INDEX_SCOPE) != 0
}

func (this *Formalizer) SetIndexScope() {
	this.flags |= FORM_INDEX_SCOPE
}

func (this *Formalizer) ClearIndexScope() {
	this.flags &^= FORM_INDEX_SCOPE
}

func (this *Formalizer) VisitAny(expr *Any) (interface{}, error) {
	err := this.PushBindings(expr.Bindings(), true)
	if err != nil {
		return nil, err
	}

	defer this.PopBindings()

	err = expr.MapChildren(this)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (this *Formalizer) VisitEvery(expr *Every) (interface{}, error) {
	err := this.PushBindings(expr.Bindings(), true)
	if err != nil {
		return nil, err
	}

	defer this.PopBindings()

	err = expr.MapChildren(this)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (this *Formalizer) VisitAnyEvery(expr *AnyEvery) (interface{}, error) {
	err := this.PushBindings(expr.Bindings(), true)
	if err != nil {
		return nil, err
	}

	defer this.PopBindings()

	err = expr.MapChildren(this)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (this *Formalizer) VisitArray(expr *Array) (interface{}, error) {
	err := this.PushBindings(expr.Bindings(), true)
	if err != nil {
		return nil, err
	}

	defer this.PopBindings()

	err = expr.MapChildren(this)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (this *Formalizer) VisitFirst(expr *First) (interface{}, error) {
	err := this.PushBindings(expr.Bindings(), true)
	if err != nil {
		return nil, err
	}

	defer this.PopBindings()

	err = expr.MapChildren(this)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (this *Formalizer) VisitObject(expr *Object) (interface{}, error) {
	err := this.PushBindings(expr.Bindings(), true)
	if err != nil {
		return nil, err
	}

	defer this.PopBindings()

	err = expr.MapChildren(this)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

/*
Formalize Identifier.
*/
func (this *Formalizer) VisitIdentifier(expr *Identifier) (interface{}, error) {
	identifier := expr.Identifier()

	ident_val, ok := this.allowed.Field(identifier)
	if ok {
		// if sarging for index, for index keys or index conditions,
		// don't match with keyspace alias
		// (i.e., don't match an index key name with a keyspace alias)
		// however if this is a keyspace alias added in previous formalization
		// process then treat it as a keyspace alias
		ident_flags := uint32(ident_val.ActualForIndex().(int64))
		keyspace_flags := ident_flags & IDENT_IS_KEYSPACE
		variable_flags := ident_flags & IDENT_IS_VARIABLE
		static_flags := ident_flags & IDENT_IS_STATIC_VAR
		unnest_flags := ident_flags & IDENT_IS_UNNEST_ALIAS
		expr_term_flags := ident_flags & IDENT_IS_EXPR_TERM
		subq_term_flags := ident_flags & IDENT_IS_SUBQ_TERM
		group_as_flags := ident_flags & IDENT_IS_GROUP_AS
		correlated_flags := ident_flags & IDENT_IS_CORRELATED
		lateral_flags := ident_flags & IDENT_IS_LATERAL_CORR
		with_flags := ident_flags & IDENT_IS_WITH_ALIAS
		if !this.indexScope() || keyspace_flags == 0 || expr.IsKeyspaceAlias() {
			this.identifiers.SetField(identifier, ident_val)
			// for user specified keyspace alias (such as alias.c1)
			// set flag to indicate it's keyspace
			if keyspace_flags != 0 && !expr.IsKeyspaceAlias() {
				expr.SetKeyspaceAlias(true)
			}
			if variable_flags != 0 && !expr.IsBindingVariable() {
				expr.SetBindingVariable(true)
			}
			if static_flags != 0 && !expr.IsStaticVariable() {
				expr.SetStaticVariable(true)
			}
			if unnest_flags != 0 && !expr.IsUnnestAlias() {
				expr.SetUnnestAlias(true)
			}
			if expr_term_flags != 0 && !expr.IsExprTermAlias() {
				expr.SetExprTermAlias(true)
			}
			if subq_term_flags != 0 && !expr.IsSubqTermAlias() {
				expr.SetSubqTermAlias(true)
			}
			if group_as_flags != 0 && !expr.IsGroupAsAlias() {
				expr.SetGroupAsAlias(true)
			}
			if !expr.IsCorrelated() && (correlated_flags != 0 || this.CheckCorrelation(identifier)) {
				expr.SetCorrelated(true)
			}
			if lateral_flags != 0 && !expr.IsLateralCorr() {
				expr.SetLateralCorr(true)
			}
			if with_flags != 0 && !expr.IsWithAlias() {
				expr.SetWithAlias(true)
			}
			return expr, nil
		}
	}

	if wi, ok := this.withs[identifier]; ok {
		if wi.correlated {
			if this.correlation == nil {
				this.correlation = make(map[string]uint32, len(wi.correlation))
			}
			for k, v := range wi.correlation {
				this.correlation[k] = v
			}
		}
		expr.SetWithAlias(true)
		return expr, nil
	}

	if this.keyspace == "" {
		return nil, errors.NewAmbiguousReferenceError(identifier, expr.ErrorContext())
	}

	if this.mapKeyspace() {
		return expr, nil
	} else {
		keyspaceIdent := NewIdentifier(this.keyspace)
		keyspaceIdent.SetKeyspaceAlias(true)
		return NewField(keyspaceIdent, NewFieldName(identifier, expr.CaseInsensitive())), nil
	}
}

/*
Formalize SELF functions defined on indexes.
*/
func (this *Formalizer) VisitSelf(expr *Self) (interface{}, error) {
	if this.mapSelf() {
		keyspaceIdent := NewIdentifier(this.keyspace)
		keyspaceIdent.SetKeyspaceAlias(true)
		return keyspaceIdent, nil
	} else {
		return expr, nil
	}
}

/*
Formalize META() functions defined on indexes.
*/
func (this *Formalizer) VisitFunction(expr Function) (interface{}, error) {
	if !this.mapKeyspace() {
		fnName := expr.Name()
		if fnName == "meta" || fnName == "search_meta" || fnName == "search_score" {
			if len(expr.Operands()) == 0 {
				if this.keyspace != "" {
					keyspaceIdent := NewIdentifier(this.keyspace)
					keyspaceIdent.SetKeyspaceAlias(true)
					var op Expression
					op = keyspaceIdent
					if fnName == "search_meta" || fnName == "search_score" {
						op = NewField(keyspaceIdent, NewFieldName(DEF_OUTNAME, false))
					}
					return expr.Constructor()(op), nil
				} else {
					return nil, errors.NewAmbiguousMetaError(fnName, expr.ErrorContext())
				}
			} else if len(expr.Operands()) == 1 && (fnName == "search_meta" || fnName == "search_score") {
				op := expr.Operands()[0]
				if keyspaceIdent, ok := op.(*Identifier); ok {
					alias := this.keyspace
					if this.keyspace == "" {
						if _, ok = this.Allowed().Field(keyspaceIdent.Alias()); ok {
							alias = keyspaceIdent.Alias()
						}
					}
					if keyspaceIdent.Alias() == alias {
						op = NewField(keyspaceIdent, NewFieldName(DEF_OUTNAME, false))
						return expr.Constructor()(op), nil
					}
				}
			}
		}
	}

	return expr, expr.MapChildren(this.mapper)
}

/*
Visitor method for Subqueries. Call formalize.
*/
func (this *Formalizer) VisitSubquery(expr Subquery) (interface{}, error) {
	err := expr.Formalize(this)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

/*
Create new scope containing bindings.
*/

func (this *Formalizer) PushBindings(bindings Bindings, push bool) (err error) {
	allowed := this.allowed
	identifiers := this.identifiers
	aliases := this.aliases

	if push {
		allowed = value.NewScopeValue(make(map[string]interface{}, len(bindings)), this.allowed)
		identifiers = value.NewScopeValue(make(map[string]interface{}, 16), this.identifiers)
		aliases = value.NewScopeValue(make(map[string]interface{}, len(bindings)), this.aliases)
	}

	var expr Expression
	var ident_flags uint32
	for _, b := range bindings {
		expr, err = this.Map(b.Expression())
		if err != nil {
			return
		}

		b.SetExpression(expr)

		// check for correlated reference in binding expr
		correlated := this.CheckCorrelated()

		if ident_val, ok := allowed.Field(b.Variable()); ok {
			ident_flags = uint32(ident_val.ActualForIndex().(int64))
			tmp_flags1 := ident_flags & IDENT_IS_KEYSPACE
			tmp_flags2 := ident_flags &^ IDENT_IS_KEYSPACE
			// when sarging index keys, allow variables used in index definition
			// to be the same as a keyspace alias
			if !this.indexScope() || tmp_flags1 == 0 || tmp_flags2 != 0 {
				var errContext string
				if b.Expression() != nil {
					errContext = b.Expression().ErrorContext()
				}
				err = errors.NewDuplicateVariableError(b.Variable(), errContext)
				return
			}
		} else {
			ident_flags = 0
		}

		ident_flags |= IDENT_IS_VARIABLE
		if b.Static() {
			ident_flags |= IDENT_IS_STATIC_VAR
		}
		if correlated {
			ident_flags |= IDENT_IS_CORRELATED
		}
		ident_val := value.NewValue(ident_flags)
		allowed.SetField(b.Variable(), ident_val)
		aliases.SetField(b.Variable(), ident_val)

		if b.NameVariable() != "" {
			if ident_val, ok := allowed.Field(b.NameVariable()); ok {
				ident_flags = uint32(ident_val.ActualForIndex().(int64))
				tmp_flags1 := ident_flags & IDENT_IS_KEYSPACE
				tmp_flags2 := ident_flags &^ IDENT_IS_KEYSPACE
				if !this.indexScope() || tmp_flags1 == 0 || tmp_flags2 != 0 {
					var errContext string
					if b.Expression() != nil {
						errContext = b.Expression().ErrorContext()
					}
					err = errors.NewDuplicateVariableError(b.NameVariable(), errContext)
					return
				}
			} else {
				ident_flags = 0
			}

			ident_flags |= IDENT_IS_VARIABLE
			ident_val := value.NewValue(ident_flags)
			allowed.SetField(b.NameVariable(), ident_val)
			aliases.SetField(b.NameVariable(), ident_val)
		}
	}

	if push {
		this.allowed = allowed
		this.identifiers = identifiers
		this.aliases = aliases
	}
	return
}

/*
Restore scope to parent's scope.
*/
func (this *Formalizer) PopBindings() {

	currLevelAllowed := this.Allowed().GetValue().Fields()
	currLevelIndentfiers := this.Identifiers().GetValue().Fields()

	this.allowed = this.allowed.Parent().(*value.ScopeValue)
	this.identifiers = this.identifiers.Parent().(*value.ScopeValue)
	this.aliases = this.aliases.Parent().(*value.ScopeValue)

	// Identifiers that are used in current level but not defined in the current level scope move to parent
	for ident, ident_val := range currLevelIndentfiers {
		if currLevelAllowed != nil {
			if _, ok := currLevelAllowed[ident]; !ok {
				this.identifiers.SetField(ident, ident_val)
			}
		}
	}
}

func (this *Formalizer) Copy() *Formalizer {
	f := NewFormalizer(this.keyspace, nil)
	if len(this.withs) > 0 {
		f.withs = make(map[string]*WithInfo, len(this.withs))
		for with, v := range this.withs {
			f.withs[with] = v.Copy()
		}
	}
	f.allowed = this.allowed.Copy().(*value.ScopeValue)
	f.identifiers = this.identifiers.Copy().(*value.ScopeValue)
	f.aliases = this.aliases.Copy().(*value.ScopeValue)
	f.flags = this.flags
	return f
}

func (this *Formalizer) SetKeyspace(keyspace string) {
	this.keyspace = keyspace

	if !this.mapKeyspace() && keyspace != "" {
		this.SetAllowedAlias(keyspace, true)
	}
}

func (this *Formalizer) Keyspace() string {
	return this.keyspace
}

func (this *Formalizer) Allowed() *value.ScopeValue {
	return this.allowed
}

func (this *Formalizer) Identifiers() *value.ScopeValue {
	return this.identifiers
}

func (this *Formalizer) Aliases() *value.ScopeValue {
	return this.aliases
}

// Argument must be non-nil
func (this *Formalizer) SetIdentifiers(identifiers *value.ScopeValue) {
	this.identifiers = identifiers
}

func (this *Formalizer) SetAlias(alias string) {
	if alias != "" {
		// we treat alias for keyspace as well as equivalent such as
		// subquery term, expression term, as same to keyspace
		var ident_flags uint32 = IDENT_IS_KEYSPACE
		this.aliases.SetField(alias, value.NewValue(ident_flags))
	}
}

// alias must be non-empty
func (this *Formalizer) SetAllowedAlias(alias string, isKeyspace bool) {
	var ident_flags uint32
	if isKeyspace {
		ident_flags = IDENT_IS_KEYSPACE
	} else {
		ident_flags = IDENT_IS_UNKNOWN
	}
	this.allowed.SetField(alias, value.NewValue(ident_flags))
}

// alias must be non-empty
func (this *Formalizer) SetAllowedUnnestAlias(alias string) {
	ident_flags := uint32(IDENT_IS_KEYSPACE | IDENT_IS_UNNEST_ALIAS)
	this.allowed.SetField(alias, value.NewValue(ident_flags))
}

// alias must be non-empty
func (this *Formalizer) SetAllowedExprTermAlias(alias string) {
	ident_flags := uint32(IDENT_IS_KEYSPACE | IDENT_IS_EXPR_TERM)
	if _, ok := this.withs[alias]; ok {
		ident_flags |= uint32(IDENT_IS_WITH_ALIAS)
	}
	this.allowed.SetField(alias, value.NewValue(ident_flags))
}

// alias must be non-empty
func (this *Formalizer) SetAllowedSubqTermAlias(alias string) {
	ident_flags := uint32(IDENT_IS_KEYSPACE | IDENT_IS_SUBQ_TERM)
	this.allowed.SetField(alias, value.NewValue(ident_flags))
}

// alias must be non-empty
func (this *Formalizer) SetAllowedGroupAsAlias(alias string) {
	ident_flags := uint32(IDENT_IS_GROUP_AS)
	this.allowed.SetField(alias, value.NewValue(ident_flags))
}

func (this *Formalizer) WithAlias(alias string) bool {
	if this.withs != nil {
		_, ok := this.withs[alias]
		return ok
	}
	return false
}

func (this *Formalizer) WithInfo(alias string) *WithInfo {
	if this.withs != nil {
		if info, ok := this.withs[alias]; ok {
			return info
		}
	}
	return nil
}

func (this *Formalizer) ProcessWiths(withs Bindings) error {
	if this.withs == nil {
		this.withs = make(map[string]*WithInfo, len(withs))
	}

	for _, with := range withs {
		errContext := with.expr.ErrorContext()
		if _, ok := this.withs[with.variable]; ok {
			return errors.NewDuplicateAliasError("WITH clause", with.variable+errContext, "semantics.with.duplicate_with_alias")
		}
		if _, ok := this.allowed.Field(with.variable); ok {
			return errors.NewDuplicateAliasError("WITH clause", with.variable+errContext, "semantics.with.duplicate_with_alias")
		}

		f := NewFormalizer("", this)
		expr, err := f.Map(with.expr)
		if err != nil {
			return err
		}
		with.expr = expr

		// check for correlation
		var correlated bool
		var correlation map[string]uint32
		if f.CheckCorrelated() {
			correlated = true
			correlation = f.GetCorrelation()
		}
		this.withs[with.variable] = newWithInfo(correlated, correlation)
	}
	return nil
}

func (this *Formalizer) SetPermanentWiths(withs Bindings) {
	if this.withs == nil {
		this.withs = make(map[string]*WithInfo, len(withs))
	}
	for _, b := range withs {
		this.withs[b.Variable()] = newPermanentWithInfo()
	}
}

func (this *Formalizer) SaveWiths(isSubq bool) map[string]*WithInfo {
	withs := this.withs
	this.withs = make(map[string]*WithInfo, len(withs))
	for k, v := range withs {
		if isSubq || v.permanent {
			this.withs[k] = v.Copy()
		}
	}
	return withs
}

func (this *Formalizer) RestoreWiths(withs map[string]*WithInfo) {
	this.withs = withs
}

func (this *Formalizer) GetCorrelation() map[string]uint32 {
	return this.correlation
}

func (this *Formalizer) CheckCorrelated() bool {
	immediate := this.allowed.GetValue().Fields()

	correlated := false
	for id, id_val := range this.identifiers.Fields() {
		if _, ok := immediate[id]; !ok {
			id_flags := uint32(value.NewValue(id_val).ActualForIndex().(int64))
			if this.WithAlias(id) {
				if (id_flags & IDENT_IS_CORRELATED) == 0 {
					continue
				}
				// for a correlated WITH variable, the actual correlation would have
				// already been added to this.correlation when the WITH expression
				// is being checked (in PushBindings()), there is no need to add
				// the WITH alias itself to this.correlation
			} else {
				if this.correlation == nil {
					this.correlation = make(map[string]uint32)
				}
				this.correlation[id] |= id_flags
			}
			correlated = true
		}
	}

	return correlated
}

func (this *Formalizer) CheckCorrelation(alias string) bool {
	immediate := this.allowed.GetValue().Fields()
	_, ok := immediate[alias]
	return !ok
}

func (this *Formalizer) AddCorrelatedIdentifiers(correlation map[string]uint32) error {
	for k, _ := range correlation {
		if _, ok := this.identifiers.Field(k); !ok {
			v, ok1 := this.allowed.Field(k)
			if !ok1 {
				return errors.NewFormalizerInternalError(fmt.Sprintf("correlation reference %s is not in allowed", k))
			}
			this.identifiers.SetField(k, v)
		}
	}
	return nil
}
