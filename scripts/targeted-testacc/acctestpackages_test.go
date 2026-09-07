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
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestFindAccTestPackages_SyntheticTree(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "internal/kibana/space/resource_test.go", "package space\n\nfunc TestAccSpace_basic(t *testing.T) {}\n")
	writeFile(t, root, "internal/kibana/space/resource_unit_test.go", "package space\n\nfunc TestUnitSpace(t *testing.T) {}\n")
	writeFile(t, root, "internal/kibana/dashboard/dashboard_test.go", "package dashboard\n\nfunc TestAccDashboard(t *testing.T) {}\n")
	writeFile(t, root, "internal/kibana/dashboard/panel/lens/lens_test.go", "package lens\n\nfunc TestAccLensPanel(t *testing.T) {}\n")
	writeFile(t, root, "internal/fleet/policy/resource.go", "package policy\n")
	writeFile(t, root, "internal/fleet/policy/resource_test.go", "package policy\n\nfunc TestAccPolicy(t *testing.T) {}\n")
	// Acceptance suite without the TestAcc prefix, driven via resource.Test.
	writeFile(t, root, "internal/kibana/synthetics/monitor/acc_test.go",
		"package monitor_test\n\nimport \"github.com/hashicorp/terraform-plugin-testing/helper/resource\"\n\nfunc TestSyntheticMonitor(t *testing.T) { resource.Test(t, resource.TestCase{}) }\n")
	// No _test.go file at all.
	writeFile(t, root, "internal/pkg/resource.go", "package pkg\n")

	got, err := FindAccTestPackages([]string{"internal"}, "github.com/example/mod")
	if err != nil {
		t.Fatalf("FindAccTestPackages: %v", err)
	}

	want := []string{
		"github.com/example/mod/internal/fleet/policy",
		"github.com/example/mod/internal/kibana/dashboard",
		"github.com/example/mod/internal/kibana/dashboard/panel/lens",
		"github.com/example/mod/internal/kibana/space",
		"github.com/example/mod/internal/kibana/synthetics/monitor",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindAccTestPackages = %v, want %v", got, want)
	}
}

func TestFindAccTestPackages_NoPackages(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "foo.go", "package foo\n")

	got, err := FindAccTestPackages([]string{"."}, "github.com/example/mod")
	if err != nil {
		t.Fatalf("FindAccTestPackages: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no packages, got %v", got)
	}
}

func TestIsAccTestFile(t *testing.T) {
	root := t.TempDir()

	cases := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "with-acc-test",
			content: "package foo\n\nfunc TestAccSomething(t *testing.T) {}\n",
			want:    true,
		},
		{
			name:    "unit-test-only",
			content: "package foo\n\nfunc TestSomething(t *testing.T) {}\n",
			want:    false,
		},
		{
			name:    "lowercase-acc",
			content: "package foo\n\nfunc TestaccSomething(t *testing.T) {}\n",
			want:    false,
		},
		{
			name:    "acc-in-comment",
			content: "package foo\n\n// func TestAccSomething(t *testing.T) {}\n",
			want:    false,
		},
		{
			name:    "harness-resource-test",
			content: "package foo\n\nimport \"github.com/hashicorp/terraform-plugin-testing/helper/resource\"\n\nfunc TestSomething(t *testing.T) {\n\tresource.Test(t, resource.TestCase{})\n}\n",
			want:    true,
		},
		{
			name:    "harness-resource-paralleltest",
			content: "package foo\n\nimport \"github.com/hashicorp/terraform-plugin-testing/helper/resource\"\n\nfunc TestSomething(t *testing.T) {\n\tresource.ParallelTest(t, resource.TestCase{})\n}\n",
			want:    true,
		},
		{
			name:    "harness-aliased-import",
			content: "package foo\n\nimport r \"github.com/hashicorp/terraform-plugin-testing/helper/resource\"\n\nfunc TestSomething(t *testing.T) {\n\tr.Test(t, r.TestCase{})\n}\n",
			want:    true,
		},
		{
			name:    "other-package-paralleltest",
			content: "package foo\n\nimport \"sync\"\n\nvar l sync.Mutex\n\nfunc other() { l.ParallelTest() }\n",
			want:    false,
		},
		{
			name:    "harness-call-in-comment",
			content: "package foo\n\nimport \"github.com/hashicorp/terraform-plugin-testing/helper/resource\"\n\n// resource.Test(t, resource.TestCase{})\n",
			want:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(root, tc.name+".go")
			if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
				t.Fatalf("write file: %v", err)
			}
			got, err := isAccTestFile(path)
			if err != nil {
				t.Fatalf("isAccTestFile: %v", err)
			}
			if got != tc.want {
				t.Errorf("isAccTestFile = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestAcceptancePackageEnumerationGuard fails when any package in the repository
//
// Named TestAcceptance... rather than TestAcc... so the stackless unit-test
// job (`go test ./... -skip '^TestAcc'`) still runs it — the `^TestAcc`
// filter would otherwise exclude the one guard whose purpose is to catch
// acceptance suites that silently stop running before they go green.
// (outside vendor/, .git/, and testdata/ fixtures) that declares a func
// TestAcc or invokes the acceptance harness (resource.Test /
// resource.ParallelTest) is missing from FindAccTestPackages output.
// Acceptance suites that do not follow the TestAcc naming convention (e.g.
// internal/kibana/synthetics, provider/provider_test.go) are only selected
// when the enumeration recognizes harness invocations; a regression here
// makes those suites silently unreachable — a PR changing them would select
// zero packages and go green. The detection below is deliberately text-based
// so it does not share the AST analysis it guards against.
func TestAcceptancePackageEnumerationGuard(t *testing.T) {
	root := repoRoot(t)
	modulePath, err := currentModulePath()
	if err != nil {
		t.Fatalf("cannot resolve module path: %v", err)
	}
	// FindAccTestPackages walks root-relative paths; run from the module root
	// so the root arguments resolve.
	t.Chdir(root)

	got, err := FindAccTestPackages(accTestEnumerationRoots, modulePath)
	if err != nil {
		t.Fatalf("FindAccTestPackages: %v", err)
	}
	gotSet := make(map[string]struct{}, len(got))
	for _, pkg := range got {
		gotSet[pkg] = struct{}{}
	}

	var missing []string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if d.IsDir() {
			// Skip vendored code, git metadata, top-level testdata dirs (analysis/
			// acctestconfigdirlint/testdata/src/... contains analyzer fixtures
			// that use resource.Test and are not real suites), and the tool's own
			// package (its unit tests textually reference resource.Test but are
			// not acceptance suites).
			if rel == "vendor" || rel == ".git" || rel == "testdata" || rel == "scripts" {
				return fs.SkipDir
			}
			return nil
		}
		if strings.Contains(filepath.ToSlash(rel), "/testdata/") || !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if !textContentHasAcceptanceTest(string(data)) {
			return nil
		}
		pkg := modulePath + "/" + filepath.ToSlash(filepath.Dir(rel))
		if _, ok := gotSet[pkg]; !ok {
			missing = append(missing, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repository: %v", err)
	}

	if len(missing) > 0 {
		t.Errorf("packages in the repository declare acceptance tests but are missing from FindAccTestPackages: %v\n"+
			"The enumeration must recognize both func TestAcc declarations and resource.Test/resource.ParallelTest invocations, "+
			"or these acceptance suites are unreachable through every selection path.", missing)
	}
}

// textContentHasAcceptanceTest reports whether source text (a *_test.go
// file) declares a func TestAcc or invokes resource.Test/resource.ParallelTest.
// Comment lines are ignored.
func textContentHasAcceptanceTest(content string) bool {
	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		if strings.HasPrefix(trimmed, "func TestAcc") {
			return true
		}
		if strings.Contains(trimmed, "resource.Test(") || strings.Contains(trimmed, "resource.ParallelTest(") {
			return true
		}
	}
	return false
}
