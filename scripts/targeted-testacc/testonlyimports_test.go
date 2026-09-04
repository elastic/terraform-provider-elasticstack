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
	"strings"
	"testing"
)

// TestTestOnlyImportedPackagesGuard fails when a package under internal/ is
// imported only from test files (no non-test importer exists in the module)
// and is neither covered by a force-all prefix nor entity-declaring. Such
// packages are invisible to both selection phases: phase 1 walks non-test
// imports only, and phase 2 only covers packages that declare Terraform
// entities. A change to such a package would silently skip relevant
// acceptance tests, so when a new shared test-helper package is introduced
// it must either be added to forceAllPrefixes or declare entities.
func TestTestOnlyImportedPackagesGuard(t *testing.T) {
	modulePath, err := currentModulePath()
	if err != nil {
		t.Skipf("cannot resolve module path: %v", err)
	}
	moduleInternal := modulePath + "/internal/"

	nonTestImported := map[string]bool{}
	importsOut, err := exec.Command("go", "list", "-f",
		"{{.ImportPath}} {{join .Imports \" \"}}", "./internal/...").Output()
	if err != nil {
		t.Skipf("go list failed: %v", err)
	}
	for _, fields := range scanGoList(string(importsOut)) {
		for _, imp := range fields[1:] {
			if strings.HasPrefix(imp, moduleInternal) {
				nonTestImported[imp] = true
			}
		}
	}

	testImportsOut, err := exec.Command("go", "list", "-f",
		"{{.ImportPath}} {{join .TestImports \" \"}} {{join .XTestImports \" \"}}", "./internal/...").Output()
	if err != nil {
		t.Skipf("go list failed: %v", err)
	}
	testImported := map[string]bool{}
	for _, fields := range scanGoList(string(testImportsOut)) {
		for _, imp := range fields[1:] {
			if imp != fields[0] && strings.HasPrefix(imp, moduleInternal) {
				testImported[imp] = true
			}
		}
	}

	var unguarded []string
	for pkg := range testImported {
		if nonTestImported[pkg] {
			continue
		}
		dir := strings.TrimPrefix(pkg, modulePath+"/")
		if matchesForceAll(dir) {
			continue
		}
		entities, err := ExtractEntities(dir)
		if err != nil {
			t.Errorf("extract entities for %s: %v", dir, err)
			continue
		}
		if len(entities) == 0 {
			unguarded = append(unguarded, dir)
		}
	}

	if len(unguarded) > 0 {
		t.Errorf("packages under internal/ are imported only from test files but are neither force-all nor entity-declaring: %v\n"+
			"Phase 1 uses non-test imports only, so these packages are invisible to selection. Add them to forceAllPrefixes or ensure they declare Terraform entities.", unguarded)
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
