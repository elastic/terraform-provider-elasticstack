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

package acctestconfigdirlint

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// isValidEmbeddedCompatConfig reports whether expr is a reference to a package-level
// string variable populated by //go:embed from testdata/.../main.tf (repository convention
// for ExternalProviders compatibility steps).
func isValidEmbeddedCompatConfig(pass *analysis.Pass, expr ast.Expr) bool {
	expr = unwrapParenExpr(expr)
	id, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	v, ok := pass.TypesInfo.Uses[id].(*types.Var)
	if !ok {
		return false
	}
	if v.Pkg() != pass.Pkg {
		return false
	}
	if v.Parent() != pass.Pkg.Scope() {
		return false
	}
	if !isStringKind(v.Type()) {
		return false
	}
	gd, ok := findGenDeclForVar(pass, v)
	if !ok || gd.Tok != token.VAR {
		return false
	}
	pos := pass.Fset.Position(gd.TokPos)
	paths := goEmbedPathsAboveLine(pass, pos.Filename, pos.Line)
	for _, p := range paths {
		if isTestdataMainTFEmbedPath(p) {
			return true
		}
	}
	return false
}

func unwrapParenExpr(expr ast.Expr) ast.Expr {
	for {
		p, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = p.X
	}
}

func isStringKind(typ types.Type) bool {
	if typ == nil {
		return false
	}
	b, ok := typ.Underlying().(*types.Basic)
	return ok && b.Kind() == types.String
}

func findGenDeclForVar(pass *analysis.Pass, v *types.Var) (*ast.GenDecl, bool) {
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					if name == nil {
						continue
					}
					if pass.TypesInfo.Defs[name] == v {
						return gd, true
					}
				}
			}
		}
	}
	return nil, false
}

// goEmbedPathsAboveLine collects //go:embed path tokens from consecutive directive lines
// immediately above the line containing the `var` keyword (1-based line number).
func goEmbedPathsAboveLine(pass *analysis.Pass, filename string, varKeywordLine1Based int) []string {
	if pass.ReadFile == nil {
		return nil
	}
	content, err := pass.ReadFile(filename)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(content), "\n")
	var paths []string
	for i := varKeywordLine1Based - 2; i >= 0; i-- {
		if i >= len(lines) {
			continue
		}
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		const prefix = "//go:embed"
		if strings.HasPrefix(line, prefix) {
			rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			for _, p := range strings.Fields(rest) {
				p = strings.Trim(p, "`\"")
				paths = append(paths, p)
			}
			continue
		}
		break
	}
	for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
		paths[i], paths[j] = paths[j], paths[i]
	}
	return paths
}

func isTestdataMainTFEmbedPath(path string) bool {
	if path == "testdata/main.tf" {
		return true
	}
	return strings.HasPrefix(path, "testdata/") && strings.HasSuffix(path, "/main.tf")
}
