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
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	clientReceiver          = "Client"
	clientWithRespReceiver  = "ClientWithResponses"
	requestConstructorStart = "New"
	requestConstructorEnd   = "Request"
	requestWithBodySuffix   = "WithBody"
)

// opSpec identifies an API operation as the generated client sees it.
type opSpec struct {
	Method     string
	Path       string
	ParamsType types.Type
	BodyType   types.Type
	// RawBody is set for client methods that take the request body as an opaque io.Reader.
	RawBody bool
}

type opIndex struct {
	methods map[string]opSpec
}

func (i *opIndex) method(name string) (opSpec, bool) {
	spec, ok := i.methods[name]
	return spec, ok
}

type constructorInfo struct {
	name      string
	httpVerb  string
	path      string
	delegates []string
}

// buildOpIndex maps every generated client method to its operation by reading the request
// constructors, so oapi-codegen naming rules are never reimplemented.
func buildOpIndex(kbapi *packages.Package) (*opIndex, error) {
	constructors := make(map[string]*constructorInfo)
	var clientMethods, responseMethods []*ast.FuncDecl
	for _, file := range kbapi.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			switch receiverName(fn) {
			case "":
				if isRequestConstructor(fn.Name.Name) {
					constructors[fn.Name.Name] = readConstructor(fn)
				}
			case clientReceiver:
				if fn.Name.IsExported() {
					clientMethods = append(clientMethods, fn)
				}
			case clientWithRespReceiver:
				if fn.Name.IsExported() {
					responseMethods = append(responseMethods, fn)
				}
			}
		}
	}

	resolved := make(map[string]opSpec, len(constructors))
	for name := range constructors {
		method, path, err := resolveConstructor(constructors, name, map[string]bool{})
		if err != nil {
			return nil, err
		}
		spec := opSpec{Method: method, Path: path}
		spec.ParamsType, spec.BodyType = constructorRequestTypes(kbapi, name)
		resolved[name] = spec
	}
	perOp := mergeConstructorTypes(resolved)

	idx := &opIndex{methods: make(map[string]opSpec)}
	for _, fn := range clientMethods {
		ctor := calledConstructor(fn, constructors)
		if ctor == "" {
			return nil, fmt.Errorf("kbapi client method %s has no request constructor", fn.Name.Name)
		}
		spec := perOp[operationKey(resolved[ctor].Method, resolved[ctor].Path)]
		spec.RawBody = strings.HasSuffix(ctor, requestWithBodySuffix)
		idx.methods[fn.Name.Name] = spec
	}
	for _, fn := range responseMethods {
		target := calledClientMethod(fn, idx)
		if target == "" {
			return nil, fmt.Errorf("kbapi client method %s does not delegate to a client method with a request constructor", fn.Name.Name)
		}
		idx.methods[fn.Name.Name] = idx.methods[target]
	}
	return idx, nil
}

func mergeConstructorTypes(resolved map[string]opSpec) map[string]opSpec {
	perOp := make(map[string]opSpec)
	for _, spec := range resolved {
		key := operationKey(spec.Method, spec.Path)
		merged := perOp[key]
		merged.Method, merged.Path = spec.Method, spec.Path
		if merged.ParamsType == nil {
			merged.ParamsType = spec.ParamsType
		}
		if merged.BodyType == nil {
			merged.BodyType = spec.BodyType
		}
		perOp[key] = merged
	}
	return perOp
}

func isRequestConstructor(name string) bool {
	return strings.HasPrefix(name, requestConstructorStart) &&
		(strings.HasSuffix(name, requestConstructorEnd) || strings.HasSuffix(name, requestConstructorEnd+requestWithBodySuffix))
}

func receiverName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	expr := fn.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return "?"
}

func readConstructor(fn *ast.FuncDecl) *constructorInfo {
	info := &constructorInfo{name: fn.Name.Name}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.SelectorExpr:
			if pkg, ok := node.X.(*ast.Ident); ok && pkg.Name == "http" {
				if verb, ok := strings.CutPrefix(node.Sel.Name, "Method"); ok {
					info.httpVerb = strings.ToUpper(verb)
				}
			}
		case *ast.AssignStmt:
			if path, ok := operationPathLiteral(node); ok {
				info.path = path
			}
		case *ast.CallExpr:
			if ident, ok := node.Fun.(*ast.Ident); ok && isRequestConstructor(ident.Name) {
				info.delegates = append(info.delegates, ident.Name)
			}
		}
		return true
	})
	return info
}

func operationPathLiteral(assign *ast.AssignStmt) (string, bool) {
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return "", false
	}
	lhs, ok := assign.Lhs[0].(*ast.Ident)
	if !ok || lhs.Name != "operationPath" {
		return "", false
	}
	call, ok := assign.Rhs[0].(*ast.CallExpr)
	if !ok || len(call.Args) == 0 {
		return "", false
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return "", false
	}
	path, err := strconv.Unquote(lit.Value)
	return path, err == nil
}

func resolveConstructor(all map[string]*constructorInfo, name string, seen map[string]bool) (string, string, error) {
	info, ok := all[name]
	if !ok {
		return "", "", fmt.Errorf("request constructor %s not found", name)
	}
	if info.httpVerb != "" && info.path != "" {
		return info.httpVerb, info.path, nil
	}
	if seen[name] {
		return "", "", fmt.Errorf("request constructor %s delegates in a cycle", name)
	}
	seen[name] = true
	for _, delegate := range info.delegates {
		if method, path, err := resolveConstructor(all, delegate, seen); err == nil {
			return method, path, nil
		}
	}
	return "", "", fmt.Errorf("request constructor %s has no literal HTTP method and path", name)
}

func constructorRequestTypes(kbapi *packages.Package, name string) (params, body types.Type) {
	fn, ok := kbapi.Types.Scope().Lookup(name).(*types.Func)
	if !ok {
		return nil, nil
	}
	sig := fn.Type().(*types.Signature)
	for v := range sig.Params().Variables() {
		t := v.Type()
		if ptr, ok := t.(*types.Pointer); ok {
			t = ptr.Elem()
		}
		alias, isAlias := t.(*types.Alias)
		var typeName string
		switch {
		case isAlias:
			typeName = alias.Obj().Name()
		default:
			if named, ok := t.(*types.Named); ok {
				typeName = named.Obj().Name()
			}
		}
		switch {
		case strings.HasSuffix(typeName, "Params"):
			params = types.Unalias(t)
		case strings.HasSuffix(typeName, "JSONRequestBody"):
			body = types.Unalias(t)
		}
	}
	return params, body
}

func calledConstructor(fn *ast.FuncDecl, constructors map[string]*constructorInfo) string {
	var found string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && constructors[ident.Name] != nil && found == "" {
				found = ident.Name
			}
		}
		return found == ""
	})
	return found
}

func calledClientMethod(fn *ast.FuncDecl, idx *opIndex) string {
	var found string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if _, known := idx.methods[sel.Sel.Name]; known && found == "" {
					found = sel.Sel.Name
				}
			}
		}
		return found == ""
	})
	return found
}
