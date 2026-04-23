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
	"os/exec"
	"strings"

	"github.com/google/go-github/v85/github"
)

func listCommitShasInRange(compareRange string) ([]string, error) {
	rangeArg := compareRange
	if rangeArg == "" {
		rangeArg = "HEAD"
	}
	cmd := exec.Command("git", "log", "--format=%H", rangeArg)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		sha := strings.TrimSpace(line)
		if sha != "" {
			result = append(result, sha)
		}
	}
	return result, nil
}

func labelNames(labels []*github.Label) []string {
	result := make([]string, 0, len(labels))
	for _, label := range labels {
		if label != nil && label.Name != nil {
			result = append(result, *label.Name)
		}
	}
	return result
}

func resolveMergedPullRequests(ctx context.Context, client *github.Client, owner, repo string, commitShas []string) ([]pullRequestRecord, error) {
	seen := map[int]bool{}
	result := make([]pullRequestRecord, 0)
	for _, sha := range commitShas {
		prs, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, sha, nil)
		if err != nil {
			return nil, err
		}
		for _, pr := range prs {
			if pr == nil || seen[pr.GetNumber()] || pr.GetState() != "closed" || pr.MergedAt == nil {
				continue
			}
			seen[pr.GetNumber()] = true
			result = append(result, pullRequestRecord{
				Number:         pr.GetNumber(),
				Title:          pr.GetTitle(),
				URL:            pr.GetHTMLURL(),
				MergeCommitSHA: pr.GetMergeCommitSHA(),
				Author: func() string {
					if pr.User != nil {
						return pr.User.GetLogin()
					}
					return "unknown"
				}(),
				Labels: labelNames(pr.Labels),
				Body:   pr.GetBody(),
			})
		}
	}
	return result, nil
}
