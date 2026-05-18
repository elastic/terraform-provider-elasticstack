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
	_ "embed"
	"strings"
	"testing"
)

//go:embed testdata/failure_comment_empty.want
var failureCommentEmptyWant []byte

//go:embed testdata/failure_comment_one_error.want
var failureCommentOneErrorWant []byte

func TestFindExistingComment(t *testing.T) {
	t.Parallel()

	marker := MarkerForPRCheck
	first := Comment{ID: 1, UserLogin: githubActionsBotLogin, Body: marker + "\nfirst"}
	second := Comment{ID: 2, UserLogin: githubActionsBotLogin, Body: marker + "\nsecond"}

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		if FindExistingComment(nil, marker) != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("no bot marker", func(t *testing.T) {
		t.Parallel()
		comments := []Comment{
			{UserLogin: githubActionsBotLogin, Body: "some other content"},
			{UserLogin: "octocat", Body: marker},
		}
		if FindExistingComment(comments, marker) != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("bot match wins", func(t *testing.T) {
		t.Parallel()
		comments := []Comment{
			{UserLogin: "octocat", Body: marker},
			first,
		}
		if got := FindExistingComment(comments, marker); got == nil || got.ID != first.ID {
			t.Fatalf("got %+v want id %d", got, first.ID)
		}
	})

	t.Run("non-bot ignores marker", func(t *testing.T) {
		t.Parallel()
		comments := []Comment{
			{UserLogin: "dependabot[bot]", Body: marker + "\nsome"},
			{UserLogin: "octocat", Body: marker + "\nsome"},
		}
		if FindExistingComment(comments, marker) != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("bot without marker substring", func(t *testing.T) {
		t.Parallel()
		comments := []Comment{
			{UserLogin: githubActionsBotLogin, Body: "no marker here"},
		}
		if FindExistingComment(comments, marker) != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("first of two matches", func(t *testing.T) {
		t.Parallel()
		comments := []Comment{first, second}
		if got := FindExistingComment(comments, marker); got == nil || got.ID != first.ID {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("ghost user ignored", func(t *testing.T) {
		t.Parallel()
		comments := []Comment{
			{UserLogin: "", Body: marker + "\nx"},
		}
		if FindExistingComment(comments, marker) != nil {
			t.Fatal("expected nil")
		}
	})
}

func TestBuildPassCommentBody(t *testing.T) {
	t.Parallel()
	marker := "<!-- z -->"
	body := BuildPassCommentBody(marker)
	if !strings.Contains(body, marker) {
		t.Fatal("missing marker")
	}
	if !strings.Contains(body, "passed") {
		t.Fatal("missing passed")
	}
	if !strings.Contains(body, "## Changelog") {
		t.Fatal("missing ## Changelog")
	}
}

func TestBuildNoChangelogPassCommentBody(t *testing.T) {
	t.Parallel()
	marker := "<!-- z -->"
	body := BuildNoChangelogPassCommentBody(marker)
	if !strings.Contains(body, marker) || !strings.Contains(body, "passed") || !strings.Contains(body, "no-changelog") {
		t.Fatalf("body=%q", body)
	}
}

func TestBuildFailureCommentBody(t *testing.T) {
	t.Parallel()

	marker := MarkerForPRCheck

	t.Run("parity empty errors", func(t *testing.T) {
		t.Parallel()
		got := BuildFailureCommentBody(marker, []string{})
		if got != string(failureCommentEmptyWant) {
			t.Fatalf("unexpected body:\n%s", got)
		}
	})

	t.Run("parity one error", func(t *testing.T) {
		t.Parallel()
		got := BuildFailureCommentBody(marker, []string{"only one error"})
		if got != string(failureCommentOneErrorWant) {
			t.Fatalf("unexpected body:\n%s", got)
		}
	})

	t.Run("includes marker bullets and hints", func(t *testing.T) {
		t.Parallel()
		body := BuildFailureCommentBody(marker, []string{"some error"})
		for _, needle := range []string{marker, "- some error", "<details>", "Expected format", "Customer impact:", "no-changelog"} {
			if !strings.Contains(body, needle) {
				t.Fatalf("missing %q in body", needle)
			}
		}
	})

	t.Run("bullet counts single and multi", func(t *testing.T) {
		t.Parallel()

		body1 := BuildFailureCommentBody(marker, []string{"only one error"})
		if countBulletLines(body1) != 1 {
			t.Fatalf("want 1 bullet line, body=%s", body1)
		}

		errs := []string{"error one", "error two", "error three"}
		body3 := BuildFailureCommentBody(marker, errs)
		if countBulletLines(body3) != 3 {
			t.Fatalf("want 3 bullet lines, body=%s", body3)
		}
	})
}

func countBulletLines(s string) int {
	n := 0
	for line := range strings.SplitSeq(s, "\n") {
		if strings.HasPrefix(line, "- ") {
			n++
		}
	}
	return n
}
