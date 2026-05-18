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
	"errors"
	"regexp"
	"strings"
)

// ErrNoChangelogSection is returned by Parse when the PR body lacks a "## Changelog" heading.
var ErrNoChangelogSection = errors.New("no ## Changelog section found in PR body")

const changelogSectionNotFoundValidateMsg = "no ## Changelog section found in PR body"

var changelogHeadingStart = regexp.MustCompile(`^##\s+Changelog`)

// Terminator mirrors JS: /^##\s/.test(line) outside fences.
var level2HeadingAfterChangelogTerminator = regexp.MustCompile(`^##\s`)

var customerImpactRE = regexp.MustCompile(`(?m)^Customer impact:\s*(.+)$`)
var summaryRE = regexp.MustCompile(`(?m)^Summary:\s*(.+)$`)

// RuleCBreakingOnlyWhenBreakingImpactMsg is the verbatim Rule C diagnostic from the PR changelog authoring spec.
const RuleCBreakingOnlyWhenBreakingImpactMsg = "### Breaking changes section is only allowed when Customer impact: breaking; " +
	"change to Customer impact: breaking or remove the ### Breaking changes heading."

// Section is the canonical parsed representation of a PR body's ## Changelog section.
type Section struct {
	CustomerImpact CustomerImpact `json:"customer_impact"`

	ImpactPresent bool `json:"impact_present,omitempty"`

	// ImpactRaw mirrors parseChangelogSection's trimmed customerImpact string (empty means missing capture).
	ImpactRaw string `json:"impact_raw,omitempty"`

	Summary                string `json:"summary,omitempty"`
	BreakingChanges        string `json:"breaking_changes,omitempty"`
	BreakingHeadingPresent bool   `json:"breaking_heading_present"`

	Raw string `json:"raw"`
}

// ValidateOpts configures validateChangelogSectionFull parity checks.
//
// By default (zero value), JS parity treats rule C ("### Breaking changes is only allowed when Customer impact:
// breaking...") as enabled, matching `{ enforceBreakingImpactMatch: true }` in validateChangelogSectionFull.
//
// Set RelaxBreakingImpactMatch to replicate changelog-renderer.js release-time rendering, which disables that
// check via `{ enforceBreakingImpactMatch: false }`.
type ValidateOpts struct {
	RelaxBreakingImpactMatch bool
}

// Parse parses a PR body's ## Changelog section into Section (parseChangelogSectionFull parity).
//
// It returns ErrNoChangelogSection when ## Changelog is absent (including empty bodies).
//
// Parsing is intentionally non-throwing beyond missing sections; callers apply ValidateChangelogSection /
// ValidateChangelogSectionFull to mirror the JS layering.
func Parse(body []byte) (Section, error) {
	sec, err := parseChangelogSection(string(body))
	if err != nil {
		return Section{}, err
	}
	return sec, nil
}

func parseChangelogSection(body string) (Section, error) {
	raw, ok := extractChangelogSection(body)
	if !ok {
		return Section{}, ErrNoChangelogSection
	}

	sec := Section{Raw: raw}

	if matches := customerImpactRE.FindStringSubmatch(raw); matches != nil {
		sec.ImpactRaw = strings.TrimSpace(matches[1])
		sec.ImpactPresent = sec.ImpactRaw != ""
		if v, okImpact := ParseCustomerImpact(sec.ImpactRaw); okImpact {
			sec.CustomerImpact = v
		}
	}

	if matches := summaryRE.FindStringSubmatch(raw); matches != nil {
		sec.Summary = strings.TrimSpace(matches[1])
	}

	sec.BreakingHeadingPresent = BreakingHeadingPresent(raw)

	if br, okBreaking := ExtractBreakingChanges(raw); okBreaking {
		sec.BreakingChanges = br
	}

	return sec, nil
}

func extractChangelogSection(body string) (inner string, ok bool) {
	if body == "" {
		return "", false
	}

	lines := strings.Split(body, "\n")
	var inChangelog bool
	var fenceType byte // 0 none, '`' backticks, '~' tilde
	var innerLines []string

	for _, ls := range lines {
		line := ls

		if changelogHeadingStart.MatchString(strings.TrimSuffix(line, "\r")) {
			inChangelog = true
			continue
		}

		if !inChangelog {
			continue
		}

		switch {
		case fenceType == 0 && strings.HasPrefix(line, "```"):
			fenceType = '`'
		case fenceType == 0 && strings.HasPrefix(line, "~~~"):
			fenceType = '~'
		case fenceType == '`' && strings.HasPrefix(line, "```"):
			fenceType = 0
		case fenceType == '~' && strings.HasPrefix(line, "~~~"):
			fenceType = 0
		}

		if fenceType == 0 && level2HeadingAfterChangelogTerminator.MatchString(line) {
			break
		}

		innerLines = append(innerLines, line)
	}

	if !inChangelog {
		return "", false
	}

	return strings.Join(innerLines, "\n"), true
}

// ValidateChangelogSection mirrors validateChangelogSection from pr-changelog-parser.js.
func ValidateChangelogSection(parsed *Section) (bool, []string) {
	errs := validateChangelogSection(parsed)
	return len(errs) == 0, errs
}

func validateChangelogSection(parsed *Section) []string {
	errs := make([]string, 0)
	if parsed == nil {
		return []string{changelogSectionNotFoundValidateMsg}
	}

	impact := parsed.ImpactRaw
	if strings.TrimSpace(impact) == "" {
		errs = append(errs, "Missing required field: Customer impact")
		return errs
	}

	if _, ok := ParseCustomerImpact(impact); !ok {
		errs = append(errs, `Invalid Customer impact value: "`+impact+`". Must be one of: `+strings.Join(customerImpactIDs(), ", "))
		return errs
	}

	// Summary parity: treat missing Summary line OR whitespace-only capture as absent (requires summary whenever impact != none).
	if impact != impactLiteralNone && parsed.Summary == "" {
		errs = append(errs, `Missing required field: Summary (required when Customer impact is not "none")`)
	}

	return errs
}

// ValidateChangelogSectionFull mirrors validateChangelogSectionFull from pr-changelog-parser.js.
func ValidateChangelogSectionFull(parsed *Section, opts ValidateOpts) (bool, []string) {
	base := validateChangelogSection(parsed)
	if parsed == nil {
		return len(base) == 0, base
	}

	errorsOut := append([]string{}, base...)
	enforceBreakingImpactMatch := !opts.RelaxBreakingImpactMatch

	if parsed.BreakingHeadingPresent && parsed.BreakingChanges == "" {
		errorsOut = append(errorsOut, "### Breaking changes section is present but contains no content")
	}

	if parsed.ImpactRaw == impactLiteralBreaking && !parsed.BreakingHeadingPresent {
		errorsOut = append(errorsOut, "Customer impact: breaking requires a ### Breaking changes subsection")
	}

	if enforceBreakingImpactMatch &&
		parsed.BreakingHeadingPresent &&
		parsed.ImpactRaw != "" &&
		ParsedImpactIsKnownContractValue(parsed.ImpactRaw) &&
		parsed.ImpactRaw != impactLiteralBreaking {

		errorsOut = append(errorsOut, RuleCBreakingOnlyWhenBreakingImpactMsg)
	}

	return len(errorsOut) == 0, errorsOut
}

// ParsedImpactIsKnownContractValue mirrors VALID_CUSTOMER_IMPACTS.has(customerImpact) in JS without mapping to enums.
func ParsedImpactIsKnownContractValue(trimmedImpact string) bool {
	_, ok := ParseCustomerImpact(trimmedImpact)
	return ok
}
