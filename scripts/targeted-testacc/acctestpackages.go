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
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// accTestEnumerationRoots is the set of repository roots whose packages are
// enumerated as acceptance test packages and grepped by phase 2. It includes
// provider/ because provider/provider_test.go and provider/factory_test.go
// drive real live-stack acceptance suites via resource.Test without any
// TestAcc-prefixed function, with fixtures under provider/testdata/**.
var accTestEnumerationRoots = []string{"internal", "provider"}

// FindAccTestPackages walks roots and returns the import paths of all Go
// acceptance test packages: packages with at least one *_test.go file that
// declares a func TestAcc or invokes the plugin-testing acceptance harness
// (resource.Test / resource.ParallelTest).
func FindAccTestPackages(roots []string, modulePath string) ([]string, error) {
	seen := make(map[string]struct{})

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if !strings.HasSuffix(name, "_test.go") {
			return nil
		}

		ok, err := isAccTestFile(path)
		if err != nil {
			return fmt.Errorf("scan %s: %w", path, err)
		}
		if !ok {
			return nil
		}

		dir := filepath.Dir(path)
		if _, exists := seen[dir]; exists {
			return nil
		}
		seen[dir] = struct{}{}
		return nil
	}

	for _, root := range roots {
		if err := filepath.WalkDir(root, walkFn); err != nil {
			return nil, err
		}
	}

	result := make([]string, 0, len(seen))
	for dir := range seen {
		result = append(result, modulePath+"/"+filepath.ToSlash(dir))
	}
	return stringsSorted(result), nil
}

// isAccTestFile reports whether path contains acceptance tests: at least
// one func TestAcc, or an invocation of the plugin-testing acceptance
// harness (resource.Test / resource.ParallelTest). Some acceptance suites
// predate the TestAcc naming convention (e.g. internal/kibana/synthetics)
// and drive resource.TestCase purely via resource.Test, so the harness
// invocation is the authoritative signal alongside the name prefix.
func isAccTestFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, data, parser.AllErrors)
	if err != nil {
		// Files that cannot be parsed cannot declare valid acceptance tests.
		return false, nil
	}

	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if ok && strings.HasPrefix(fn.Name.Name, "TestAcc") {
			return true, nil
		}
	}
	return invokesAccHarness(f), nil
}

// invokesAccHarness reports whether f calls resource.Test or
// resource.ParallelTest, including through an aliased import of
// helper/resource.
func invokesAccHarness(f *ast.File) bool {
	idents := make(map[string]struct{})
	for _, imp := range f.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil || path != "github.com/hashicorp/terraform-plugin-testing/helper/resource" {
			continue
		}
		name := "resource"
		if imp.Name != nil {
			name = imp.Name.Name
		}
		if name != "." && name != "_" {
			idents[name] = struct{}{}
		}
	}
	if len(idents) == 0 {
		return false
	}

	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name != "Test" && sel.Sel.Name != "ParallelTest" {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if _, ok := idents[ident.Name]; ok {
			found = true
			return false
		}
		return true
	})
	return found
}
