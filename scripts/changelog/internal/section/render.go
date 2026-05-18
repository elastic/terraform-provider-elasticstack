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
	"fmt"
	"regexp"
	"slices"
	"strings"
)

const noChangelogLabel = "no-changelog"

const rendererExcludedImpactNoneReason = "Customer impact: none"

var bulletLeading = regexp.MustCompile(`^[-*+]\s*`)
var leadingWhitespace = regexp.MustCompile(`^\s+`)

// RenderOpts selects PR metadata rendered into CHANGELOG citations.
//
// changelog-renderer.js currently only consumes PR number + URL for deterministic output.
// Additional authoring metadata (contributor attribution, changelog links, etc.) is reserved here.
type RenderOpts struct {
	PRNumber int
	PRURL    string
}

type MergedPR struct {
	Number         int
	Title          string
	URL            string
	Labels         []string
	Body           string
	MergeCommitSHA string
	AuthorLogin    string
}

type AssemblyError struct {
	PRNumber int
	PRURL    string
	Reason   string
}

type IncludedPR struct {
	PRNumber        int
	PRURL           string
	Summary         string
	BreakingChanges *string
}

type ExcludedPR struct {
	PRNumber        int
	PRURL           string
	Reason          string
	BreakingChanges *string
}

type RenderResult struct {
	Success     bool
	SectionBody string
	Errors      []AssemblyError
	Included    []IncludedPR
	Excluded    []ExcludedPR
}

// NormalizeBulletPrefix matches normalizeBulletPrefix in changelog-renderer.js.
func NormalizeBulletPrefix(line string) string {
	stripped := bulletLeading.ReplaceAllString(line, "")
	stripped = leadingWhitespace.ReplaceAllString(stripped, "")
	return "- " + stripped
}

// BuildCitation matches buildCitation in changelog-renderer.js.
func BuildCitation(prNumber int, prURL string) string {
	return fmt.Sprintf(`([#%d](%s))`, prNumber, prURL)
}

// BuildChangeBullet matches buildChangeBullet in changelog-renderer.js.
func BuildChangeBullet(summary string, prNumber int, prURL string) string {
	ns := NormalizeBulletPrefix(strings.TrimSpace(summary))
	citation := BuildCitation(prNumber, prURL)
	return fmt.Sprintf("%s %s", ns, citation)
}

func rendererTreatsImpactAsNone(impactRaw string) bool {
	return strings.ToLower(strings.TrimSpace(impactRaw)) == impactLiteralNone
}

func trimBreakingBlockEnd(s string) string {
	return strings.TrimRightFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == '\v'
	})
}

// Render emits a deterministic markdown fragment for a single Included PR shaped like changelog-renderer.js'
// aggregated section parts (before multi-PR batching in RenderChangelogSection).
func Render(sec Section, opts RenderOpts) string {
	var sectionParts []string

	if strings.TrimSpace(sec.BreakingChanges) != "" {
		sectionParts = append(sectionParts, "### Breaking changes", "", trimBreakingBlockEnd(sec.BreakingChanges), "")
	}

	summaryTrim := strings.TrimSpace(sec.Summary)
	if !rendererTreatsImpactAsNone(sec.ImpactRaw) && summaryTrim != "" {
		bullet := BuildChangeBullet(summaryTrim, opts.PRNumber, opts.PRURL)
		sectionParts = append(sectionParts, "### Changes", "", bullet, "")
	}

	for len(sectionParts) > 0 && sectionParts[len(sectionParts)-1] == "" {
		sectionParts = sectionParts[:len(sectionParts)-1]
	}

	return strings.Join(sectionParts, "\n")
}

// RenderChangelogSection matches renderChangelogSection in changelog-renderer.js.
func RenderChangelogSection(merged []MergedPR) RenderResult {
	var included []IncludedPR
	var excluded []ExcludedPR

	var changeBullets []string
	var breakingChangeBlocks []string

	var errs []AssemblyError

	for _, pr := range merged {
		prNumber := pr.Number
		prURL := pr.URL
		labelNames := pr.Labels
		if labelNames == nil {
			labelNames = []string{}
		}

		if containsExactLabel(labelNames, noChangelogLabel) {
			excluded = append(excluded, ExcludedPR{PRNumber: prNumber, PRURL: prURL, Reason: "no-changelog label"})
			continue
		}

		parsed, err := Parse([]byte(pr.Body))
		if err != nil {
			errs = append(errs, AssemblyError{
				PRNumber: prNumber,
				PRURL:    prURL,
				Reason: fmt.Sprintf(
					`PR #%d (%s) has no parseable ## Changelog section and is not labeled 'no-changelog'. Add a ## Changelog section to the PR body or apply the no-changelog label.`,
					prNumber,
					prURL,
				),
			})
			continue
		}

		sec := parsed

		valid, validationErrors := ValidateChangelogSectionFull(&sec, ValidateOpts{RelaxBreakingImpactMatch: true})
		if !valid {
			var reason string
			if strings.TrimSpace(sec.ImpactRaw) == "" {
				reason = fmt.Sprintf("PR #%d: ## Changelog section is missing the required Customer impact field", prNumber)
			} else {
				reason = fmt.Sprintf(
					"PR #%d: ## Changelog section failed validation: %s",
					prNumber,
					strings.Join(validationErrors, "; "),
				)
			}
			errs = append(errs, AssemblyError{PRNumber: prNumber, PRURL: prURL, Reason: reason})
			continue
		}

		breakingMarkdown := ""
		if strings.TrimSpace(sec.BreakingChanges) != "" {
			breakingMarkdown = trimBreakingBlockEnd(sec.BreakingChanges)
		}

		hasBreakingMarkdown := breakingMarkdown != ""

		// Collect breaking-change blocks regardless of customerImpact (changelog-renderer.js).
		if hasBreakingMarkdown {
			breakingChangeBlocks = append(breakingChangeBlocks, breakingMarkdown)
		}

		if rendererTreatsImpactAsNone(sec.ImpactRaw) {
			ex := ExcludedPR{
				PRNumber: prNumber,
				PRURL:    prURL,
				Reason:   rendererExcludedImpactNoneReason,
			}
			if hasBreakingMarkdown {
				bc := breakingMarkdown
				ex.BreakingChanges = &bc
			}

			excluded = append(excluded, ex)

			continue
		}

		if strings.TrimSpace(sec.Summary) == "" {
			errs = append(errs, AssemblyError{
				PRNumber: prNumber,
				PRURL:    prURL,
				Reason: fmt.Sprintf(
					`PR #%d (%s) has Customer impact: %s but is missing the required Summary field.`,
					prNumber,
					prURL,
					sec.ImpactRaw,
				),
			})
			continue
		}

		bullet := BuildChangeBullet(strings.TrimSpace(sec.Summary), prNumber, prURL)
		changeBullets = append(changeBullets, bullet)

		inc := IncludedPR{
			PRNumber: prNumber,
			PRURL:    prURL,
			Summary:  strings.TrimSpace(sec.Summary),
		}
		if hasBreakingMarkdown {
			bc := breakingMarkdown
			inc.BreakingChanges = &bc
		}

		included = append(included, inc)
	}

	if len(errs) > 0 {
		return RenderResult{Success: false, SectionBody: "", Errors: errs, Included: included, Excluded: excluded}
	}

	sectionBody := renderMergedSectionBodies(breakingChangeBlocks, changeBullets)
	return RenderResult{
		Success:     true,
		SectionBody: sectionBody,
		Errors:      []AssemblyError{},
		Included:    included,
		Excluded:    excluded,
	}
}

func renderMergedSectionBodies(breakingChangeBlocks []string, changeBullets []string) string {
	var sectionParts []string

	if len(breakingChangeBlocks) > 0 {
		sectionParts = append(sectionParts, "### Breaking changes", "")
		for _, bc := range breakingChangeBlocks {
			sectionParts = append(sectionParts, trimBreakingBlockEnd(bc), "")
		}
	}

	if len(changeBullets) > 0 {
		sectionParts = append(sectionParts, "### Changes", "")
		sectionParts = append(sectionParts, changeBullets...)
		sectionParts = append(sectionParts, "")
	}

	for len(sectionParts) > 0 && sectionParts[len(sectionParts)-1] == "" {
		sectionParts = sectionParts[:len(sectionParts)-1]
	}

	if len(sectionParts) == 0 {
		return ""
	}
	return strings.Join(sectionParts, "\n")
}

func containsExactLabel(labels []string, want string) bool {
	return slices.Contains(labels, want)
}
