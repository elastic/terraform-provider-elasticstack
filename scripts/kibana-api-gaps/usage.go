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
	"go/types"
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
// were populated somewhere this analysis cannot see.
func (u *usageSet) markConversion(info *types.Info, call *ast.CallExpr) {
	if tv, ok := info.Types[call.Fun]; ok && tv.IsType() {
		u.markOpaqueType(tv.Type)
	}
}

// markWholeValueCopy records struct assignments such as `*out = *in` or `dst = src`, which copy
// every field from a value built elsewhere. Literals and calls are not copies: the fields set
// by the code that builds them are counted through its own selections and literals.
func (u *usageSet) markWholeValueCopy(info *types.Info, assign *ast.AssignStmt) {
	if len(assign.Lhs) != len(assign.Rhs) {
		return
	}
	for i, rhs := range assign.Rhs {
		switch ast.Unparen(rhs).(type) {
		case *ast.StarExpr, *ast.Ident, *ast.SelectorExpr:
		default:
			continue
		}
		t := info.TypeOf(assign.Lhs[i])
		if t == nil {
			continue
		}
		if _, isStruct := derefType(t).Underlying().(*types.Struct); isStruct && !isPointer(t) {
			u.markOpaqueType(t)
		}
	}
}

func isPointer(t types.Type) bool {
	_, ok := types.Unalias(t).(*types.Pointer)
	return ok
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
