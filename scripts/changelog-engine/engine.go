package changelogengine

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	github "github.com/google/go-github/v85/github"
	"golang.org/x/oauth2"
)

type Mode string

const (
	ModeUnreleased Mode = "unreleased"
	ModeRelease    Mode = "release"
)

var (
	semverTagPattern      = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	releaseBranchPattern  = regexp.MustCompile(`^prep-release-(.+)$`)
	changelogHeaderRegexp = regexp.MustCompile(`^##\s+Changelog`)
	topSectionRegexp      = regexp.MustCompile(`^##\s`)
	subSectionRegexp      = regexp.MustCompile(`^#{2,3}\s`)
	breakingHeaderRegexp  = regexp.MustCompile(`^###\s+Breaking changes`)
)

type Config struct {
	Mode          Mode
	TargetVersion string
	Owner         string
	Repo          string
	Token         string
	ChangelogPath string
	Now           time.Time
}

type Engine struct {
	client               *github.Client
	config               Config
	gitExec              func(args ...string) ([]byte, error)
	listPRsForCommitFunc func(ctx context.Context, owner, repo, sha string) ([]*github.PullRequest, error)
}

type ReleaseContext struct {
	Mode               Mode
	TargetVersion      string
	TargetBranch       string
	PreviousTag        string
	CompareRange       string
	ExcludedTag        string
	ExcludedCurrentTag bool
}

type PullRequestRecord struct {
	Number         int      `json:"number"`
	Title          string   `json:"title"`
	URL            string   `json:"url"`
	MergeCommitSHA string   `json:"merge_commit_sha"`
	Author         string   `json:"author"`
	Labels         []string `json:"labels"`
	Body           string   `json:"body"`
}

type Outputs struct {
	Mode                 string `json:"mode"`
	TargetVersion        string `json:"target_version"`
	TargetBranch         string `json:"target_branch"`
	PreviousTag          string `json:"previous_tag"`
	CompareRange         string `json:"compare_range"`
	SectionHeader        string `json:"section_header"`
	HasChanges           bool   `json:"has_changes"`
	HasUserFacingChanges bool   `json:"has_user_facing_changes"`
	PRCount              int    `json:"pr_count"`
}

type RunResult struct {
	Outputs      Outputs
	PullRequests []PullRequestRecord
	UpdatedBody  string
}

type RenderResult struct {
	Success     bool
	SectionBody string
	Errors      []AssemblyError
	Included    []IncludedPR
	Excluded    []ExcludedPR
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
	BreakingChanges string
}

type ExcludedPR struct {
	PRNumber        int
	PRURL           string
	Reason          string
	BreakingChanges string
}

type ParsedChangelog struct {
	CustomerImpact              string
	Summary                     string
	BreakingChanges             string
	BreakingChangesHeadingFound bool
}

func New(config Config) (*Engine, error) {
	if config.Mode != ModeUnreleased && config.Mode != ModeRelease {
		return nil, fmt.Errorf("unsupported mode %q", config.Mode)
	}
	if config.Mode == ModeRelease && config.TargetVersion == "" {
		return nil, errors.New("release mode requires target version")
	}
	if config.Owner == "" || config.Repo == "" {
		return nil, errors.New("owner and repo are required")
	}
	if config.Token == "" {
		return nil, errors.New("token is required")
	}
	if config.ChangelogPath == "" {
		config.ChangelogPath = "CHANGELOG.md"
	}
	if config.Now.IsZero() {
		config.Now = time.Now().UTC()
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token})
	client := github.NewClient(oauth2.NewClient(context.Background(), ts))

	engine := &Engine{
		client: client,
		config: config,
		gitExec: func(args ...string) ([]byte, error) {
			cmd := exec.Command("git", args...)
			return cmd.CombinedOutput()
		},
	}
	engine.listPRsForCommitFunc = func(ctx context.Context, owner, repo, sha string) ([]*github.PullRequest, error) {
		prs, _, err := engine.client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, sha, nil)
		return prs, err
	}
	return engine, nil
}

func (e *Engine) Run(ctx context.Context) (*RunResult, error) {
	releaseContext, err := e.ResolveReleaseContext()
	if err != nil {
		return nil, err
	}

	prs, err := e.ResolveMergedPullRequests(ctx, releaseContext.CompareRange)
	if err != nil {
		return nil, err
	}

	rendered := RenderChangelogSection(prs)
	if !rendered.Success {
		var reasons []string
		for _, item := range rendered.Errors {
			reasons = append(reasons, item.Reason)
		}
		return nil, fmt.Errorf("changelog assembly failed:\n%s", strings.Join(reasons, "\n"))
	}

	current, err := os.ReadFile(e.config.ChangelogPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read changelog: %w", err)
	}

	sectionHeader := buildSectionHeader(releaseContext.Mode, releaseContext.TargetVersion, e.config.Now)
	newSectionContent := sectionHeader
	if rendered.SectionBody != "" {
		newSectionContent += "\n\n" + rendered.SectionBody
	}

	updated := RewriteChangelogSection(string(current), newSectionContent, releaseContext.Mode, releaseContext.TargetVersion)
	if err := os.WriteFile(e.config.ChangelogPath, []byte(updated), 0o644); err != nil {
		return nil, fmt.Errorf("write changelog: %w", err)
	}

	return &RunResult{
		Outputs: Outputs{
			Mode:                 string(releaseContext.Mode),
			TargetVersion:        releaseContext.TargetVersion,
			TargetBranch:         releaseContext.TargetBranch,
			PreviousTag:          releaseContext.PreviousTag,
			CompareRange:         releaseContext.CompareRange,
			SectionHeader:        sectionHeader,
			HasChanges:           len(rendered.Included) > 0 || len(rendered.Excluded) > 0,
			HasUserFacingChanges: len(rendered.Included) > 0,
			PRCount:              len(prs),
		},
		PullRequests: prs,
		UpdatedBody:  updated,
	}, nil
}

func (e *Engine) ResolveReleaseContext() (ReleaseContext, error) {
	tags, err := e.semverTags()
	if err != nil {
		return ReleaseContext{}, err
	}

	ctx := ReleaseContext{
		Mode:          e.config.Mode,
		TargetVersion: e.config.TargetVersion,
		TargetBranch:  "generated-changelog",
	}
	if ctx.Mode == ModeRelease {
		ctx.TargetBranch = fmt.Sprintf("prep-release-%s", ctx.TargetVersion)
		ctx.ExcludedTag = "v" + ctx.TargetVersion
	}

	candidates := tags
	if ctx.ExcludedTag != "" {
		candidates = nil
		for _, tag := range tags {
			if tag == ctx.ExcludedTag {
				ctx.ExcludedCurrentTag = true
				continue
			}
			candidates = append(candidates, tag)
		}
	}
	if len(candidates) > 0 {
		ctx.PreviousTag = candidates[0]
		ctx.CompareRange = ctx.PreviousTag + "..HEAD"
	} else {
		ctx.CompareRange = "HEAD"
	}

	return ctx, nil
}

func (e *Engine) ResolveMergedPullRequests(ctx context.Context, compareRange string) ([]PullRequestRecord, error) {
	shas, err := e.commitSHAs(compareRange)
	if err != nil {
		return nil, err
	}

	byNumber := map[int]PullRequestRecord{}
	for _, sha := range shas {
		prs, err := e.listPRsForCommitFunc(ctx, e.config.Owner, e.config.Repo, sha)
		if err != nil {
			return nil, fmt.Errorf("list pull requests for commit %s: %w", sha, err)
		}
		for _, pr := range prs {
			if pr.GetState() != "closed" || pr.GetMergedAt().IsZero() {
				continue
			}
			if _, exists := byNumber[pr.GetNumber()]; exists {
				continue
			}
			byNumber[pr.GetNumber()] = PullRequestRecord{
				Number:         pr.GetNumber(),
				Title:          pr.GetTitle(),
				URL:            pr.GetHTMLURL(),
				MergeCommitSHA: pr.GetMergeCommitSHA(),
				Author:         pr.GetUser().GetLogin(),
				Labels:         labelNames(pr.Labels),
				Body:           pr.GetBody(),
			}
		}
	}

	result := make([]PullRequestRecord, 0, len(byNumber))
	for _, pr := range byNumber {
		result = append(result, pr)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Number < result[j].Number })
	return result, nil
}

func labelNames(labels []*github.Label) []string {
	result := make([]string, 0, len(labels))
	for _, label := range labels {
		if name := label.GetName(); name != "" {
			result = append(result, name)
		}
	}
	return result
}

func (e *Engine) semverTags() ([]string, error) {
	output, err := e.gitExec("tag", "--list", "v[0-9]*.[0-9]*.[0-9]*", "--sort=-version:refname")
	if err != nil {
		return nil, fmt.Errorf("list git tags: %w (%s)", err, strings.TrimSpace(string(output)))
	}
	var tags []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		tag := strings.TrimSpace(scanner.Text())
		if semverTagPattern.MatchString(tag) {
			tags = append(tags, tag)
		}
	}
	return tags, scanner.Err()
}

func (e *Engine) commitSHAs(compareRange string) ([]string, error) {
	args := []string{"log", "--format=%H"}
	if compareRange != "" {
		args = append(args, compareRange)
	}
	output, err := e.gitExec(args...)
	if err != nil {
		return nil, fmt.Errorf("git log %s: %w (%s)", compareRange, err, strings.TrimSpace(string(output)))
	}
	var shas []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		sha := strings.TrimSpace(scanner.Text())
		if sha != "" {
			shas = append(shas, sha)
		}
	}
	return shas, scanner.Err()
}

func RenderChangelogSection(mergedPRs []PullRequestRecord) RenderResult {
	var result RenderResult
	var changeBullets []string
	var breakingBlocks []string

	for _, pr := range mergedPRs {
		if slices.Contains(pr.Labels, "no-changelog") {
			result.Excluded = append(result.Excluded, ExcludedPR{PRNumber: pr.Number, PRURL: pr.URL, Reason: "no-changelog label"})
			continue
		}

		parsed := ParseChangelogSection(pr.Body)
		if parsed == nil {
			result.Errors = append(result.Errors, AssemblyError{PRNumber: pr.Number, PRURL: pr.URL, Reason: fmt.Sprintf("PR #%d (%s) has no parseable ## Changelog section and is not labeled 'no-changelog'. Add a ## Changelog section to the PR body or apply the no-changelog label.", pr.Number, pr.URL)})
			continue
		}

		validationErrors := ValidateChangelogSection(parsed)
		if len(validationErrors) > 0 {
			reason := strings.Join(validationErrors, "; ")
			if parsed.CustomerImpact == "" {
				reason = fmt.Sprintf("PR #%d: ## Changelog section is missing the required Customer impact field", pr.Number)
			} else {
				reason = fmt.Sprintf("PR #%d: ## Changelog section failed validation: %s", pr.Number, reason)
			}
			result.Errors = append(result.Errors, AssemblyError{PRNumber: pr.Number, PRURL: pr.URL, Reason: reason})
			continue
		}

		if parsed.BreakingChanges != "" {
			breakingBlocks = append(breakingBlocks, strings.TrimRight(parsed.BreakingChanges, "\n"))
		}

		if strings.EqualFold(parsed.CustomerImpact, "none") {
			excluded := ExcludedPR{PRNumber: pr.Number, PRURL: pr.URL, Reason: "Customer impact: none"}
			if parsed.BreakingChanges != "" {
				excluded.BreakingChanges = parsed.BreakingChanges
			}
			result.Excluded = append(result.Excluded, excluded)
			continue
		}

		bullet := buildChangeBullet(parsed.Summary, pr.Number, pr.URL)
		changeBullets = append(changeBullets, bullet)
		result.Included = append(result.Included, IncludedPR{PRNumber: pr.Number, PRURL: pr.URL, Summary: parsed.Summary, BreakingChanges: parsed.BreakingChanges})
	}

	if len(result.Errors) > 0 {
		return result
	}

	var parts []string
	if len(breakingBlocks) > 0 {
		parts = append(parts, "### Breaking changes", "")
		for _, block := range breakingBlocks {
			parts = append(parts, block, "")
		}
	}
	if len(changeBullets) > 0 {
		parts = append(parts, "### Changes", "")
		parts = append(parts, changeBullets...)
	}
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	result.Success = true
	result.SectionBody = strings.Join(parts, "\n")
	return result
}

func buildChangeBullet(summary string, prNumber int, prURL string) string {
	return fmt.Sprintf("%s ([#%d](%s))", normalizeBulletPrefix(strings.TrimSpace(summary)), prNumber, prURL)
}

func normalizeBulletPrefix(line string) string {
	line = strings.TrimLeft(line, " \t")
	line = strings.TrimPrefix(line, "-")
	line = strings.TrimPrefix(line, "*")
	line = strings.TrimPrefix(line, "+")
	line = strings.TrimLeft(line, " \t")
	return "- " + line
}

func ParseChangelogSection(body string) *ParsedChangelog {
	section := extractSection(body, changelogHeaderRegexp, topSectionRegexp)
	if section == "" && !changelogHeaderRegexp.MatchString(body) {
		return nil
	}

	parsed := &ParsedChangelog{}
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "Customer impact:"):
			parsed.CustomerImpact = strings.TrimSpace(strings.TrimPrefix(trimmed, "Customer impact:"))
		case strings.HasPrefix(trimmed, "Summary:"):
			parsed.Summary = strings.TrimSpace(strings.TrimPrefix(trimmed, "Summary:"))
		case breakingHeaderRegexp.MatchString(trimmed):
			parsed.BreakingChangesHeadingFound = true
		}
	}
	parsed.BreakingChanges = extractSection(section, breakingHeaderRegexp, subSectionRegexp)
	return parsed
}

func ValidateChangelogSection(parsed *ParsedChangelog) []string {
	if parsed == nil {
		return []string{"No ## Changelog section found in PR body"}
	}

	var errs []string
	switch parsed.CustomerImpact {
	case "":
		errs = append(errs, "Missing required field: Customer impact")
	case "none", "fix", "enhancement", "breaking":
	default:
		errs = append(errs, fmt.Sprintf("Invalid Customer impact value: %q. Must be one of: none, fix, enhancement, breaking", parsed.CustomerImpact))
	}

	if parsed.CustomerImpact != "" && !strings.EqualFold(parsed.CustomerImpact, "none") && parsed.Summary == "" {
		errs = append(errs, "Missing required field: Summary (required when Customer impact is not \"none\")")
	}
	if parsed.BreakingChangesHeadingFound && parsed.BreakingChanges == "" {
		errs = append(errs, "### Breaking changes section is present but contains no content")
	}
	if strings.EqualFold(parsed.CustomerImpact, "breaking") && !parsed.BreakingChangesHeadingFound {
		errs = append(errs, "Customer impact: breaking requires a ### Breaking changes subsection")
	}
	return errs
}

func extractSection(body string, start *regexp.Regexp, end *regexp.Regexp) string {
	var lines []string
	inSection := false
	var fence string
	for _, line := range strings.Split(body, "\n") {
		if !inSection && start.MatchString(line) {
			inSection = true
			continue
		}
		if !inSection {
			continue
		}
		if fence == "" {
			switch {
			case strings.HasPrefix(line, "```"):
				fence = "```"
			case strings.HasPrefix(line, "~~~"):
				fence = "~~~"
			case end.MatchString(line):
				return strings.TrimRight(strings.Join(lines, "\n"), "\n")
			}
		} else if strings.HasPrefix(line, fence) {
			fence = ""
		}
		lines = append(lines, line)
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func RewriteChangelogSection(content, newSectionContent string, mode Mode, targetVersion string) string {
	lines := strings.Split(content, "\n")
	targetStart := -1
	if mode == ModeUnreleased {
		for i, line := range lines {
			if strings.HasPrefix(line, "## [Unreleased]") {
				targetStart = i
				break
			}
		}
	} else {
		header := fmt.Sprintf("## [%s]", targetVersion)
		for i, line := range lines {
			if strings.HasPrefix(line, header) {
				targetStart = i
				break
			}
		}
	}

	if targetStart == -1 {
		if mode == ModeRelease {
			for i, line := range lines {
				if strings.HasPrefix(line, "## [Unreleased]") {
					end := findSectionEnd(lines, i)
					before := append([]string{}, lines[:end]...)
					after := lines[end:]
					return strings.Join(append(append(before, "", newSectionContent), after...), "\n")
				}
			}
		}
		if strings.TrimSpace(content) == "" {
			return newSectionContent
		}
		return newSectionContent + "\n\n" + content
	}

	end := findSectionEnd(lines, targetStart)
	before := append([]string{}, lines[:targetStart]...)
	after := lines[end:]
	for len(before) > 0 && before[len(before)-1] == "" {
		before = before[:len(before)-1]
	}
	for len(after) > 0 && after[0] == "" {
		after = after[1:]
	}
	parts := append([]string{}, before...)
	if len(parts) > 0 {
		parts = append(parts, "")
	}
	parts = append(parts, newSectionContent)
	if len(after) > 0 {
		parts = append(parts, "")
		parts = append(parts, after...)
	}
	return strings.Join(parts, "\n")
}

func findSectionEnd(lines []string, start int) int {
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			return i
		}
	}
	return len(lines)
}

func buildSectionHeader(mode Mode, targetVersion string, now time.Time) string {
	if mode == ModeRelease {
		return fmt.Sprintf("## [%s] - %s", targetVersion, now.UTC().Format("2006-01-02"))
	}
	return "## [Unreleased]"
}

func ResolveReleaseMode(eventName, headBranch string) (Mode, string, string) {
	if (eventName == "pull_request" || eventName == "pull_request_target") && releaseBranchPattern.MatchString(headBranch) {
		match := releaseBranchPattern.FindStringSubmatch(headBranch)
		return ModeRelease, match[1], headBranch
	}
	return ModeUnreleased, "", "generated-changelog"
}
