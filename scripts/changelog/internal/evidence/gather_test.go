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

package evidence_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/engine"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/evidence"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/section"
)

func TestClassifyPullRequestForChangelog_userFacingLabelWinsOverInternalDocs(t *testing.T) {
	t.Parallel()

	res := evidence.ClassifyPullRequestForChangelog(
		[]string{"bug", "documentation"},
		"octocat",
		[]string{"docs/guide.md"},
	)

	if res.Classification != "user-facing" || res.InclusionRationale == nil ||
		!strings.Contains(*res.InclusionRationale, "bug") || res.ExclusionRationale != nil {
		t.Fatalf("unexpected: %+v", res)
	}
}

func TestClassifyPullRequestForChangelog_openspecOnlyInternal(t *testing.T) {
	t.Parallel()

	res := evidence.ClassifyPullRequestForChangelog(
		[]string{},
		"octocat",
		[]string{"openspec/changes/example/tasks.md"},
	)

	if res.Classification != "internal" || res.ExclusionRationale == nil ||
		!strings.Contains(*res.ExclusionRationale, "openspec") {
		t.Fatalf("unexpected: %+v", res)
	}
}

func TestClassifyPullRequestForChangelog_providerPathOverridesInternalLabel(t *testing.T) {
	t.Parallel()

	res := evidence.ClassifyPullRequestForChangelog(
		[]string{"internal"},
		"octocat",
		[]string{"internal/provider/resource.go"},
	)

	if res.Classification != "user-facing" || res.InclusionRationale == nil ||
		!strings.Contains(*res.InclusionRationale, "provider implementation paths") ||
		res.ExclusionRationale != nil {
		t.Fatalf("unexpected: %+v", res)
	}
}

func TestClassifyPullRequestForChangelog_automatedDependabotInternal(t *testing.T) {
	t.Parallel()

	res := evidence.ClassifyPullRequestForChangelog(
		[]string{},
		"dependabot[bot]",
		[]string{"go.mod"},
	)

	if res.Classification != "internal" || res.ExclusionRationale == nil ||
		!strings.Contains(*res.ExclusionRationale, "dependabot") {
		t.Fatalf("unexpected: %+v", res)
	}
}

func TestParseCommitShas(t *testing.T) {
	t.Parallel()

	got := evidence.ParseCommitShas("\nabc\n\ndef \n")
	if !reflect.DeepEqual(got, []string{"abc", "def"}) {
		t.Fatalf("got %#v", got)
	}
}

func TestSelectMergedPullRequests_orderAndDedup(t *testing.T) {
	t.Parallel()

	open := MergeCand(2, "open", false, "x")
	noMerge := MergeCand(3, "closed", false, "z")

	got := evidence.SelectMergedPullRequests([]evidence.MergeCandidate{
		MergeCand(1, "closed", true, "first"),
		open,
		MergeCand(1, "closed", true, "dup"),
		noMerge,
	})

	if len(got) != 1 || got[0].Number != 1 {
		t.Fatalf("got %+v", got)
	}
	if got[0].Title != "first" {
		t.Fatalf("first-wins title: got %q want first", got[0].Title)
	}
}

func MergeCand(n int, state string, merged bool, title string) evidence.MergeCandidate {
	return evidence.MergeCandidate{Number: n, State: state, MergedAt: merged, Title: title}
}

func TestBuildPullRequestEvidence_normalizedRow(t *testing.T) {
	t.Parallel()

	wantInc := "Has user-facing label(s): enhancement"

	got := evidence.BuildPullRequestEvidence(
		42,
		"Example PR",
		"https://example.test/pr/42",
		"abc123",
		"contributor",
		[]string{"enhancement"},
		[]string{"pkg/example.go", "docs/guide.md"},
	)

	wantIncPtr := wantInc

	want := evidence.PullRequestEvidence{
		Number:             42,
		Title:              "Example PR",
		URL:                "https://example.test/pr/42",
		MergeCommitSHA:     "abc123",
		Author:             "contributor",
		Labels:             []string{"enhancement"},
		TouchedFiles:       []string{"pkg/example.go", "docs/guide.md"},
		Classification:     "user-facing",
		InclusionRationale: &wantIncPtr,
		ExclusionRationale: nil,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

// pullFilesLister matches evidence.GatherOptions.ListPullRequestFilenames.
type pullFilesLister func(context.Context, string, string, int) ([]string, error)

// gatherStub implements engine.MergedPRGatherer for Gather orchestration tests.
type gatherStub struct {
	Recs            []section.MergedPR
	Warn            []string
	Err             error
	GotCompareRange string
}

func (g *gatherStub) GatherMergedPRs(
	ctx context.Context, owner, repo, compareRange string,
) ([]section.MergedPR, []string, error) {
	_, _, _ = ctx, owner, repo
	g.GotCompareRange = compareRange
	if g.Err != nil {
		return nil, append([]string(nil), g.Warn...), g.Err
	}
	out := append([]section.MergedPR(nil), g.Recs...)
	return out, append([]string(nil), g.Warn...), nil
}

func TestGather_orchestration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fixedNow := time.Date(2026, 5, 1, 15, 30, 0, 123000000, time.UTC)
	fixedNowFn := func() time.Time { return fixedNow }

	baseOpts := func(stub engine.MergedPRGatherer, listFn pullFilesLister) evidence.GatherOptions {
		return evidence.GatherOptions{
			Owner:                    "elastic",
			Repo:                     "repo",
			CompareRange:             "v9..HEAD",
			TargetVersion:            "2.0.0",
			PreviousTag:              "v1.0.0",
			Mode:                     "release",
			PRGatherer:               stub,
			ListPullRequestFilenames: listFn,
			Now:                      fixedNowFn,
		}
	}

	threePRStub := func() *gatherStub {
		return &gatherStub{
			Recs: []section.MergedPR{
				{
					Number: 101, Title: "UF", URL: "https://ex/101",
					Labels: []string{"bug"}, Body: "",
					MergeCommitSHA: "sha101", AuthorLogin: "a1",
				},
				{
					Number: 102, Title: "In", URL: "https://ex/102",
					Labels: []string{"internal"}, Body: "",
					MergeCommitSHA: "sha102", AuthorLogin: "a2",
				},
				{
					Number: 103, Title: "Un", URL: "https://ex/103",
					Labels: []string{}, Body: "",
					MergeCommitSHA: "sha103", AuthorLogin: "a3",
				},
			},
		}
	}

	listByPR := func(errOnPR int, errVal error) pullFilesLister {
		return func(_ context.Context, _, _ string, pr int) ([]string, error) {
			if pr == errOnPR {
				return nil, errVal
			}
			switch pr {
			case 101:
				return []string{"CHANGELOG.md"}, nil
			case 102:
				return []string{"README.md"}, nil
			case 103:
				return []string{"misc.txt"}, nil
			default:
				return nil, nil
			}
		}
	}

	tests := []struct {
		name              string
		opts              evidence.GatherOptions
		wantErr           bool
		wantErrContains   string
		wantWarnSubstr    []string
		wantGeneratedAt   string
		wantPRCount       int
		wantUF            int
		wantInternal      int
		wantUncertain     int
		wantTargetSection string
		wantCompareInMan  string
		checkTitles       []string // len must match merged PR order when set
		checkEmptyFilesPR int      // PR number expected to have nil/empty touched_files
		assertStubHEAD    bool
	}{
		{
			name: "success_three_prs_manifest_shape",
			opts: func() evidence.GatherOptions {
				st := threePRStub()
				o := baseOpts(st, listByPR(0, nil))
				o.CompareRange = "v9..HEAD"
				return o
			}(),
			wantGeneratedAt:   evidence.FormatGeneratedAtISO(fixedNow),
			wantPRCount:       3,
			wantUF:            1,
			wantInternal:      1,
			wantUncertain:     1,
			wantTargetSection: "## [2.0.0] - 2026-05-01",
			wantCompareInMan:  "v9..HEAD",
			checkTitles:       []string{"UF", "In", "Un"},
		},
		{
			name: "pr_gatherer_error",
			opts: func() evidence.GatherOptions {
				st := &gatherStub{Err: errors.New("upstream")}
				o := baseOpts(st, listByPR(0, nil))
				return o
			}(),
			wantErr:         true,
			wantErrContains: "gather merged pull requests",
			wantWarnSubstr:  nil,
			wantPRCount:     -1,
		},
		{
			name: "validation_empty_owner",
			opts: func() evidence.GatherOptions {
				o := baseOpts(threePRStub(), listByPR(0, nil))
				o.Owner = ""
				return o
			}(),
			wantErr:         true,
			wantErrContains: "owner must be non-empty",
			wantPRCount:     -1,
		},
		{
			name: "list_files_error_one_pr",
			opts: func() evidence.GatherOptions {
				st := threePRStub()
				o := baseOpts(st, listByPR(102, errors.New("rate limit")))
				return o
			}(),
			wantGeneratedAt:   evidence.FormatGeneratedAtISO(fixedNow),
			wantPRCount:       3,
			wantUF:            1,
			wantInternal:      1,
			wantUncertain:     1,
			wantTargetSection: "## [2.0.0] - 2026-05-01",
			wantCompareInMan:  "v9..HEAD",
			wantWarnSubstr:    []string{"Failed to list files for PR #102"},
			checkEmptyFilesPR: 102,
		},
		{
			name: "empty_pr_list",
			opts: func() evidence.GatherOptions {
				st := &gatherStub{Recs: nil}
				o := baseOpts(st, listByPR(0, nil))
				return o
			}(),
			wantGeneratedAt:   evidence.FormatGeneratedAtISO(fixedNow),
			wantPRCount:       0,
			wantUF:            0,
			wantInternal:      0,
			wantUncertain:     0,
			wantTargetSection: "## [2.0.0] - 2026-05-01",
			wantCompareInMan:  "v9..HEAD",
		},
		{
			name: "default_compare_range_HEAD",
			opts: func() evidence.GatherOptions {
				st := &gatherStub{Recs: []section.MergedPR{
					{Number: 1, Title: "Only", URL: "u", Labels: []string{"bug"}, MergeCommitSHA: "m", AuthorLogin: "u"},
				}}
				o := baseOpts(st, listByPR(0, nil))
				o.CompareRange = ""
				return o
			}(),
			wantGeneratedAt:   evidence.FormatGeneratedAtISO(fixedNow),
			wantPRCount:       1,
			wantUF:            1,
			wantInternal:      0,
			wantUncertain:     0,
			wantTargetSection: "## [2.0.0] - 2026-05-01",
			wantCompareInMan:  "HEAD",
			assertStubHEAD:    true,
			checkTitles:       []string{"Only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stub, _ := tt.opts.PRGatherer.(*gatherStub)

			man, warns, err := evidence.Gather(ctx, tt.opts)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("err: %v (want contains %q)", err, tt.wantErrContains)
				}
				if !reflect.DeepEqual(man, evidence.Manifest{}) {
					t.Fatalf("want empty manifest on error, got %+v", man)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, sub := range tt.wantWarnSubstr {
				found := false
				for _, w := range warns {
					if strings.Contains(w, sub) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("missing warning containing %q in %#v", sub, warns)
				}
			}
			if man.GeneratedAt != tt.wantGeneratedAt {
				t.Fatalf("generated_at: got %q want %q", man.GeneratedAt, tt.wantGeneratedAt)
			}
			if man.TargetSection != tt.wantTargetSection {
				t.Fatalf("target_section: got %q want %q", man.TargetSection, tt.wantTargetSection)
			}
			if man.CompareRange != tt.wantCompareInMan {
				t.Fatalf("compare_range: got %q want %q", man.CompareRange, tt.wantCompareInMan)
			}
			if tt.wantPRCount >= 0 && man.PRCount != tt.wantPRCount {
				t.Fatalf("pr_count: got %d want %d", man.PRCount, tt.wantPRCount)
			}
			if man.UserFacingCount != tt.wantUF {
				t.Fatalf("user_facing_count: got %d want %d", man.UserFacingCount, tt.wantUF)
			}
			if man.InternalCount != tt.wantInternal {
				t.Fatalf("internal_count: got %d want %d", man.InternalCount, tt.wantInternal)
			}
			if man.UncertainCount != tt.wantUncertain {
				t.Fatalf("uncertain_count: got %d want %d", man.UncertainCount, tt.wantUncertain)
			}
			if len(tt.checkTitles) != 0 {
				if len(man.PullRequests) != len(tt.checkTitles) {
					t.Fatalf("pull_requests len %d want %d", len(man.PullRequests), len(tt.checkTitles))
				}
				for i, wantTitle := range tt.checkTitles {
					if man.PullRequests[i].Title != wantTitle {
						t.Fatalf("PR[%d].Title: got %q want %q", i, man.PullRequests[i].Title, wantTitle)
					}
				}
			}
			if tt.checkEmptyFilesPR != 0 {
				for _, row := range man.PullRequests {
					if row.Number == tt.checkEmptyFilesPR {
						if len(row.TouchedFiles) != 0 {
							t.Fatalf("PR %d touched_files: got %#v want empty", row.Number, row.TouchedFiles)
						}
						return
					}
				}
				t.Fatalf("missing PR %d in manifest", tt.checkEmptyFilesPR)
			}
			if tt.assertStubHEAD && stub != nil && stub.GotCompareRange != "HEAD" {
				t.Fatalf("gatherer compare range: got %q want HEAD", stub.GotCompareRange)
			}
		})
	}
}

func TestBuildEvidenceManifest_counts_target_section(t *testing.T) {
	t.Parallel()

	genAt := time.Date(2026, 4, 17, 8, 0, 0, 0, time.UTC)

	got := evidence.BuildEvidenceManifest(
		"release",
		"1.2.3",
		"v1.2.2",
		"v1.2.2..HEAD",
		[]evidence.PullRequestEvidence{
			{Classification: "user-facing"},
			{Classification: "internal"},
			{Classification: "uncertain"},
		},
		genAt,
	)

	want := evidence.Manifest{
		GeneratedAt:       "2026-04-17T08:00:00.000Z",
		Mode:              "release",
		TargetSection:     "## [1.2.3] - 2026-04-17",
		TargetSectionMode: "release",
		TargetVersion:     "1.2.3",
		PreviousTag:       "v1.2.2",
		CompareRange:      "v1.2.2..HEAD",
		PRCount:           3,
		UserFacingCount:   1,
		InternalCount:     1,
		UncertainCount:    1,
		PullRequests: []evidence.PullRequestEvidence{
			{Classification: "user-facing"},
			{Classification: "internal"},
			{Classification: "uncertain"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}
