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
	"reflect"
	"sort"
	"testing"
)

func TestSelectPackages_ForceAll(t *testing.T) {
	all := []string{"a", "b", "c", "d"}
	phase1 := []string{"a"}
	phase2 := []string{"b"}

	got := SelectPackages(true, phase1, phase2, all, 70.0)
	if !reflect.DeepEqual(got, all) {
		t.Errorf("SelectPackages(forceAll) = %v, want %v", got, all)
	}
}

func TestSelectPackages_RunAllThreshold(t *testing.T) {
	all := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10"}

	cases := []struct {
		name   string
		phase1 []string
		phase2 []string
		want   []string
	}{
		{
			name:   "below threshold selects subset",
			phase1: []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"},
			phase2: []string{},
			want:   []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"},
		},
		{
			name:   "exceeds threshold selects all",
			phase1: []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9"},
			phase2: []string{},
			want:   all,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SelectPackages(false, tc.phase1, tc.phase2, all, 70.0)
			sort.Strings(got)
			want := append([]string(nil), tc.want...)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("SelectPackages = %v, want %v", got, want)
			}
		})
	}
}

func TestSelectPackages_UnionAndDeduplicate(t *testing.T) {
	all := []string{"a", "b", "c", "d"}
	phase1 := []string{"a", "b"}
	phase2 := []string{"b", "c"}

	got := SelectPackages(false, phase1, phase2, all, 100.0)
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SelectPackages = %v, want %v", got, want)
	}
}

func TestSelectPackages_ThresholdAtBoundary(t *testing.T) {
	all := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10"}

	// thresholdCount is floor(70/100 * 10) = 7. Exactly 7 packages should
	// remain as the union, not collapse to all.
	got := SelectPackages(false, []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, nil, all, 70.0)
	want := []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SelectPackages = %v, want %v", got, want)
	}
}

func TestBuildShardPlan(t *testing.T) {
	packages := func(n int) []string {
		out := make([]string, n)
		for i := range n {
			out[i] = fmt.Sprintf("example.com/mod/pkg%02d", i)
		}
		return out
	}

	cases := []struct {
		name          string
		packages      []string
		totalShards   int
		minShardPkgs  int
		wantHasPkg    bool
		wantShardLens []int
	}{
		{
			name:          "empty selection yields single empty shard",
			packages:      nil,
			totalShards:   2,
			minShardPkgs:  30,
			wantHasPkg:    false,
			wantShardLens: []int{0},
		},
		{
			name:          "single shard keeps all packages",
			packages:      packages(60),
			totalShards:   1,
			minShardPkgs:  30,
			wantHasPkg:    true,
			wantShardLens: []int{60},
		},
		{
			name:          "small set uses one shard",
			packages:      packages(8),
			totalShards:   2,
			minShardPkgs:  30,
			wantHasPkg:    true,
			wantShardLens: []int{8},
		},
		{
			name:          "large set splits round-robin",
			packages:      packages(60),
			totalShards:   2,
			minShardPkgs:  30,
			wantHasPkg:    true,
			wantShardLens: []int{30, 30},
		},
		{
			name:         "requested count capped to package count",
			packages:     packages(30),
			totalShards:  31,
			minShardPkgs: 30,
			wantHasPkg:   true,
			wantShardLens: func() []int {
				out := make([]int, 30)
				for i := range out {
					out[i] = 1
				}
				return out
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := BuildShardPlan(tc.packages, tc.totalShards, tc.minShardPkgs)

			if plan.HasPackages != tc.wantHasPkg {
				t.Errorf("BuildShardPlan.HasPackages = %v, want %v", plan.HasPackages, tc.wantHasPkg)
			}
			if !reflect.DeepEqual(plan.SelectedPackages, stringsSorted(tc.packages)) {
				t.Errorf("BuildShardPlan.SelectedPackages = %v, want %v", plan.SelectedPackages, stringsSorted(tc.packages))
			}
			if len(plan.Shards) != len(tc.wantShardLens) {
				t.Fatalf("BuildShardPlan produced %d shards, want %d: %v", len(plan.Shards), len(tc.wantShardLens), plan.Shards)
			}
			for i, want := range tc.wantShardLens {
				if len(plan.Shards[i]) != want {
					t.Errorf("shard %d has %d packages, want %d", i, len(plan.Shards[i]), want)
				}
			}

			// Every nonempty plan shard contains at least one package.
			if plan.HasPackages {
				for i, shard := range plan.Shards {
					if len(shard) == 0 {
						t.Errorf("nonempty plan shard %d is empty", i)
					}
				}
			}

			// Round-robin from the sorted list: shard k holds positions k, k+shardCount, ...
			if plan.HasPackages && len(plan.Shards) > 1 {
				shardCount := len(plan.Shards)
				for k, shard := range plan.Shards {
					for j, pkg := range shard {
						if want := plan.SelectedPackages[j*shardCount+k]; pkg != want {
							t.Errorf("shard %d position %d = %s, want %s", k, j, pkg, want)
						}
					}
				}
			}

			if err := ValidateShardPlan(plan); err != nil {
				t.Errorf("ValidateShardPlan rejected a plan BuildShardPlan produced: %v", err)
			}
		})
	}
}
