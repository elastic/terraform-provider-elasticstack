package main

import (
	"strings"
	"testing"

	"github.com/google/go-github/v74/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluate(t *testing.T) {
	t.Parallel()

	makeCommit := func(login string) *github.RepositoryCommit {
		return &github.RepositoryCommit{
			Author: &github.User{Login: github.Ptr(login)},
		}
	}

	makeFile := func(name string) *github.CommitFile {
		return &github.CommitFile{Filename: github.Ptr(name)}
	}

	makeCheck := func(status string, conclusion string) *github.CheckRun {
		return &github.CheckRun{
			Status:     github.Ptr(status),
			Conclusion: github.Ptr(conclusion),
		}
	}

	baseInput := EvaluationInput{
		PullRequest: &github.PullRequest{
			State:     github.Ptr("open"),
			Draft:     github.Ptr(false),
			Additions: github.Ptr(120),
			Deletions: github.Ptr(90),
			User: &github.User{
				Login: github.Ptr("github-copilot[bot]"),
			},
		},
		Commits: []*github.RepositoryCommit{
			makeCommit("github-copilot[bot]"),
			makeCommit("github-copilot[bot]"),
		},
		Files: []*github.CommitFile{
			makeFile("internal/foo/resource_test.go"),
			makeFile("examples/main.tf"),
		},
		CombinedState: "success",
		CheckRuns: []*github.CheckRun{
			makeCheck("completed", "success"),
			makeCheck("completed", "neutral"),
		},
		ApproverLogin: "github-actions[bot]",
		Reviews: []*github.PullRequestReview{
			{
				State: github.Ptr("COMMENTED"),
				User:  &github.User{Login: github.Ptr("github-actions[bot]")},
			},
		},
	}

	tests := []struct {
		name            string
		mutate          func(in *EvaluationInput)
		wantApprove     bool
		wantAlreadySeen bool
		wantReason      string
	}{
		{
			name:        "approves when all gates pass",
			mutate:      func(in *EvaluationInput) {},
			wantApprove: true,
			wantReason:  "all gates passed",
		},
		{
			name: "approves when commits are authored by Copilot app login",
			mutate: func(in *EvaluationInput) {
				in.Commits[0] = makeCommit("Copilot")
				in.Commits[1] = makeCommit("Copilot")
			},
			wantApprove: true,
			wantReason:  "all gates passed",
		},
		{
			name: "approves dependabot without copilot-only gates",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.User = &github.User{Login: github.Ptr("dependabot[bot]")}
				in.Commits = []*github.RepositoryCommit{makeCommit("octocat")}
				in.Files = []*github.CommitFile{makeFile("README.md")}
				in.PullRequest.Additions = github.Ptr(500)
				in.PullRequest.Deletions = github.Ptr(500)
			},
			wantApprove: true,
			wantReason:  "all gates passed",
		},
		{
			name: "rejects non copilot commit author",
			mutate: func(in *EvaluationInput) {
				in.Commits[1] = makeCommit("octocat")
			},
			wantApprove: false,
			wantReason:  "not all commits are authored by allowed Copilot identities",
		},
		{
			name: "rejects disallowed file type",
			mutate: func(in *EvaluationInput) {
				in.Files = append(in.Files, makeFile("README.md"))
			},
			wantApprove: false,
			wantReason:  "pull request contains files outside *_test.go and *.tf",
		},
		{
			name: "rejects exactly 300 edited lines",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.Additions = github.Ptr(200)
				in.PullRequest.Deletions = github.Ptr(100)
			},
			wantApprove: false,
			wantReason:  "edited lines must be < 300",
		},
		{
			name: "rejects over 300 edited lines",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.Additions = github.Ptr(250)
				in.PullRequest.Deletions = github.Ptr(80)
			},
			wantApprove: false,
			wantReason:  "edited lines must be < 300",
		},
		{
			name: "rejects when no category matches",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.User = &github.User{Login: github.Ptr("octocat")}
			},
			wantApprove: false,
			wantReason:  "did not match any auto-approve category",
		},
		{
			name: "rejects failing combined status",
			mutate: func(in *EvaluationInput) {
				in.CombinedState = "failure"
			},
			wantApprove: false,
			wantReason:  "not all checks are successful",
		},
		{
			name: "rejects dependabot when shared checks fail",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.User = &github.User{Login: github.Ptr("dependabot[bot]")}
				in.CombinedState = "failure"
				in.Commits = []*github.RepositoryCommit{makeCommit("octocat")}
				in.Files = []*github.CommitFile{makeFile("README.md")}
			},
			wantApprove: false,
			wantReason:  "not all checks are successful",
		},
		{
			name: "rejects incomplete check run",
			mutate: func(in *EvaluationInput) {
				in.CheckRuns[0] = makeCheck("queued", "")
			},
			wantApprove: false,
			wantReason:  "not all checks are successful",
		},
		{
			name: "rejects failed check run conclusion",
			mutate: func(in *EvaluationInput) {
				in.CheckRuns[0] = makeCheck("completed", "failure")
			},
			wantApprove: false,
			wantReason:  "not all checks are successful",
		},
		{
			name: "skips when approver already approved",
			mutate: func(in *EvaluationInput) {
				in.Reviews = append(in.Reviews, &github.PullRequestReview{
					State: github.Ptr("APPROVED"),
					User:  &github.User{Login: github.Ptr("github-actions[bot]")},
				})
			},
			wantApprove:     false,
			wantAlreadySeen: true,
			wantReason:      "approver has already submitted an approval review",
		},
		{
			name: "rejects non open pull request",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.State = github.Ptr("closed")
			},
			wantApprove: false,
			wantReason:  "pull request is not open",
		},
		{
			name: "rejects draft pull request",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.Draft = github.Ptr(true)
			},
			wantApprove: false,
			wantReason:  "pull request is draft",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			input := cloneInput(baseInput)
			tc.mutate(&input)

			got := Evaluate(input)
			assert.Equal(t, tc.wantApprove, got.ShouldApprove)
			assert.Equal(t, tc.wantAlreadySeen, got.AlreadyApproved)
			require.NotEmpty(t, got.Reasons)
			assert.True(t, hasReasonContaining(got.Reasons, tc.wantReason))
		})
	}
}

func hasReasonContaining(reasons []string, expected string) bool {
	for _, reason := range reasons {
		if strings.Contains(reason, expected) {
			return true
		}
	}
	return false
}

func cloneInput(in EvaluationInput) EvaluationInput {
	out := in

	out.PullRequest = &github.PullRequest{
		State:     github.Ptr(in.PullRequest.GetState()),
		Draft:     github.Ptr(in.PullRequest.GetDraft()),
		Additions: github.Ptr(in.PullRequest.GetAdditions()),
		Deletions: github.Ptr(in.PullRequest.GetDeletions()),
		User: &github.User{
			Login: github.Ptr(in.PullRequest.GetUser().GetLogin()),
		},
	}

	out.Commits = make([]*github.RepositoryCommit, 0, len(in.Commits))
	for _, c := range in.Commits {
		out.Commits = append(out.Commits, &github.RepositoryCommit{
			Author: &github.User{
				Login: github.Ptr(c.Author.GetLogin()),
			},
		})
	}

	out.Files = make([]*github.CommitFile, 0, len(in.Files))
	for _, f := range in.Files {
		out.Files = append(out.Files, &github.CommitFile{
			Filename: github.Ptr(f.GetFilename()),
		})
	}

	out.CheckRuns = make([]*github.CheckRun, 0, len(in.CheckRuns))
	for _, r := range in.CheckRuns {
		out.CheckRuns = append(out.CheckRuns, &github.CheckRun{
			Status:     github.Ptr(r.GetStatus()),
			Conclusion: github.Ptr(r.GetConclusion()),
		})
	}

	out.Reviews = make([]*github.PullRequestReview, 0, len(in.Reviews))
	for _, review := range in.Reviews {
		out.Reviews = append(out.Reviews, &github.PullRequestReview{
			State: github.Ptr(review.GetState()),
			User: &github.User{
				Login: github.Ptr(review.GetUser().GetLogin()),
			},
		})
	}

	return out
}
