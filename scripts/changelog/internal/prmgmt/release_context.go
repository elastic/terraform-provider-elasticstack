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

// Package prmgmt implements GitHub pull request maintenance for changelog-generation
// workflows (parity with changelog-pr-management.js and changelog-release-context.js).
package prmgmt

import (
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/semver"
)

const (
	// EventPullRequest is the GitHub Actions pull_request event name.
	EventPullRequest = "pull_request"
	// EventPullRequestTarget is the pull_request_target event name.
	EventPullRequestTarget = "pull_request_target"
	// WorkflowModeUnreleased is the unreleased changelog engine mode.
	WorkflowModeUnreleased = "unreleased"
	// WorkflowModeRelease is the release changelog engine mode.
	WorkflowModeRelease = "release"
)

const defaultUnreleasedTargetBranchName = "generated-changelog"

var releaseBranchRegexp = regexp.MustCompile(`^prep-release-(.+)$`)

// ReleaseModeResolution captures resolveReleaseMode output from changelog-release-context.js.
type ReleaseModeResolution struct {
	Mode          string
	TargetVersion string
	TargetBranch  string
}

// ResolveReleaseMode mirrors changelog-release-context.js resolveReleaseMode.
func ResolveReleaseMode(eventName, headBranch string) ReleaseModeResolution {
	mode := WorkflowModeUnreleased
	targetVersion := ""
	targetBranch := defaultUnreleasedTargetBranchName

	switch eventName {
	case EventPullRequest, EventPullRequestTarget:
		if m := releaseBranchRegexp.FindStringSubmatch(headBranch); len(m) == 2 {
			mode = WorkflowModeRelease
			targetVersion = m[1]
			targetBranch = headBranch
		}
	}

	return ReleaseModeResolution{
		Mode:          mode,
		TargetVersion: targetVersion,
		TargetBranch:  targetBranch,
	}
}

// ReleaseContext merges mode resolution with semver-selected compare metadata
// (buildReleaseContext in changelog-release-context.js).
type ReleaseContext struct {
	ReleaseModeResolution
	semver.PreviousTagResult
	CompareRange string
}

// BuildReleaseContext mirrors changelog-release-context.js buildReleaseContext.
func BuildReleaseContext(eventName, headBranch string, tags []semver.Tag) ReleaseContext {
	rm := ResolveReleaseMode(eventName, headBranch)
	prev := semver.SelectPreviousTag(tags, rm.Mode, rm.TargetVersion)
	cr := semver.BuildCompareRange(prev.PreviousTag)
	return ReleaseContext{
		ReleaseModeResolution: rm,
		PreviousTagResult:     prev,
		CompareRange:          cr,
	}
}
