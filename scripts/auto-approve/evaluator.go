package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-github/v74/github"
)

const (
	maxEditedLines = 300
)

var allowedCheckConclusions = map[string]struct{}{
	"success": {},
	"neutral": {},
	"skipped": {},
}

var allowedCopilotAuthorLogins = map[string]struct{}{
	"github-copilot[bot]": {},
	"Copilot":             {},
}

type EvaluationInput struct {
	PullRequest       *github.PullRequest
	Commits           []*github.RepositoryCommit
	Files             []*github.CommitFile
	CombinedState     string
	CheckRuns         []*github.CheckRun
	Reviews           []*github.PullRequestReview
	ApproverLogin     string
	RepositoryOwner   string
	RepositoryName    string
	PullRequestNumber int
}

type EvaluationResult struct {
	ShouldApprove   bool     `json:"should_approve"`
	AlreadyApproved bool     `json:"already_approved"`
	Reasons         []string `json:"reasons"`
}

func Evaluate(input EvaluationInput) EvaluationResult {
	reasons := make([]string, 0)

	if input.PullRequest == nil {
		return EvaluationResult{
			ShouldApprove: false,
			Reasons:       []string{"missing pull request payload"},
		}
	}

	if input.PullRequest.GetState() != "open" {
		reasons = append(reasons, "pull request is not open")
	}

	if input.PullRequest.GetDraft() {
		reasons = append(reasons, "pull request is draft")
	}

	if !allCommitsByCopilot(input.Commits) {
		reasons = append(reasons, fmt.Sprintf("not all commits are authored by allowed Copilot identities (%s)", strings.Join(sortedCopilotAuthorLogins(), ", ")))
	}

	if !filesAllowed(input.Files) {
		reasons = append(reasons, "pull request contains files outside *_test.go and *.tf")
	}

	if !withinDiffThreshold(input.PullRequest.GetAdditions(), input.PullRequest.GetDeletions()) {
		reasons = append(reasons, fmt.Sprintf("edited lines must be < %d", maxEditedLines))
	}

	if !checksSuccessful(input.CombinedState, input.CheckRuns) {
		reasons = append(reasons, "not all checks are successful")
	}

	if approverAlreadyApproved(input.Reviews, input.ApproverLogin) {
		return EvaluationResult{
			ShouldApprove:   false,
			AlreadyApproved: true,
			Reasons:         []string{"approver has already submitted an approval review"},
		}
	}

	if len(reasons) > 0 {
		return EvaluationResult{
			ShouldApprove: false,
			Reasons:       reasons,
		}
	}

	return EvaluationResult{
		ShouldApprove: true,
		Reasons:       []string{"all gates passed"},
	}
}

func allCommitsByCopilot(commits []*github.RepositoryCommit) bool {
	if len(commits) == 0 {
		return false
	}

	for _, commit := range commits {
		if commit == nil || commit.Author == nil || commit.Author.Login == nil {
			return false
		}
		if _, ok := allowedCopilotAuthorLogins[commit.Author.GetLogin()]; !ok {
			return false
		}
	}

	return true
}

func sortedCopilotAuthorLogins() []string {
	logins := make([]string, 0, len(allowedCopilotAuthorLogins))
	for login := range allowedCopilotAuthorLogins {
		logins = append(logins, login)
	}
	sort.Strings(logins)
	return logins
}

func filesAllowed(files []*github.CommitFile) bool {
	for _, file := range files {
		if file == nil {
			return false
		}
		filename := file.GetFilename()
		if strings.HasSuffix(filename, "_test.go") || strings.HasSuffix(filename, ".tf") {
			continue
		}
		return false
	}
	return true
}

func withinDiffThreshold(additions int, deletions int) bool {
	return additions+deletions < maxEditedLines
}

func checksSuccessful(combinedState string, checkRuns []*github.CheckRun) bool {
	if combinedState != "success" {
		return false
	}

	for _, run := range checkRuns {
		if run == nil {
			return false
		}

		if run.GetStatus() != "completed" {
			return false
		}

		if _, ok := allowedCheckConclusions[run.GetConclusion()]; !ok {
			return false
		}
	}

	return true
}

func approverAlreadyApproved(reviews []*github.PullRequestReview, approverLogin string) bool {
	if approverLogin == "" {
		return false
	}

	for _, review := range reviews {
		if review == nil || review.User == nil {
			continue
		}
		if review.User.GetLogin() == approverLogin && strings.EqualFold(review.GetState(), "APPROVED") {
			return true
		}
	}

	return false
}
