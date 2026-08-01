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
}

// Classifier maps changed file paths to Go package import paths and detects
// force-all prefixes. It also filters out non-Go/non-testdata files.
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
	// HasCode is true when at least one changed file is a Go file or a testdata
	// file that maps to a Go package.
	HasCode bool
}

// Classify maps changed file paths to their owning Go package import paths.
// Files outside Go packages are ignored. Files under testdata/ are attributed
// to the nearest ancestor directory that contains a .go file.
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
	if !isRelevantFile(file) {
		return "", false
	}

	dir := filepath.Dir(file)

	if strings.HasSuffix(file, ".go") {
		// A .go file belongs to its own directory.
		return dir, true
	}

	// testdata file: walk up to the nearest directory containing a .go file.
	for {
		if dir == "." || dir == "/" || dir == "" {
			return "", false
		}
		if hasGoFile(dir) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// isRelevantFile reports whether a changed file can contribute to package
// selection. We consider .go files and any file under a testdata/ directory.
func isRelevantFile(file string) bool {
	if strings.HasSuffix(file, ".go") {
		return true
	}
	if strings.Contains(file, "/testdata/") {
		return true
	}
	return false
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
	return false
}
