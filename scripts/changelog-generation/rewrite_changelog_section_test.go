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
)

// ---------------------------------------------------------------------------
// Sample CHANGELOG content
// ---------------------------------------------------------------------------

const sampleCL = `## [Unreleased]

### Changes

- Existing unreleased entry (#100)

## [0.14.3] - 2026-03-02

### Changes

- Stable release entry (#99)

## [0.14.2] - 2026-02-19

### Changes

- Older entry (#98)

[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.3...HEAD
[0.14.3]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.2...v0.14.3
[0.14.2]: https://github.com/elastic/terraform-provider-elasticstack/releases/tag/v0.14.2
`

// ---------------------------------------------------------------------------
// Constants for repeated test strings
// ---------------------------------------------------------------------------

const (
	header0143     = "## [0.14.3] - 2026-03-02"
	header0142     = "## [0.14.2] - 2026-02-19"
	newBodyChanges = "### Changes\n\n- New entry (#200)"
	newBodyRelease = "### Changes\n\n- Release feature (#201)"
)

// ---------------------------------------------------------------------------
// parseChangelog tests
// ---------------------------------------------------------------------------

func TestParseChangelog_IdentifiesAllSections(t *testing.T) {
	parsed := parseChangelog(sampleCL)
	if len(parsed.sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(parsed.sections))
	}
	if parsed.sections[0].header != headerUnreleased {
		t.Errorf("expected section[0] header '## [Unreleased]', got %q", parsed.sections[0].header)
	}
	if parsed.sections[1].header != header0143 {
		t.Errorf("expected section[1] header %q, got %q", header0143, parsed.sections[1].header)
	}
	if parsed.sections[2].header != header0142 {
		t.Errorf("expected section[2] header %q, got %q", header0142, parsed.sections[2].header)
	}
}

func TestParseChangelog_FooterSeparatedFromLastSection(t *testing.T) {
	parsed := parseChangelog(sampleCL)
	if !contains(parsed.footer, "[Unreleased]:") {
		t.Error("footer should contain [Unreleased]:")
	}
	if !contains(parsed.footer, "[0.14.3]:") {
		t.Error("footer should contain [0.14.3]:")
	}
	// Footer should not be part of any section body
	for _, section := range parsed.sections {
		if contains(section.body, "[Unreleased]: https") {
			t.Errorf("section %q body should not contain link definitions", section.header)
		}
	}
}

func TestParseChangelog_RoundTrips(t *testing.T) {
	parsed := parseChangelog(sampleCL)
	roundTripped := serialiseChangelog(parsed)
	if !contains(roundTripped, headerUnreleased) {
		t.Error("round-tripped output should contain ## [Unreleased]")
	}
	if !contains(roundTripped, header0143) {
		t.Errorf("round-tripped output should contain %q", header0143)
	}
	if !contains(roundTripped, "[Unreleased]: https") {
		t.Error("round-tripped output should contain link footer")
	}
}

// ---------------------------------------------------------------------------
// rewriteUnreleased tests
// ---------------------------------------------------------------------------

func TestRewriteUnreleased_ReplacesBody(t *testing.T) {
	result, err := rewriteUnreleased(sampleCL, newBodyChanges)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(result, headerUnreleased) {
		t.Error("result should contain ## [Unreleased]")
	}
	if !contains(result, "New entry (#200)") {
		t.Error("result should contain new entry")
	}
	if contains(result, "Existing unreleased entry") {
		t.Error("result should not contain old entry")
	}
}

func TestRewriteUnreleased_PreservesOtherSections(t *testing.T) {
	result, err := rewriteUnreleased(sampleCL, newBodyChanges)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(result, header0143) {
		t.Error("result should contain 0.14.3 section")
	}
	if !contains(result, "Stable release entry (#99)") {
		t.Error("result should contain stable release entry")
	}
	if !contains(result, header0142) {
		t.Error("result should contain 0.14.2 section")
	}
	if !contains(result, "Older entry (#98)") {
		t.Error("result should contain older entry")
	}
}

func TestRewriteUnreleased_PreservesLinkFooter(t *testing.T) {
	result, err := rewriteUnreleased(sampleCL, newBodyChanges)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(result, "[Unreleased]: https://") {
		t.Error("result should contain Unreleased link definition")
	}
	if !contains(result, "[0.14.3]: https://") {
		t.Error("result should contain 0.14.3 link definition")
	}
}

func TestRewriteUnreleased_ThrowsWhenMissingUnreleased(t *testing.T) {
	noUnreleased := "## [0.14.3] - 2026-03-02\n\n- Entry (#1)\n"
	_, err := rewriteUnreleased(noUnreleased, "- New (#2)")
	if err == nil {
		t.Error("expected error when Unreleased section is missing")
	}
	if !contains(err.Error(), "Unreleased") || !contains(err.Error(), "not found") {
		t.Errorf("error should mention Unreleased not found, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// rewriteRelease tests — insert new section
// ---------------------------------------------------------------------------

func TestRewriteRelease_InsertsNewSection(t *testing.T) {
	result := rewriteRelease(sampleCL, "0.14.4", "2026-04-16", newBodyRelease)
	if !contains(result, "## [0.14.4] - 2026-04-16") {
		t.Error("result should contain new release section header")
	}
	if !contains(result, "Release feature (#201)") {
		t.Error("result should contain release feature")
	}
}

func TestRewriteRelease_NewSectionBeforeExistingVersioned(t *testing.T) {
	result := rewriteRelease(sampleCL, "0.14.4", "2026-04-16", newBodyRelease)
	pos0144 := strings.Index(result, "## [0.14.4]")
	pos0143 := strings.Index(result, "## [0.14.3]")
	if pos0144 >= pos0143 {
		t.Error("new section should appear before 0.14.3")
	}
}

func TestRewriteRelease_PreservesUnreleasedBody(t *testing.T) {
	result := rewriteRelease(sampleCL, "0.14.4", "2026-04-16", newBodyRelease)
	if !contains(result, headerUnreleased) {
		t.Error("result should contain ## [Unreleased]")
	}
	if !contains(result, "Existing unreleased entry") {
		t.Error("result should preserve existing unreleased content")
	}
}

func TestRewriteRelease_PreservesOtherSections(t *testing.T) {
	result := rewriteRelease(sampleCL, "0.14.4", "2026-04-16", newBodyRelease)
	if !contains(result, "Stable release entry (#99)") {
		t.Error("result should contain stable release entry")
	}
	if !contains(result, "Older entry (#98)") {
		t.Error("result should contain older entry")
	}
}

func TestRewriteRelease_PreservesLinkFooter(t *testing.T) {
	result := rewriteRelease(sampleCL, "0.14.4", "2026-04-16", newBodyRelease)
	if !contains(result, "[Unreleased]: https://") {
		t.Error("result should contain Unreleased link definition")
	}
	if !contains(result, "[0.14.3]: https://") {
		t.Error("result should contain 0.14.3 link definition")
	}
}

// ---------------------------------------------------------------------------
// rewriteRelease tests — replace existing section
// ---------------------------------------------------------------------------

const sampleWithExistingRelease = `## [Unreleased]

- No unreleased changes

## [0.14.4] - 2026-04-01

### Changes

- Old release content (#150)

## [0.14.3] - 2026-03-02

### Changes

- Stable release entry (#99)

[Unreleased]: https://example.com
[0.14.4]: https://example.com
[0.14.3]: https://example.com
`

func TestRewriteRelease_ReplacesExistingSection(t *testing.T) {
	newBody := "### Changes\n\n- Updated release content (#160)"
	result := rewriteRelease(sampleWithExistingRelease, "0.14.4", "2026-04-16", newBody)
	if !contains(result, "## [0.14.4] - 2026-04-16") {
		t.Error("result should contain updated release section header")
	}
	if !contains(result, "Updated release content (#160)") {
		t.Error("result should contain updated content")
	}
	if contains(result, "Old release content (#150)") {
		t.Error("result should not contain old content")
	}
}

func TestRewriteRelease_DoesNotDuplicate(t *testing.T) {
	newBody := "### Changes\n\n- Updated release content (#160)"
	result := rewriteRelease(sampleWithExistingRelease, "0.14.4", "2026-04-16", newBody)
	count := strings.Count(result, "## [0.14.4]")
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of ## [0.14.4], got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestRewriteUnreleased_OnlyUnreleasedSection(t *testing.T) {
	minimal := "## [Unreleased]\n\n- Old entry (#1)\n\n[Unreleased]: https://example.com\n"
	result, err := rewriteUnreleased(minimal, "### Changes\n\n- New entry (#2)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(result, headerUnreleased) {
		t.Error("result should contain ## [Unreleased]")
	}
	if !contains(result, "New entry (#2)") {
		t.Error("result should contain new entry")
	}
	if contains(result, "Old entry") {
		t.Error("result should not contain old entry")
	}
}

// ---------------------------------------------------------------------------
// Round-trip tests
// ---------------------------------------------------------------------------

func TestRoundTrip_RewriteUnreleased(t *testing.T) {
	newBody := "### Changes\n\n- Round-trip entry (#300)"
	rewritten, err := rewriteUnreleased(sampleCL, newBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reparsed := parseChangelog(rewritten)

	if len(reparsed.sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(reparsed.sections))
	}
	if reparsed.sections[0].header != headerUnreleased {
		t.Errorf("expected section[0] '## [Unreleased]', got %q", reparsed.sections[0].header)
	}
	if reparsed.sections[1].header != header0143 {
		t.Errorf("expected section[1] %q, got %q", header0143, reparsed.sections[1].header)
	}
	if reparsed.sections[2].header != header0142 {
		t.Errorf("expected section[2] %q, got %q", header0142, reparsed.sections[2].header)
	}
	if !contains(reparsed.sections[0].body, "Round-trip entry (#300)") {
		t.Error("unreleased section should contain new entry")
	}
	if !contains(reparsed.sections[1].body, "Stable release entry (#99)") {
		t.Error("0.14.3 section should be unchanged")
	}
	if !contains(reparsed.sections[2].body, "Older entry (#98)") {
		t.Error("0.14.2 section should be unchanged")
	}
	if !contains(reparsed.footer, "[Unreleased]:") {
		t.Error("footer should contain [Unreleased]:")
	}
	if !contains(reparsed.footer, "[0.14.3]:") {
		t.Error("footer should contain [0.14.3]:")
	}
}

func TestRoundTrip_RewriteReleaseInsert(t *testing.T) {
	newBody := "### Changes\n\n- Round-trip release (#400)"
	rewritten := rewriteRelease(sampleCL, "0.14.4", "2026-04-16", newBody)

	reparsed := parseChangelog(rewritten)

	// Four sections: Unreleased + new 0.14.4 + 0.14.3 + 0.14.2
	if len(reparsed.sections) != 4 {
		t.Fatalf("expected 4 sections, got %d", len(reparsed.sections))
	}
	if reparsed.sections[0].header != headerUnreleased {
		t.Errorf("expected section[0] '## [Unreleased]', got %q", reparsed.sections[0].header)
	}
	if reparsed.sections[1].header != "## [0.14.4] - 2026-04-16" {
		t.Errorf("expected section[1] '## [0.14.4] - 2026-04-16', got %q", reparsed.sections[1].header)
	}
	if reparsed.sections[2].header != header0143 {
		t.Errorf("expected section[2] %q, got %q", header0143, reparsed.sections[2].header)
	}
	if reparsed.sections[3].header != header0142 {
		t.Errorf("expected section[3] %q, got %q", header0142, reparsed.sections[3].header)
	}
	if !contains(reparsed.sections[1].body, "Round-trip release (#400)") {
		t.Error("new release section should contain new entry")
	}
	if !contains(reparsed.sections[0].body, "Existing unreleased entry") {
		t.Error("unreleased section should be preserved")
	}
	if !contains(reparsed.sections[2].body, "Stable release entry (#99)") {
		t.Error("0.14.3 section should be unchanged")
	}
	if !contains(reparsed.footer, "[Unreleased]:") {
		t.Error("footer should be preserved")
	}
}

func TestRoundTrip_RewriteReleaseReplace(t *testing.T) {
	newBody := "### Changes\n\n- Updated via round-trip (#500)"
	rewritten := rewriteRelease(sampleWithExistingRelease, "0.14.4", "2026-05-01", newBody)

	reparsed := parseChangelog(rewritten)

	// Still three sections: Unreleased + updated 0.14.4 + 0.14.3
	if len(reparsed.sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(reparsed.sections))
	}
	if reparsed.sections[1].header != "## [0.14.4] - 2026-05-01" {
		t.Errorf("expected updated header '## [0.14.4] - 2026-05-01', got %q", reparsed.sections[1].header)
	}
	if !contains(reparsed.sections[1].body, "Updated via round-trip (#500)") {
		t.Error("section should contain new content")
	}
	if contains(reparsed.sections[1].body, "Old release content") {
		t.Error("section should not contain old content")
	}
	count := strings.Count(rewritten, "## [0.14.4]")
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of ## [0.14.4], got %d", count)
	}
}
