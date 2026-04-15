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
// string variable populated by //go:embed from a Terraform file under testdata/
// (repository convention for ExternalProviders compatibility steps).
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
	vs, ok := findValueSpecForVar(pass, v)
	if !ok || vs == nil || len(vs.Names) == 0 || vs.Names[0] == nil {
		return false
	}
	pos := pass.Fset.Position(vs.Names[0].Pos())
	paths := goEmbedPathsAboveValueSpec(pass, pos.Filename, pos.Line)
	for _, p := range paths {
		if isTestdataTFEmbedPath(p) {
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

func findValueSpecForVar(pass *analysis.Pass, v *types.Var) (*ast.ValueSpec, bool) {
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
						return vs, true
					}
				}
			}
		}
	}
	return nil, false
}

// goEmbedPathsAboveValueSpec collects //go:embed path tokens from directive lines above the
// ValueSpec (anchored at the first bound identifier's line). Ordinary // line comments between
// //go:embed and the declaration are skipped, matching allowed go:embed layouts. For
// parenthesized `var (` blocks, lines containing only `(`, `)`, or `var` / `var (` are skipped
// so a //go:embed placed directly above an inner declaration is still found.
func goEmbedPathsAboveValueSpec(pass *analysis.Pass, filename string, valueSpecNameLine1Based int) []string {
	if pass.ReadFile == nil {
		return nil
	}
	content, err := pass.ReadFile(filename)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(content), "\n")
	var paths []string
	for i := valueSpecNameLine1Based - 2; i >= 0; i-- {
		if i >= len(lines) {
			continue
		}
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if isVarGroupBoundaryLine(line) {
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
		// Other // line comments may appear between //go:embed and the variable (valid in Go).
		if strings.HasPrefix(line, "//") {
			continue
		}
		break
	}
	for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
		paths[i], paths[j] = paths[j], paths[i]
	}
	return paths
}

func isVarGroupBoundaryLine(line string) bool {
	s := strings.TrimSpace(line)
	switch s {
	case "(", ")", "var":
		return true
	}
	if strings.HasPrefix(s, "var") {
		rest := strings.TrimSpace(strings.TrimPrefix(s, "var"))
		if rest == "(" || rest == "()" {
			return true
		}
		if len(rest) > 0 && rest[0] == '(' {
			after := strings.TrimSpace(rest[1:])
			if after == "" || after == ")" {
				return true
			}
		}
	}
	return false
}

// isTestdataTFEmbedPath reports whether path matches the repository contract for
// compatibility-step fixtures: under testdata/, end with .tf, and contain no ".",
// "..", or empty path segments that could escape the fixture tree after normalization.
func isTestdataTFEmbedPath(path string) bool {
	if path == "" || strings.Contains(path, "\\") {
		return false
	}
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "../") {
		return false
	}
	if !strings.HasPrefix(path, "testdata/") || !strings.HasSuffix(path, ".tf") {
		return false
	}
	for _, seg := range strings.Split(path, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return false
		}
	}
	return true
}
