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

// FindTestConsumersMulti walks root once and reports, per owning package, the
// deduplicated set of entity names found in that package's *.tf and *_test.go
// files. The walk cost is independent of the number of entity names: each
// candidate file is read once and checked against every name.
func FindTestConsumersMulti(root, modulePath string, entityNames []string) (map[string][]string, error) {
	seen := make(map[string]map[string]struct{})

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

		matched := matchedNames(data, entityNames)
		if len(matched) == 0 {
			return nil
		}

		pkgDir, ok := owningPackageDir(path)
		if !ok {
			return nil
		}
		importPath := modulePath + "/" + pkgDir
		if seen[importPath] == nil {
			seen[importPath] = make(map[string]struct{})
		}
		for _, name := range matched {
			seen[importPath][name] = struct{}{}
		}
		return nil
	}

	if err := filepath.WalkDir(root, walkFn); err != nil {
		return nil, err
	}

	result := make(map[string][]string, len(seen))
	for pkg, names := range seen {
		result[pkg] = stringsSorted(mapKeys(names))
	}
	return result, nil
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

// matchedNames returns the subset of needles contained in data.
func matchedNames(data []byte, needles []string) []string {
	matched := make([]string, 0)
	for _, n := range needles {
		if bytes.Contains(data, []byte(n)) {
			matched = append(matched, n)
		}
	}
	return matched
}

// mapKeys returns the sorted keys of a set map.
func mapKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	return keys
}
