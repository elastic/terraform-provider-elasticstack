package main

import (
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
			phase1: []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8"},
			phase2: []string{"p9"},
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

func TestApplyShard(t *testing.T) {
	makePackages := func(n int) []string {
		out := make([]string, n)
		formatDigit := func(v int) string {
			if v == 0 {
				return "0"
			}
			s := ""
			for v > 0 {
				s = string(rune('0'+v%10)) + s
				v /= 10
			}
			return s
		}
		for i := range n {
			out[i] = "pkg" + string(rune('a'+i%26)) + formatDigit(i/26)
		}
		return out
	}

	cases := []struct {
		name     string
		packages []string
		total    int
		index    int
		minShard int
		want     []string
	}{
		{
			name:     "out of range shard",
			packages: []string{"p0", "p1"},
			total:    2,
			index:    2,
			minShard: 30,
			want:     nil,
		},
		{
			name:     "small set shard index zero",
			packages: []string{"p0", "p1", "p2", "p3", "p4"},
			total:    2,
			index:    0,
			minShard: 30,
			want:     []string{"p0", "p1", "p2", "p3", "p4"},
		},
		{
			name:     "small set shard index positive suppressed",
			packages: []string{"p0", "p1", "p2", "p3", "p4"},
			total:    2,
			index:    1,
			minShard: 30,
			want:     nil,
		},
		{
			name:     "large set even positions",
			packages: makePackages(60),
			total:    2,
			index:    0,
			minShard: 30,
			want: func() []string {
				out := make([]string, 0, 30)
				for i := 0; i < 60; i += 2 {
					out = append(out, makePackages(60)[i])
				}
				return out
			}(),
		},
		{
			name:     "large set odd positions",
			packages: makePackages(60),
			total:    2,
			index:    1,
			minShard: 30,
			want: func() []string {
				out := make([]string, 0, 30)
				for i := 1; i < 60; i += 2 {
					out = append(out, makePackages(60)[i])
				}
				return out
			}(),
		},
		{
			name:     "large set split four ways",
			packages: makePackages(40),
			total:    4,
			index:    1,
			minShard: 30,
			want:     []string{"pkgb0", "pkgf0", "pkgj0", "pkgn0", "pkgr0", "pkgv0", "pkgz0", "pkgd1", "pkgh1", "pkgl1"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ApplyShard(tc.packages, tc.total, tc.index, tc.minShard)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ApplyShard = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestApplyShard_TotalShardsOne(t *testing.T) {
	packages := []string{"p0", "p1", "p2"}

	got := ApplyShard(packages, 1, 0, 30)
	if !reflect.DeepEqual(got, packages) {
		t.Errorf("ApplyShard = %v, want %v", got, packages)
	}
}

func TestApplyShard_DifferentMinShardPackages(t *testing.T) {
	// Exactly at the threshold (count == minShardPackages) uses round-robin.
	packages := []string{"p0", "p1", "p2"}

	got := ApplyShard(packages, 2, 1, 3)
	want := []string{"p1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ApplyShard = %v, want %v", got, want)
	}
}
