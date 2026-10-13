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
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// providerClientsPkgSuffix identifies the provider's connection and version-check plumbing. It
// is never expanded: it returns clients rather than calling operations for a resource, and
// interface resolution would otherwise credit every resource with its version-check calls.
const providerClientsPkgSuffix = "/internal/clients"

// scanUnit is a piece of syntax reachable from a resource, with the type information to read it.
type scanUnit struct {
	pkg  *packages.Package
	node ast.Node
}

// reachResult is the over-approximated set of code a resource can execute.
type reachResult struct {
	// Methods holds the kbapi client method names referenced within the reachable code.
	Methods map[string]bool
	// Funcs holds the full names of the reachable module functions.
	Funcs map[string]bool
	units []scanUnit
	// implementations resolves an interface method to every module implementation.
	implementations func(iface *types.Interface, name string) []*types.Func
}

type declSite struct {
	pkg  *packages.Package
	node ast.Node
}

// moduleIndex holds module-wide lookups shared by every resource's reachability walk.
type moduleIndex struct {
	funcs      map[*types.Func]declSite
	vars       map[*types.Var]declSite
	namedTypes []*types.Named
	// genericTypes cannot be tested with types.Implements without an instantiation, so any
	// method of a matching name is conservatively included.
	genericTypes []*types.Named
	implCache    map[implKey][]*types.Func
}

type implKey struct {
	iface *types.Interface
	name  string
}

func (m *loadedModule) moduleIndex() *moduleIndex {
	m.indexOnce.Do(func() { m.index = buildModuleIndex(m) })
	return m.index
}

func buildModuleIndex(mod *loadedModule) *moduleIndex {
	idx := &moduleIndex{
		funcs:     make(map[*types.Func]declSite),
		vars:      make(map[*types.Var]declSite),
		implCache: make(map[implKey][]*types.Func),
	}
	for _, pkg := range mod.pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				indexDecl(idx, pkg, decl)
			}
		}
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			tn, ok := scope.Lookup(name).(*types.TypeName)
			if !ok || tn.IsAlias() {
				continue
			}
			named, ok := tn.Type().(*types.Named)
			if !ok || types.IsInterface(named) {
				continue
			}
			if named.TypeParams().Len() > 0 {
				idx.genericTypes = append(idx.genericTypes, named)
				continue
			}
			idx.namedTypes = append(idx.namedTypes, named)
		}
	}
	return idx
}

func indexDecl(idx *moduleIndex, pkg *packages.Package, decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if fn, ok := pkg.TypesInfo.Defs[d.Name].(*types.Func); ok && d.Body != nil {
			idx.funcs[fn.Origin()] = declSite{pkg: pkg, node: d}
		}
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Values) == 0 {
				continue
			}
			for _, name := range vs.Names {
				if v, ok := pkg.TypesInfo.Defs[name].(*types.Var); ok {
					idx.vars[v] = declSite{pkg: pkg, node: vs}
				}
			}
		}
	}
}

// computeReach builds the reference graph rooted at every declaration in the resource's package.
// Every reference to a module function, whether a call or a function value, is followed, and
// interface method calls are resolved to all module implementations.
func computeReach(mod *loadedModule, kbapiPath, pkgPath string) *reachResult {
	result := &reachResult{Methods: make(map[string]bool), Funcs: make(map[string]bool)}
	root, ok := mod.byPath[pkgPath]
	if !ok {
		return result
	}
	idx := mod.moduleIndex()
	w := &reachWalker{idx: idx, kbapiPath: kbapiPath, result: result, visited: make(map[ast.Node]bool)}
	result.implementations = w.implementations
	for _, file := range root.Syntax {
		w.scan(root, file)
	}
	return result
}

type reachWalker struct {
	idx       *moduleIndex
	kbapiPath string
	result    *reachResult
	visited   map[ast.Node]bool
}

func (w *reachWalker) scan(pkg *packages.Package, node ast.Node) {
	if w.visited[node] {
		return
	}
	w.visited[node] = true
	w.result.units = append(w.result.units, scanUnit{pkg: pkg, node: node})
	ast.Inspect(node, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		switch obj := pkg.TypesInfo.Uses[ident].(type) {
		case *types.Func:
			w.follow(obj)
		case *types.Var:
			if site, ok := w.idx.vars[obj.Origin()]; ok {
				w.scan(site.pkg, site.node)
			}
		case *types.TypeName:
			w.followMethods(obj)
		}
		return true
	})
}

func (w *reachWalker) follow(fn *types.Func) {
	origin := fn.Origin()
	if pkg := origin.Pkg(); pkg != nil && pkg.Path() == w.kbapiPath {
		if isClientMethod(origin) {
			w.result.Methods[origin.Name()] = true
		}
		return
	}
	if iface := receiverInterface(origin); iface != nil {
		for _, impl := range w.implementations(iface, origin.Name()) {
			w.followDecl(impl)
		}
		return
	}
	w.followDecl(origin)
}

// followMethods scans every method of a referenced module type: frameworks and envelopes call
// them through interfaces, so no identifier in the resource's own code names them.
func (w *reachWalker) followMethods(tn *types.TypeName) {
	named, ok := types.Unalias(tn.Type()).(*types.Named)
	if !ok || (tn.Pkg() != nil && tn.Pkg().Path() == w.kbapiPath) {
		return
	}
	for method := range named.Origin().Methods() {
		w.followDecl(method)
	}
}

func (w *reachWalker) followDecl(fn *types.Func) {
	site, ok := w.idx.funcs[fn.Origin()]
	if !ok || strings.HasSuffix(site.pkg.PkgPath, providerClientsPkgSuffix) {
		return
	}
	w.result.Funcs[fn.FullName()] = true
	w.scan(site.pkg, site.node)
}

func isClientMethod(fn *types.Func) bool {
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return false
	}
	recv := sig.Recv().Type()
	if ptr, ok := recv.(*types.Pointer); ok {
		recv = ptr.Elem()
	}
	named, ok := recv.(*types.Named)
	if !ok {
		return false
	}
	return slices.Contains([]string{clientReceiver, clientWithRespReceiver, "ClientInterface"}, named.Obj().Name())
}

func receiverInterface(fn *types.Func) *types.Interface {
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return nil
	}
	switch t := sig.Recv().Type().Underlying().(type) {
	case *types.Interface:
		return t
	default:
		return nil
	}
}

func (w *reachWalker) implementations(iface *types.Interface, name string) []*types.Func {
	key := implKey{iface: iface, name: name}
	if cached, ok := w.idx.implCache[key]; ok {
		return cached
	}
	var impls []*types.Func
	for _, named := range w.idx.namedTypes {
		for _, t := range []types.Type{named, types.NewPointer(named)} {
			if !types.Implements(t, iface) {
				continue
			}
			obj, _, _ := types.LookupFieldOrMethod(t, true, named.Obj().Pkg(), name)
			if fn, ok := obj.(*types.Func); ok {
				impls = append(impls, fn)
			}
			break
		}
	}
	// Generic types are matched by method name only, so this intentionally over-includes:
	// usage is over-counted, which can only hide gaps, never invent them.
	for _, generic := range w.idx.genericTypes {
		for method := range generic.Methods() {
			if method.Name() == name {
				impls = append(impls, method)
			}
		}
	}
	w.idx.implCache[key] = impls
	return impls
}
