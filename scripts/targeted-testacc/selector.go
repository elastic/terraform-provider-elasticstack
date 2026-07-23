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
	"math"
	"sort"
)

// SelectPackages combines phase 1 and phase 2 results, applies the run-all
// threshold, and returns the final sorted package list.
//
// If forceAll is true, the full accTestPackages set is returned. Otherwise,
// phase1 and phase2 are unioned and deduplicated. If the union size exceeds
// runAllThresholdPct percent of len(accTestPackages), the full set is returned.
func SelectPackages(forceAll bool, phase1, phase2, accTestPackages []string, runAllThresholdPct float64) []string {
	all := make([]string, len(accTestPackages))
	copy(all, accTestPackages)
	sort.Strings(all)
	all = uniqStrings(all)

	if forceAll {
		return all
	}

	unionSet := make(map[string]struct{})
	for _, p := range phase1 {
		unionSet[p] = struct{}{}
	}
	for _, p := range phase2 {
		unionSet[p] = struct{}{}
	}

	union := make([]string, 0, len(unionSet))
	for p := range unionSet {
		union = append(union, p)
	}
	sort.Strings(union)

	thresholdCount := int(math.Floor(runAllThresholdPct / 100.0 * float64(len(all))))
	if len(union) > thresholdCount {
		return all
	}
	return union
}

// ApplyShard returns the slice of packages assigned to the given shard.
//
// Rules:
//   - If shardIndex >= totalShards: empty.
//   - If len(packages) < minShardPackages and totalShards > 1:
//     shardIndex == 0 returns all packages; shardIndex > 0 returns empty.
//   - Otherwise, round-robin: emit packages whose position%totalShards == shardIndex.
func ApplyShard(packages []string, totalShards, shardIndex, minShardPackages int) []string {
	if totalShards <= 0 {
		return nil
	}
	if shardIndex >= totalShards {
		return nil
	}

	if len(packages) < minShardPackages && totalShards > 1 {
		if shardIndex == 0 {
			return packages
		}
		return nil
	}

	var out []string
	for i, pkg := range packages {
		if i%totalShards == shardIndex {
			out = append(out, pkg)
		}
	}
	return out
}
