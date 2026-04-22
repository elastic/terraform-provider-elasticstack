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
	"testing"
)

// ---------------------------------------------------------------------------
// Helper factories
// ---------------------------------------------------------------------------

func makeEvidence(prNumbers ...int) *evidenceManifest {
	e := &evidenceManifest{TargetSection: headerUnreleased}
	for _, n := range prNumbers {
		e.PullRequests = append(e.PullRequests, evidencePR{Number: n, Title: "PR"})
	}
	return e
}

func makeProvenance(bullets []provenanceBullet) *provenanceFile {
	return &provenanceFile{Bullets: bullets}
}

// ---------------------------------------------------------------------------
// extractPRReferences tests
// ---------------------------------------------------------------------------

func TestExtractPRReferences_Single(t *testing.T) {
	refs := extractPRReferences("- Fix foo (#123)")
	if len(refs) != 1 || refs[0] != 123 {
		t.Errorf("expected [123], got %v", refs)
	}
}

func TestExtractPRReferences_Multiple(t *testing.T) {
	refs := extractPRReferences("- Fix foo (#123) and bar (#456)")
	if len(refs) != 2 || refs[0] != 123 || refs[1] != 456 {
		t.Errorf("expected [123, 456], got %v", refs)
	}
}

func TestExtractPRReferences_IgnoresURLPaths(t *testing.T) {
	refs := extractPRReferences("see https://github.com/owner/repo/pull/789")
	if len(refs) != 0 {
		t.Errorf("expected [], got %v", refs)
	}
}

func TestExtractPRReferences_Deduplicates(t *testing.T) {
	refs := extractPRReferences("- Fix (#123) and also (#123)")
	if len(refs) != 1 || refs[0] != 123 {
		t.Errorf("expected [123], got %v", refs)
	}
}

// ---------------------------------------------------------------------------
// extractBulletLines tests
// ---------------------------------------------------------------------------

func TestExtractBulletLines_DashBullets(t *testing.T) {
	text := "## [Unreleased]\n\n### Changes\n\n- Foo (#1)\n- Bar (#2)\n"
	bullets := extractBulletLines(text)
	if len(bullets) != 2 {
		t.Errorf("expected 2 bullets, got %d", len(bullets))
	}
	if bullets[0] == "" || !contains(bullets[0], "Foo") {
		t.Errorf("expected first bullet to contain 'Foo', got %q", bullets[0])
	}
}

func TestExtractBulletLines_AsteriskBullets(t *testing.T) {
	text := "* Fix thing (#3)\n* Add widget (#4)"
	bullets := extractBulletLines(text)
	if len(bullets) != 2 {
		t.Errorf("expected 2 bullets, got %d", len(bullets))
	}
}

func TestExtractBulletLines_SkipsNonBulletLines(t *testing.T) {
	text := "## [Unreleased]\n\n### Changes\n\nSome paragraph text.\n\n- Bullet (#5)"
	bullets := extractBulletLines(text)
	if len(bullets) != 1 {
		t.Errorf("expected 1 bullet, got %d", len(bullets))
	}
}

// ---------------------------------------------------------------------------
// looksLikeCommitNarration tests
// ---------------------------------------------------------------------------

func TestLooksLikeCommitNarration_40CharSHA(t *testing.T) {
	result := looksLikeCommitNarration("- Fix a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2")
	if !result {
		t.Error("expected true for 40-char SHA")
	}
}

func TestLooksLikeCommitNarration_ShortHexNoPRRef(t *testing.T) {
	result := looksLikeCommitNarration("- Fix abc1234 issue")
	if !result {
		t.Error("expected true for 7-char hex without PR ref")
	}
}

func TestLooksLikeCommitNarration_NormalBulletWithPRRef(t *testing.T) {
	result := looksLikeCommitNarration("- Fix thing (#123)")
	if result {
		t.Error("expected false for normal bullet with PR ref")
	}
}

func TestLooksLikeCommitNarration_CleanBullet(t *testing.T) {
	result := looksLikeCommitNarration("- Add new resource for Fleet output")
	if result {
		t.Error("expected false for clean bullet")
	}
}

// ---------------------------------------------------------------------------
// extractSectionFromChangelog tests
// ---------------------------------------------------------------------------

const sampleChangelog = `## [Unreleased]

### Changes

- Foo (#1)
- Bar (#2)

## [0.14.3] - 2026-03-02

### Changes

- Baz (#3)

[Unreleased]: https://example.com
[0.14.3]: https://example.com
`

func TestExtractSectionFromChangelog_Unreleased(t *testing.T) {
	section, found := extractSectionFromChangelog(sampleChangelog, headerUnreleased)
	if !found {
		t.Fatal("expected to find section")
	}
	if !contains(section, headerUnreleased) {
		t.Error("section should contain header")
	}
	if !contains(section, "Foo (#1)") {
		t.Error("section should contain Foo (#1)")
	}
	if contains(section, "0.14.3") {
		t.Error("section should not contain 0.14.3")
	}
}

func TestExtractSectionFromChangelog_Versioned(t *testing.T) {
	section, found := extractSectionFromChangelog(sampleChangelog, "## [0.14.3]")
	if !found {
		t.Fatal("expected to find section")
	}
	if !contains(section, "Baz (#3)") {
		t.Error("section should contain Baz (#3)")
	}
	if contains(section, "Foo") {
		t.Error("section should not contain Foo")
	}
}

func TestExtractSectionFromChangelog_Missing(t *testing.T) {
	_, found := extractSectionFromChangelog(sampleChangelog, "## [9.9.9]")
	if found {
		t.Error("expected not to find section")
	}
}

// ---------------------------------------------------------------------------
// validateProvenance tests — valid cases
// ---------------------------------------------------------------------------

func TestValidateProvenance_ValidProvenanceAndChangelog(t *testing.T) {
	evidence := makeEvidence(123, 456)
	provenance := makeProvenance([]provenanceBullet{
		{Text: "Fix foo (#123)", PRNumbers: []int{123}},
		{Text: "Add bar (#456)", PRNumbers: []int{456}},
	})
	changelogSection := "## [Unreleased]\n\n### Changes\n\n- Fix foo (#123)\n- Add bar (#456)\n"

	result := validateProvenance(validateProvenanceParams{
		evidence:         evidence,
		provenance:       provenance,
		changelogSection: changelogSection,
	})
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidateProvenance_EmptyBulletsAndNoChangelogEntries(t *testing.T) {
	evidence := makeEvidence()
	provenance := makeProvenance(nil)
	// No bullet lines in section
	changelogSection := "## [Unreleased]\n\nNo unreleased changes.\n"

	result := validateProvenance(validateProvenanceParams{
		evidence:         evidence,
		provenance:       provenance,
		changelogSection: changelogSection,
	})
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

// ---------------------------------------------------------------------------
// validateProvenance tests — rejection cases
// ---------------------------------------------------------------------------

func TestValidateProvenance_ProvenanceReferencesUnknownPR(t *testing.T) {
	evidence := makeEvidence(123)
	provenance := makeProvenance([]provenanceBullet{
		{Text: "Fix foo (#999)", PRNumbers: []int{999}},
	})
	changelogSection := "## [Unreleased]\n\n- Fix foo (#999)\n"

	result := validateProvenance(validateProvenanceParams{
		evidence:         evidence,
		provenance:       provenance,
		changelogSection: changelogSection,
	})
	if result.Valid {
		t.Error("expected invalid")
	}
	if !anyContains(result.Errors, "#999") {
		t.Errorf("expected error mentioning #999, got: %v", result.Errors)
	}
}

func TestValidateProvenance_ChangelogReferencesUnknownPR(t *testing.T) {
	evidence := makeEvidence(123)
	provenance := makeProvenance([]provenanceBullet{
		{Text: "Fix foo (#123)", PRNumbers: []int{123}},
	})
	changelogSection := "## [Unreleased]\n\n- Fix foo (#123)\n- Fabricated (#888)\n"

	result := validateProvenance(validateProvenanceParams{
		evidence:         evidence,
		provenance:       provenance,
		changelogSection: changelogSection,
	})
	if result.Valid {
		t.Error("expected invalid")
	}
	if !anyContains(result.Errors, "#888") {
		t.Errorf("expected error mentioning #888, got: %v", result.Errors)
	}
}

func TestValidateProvenance_BulletHasNoPRNumbers(t *testing.T) {
	evidence := makeEvidence(123)
	provenance := makeProvenance([]provenanceBullet{
		{Text: "Fix something", PRNumbers: []int{}},
	})
	changelogSection := "## [Unreleased]\n\n- Fix something\n"

	result := validateProvenance(validateProvenanceParams{
		evidence:         evidence,
		provenance:       provenance,
		changelogSection: changelogSection,
	})
	if result.Valid {
		t.Error("expected invalid")
	}
	if !anyContains(result.Errors, "no pr_numbers") {
		t.Errorf("expected error mentioning 'no pr_numbers', got: %v", result.Errors)
	}
}

func TestValidateProvenance_SectionHeaderMismatch(t *testing.T) {
	evidence := makeEvidence(123)
	evidence.TargetSection = headerUnreleased
	provenance := makeProvenance([]provenanceBullet{
		{Text: "Fix foo (#123)", PRNumbers: []int{123}},
	})
	// Section starts with wrong header
	changelogSection := "## [0.14.4] - 2026-04-16\n\n- Fix foo (#123)\n"

	result := validateProvenance(validateProvenanceParams{
		evidence:         evidence,
		provenance:       provenance,
		changelogSection: changelogSection,
		expectedHeader:   headerUnreleased,
	})
	if result.Valid {
		t.Error("expected invalid")
	}
	if !anyContains(result.Errors, "does not match") {
		t.Errorf("expected error mentioning 'does not match', got: %v", result.Errors)
	}
}

func TestValidateProvenance_CommitSHAInBullet(t *testing.T) {
	evidence := makeEvidence(123)
	provenance := makeProvenance([]provenanceBullet{
		{Text: "Fix (#123)", PRNumbers: []int{123}},
	})
	sha := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	changelogSection := "## [Unreleased]\n\n- Fix via " + sha + " (#123)\n"

	result := validateProvenance(validateProvenanceParams{
		evidence:         evidence,
		provenance:       provenance,
		changelogSection: changelogSection,
	})
	if result.Valid {
		t.Error("expected invalid")
	}
	if !anyContains(result.Errors, "commit-level narration") {
		t.Errorf("expected error mentioning 'commit-level narration', got: %v", result.Errors)
	}
}

func TestValidateProvenance_BulletWithoutPRRef(t *testing.T) {
	evidence := makeEvidence()
	provenance := makeProvenance(nil)
	changelogSection := "## [Unreleased]\n\n- Some generic improvement\n"

	result := validateProvenance(validateProvenanceParams{
		evidence:         evidence,
		provenance:       provenance,
		changelogSection: changelogSection,
	})
	if result.Valid {
		t.Error("expected invalid")
	}
	if !anyContains(result.Errors, "no PR reference") {
		t.Errorf("expected error mentioning 'no PR reference', got: %v", result.Errors)
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

func anyContains(ss []string, substr string) bool {
	for _, s := range ss {
		if contains(s, substr) {
			return true
		}
	}
	return false
}
