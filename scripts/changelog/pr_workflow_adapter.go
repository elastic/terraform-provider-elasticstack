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

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/githubx"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/prmgmt"
	"github.com/google/go-github/v89/github"
)

// changelogGitHubRESTAdapter binds go-github pulls/issues calls to prmgmt.ChangelogWorkflowREST.
type changelogGitHubRESTAdapter struct {
	inner *githubx.ChangelogWorkflowPullRequests
}

func newChangelogRESTAdapter(client *github.Client) prmgmt.ChangelogWorkflowREST {
	return &changelogGitHubRESTAdapter{
		inner: &githubx.ChangelogWorkflowPullRequests{Client: client},
	}
}

func (a *changelogGitHubRESTAdapter) ListOpenPullRequestsByHead(ctx context.Context, owner, repo, headRef, baseBranch string) ([]prmgmt.PullRequestRef, error) {
	items, err := a.inner.ListOpenPullRequestsByHead(ctx, owner, repo, headRef, baseBranch)
	if err != nil {
		return nil, err
	}
	out := make([]prmgmt.PullRequestRef, len(items))
	for i, item := range items {
		out[i] = prmgmt.PullRequestRef{
			Number: item.Number,
			URL:    item.URL,
		}
	}
	return out, nil
}

func (a *changelogGitHubRESTAdapter) CreatePullRequest(ctx context.Context, owner, repo string, title, body, head, base string) (*prmgmt.PullRequestRef, error) {
	created, err := a.inner.CreatePullRequest(ctx, owner, repo, title, body, head, base)
	if err != nil {
		return nil, err
	}
	return &prmgmt.PullRequestRef{
		Number: created.Number,
		URL:    created.URL,
	}, nil
}

func (a *changelogGitHubRESTAdapter) UpdatePullRequestBody(ctx context.Context, owner, repo string, number int, body string) error {
	return a.inner.UpdatePullRequestBody(ctx, owner, repo, number, body)
}

func (a *changelogGitHubRESTAdapter) AddIssueLabels(ctx context.Context, owner, repo string, issueNumber int, labels []string) error {
	return a.inner.AddIssueLabels(ctx, owner, repo, issueNumber, labels)
}
