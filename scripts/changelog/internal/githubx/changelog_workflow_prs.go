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
	"context"
	"fmt"

	"github.com/google/go-github/v89/github"
)

// ChangelogPullSummary captures number and HTML URL returned from pull request APIs.
type ChangelogPullSummary struct {
	Number int
	URL    string
}

// ChangelogWorkflowPullRequests implements GitHub pulls/issues calls used by
// changelog-generation workflow helpers (changelog-pr-management.js parity).
type ChangelogWorkflowPullRequests struct {
	Client *github.Client
}

// ListOpenPullRequestsByHead lists open pulls where Head is owner:branchSlug and Base is baseBranch.
func (w *ChangelogWorkflowPullRequests) ListOpenPullRequestsByHead(ctx context.Context, owner, repo, headRef, baseBranch string) ([]ChangelogPullSummary, error) {
	if w.Client == nil {
		return nil, fmt.Errorf("github client required")
	}
	opts := &github.PullRequestListOptions{
		State: "open",
		Head:  headRef,
		Base:  baseBranch,
	}
	pulls, _, err := w.Client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("list pull requests: %w", err)
	}
	out := make([]ChangelogPullSummary, 0, len(pulls))
	for _, pr := range pulls {
		out = append(out, ChangelogPullSummary{
			Number: pr.GetNumber(),
			URL:    pr.GetHTMLURL(),
		})
	}
	return out, nil
}

// CreatePullRequest opens a pull request (mirrors pulls.create inputs used by changelog workflow).
func (w *ChangelogWorkflowPullRequests) CreatePullRequest(ctx context.Context, owner, repo string, title, body, head, base string) (*ChangelogPullSummary, error) {
	if w.Client == nil {
		return nil, fmt.Errorf("github client required")
	}
	titleCopy, headCopy, baseCopy, bodyCopy := title, head, base, body
	newPR := &github.NewPullRequest{
		Title: &titleCopy,
		Head:  &headCopy,
		Base:  &baseCopy,
		Body:  &bodyCopy,
	}
	pr, _, err := w.Client.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("create pull request: %w", err)
	}
	return &ChangelogPullSummary{
		Number: pr.GetNumber(),
		URL:    pr.GetHTMLURL(),
	}, nil
}

// UpdatePullRequestBody sets the pull body (mirrors pulls.update body-only updates).
func (w *ChangelogWorkflowPullRequests) UpdatePullRequestBody(ctx context.Context, owner, repo string, number int, body string) error {
	if w.Client == nil {
		return fmt.Errorf("github client required")
	}
	bodyCopy := body
	req := &github.PullRequest{Body: &bodyCopy}
	if _, _, err := w.Client.PullRequests.Edit(ctx, owner, repo, number, req); err != nil {
		return fmt.Errorf("edit pull request body: %w", err)
	}
	return nil
}

// AddIssueLabels attaches labels to a pull (issues.addLabels parity).
func (w *ChangelogWorkflowPullRequests) AddIssueLabels(ctx context.Context, owner, repo string, issueNumber int, labels []string) error {
	if w.Client == nil {
		return fmt.Errorf("github client required")
	}
	if _, _, err := w.Client.Issues.AddLabelsToIssue(ctx, owner, repo, issueNumber, labels); err != nil {
		return fmt.Errorf("add labels to pull request #%d: %w", issueNumber, err)
	}
	return nil
}
