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

	"github.com/google/go-github/v89/github"
)

const pullAssociationsPageSize = 100

// PullRequestsAssociatedWithCommit lists merged/closed associations for sha (paginated).
func PullRequestsAssociatedWithCommit(
	ctx context.Context,
	client *github.Client,
	owner, repo, sha string,
) ([]*github.PullRequest, error) {
	opts := &github.ListOptions{PerPage: pullAssociationsPageSize}
	var all []*github.PullRequest
	for {
		slice, resp, err := client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, sha, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, slice...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}
