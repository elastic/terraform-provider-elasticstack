package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-github/v74/github"
)

const (
	maxEditedLines = 1000
)

var allowedCopilotAuthorLogins = map[string]struct{}{
	"github-copilot[bot]": {},
	"Copilot":             {},
	"tobio":               {},
}

type EvaluationInput struct {
	PullRequest       *github.PullRequest
	Commits           []*github.RepositoryCommit
	Files             []*github.CommitFile
	Reviews           []*github.PullRequestReview
	ApproverLogin     string
	RepositoryOwner   string
	RepositoryName    string
	PullRequestNumber int
}

type EvaluationResult struct {
	CategoryMatched string   `json:"category_matched,omitempty"`
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

	category := matchedCategory(input.PullRequest)
	if category == "" {
		reasons = append(reasons, "pull request did not match any auto-approve category")
	} else {
		reasons = append(reasons, evaluateCategoryGates(category, input)...)
	}

	if approverAlreadyApproved(input.Reviews, input.ApproverLogin) {
		return EvaluationResult{
			CategoryMatched: category,
			ShouldApprove:   false,
			AlreadyApproved: true,
			Reasons:         []string{"approver has already submitted an approval review"},
		}
	}

	if len(reasons) > 0 {
		return EvaluationResult{
			CategoryMatched: category,
			ShouldApprove:   false,
			Reasons:         reasons,
		}
	}

	return EvaluationResult{
		CategoryMatched: category,
		ShouldApprove:   true,
		Reasons:         []string{"all gates passed"},
	}
}

func matchedCategory(pr *github.PullRequest) string {
	if pr == nil || pr.User == nil {
		return ""
	}

	author := pr.User.GetLogin()
	if _, ok := allowedCopilotAuthorLogins[author]; ok {
		return "copilot"
	}
	if author == "dependabot[bot]" {
		return "dependabot"
	}
	return ""
}

func evaluateCategoryGates(category string, input EvaluationInput) []string {
	switch category {
	case "copilot":
		return evaluateCopilotCategory(input)
	case "dependabot":
		return nil
	default:
		return []string{fmt.Sprintf("unknown auto-approve category %q", category)}
	}
}

func evaluateCopilotCategory(input EvaluationInput) []string {
	reasons := make([]string, 0)
	if !allCommitsByCopilot(input.Commits) {
		reasons = append(reasons, fmt.Sprintf("not all commits are authored by allowed Copilot identities (%s)", strings.Join(sortedCopilotAuthorLogins(), ", ")))
	}
	if !filesAllowed(input.Files) {
		reasons = append(reasons, "pull request contains files outside *_test.go and *.tf")
	}
	if !withinDiffThreshold(input.PullRequest.GetAdditions(), input.PullRequest.GetDeletions()) {
		reasons = append(reasons, fmt.Sprintf("edited lines must be < %d", maxEditedLines))
	}
	return reasons
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
