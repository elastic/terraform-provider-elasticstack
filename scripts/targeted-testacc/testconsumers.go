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
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindTestConsumers searches root (typically "internal") recursively for files
// whose contents contain entityName. It considers *.tf and *_test.go files.
// Each matching file is mapped to its owning Go package import path (the nearest
// ancestor directory containing a .go file). The deduplicated set of import
// paths is returned.
func FindTestConsumers(root, modulePath, entityName string) ([]string, error) {
	return FindTestConsumersMulti(root, modulePath, []string{entityName})
}

// FindTestConsumersMulti walks root once and reports the deduplicated import
// paths of packages containing files that mention any of entityNames. The walk
// cost is independent of the number of entity names: each candidate file is
// read once and checked against every name.
func FindTestConsumersMulti(root, modulePath string, entityNames []string) ([]string, error) {
	seen := make(map[string]struct{})

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if !strings.HasSuffix(name, ".tf") && !strings.HasSuffix(name, "_test.go") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("scan %s: %w", path, err)
		}

		if !fileContainsAny(data, entityNames) {
			return nil
		}

		pkgDir, ok := owningPackageDir(path)
		if !ok {
			return nil
		}
		importPath := modulePath + "/" + pkgDir
		seen[importPath] = struct{}{}
		return nil
	}

	if err := filepath.WalkDir(root, walkFn); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(seen))
	for pkg := range seen {
		result = append(result, pkg)
	}
	return stringsSorted(result), nil
}

// owningPackageDir returns the nearest ancestor directory of path that
// contains at least one .go file.
func owningPackageDir(path string) (string, bool) {
	dir := filepath.Dir(path)
	for {
		if dir == "." || dir == "/" || dir == "" {
			return "", false
		}
		if hasGoFile(dir) {
			return filepath.ToSlash(dir), true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// fileContainsAny reports whether data contains any of needles.
func fileContainsAny(data []byte, needles []string) bool {
	for _, n := range needles {
		if bytes.Contains(data, []byte(n)) {
			return true
		}
	}
	return false
}
