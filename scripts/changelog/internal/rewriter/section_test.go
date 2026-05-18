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

package rewriter

import (
	"strings"
	"testing"
)

func TestFindSectionEndStopsAtNextHeading(t *testing.T) {
	t.Parallel()
	lines := []string{"## [Unreleased]", "a", "## [1.0.0] - x", "tail"}
	if got := FindSectionEnd(lines, 0); got != 2 {
		t.Fatalf("FindSectionEnd = %d, want 2", got)
	}
}

func TestFindSectionEndEndOfFile(t *testing.T) {
	t.Parallel()
	lines := []string{"## [Unreleased]", "only"}
	if got := FindSectionEnd(lines, 0); got != 2 {
		t.Fatalf("FindSectionEnd = %d, want 2", got)
	}
}

func TestRewriteSectionUnreleasedReplacesUnreleasedBody(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{
		"# Changelog",
		"",
		"## [Unreleased]",
		"old",
		"",
		"## [1.0.0] - 2020-01-01",
		"released",
	}, "\n")
	rewrite := SectionRewrite{
		Header: "[Unreleased]",
		Body:   "\n### Changes\n\n- fresh ([#1](u))",
	}
	out := mustRewrite(t, []byte(before), rewrite, ModeUnreleased, "")
	s := string(out)
	if !strings.Contains(s, "### Changes") {
		t.Fatalf("expected ### Changes")
	}
	if strings.Contains(s, "\nold\n") {
		t.Fatalf("did not expect old body preserved")
	}
	if !strings.Contains(s, "## [1.0.0]") {
		t.Fatalf("expected downstream section intact")
	}
}

func TestRewriteSectionReleaseReplacesUnreleasedWithVersionSection(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{"# C", "", "## [Unreleased]", "work", "", "## [0.9.0]", "x"}, "\n")
	rewrite := SectionRewrite{
		Header: "[1.0.0] - 2025-01-01",
		Body:   "\n### Changes\n\n- x ([#2](u))",
	}
	out := mustRewrite(t, []byte(before), rewrite, ModeRelease, "1.0.0")
	s := string(out)
	if strings.Contains(s, "## [Unreleased]") {
		t.Fatalf("did not expect Unreleased heading after release rewrite")
	}
	newIdx := strings.Index(s, "## [1.0.0]")
	oldIdx := strings.Index(s, "## [0.9.0]")
	if newIdx == -1 || oldIdx == -1 || newIdx >= oldIdx {
		t.Fatalf("expected new section before ## [0.9.0], got indices %d,%d", newIdx, oldIdx)
	}
}

func TestRewriteSectionReleasePrependsWhenMissingHeadings(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{"# Changelog", "", "## [0.9.0]", "prior release"}, "\n")
	rewrite := SectionRewrite{
		Header: "[1.0.0] - 2026-06-01",
		Body:   "\n### Changes\n\n- leap ([#501](https://example/501))",
	}
	out := mustRewrite(t, []byte(before), rewrite, ModeRelease, "1.0.0")
	s := string(out)
	if !strings.HasPrefix(s, "## [1.0.0]") {
		t.Fatalf("want prefix ## [1.0.0], got %q", firstLine(s))
	}
	if strings.Contains(s, "## [Unreleased]") {
		t.Fatalf("did not expect Unreleased heading")
	}
	vNew := strings.Index(s, "## [1.0.0]")
	vPrev := strings.Index(s, "## [0.9.0]")
	if vNew == -1 || vPrev == -1 || vNew >= vPrev {
		t.Fatalf("expected new heading before ## [0.9.0], got indices %d,%d", vNew, vPrev)
	}
	if !strings.Contains(s, "prior release") {
		t.Fatalf("expected prior release retained")
	}
}

func TestRewriteSectionReleaseReplaceVersionWithoutUnreleased(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{
		"# Changelog",
		"",
		"## [1.0.0] - stale-date",
		"",
		"- stale bullet ([#11](https://example/11))",
		"",
		"## [0.9.0]",
		"older",
	}, "\n")
	rewrite := SectionRewrite{
		Header: "[1.0.0] - 2026-06-15",
		Body:   "\n### Changes\n\n- current ([#502](https://example/502))",
	}
	out := mustRewrite(t, []byte(before), rewrite, ModeRelease, "1.0.0")
	s := string(out)
	if strings.Contains(s, "## [Unreleased]") {
		t.Fatalf("did not expect Unreleased heading")
	}
	if headingsWithPrefixLine(s, "## [1.0.0]") != 1 {
		t.Fatalf("wanted exactly one ## [1.0.0] heading line prefix")
	}
	if !strings.Contains(s, "- current") {
		t.Fatalf("expected current bullet present")
	}
	if strings.Contains(s, "- stale bullet") {
		t.Fatalf("did not expect stale bullet")
	}
	vTen := strings.Index(s, "## [1.0.0]")
	vNine := strings.Index(s, "## [0.9.0]")
	if vTen == -1 || vNine == -1 || vTen >= vNine {
		t.Fatalf("expected ordering 1.0.0 before 0.9.0, indices %d,%d", vTen, vNine)
	}
}

func TestRewriteSectionReleaseRerunCollapsesUnreleasedWhenVersionExists(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{
		"# Changelog",
		"",
		"## [Unreleased]",
		"stale unreleased",
		"",
		"## [1.0.0] - 2020-06-01",
		"",
		"### Changes",
		"",
		"- obsolete ([#10](https://example/10))",
		"",
		"## [0.9.0]",
		"prior",
	}, "\n")
	rewrite := SectionRewrite{
		Header: "[1.0.0] - 2026-05-12",
		Body:   "\n### Changes\n\n- refreshed ([#999](https://example/999))",
	}
	out := mustRewrite(t, []byte(before), rewrite, ModeRelease, "1.0.0")
	s := string(out)
	if strings.Contains(s, "## [Unreleased]") {
		t.Fatalf("did not expect Unreleased heading")
	}
	if headingsWithPrefixLine(s, "## [1.0.0]") != 1 {
		t.Fatalf("wanted exactly one ## [1.0.0] heading line prefix")
	}
	if !strings.Contains(s, "refreshed") {
		t.Fatalf("expected refreshed bullet")
	}
	if strings.Contains(s, "stale unreleased") {
		t.Fatalf("did not expect stale unreleased text")
	}
	if strings.Contains(s, "- obsolete") {
		t.Fatalf("did not expect obsolete bullet")
	}
	newIdx := strings.Index(s, "## [1.0.0]")
	prevIdx := strings.Index(s, "## [0.9.0]")
	if newIdx == -1 || prevIdx == -1 || newIdx >= prevIdx {
		t.Fatalf("expected ordering 1.0.0 before 0.9.0, indices %d,%d", newIdx, prevIdx)
	}
}

func TestRewriteSectionReleaseDedupes2857(t *testing.T) {
	t.Parallel()
	releaseBody := "\n### Changes\n\n" +
		"- First ship ([#2840](https://github.com/elastic/terraform-provider-elasticstack/pull/2840))\n" +
		"- Second ship ([#2841](https://github.com/elastic/terraform-provider-elasticstack/pull/2841))\n"

	unreleasedFixture := "# Log\n\n## [Unreleased]" + releaseBody + "## [0.14.0] - older\nprior\n"

	version := "0.15.0"
	header := "## [" + version + "] - 2026-05-11"
	rewrite := sectionRewriteFromTwoLineHeader(strings.TrimPrefix(header, "## "), releaseBody)
	out := mustRewrite(t, []byte(unreleasedFixture), rewrite, ModeRelease, version)
	s := string(out)
	if strings.Contains(s, "## [Unreleased]") {
		t.Fatalf("did not expect Unreleased heading")
	}
	if strings.Count(s, "First ship") != 1 || strings.Count(s, "Second ship") != 1 {
		t.Fatalf("expected single occurrence of bullets, got counts %d,%d",
			strings.Count(s, "First ship"), strings.Count(s, "Second ship"))
	}
	if headingsAtLineStarts(s, "["+version+"]") != 1 {
		t.Fatalf("wanted exactly one [0.15.0] release heading marker")
	}
}

func headingsAtLineStarts(s, bracketedVersion string) int {
	want := "## " + bracketedVersion
	count := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, want) {
			count++
		}
	}
	return count
}

func sectionRewriteFromTwoLineHeader(header, body string) SectionRewrite {
	return SectionRewrite{
		Header: header,
		Body:   body,
	}
}

func TestRewriteSectionPrependsUnreleasedWhenMissingHeading(t *testing.T) {
	t.Parallel()
	before := "# T\n\n## [1.0.0]\nx"
	rewrite := SectionRewrite{
		Header: "[Unreleased]",
		Body:   "\n### Changes\n\n- y ([#3](u))",
	}
	out := mustRewrite(t, []byte(before), rewrite, ModeUnreleased, "")
	s := string(out)
	if !strings.HasPrefix(s, "## [Unreleased]") {
		t.Fatalf("expected leading Unreleased heading, got %q", firstLine(s))
	}
}

func TestRewriteSectionIdempotent(t *testing.T) {
	t.Parallel()
	content := strings.Join([]string{
		"# Changelog",
		"",
		"## [Unreleased]",
		"old body",
		"",
		"## [1.0.0] - 2020",
		"tail",
	}, "\n")
	rewrite := SectionRewrite{
		Header: "[Unreleased]",
		Body:   "new unified content\n(without embedded headings)",
	}
	firstPass := mustRewrite(t, []byte(content), rewrite, ModeUnreleased, "")
	secondPass := mustRewrite(t, firstPass, rewrite, ModeUnreleased, "")
	if string(firstPass) != string(secondPass) {
		t.Fatalf("RewriteSection should be idempotent; first != second:\n%s\nvs\n%s", firstPass, secondPass)
	}
}

func TestRewriteSectionReleaseIdempotentDualRange(t *testing.T) {
	t.Parallel()
	releaseBody := "\n### Changes\n\n- Bullet ([#2840](https://example))\n"

	unreleasedFixture := "# Log\n\n## [Unreleased]" + releaseBody + "## [0.14.0] - older\nprior\n"

	version := "0.15.0"
	header := "[" + version + "] - 2026-05-11"
	rewrite := sectionRewriteFromTwoLineHeader(header, releaseBody)

	firstPass := mustRewrite(t, []byte(unreleasedFixture), rewrite, ModeRelease, version)
	secondPass := mustRewrite(t, firstPass, rewrite, ModeRelease, version)
	if string(firstPass) != string(secondPass) {
		t.Fatalf("RewriteSection release re-run must be stable")
	}
}

func headingsWithPrefixLine(markdown string, headingPrefix string) int {
	count := 0
	for _, line := range strings.Split(markdown, "\n") {
		if strings.HasPrefix(line, headingPrefix) {
			count++
		}
	}
	return count
}

func mustRewrite(tb testing.TB, content []byte, rewrite SectionRewrite, mode RewriteMode, targetVersion string) []byte {
	tb.Helper()
	out, err := RewriteSection(content, rewrite, mode, targetVersion)
	if err != nil {
		tb.Fatalf("RewriteSection error: %v", err)
	}
	return out
}

func firstLine(s string) string {
	idx := strings.IndexByte(s, '\n')
	if idx == -1 {
		return s
	}
	return s[:idx]
}
