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
	"errors"
	"fmt"
	"strings"
)

const (
	unreleasedHeadingPrefix = "## [Unreleased]"
	sectionHeadingPrefix    = "## "
)

// errSectionBodyContainsH2Heading is returned when Body includes a Markdown H2 heading
// (“## …”), which RewriteSection forbids — the rewriter emits the canonical heading itself.
var errSectionBodyContainsH2Heading = errors.New(
	"rewriter: section body must not contain a level-2 markdown heading",
)

// RewriteMode selects which CHANGELOG section heading rewriteChangelogSection targets.
type RewriteMode int

const (
	// ModeUnreleased updates the ## [Unreleased] section.
	ModeUnreleased RewriteMode = iota
	// ModeRelease updates by target version (and may remove Unreleased in the same pass).
	ModeRelease
)

// SectionRewrite describes a CHANGELOG release or Unreleased section to write in place.
type SectionRewrite struct {
	// Header is the literal heading without the leading "## ", e.g.
	// "[Unreleased]" or "[1.2.3] - 2025-05-01".
	Header string
	// Body is the Markdown rendered under Header. It MUST NOT contain any "## …" headings;
	// use "### …" subsections only. Embedding "## …" in Body makes RewriteSection return an error.
	Body string
}

func (r SectionRewrite) fullSectionMarkdown() string {
	return sectionHeadingPrefix + r.Header + "\n" + r.Body
}

// FindSectionEnd returns the exclusive index of the first ## line after startIndex,
// or len(lines) if none — matching changelog-rewriter.js findSectionEnd.
func FindSectionEnd(lines []string, startIndex int) int {
	for i := startIndex + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], sectionHeadingPrefix) {
			return i
		}
	}
	return len(lines)
}

// RewriteSection mirrors rewriteChangelogSection from changelog-rewriter.js.
//
// When mode is ModeRelease, targetVersion is the semver without a leading 'v';
// otherwise it is ignored.
//
// RewriteSection returns an error when rewrite.Body introduces an H2 Markdown heading or mode is unrecognized.
func RewriteSection(content []byte, rewrite SectionRewrite, mode RewriteMode, targetVersion string) ([]byte, error) {
	if err := validateSectionRewriteBody(rewrite.Body); err != nil {
		return nil, err
	}
	newSectionContent := rewrite.fullSectionMarkdown()
	lines := strings.Split(string(content), "\n")

	var targetStart int
	switch mode {
	case ModeUnreleased:
		targetStart = indexOfHeading(lines, unreleasedHeadingMatch)
	case ModeRelease:
		prefix := sectionHeadingPrefix + "[" + targetVersion + "]"
		targetStart = indexOfLinePrefix(lines, prefix)
	default:
		return nil, fmt.Errorf("rewriter: unknown rewrite mode %d", mode)
	}

	if mode == ModeRelease {
		return []byte(rewriteRelease(lines, newSectionContent, targetStart, string(content))), nil
	}

	if targetStart == -1 {
		return []byte(newSectionContent + "\n\n" + string(content)), nil
	}

	sectionEnd := FindSectionEnd(lines, targetStart)
	return []byte(spliceSingleSection(lines, targetStart, sectionEnd, newSectionContent)), nil
}

func validateSectionRewriteBody(body string) error {
	switch {
	case strings.HasPrefix(body, "## "):
		return errSectionBodyContainsH2Heading
	case strings.Contains(body, "\n## "):
		return errSectionBodyContainsH2Heading
	default:
		return nil
	}
}

func unreleasedHeadingMatch(line string) bool {
	return strings.HasPrefix(line, unreleasedHeadingPrefix)
}

func indexOfHeading(lines []string, match func(string) bool) int {
	for i, line := range lines {
		if match(line) {
			return i
		}
	}
	return -1
}

func indexOfLinePrefix(lines []string, prefix string) int {
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return i
		}
	}
	return -1
}

func rewriteRelease(lines []string, newSectionContent string, targetStart int, original string) string {
	unreleasedStart := indexOfHeading(lines, unreleasedHeadingMatch)
	var ranges [][2]int
	if unreleasedStart != -1 {
		ranges = append(ranges, [2]int{unreleasedStart, FindSectionEnd(lines, unreleasedStart)})
	}
	if targetStart != -1 {
		ranges = append(ranges, [2]int{targetStart, FindSectionEnd(lines, targetStart)})
	}
	ranges = sortRangesByStart(ranges)

	if len(ranges) == 0 {
		return newSectionContent + "\n\n" + original
	}

	return spliceReleaseSectionRanges(lines, ranges, newSectionContent)
}

func sortRangesByStart(ranges [][2]int) [][2]int {
	// Insertion sort — tiny n (≤2 in practice).
	for i := 1; i < len(ranges); i++ {
		for j := i; j > 0 && ranges[j-1][0] > ranges[j][0]; j-- {
			ranges[j-1], ranges[j] = ranges[j], ranges[j-1]
		}
	}
	return ranges
}

func spliceReleaseSectionRanges(lines []string, ranges [][2]int, newSectionContent string) string {
	first := ranges[0]
	before := lines[:first[0]]
	before = trimTrailingEmptyLines(before)

	parts := append([]string{}, before...)
	parts = appendNonemptySeparator(parts)
	parts = append(parts, newSectionContent)

	cursor := first[1]
	for i := 1; i < len(ranges); i++ {
		r := ranges[i]
		parts = append(parts, lines[cursor:r[0]]...)
		cursor = r[1]
	}

	parts = appendRemainderAfterSkippingLeadingBlankLines(parts, lines[cursor:])
	return strings.Join(parts, "\n")
}

func spliceSingleSection(lines []string, targetStart, sectionEnd int, newSectionContent string) string {
	before := lines[:targetStart]
	before = trimTrailingEmptyLines(before)

	parts := append([]string{}, before...)
	parts = appendNonemptySeparator(parts)
	parts = append(parts, newSectionContent)

	parts = appendRemainderAfterSkippingLeadingBlankLines(parts, lines[sectionEnd:])
	return strings.Join(parts, "\n")
}

func trimTrailingEmptyLines(lines []string) []string {
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func appendNonemptySeparator(parts []string) []string {
	if len(parts) > 0 {
		parts = append(parts, "")
	}
	return parts
}

func appendRemainderAfterSkippingLeadingBlankLines(parts []string, after []string) []string {
	start := skipLeadingEmptyLines(after)
	if start < len(after) {
		parts = append(parts, "")
		parts = append(parts, after[start:]...)
	}
	return parts
}

func skipLeadingEmptyLines(lines []string) int {
	i := 0
	for i < len(lines) && lines[i] == "" {
		i++
	}
	return i
}
