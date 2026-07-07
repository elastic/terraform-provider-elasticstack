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

const issueCommentsPageSize = 100

// IssueComment summarizes a REST issue comment (PR comments use the issue comments API).
type IssueComment struct {
	ID        int64
	Body      string
	UserLogin string
}

// ListIssueComments lists all comments for issueNumber (paginated).
func ListIssueComments(ctx context.Context, client *github.Client, owner, repo string, issueNumber int) ([]IssueComment, error) {
	opts := &github.IssueListCommentsOptions{ListOptions: github.ListOptions{PerPage: issueCommentsPageSize}}
	var out []IssueComment
	for {
		comments, resp, err := client.Issues.ListComments(ctx, owner, repo, issueNumber, opts)
		if err != nil {
			return nil, err
		}
		for _, ic := range comments {
			if ic == nil {
				continue
			}
			body := ""
			if ic.Body != nil {
				body = *ic.Body
			}
			login := ""
			if u := ic.GetUser(); u != nil {
				login = u.GetLogin()
			}
			out = append(out, IssueComment{
				ID:        ic.GetID(),
				Body:      body,
				UserLogin: login,
			})
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return out, nil
}

// CreateIssueComment adds a comment to issueNumber.
func CreateIssueComment(ctx context.Context, client *github.Client, owner, repo string, issueNumber int, body string) error {
	bodyCopy := body
	in := &github.IssueComment{Body: &bodyCopy}
	_, _, err := client.Issues.CreateComment(ctx, owner, repo, issueNumber, in)
	return err
}

// UpdateIssueComment edits an existing issue comment's body.
func UpdateIssueComment(ctx context.Context, client *github.Client, owner, repo string, commentID int64, body string) error {
	bodyCopy := body
	edit := &github.IssueComment{Body: &bodyCopy}
	_, _, err := client.Issues.EditComment(ctx, owner, repo, commentID, edit)
	return err
}
