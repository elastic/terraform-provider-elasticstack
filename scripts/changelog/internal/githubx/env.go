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

package githubx

import (
	"fmt"
	"os"
	"strings"
)

const (
	// EnvGitHubRepository is the actions-provided repo slug "owner/name".
	EnvGitHubRepository = "GITHUB_REPOSITORY"
	// EnvGitHubEventPath carries the workflow event payload JSON path.
	EnvGitHubEventPath = "GITHUB_EVENT_PATH"
	// EnvGitHubToken carries the bearer token used by REST calls.
	EnvGitHubToken = "GITHUB_TOKEN"
	// EnvGitHubOutput names the file path for step outputs.
	EnvGitHubOutput = "GITHUB_OUTPUT"
	// EnvMode selects unreleased vs release engine behaviour.
	EnvMode = "MODE"
	// EnvTargetVersion is the semver X.Y.Z for release mode (no leading v).
	EnvTargetVersion = "TARGET_VERSION"
	// EnvTargetBranch optionally overrides the branch name written to outputs.
	EnvTargetBranch = "TARGET_BRANCH"
	// EnvChangelogPath overrides the CHANGELOG.md path.
	EnvChangelogPath = "CHANGELOG_PATH"
	// EnvPreviousTag is the tag spanning the changelog compare range baseline.
	EnvPreviousTag = "PREVIOUS_TAG"
	// EnvCompareRange is git rev-list range (e.g. v1.0.0..HEAD).
	EnvCompareRange = "COMPARE_RANGE"
	// EnvInput* mirror actions/core getInput env fallbacks used by github-script steps.
	EnvInputPreviousTag   = "INPUT_PREVIOUS_TAG"
	EnvInputCompareRange  = "INPUT_COMPARE_RANGE"
	EnvInputMode          = "INPUT_MODE"
	EnvInputTargetVersion = "INPUT_TARGET_VERSION"
)

// GitHubToken returns trimmed GITHUB_TOKEN (may be empty).
func GitHubToken() string {
	return strings.TrimSpace(os.Getenv(EnvGitHubToken))
}

// GitHubOutputPath returns trimmed GITHUB_OUTPUT (may be empty).
func GitHubOutputPath() string {
	return strings.TrimSpace(os.Getenv(EnvGitHubOutput))
}

// ParseGitHubOwnerRepo parses github.repository-style "owner/name".
func ParseGitHubOwnerRepo(repository string) (owner, repo string, err error) {
	repository = strings.TrimSpace(repository)
	parts := strings.Split(repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid GITHUB_REPOSITORY value %q", repository)
	}
	return parts[0], parts[1], nil
}

// OwnerRepoFromEnv reads GITHUB_REPOSITORY.
func OwnerRepoFromEnv() (owner, repo string, err error) {
	return ParseGitHubOwnerRepo(os.Getenv(EnvGitHubRepository))
}
