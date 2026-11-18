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
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Member is one unused request field inside a gap group.
type Member struct {
	Field   string `json:"field"`
	Pointer string `json:"pointer"`
	XState  string `json:"x_state,omitempty"`
}

// SkippedShape records a schema subtree the walk treated as covered because it cannot follow it.
type SkippedShape struct {
	Resource string `json:"resource"`
	Pointer  string `json:"pointer"`
	Shape    string `json:"shape"`
}

// propGap is the unused fields found under one top-level request field.
type propGap struct {
	name     string
	ref      schemaRef
	required bool
	members  []Member
	// query marks a query parameter, as opposed to a request body property.
	query bool
}

type walker struct {
	doc      *oasDoc
	usage    *usageSet
	defined  definedTypes
	resource string
	skipped  []SkippedShape
	active   map[string]bool
}

func newWalker(doc *oasDoc, usage *usageSet, defined definedTypes, resource string) *walker {
	return &walker{doc: doc, usage: usage, defined: defined, resource: resource, active: make(map[string]bool)}
}

// isOpaque also checks the type a defined type was declared from, since JSON decoding may
// target either.
func (w *walker) isOpaque(t types.Type) bool {
	if w.usage.opaque(t) {
		return true
	}
	if named, ok := t.(*types.Named); ok {
		if source, ok := w.defined[named.Obj()]; ok {
			return w.usage.opaque(source)
		}
	}
	return false
}

func (w *walker) skip(pointer, shape string) {
	w.skipped = append(w.skipped, SkippedShape{Resource: w.resource, Pointer: pointer, Shape: shape})
}

// bodyGaps walks a request body schema against its Go type in lockstep and returns one
// entry per top-level request field.
func (w *walker) bodyGaps(schema schemaRef, goType types.Type) ([]propGap, error) {
	v, err := w.doc.view(schema)
	if err != nil {
		return nil, err
	}
	if v.unsupported != "" {
		w.skip(schema.pointer, v.unsupported)
		return nil, nil
	}
	t := derefType(goType)
	if w.isOpaque(t) {
		return nil, nil
	}
	if variants, ok := w.defined.unionVariants(t); ok {
		return unionBranches(w, v, variants, schema.pointer, w.bodyGaps)
	}
	if _, ok := t.Underlying().(*types.Struct); !ok || len(v.properties) == 0 {
		w.skip(schema.pointer, "request body is not an object with properties")
		return nil, nil
	}
	return w.structGaps(v, t, "")
}

// unionBranches walks each oneOf/anyOf branch against the matching union variant and joins the
// results.
func unionBranches[T any](w *walker, v *schemaView, variants []types.Type, pointer string, walk func(schemaRef, types.Type) ([]T, error)) ([]T, error) {
	if len(v.oneOf) == 0 {
		w.skip(pointer, "union Go type without oneOf or anyOf")
		return nil, nil
	}
	if len(v.oneOf) != len(variants) {
		return nil, fmt.Errorf("%s: %d oneOf/anyOf branches but %d union variants in Go type", pointer, len(v.oneOf), len(variants))
	}
	var out []T
	for i, branch := range v.oneOf {
		found, err := walk(branch, variants[i])
		if err != nil {
			return nil, err
		}
		out = append(out, found...)
	}
	return out, nil
}

// queryGaps matches query parameters to the fields of the operation's Params struct.
func (w *walker) queryGaps(params []oasParam, paramsType types.Type, opName string) ([]propGap, error) {
	if len(params) == 0 {
		return nil, nil
	}
	if paramsType == nil {
		return nil, fmt.Errorf("%s has query parameters but its client method takes no params struct", opName)
	}
	fields := jsonFields(paramsType)
	var out []propGap
	for _, param := range params {
		if param.deprecated {
			continue
		}
		entry, ok := fields[param.name]
		if !ok {
			return nil, fmt.Errorf("unmatched json tag: query parameter %q of %s has no field in Go type %s", param.name, opName, paramsType)
		}
		gap := propGap{name: param.name, ref: param.ref, required: boolValue(param.ref.node, "required"), query: true}
		if !w.usage.fieldUsed(entry.owner, entry.field.Name()) {
			gap.members = []Member{{Field: param.name, Pointer: param.ref.pointer, XState: stringValue(param.ref.node, "x-state")}}
		} else if schema := mappingValue(param.ref.node, "schema"); schema != nil {
			members, err := w.gaps(schemaRef{node: schema, pointer: param.ref.pointer + "/schema"}, entry.field.Type(), param.name)
			if err != nil {
				return nil, err
			}
			gap.members = members
		}
		out = append(out, gap)
	}
	return out, nil
}

// gaps returns the unused fields beneath a schema node whose own Go field is in use.
func (w *walker) gaps(ref schemaRef, goType types.Type, path string) ([]Member, error) {
	v, err := w.doc.view(ref)
	if err != nil {
		return nil, err
	}
	return w.viewGaps(ref, v, goType, path)
}

// viewGaps is gaps for a ref whose view the caller has already built.
func (w *walker) viewGaps(ref schemaRef, v *schemaView, goType types.Type, path string) ([]Member, error) {
	if v.deprecated || v.readOnly {
		return nil, nil
	}
	if v.unsupported != "" {
		w.skip(ref.pointer, v.unsupported)
		return nil, nil
	}
	t := derefType(goType)
	if w.isOpaque(t) {
		return nil, nil
	}
	key := v.pointer + "|" + types.TypeString(t, nil)
	if w.active[key] {
		return nil, nil
	}
	w.active[key] = true
	defer delete(w.active, key)

	switch u := t.Underlying().(type) {
	case *types.Slice:
		return w.elementGaps(v, u.Elem(), path)
	case *types.Array:
		return w.elementGaps(v, u.Elem(), path)
	case *types.Struct:
		return w.objectGaps(v, t, ref.pointer, path)
	case *types.Map:
		w.skip(ref.pointer, "map-valued field: dynamic keys cannot be proven unused")
		return nil, nil
	case *types.Interface:
		w.skip(ref.pointer, "untyped value: contents cannot be proven unused")
		return nil, nil
	default:
		return nil, nil
	}
}

func (w *walker) elementGaps(v *schemaView, elem types.Type, path string) ([]Member, error) {
	if v.items == nil {
		return nil, nil
	}
	return w.gaps(*v.items, elem, path+"[]")
}

func (w *walker) objectGaps(v *schemaView, t types.Type, pointer, path string) ([]Member, error) {
	if variants, ok := w.defined.unionVariants(t); ok {
		return unionBranches(w, v, variants, pointer, func(branch schemaRef, variant types.Type) ([]Member, error) {
			return w.gaps(branch, variant, path)
		})
	}
	if len(v.oneOf) > 0 {
		w.skip(pointer, "oneOf or anyOf without a union Go type")
		return nil, nil
	}
	props, err := w.structGaps(v, t, path)
	if err != nil {
		return nil, err
	}
	var members []Member
	for _, prop := range props {
		members = append(members, prop.members...)
	}
	return members, nil
}

// structGaps matches the properties of an object schema to the JSON fields of a Go struct.
// A property without a field, or a field without a property, is a tool error.
func (w *walker) structGaps(v *schemaView, t types.Type, path string) ([]propGap, error) {
	fields := jsonFields(t)
	matched := make(map[string]bool, len(v.properties))
	var out []propGap
	for _, prop := range v.properties {
		entry, ok := fields[prop.name]
		if !ok {
			return nil, fmt.Errorf("unmatched json tag: OAS property %q at %s has no field in Go type %s", prop.name, prop.ref.pointer, t)
		}
		matched[prop.name] = true
		child, err := w.doc.view(prop.ref)
		if err != nil {
			return nil, err
		}
		if child.deprecated || child.readOnly {
			continue
		}
		if child.unsupported != "" {
			w.skip(prop.ref.pointer, child.unsupported)
			continue
		}
		fieldPath := joinFieldPath(path, prop.name)
		gap := propGap{name: prop.name, ref: prop.ref, required: prop.required}
		if w.usage.fieldUsed(entry.owner, entry.field.Name()) {
			gap.members, err = w.viewGaps(prop.ref, child, entry.field.Type(), fieldPath)
			if err != nil {
				return nil, err
			}
		} else {
			gap.members = []Member{{Field: fieldPath, Pointer: prop.ref.pointer, XState: child.xState}}
		}
		out = append(out, gap)
	}
	if err := unmatchedGoFields(fields, matched, t); err != nil {
		return nil, err
	}
	return out, nil
}

func unmatchedGoFields(fields map[string]jsonField, matched map[string]bool, t types.Type) error {
	var extra []string
	for name := range fields {
		if !matched[name] {
			extra = append(extra, name)
		}
	}
	if len(extra) == 0 {
		return nil
	}
	sort.Strings(extra)
	return fmt.Errorf("unmatched json tag: Go type %s has fields %s with no OAS property", t, strings.Join(extra, ", "))
}

func joinFieldPath(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}

// definedTypes maps a defined type declared as `type A B` to B. Methods are not inherited by
// A, so the union methods of oapi-codegen request bodies live on B.
type definedTypes map[*types.TypeName]*types.Named

func buildDefinedTypes(pkg *packages.Package) definedTypes {
	defs := make(definedTypes)
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Assign.IsValid() {
					continue
				}
				rhs, ok := pkg.TypesInfo.TypeOf(ts.Type).(*types.Named)
				if !ok {
					continue
				}
				if obj, ok := pkg.TypesInfo.Defs[ts.Name].(*types.TypeName); ok {
					defs[obj] = rhs
				}
			}
		}
	}
	return defs
}

// unionVariants returns the variant types of an oapi-codegen union, in declaration order of
// their As methods, which follows the order of the oneOf/anyOf branches.
func (d definedTypes) unionVariants(t types.Type) ([]types.Type, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return nil, false
	}
	if source, ok := d[named.Obj()]; ok && named.NumMethods() == 0 {
		named = source
	}
	st, ok := named.Underlying().(*types.Struct)
	if !ok || !hasUnionField(st) {
		return nil, false
	}
	var methods []*types.Func
	for method := range named.Methods() {
		if strings.HasPrefix(method.Name(), "As") && method.Exported() {
			methods = append(methods, method)
		}
	}
	sort.Slice(methods, func(i, j int) bool { return methods[i].Pos() < methods[j].Pos() })
	variants := make([]types.Type, 0, len(methods))
	for _, method := range methods {
		sig := method.Type().(*types.Signature)
		if sig.Results().Len() == 0 {
			continue
		}
		variants = append(variants, sig.Results().At(0).Type())
	}
	return variants, true
}

func hasUnionField(st *types.Struct) bool {
	return st.NumFields() == 1 && st.Field(0).Name() == "union"
}
