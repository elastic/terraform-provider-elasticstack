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
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------
// Regex patterns
// ---------------------------------------------------------------------------

// commitSHARE matches a 40-character hex SHA (commit hash).
var commitSHARE = regexp.MustCompile(`(?i)\b[0-9a-f]{40}\b`)

// shortSHARE matches a short (7-12 char) hex SHA.
var shortSHARE = regexp.MustCompile(`(?i)\b[0-9a-f]{7,12}\b`)

// prRefRE matches #NNN not preceded by / (which would be a URL path segment).
// We simulate the negative lookbehind by extracting the full match and checking the preceding byte.
var prRefRE = regexp.MustCompile(`(^|[^/])#(\d+)`)

// ---------------------------------------------------------------------------
// Core helper functions
// ---------------------------------------------------------------------------

// extractPRReferences extracts all #NNN PR number references from markdown text,
// excluding references that appear after a / (URL paths).
func extractPRReferences(text string) []int {
	matches := prRefRE.FindAllStringSubmatch(text, -1)
	seen := map[int]bool{}
	var result []int
	for _, m := range matches {
		n, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		if !seen[n] {
			seen[n] = true
			result = append(result, n)
		}
	}
	return result
}

// extractBulletLines extracts bullet lines from a changelog section.
var bulletLineRE = regexp.MustCompile(`^\s*[*-]\s`)

func extractBulletLines(sectionContent string) []string {
	var bullets []string
	for line := range strings.SplitSeq(sectionContent, "\n") {
		if bulletLineRE.MatchString(line) {
			bullets = append(bullets, line)
		}
	}
	return bullets
}

// looksLikeCommitNarration returns true if a line appears to reference a commit SHA.
var prNumInlineRE = regexp.MustCompile(`#\d+`)

func looksLikeCommitNarration(line string) bool {
	if commitSHARE.MatchString(line) {
		return true
	}
	// Short hex SHA without PR reference is suspicious
	if shortSHARE.MatchString(line) && !prNumInlineRE.MatchString(line) {
		return true
	}
	return false
}

// extractSectionFromChangelog extracts a named section from CHANGELOG.md content.
// Returns the content from the section header up to (but not including) the next ## header.
// Uses exact prefix matching: header must equal the line, or line starts with header+" " or header+"\t".
func extractSectionFromChangelog(changelogContent, header string) (string, bool) {
	lines := strings.Split(changelogContent, "\n")
	inSection := false
	var sectionLines []string

	headerMatches := func(line string) bool {
		return line == header ||
			strings.HasPrefix(line, header+" ") ||
			strings.HasPrefix(line, header+"\t")
	}

	for _, line := range lines {
		if !inSection {
			if headerMatches(line) {
				inSection = true
				sectionLines = append(sectionLines, line)
			}
		} else {
			// Stop at the next ## header (but not the same one)
			if strings.HasPrefix(line, "## ") && !headerMatches(line) {
				break
			}
			sectionLines = append(sectionLines, line)
		}
	}

	if !inSection {
		return "", false
	}
	return strings.Join(sectionLines, "\n"), true
}

// ---------------------------------------------------------------------------
// JSON input types
// ---------------------------------------------------------------------------

type evidenceManifest struct {
	TargetSection string       `json:"target_section"`
	PullRequests  []evidencePR `json:"pull_requests"`
}

type evidencePR struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

type provenanceFile struct {
	Bullets []provenanceBullet `json:"bullets"`
}

type provenanceBullet struct {
	Text      string `json:"text"`
	PRNumbers []int  `json:"pr_numbers"`
}

// ---------------------------------------------------------------------------
// Core validation
// ---------------------------------------------------------------------------

type validationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

type validateProvenanceParams struct {
	evidence         *evidenceManifest
	provenance       *provenanceFile
	changelogSection string
	expectedHeader   string
}

func validateProvenance(p validateProvenanceParams) validationResult {
	var errors []string
	var warnings []string

	// Build a set of known PR numbers from the evidence manifest
	evidencePRNumbers := map[int]bool{}
	for _, pr := range p.evidence.PullRequests {
		evidencePRNumbers[pr.Number] = true
	}

	// --- Check 1: provenance bullets map to known PRs ---
	bullets := p.provenance.Bullets
	for _, bullet := range bullets {
		if len(bullet.PRNumbers) == 0 {
			text := bullet.Text
			if text == "" {
				text = "(no text)"
			}
			errors = append(errors, fmt.Sprintf("Provenance bullet has no pr_numbers: %q", text))
		}
		for _, prNum := range bullet.PRNumbers {
			if !evidencePRNumbers[prNum] {
				errors = append(errors, fmt.Sprintf(
					"Provenance bullet references PR #%d which is NOT in the evidence manifest: %q",
					prNum, bullet.Text,
				))
			}
		}
	}

	// --- Check 2: every #NNN in the changelog markdown is backed by evidence ---
	changelogPRRefs := extractPRReferences(p.changelogSection)
	for _, prNum := range changelogPRRefs {
		if !evidencePRNumbers[prNum] {
			errors = append(errors, fmt.Sprintf(
				"Changelog references PR #%d which is NOT in the evidence manifest (fabricated reference?)",
				prNum,
			))
		}
	}

	// --- Check 3: no commit-level narration ---
	bulletLines := extractBulletLines(p.changelogSection)
	for _, line := range bulletLines {
		if looksLikeCommitNarration(line) {
			errors = append(errors, fmt.Sprintf(
				"Changelog bullet appears to contain commit-level narration (SHA found): %q",
				strings.TrimSpace(line),
			))
		}
		// Reject bullets without any PR reference
		if !prNumInlineRE.MatchString(line) {
			errors = append(errors, fmt.Sprintf(
				"Changelog bullet has no PR reference (#NNN): %q",
				strings.TrimSpace(line),
			))
		}
	}

	// --- Check 4: section header matches expected ---
	resolvedExpectedHeader := p.expectedHeader
	if resolvedExpectedHeader == "" {
		resolvedExpectedHeader = p.evidence.TargetSection
	}
	if resolvedExpectedHeader != "" {
		lines := strings.Split(p.changelogSection, "\n")
		var firstNonEmpty string
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				firstNonEmpty = l
				break
			}
		}
		if firstNonEmpty != "" && strings.TrimSpace(firstNonEmpty) != resolvedExpectedHeader {
			errors = append(errors, fmt.Sprintf(
				"Changelog section header %q does not match expected %q",
				strings.TrimSpace(firstNonEmpty),
				resolvedExpectedHeader,
			))
		}
	}

	valid := len(errors) == 0
	// Return non-nil slices for clean JSON output
	if errors == nil {
		errors = []string{}
	}
	if warnings == nil {
		warnings = []string{}
	}
	return validationResult{Valid: valid, Errors: errors, Warnings: warnings}
}

// ---------------------------------------------------------------------------
// CLI entry point
// ---------------------------------------------------------------------------

func runValidateProvenance(args []string) {
	// Parse --key value pairs
	argMap := parseArgMap(args)

	evidencePath := argMapGet(argMap, "evidence", os.Getenv("EVIDENCE_PATH"), "/tmp/gh-aw/agent/evidence.json")
	provenancePath := argMapGet(argMap, "provenance", os.Getenv("PROVENANCE_PATH"), "/tmp/gh-aw/agent/provenance.json")
	changelogPath := argMapGet(argMap, "changelog", os.Getenv("CHANGELOG_PATH"), "CHANGELOG.md")
	sectionHeaderArg := argMapGet(argMap, "section-header", os.Getenv("SECTION_HEADER"), "")

	writeResult := func(r validationResult) {
		out, _ := json.MarshalIndent(r, "", "  ")
		fmt.Println(string(out))
	}

	evidenceRaw, err := os.ReadFile(evidencePath)
	if err != nil {
		r := validationResult{Valid: false, Errors: []string{fmt.Sprintf("Cannot read evidence manifest: %s", err.Error())}, Warnings: []string{}}
		writeResult(r)
		os.Exit(1)
	}
	var evidence evidenceManifest
	if err := json.Unmarshal(evidenceRaw, &evidence); err != nil {
		r := validationResult{Valid: false, Errors: []string{fmt.Sprintf("Cannot parse evidence manifest: %s", err.Error())}, Warnings: []string{}}
		writeResult(r)
		os.Exit(1)
	}

	provenanceRaw, err := os.ReadFile(provenancePath)
	if err != nil {
		r := validationResult{Valid: false, Errors: []string{fmt.Sprintf("Cannot read provenance file: %s", err.Error())}, Warnings: []string{}}
		writeResult(r)
		os.Exit(1)
	}
	var provenance provenanceFile
	if err := json.Unmarshal(provenanceRaw, &provenance); err != nil {
		r := validationResult{Valid: false, Errors: []string{fmt.Sprintf("Cannot parse provenance file: %s", err.Error())}, Warnings: []string{}}
		writeResult(r)
		os.Exit(1)
	}

	changelogContent, err := os.ReadFile(changelogPath)
	if err != nil {
		r := validationResult{Valid: false, Errors: []string{fmt.Sprintf("Cannot read CHANGELOG.md: %s", err.Error())}, Warnings: []string{}}
		writeResult(r)
		os.Exit(1)
	}

	targetHeader := sectionHeaderArg
	if targetHeader == "" {
		targetHeader = evidence.TargetSection
	}
	if targetHeader == "" {
		targetHeader = headerUnreleased
	}

	// Extract just the section heading prefix for matching (strip date part for lookup)
	headerPrefix := strings.SplitN(targetHeader, " - ", 2)[0]
	changelogSection, found := extractSectionFromChangelog(string(changelogContent), headerPrefix)
	if !found {
		r := validationResult{
			Valid:    false,
			Errors:   []string{fmt.Sprintf("Section %q not found in %s", targetHeader, changelogPath)},
			Warnings: []string{},
		}
		writeResult(r)
		os.Exit(1)
	}

	result := validateProvenance(validateProvenanceParams{
		evidence:         &evidence,
		provenance:       &provenance,
		changelogSection: changelogSection,
		expectedHeader:   targetHeader,
	})

	writeResult(result)

	if !result.Valid {
		fmt.Fprintf(os.Stderr, "Validation failed: %d error(s)\n", len(result.Errors))
		os.Exit(1)
	}

	if len(result.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "Validation passed (%d warning(s))\n", len(result.Warnings))
	} else {
		fmt.Fprintln(os.Stderr, "Validation passed")
	}
}
