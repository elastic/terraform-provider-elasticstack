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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

func ReadArtifact(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read artifact: %w", err)
	}
	var versions []string
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, fmt.Errorf("parse artifact: %w", err)
	}
	return versions, nil
}

func WriteArtifact(path string, versions []string) error {
	sorted := SortVersions(versions)
	data, err := json.MarshalIndent(sorted, "", "  ")
	if err != nil {
		return fmt.Errorf("encode artifact: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create artifact directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write artifact: %w", err)
	}
	return nil
}

func VersionsEqual(a, b []string) bool {
	return slices.Equal(SortVersions(a), SortVersions(b))
}

func SortVersions(versions []string) []string {
	out := append([]string(nil), versions...)
	sort.SliceStable(out, func(i, j int) bool {
		iSnap := strings.HasSuffix(out[i], "-SNAPSHOT")
		jSnap := strings.HasSuffix(out[j], "-SNAPSHOT")
		if iSnap != jSnap {
			return !iSnap
		}
		return versionLess(out[i], out[j])
	})
	return out
}
