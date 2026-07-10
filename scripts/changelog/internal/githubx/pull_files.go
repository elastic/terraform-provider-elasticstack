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

const pullFilesPageSize = 100

// ListPullRequestFilenames returns each changed file path for pullNumber (paginated).
// Mirrors pulls.listFiles usage in the changelog evidence gather path.
func ListPullRequestFilenames(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	pullNumber int,
) ([]string, error) {
	opts := &github.ListOptions{PerPage: pullFilesPageSize}
	var names []string
	for {
		files, resp, err := client.PullRequests.ListFiles(ctx, owner, repo, pullNumber, opts)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if f != nil && f.GetFilename() != "" {
				names = append(names, f.GetFilename())
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return names, nil
}
