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
	"fmt"
	"regexp"
	"strings"
)

var (
	validCustomerImpacts = map[string]bool{
		"none":        true,
		"fix":         true,
		"enhancement": true,
		"breaking":    true,
	}
	changelogHeadingRegexp   = regexp.MustCompile(`^##\s+Changelog`)
	sectionHeadingRegexp     = regexp.MustCompile(`^##\s`)
	breakingHeadingRegexp    = regexp.MustCompile(`^###\s+Breaking changes`)
	subsectionHeadingRegexp  = regexp.MustCompile(`^#{2,3}\s`)
	customerImpactLineRegexp = regexp.MustCompile(`^Customer impact:\s*(.+)$`)
	summaryLineRegexp        = regexp.MustCompile(`^Summary:\s*(.+)$`)
)

type parsedChangelogSection struct {
	CustomerImpact                string
	Summary                       string
	BreakingChanges               string
	BreakingChangesHeadingPresent bool
}

func extractChangelogSection(body string) string {
	if body == "" {
		return ""
	}
	lines := strings.Split(body, "\n")
	inChangelog := false
	fenceType := ""
	content := make([]string, 0)

	for _, line := range lines {
		if changelogHeadingRegexp.MatchString(line) {
			inChangelog = true
			continue
		}
		if !inChangelog {
			continue
		}
		switch {
		case fenceType == "" && strings.HasPrefix(line, "```"):
			fenceType = "`"
		case fenceType == "" && strings.HasPrefix(line, "~~~"):
			fenceType = "~"
		case fenceType == "`" && strings.HasPrefix(line, "```"):
			fenceType = ""
		case fenceType == "~" && strings.HasPrefix(line, "~~~"):
			fenceType = ""
		}
		if fenceType == "" && sectionHeadingRegexp.MatchString(line) {
			break
		}
		content = append(content, line)
	}

	if !inChangelog {
		return ""
	}
	return strings.Join(content, "\n")
}

func extractBreakingChanges(section string) string {
	if section == "" {
		return ""
	}
	lines := strings.Split(section, "\n")
	inBreaking := false
	fenceType := ""
	content := make([]string, 0)

	for _, line := range lines {
		if breakingHeadingRegexp.MatchString(line) {
			inBreaking = true
			continue
		}
		if !inBreaking {
			continue
		}
		switch {
		case fenceType == "" && strings.HasPrefix(line, "```"):
			fenceType = "`"
		case fenceType == "" && strings.HasPrefix(line, "~~~"):
			fenceType = "~"
		case fenceType == "`" && strings.HasPrefix(line, "```"):
			fenceType = ""
		case fenceType == "~" && strings.HasPrefix(line, "~~~"):
			fenceType = ""
		}
		if fenceType == "" && subsectionHeadingRegexp.MatchString(line) {
			break
		}
		content = append(content, line)
	}

	if !inBreaking {
		return ""
	}
	return strings.TrimRight(strings.Join(content, "\n"), "\n")
}

func parseChangelogSectionFull(body string) *parsedChangelogSection {
	section := extractChangelogSection(body)
	if section == "" {
		return nil
	}

	customerImpact := ""
	summary := ""
	for _, line := range strings.Split(section, "\n") {
		if matches := customerImpactLineRegexp.FindStringSubmatch(line); len(matches) == 2 {
			customerImpact = strings.TrimSpace(matches[1])
		}
		if matches := summaryLineRegexp.FindStringSubmatch(line); len(matches) == 2 {
			summary = strings.TrimSpace(matches[1])
		}
	}

	headingPresent := false
	for _, line := range strings.Split(section, "\n") {
		if breakingHeadingRegexp.MatchString(line) {
			headingPresent = true
			break
		}
	}

	return &parsedChangelogSection{
		CustomerImpact:                customerImpact,
		Summary:                       summary,
		BreakingChanges:               extractBreakingChanges(section),
		BreakingChangesHeadingPresent: headingPresent,
	}
}

func validateChangelogSectionFull(parsed *parsedChangelogSection) []string {
	if parsed == nil {
		return []string{"No ## Changelog section found in PR body"}
	}
	var errs []string
	if parsed.CustomerImpact == "" {
		errs = append(errs, "Missing required field: Customer impact")
	} else if !validCustomerImpacts[parsed.CustomerImpact] {
		errs = append(errs, fmt.Sprintf("Invalid Customer impact value: %q", parsed.CustomerImpact))
	}
	if parsed.CustomerImpact != "" && parsed.CustomerImpact != "none" && validCustomerImpacts[parsed.CustomerImpact] && parsed.Summary == "" {
		errs = append(errs, "Missing required field: Summary (required when Customer impact is not \"none\")")
	}
	if parsed.BreakingChangesHeadingPresent && parsed.BreakingChanges == "" {
		errs = append(errs, "### Breaking changes section is present but contains no content")
	}
	if parsed.CustomerImpact == "breaking" && !parsed.BreakingChangesHeadingPresent {
		errs = append(errs, "Customer impact: breaking requires a ### Breaking changes subsection")
	}
	return errs
}
