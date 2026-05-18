// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package section

import (
	"embed"
	"errors"
	"path"
	"strings"
	"testing"
)

//go:embed testdata/*.md
var parserBodies embed.FS

func fixtureBody(tb testing.TB, name string) string {
	tb.Helper()
	b, err := parserBodies.ReadFile(path.Join("testdata", name))
	if err != nil {
		tb.Fatalf("load fixture %s: %v", name, err)
	}
	return string(b)
}

func trimStartFixture(s string) string {
	return strings.TrimLeft(s, "\n\r\t ")
}

func mustParse(tb testing.TB, body string) Section {
	tb.Helper()
	sec, err := Parse([]byte(body))
	if err != nil {
		tb.Fatalf("Parse: %v", err)
	}
	return sec
}

func errorsAny(errs []string, pred func(string) bool) bool {
	for _, e := range errs {
		if pred(e) {
			return true
		}
	}
	return false
}

// Small fixtures (verbatim trimStart parity with pr-changelog-parser.test.mjs).

const bodyNoneNoSummary = `## Changelog

Customer impact: none
`

const bodyNoChangelog = `## Description

No changelog section here.

## Notes

Some notes.
`

const bodyInvalidImpact = `## Changelog

Customer impact: patch
Summary: Some change
`

const bodyFixMissingSummary = `## Changelog

Customer impact: fix
`

const bodyBreakingChangesEmpty = `## Changelog

Customer impact: breaking
Summary: A breaking change

### Breaking changes

## Other section
`

const bodyEnhancementWithSummary = `## Changelog

Customer impact: enhancement
Summary: Add support for index lifecycle management
`

const bodyBreakingNoBreakingSection = `## Changelog

Customer impact: breaking
Summary: Remove deprecated attribute
`

const bodyFixWithBreakingSection = `## Changelog

Customer impact: fix
Summary: A fix with breaking section

### Breaking changes

Some content.
`

const bodyEnhancementWithBreakingSection = `## Changelog

Customer impact: enhancement
Summary: An enhancement with breaking section

### Breaking changes

Some content.
`

const bodyNoneWithBreakingSection = `## Changelog

Customer impact: none

### Breaking changes

Some content.
`

const bodyInvalidImpactWithBreakingSection = `## Changelog

Customer impact: patch
Summary: Some change with breaking section

### Breaking changes

Some content.
`

const bodySummaryEmptyValue = `## Changelog

Customer impact: fix
Summary:
`

const bodyChangelogLastSection = `## Description

Fixes a bug.

## Changelog

Customer impact: fix
Summary: Fix handling of nil pointer in cluster client
`

func TestParse_returnsErrorWhenNoChangelogSection(t *testing.T) {
	_, err := Parse([]byte(trimStartFixture(bodyNoChangelog)))
	if !errors.Is(err, ErrNoChangelogSection) {
		t.Fatalf("expected ErrNoChangelogSection, got %v", err)
	}
}

func TestParse_returnsErrorForEmptyString(t *testing.T) {
	_, err := Parse([]byte(""))
	if !errors.Is(err, ErrNoChangelogSection) {
		t.Fatalf("expected ErrNoChangelogSection, got %v", err)
	}
}

func TestParse_parsesFixWithSummary(t *testing.T) {
	body := fixtureBody(t, "parse_fix_with_summary.md")
	sec := mustParse(t, body)
	if sec.ImpactRaw != "fix" || sec.Summary != "Correct handling of empty API responses" {
		t.Fatalf("unexpected: %#v", sec)
	}
	if sec.BreakingChanges != "" {
		t.Fatalf("expected empty breaking, got %q", sec.BreakingChanges)
	}
}

func TestParse_parsesNoneWithNoSummary(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyNoneNoSummary))
	if sec.ImpactRaw != "none" || sec.Summary != "" {
		t.Fatalf("unexpected: %#v", sec)
	}
}

func TestParse_parsesBreakingWithExtractedBody(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_breaking_full.md"))
	if sec.ImpactRaw != "breaking" {
		t.Fatalf("impact: %q", sec.ImpactRaw)
	}
	wantSum := "Remove deprecated attribute from elasticstack_kibana_slo"
	if sec.Summary != wantSum {
		t.Fatalf("summary got %q want %q", sec.Summary, wantSum)
	}
	if !strings.Contains(sec.BreakingChanges, "legacy_mode") {
		t.Fatalf("breaking: %q", sec.BreakingChanges)
	}
	if !strings.Contains(sec.BreakingChanges, "```hcl") {
		t.Fatalf("breaking: %q", sec.BreakingChanges)
	}
}

func TestParse_preservesInvalidImpactLiteral(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyInvalidImpact))
	if sec.ImpactRaw != "patch" {
		t.Fatalf("impact: %q", sec.ImpactRaw)
	}
}

func TestValidate_fixWithSummaryValid(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_fix_with_summary.md"))
	ok, errs := ValidateChangelogSection(&sec)
	if !ok || len(errs) != 0 {
		t.Fatalf("ok=%v errs=%#v", ok, errs)
	}
}

func TestValidate_noneNoSummaryValid(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyNoneNoSummary))
	ok, errs := ValidateChangelogSection(&sec)
	if !ok {
		t.Fatalf("errs=%#v", errs)
	}
}

func TestValidate_breakingFullValid(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_breaking_full.md"))
	ok, errs := ValidateChangelogSection(&sec)
	if !ok {
		t.Fatalf("errs=%#v", errs)
	}
}

func TestValidate_invalidImpact(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyInvalidImpact))
	ok, errs := ValidateChangelogSection(&sec)
	if ok || !errorsAny(errs, func(e string) bool { return strings.Contains(e, "patch") }) {
		t.Fatalf("ok=%v errs=%#v", ok, errs)
	}
}

func TestValidate_fixMissingSummaryInvalid(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyFixMissingSummary))
	ok, errs := ValidateChangelogSection(&sec)
	if ok || !errorsAny(errs, func(e string) bool { return strings.Contains(e, "Summary") }) {
		t.Fatalf("ok=%v errs=%#v", ok, errs)
	}
}

func TestValidate_returnsErrorWhenParsedNil(t *testing.T) {
	ok, errs := ValidateChangelogSection(nil)
	if ok || len(errs) == 0 {
		t.Fatalf("ok=%v errs=%#v", ok, errs)
	}
}

func TestValidateFull_emptyBreakingHeading(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyBreakingChangesEmpty))
	ok, errs := ValidateChangelogSectionFull(&sec, ValidateOpts{})
	if ok || !errorsAny(errs, func(e string) bool {
		return strings.Contains(e, "Breaking changes") && strings.Contains(e, "no content")
	}) {
		t.Fatalf("ok=%v errs=%#v", ok, errs)
	}
}

func TestExtractBreakingChanges_absent(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_fix_with_summary.md"))
	_, ok := ExtractBreakingChanges(sec.Raw)
	if ok {
		t.Fatal("expected absent")
	}
}

func TestExtractBreakingChanges_emptyBody(t *testing.T) {
	_, ok := ExtractBreakingChanges("")
	if ok {
		t.Fatal("expected absent")
	}
}

func TestExtractBreakingChanges_fullMarkdown(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_breaking_full.md"))
	c, ok := ExtractBreakingChanges(sec.Raw)
	if !ok {
		t.Fatal("expected content")
	}
	if !strings.Contains(c, "legacy_mode") || !strings.Contains(c, "- Attribute") || !strings.Contains(c, "```hcl") || strings.Contains(c, "## Other section") {
		t.Fatalf("unexpected: %q", c)
	}
}

func TestExtractBreakingChanges_headingButEmpty(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyBreakingChangesEmpty))
	_, ok := ExtractBreakingChanges(sec.Raw)
	if ok {
		t.Fatal("expected absent")
	}
}

func TestExtractBreakingChanges_fencedOnly(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_breaking_changes_fenced_only.md"))
	c, ok := ExtractBreakingChanges(sec.Raw)
	if !ok {
		t.Fatal("expected content")
	}
	if !strings.Contains(c, "```json") || !strings.Contains(c, `"removed": true`) {
		t.Fatalf("bad: %q", c)
	}
}

func TestParse_breakingHeadingOutsideChangelogIgnored(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_breaking_before_changelog.md"))
	if sec.ImpactRaw != "fix" {
		t.Fatal(sec.ImpactRaw)
	}
	if sec.BreakingChanges != "" {
		t.Fatalf("want empty, got %q", sec.BreakingChanges)
	}
}

func TestParseFull_breakingHeadingOutsideChangelog(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_breaking_before_changelog.md"))
	if sec.BreakingHeadingPresent {
		t.Fatal("expected false")
	}
	if sec.BreakingChanges != "" {
		t.Fatal()
	}
}

func TestParse_enhancementSummary(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyEnhancementWithSummary))
	if sec.ImpactRaw != "enhancement" || sec.Summary != "Add support for index lifecycle management" || sec.BreakingChanges != "" {
		t.Fatalf("%+v", sec)
	}
}

func TestValidate_enhancementSummary(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyEnhancementWithSummary))
	ok, errs := ValidateChangelogSection(&sec)
	if !ok {
		t.Fatal(errs)
	}
}

func TestValidateFull_breakingFullPositive(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_breaking_full.md"))
	ok, errs := ValidateChangelogSectionFull(&sec, ValidateOpts{})
	if !ok {
		t.Fatal(errs)
	}
}

func TestValidateFull_breakingNeedsSubsection(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyBreakingNoBreakingSection))
	ok, errs := ValidateChangelogSectionFull(&sec, ValidateOpts{})
	if ok || !errorsAny(errs, func(e string) bool { return strings.Contains(e, "breaking") && strings.Contains(e, "Breaking changes") }) {
		t.Fatalf("ok=%v errs=%#v", ok, errs)
	}
}

func TestValidateFull_ruleC_fixBreakingHeading(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyFixWithBreakingSection))
	ok, errs := ValidateChangelogSectionFull(&sec, ValidateOpts{})
	if ok || !errorsAny(errs, func(e string) bool {
		return strings.Contains(e, "Breaking changes") && strings.Contains(e, "breaking") && strings.Contains(e, "remove")
	}) {
		t.Fatalf("ok=%v errs=%#v", ok, errs)
	}
}

func TestValidateFull_ruleC_enhancementBreakingHeading(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyEnhancementWithBreakingSection))
	ok, errs := ValidateChangelogSectionFull(&sec, ValidateOpts{})
	if ok || !errorsAny(errs, func(e string) bool {
		return strings.Contains(e, "Breaking changes") && strings.Contains(e, "breaking") && strings.Contains(e, "remove")
	}) {
		t.Fatalf("ok=%v errs=%#v", ok, errs)
	}
}

func TestValidateFull_ruleC_noneBreakingHeading(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyNoneWithBreakingSection))
	ok, errs := ValidateChangelogSectionFull(&sec, ValidateOpts{})
	if ok || !errorsAny(errs, func(e string) bool {
		return strings.Contains(e, "Breaking changes") && strings.Contains(e, "breaking") && strings.Contains(e, "remove")
	}) {
		t.Fatalf("ok=%v errs=%#v", ok, errs)
	}
}

func TestValidateFull_invalidImpactDoesNotTriggerRuleC(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyInvalidImpactWithBreakingSection))
	ok, errs := ValidateChangelogSectionFull(&sec, ValidateOpts{})
	if ok {
		t.Fatal("expected invalid")
	}
	if !errorsAny(errs, func(e string) bool { return strings.Contains(e, "Invalid Customer impact") }) {
		t.Fatalf("errs=%#v", errs)
	}
	for _, e := range errs {
		if strings.Contains(e, "Breaking changes") && strings.Contains(e, "remove") {
			t.Fatalf("unexpected rule C: %q", e)
		}
	}
}

func TestValidateFull_breakingWithContentValid(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_breaking_full.md"))
	ok, errs := ValidateChangelogSectionFull(&sec, ValidateOpts{})
	if !ok {
		t.Fatal(errs)
	}
}

func TestValidate_summaryEmptyLineInvalid(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodySummaryEmptyValue))
	ok, errs := ValidateChangelogSection(&sec)
	if ok || !errorsAny(errs, func(e string) bool { return strings.Contains(e, "Summary") }) {
		t.Fatalf("ok=%v errs=%#v", ok, errs)
	}
}

func TestParse_changelogLastSection(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyChangelogLastSection))
	if sec.ImpactRaw != "fix" || sec.Summary != "Fix handling of nil pointer in cluster client" || sec.BreakingChanges != "" {
		t.Fatalf("%+v", sec)
	}
}

func TestValidate_changelogLastSection(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyChangelogLastSection))
	ok, errs := ValidateChangelogSection(&sec)
	if !ok {
		t.Fatal(errs)
	}
}

func TestParse_multiFollowingSections(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_changelog_multi_follow.md"))
	if sec.ImpactRaw != "enhancement" || sec.Summary != "Add support for cross-cluster replication" || sec.BreakingChanges != "" {
		t.Fatalf("%+v", sec)
	}
}

func TestExtractBreakingChanges_endMarkerCutsContent(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_end_marker.md"))
	c, ok := ExtractBreakingChanges(sec.Raw)
	if !ok || !strings.Contains(c, "Content before marker.") || strings.Contains(c, "Content after marker.") {
		t.Fatalf("got %q ok=%v", c, ok)
	}
}

func TestExtractBreakingChanges_endMarkerExtraWhitespace(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_end_marker_whitespace.md"))
	c, ok := ExtractBreakingChanges(sec.Raw)
	if !ok || !strings.Contains(c, "Content before marker.") || strings.Contains(c, "Content after marker.") {
		t.Fatalf("got %q ok=%v", c, ok)
	}
}

func TestExtractBreakingChanges_endMarkerIgnoredInsideBacktickFence(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_end_marker_fence_backtick.md"))
	c, ok := ExtractBreakingChanges(sec.Raw)
	if !ok || !strings.Contains(c, "<!-- /breaking-changes -->") || !strings.Contains(c, "Content after fence.") {
		t.Fatalf("got %q ok=%v", c, ok)
	}
}

func TestExtractBreakingChanges_endMarkerIgnoredInsideTildeFence(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_end_marker_fence_tilde.md"))
	c, ok := ExtractBreakingChanges(sec.Raw)
	if !ok || !strings.Contains(c, "<!-- /breaking-changes -->") || !strings.Contains(c, "Content after fence.") {
		t.Fatalf("got %q ok=%v", c, ok)
	}
}

func TestExtractBreakingChanges_markerBeforeChangelogHeadingIgnored(t *testing.T) {
	body := fixtureBody(t, "parse_end_marker_before_heading.md")
	raw, okInner := extractChangelogSection(body)
	if !okInner {
		t.Fatal("expected changelog inner")
	}
	c, ok := ExtractBreakingChanges(raw)
	if !ok || !strings.Contains(c, "Actual breaking change content.") {
		t.Fatalf("got %q ok=%v", c, ok)
	}
}

func TestParse_tildeFenceBreakingSubsection(t *testing.T) {
	sec := mustParse(t, fixtureBody(t, "parse_tilde_fence_breaking.md"))
	full := sec
	if full.ImpactRaw != "fix" {
		t.Fatal(full.ImpactRaw)
	}
	br, ok := ExtractBreakingChanges(full.Raw)
	if !ok || !strings.Contains(br, "tilde-fenced block") {
		t.Fatalf("breaking: %q", br)
	}
}

func TestParse_tildeFenceGuardsChangelogTerminator(t *testing.T) {
	body := strings.Join([]string{
		"## Changelog",
		"Customer impact: fix",
		"Summary: Fix tilde fence edge case",
		"",
		"~~~",
		"## fake heading inside tilde block",
		"~~~",
		"",
		"## Real next section",
	}, "\n")
	sec := mustParse(t, body)
	if sec.ImpactRaw != "fix" || sec.Summary != "Fix tilde fence edge case" {
		t.Fatalf("%+v", sec)
	}
}

func TestValidate_returnsErrorWhenParsedNil_message(t *testing.T) {
	_, errs := ValidateChangelogSection(nil)
	if len(errs) != 1 || errs[0] != "No ## Changelog section found in PR body" {
		t.Fatalf("got %#v", errs)
	}
}

func TestValidateFull_ruleC_verbatimMessageFromSpec(t *testing.T) {
	sec := mustParse(t, trimStartFixture(bodyFixWithBreakingSection))
	ok, errs := ValidateChangelogSectionFull(&sec, ValidateOpts{})
	if ok {
		t.Fatal("expected invalid")
	}
	var found bool
	for _, e := range errs {
		if e == ruleCBreakingOnlyWhenBreakingImpactMsg {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("want exact rule C string, got %#v", errs)
	}
}

func simulateAuthoringGate(prBody string) (changelogPresent bool, changelogValid bool, validationErrs []string) {
	sec, err := Parse([]byte(prBody))
	if err != nil {
		if errors.Is(err, ErrNoChangelogSection) {
			return false, false, nil
		}
		return false, false, nil
	}
	valid, ve := ValidateChangelogSectionFull(&sec, ValidateOpts{})
	if valid {
		return true, true, nil
	}
	return true, false, ve
}

func TestAuthoring_gateNoChangelogLabelBodyStillNullParse(t *testing.T) {
	prBody := "## Description\n\nThis is an internal refactor.\n"
	_, err := Parse([]byte(prBody))
	if !errors.Is(err, ErrNoChangelogSection) {
		t.Fatal(err)
	}
}

func TestAuthoring_gateValidFixPasses(t *testing.T) {
	prBody := strings.Join([]string{
		"## Description",
		"",
		"This fixes a regression.",
		"",
		"## Changelog",
		"",
		"Customer impact: fix",
		"Summary: Correct handling of empty API responses",
		"",
	}, "\n")
	present, valid, errs := simulateAuthoringGate(prBody)
	if !present || !valid || len(errs) != 0 {
		t.Fatalf("%v %v %#v", present, valid, errs)
	}
}

func TestAuthoring_gateValidEnhancementPasses(t *testing.T) {
	prBody := strings.Join([]string{
		"## Changelog",
		"",
		"Customer impact: enhancement",
		"Summary: Add support for index lifecycle management",
		"",
	}, "\n")
	present, valid, _ := simulateAuthoringGate(prBody)
	if !present || !valid {
		t.Fatal(present, valid)
	}
}

func TestAuthoring_gateValidNonePasses(t *testing.T) {
	present, valid, _ := simulateAuthoringGate("## Changelog\n\nCustomer impact: none\n")
	if !present || !valid {
		t.Fatal(present, valid)
	}
}

func TestAuthoring_gateValidBreakingPasses(t *testing.T) {
	prBody := strings.Join([]string{
		"## Changelog",
		"",
		"Customer impact: breaking",
		"Summary: Remove deprecated attribute from elasticstack_kibana_slo",
		"",
		"### Breaking changes",
		"",
		"The `legacy_mode` attribute has been removed. Update your configs.",
		"",
	}, "\n")
	present, valid, errs := simulateAuthoringGate(prBody)
	if !present || !valid || len(errs) != 0 {
		t.Fatal(present, valid, errs)
	}
}

func TestAuthoring_gateInvalidImpactFails(t *testing.T) {
	prBody := "## Changelog\n\nCustomer impact: patch\nSummary: Some change\n"
	present, valid, errs := simulateAuthoringGate(prBody)
	if !present || valid || len(errs) == 0 {
		t.Fatal(present, valid, errs)
	}
	ok := errorsAny(errs, func(e string) bool { return strings.Contains(e, "patch") })
	if !ok {
		t.Fatal(errs)
	}
}

func TestAuthoring_gateFixWithoutSummaryFails(t *testing.T) {
	prBody := "## Changelog\n\nCustomer impact: fix\n"
	present, valid, errs := simulateAuthoringGate(prBody)
	if !present || valid {
		t.Fatal(present, valid)
	}
	if !errorsAny(errs, func(e string) bool { return strings.Contains(e, "Summary") }) {
		t.Fatal(errs)
	}
}

func TestAuthoring_gateBreakingWithoutBreakingSubsectionFails(t *testing.T) {
	prBody := "## Changelog\n\nCustomer impact: breaking\nSummary: Remove old API\n"
	present, valid, errs := simulateAuthoringGate(prBody)
	if !present || valid {
		t.Fatal(present, valid)
	}
	if !errorsAny(errs, func(e string) bool { return strings.Contains(e, "Breaking changes") }) {
		t.Fatal(errs)
	}
}

func TestAuthoring_gateEmptyBreakingHeadingFails(t *testing.T) {
	prBody := strings.Join([]string{
		"## Changelog",
		"",
		"Customer impact: breaking",
		"Summary: A breaking change",
		"",
		"### Breaking changes",
		"",
		"## Other section",
		"",
	}, "\n")
	present, valid, errs := simulateAuthoringGate(prBody)
	if !present || valid {
		t.Fatal(present, valid)
	}
	if !errorsAny(errs, func(e string) bool {
		return strings.Contains(e, "Breaking changes") && strings.Contains(e, "no content")
	}) {
		t.Fatal(errs)
	}
}

func TestAuthoring_gateMissingSectionNotPresent(t *testing.T) {
	prBody := "## Description\n\nNo changelog section here.\n\n## Notes\n\nSome notes.\n"
	present, valid, errs := simulateAuthoringGate(prBody)
	if present || valid || len(errs) != 0 {
		t.Fatal(present, valid, errs)
	}
}

func TestAuthoring_gateEmptyPRBodyMissing(t *testing.T) {
	present, valid, errs := simulateAuthoringGate("")
	if present || valid || len(errs) != 0 {
		t.Fatal(present, valid, errs)
	}
}
