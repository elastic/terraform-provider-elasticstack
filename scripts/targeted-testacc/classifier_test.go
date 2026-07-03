package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
	return path
}

func TestClassifier_Classify_MapsGoFileToPackage(t *testing.T) {
	root := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	writeFile(t, root, "internal/kibana/slo/resource.go", "package slo")

	c := NewClassifier("github.com/example/mod")
	res, err := c.Classify([]string{"internal/kibana/slo/resource.go"})
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}

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
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	writeFile(t, root, "internal/kibana/slo/testdata/main.tf", "resource {}\n")
	writeFile(t, root, "internal/kibana/slo/resource.go", "package slo")

	c := NewClassifier("github.com/example/mod")
	res, err := c.Classify([]string{"internal/kibana/slo/testdata/main.tf"})
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}

	want := []string{"github.com/example/mod/internal/kibana/slo"}
	if !reflect.DeepEqual(res.Packages, want) {
		t.Errorf("packages = %v, want %v", res.Packages, want)
	}
}

func TestClassifier_Classify_IgnoresNonRelevantFiles(t *testing.T) {
	root := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	writeFile(t, root, "docs/index.md", "# docs\n")

	c := NewClassifier("github.com/example/mod")
	res, err := c.Classify([]string{"docs/index.md", "README.md"})
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}

	if len(res.Packages) != 0 {
		t.Errorf("packages = %v, want empty", res.Packages)
	}
	if res.HasCode {
		t.Errorf("HasCode = true, want false")
	}
}

func TestClassifier_Classify_DeduplicatesPackages(t *testing.T) {
	root := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	writeFile(t, root, "internal/a/resource.go", "package a")

	c := NewClassifier("github.com/example/mod")
	res, err := c.Classify([]string{"internal/a/resource.go", "internal/a/resource_test.go"})
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}

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
			res, err := c.Classify([]string{file})
			if err != nil {
				t.Fatalf("Classify: %v", err)
			}
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
			res, err := c.Classify([]string{file})
			if err != nil {
				t.Fatalf("Classify: %v", err)
			}
			if res.ForceAll {
				t.Errorf("ForceAll = true for %s, want false", file)
			}
		})
	}
}

func TestPackageDir_TestdataNestedUnderSubdirectory(t *testing.T) {
	root := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

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
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWd)
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	writeFile(t, root, "internal/z/main.go", "package z")
	writeFile(t, root, "internal/a/main.go", "package a")
	writeFile(t, root, "internal/m/main.go", "package m")

	c := NewClassifier("github.com/example/mod")
	res, err := c.Classify([]string{"internal/z/main.go", "internal/a/main.go", "internal/m/main.go"})
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}

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
