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
	"slices"
	"strings"
)

// Force-all prefixes. When any changed file path matches one of these,
// the tool selects the full acceptance test package set.
var forceAllPrefixes = []string{
	"provider/",
	"internal/acctest/",
	"internal/clients/",
	"internal/entitycore/",
	"generated/",
	"xpprovider/",
	".github/workflows/",
	// The embedded example tree: consumed only from test code
	// (internal/acctest/examples_plan_test.go embeds examples/ and drives
	// TestAccExamples_planOnly over every embedded example), so examples/ is
	// outside the enumeration roots and outside the go list pattern that
	// builds the import graph, so phase 1 has no reverse-dependency edges
	// into it, and is excluded from the unit-test job by -skip '^TestAcc'.
	// An examples-only PR would otherwise select zero packages on every shard.
	"examples/",
	// The tool's own package: a PR touching only scripts/targeted-testacc/
	// would otherwise select zero acceptance packages, so changes to the
	// selection tool are always exercised by the full acceptance suite.
	"scripts/targeted-testacc/",
	// Shared acceptance-test helper packages: imported only from test files,
	// so the phase-1 reverse-dependency walk (non-test imports) cannot see
	// them and they do not declare Terraform entities for phase 2.
	"internal/kibana/dashboard/dashboardacctest/",
	"internal/kibana/dashboard/panelkit/contracttest/",
	"internal/providerfwtest/",
}

// Force-all files. When any changed file path equals one of these, the tool
// selects the full acceptance test package set. These are module-level files
// that affect every build, so a diff touching them cannot be narrowed to a
// subset of packages.
var forceAllFiles = []string{
	"go.mod",
	"go.sum",
	"Makefile",
	"main.go",
	".terraform-version",
	".env.template",
}

// isForceAllDockerComposeFile reports whether the file name matches the
// docker-compose compose-file glob (docker-compose*.yml or docker-compose*.yaml,
// at any repository path) that triggers a force-all run.
func isForceAllDockerComposeFile(file string) bool {
	name := filepath.Base(file)
	if !strings.HasPrefix(name, "docker-compose") {
		return false
	}
	return strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")
}

// Classifier maps changed file paths to Go package import paths and detects
// force-all prefixes. Go files belong to their own directory; every other
// file is attributed to its nearest ancestor directory that contains a .go
// file (or ignored when no such directory exists).
type Classifier struct {
	ModulePath string
}

// NewClassifier creates a classifier for the given module import path.
func NewClassifier(modulePath string) *Classifier {
	return &Classifier{ModulePath: modulePath}
}

// ClassifyResult holds the output of classifying the changed file set.
type ClassifyResult struct {
	// ForceAll is true when at least one changed file triggers a full run.
	ForceAll bool
	// Packages is the deduplicated set of changed Go package import paths.
	Packages []string
	// HasCode is true when at least one changed file maps to a Go package: a
	// .go file, or any other file whose nearest ancestor directory with a
	// .go file exists (testdata fixtures, go:embed'ed description files,
	// fixture directories not named exactly testdata).
	HasCode bool
}

// Classify maps changed file paths to their owning Go package import paths.
// Files outside Go packages (no ancestor directory contains a .go file) are
// ignored. Every other file is attributed to its nearest ancestor directory
// that contains a .go file: testdata fixtures, go:embed'ed schema
// description files, and fixture directories that are not named exactly
// testdata all belong to the package that embeds or reads them.
func (c *Classifier) Classify(changedFiles []string) *ClassifyResult {
	res := &ClassifyResult{}

	seen := make(map[string]struct{})
	for _, file := range changedFiles {
		file = filepath.ToSlash(file)

		if matchesForceAll(file) {
			res.ForceAll = true
		}

		pkgDir, ok := c.packageDir(file)
		if !ok {
			continue
		}

		res.HasCode = true

		// A changed .go path whose directory no longer exists means the diff
		// deletes (part of) a package. The selection cannot reason about
		// deleted code, so fail safe in the conservative direction: run the
		// full suite.
		if _, err := os.Stat(pkgDir); err != nil {
			res.ForceAll = true
			continue
		}

		importPath := c.ModulePath + "/" + pkgDir
		if _, exists := seen[importPath]; exists {
			continue
		}
		seen[importPath] = struct{}{}
		res.Packages = append(res.Packages, importPath)
	}

	return res
}

// packageDir returns the directory path (relative to the module root) that
// owns the changed file, and whether such a directory exists.
func (c *Classifier) packageDir(file string) (string, bool) {
	if strings.HasSuffix(file, ".go") {
		// A .go file belongs to its own directory even when the diff has
		// already deleted that directory: Classify's fail-safe then triggers a
		// full run instead of attributing the deleted file to an ancestor.
		return filepath.ToSlash(filepath.Dir(file)), true
	}

	// Any other changed file is attributed to its nearest ancestor directory
	// that contains a .go file. This covers files under testdata/ as well as
	// go:embed'ed content (e.g. descriptions/*.md schema inputs) and fixture
	// directories that are not named exactly testdata (e.g. test_data/*.tf).
	return owningPackageDir(file)
}

// hasGoFile reports whether dir contains at least one .go file.
func hasGoFile(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".go") {
			return true
		}
	}
	return false
}

func matchesForceAll(file string) bool {
	for _, prefix := range forceAllPrefixes {
		if strings.HasPrefix(file, prefix) {
			return true
		}
	}
	if slices.Contains(forceAllFiles, file) {
		return true
	}
	return isForceAllDockerComposeFile(file)
}
