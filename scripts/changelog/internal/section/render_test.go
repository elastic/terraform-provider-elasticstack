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

package section

import (
	"strings"
	"testing"
)

const (
	testNormalizedFixBulletWant = "- fix bug"

	// Fixtures repeat the changelog contract line verbatim for renderer parity assertions.
	testPRBodyChangelogCustomerImpactNoneSuffix = "## Changelog\nCustomer impact: none\n"
)

func assertRendererSuccess(t *testing.T, res RenderResult) {
	t.Helper()
	if !res.Success || len(res.Errors) != 0 {
		t.Fatalf("expected success got errors=%#v", res.Errors)
	}
}

func TestRenderer_singleFixSummaryBullet(t *testing.T) {
	res := RenderChangelogSection([]MergedPR{{
		Number: 42,
		URL:    "https://github.com/org/repo/pull/42",
		Labels: []string{},
		Body:   "## Changelog\nCustomer impact: fix\nSummary: fixed a nasty bug\n",
	}})
	assertRendererSuccess(t, res)
	if len(res.Included) != 1 || len(res.Excluded) != 0 {
		t.Fatal(res)
	}
	got := res.SectionBody
	if !strings.Contains(got, "### Changes") {
		t.Fatal(got)
	}
	want := "- fixed a nasty bug ([#42](https://github.com/org/repo/pull/42))"
	if !strings.Contains(got, want) {
		t.Fatal(got)
	}
}

func TestRenderer_noneExcludedNoBullets(t *testing.T) {
	res := RenderChangelogSection([]MergedPR{{
		Number: 7,
		URL:    "https://github.com/org/repo/pull/7",
		Labels: []string{},
		Body:   testPRBodyChangelogCustomerImpactNoneSuffix,
	}})
	assertRendererSuccess(t, res)
	if len(res.Included) != 0 || len(res.Excluded) != 1 {
		t.Fatal(res)
	}
	if res.Excluded[0].Reason != rendererExcludedImpactNoneReason {
		t.Fatal(res.Excluded[0])
	}
	if res.SectionBody != "" {
		t.Fatalf("want empty sectionBody, got %q", res.SectionBody)
	}
}

func TestRenderer_noChangelogLabelExcluded(t *testing.T) {
	res := RenderChangelogSection([]MergedPR{{
		Number: 3,
		URL:    "https://github.com/org/repo/pull/3",
		Labels: []string{"no-changelog"},
		Body:   "",
	}})
	assertRendererSuccess(t, res)
	if len(res.Included) != 0 || len(res.Excluded) != 1 {
		t.Fatal(res)
	}
	if res.Excluded[0].Reason != "no-changelog label" {
		t.Fatal(res.Excluded[0])
	}
}

func TestRenderer_missingCustomerImpactFailsAssembly(t *testing.T) {
	res := RenderChangelogSection([]MergedPR{{
		Number: 5,
		URL:    "https://github.com/org/repo/pull/5",
		Labels: []string{},
		Body:   "## Changelog\nSummary: something happened\n",
	}})
	if res.Success || len(res.Errors) != 1 {
		t.Fatal(res)
	}
	if !strings.Contains(res.Errors[0].Reason, "missing the required Customer impact field") {
		t.Fatal(res.Errors[0].Reason)
	}
}

func TestRenderer_breakingImpactWithSubsection(t *testing.T) {
	body := strings.Join([]string{
		"## Changelog",
		"Customer impact: breaking",
		"Summary: removed the old API endpoint",
		"",
		"### Breaking changes",
		"The `/v1/legacy` endpoint has been removed. Migrate to `/v2/endpoint`.",
		"",
	}, "\n")

	res := RenderChangelogSection([]MergedPR{{
		Number: 99,
		URL:    "https://github.com/org/repo/pull/99",
		Labels: []string{},
		Body:   body,
	}})
	assertRendererSuccess(t, res)
	if len(res.Included) != 1 {
		t.Fatal(res)
	}
	sb := res.SectionBody
	if !strings.Contains(sb, "### Breaking changes") || !strings.Contains(sb, "### Changes") {
		t.Fatal(sb)
	}
	if !strings.Contains(sb, "- removed the old API endpoint ([#99](https://github.com/org/repo/pull/99))") ||
		!strings.Contains(sb, "The `/v1/legacy` endpoint has been removed") {
		t.Fatal(sb)
	}
}

func TestRenderer_invalidImpactFailsAssembly(t *testing.T) {
	res := RenderChangelogSection([]MergedPR{{
		Number: 11,
		URL:    "https://github.com/org/repo/pull/11",
		Labels: []string{},
		Body:   "## Changelog\nCustomer impact: refactor\nSummary: cleaned up internals\n",
	}})
	if res.Success || len(res.Errors) != 1 {
		t.Fatal(res)
	}
	if !strings.Contains(res.Errors[0].Reason, "failed validation") {
		t.Fatal(res.Errors[0].Reason)
	}
}

func TestRenderer_missingChangelogFailsAssembly(t *testing.T) {
	res := RenderChangelogSection([]MergedPR{{
		Number: 20,
		URL:    "https://github.com/org/repo/pull/20",
		Labels: []string{},
		Body:   "## Description\nSome description but no changelog section.",
	}})
	if res.Success || len(res.Errors) != 1 {
		t.Fatal(res)
	}
	if !strings.Contains(res.Errors[0].Reason, "no parseable ## Changelog section") {
		t.Fatal(res.Errors[0].Reason)
	}
}

func TestRenderer_mixedBatchCombinedOutput(t *testing.T) {
	breakingBody := strings.Join([]string{
		"## Changelog",
		"Customer impact: breaking",
		"Summary: Remove the old authentication endpoint",
		"",
		"### Breaking changes",
		"",
		"The `/v1/auth` endpoint has been removed. Use `/v2/auth` instead.",
		"",
	}, "\n")

	res := RenderChangelogSection([]MergedPR{
		{
			Number: 101,
			URL:    "https://github.com/org/repo/pull/101",
			Labels: []string{},
			Body:   "## Changelog\nCustomer impact: fix\nSummary: Fix the widget factory\n",
		},
		{
			Number: 102,
			URL:    "https://github.com/org/repo/pull/102",
			Labels: []string{"no-changelog"},
			Body:   "",
		},
		{
			Number: 103,
			URL:    "https://github.com/org/repo/pull/103",
			Labels: []string{},
			Body:   testPRBodyChangelogCustomerImpactNoneSuffix,
		},
		{
			Number: 104,
			URL:    "https://github.com/org/repo/pull/104",
			Labels: []string{},
			Body:   breakingBody,
		},
	})

	assertRendererSuccess(t, res)
	if len(res.Errors) != 0 {
		t.Fatalf("errs %#v", res.Errors)
	}
	if len(res.Included) != 2 || len(res.Excluded) != 2 {
		t.Fatal(res.Included, res.Excluded)
	}
	exMap := map[int]string{}
	for _, e := range res.Excluded {
		exMap[e.PRNumber] = e.Reason
	}
	if exMap[102] != "no-changelog label" || exMap[103] != rendererExcludedImpactNoneReason {
		t.Fatal(exMap)
	}
	sb := res.SectionBody
	if !strings.Contains(sb, "### Breaking changes") || !strings.Contains(sb, "### Changes") ||
		!strings.Contains(sb, "/v1/auth") ||
		!strings.Contains(sb, "- Fix the widget factory ([#101](https://github.com/org/repo/pull/101))") ||
		!strings.Contains(sb, "- Remove the old authentication endpoint ([#104](https://github.com/org/repo/pull/104))") ||
		strings.Contains(sb, "#102") || strings.Contains(sb, "#103") {
		t.Fatal(sb)
	}
}

func TestRenderer_breakingIncludedCarriesBreakingMarkdown(t *testing.T) {
	body := strings.Join([]string{
		"## Changelog",
		"Customer impact: breaking",
		"Summary: An internal refactor removes an undocumented field",
		"",
		"### Breaking changes",
		"This internal refactor technically removes an undocumented field.",
		"",
	}, "\n")

	res := RenderChangelogSection([]MergedPR{{
		Number: 55,
		URL:    "https://github.com/org/repo/pull/55",
		Labels: []string{},
		Body:   body,
	}})
	assertRendererSuccess(t, res)
	if len(res.Included) != 1 || len(res.Excluded) != 0 {
		t.Fatal(res)
	}
	inc := res.Included[0]
	if inc.Summary != "An internal refactor removes an undocumented field" {
		t.Fatal(inc.Summary)
	}
	if inc.BreakingChanges == nil || !strings.Contains(*inc.BreakingChanges, "undocumented field") {
		t.Fatal(inc.BreakingChanges)
	}
	if !strings.Contains(res.SectionBody, "### Breaking changes") || !strings.Contains(res.SectionBody, "undocumented field") {
		t.Fatal(res.SectionBody)
	}
}

func TestRenderer_noneExcludedBreakingChangesUnset(t *testing.T) {
	res := RenderChangelogSection([]MergedPR{{
		Number: 7,
		URL:    "https://github.com/org/repo/pull/7",
		Labels: []string{},
		Body:   testPRBodyChangelogCustomerImpactNoneSuffix,
	}})
	assertRendererSuccess(t, res)
	if len(res.Excluded) != 1 || res.Excluded[0].Reason != rendererExcludedImpactNoneReason {
		t.Fatal(res)
	}
	if res.Excluded[0].BreakingChanges != nil {
		t.Fatal("expected nil BreakingChanges pointer")
	}
}

func TestRenderer_noneWithBreakingRelaxMatchStillRendered(t *testing.T) {
	body := strings.Join([]string{
		"## Changelog",
		"Customer impact: " + impactLiteralNone,
		"",
		"### Breaking changes",
		"This is a legacy internal change with breaking implications.",
	}, "\n")

	res := RenderChangelogSection([]MergedPR{{
		Number: 56,
		URL:    "https://github.com/org/repo/pull/56",
		Labels: []string{},
		Body:   body,
	}})
	assertRendererSuccess(t, res)
	if len(res.Excluded) != 1 || res.Excluded[0].Reason != rendererExcludedImpactNoneReason {
		t.Fatal(res)
	}
	if res.Excluded[0].BreakingChanges == nil ||
		!strings.Contains(*res.Excluded[0].BreakingChanges, "legacy internal change") {
		t.Fatal(res.Excluded[0].BreakingChanges)
	}
	sb := res.SectionBody
	if !strings.Contains(sb, "### Breaking changes") || !strings.Contains(sb, "legacy internal change") {
		t.Fatal(sb)
	}
}

func TestNormalizeBulletPrefix_edgeCases(t *testing.T) {
	want := testNormalizedFixBulletWant
	if NormalizeBulletPrefix("-fix bug") != want ||
		NormalizeBulletPrefix("- fix bug") != want ||
		NormalizeBulletPrefix("* fix bug") != want ||
		NormalizeBulletPrefix("+fix bug") != want {
		t.Fatal("unexpected bullet normalization")
	}
}

func TestRender_matchesEngineForSingleFixPR(t *testing.T) {
	body := "## Changelog\nCustomer impact: fix\nSummary: Alpha line\n"
	sec, err := Parse([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	out := Render(sec, RenderOpts{PRNumber: 42, PRURL: "https://example.com/pr/42"})
	res := RenderChangelogSection([]MergedPR{{
		Number: 42,
		URL:    "https://example.com/pr/42",
		Labels: []string{},
		Body:   body,
	}})
	assertRendererSuccess(t, res)
	if out != res.SectionBody {
		t.Fatalf("Render=%q\nbatch=%q", out, res.SectionBody)
	}
}
