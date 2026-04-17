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
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type kibanaoapiIndex struct {
	// funcName -> base file name (e.g. alerting_rule.go)
	funcFile map[string]string
	// fileName -> file content (pre-loaded for O(1) symbol lookups)
	fileContent map[string]string
}

func buildKibanaOAPIIndex(repoRoot string) (*kibanaoapiIndex, error) {
	dir := filepath.Join(repoRoot, "internal/clients/kibanaoapi")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	idx := &kibanaoapiIndex{
		funcFile:    make(map[string]string),
		fileContent: make(map[string]string),
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		idx.fileContent[e.Name()] = string(data)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, data, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || !fn.Name.IsExported() || fn.Recv != nil {
				continue
			}
			idx.funcFile[fn.Name.Name] = e.Name()
		}
	}
	return idx, nil
}

func kibanaOAPICallsFromPaths(paths []string) (map[string]struct{}, error) {
	calls := make(map[string]struct{})
	for _, p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			err := filepath.WalkDir(p, func(sub string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if !strings.HasSuffix(sub, ".go") {
					return nil
				}
				return scanFileForKibanaOAPICalls(sub, calls)
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		if err := scanFileForKibanaOAPICalls(p, calls); err != nil {
			return nil, err
		}
	}
	return calls, nil
}

func scanFileForKibanaOAPICalls(path string, calls map[string]struct{}) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || id.Name != "kibanaoapi" {
			return true
		}
		if sel.Sel != nil && sel.Sel.Name != "" {
			calls[sel.Sel.Name] = struct{}{}
		}
		return true
	})
	return nil
}

func containsGoSymbol(src, symbol string) bool {
	if symbol == "" {
		return false
	}
	// Word-boundary match for generated identifiers (types, client methods).
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(symbol) + `\b`)
	return re.MatchString(src)
}

// matchHighConfidence returns symbols from changed that are referenced by this entity through
// its implementation paths and/or kibanaoapi files whose sources reference those symbols when
// the entity calls exported helpers from those files.
func matchHighConfidence(scanPaths []string, oapi *kibanaoapiIndex, changed []string) (matched []string, err error) {
	if len(changed) == 0 || len(scanPaths) == 0 {
		return nil, nil
	}

	calls, err := kibanaOAPICallsFromPaths(scanPaths)
	if err != nil {
		return nil, err
	}

	var srcBuf strings.Builder
	for _, p := range scanPaths {
		fi, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			err := filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || !strings.HasSuffix(path, ".go") {
					return nil
				}
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				srcBuf.Write(b)
				srcBuf.WriteByte('\n')
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		srcBuf.Write(b)
		srcBuf.WriteByte('\n')
	}
	entitySrc := srcBuf.String()

	for _, sym := range changed {
		if containsGoSymbol(entitySrc, sym) {
			matched = append(matched, sym)
			continue
		}
		// Use pre-loaded index to avoid O(symbols × files) disk reads.
		for fname, content := range oapi.fileContent {
			if !containsGoSymbol(content, sym) {
				continue
			}
			for fn := range calls {
				if oapi.funcFile[fn] == fname {
					matched = append(matched, sym)
					goto nextSym
				}
			}
		}
	nextSym:
	}

	return sortStrings(matched), nil
}

func sortStrings(s []string) []string {
	sort.Strings(s)
	return s
}
