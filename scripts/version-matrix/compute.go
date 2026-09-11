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
	"slices"
	"strings"
)

func ComputeDesired(tags []string, snapshot string, pinned []string, probe func(string) bool) []string {
	computed := LatestPatchPerMinor(tags)
	out := make([]string, 0, len(computed)+1)
	promotionBlocked := false
	for _, version := range computed {
		if slices.Contains(pinned, version) || probe == nil || probe(version) {
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
	return SortVersions(out)
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
