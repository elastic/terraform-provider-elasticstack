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
	"strings"
	"testing"

	"github.com/google/go-github/v84/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluate(t *testing.T) {
	t.Parallel()

	makeCommit := func(login string) *github.RepositoryCommit {
		return &github.RepositoryCommit{
			Author: &github.User{Login: new(login)},
		}
	}

	makeFile := func(name string) *github.CommitFile {
		return &github.CommitFile{Filename: new(name)}
	}

	baseInput := EvaluationInput{
		PullRequest: &github.PullRequest{
			State:     new("open"),
			Draft:     new(false),
			Additions: new(120),
			Deletions: new(90),
			User: &github.User{
				Login: new("github-copilot[bot]"),
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
		ApproverLogin: "github-actions[bot]",
		Reviews: []*github.PullRequestReview{
			{
				State: new("COMMENTED"),
				User:  &github.User{Login: new("github-actions[bot]")},
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
			mutate:      func(_ *EvaluationInput) {},
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
				in.PullRequest.User = &github.User{Login: new("dependabot[bot]")}
				in.Commits = []*github.RepositoryCommit{makeCommit("octocat")}
				in.Files = []*github.CommitFile{makeFile("README.md")}
				in.PullRequest.Additions = new(500)
				in.PullRequest.Deletions = new(500)
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
			name: "approves copilot PR with total edits between 300 and 1000",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.Additions = new(400)
				in.PullRequest.Deletions = new(500)
			},
			wantApprove: true,
			wantReason:  "all gates passed",
		},
		{
			name: "rejects exactly 1000 edited lines",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.Additions = new(600)
				in.PullRequest.Deletions = new(400)
			},
			wantApprove: false,
			wantReason:  "edited lines must be < 1000",
		},
		{
			name: "rejects over 1000 edited lines",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.Additions = new(600)
				in.PullRequest.Deletions = new(500)
			},
			wantApprove: false,
			wantReason:  "edited lines must be < 1000",
		},
		{
			name: "rejects when no category matches",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.User = &github.User{Login: new("octocat")}
			},
			wantApprove: false,
			wantReason:  "did not match any auto-approve category",
		},
		{
			name: "skips when approver already approved",
			mutate: func(in *EvaluationInput) {
				in.Reviews = append(in.Reviews, &github.PullRequestReview{
					State: new("APPROVED"),
					User:  &github.User{Login: new("github-actions[bot]")},
				})
			},
			wantApprove:     false,
			wantAlreadySeen: true,
			wantReason:      "approver has already submitted an approval review",
		},
		{
			name: "rejects non open pull request",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.State = new("closed")
			},
			wantApprove: false,
			wantReason:  "pull request is not open",
		},
		{
			name: "approves draft pull request when all gates pass",
			mutate: func(in *EvaluationInput) {
				in.PullRequest.Draft = new(true)
			},
			wantApprove: true,
			wantReason:  "all gates passed",
		},
	}

	for _, tc := range tests {
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
		State:     new(in.PullRequest.GetState()),
		Draft:     new(in.PullRequest.GetDraft()),
		Additions: new(in.PullRequest.GetAdditions()),
		Deletions: new(in.PullRequest.GetDeletions()),
		User: &github.User{
			Login: new(in.PullRequest.GetUser().GetLogin()),
		},
	}

	out.Commits = make([]*github.RepositoryCommit, 0, len(in.Commits))
	for _, c := range in.Commits {
		out.Commits = append(out.Commits, &github.RepositoryCommit{
			Author: &github.User{
				Login: new(c.Author.GetLogin()),
			},
		})
	}

	out.Files = make([]*github.CommitFile, 0, len(in.Files))
	for _, f := range in.Files {
		out.Files = append(out.Files, &github.CommitFile{
			Filename: new(f.GetFilename()),
		})
	}

	out.Reviews = make([]*github.PullRequestReview, 0, len(in.Reviews))
	for _, review := range in.Reviews {
		out.Reviews = append(out.Reviews, &github.PullRequestReview{
			State: new(review.GetState()),
			User: &github.User{
				Login: new(review.GetUser().GetLogin()),
			},
		})
	}

	return out
}
