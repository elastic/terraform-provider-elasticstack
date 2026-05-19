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

package prcheck

import (
	"context"
	"fmt"
)

// IssueCommentsREST lists and mutates GitHub issue comments (used for PR threads).
type IssueCommentsREST interface {
	ListIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]Comment, error)
	CreateIssueComment(ctx context.Context, owner, repo string, issueNumber int, body string) error
	UpdateIssueComment(ctx context.Context, owner, repo string, commentID int64, body string) error
}

// UpsertVerdictIssueComment mirrors pr-changelog-check/check.js comment upsert semantics.
func UpsertVerdictIssueComment(ctx context.Context, rest IssueCommentsREST, owner, repo string, issueNumber int, verdict Verdict) error {
	marker := MarkerForPRCheck
	raw, err := rest.ListIssueComments(ctx, owner, repo, issueNumber)
	if err != nil {
		return err
	}
	existing := FindExistingComment(raw, marker)

	switch {
	case verdict.NoChangelogSkip:
		if existing != nil {
			body := BuildNoChangelogPassCommentBody(marker)
			return rest.UpdateIssueComment(ctx, owner, repo, existing.ID, body)
		}
	case verdict.Status == StatusFail:
		body := BuildFailureCommentBody(marker, verdict.Errors)
		if existing != nil {
			return rest.UpdateIssueComment(ctx, owner, repo, existing.ID, body)
		}
		return rest.CreateIssueComment(ctx, owner, repo, issueNumber, body)
	case verdict.Status == StatusPass:
		if existing != nil {
			body := BuildPassCommentBody(marker)
			return rest.UpdateIssueComment(ctx, owner, repo, existing.ID, body)
		}
	default:
		return fmt.Errorf("unexpected verdict %+v", verdict)
	}
	return nil
}
