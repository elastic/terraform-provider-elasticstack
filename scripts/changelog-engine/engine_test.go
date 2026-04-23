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
	engine.listPRsForCommitFunc = func(ctx context.Context, owner, repo, sha string) ([]*github.PullRequest, error) {
		return []*github.PullRequest{{
			Number:         github.Ptr(11),
			HTMLURL:        github.Ptr("https://example.test/pr/11"),
			State:          github.Ptr("closed"),
			MergedAt:       &github.Timestamp{Time: time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)},
			MergeCommitSHA: github.Ptr("abc123"),
			Body:           github.Ptr("## Changelog\nCustomer impact: fix\nSummary: Add a new thing"),
			User:           &github.User{Login: github.Ptr("octocat")},
		}}, nil
	}

	result, err := engine.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "## [Unreleased]", result.Outputs.SectionHeader)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "- Add a new thing ([#11](https://example.test/pr/11))")
}

func testEngine(t *testing.T) *Engine {
	t.Helper()
	engine, err := New(Config{Mode: ModeUnreleased, Owner: "elastic", Repo: "repo", Token: "token", ChangelogPath: "CHANGELOG.md"})
	require.NoError(t, err)
	return engine
}
