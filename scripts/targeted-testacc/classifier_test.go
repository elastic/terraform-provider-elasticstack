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
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func TestClassifier_Classify_MapsGoFileToPackage(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "internal/kibana/slo/resource.go", "package slo")

	c := NewClassifier("github.com/example/mod")
	res := c.Classify([]string{"internal/kibana/slo/resource.go"})

	want := []string{"github.com/example/mod/internal/kibana/slo"}
	if !reflect.DeepEqual(res.Packages, want) {
		t.Errorf("packages = %v, want %v", res.Packages, want)
	}
	if !res.HasCode {
		t.Errorf("HasCode = false, want true")
	}
	if res.ForceAll {
		t.Errorf("ForceAll = true, want false")
	}
}

func TestClassifier_Classify_MapsTestdataFileToAncestorPackage(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "internal/kibana/slo/testdata/main.tf", "resource {}\n")
	writeFile(t, root, "internal/kibana/slo/resource.go", "package slo")

	c := NewClassifier("github.com/example/mod")
	res := c.Classify([]string{"internal/kibana/slo/testdata/main.tf"})

	want := []string{"github.com/example/mod/internal/kibana/slo"}
	if !reflect.DeepEqual(res.Packages, want) {
		t.Errorf("packages = %v, want %v", res.Packages, want)
	}
}

func TestClassifier_Classify_IgnoresNonRelevantFiles(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "docs/index.md", "# docs\n")

	c := NewClassifier("github.com/example/mod")
	res := c.Classify([]string{"docs/index.md", "README.md"})

	if len(res.Packages) != 0 {
		t.Errorf("packages = %v, want empty", res.Packages)
	}
	if res.HasCode {
		t.Errorf("HasCode = true, want false")
	}
}

func TestClassifier_Classify_DeduplicatesPackages(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "internal/a/resource.go", "package a")

	c := NewClassifier("github.com/example/mod")
	res := c.Classify([]string{"internal/a/resource.go", "internal/a/resource_test.go"})

	want := []string{"github.com/example/mod/internal/a"}
	if !reflect.DeepEqual(res.Packages, want) {
		t.Errorf("packages = %v, want %v", res.Packages, want)
	}
}

func TestClassifier_Classify_ForceAllPrefixes(t *testing.T) {
	prefixes := []string{
		"provider/config.go",
		"internal/acctest/provider_factory.go",
		"internal/clients/clients.go",
		"internal/entitycore/resource.go",
		"generated/kibana/client.go",
	}

	for _, file := range prefixes {
		t.Run(file, func(t *testing.T) {
			c := NewClassifier("github.com/example/mod")
			res := c.Classify([]string{file})
			if !res.ForceAll {
				t.Errorf("ForceAll = false for %s, want true", file)
			}
		})
	}
}

func TestClassifier_Classify_NoForceAllForSimilarPaths(t *testing.T) {
	files := []string{
		"internal/clientspkg/client.go",
		"providerx/config.go",
		"internal/entitycorepkg/base.go",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			c := NewClassifier("github.com/example/mod")
			res := c.Classify([]string{file})
			if res.ForceAll {
				t.Errorf("ForceAll = true for %s, want false", file)
			}
		})
	}
}

func TestPackageDir_TestdataNestedUnderSubdirectory(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "internal/pkg/testdata/sub/main.tf", "")
	writeFile(t, root, "internal/pkg/resource.go", "package pkg")

	c := NewClassifier("github.com/example/mod")
	dir, ok := c.packageDir("internal/pkg/testdata/sub/main.tf")
	if !ok {
		t.Fatalf("packageDir returned false, want true")
	}
	want := "internal/pkg"
	if dir != want {
		t.Errorf("packageDir = %q, want %q", dir, want)
	}
}

func TestClassifyResult_PackagesAreSorted(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "internal/z/main.go", "package z")
	writeFile(t, root, "internal/a/main.go", "package a")
	writeFile(t, root, "internal/m/main.go", "package m")

	c := NewClassifier("github.com/example/mod")
	res := c.Classify([]string{"internal/z/main.go", "internal/a/main.go", "internal/m/main.go"})

	want := []string{
		"github.com/example/mod/internal/a",
		"github.com/example/mod/internal/m",
		"github.com/example/mod/internal/z",
	}
	sort.Strings(res.Packages)
	if !reflect.DeepEqual(res.Packages, want) {
		t.Errorf("packages = %v, want %v", res.Packages, want)
	}
}
