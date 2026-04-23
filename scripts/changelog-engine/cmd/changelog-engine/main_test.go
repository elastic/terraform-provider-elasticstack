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
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	changelogengine "github.com/elastic/terraform-provider-elasticstack/scripts/changelog-engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWritesJSONOutputAndGitHubOutputsSeparately(t *testing.T) {
	oldRunEngine := runEngine
	t.Cleanup(func() {
		runEngine = oldRunEngine
	})

	runEngine = func(_ context.Context, cfg changelogengine.Config) (*changelogengine.RunResult, error) {
		return &changelogengine.RunResult{
			Outputs: changelogengine.Outputs{
				Mode:                 string(cfg.Mode),
				TargetVersion:        cfg.TargetVersion,
				TargetBranch:         "generated-changelog",
				PreviousTag:          "v1.0.0",
				CompareRange:         "v1.0.0..HEAD",
				SectionHeader:        "## [Unreleased]",
				HasChanges:           true,
				HasUserFacingChanges: true,
				PRCount:              1,
			},
			PullRequests: []changelogengine.PullRequestRecord{{Number: 11}},
			UpdatedBody:  "# Changelog",
		}, nil
	}

	dir := t.TempDir()
	jsonOutputPath := filepath.Join(dir, "result.json")
	githubOutputPath := filepath.Join(dir, "github-output.txt")
	t.Setenv("GITHUB_OUTPUT", githubOutputPath)

	result, err := run(context.Background(), changelogengine.Config{
		Mode:          changelogengine.ModeUnreleased,
		Owner:         "elastic",
		Repo:          "repo",
		Token:         "token",
		ChangelogPath: filepath.Join(dir, "CHANGELOG.md"),
	}, jsonOutputPath)
	require.NoError(t, err)
	require.NotNil(t, result)

	payload, err := os.ReadFile(jsonOutputPath)
	require.NoError(t, err)
	var persisted changelogengine.RunResult
	require.NoError(t, json.Unmarshal(payload, &persisted))
	assert.Equal(t, "generated-changelog", persisted.Outputs.TargetBranch)
	assert.Equal(t, 1, persisted.Outputs.PRCount)

	githubOutput, err := os.ReadFile(githubOutputPath)
	require.NoError(t, err)
	assert.Contains(t, string(githubOutput), "mode=unreleased")
	assert.Contains(t, string(githubOutput), "target_branch=generated-changelog")
	assert.NotContains(t, string(githubOutput), "\"outputs\"")
}
