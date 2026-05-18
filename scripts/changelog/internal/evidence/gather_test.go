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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/evidence"
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
