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
	"bufio"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// repoRoot returns the module root directory for the repository containing
// this test, via `go env GOMOD`.
func repoRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatalf("go env GOMOD failed: %v", err)
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		t.Fatalf("go env GOMOD returned empty module file path")
	}
	return filepath.Dir(root)
}

// goList runs `go list` with -f format over pattern from the repository root,
// failing the test on any infrastructure error (a skip here would let the
// guard silently pass).
func goList(t *testing.T, repoRoot, format, pattern string) string {
	t.Helper()
	cmd := exec.Command("go", "list", "-f", format, pattern)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list %s failed: %v", pattern, err)
	}
	return string(out)
}

// TestImportGraphExcludesTestOnlyImports keeps Phase 1 limited to production
// dependencies. Shared acceptance-test helpers are handled by the force-all
// prefix table rather than a transitive test-import graph.
func TestImportGraphExcludesTestOnlyImports(t *testing.T) {
	root := repoRoot(t)
	modulePath, err := currentModulePath()
	if err != nil {
		t.Fatalf("cannot resolve module path: %v", err)
	}
	modulePrefix := modulePath + "/"

	t.Chdir(root)
	g, err := BuildImportGraph()
	if err != nil {
		t.Fatalf("BuildImportGraph: %v", err)
	}

	productionOut := goList(t, root, "{{.ImportPath}} {{join .Imports \" \"}}", "./internal/... ./provider/...")
	productionImports := make(map[string]map[string]struct{})
	for _, fields := range scanGoList(productionOut) {
		imports := make(map[string]struct{}, len(fields)-1)
		for _, imp := range fields[1:] {
			imports[imp] = struct{}{}
		}
		productionImports[fields[0]] = imports
	}

	out := goList(t, root, "{{.ImportPath}} {{join .TestImports \" \"}} {{join .XTestImports \" \"}}", "./internal/... ./provider/...")
	var leaked []string
	for _, fields := range scanGoList(out) {
		pkg := fields[0]
		for _, imp := range fields[1:] {
			if imp == pkg || !strings.HasPrefix(imp, modulePrefix) {
				continue
			}
			if _, productionImport := productionImports[pkg][imp]; productionImport {
				continue
			}
			if slices.Contains(g.Forward[pkg], imp) {
				leaked = append(leaked, pkg+" -> "+imp)
			}
		}
	}
	if len(leaked) > 0 {
		t.Errorf("import graph includes test-only edges: %v", leaked)
	}

	aliasPackage := modulePath + "/internal/elasticsearch/index/alias"
	acctestPackage := modulePath + "/internal/acctest"
	if slices.Contains(g.Forward[aliasPackage], acctestPackage) {
		t.Errorf("import graph includes the test-only edge %s -> %s", aliasPackage, acctestPackage)
	}
}

// scanGoList splits each non-empty line of go list output into fields.
func scanGoList(out string) [][]string {
	var results [][]string
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		fields := strings.Fields(strings.TrimSpace(sc.Text()))
		if len(fields) > 0 {
			results = append(results, fields)
		}
	}
	return results
}
