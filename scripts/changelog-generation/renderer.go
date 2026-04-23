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
	"slices"
	"strings"
)

const noChangelogLabel = "no-changelog"

type pullRequestRecord struct {
	Number         int
	Title          string
	URL            string
	MergeCommitSHA string
	Author         string
	Labels         []string
	Body           string
}

type includedPR struct {
	PRNumber        int
	PRURL           string
	Summary         string
	BreakingChanges string
}

type excludedPR struct {
	PRNumber        int
	PRURL           string
	Reason          string
	BreakingChanges string
}

type assemblyError struct {
	PRNumber int
	PRURL    string
	Reason   string
}

type renderResult struct {
	Success     bool
	SectionBody string
	Errors      []assemblyError
	Included    []includedPR
	Excluded    []excludedPR
}

func normalizeBulletPrefix(line string) string {
	trimmed := strings.TrimLeft(line, " ")
	trimmed = strings.TrimPrefix(trimmed, "-")
	trimmed = strings.TrimPrefix(trimmed, "*")
	trimmed = strings.TrimPrefix(trimmed, "+")
	trimmed = strings.TrimLeft(trimmed, " ")
	return "- " + trimmed
}

func buildCitation(prNumber int, prURL string) string {
	return fmt.Sprintf("([#%d](%s))", prNumber, prURL)
}

func buildChangeBullet(summary string, prNumber int, prURL string) string {
	return fmt.Sprintf("%s %s", normalizeBulletPrefix(strings.TrimSpace(summary)), buildCitation(prNumber, prURL))
}

func hasLabel(labels []string, target string) bool {
	return slices.Contains(labels, target)
}

func renderChangelogSection(mergedPRs []pullRequestRecord) renderResult {
	result := renderResult{
		Success:  true,
		Errors:   []assemblyError{},
		Included: []includedPR{},
		Excluded: []excludedPR{},
	}
	changeBullets := []string{}
	breakingChangeBlocks := []string{}

	for _, pr := range mergedPRs {
		if hasLabel(pr.Labels, noChangelogLabel) {
			result.Excluded = append(result.Excluded, excludedPR{PRNumber: pr.Number, PRURL: pr.URL, Reason: "no-changelog label"})
			continue
		}

		parsed := parseChangelogSectionFull(pr.Body)
		if parsed == nil {
			result.Success = false
			result.Errors = append(result.Errors, assemblyError{
				PRNumber: pr.Number,
				PRURL:    pr.URL,
				Reason: fmt.Sprintf(
					"PR #%d (%s) has no parseable ## Changelog section and is not labeled 'no-changelog'. Add a ## Changelog section to the PR body or apply the no-changelog label.",
					pr.Number,
					pr.URL,
				),
			})
			continue
		}

		validationErrors := validateChangelogSectionFull(parsed)
		if len(validationErrors) > 0 {
			result.Success = false
			reason := fmt.Sprintf("PR #%d: ## Changelog section failed validation: %s", pr.Number, strings.Join(validationErrors, "; "))
			if parsed.CustomerImpact == "" {
				reason = fmt.Sprintf("PR #%d: ## Changelog section is missing the required Customer impact field", pr.Number)
			}
			result.Errors = append(result.Errors, assemblyError{PRNumber: pr.Number, PRURL: pr.URL, Reason: reason})
			continue
		}

		if parsed.BreakingChanges != "" {
			breakingChangeBlocks = append(breakingChangeBlocks, strings.TrimRight(parsed.BreakingChanges, "\n"))
		}

		if strings.ToLower(strings.TrimSpace(parsed.CustomerImpact)) == "none" {
			result.Excluded = append(result.Excluded, excludedPR{PRNumber: pr.Number, PRURL: pr.URL, Reason: "Customer impact: none", BreakingChanges: parsed.BreakingChanges})
			continue
		}

		changeBullets = append(changeBullets, buildChangeBullet(parsed.Summary, pr.Number, pr.URL))
		result.Included = append(result.Included, includedPR{PRNumber: pr.Number, PRURL: pr.URL, Summary: parsed.Summary, BreakingChanges: parsed.BreakingChanges})
	}

	if !result.Success {
		return result
	}

	parts := []string{}
	if len(breakingChangeBlocks) > 0 {
		parts = append(parts, "### Breaking changes", "")
		for _, block := range breakingChangeBlocks {
			parts = append(parts, block, "")
		}
	}
	if len(changeBullets) > 0 {
		parts = append(parts, "### Changes", "")
		parts = append(parts, changeBullets...)
		parts = append(parts, "")
	}
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	result.SectionBody = strings.Join(parts, "\n")
	return result
}
