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
	"strings"
	"testing"
	"time"
)

func TestParseChangelogSectionFull(t *testing.T) {
	body := "## Changelog\nCustomer impact: enhancement\nSummary: Add thing\n\n### Breaking changes\nBe careful\n\n## Other\nignored"
	parsed := parseChangelogSectionFull(body)
	if parsed == nil {
		t.Fatal("expected parsed section")
	}
	if parsed.CustomerImpact != "enhancement" || parsed.Summary != "Add thing" || parsed.BreakingChanges != "Be careful" {
		t.Fatalf("unexpected parsed result: %#v", parsed)
	}
}

func TestValidateChangelogSectionFull(t *testing.T) {
	parsed := &parsedChangelogSection{CustomerImpact: "breaking", Summary: "oops", BreakingChangesHeadingPresent: false}
	errs := validateChangelogSectionFull(parsed)
	if len(errs) == 0 {
		t.Fatal("expected validation error")
	}
}

func TestRenderChangelogSection(t *testing.T) {
	prs := []pullRequestRecord{
		{Number: 1, URL: "https://example/1", Labels: []string{"no-changelog"}, Body: ""},
		{Number: 2, URL: "https://example/2", Labels: nil, Body: "## Changelog\nCustomer impact: none"},
		{Number: 3, URL: "https://example/3", Labels: nil, Body: "## Changelog\nCustomer impact: enhancement\nSummary: Add thing\n\n### Breaking changes\nCareful"},
	}
	result := renderChangelogSection(prs)
	if !result.Success {
		t.Fatalf("expected success, got errors: %#v", result.Errors)
	}
	if !strings.Contains(result.SectionBody, "### Breaking changes") || !strings.Contains(result.SectionBody, "### Changes") {
		t.Fatalf("unexpected section body: %s", result.SectionBody)
	}
	if len(result.Excluded) != 2 || len(result.Included) != 1 {
		t.Fatalf("unexpected included/excluded counts: %#v %#v", result.Included, result.Excluded)
	}
}

func TestRewriteChangelogSection(t *testing.T) {
	content := "# Changelog\n\n## [Unreleased]\n\nOld\n\n## [0.14.4] - 2026-01-01\n\nPrevious"
	_, newSection := buildSectionContent("unreleased", "", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), "### Changes\n\n- New")
	updated := rewriteChangelogSection(content, newSection, "unreleased", "")
	if !strings.Contains(updated, "- New") || strings.Contains(updated, "Old") {
		t.Fatalf("unexpected updated changelog: %s", updated)
	}
}

func TestInsertReleaseSectionAfterUnreleased(t *testing.T) {
	content := "# Changelog\n\n## [Unreleased]\n\n### Changes\n\n- Pending\n\n## [0.14.4] - 2026-01-01\n\nPrevious"
	_, newSection := buildSectionContent("release", "0.14.5", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), "### Changes\n\n- Released")
	updated := rewriteChangelogSection(content, newSection, "release", "0.14.5")
	if !strings.Contains(updated, "## [0.14.5] - 2026-01-02") {
		t.Fatalf("missing release section: %s", updated)
	}
}

// TestRewritePreservesAdjacentSections verifies that replacing the Unreleased
// section does not destroy surrounding sections (title, older releases).
func TestRewritePreservesAdjacentSections(t *testing.T) {
	content := "# Changelog\n\n## [Unreleased]\n\n- Old unreleased\n\n## [0.14.4] - 2026-01-01\n\n- Previous release\n\n## [0.14.3] - 2025-12-01\n\n- Even older"
	_, newSection := buildSectionContent("unreleased", "", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), "### Changes\n\n- New")
	updated := rewriteChangelogSection(content, newSection, "unreleased", "")
	if !strings.Contains(updated, "# Changelog") {
		t.Errorf("title heading lost: %s", updated)
	}
	if !strings.Contains(updated, "## [0.14.4] - 2026-01-01") {
		t.Errorf("adjacent release section lost: %s", updated)
	}
	if !strings.Contains(updated, "## [0.14.3] - 2025-12-01") {
		t.Errorf("older release section lost: %s", updated)
	}
	if !strings.Contains(updated, "- New") {
		t.Errorf("new content missing: %s", updated)
	}
	if strings.Contains(updated, "- Old unreleased") {
		t.Errorf("old unreleased content not replaced: %s", updated)
	}
}

// TestReplaceExistingReleaseSection verifies that re-running in release mode
// replaces an already-written release section rather than duplicating it.
func TestReplaceExistingReleaseSection(t *testing.T) {
	content := "# Changelog\n\n## [Unreleased]\n\n## [0.14.5] - 2026-01-01\n\n### Changes\n\n- Old release entry\n\n## [0.14.4] - 2025-12-01\n\n- Previous"
	_, newSection := buildSectionContent("release", "0.14.5", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), "### Changes\n\n- Updated entry")
	updated := rewriteChangelogSection(content, newSection, "release", "0.14.5")
	if strings.Contains(updated, "- Old release entry") {
		t.Errorf("old release entry not replaced: %s", updated)
	}
	if !strings.Contains(updated, "- Updated entry") {
		t.Errorf("updated entry missing: %s", updated)
	}
	if !strings.Contains(updated, "## [0.14.4] - 2025-12-01") {
		t.Errorf("older section lost: %s", updated)
	}
	// Section header should appear exactly once
	count := strings.Count(updated, "## [0.14.5]")
	if count != 1 {
		t.Errorf("expected exactly 1 release section header, got %d: %s", count, updated)
	}
}

// TestRewriteEmptySectionBody verifies that a section with no user-facing
// changes still writes the header (no content lost or duplicated).
func TestRewriteEmptySectionBody(t *testing.T) {
	content := "# Changelog\n\n## [Unreleased]\n\n- Pending\n\n## [0.14.4] - 2026-01-01\n\n- Previous"
	_, newSection := buildSectionContent("unreleased", "", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), "")
	updated := rewriteChangelogSection(content, newSection, "unreleased", "")
	if !strings.Contains(updated, "## [Unreleased]") {
		t.Errorf("Unreleased header missing: %s", updated)
	}
	if strings.Contains(updated, "- Pending") {
		t.Errorf("old pending entry not removed: %s", updated)
	}
	if !strings.Contains(updated, "## [0.14.4] - 2026-01-01") {
		t.Errorf("adjacent section lost: %s", updated)
	}
}
