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

package evidence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/engine"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/section"
)

// Label sets mirrored from changelog-pr-evidence.js.
var (
	UserFacingLabels = []string{
		"enhancement", "bug", "feature", "breaking-change",
		"deprecation", "new-resource", "new-data-source",
	}
	InternalLabels = []string{
		"dependencies", "chore", "internal", "documentation", "ci", "test", "openspec",
	}
	// ProviderPathPrefixes mirrors PROVIDER_PATH_PREFIXES in changelog-pr-evidence.js.
	ProviderPathPrefixes = []string{
		"internal/", "pkg/", "libs/", "provider/", "go.mod", "go.sum",
	}
)

func labelSet(slice []string) map[string]struct{} {
	out := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		out[s] = struct{}{}
	}
	return out
}

var (
	userFacingSet = labelSet(UserFacingLabels)
	internalSet   = labelSet(InternalLabels)
)

const (
	botLoginDependabotBot    = "dependabot[bot]"
	botLoginDependabotLegacy = "dependabot"
	botLoginGitHubActionsBot = "github-actions[bot]"
	openspecPathPrefix       = "openspec/"
	comparePlaceholderHEAD   = "HEAD"
	commaJoinDelimiter       = ", "
)

// MergeCandidate selects merged closed PR rows (changelog-pr-evidence.js selectMergedPullRequests).
type MergeCandidate struct {
	Number   int
	State    string
	MergedAt bool
	Title    string // carried through for reorder tests (unused in filtering)
}

// SelectMergedPullRequests keeps merged closed PRs and de-duplicates by number (first occurrence wins).
func SelectMergedPullRequests(prs []MergeCandidate) []MergeCandidate {
	seen := make(map[int]struct{})
	var merged []MergeCandidate

	for _, pr := range prs {
		if pr.State == "closed" && pr.MergedAt {
			if _, dup := seen[pr.Number]; dup {
				continue
			}
			seen[pr.Number] = struct{}{}
			merged = append(merged, pr)
		}
	}
	return merged
}

// ParseCommitShas trims and filters git log `%H` output like parseCommitShas in changelog-pr-evidence.js.
func ParseCommitShas(raw string) []string {
	var out []string
	for ln := range strings.SplitSeq(raw, "\n") {
		s := strings.TrimSpace(ln)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func filterUserFacing(labels []string) []string {
	var names []string
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if _, ok := userFacingSet[label]; ok {
			names = append(names, label)
		}
	}
	return names
}

func filterInternalListed(labels []string) []string {
	var names []string
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if _, ok := internalSet[label]; ok {
			names = append(names, label)
		}
	}
	return names
}

func isAutomatedPRAuthor(login string) bool {
	switch login {
	case botLoginDependabotBot, botLoginDependabotLegacy, botLoginGitHubActionsBot:
		return true
	default:
		return false
	}
}

func touchesProviderCode(files []string) bool {
	for _, file := range files {
		fn := strings.TrimSpace(file)
		for _, pre := range ProviderPathPrefixes {
			if strings.HasPrefix(fn, pre) {
				return true
			}
		}
	}
	return false
}

func isOpenspecOnly(files []string) bool {
	if len(files) == 0 {
		return false
	}
	for _, file := range files {
		fn := strings.TrimSpace(file)
		if !strings.HasPrefix(fn, openspecPathPrefix) {
			return false
		}
	}
	return true
}

// ClassificationResult holds classifyPullRequestForChangelog output separate from persisted evidence rows.
type ClassificationResult struct {
	Classification     string
	InclusionRationale *string
	ExclusionRationale *string
}

// ClassifyPullRequestForChangelog mirrors changelog-pr-evidence.js classifyPullRequestForChangelog.
func ClassifyPullRequestForChangelog(labels []string, userLogin string, files []string) ClassificationResult {
	var hasUF, hasInternal bool
	for _, l := range labels {
		bl := strings.TrimSpace(l)
		if _, ok := userFacingSet[bl]; ok {
			hasUF = true
		}
		if _, ok := internalSet[bl]; ok {
			hasInternal = true
		}
	}

	isBot := isAutomatedPRAuthor(strings.TrimSpace(userLogin))
	touchesProv := touchesProviderCode(files)
	openspecONLY := isOpenspecOnly(files)

	var (
		clf       string
		inclusion *string
		exclusion *string
	)

	switch {
	case isBot:
		clf = classInternal
		msg := fmt.Sprintf("Automated PR by %s", strings.TrimSpace(userLogin))
		exclusion = &msg
	case openspecONLY:
		clf = classInternal
		msg := "Touches only openspec/ files — no provider code changes"
		exclusion = &msg
	case hasUF:
		clf = classUserFacing
		names := filterUserFacing(labels)
		msg := fmt.Sprintf("Has user-facing label(s): %s", strings.Join(names, commaJoinDelimiter))
		inclusion = &msg
	case hasInternal && !touchesProv:
		clf = classInternal
		names := filterInternalListed(labels)
		msg := fmt.Sprintf(
			"Has internal label(s): %s and does not touch provider code",
			strings.Join(names, commaJoinDelimiter),
		)
		exclusion = &msg
	case touchesProv:
		clf = classUserFacing
		msg := "Touches provider implementation paths — presumed user-facing"
		inclusion = &msg
	default:
		clf = classUncertain
		msg := "Classification uncertain — agent to decide"
		inclusion = &msg
	}

	return ClassificationResult{
		Classification:     clf,
		InclusionRationale: inclusion,
		ExclusionRationale: exclusion,
	}
}

// BuildPullRequestEvidence mirrors changelog-pr-evidence.js buildPullRequestEvidence.
func BuildPullRequestEvidence(
	number int,
	title, url, mergeCommitSHA string,
	userLogin string,
	labels []string,
	touchedFiles []string,
) PullRequestEvidence {
	cl := ClassifyPullRequestForChangelog(labels, userLogin, touchedFiles)
	return PullRequestEvidence{
		Number:             number,
		Title:              title,
		URL:                url,
		MergeCommitSHA:     mergeCommitSHA,
		Author:             userLoginFallback(userLogin),
		Labels:             append([]string{}, labels...),
		TouchedFiles:       append([]string{}, touchedFiles...),
		Classification:     cl.Classification,
		InclusionRationale: cloneStrPtr(cl.InclusionRationale),
		ExclusionRationale: cloneStrPtr(cl.ExclusionRationale),
	}
}

func userLoginFallback(login string) string {
	if strings.TrimSpace(login) == "" {
		return "unknown"
	}
	return strings.TrimSpace(login)
}

func cloneStrPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func evidenceFromMergedPR(pr section.MergedPR, filenames []string) PullRequestEvidence {
	labelCopy := append([]string{}, pr.Labels...)
	return BuildPullRequestEvidence(
		pr.Number,
		pr.Title,
		pr.URL,
		pr.MergeCommitSHA,
		pr.AuthorLogin,
		labelCopy,
		filenames,
	)
}

// GatherOptions configures PR evidence aggregation (changelog/gather-pr-evidence.js parity).
type GatherOptions struct {
	Owner                    string
	Repo                     string
	CompareRange             string // e.g. "v1.2.2..HEAD" or "HEAD"; empty behaves like HEAD
	TargetVersion            string // semver tag text for release manifests
	PreviousTag              string // informational only in manifest.previous_tag
	Mode                     string
	PRGatherer               engine.MergedPRGatherer
	ListPullRequestFilenames func(context.Context, string, string, int) ([]string, error)
	Now                      func() time.Time
}

// Gather lists merged PRs in the git range with file lists, mirrors gather-pr-evidence.js + changelog-pr-evidence.js.
//
// Returned warnings mirror core.warning from the JS runner (listing commits failed, listing PRs/files failed).
func Gather(ctx context.Context, opts GatherOptions) (Manifest, []string, error) {
	switch {
	case strings.TrimSpace(opts.Owner) == "":
		return Manifest{}, nil, errors.New("evidence: owner must be non-empty")
	case strings.TrimSpace(opts.Repo) == "":
		return Manifest{}, nil, errors.New("evidence: repo must be non-empty")
	case opts.PRGatherer == nil:
		return Manifest{}, nil, errors.New("evidence: merged PR gatherer must be non-nil")
	case opts.ListPullRequestFilenames == nil:
		return Manifest{}, nil, errors.New("evidence: list PR files callback must be non-nil")
	case opts.Now == nil:
		return Manifest{}, nil, errors.New("evidence: time provider must be non-nil")
	}

	compareRange := strings.TrimSpace(opts.CompareRange)
	if compareRange == "" {
		compareRange = comparePlaceholderHEAD
	}

	mergedRecords, warns, gatherErr := opts.PRGatherer.GatherMergedPRs(ctx, opts.Owner, opts.Repo, compareRange)
	warnings := append([]string(nil), warns...)
	if gatherErr != nil {
		return Manifest{}, warnings, fmt.Errorf("gather merged pull requests: %w", gatherErr)
	}

	var evidence []PullRequestEvidence

	for _, pr := range mergedRecords {
		fn := opts.ListPullRequestFilenames

		paths, fileErr := fn(ctx, opts.Owner, opts.Repo, pr.Number)
		if fileErr != nil {
			msg := fmt.Sprintf("Failed to list files for PR #%d: %v", pr.Number, fileErr)
			warnings = append(warnings, msg)
			paths = nil
		}

		evidence = append(evidence, evidenceFromMergedPR(pr, paths))
	}

	generatedAt := opts.Now()

	return BuildEvidenceManifest(
		opts.Mode,
		opts.TargetVersion,
		opts.PreviousTag,
		compareRange,
		evidence,
		generatedAt,
	), warnings, nil
}
