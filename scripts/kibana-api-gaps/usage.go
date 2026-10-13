// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
)

// usageSet records which struct fields a resource's reachable code reads or writes, and which
// types it handles as opaque whole values through encoding/json.
type usageSet struct {
	named       map[*types.TypeName]map[string]bool
	anon        map[string][]*types.Struct
	opaqueNamed map[*types.TypeName]bool
}

func collectUsage(reach *reachResult) *usageSet {
	u := &usageSet{
		named:       make(map[*types.TypeName]map[string]bool),
		anon:        make(map[string][]*types.Struct),
		opaqueNamed: make(map[*types.TypeName]bool),
	}
	reflecting := reflectingFuncs(reach)
	for _, unit := range reach.units {
		info := unit.pkg.TypesInfo
		ast.Inspect(unit.node, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.SelectorExpr:
				u.markSelection(info, node)
			case *ast.CompositeLit:
				u.markCompositeLit(info, node)
			case *ast.CallExpr:
				u.markConversion(info, node)
				u.markJSONCall(info, node)
				u.markReflectCall(info, node)
				u.markReflectingCallSite(info, node, reflecting)
			case *ast.AssignStmt:
				u.markWholeValueCopy(info, node)
			}
			return true
		})
	}
	return u
}

func (u *usageSet) markSelection(info *types.Info, sel *ast.SelectorExpr) {
	selection := info.Selections[sel]
	if selection == nil || selection.Kind() != types.FieldVal {
		return
	}
	current := selection.Recv()
	for _, index := range selection.Index() {
		owner := derefType(current)
		st, ok := owner.Underlying().(*types.Struct)
		if !ok {
			return
		}
		field := st.Field(index)
		u.markField(owner, field.Name())
		current = field.Type()
	}
}

func (u *usageSet) markCompositeLit(info *types.Info, lit *ast.CompositeLit) {
	litType := info.TypeOf(lit)
	if litType == nil {
		return
	}
	owner := derefType(litType)
	st, ok := owner.Underlying().(*types.Struct)
	if !ok {
		return
	}
	for i, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if key, ok := kv.Key.(*ast.Ident); ok {
				u.markField(owner, key.Name)
			}
			continue
		}
		if i < st.NumFields() {
			u.markField(owner, st.Field(i).Name())
		}
	}
}

func (u *usageSet) markField(owner types.Type, goName string) {
	if named, ok := owner.(*types.Named); ok {
		obj := named.Origin().Obj()
		if u.named[obj] == nil {
			u.named[obj] = make(map[string]bool)
		}
		u.named[obj][goName] = true
		return
	}
	if st, ok := owner.Underlying().(*types.Struct); ok {
		u.anon[goName] = append(u.anon[goName], st)
	}
}

func (u *usageSet) fieldUsed(owner types.Type, goName string) bool {
	owner = derefType(owner)
	if named, ok := owner.(*types.Named); ok {
		return u.named[named.Origin().Obj()][goName]
	}
	st, ok := owner.Underlying().(*types.Struct)
	if !ok {
		return false
	}
	for _, candidate := range u.anon[goName] {
		if types.Identical(candidate, st) {
			return true
		}
	}
	return false
}

func (u *usageSet) opaque(t types.Type) bool {
	named, ok := derefType(t).(*types.Named)
	return ok && u.opaqueNamed[named.Origin().Obj()]
}

// markConversion records a conversion into a type as a whole-value construction: the fields
// were populated somewhere this analysis cannot see. Converting a composite literal is not
// one, because the literal's own keys are already counted.
func (u *usageSet) markConversion(info *types.Info, call *ast.CallExpr) {
	tv, ok := info.Types[call.Fun]
	if !ok || !tv.IsType() || len(call.Args) != 1 || isCompositeLiteral(call.Args[0]) {
		return
	}
	u.markOpaqueType(tv.Type)
}

func isCompositeLiteral(expr ast.Expr) bool {
	expr = ast.Unparen(expr)
	if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		expr = ast.Unparen(unary.X)
	}
	_, ok := expr.(*ast.CompositeLit)
	return ok
}

// markWholeValueCopy records dereferencing copies such as `*out = *in`, which overwrite every
// field of a value with one built elsewhere. Plain variable and selector copies are not
// treated as opaque: they would hide real gaps whenever any local struct is copied.
func (u *usageSet) markWholeValueCopy(info *types.Info, assign *ast.AssignStmt) {
	if len(assign.Lhs) != len(assign.Rhs) {
		return
	}
	for i, lhs := range assign.Lhs {
		if _, deref := ast.Unparen(lhs).(*ast.StarExpr); !deref {
			continue
		}
		if _, deref := ast.Unparen(assign.Rhs[i]).(*ast.StarExpr); !deref {
			continue
		}
		if t := info.TypeOf(lhs); t != nil {
			u.markOpaqueType(t)
		}
	}
}

// markJSONCall records the targets of whole-value encoding/json decodes as opaque types.
// Marshal and Encode are deliberately ignored: serialising a value says nothing about who
// populated its fields, so counting them would hide real gaps.
func (u *usageSet) markJSONCall(info *types.Info, call *ast.CallExpr) {
	argIndex, ok := jsonValueArgument(info, call)
	if !ok || argIndex >= len(call.Args) {
		return
	}
	u.markOpaqueType(info.TypeOf(call.Args[argIndex]))
}

func jsonValueArgument(info *types.Info, call *ast.CallExpr) (int, bool) {
	var ident *ast.Ident
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		ident = fun.Sel
	case *ast.Ident:
		ident = fun
	default:
		return 0, false
	}
	fn, ok := info.Uses[ident].(*types.Func)
	if !ok || fn.Pkg() == nil || fn.Pkg().Path() != "encoding/json" {
		return 0, false
	}
	switch fn.Name() {
	case "Unmarshal":
		return 1, true
	case "Decode":
		return 0, true
	default:
		return 0, false
	}
}

func (u *usageSet) markOpaqueType(t types.Type) {
	if t == nil {
		return
	}
	t = derefType(t)
	switch typed := t.(type) {
	case *types.Named:
		u.opaqueNamed[typed.Origin().Obj()] = true
	case *types.Slice:
		u.markOpaqueType(typed.Elem())
	case *types.Array:
		u.markOpaqueType(typed.Elem())
	case *types.Map:
		u.markOpaqueType(typed.Elem())
	}
}

// reflectingFuncs finds the reachable functions that hand a value of non-concrete static type
// (a type parameter or an interface) to the reflect package. Which struct they inspect is only
// known at their call sites.
func reflectingFuncs(reach *reachResult) map[*types.Func]bool {
	found := make(map[*types.Func]bool)
	for _, unit := range reach.units {
		decl, ok := unit.node.(*ast.FuncDecl)
		if !ok {
			continue
		}
		info := unit.pkg.TypesInfo
		fn, ok := info.Defs[decl.Name].(*types.Func)
		if !ok {
			continue
		}
		ast.Inspect(decl, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok && isReflectCall(info, call) && slices.ContainsFunc(call.Args, func(arg ast.Expr) bool { return !concreteType(info.TypeOf(arg)) }) {
				found[fn.Origin()] = true
			}
			return true
		})
	}
	return found
}

func concreteType(t types.Type) bool {
	if t == nil {
		return false
	}
	switch derefType(t).(type) {
	case *types.TypeParam:
		return false
	}
	return !types.IsInterface(derefType(t))
}

func isReflectCall(info *types.Info, call *ast.CallExpr) bool {
	fn := calledFunc(info, call)
	return fn != nil && fn.Pkg() != nil && fn.Pkg().Path() == "reflect" && fn.Type().(*types.Signature).Recv() == nil
}

func calledFunc(info *types.Info, call *ast.CallExpr) *types.Func {
	ident := funcIdent(call)
	if ident == nil {
		return nil
	}
	fn, _ := info.Uses[ident].(*types.Func)
	return fn
}

// markReflectCall treats a concrete value handed to the reflect package as an opaque whole:
// reflection can read and write any of its fields by name.
func (u *usageSet) markReflectCall(info *types.Info, call *ast.CallExpr) {
	if !isReflectCall(info, call) {
		return
	}
	for _, arg := range call.Args {
		if t := info.TypeOf(arg); concreteType(t) {
			u.markOpaqueType(t)
		}
	}
}

// markReflectingCallSite marks every argument type and type argument of a call to a reflecting
// function as opaque, which over-approximates what the callee inspects.
func (u *usageSet) markReflectingCallSite(info *types.Info, call *ast.CallExpr, reflecting map[*types.Func]bool) {
	fn := calledFunc(info, call)
	if fn == nil || !reflecting[fn.Origin()] {
		return
	}
	for _, arg := range call.Args {
		u.markOpaqueType(info.TypeOf(arg))
	}
	if ident := funcIdent(call); ident != nil {
		if inst, ok := info.Instances[ident]; ok {
			for i := range inst.TypeArgs.Len() {
				u.markOpaqueType(inst.TypeArgs.At(i))
			}
		}
	}
}

func funcIdent(call *ast.CallExpr) *ast.Ident {
	fun := ast.Unparen(call.Fun)
	switch f := fun.(type) {
	case *ast.IndexExpr:
		fun = f.X
	case *ast.IndexListExpr:
		fun = f.X
	}
	switch f := fun.(type) {
	case *ast.Ident:
		return f
	case *ast.SelectorExpr:
		return f.Sel
	}
	return nil
}
