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
	"strings"
)

// FindAccTestPackages walks root (typically "internal") and returns the import
// paths of all Go packages that define at least one func TestAcc in a
// *_test.go file.
func FindAccTestPackages(root, modulePath string) ([]string, error) {
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

	if err := filepath.WalkDir(root, walkFn); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(seen))
	for dir := range seen {
		result = append(result, modulePath+"/"+filepath.ToSlash(dir))
	}
	return stringsSorted(result), nil
}

// isAccTestFile reports whether path declares at least one func TestAcc.
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
	return false, nil
}
