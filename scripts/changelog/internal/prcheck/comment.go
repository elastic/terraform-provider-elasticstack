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

import "strings"

// MarkerForPRCheck is embedded in verdict comments so later runs locate the same comment.
const MarkerForPRCheck = "<!-- pr-changelog-check -->"

const githubActionsBotLogin = "github-actions[bot]"

// Comment is a minimal snapshot of an issue comment for upsert lookups.
type Comment struct {
	ID        int64
	Body      string
	UserLogin string
}

// FindExistingComment returns the first github-actions[bot] comment containing marker, or nil.
func FindExistingComment(comments []Comment, marker string) *Comment {
	for i := range comments {
		c := &comments[i]
		if strings.TrimSpace(c.UserLogin) != githubActionsBotLogin {
			continue
		}
		if strings.Contains(c.Body, marker) {
			return &comments[i]
		}
	}
	return nil
}

// BuildPassCommentBody matches the legacy pass comment template verbatim.
func BuildPassCommentBody(marker string) string {
	return marker + "\n:white_check_mark: **PR Changelog Check passed** — the `## Changelog` section looks good."
}

// BuildNoChangelogPassCommentBody matches the legacy no-changelog pass comment template verbatim.
func BuildNoChangelogPassCommentBody(marker string) string {
	return marker + "\n:white_check_mark: **PR Changelog Check passed** — `no-changelog` label is set."
}

// BuildFailureCommentBody matches the legacy failure comment template (byte-compatible).
func BuildFailureCommentBody(marker string, errs []string) string {
	errLines := make([]string, len(errs))
	for i, e := range errs {
		errLines[i] = "- " + e
	}
	errorList := strings.Join(errLines, "\n")
	lines := []string{
		marker,
		":x: **PR Changelog Check failed**",
		"",
		"The following issues were found with the `## Changelog` section:",
		"",
		errorList,
		"",
		"<details>",
		"<summary>Expected format</summary>",
		"",
		"```",
		"## Changelog",
		"Customer impact: <none|fix|enhancement|breaking>",
		"Summary: <one-line description>  (required when Customer impact is not \"none\")",
		"",
		"### Breaking changes",
		"<free-form markdown>  (required when Customer impact is \"breaking\")",
		"<!-- /breaking-changes -->  (optional — ends the block early)",
		"```",
		"",
		"Or add the `no-changelog` label to bypass this check.",
		"</details>",
	}
	return strings.Join(lines, "\n")
}
