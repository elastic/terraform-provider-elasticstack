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
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v85/github"
)

type engineResult struct {
	Mode                 string
	TargetVersion        string
	TargetBranch         string
	PreviousTag          string
	CompareRange         string
	SectionHeader        string
	HasUserFacingChanges bool
	PullRequests         []pullRequestRecord
	IncludedPullRequests []includedPR
	ExcludedPullRequests []excludedPR
}

func runChangelogEngine(ctx context.Context, client *github.Client, owner, repo, mode, targetVersion, changelogPath string, generatedAt time.Time) (engineResult, error) {
	if client == nil {
		return engineResult{}, fmt.Errorf("github client is required")
	}
	if owner == "" || repo == "" {
		return engineResult{}, fmt.Errorf("owner and repo are required")
	}
	if changelogPath == "" {
		changelogPath = "CHANGELOG.md"
	}

	tags, err := listSemverTags()
	if err != nil {
		return engineResult{}, err
	}
	releaseCtx, err := resolveReleaseContext(mode, targetVersion, tags)
	if err != nil {
		return engineResult{}, err
	}
	commitShas, err := listCommitShasInRange(releaseCtx.CompareRange)
	if err != nil {
		return engineResult{}, err
	}
	mergedPRs, err := resolveMergedPullRequests(ctx, client, owner, repo, commitShas)
	if err != nil {
		return engineResult{}, err
	}
	rendered := renderChangelogSection(mergedPRs)
	if !rendered.Success {
		return engineResult{}, fmt.Errorf("changelog assembly failed: %v", rendered.Errors)
	}
	currentContent := ""
	if data, err := os.ReadFile(changelogPath); err == nil {
		currentContent = string(data)
	} else if !os.IsNotExist(err) {
		return engineResult{}, err
	}
	sectionHeader, newSectionContent := buildSectionContent(releaseCtx.Mode, releaseCtx.TargetVersion, generatedAt, rendered.SectionBody)
	updated := rewriteChangelogSection(currentContent, newSectionContent, releaseCtx.Mode, releaseCtx.TargetVersion)
	if err := os.WriteFile(changelogPath, []byte(updated), 0644); err != nil {
		return engineResult{}, err
	}

	return engineResult{
		Mode:                 releaseCtx.Mode,
		TargetVersion:        releaseCtx.TargetVersion,
		TargetBranch:         releaseCtx.TargetBranch,
		PreviousTag:          releaseCtx.PreviousTag,
		CompareRange:         releaseCtx.CompareRange,
		SectionHeader:        sectionHeader,
		HasUserFacingChanges: len(rendered.Included) > 0,
		PullRequests:         mergedPRs,
		IncludedPullRequests: rendered.Included,
		ExcludedPullRequests: rendered.Excluded,
	}, nil
}
