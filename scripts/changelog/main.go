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

// Command changelog implements GitHub workflow helpers for changelog generation
// and PR changelog validation.
//
// Usage:
//
//	go run ./scripts/changelog <subcommand> [flags]
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/engine"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/evidence"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/githubx"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/prcheck"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/prmgmt"
	"github.com/google/go-github/v89/github"
)

func writeWorkflowWarning(msg string) {
	// Mirrors @actions/core.issueCommand(..., 'warning', ...) → stdout (::warning::$msg).
	d := escapeGitHubWorkflowData(msg)
	fmt.Fprintf(os.Stdout, "::warning::%s\n", d)
}

func escapeGitHubWorkflowData(msg string) string {
	msg = strings.ReplaceAll(msg, "%", "%25")
	msg = strings.ReplaceAll(msg, "\r", "%0D")
	msg = strings.ReplaceAll(msg, "\n", "%0A")
	return msg
}

const evidenceArtifactPerm fs.FileMode = 0o644

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "changelog: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer) error {
	if len(args) == 0 {
		return usageError(stderr)
	}
	switch args[0] {
	case "gather-evidence":
		return cmdGatherEvidence(args[1:], stderr)
	case "run-engine":
		return cmdRunEngine(args[1:], stderr)
	case "manage-unreleased-pr":
		return cmdManageUnreleasedPR(args[1:], stderr)
	case "refresh-release-pr":
		return cmdRefreshReleasePR(args[1:], stderr)
	case "validate-pr-section":
		return cmdValidatePRSection(args[1:], stderr)
	default:
		return usageError(stderr)
	}
}

func usageError(w io.Writer) error {
	fmt.Fprintln(w, "Usage: changelog <subcommand> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Subcommands:")
	fmt.Fprintln(w, "  gather-evidence        Build evidence manifest for a release range")
	fmt.Fprintln(w, "  run-engine             Run unreleased or release changelog engine")
	fmt.Fprintln(w, "  manage-unreleased-pr   Open or update the generated-changelog PR")
	fmt.Fprintln(w, "  refresh-release-pr     Refresh the release changelog PR")
	fmt.Fprintln(w, "  validate-pr-section    Validate PR body ## Changelog section")
	return errors.New("unknown or missing subcommand")
}

func cmdGatherEvidence(args []string, stderr io.Writer) error {
	ctx := context.Background()
	fsFlag := flag.NewFlagSet("gather-evidence", flag.ContinueOnError)
	fsFlag.SetOutput(stderr)
	modeFlag := fsFlag.String("mode", "",
		"evidence mode: unreleased|release (defaults to $MODE or unreleased)")
	targetVerFlag := fsFlag.String("target-version", "",
		"semver X.Y.Z without leading v for release (defaults to $TARGET_VERSION)")
	prevTagFlag := fsFlag.String("previous-tag", "",
		"previous git tag string (defaults to $PREVIOUS_TAG)")
	compareFlag := fsFlag.String("compare-range", "",
		"git log range (defaults to $COMPARE_RANGE or HEAD)")
	if err := fsFlag.Parse(args); err != nil {
		return err
	}

	mode := firstNonEmpty(*modeFlag,
		os.Getenv(githubx.EnvMode),
		os.Getenv(githubx.EnvInputMode),
	)
	if mode == "" {
		mode = engine.ModeUnreleased
	}

	targetVersion := firstNonEmpty(*targetVerFlag,
		os.Getenv(githubx.EnvTargetVersion),
		os.Getenv(githubx.EnvInputTargetVersion),
	)

	previousTag := firstNonEmpty(*prevTagFlag,
		os.Getenv(githubx.EnvPreviousTag),
		os.Getenv(githubx.EnvInputPreviousTag),
	)

	compareRange := firstNonEmpty(*compareFlag,
		os.Getenv(githubx.EnvCompareRange),
		os.Getenv(githubx.EnvInputCompareRange),
	)

	owner, repo, err := githubx.OwnerRepoFromEnv()
	if err != nil {
		return fmt.Errorf("gather-evidence: github repository env: %w", err)
	}

	client, err := githubx.NewGitHubClient(ctx, githubx.GitHubToken())
	if err != nil {
		return fmt.Errorf("gather-evidence: github client: %w", err)
	}

	gather := &gitMergedPRGatherer{client: client, execer: githubx.ShellGit{}}

	manifest, warns, gerr := evidence.Gather(ctx, evidence.GatherOptions{
		Owner:                    owner,
		Repo:                     repo,
		CompareRange:             compareRange,
		TargetVersion:            targetVersion,
		PreviousTag:              previousTag,
		Mode:                     mode,
		PRGatherer:               gather,
		ListPullRequestFilenames: githubxListFilesAdapter(client),
		Now:                      time.Now,
	})
	if gerr != nil {
		return fmt.Errorf("gather-evidence: %w", gerr)
	}

	for _, w := range warns {
		writeWorkflowWarning(w)
	}

	plan, err := evidence.BuildEvidenceArtifactPlan(evidence.ArtifactPlanRequest{
		Manifest: manifest,
	})
	if err != nil {
		return fmt.Errorf("gather-evidence: build artifact plan: %w", err)
	}

	if err := os.MkdirAll(plan.Directory, 0o755); err != nil {
		return fmt.Errorf("gather-evidence: mkdir %q: %w", plan.Directory, err)
	}

	if err := os.WriteFile(plan.ArtifactPath, []byte(plan.FormattedJSON), evidenceArtifactPerm); err != nil {
		return fmt.Errorf("gather-evidence: write evidence file: %w", err)
	}

	hasEvidence := len(manifest.PullRequests) > 0

	outPath := githubx.GitHubOutputPath()
	if outPath != "" {
		appendOut := func(name, value string) error {
			if perr := githubx.AppendGitHubOutput(outPath, name, value); perr != nil {
				return fmt.Errorf("GITHUB_OUTPUT (%s): %w", name, perr)
			}
			return nil
		}
		if werr := appendOut("evidence_file_path", plan.ArtifactPath); werr != nil {
			return fmt.Errorf("gather-evidence: %w", werr)
		}
		if werr := appendOut("has_evidence", boolGitHubActionsString(hasEvidence)); werr != nil {
			return fmt.Errorf("gather-evidence: %w", werr)
		}
	}

	return nil
}

func githubxListFilesAdapter(client *github.Client) func(context.Context, string, string, int) ([]string, error) {
	return func(ctx context.Context, owner, repo string, pullNumber int) ([]string, error) {
		return githubx.ListPullRequestFilenames(ctx, client, owner, repo, pullNumber)
	}
}

func cmdRunEngine(args []string, stderr io.Writer) error {
	ctx := context.Background()
	fsFlag := flag.NewFlagSet("run-engine", flag.ContinueOnError)
	fsFlag.SetOutput(stderr)
	modeFlag := fsFlag.String("mode", "",
		"changelog mode: unreleased|release (defaults to $MODE or unreleased)")
	targetVerFlag := fsFlag.String("target-version", "",
		"semver X.Y.Z without leading v for release (defaults to $TARGET_VERSION)")
	if err := fsFlag.Parse(args); err != nil {
		return err
	}

	mode := strings.TrimSpace(*modeFlag)
	if mode == "" {
		mode = strings.TrimSpace(os.Getenv(githubx.EnvMode))
	}
	if mode == "" {
		mode = engine.ModeUnreleased
	}

	targetVersion := strings.TrimSpace(*targetVerFlag)
	if targetVersion == "" {
		targetVersion = strings.TrimSpace(os.Getenv(githubx.EnvTargetVersion))
	}

	changelogPath := strings.TrimSpace(os.Getenv(githubx.EnvChangelogPath))
	if changelogPath == "" {
		changelogPath = "CHANGELOG.md"
	}

	targetBranchOverride := strings.TrimSpace(os.Getenv(githubx.EnvTargetBranch))

	owner, repo, err := githubx.OwnerRepoFromEnv()
	if err != nil {
		return fmt.Errorf("run-engine: github repository env: %w", err)
	}

	client, err := githubx.NewGitHubClient(ctx, githubx.GitHubToken())
	if err != nil {
		return err
	}

	gather := &gitMergedPRGatherer{client: client, execer: githubx.ShellGit{}}

	res, err := engine.Run(ctx, engine.Options{
		Mode:          mode,
		TargetVersion: targetVersion,
		Owner:         owner,
		Repo:          repo,
		ChangelogPath: changelogPath,
		Now:           time.Now,
		FS:            osChangelogFS{},
		Git:           githubx.ShellGit{},
		Gather:        gather,
	})
	if err != nil {
		return err
	}

	for _, w := range res.Warnings {
		writeWorkflowWarning(w)
	}

	effectiveBranch := targetBranchOverride
	if effectiveBranch == "" {
		effectiveBranch = res.TargetBranch
	}

	outPath := githubx.GitHubOutputPath()
	if outPath != "" {
		appendOut := func(name, value string) error {
			if perr := githubx.AppendGitHubOutput(outPath, name, value); perr != nil {
				return fmt.Errorf("GITHUB_OUTPUT (%s): %w", name, perr)
			}
			return nil
		}
		writes := [][2]string{
			{"mode", mode},
			{"target_version", res.TargetVersionOutput},
			{"previous_tag", res.PreviousTag},
			{"compare_range", res.CompareRange},
			{"target_branch", effectiveBranch},
			{"has_prs", boolGitHubActionsString(res.HasPRs)},
			{"has_user_facing_changes", boolGitHubActionsString(res.HasUserFacingChanges)},
			{"section_header", res.SectionHeader},
		}
		for _, kv := range writes {
			if werr := appendOut(kv[0], kv[1]); werr != nil {
				return werr
			}
		}
	}

	return nil
}

// osChangelogFS implements engine.FS backed by OS file operations.
type osChangelogFS struct{}

func (osChangelogFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (osChangelogFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func boolGitHubActionsString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func cmdManageUnreleasedPR(args []string, stderr io.Writer) error {
	ctx := context.Background()
	fs := flag.NewFlagSet("manage-unreleased-pr", flag.ContinueOnError)
	fs.SetOutput(stderr)
	compareFlag := fs.String("compare-range", "",
		"git compare range for PR body metadata (defaults to $COMPARE_RANGE)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	compareRange := firstNonEmpty(*compareFlag,
		os.Getenv(githubx.EnvCompareRange),
		os.Getenv(githubx.EnvInputCompareRange),
	)

	owner, repo, err := githubx.OwnerRepoFromEnv()
	if err != nil {
		return fmt.Errorf("manage-unreleased-pr: github repository env: %w", err)
	}

	client, err := githubx.NewGitHubClient(ctx, githubx.GitHubToken())
	if err != nil {
		return fmt.Errorf("manage-unreleased-pr: github client: %w", err)
	}

	res, err := prmgmt.ManageUnreleasedPR(ctx, prmgmt.ManageUnreleasedOptions{
		Owner:        owner,
		Repo:         repo,
		CompareRange: compareRange,
		GitHub:       newChangelogRESTAdapter(client),
		Now:          time.Now,
	})
	if err != nil {
		return fmt.Errorf("manage-unreleased-pr: %w", err)
	}

	for _, w := range res.Warnings {
		writeWorkflowWarning(w)
	}

	outPath := githubx.GitHubOutputPath()
	if outPath != "" {
		appendOut := func(name, value string) error {
			if perr := githubx.AppendGitHubOutput(outPath, name, value); perr != nil {
				return fmt.Errorf("GITHUB_OUTPUT (%s): %w", name, perr)
			}
			return nil
		}
		for _, kv := range []struct{ k, v string }{
			{"pr_action", res.Action},
			{"pr_number", fmt.Sprintf("%d", res.Number)},
			{"pr_url", res.URL},
		} {
			if werr := appendOut(kv.k, kv.v); werr != nil {
				return fmt.Errorf("manage-unreleased-pr: %w", werr)
			}
		}
	}

	return nil
}

func cmdRefreshReleasePR(args []string, stderr io.Writer) error {
	ctx := context.Background()
	fs := flag.NewFlagSet("refresh-release-pr", flag.ContinueOnError)
	fs.SetOutput(stderr)
	compareFlag := fs.String("compare-range", "",
		"git compare range for PR body metadata (defaults to $COMPARE_RANGE)")
	prNumberFlag := fs.Int("pr-number", 0,
		"release prep pull request number (defaults to pull_request.number from $GITHUB_EVENT_PATH when set)")
	targetVerFlag := fs.String("target-version", "",
		"release semver X.Y.Z without v (defaults to $TARGET_VERSION)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	prNumber := *prNumberFlag
	if prNumber <= 0 {
		num, perr := githubx.OptionalPullRequestNumberFromEventPath(os.Getenv(githubx.EnvGitHubEventPath))
		if perr != nil {
			return fmt.Errorf("refresh-release-pr: parse event: %w", perr)
		}
		prNumber = num
	}

	compareRange := firstNonEmpty(*compareFlag,
		os.Getenv(githubx.EnvCompareRange),
		os.Getenv(githubx.EnvInputCompareRange),
	)
	targetVersion := firstNonEmpty(*targetVerFlag,
		os.Getenv(githubx.EnvTargetVersion),
		os.Getenv(githubx.EnvInputTargetVersion),
	)

	owner, repo, err := githubx.OwnerRepoFromEnv()
	if err != nil {
		return fmt.Errorf("refresh-release-pr: github repository env: %w", err)
	}

	client, err := githubx.NewGitHubClient(ctx, githubx.GitHubToken())
	if err != nil {
		return fmt.Errorf("refresh-release-pr: github client: %w", err)
	}

	res, err := prmgmt.RefreshReleasePR(ctx, prmgmt.RefreshReleaseOptions{
		Owner:         owner,
		Repo:          repo,
		PRNumber:      prNumber,
		CompareRange:  compareRange,
		TargetVersion: targetVersion,
		GitHub:        newChangelogRESTAdapter(client),
		Now:           time.Now,
	})
	if err != nil {
		return fmt.Errorf("refresh-release-pr: %w", err)
	}

	if len(res.Infos) > 0 {
		fmt.Fprintf(stderr, "%s\n", res.Infos[0])
	}
	for _, w := range res.Warnings {
		writeWorkflowWarning(w)
	}
	if len(res.Infos) > 1 {
		fmt.Fprintf(stderr, "%s\n", res.Infos[1])
	}

	return nil
}

func cmdValidatePRSection(args []string, stderr io.Writer) error {
	ctx := context.Background()
	fs := flag.NewFlagSet("validate-pr-section", flag.ContinueOnError)
	fs.SetOutput(stderr)
	prNumberFlag := fs.Int("pr-number", 0,
		"pull request number (defaults to pull_request.number from $GITHUB_EVENT_PATH when set)")
	upsertCommentFlag := fs.Bool("upsert-comment", true,
		"whether to upsert the github-actions verdict comment when GITHUB_TOKEN is available")
	if err := fs.Parse(args); err != nil {
		return err
	}

	prNumber := *prNumberFlag
	if prNumber <= 0 {
		n, perr := githubx.OptionalPullRequestNumberFromEventPath(os.Getenv(githubx.EnvGitHubEventPath))
		if perr != nil {
			return fmt.Errorf("validate-pr-section: parse event: %w", perr)
		}
		prNumber = n
	}
	if prNumber <= 0 {
		return errors.New("validate-pr-section: missing pull request number (use --pr-number or a pull_request payload in GITHUB_EVENT_PATH)")
	}

	owner, repo, err := githubx.OwnerRepoFromEnv()
	if err != nil {
		return fmt.Errorf("validate-pr-section: github repository env: %w", err)
	}

	token := strings.TrimSpace(githubx.GitHubToken())
	if token == "" {
		return errors.New("validate-pr-section: missing GITHUB_TOKEN")
	}

	client, err := githubx.NewGitHubClient(ctx, token)
	if err != nil {
		return fmt.Errorf("validate-pr-section: github client: %w", err)
	}

	verdict, err := prcheck.Validate(ctx, prcheck.ValidateOptions{
		Owner:   owner,
		Repo:    repo,
		Number:  prNumber,
		Fetcher: &pullRequestFetcher{client: client},
	})
	if err != nil {
		return fmt.Errorf("validate-pr-section: %w", err)
	}

	resultJSON, err := json.Marshal(verdict)
	if err != nil {
		return fmt.Errorf("validate-pr-section: marshal verdict: %w", err)
	}

	if outPath := githubx.GitHubOutputPath(); outPath != "" {
		if err := githubx.AppendGitHubOutput(outPath, "result_json", string(resultJSON)); err != nil {
			return fmt.Errorf("validate-pr-section: write GITHUB_OUTPUT result_json: %w", err)
		}
	}

	if *upsertCommentFlag {
		if err := prcheck.UpsertVerdictIssueComment(ctx, ghIssueCommentsRESTAdapter{client: client}, owner, repo, prNumber, verdict); err != nil {
			return fmt.Errorf("validate-pr-section: verdict comment: %w", err)
		}
	}

	if verdict.Status == prcheck.StatusFail {
		return fmt.Errorf("validate-pr-section: PR changelog check failed:\n%s", strings.Join(verdict.Errors, "\n"))
	}

	return nil
}

// ghIssueCommentsRESTAdapter exposes githubx helpers through prcheck.IssueCommentsREST.
type ghIssueCommentsRESTAdapter struct {
	client *github.Client
}

func (a ghIssueCommentsRESTAdapter) ListIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]prcheck.Comment, error) {
	raw, err := githubx.ListIssueComments(ctx, a.client, owner, repo, issueNumber)
	if err != nil {
		return nil, err
	}
	out := make([]prcheck.Comment, len(raw))
	for i := range raw {
		out[i] = prcheck.Comment{
			ID:        raw[i].ID,
			Body:      raw[i].Body,
			UserLogin: raw[i].UserLogin,
		}
	}
	return out, nil
}

func (a ghIssueCommentsRESTAdapter) CreateIssueComment(ctx context.Context, owner, repo string, issueNumber int, body string) error {
	return githubx.CreateIssueComment(ctx, a.client, owner, repo, issueNumber, body)
}

func (a ghIssueCommentsRESTAdapter) UpdateIssueComment(ctx context.Context, owner, repo string, commentID int64, body string) error {
	return githubx.UpdateIssueComment(ctx, a.client, owner, repo, commentID, body)
}

// pullRequestFetcher implements prcheck.PullRequestFetcher using go-github REST pull requests.
type pullRequestFetcher struct {
	client *github.Client
}

func (f *pullRequestFetcher) GetPullRequest(ctx context.Context, owner, repo string, number int) (*prcheck.PullRequest, error) {
	if f == nil || f.client == nil {
		return nil, errors.New("pull request fetcher requires github client")
	}
	pr, _, err := f.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	body := ""
	if pr.Body != nil {
		body = *pr.Body
	}
	labelNames := make([]string, 0, len(pr.Labels))
	for _, l := range pr.Labels {
		if l == nil {
			continue
		}
		if n := strings.TrimSpace(l.GetName()); n != "" {
			labelNames = append(labelNames, n)
		}
	}
	return &prcheck.PullRequest{
		Number: pr.GetNumber(),
		Body:   body,
		Labels: labelNames,
	}, nil
}
