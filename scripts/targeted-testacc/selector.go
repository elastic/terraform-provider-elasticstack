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

	thresholdCount := RunAllThresholdCount(runAllThresholdPct, len(all))
	if len(union) > thresholdCount {
		return all
	}
	return union
}

// RunAllThresholdCount returns the run-all threshold package count
// for a total of total acceptance packages: the largest count that still
// selects the narrow union. Exported so the dry-run threshold note in
// main.go uses the identical rule instead of re-deriving it (two copies
// of the same rule can drift).
func RunAllThresholdCount(runAllThresholdPct float64, total int) int {
	return int(math.Floor(runAllThresholdPct / 100.0 * float64(total)))
}

// ShardPlan is the machine-readable shard plan emitted to stdout: the complete
// sorted selection, the ordered shard assignments, whether any package was
// selected, and why the selection was made.
type ShardPlan struct {
	HasPackages      bool       `json:"has_packages"`
	SelectedPackages []string   `json:"selected_packages"`
	Shards           [][]string `json:"shards"`
	Rationale        []string   `json:"rationale"`
}

// BuildShardPlan computes the complete shard plan for the selected package set
// under the given sharding policy.
//
//   - No packages: exactly one empty shard.
//   - len(packages) < minShardPackages and totalShards > 1: exactly one shard
//     holding every package.
//   - Otherwise: min(totalShards, len(packages)) shards, assigned round-robin
//     from the sorted package list.
func BuildShardPlan(packages []string, totalShards, minShardPackages int) *ShardPlan {
	sorted := stringsSorted(packages)

	plan := &ShardPlan{SelectedPackages: sorted}
	if len(sorted) == 0 {
		plan.Shards = [][]string{{}}
		return plan
	}
	plan.HasPackages = true

	if totalShards > 1 && len(sorted) < minShardPackages {
		plan.Shards = [][]string{sorted}
		return plan
	}

	shardCount := min(totalShards, len(sorted))
	shards := make([][]string, shardCount)
	for i, pkg := range sorted {
		shard := i % shardCount
		if shards[shard] == nil {
			shards[shard] = []string{}
		}
		shards[shard] = append(shards[shard], pkg)
	}
	plan.Shards = shards
	return plan
}

// ValidateShardPlan enforces the plan contract: zero selected packages yields
// exactly one empty assignment; a nonempty selection is covered exactly once
// by nonempty shards.
func ValidateShardPlan(plan *ShardPlan) error {
	if len(plan.SelectedPackages) == 0 {
		if len(plan.Shards) != 1 {
			return fmt.Errorf("empty plan must contain exactly one shard, got %d", len(plan.Shards))
		}
		if len(plan.Shards[0]) != 0 {
			return fmt.Errorf("empty plan shard must be empty, got %d packages", len(plan.Shards[0]))
		}
		if plan.HasPackages {
			return fmt.Errorf("empty plan must set has_packages=false")
		}
		return nil
	}

	if !plan.HasPackages {
		return fmt.Errorf("nonempty plan must set has_packages=true")
	}

	seen := make(map[string]int, len(plan.SelectedPackages))
	for shardIdx, shard := range plan.Shards {
		if len(shard) == 0 {
			return fmt.Errorf("nonempty plan shard %d is empty", shardIdx)
		}
		for _, pkg := range shard {
			if prev, dup := seen[pkg]; dup {
				return fmt.Errorf("package %s appears in shards %d and %d", pkg, prev, shardIdx)
			}
			seen[pkg] = shardIdx
		}
	}
	for _, pkg := range plan.SelectedPackages {
		if _, ok := seen[pkg]; !ok {
			return fmt.Errorf("package %s is not assigned to any shard", pkg)
		}
	}
	return nil
}
