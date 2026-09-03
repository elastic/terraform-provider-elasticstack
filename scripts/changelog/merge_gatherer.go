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
	"fmt"

	"github.com/google/go-github/v89/github"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/githubx"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/section"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/semver"
)

type gitMergedPRGatherer struct {
	client *github.Client
	execer semver.Execer
}

func (g *gitMergedPRGatherer) GatherMergedPRs(
	ctx context.Context, owner, repo, compareRange string,
) ([]section.MergedPR, []string, error) {
	shas, err := githubx.ListCommitSHAs(g.execer, compareRange)
	warnMsgs := []string{}
	if err != nil {
		shas = nil
		warnMsgs = append(warnMsgs, fmt.Sprintf("Failed to list commits in range: %v", err))
	}

	byNum := make(map[int]*github.PullRequest)
	var order []int

	for _, sha := range shas {
		prs, prErr := githubx.PullRequestsAssociatedWithCommit(ctx, g.client, owner, repo, sha)
		if prErr != nil {
			warnMsgs = append(warnMsgs, fmt.Sprintf("Failed to list PRs for commit %s: %v", sha, prErr))
			continue
		}

		for _, pr := range prs {
			if pr == nil {
				continue
			}
			if pr.GetState() != "closed" || pr.MergedAt == nil {
				continue
			}
			n := pr.GetNumber()
			if _, dup := byNum[n]; dup {
				continue
			}
			byNum[n] = pr
			order = append(order, n)
		}
	}

	out := make([]section.MergedPR, 0, len(order))
	for _, n := range order {
		pr := byNum[n]
		var labels []string
		for _, l := range pr.Labels {
			if l != nil && l.GetName() != "" {
				labels = append(labels, l.GetName())
			}
		}
		author := ""
		if u := pr.GetUser(); u != nil {
			author = u.GetLogin()
		}
		out = append(out, section.MergedPR{
			Number:         n,
			Title:          pr.GetTitle(),
			URL:            pr.GetHTMLURL(),
			Labels:         labels,
			Body:           pr.GetBody(),
			MergeCommitSHA: pr.GetMergeCommitSHA(),
			AuthorLogin:    author,
		})
	}
	return out, warnMsgs, nil
}
