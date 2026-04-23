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

package changelogengine

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	github "github.com/google/go-github/v85/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveReleaseContext(t *testing.T) {
	engine := testEngine(t)
	engine.gitExec = func(args ...string) ([]byte, error) {
		return []byte("v2.0.0\nv1.9.0\n"), nil
	}
	engine.config.Mode = ModeRelease
	engine.config.TargetVersion = "2.0.0"

	ctx, err := engine.ResolveReleaseContext()
	require.NoError(t, err)
	assert.Equal(t, ModeRelease, ctx.Mode)
	assert.Equal(t, "2.0.0", ctx.TargetVersion)
	assert.Equal(t, "prep-release-2.0.0", ctx.TargetBranch)
	assert.Equal(t, "v1.9.0", ctx.PreviousTag)
	assert.Equal(t, "v1.9.0..HEAD", ctx.CompareRange)
	assert.True(t, ctx.ExcludedCurrentTag)
}

func TestRenderChangelogSection(t *testing.T) {
	result := RenderChangelogSection([]PullRequestRecord{
		{Number: 1, URL: "https://example.test/pr/1", Body: "## Changelog\nCustomer impact: fix\nSummary: Fix a bug"},
		{Number: 2, URL: "https://example.test/pr/2", Labels: []string{"no-changelog"}},
		{Number: 3, URL: "https://example.test/pr/3", Body: "## Changelog\nCustomer impact: breaking\nSummary: Remove an API\n\n### Breaking changes\nMigration required"},
	})

	require.True(t, result.Success)
	assert.Contains(t, result.SectionBody, "### Breaking changes")
	assert.Contains(t, result.SectionBody, "### Changes")
	assert.Contains(t, result.SectionBody, "- Fix a bug ([#1](https://example.test/pr/1))")
	assert.Contains(t, result.SectionBody, "- Remove an API ([#3](https://example.test/pr/3))")
	assert.Len(t, result.Excluded, 1)
}

func TestRenderChangelogSectionValidationFailure(t *testing.T) {
	result := RenderChangelogSection([]PullRequestRecord{{Number: 1, URL: "https://example.test/pr/1", Body: "## Changelog\nSummary: Missing impact"}})
	require.False(t, result.Success)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Reason, "Customer impact")
}

func TestRewriteChangelogSection(t *testing.T) {
	content := "# Changelog\n\n## [Unreleased]\n\nOld unreleased\n\n## [1.0.0] - 2026-01-01\n\nPrevious"
	updated := RewriteChangelogSection(content, "## [Unreleased]\n\n### Changes\n\n- New entry", ModeUnreleased, "")
	assert.Contains(t, updated, "- New entry")
	assert.NotContains(t, updated, "Old unreleased")
	assert.Contains(t, updated, "## [1.0.0] - 2026-01-01")
}

func TestRunWritesChangelogAndOutputs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CHANGELOG.md")
	require.NoError(t, os.WriteFile(path, []byte("# Changelog\n\n## [Unreleased]\n"), 0o644))

	engine := testEngine(t)
	engine.config.ChangelogPath = path
	engine.config.Now = time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)
	engine.gitExec = func(args ...string) ([]byte, error) {
		switch args[0] {
		case "tag":
			return []byte("v1.0.0\n"), nil
		case "log":
			return []byte("abc123\n"), nil
		default:
			return nil, nil
		}
	}
	engine.listPRsForCommitFunc = func(_ context.Context, _, _, _ string) ([]*github.PullRequest, error) {
		return []*github.PullRequest{{
			Number:         new(11),
			HTMLURL:        new("https://example.test/pr/11"),
			State:          new("closed"),
			MergedAt:       &github.Timestamp{Time: time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)},
			MergeCommitSHA: new("abc123"),
			Body:           new("## Changelog\nCustomer impact: fix\nSummary: Add a new thing"),
			User:           &github.User{Login: new("octocat")},
		}}, nil
	}

	result, err := engine.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "## [Unreleased]", result.Outputs.SectionHeader)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "- Add a new thing ([#11](https://example.test/pr/11))")
}

func TestResolveMergedPullRequestsDeduplicatesAndFiltersMergedPRs(t *testing.T) {
	engine := testEngine(t)
	engine.gitExec = func(args ...string) ([]byte, error) {
		require.Equal(t, []string{"log", "--format=%H", "v1.0.0..HEAD"}, args)
		return []byte("sha1\nsha2\nsha3\n"), nil
	}
	engine.listPRsForCommitFunc = func(_ context.Context, _, _, sha string) ([]*github.PullRequest, error) {
		switch sha {
		case "sha1":
			return []*github.PullRequest{{
				Number:         github.Ptr(10),
				Title:          github.Ptr("merged"),
				HTMLURL:        github.Ptr("https://example.test/pr/10"),
				State:          github.Ptr("closed"),
				MergedAt:       &github.Timestamp{Time: time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)},
				MergeCommitSHA: github.Ptr("merge-sha-10"),
				Body:           github.Ptr("## Changelog\nCustomer impact: fix\nSummary: merged"),
				Labels:         []*github.Label{{Name: github.Ptr("enhancement")}},
				User:           &github.User{Login: github.Ptr("octocat")},
			}}, nil
		case "sha2":
			return []*github.PullRequest{{
				Number:   github.Ptr(10),
				State:    github.Ptr("closed"),
				MergedAt: &github.Timestamp{Time: time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)},
			}, {
				Number: github.Ptr(11),
				State:  github.Ptr("open"),
			}}, nil
		case "sha3":
			return []*github.PullRequest{{
				Number: github.Ptr(12),
				State:  github.Ptr("closed"),
			}}, nil
		default:
			return nil, nil
		}
	}

	prs, err := engine.ResolveMergedPullRequests(context.Background(), "v1.0.0..HEAD")
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, 10, prs[0].Number)
	assert.Equal(t, "merge-sha-10", prs[0].MergeCommitSHA)
	assert.Equal(t, []string{"enhancement"}, prs[0].Labels)
}

func TestNewRequiresExplicitReleaseTargetVersion(t *testing.T) {
	_, err := New(Config{Mode: ModeRelease, Owner: "elastic", Repo: "repo", Token: "token"})
	require.EqualError(t, err, "release mode requires target version")
}

func testEngine(t *testing.T) *Engine {
	t.Helper()
	engine, err := New(Config{Mode: ModeUnreleased, Owner: "elastic", Repo: "repo", Token: "token", ChangelogPath: "CHANGELOG.md"})
	require.NoError(t, err)
	return engine
}
