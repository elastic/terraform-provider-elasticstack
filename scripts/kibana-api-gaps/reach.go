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

	"golang.org/x/tools/go/packages"
)

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
	implCache  map[implKey][]*types.Func
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
			if !ok || named.TypeParams().Len() > 0 || types.IsInterface(named) {
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

func (w *reachWalker) followDecl(fn *types.Func) {
	site, ok := w.idx.funcs[fn.Origin()]
	if !ok {
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
	w.idx.implCache[key] = impls
	return impls
}
