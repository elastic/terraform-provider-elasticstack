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

package prmgmt

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	releasePullRequestPrepBranchFmt         = "prep-release-%s"
	warningPrepReleaseUnresolvedVersionSlug = "prep-release-<version>"
	warningDisplayedBaseBranch              = "main"
)

// RefreshReleaseOptions configures RefreshReleasePR.
type RefreshReleaseOptions struct {
	Owner         string
	Repo          string
	BaseBranch    string
	PRNumber      int // 0 triggers lookup via prep-release-<TargetVersion>.
	CompareRange  string
	TargetVersion string
	GitHub        ChangelogWorkflowREST
	Now           func() time.Time
}

// RefreshReleaseResult reports whether a PR body refresh ran.
type RefreshReleaseResult struct {
	Warnings []string
	Updated  bool
	Number   int
}

// BuildReleasePRBody mirrors buildReleasePRBody from changelog-pr-management.js.
func BuildReleasePRBody(targetVersion string, compareRange string, generatedDate string) string {
	lines := []string{
		fmt.Sprintf("**Generated:** %s", generatedDate),
		fmt.Sprintf("**Version:** `%s`", targetVersion),
		fmt.Sprintf("**Compare range:** `%s`", compareRange),
		"",
		"> This PR body was last refreshed by the changelog-generation workflow.",
	}
	return strings.Join(lines, "\n")
}

// FindOpenReleasePrepPRNumber mirrors findOpenReleasePrepPRNumber from changelog-pr-management.js.
func FindOpenReleasePrepPRNumber(ctx context.Context, gh ChangelogWorkflowREST, owner, repo, baseBranch, targetVersion string) (int, error) {
	if gh == nil {
		return 0, fmt.Errorf("find release prep pr: github client required")
	}
	targetVersion = strings.TrimSpace(targetVersion)
	if targetVersion == "" {
		return 0, nil
	}
	headBranch := fmt.Sprintf(releasePullRequestPrepBranchFmt, targetVersion)
	headRef := fmt.Sprintf("%s:%s", owner, headBranch)

	prs, err := gh.ListOpenPullRequestsByHead(ctx, owner, repo, headRef, baseBranch)
	if err != nil {
		return 0, fmt.Errorf("list prep-release pull requests: %w", err)
	}
	if len(prs) == 0 {
		return 0, nil
	}
	return prs[0].Number, nil
}

func releasePrepResolveWarning(targetVersion string) string {
	versionSlug := warningPrepReleaseUnresolvedVersionSlug
	if tv := strings.TrimSpace(targetVersion); tv != "" {
		versionSlug = fmt.Sprintf(releasePullRequestPrepBranchFmt, tv)
	}
	return fmt.Sprintf(
		"Could not resolve a release prep PR to refresh (no PR number and no open PR for %s → %s); "+
			"skipping PR body update",
		versionSlug,
		warningDisplayedBaseBranch,
	)
}

// RefreshReleasePR mirrors refreshReleasePR from changelog-pr-management.js.
func RefreshReleasePR(ctx context.Context, opts RefreshReleaseOptions) (RefreshReleaseResult, error) {
	if opts.GitHub == nil {
		return RefreshReleaseResult{}, fmt.Errorf("refresh release pr: github client required")
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	base := opts.BaseBranch
	if base == "" {
		base = defaultPullRequestBaseBranch
	}

	num := opts.PRNumber
	if num <= 0 {
		found, ferr := FindOpenReleasePrepPRNumber(ctx, opts.GitHub, opts.Owner, opts.Repo, base, opts.TargetVersion)
		if ferr != nil {
			return RefreshReleaseResult{}, ferr
		}
		num = found
	}
	if num <= 0 {
		return RefreshReleaseResult{
			Warnings: []string{releasePrepResolveWarning(opts.TargetVersion)},
		}, nil
	}

	generatedDate := now().UTC().Format("2006-01-02")
	body := BuildReleasePRBody(opts.TargetVersion, opts.CompareRange, generatedDate)

	if err := opts.GitHub.UpdatePullRequestBody(ctx, opts.Owner, opts.Repo, num, body); err != nil {
		return RefreshReleaseResult{}, fmt.Errorf("update pull request body: %w", err)
	}

	res := RefreshReleaseResult{
		Updated: true,
		Number:  num,
	}
	appendNoChangelogLabelWarning(ctx, &res.Warnings, opts.GitHub, opts.Owner, opts.Repo, num)
	return res, nil
}
