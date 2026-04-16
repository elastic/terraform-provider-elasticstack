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
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Core parsing helpers
// ---------------------------------------------------------------------------

// changelogSection represents a ## section in CHANGELOG.md.
type changelogSection struct {
	header string
	body   string
	// bodySet indicates that the body field was explicitly set (even if empty string).
	// This mirrors the JS behaviour where `body !== undefined`.
	bodySet bool
}

// parsedChangelog is the result of parsing CHANGELOG.md into logical blocks.
type parsedChangelog struct {
	preamble string
	sections []changelogSection
	footer   string
}

// headerUnreleased is the standard Unreleased section header in CHANGELOG.md.
const headerUnreleased = "## [Unreleased]"

// linkDefRE matches a link definition line: [label]: https://...
var linkDefRE = regexp.MustCompile(`^\[.+\]:\s*https?://`)

// parseChangelog splits CHANGELOG.md content into preamble, sections, and footer.
func parseChangelog(content string) parsedChangelog {
	lines := strings.Split(content, "\n")
	var sections []changelogSection
	preamble := ""

	var currentHeader string
	var currentBodyLines []string
	inSections := false
	hasHeader := false

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			if hasHeader {
				sections = append(sections, changelogSection{
					header:  currentHeader,
					body:    strings.Join(currentBodyLines, "\n"),
					bodySet: true,
				})
			} else if !inSections {
				// Everything before the first ## section is preamble
				preamble = strings.Join(currentBodyLines, "\n")
			}
			currentHeader = line
			currentBodyLines = nil
			inSections = true
			hasHeader = true
		} else {
			currentBodyLines = append(currentBodyLines, line)
		}
	}

	if hasHeader {
		sections = append(sections, changelogSection{
			header:  currentHeader,
			body:    strings.Join(currentBodyLines, "\n"),
			bodySet: true,
		})
	} else {
		// No sections found at all
		preamble = strings.Join(currentBodyLines, "\n")
	}

	// The footer is the trailing block of link definitions after the last section.
	// Link definitions look like: [label]: https://...
	// We extract them from the last section's body tail.
	footer := ""
	if len(sections) > 0 {
		lastSection := &sections[len(sections)-1]
		bodyLines := strings.Split(lastSection.body, "\n")

		// Find where the link footer starts: scan from end, stop at first non-empty non-link line.
		footerStart := -1
		for i, v := range slices.Backward(bodyLines) {
			if linkDefRE.MatchString(v) {
				footerStart = i
			} else if strings.TrimSpace(v) != "" {
				break
			}
		}

		if footerStart != -1 {
			footer = strings.Join(bodyLines[footerStart:], "\n")
			lastSection.body = strings.Join(bodyLines[:footerStart], "\n")
		}
	}

	return parsedChangelog{preamble: preamble, sections: sections, footer: footer}
}

// serialiseChangelog converts a parsedChangelog back to a string.
func serialiseChangelog(parsed parsedChangelog) string {
	var parts []string

	if strings.TrimSpace(parsed.preamble) != "" {
		parts = append(parts, parsed.preamble)
	}

	for _, section := range parsed.sections {
		parts = append(parts, section.header)
		if section.bodySet {
			parts = append(parts, section.body)
		}
	}

	if strings.TrimSpace(parsed.footer) != "" {
		parts = append(parts, parsed.footer)
	}

	return strings.Join(parts, "\n")
}

// normaliseSectionBody normalises new section body content:
//   - Strips leading blank lines
//   - Trims trailing whitespace/newlines
//   - Adds a leading "\n" and trailing "\n"
func normaliseSectionBody(content string) string {
	trimmed := strings.TrimRight(regexp.MustCompile(`^\n+`).ReplaceAllString(content, ""), " \t\n\r")
	return "\n" + trimmed + "\n"
}

// ---------------------------------------------------------------------------
// Rewrite operations
// ---------------------------------------------------------------------------

// rewriteUnreleased replaces the ## [Unreleased] section body.
func rewriteUnreleased(changelogContent, newBody string) (string, error) {
	parsed := parseChangelog(changelogContent)

	idx := -1
	for i, s := range parsed.sections {
		if s.header == headerUnreleased {
			idx = i
			break
		}
	}
	if idx == -1 {
		return "", fmt.Errorf("## [Unreleased] section not found in CHANGELOG.md")
	}

	parsed.sections[idx].body = normaliseSectionBody(newBody)
	parsed.sections[idx].bodySet = true

	return serialiseChangelog(parsed), nil
}

// rewriteRelease replaces or inserts the ## [x.y.z] - YYYY-MM-DD section.
// The ## [Unreleased] section is left unchanged.
func rewriteRelease(changelogContent, targetVersion, releaseDate, newBody string) string {
	parsed := parseChangelog(changelogContent)

	releaseHeader := fmt.Sprintf("## [%s] - %s", targetVersion, releaseDate)
	headerPrefix := fmt.Sprintf("## [%s]", targetVersion)

	// Find existing release section (match by version prefix, ignoring date)
	releaseIdx := -1
	for i, s := range parsed.sections {
		if strings.HasPrefix(s.header, headerPrefix) {
			releaseIdx = i
			break
		}
	}

	if releaseIdx == -1 {
		// Insert after ## [Unreleased]
		unreleasedIdx := -1
		for i, s := range parsed.sections {
			if s.header == headerUnreleased {
				unreleasedIdx = i
				break
			}
		}

		newSection := changelogSection{
			header:  releaseHeader,
			body:    normaliseSectionBody(newBody),
			bodySet: true,
		}

		if unreleasedIdx != -1 {
			// Insert after unreleased
			newSections := make([]changelogSection, 0, len(parsed.sections)+1)
			newSections = append(newSections, parsed.sections[:unreleasedIdx+1]...)
			newSections = append(newSections, newSection)
			newSections = append(newSections, parsed.sections[unreleasedIdx+1:]...)
			parsed.sections = newSections
		} else {
			// Prepend
			parsed.sections = append([]changelogSection{newSection}, parsed.sections...)
		}
	} else {
		parsed.sections[releaseIdx].header = releaseHeader
		parsed.sections[releaseIdx].body = normaliseSectionBody(newBody)
		parsed.sections[releaseIdx].bodySet = true
	}

	return serialiseChangelog(parsed)
}

// rewriteSectionResult is the JSON output of the rewrite operation.
type rewriteSectionResult struct {
	Success       bool   `json:"success"`
	ChangelogPath string `json:"changelogPath"`
	Section       string `json:"section"`
	Error         string `json:"error,omitempty"`
}

// rewriteSection is the unified entry point: rewrites the appropriate section based on mode.
func rewriteSection(changelogPath, mode, targetVersion, releaseDate, newSectionBody string) (rewriteSectionResult, error) {
	absPath, err := filepath.Abs(changelogPath)
	if err != nil {
		return rewriteSectionResult{}, err
	}

	changelogContent, err := os.ReadFile(absPath)
	if err != nil {
		return rewriteSectionResult{}, err
	}

	date := releaseDate
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	var updated string
	var sectionHeader string

	if mode == "release" {
		if targetVersion == "" {
			return rewriteSectionResult{}, fmt.Errorf("targetVersion is required when mode == \"release\"")
		}
		updated = rewriteRelease(string(changelogContent), targetVersion, date, newSectionBody)
		sectionHeader = fmt.Sprintf("## [%s] - %s", targetVersion, date)
	} else {
		updated, err = rewriteUnreleased(string(changelogContent), newSectionBody)
		if err != nil {
			return rewriteSectionResult{}, err
		}
		sectionHeader = headerUnreleased
	}

	if err := os.WriteFile(absPath, []byte(updated), 0644); err != nil {
		return rewriteSectionResult{}, err
	}

	return rewriteSectionResult{Success: true, ChangelogPath: absPath, Section: sectionHeader}, nil
}

// ---------------------------------------------------------------------------
// CLI entry point
// ---------------------------------------------------------------------------

func runRewriteChangelogSection(args []string) {
	argMap := parseArgMap(args)

	mode := argMapGet(argMap, "mode", os.Getenv("MODE"), "unreleased")
	targetVersion := argMapGet(argMap, "target-version", os.Getenv("TARGET_VERSION"), "")
	changelogPath := argMapGet(argMap, "changelog", os.Getenv("CHANGELOG_PATH"), "CHANGELOG.md")
	sectionFile := argMapGet(argMap, "section-file", os.Getenv("SECTION_FILE"), "")

	newSectionBody := argMapGet(argMap, "section-content", os.Getenv("SECTION_CONTENT"), "")

	writeResult := func(r rewriteSectionResult) {
		out, _ := json.MarshalIndent(r, "", "  ")
		fmt.Println(string(out))
	}

	if newSectionBody == "" && sectionFile != "" {
		raw, err := os.ReadFile(sectionFile)
		if err != nil {
			r := rewriteSectionResult{
				Success:       false,
				ChangelogPath: changelogPath,
				Section:       "",
				Error:         fmt.Sprintf("Cannot read section file: %s", err.Error()),
			}
			writeResult(r)
			os.Exit(1)
		}
		newSectionBody = string(raw)
	}

	if newSectionBody == "" {
		r := rewriteSectionResult{
			Success:       false,
			ChangelogPath: changelogPath,
			Section:       "",
			Error:         "No section content provided (use --section-content or --section-file)",
		}
		writeResult(r)
		os.Exit(1)
	}

	result, err := rewriteSection(changelogPath, mode, targetVersion, "", newSectionBody)
	if err != nil {
		r := rewriteSectionResult{
			Success:       false,
			ChangelogPath: changelogPath,
			Section:       "",
			Error:         err.Error(),
		}
		writeResult(r)
		os.Exit(1)
	}

	writeResult(result)
}
