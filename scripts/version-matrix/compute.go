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
	"slices"
	"strings"
)

func ComputeDesired(tags []string, snapshot string, pinned []string, probe func(string) (bool, error)) ([]string, error) {
	computed := LatestPatchPerMinor(tags)
	if len(computed) == 0 {
		return nil, fmt.Errorf("no 8.x/9.x GA tags found")
	}
	out := make([]string, 0, len(computed)+1)
	promotionBlocked := false
	for _, version := range computed {
		if slices.Contains(pinned, version) {
			out = append(out, version)
			continue
		}
		if probe == nil {
			out = append(out, version)
			continue
		}
		ok, err := probe(version)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, version)
			continue
		}
		if prev := pinnedGAForMinor(pinned, version); prev != "" {
			out = append(out, prev)
			continue
		}
		if snap := pinnedSnapshotForMinor(pinned, version); snap != "" {
			out = append(out, snap)
			promotionBlocked = true
		}
	}
	if snapshot != "" && !promotionBlocked {
		out = append(out, snapshot)
	}
	return SortVersions(out), nil
}

func pinnedSnapshotForMinor(pinned []string, version string) string {
	wantMajor, wantMinor, _ := parseLooseVersion(version)
	for _, p := range pinned {
		if !strings.HasSuffix(p, "-SNAPSHOT") {
			continue
		}
		major, minor, _ := parseLooseVersion(p)
		if major == wantMajor && minor == wantMinor {
			return p
		}
	}
	return ""
}

func pinnedGAForMinor(pinned []string, version string) string {
	wantMajor, wantMinor, _ := parseLooseVersion(version)
	for _, p := range pinned {
		if strings.HasSuffix(p, "-SNAPSHOT") {
			continue
		}
		major, minor, _ := parseLooseVersion(p)
		if major == wantMajor && minor == wantMinor {
			return p
		}
	}
	return ""
}
