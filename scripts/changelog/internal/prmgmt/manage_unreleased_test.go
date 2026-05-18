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

package prmgmt_test

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/prmgmt"
)

func TestBuildUnreleasedPRBody_generatedLabel(t *testing.T) {
	t.Parallel()
	body := prmgmt.BuildUnreleasedPRBody("v1.0.0...HEAD", "2006-01-02")
	ok, err := regexp.MatchString(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2}`, body)
	if err != nil || !ok {
		t.Fatalf("expected Generated date pattern, body=%q", body)
	}
}

func TestBuildUnreleasedPRBody_compareRangeLine(t *testing.T) {
	t.Parallel()
	body := prmgmt.BuildUnreleasedPRBody("v1.0.0...HEAD", "2006-01-02")
	if !strings.Contains(body, "**Compare range:** `v1.0.0...HEAD`") {
		t.Fatalf("compare range line missing: %q", body)
	}
}

func TestBuildUnreleasedPRBody_manualEditsNotice(t *testing.T) {
	t.Parallel()
	body := prmgmt.BuildUnreleasedPRBody("v1.0.0...HEAD", "2006-01-02")
	want := "Do not make manual edits to the `generated-changelog` branch."
	if !strings.Contains(body, want) {
		t.Fatalf("expected notice %q in %q", want, body)
	}
}

type pullUpdateCall struct {
	number int
	body   string
}

type pullCreateCall struct {
	title string
	body  string
	head  string
	base  string
}

type labelCall struct {
	issue int
	tags  []string
}

type stubChangelogREST struct {
	listResult []prmgmt.PullRequestRef
	listErr    error
	updateErr  error
	createErr  error

	updateCalls []pullUpdateCall
	createCalls []pullCreateCall
	labelCalls  []labelCall

	labelErr error
}

func (s *stubChangelogREST) ListOpenPullRequestsByHead(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.listResult, nil
}

func (s *stubChangelogREST) CreatePullRequest(_ context.Context, _, _ string, title, body, head, base string) (*prmgmt.PullRequestRef, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	s.createCalls = append(s.createCalls, pullCreateCall{title: title, body: body, head: head, base: base})
	return &prmgmt.PullRequestRef{Number: 7, URL: "https://github.com/org/repo/pull/7"}, nil
}

func (s *stubChangelogREST) UpdatePullRequestBody(_ context.Context, _, _ string, number int, body string) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updateCalls = append(s.updateCalls, pullUpdateCall{number: number, body: body})
	return nil
}

func (s *stubChangelogREST) AddIssueLabels(_ context.Context, _, _ string, issueNumber int, labels []string) error {
	s.labelCalls = append(s.labelCalls, labelCall{issue: issueNumber, tags: append([]string(nil), labels...)})
	return s.labelErr
}

func TestManageUnreleasedPR_existingOpen_updatesBody(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &stubChangelogREST{
		listResult: []prmgmt.PullRequestRef{{Number: 42, URL: "https://github.com/org/repo/pull/42"}},
	}

	res, err := prmgmt.ManageUnreleasedPR(ctx, prmgmt.ManageUnreleasedOptions{
		Owner:        "org",
		Repo:         "repo",
		CompareRange: "v1.0.0...HEAD",
		GitHub:       st,
		Now:          fixedClock(time.Date(2020, 5, 1, 15, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != "updated" || res.Number != 42 || res.URL != "https://github.com/org/repo/pull/42" {
		t.Fatalf("unexpected result: %+v", res)
	}
	if len(st.updateCalls) != 1 {
		t.Fatalf("expected one update, got %#v", st.updateCalls)
	}
	if st.updateCalls[0].number != 42 || !strings.Contains(st.updateCalls[0].body, "Do not make manual edits") {
		t.Fatalf("unexpected update %#v", st.updateCalls[0])
	}
	if len(st.createCalls) != 0 {
		t.Fatalf("create unexpectedly called %#v", st.createCalls)
	}
	if len(st.labelCalls) != 1 || st.labelCalls[0].issue != 42 {
		t.Fatalf("labels: %#v", st.labelCalls)
	}
	if len(st.labelCalls[0].tags) != 1 || st.labelCalls[0].tags[0] != "no-changelog" {
		t.Fatalf("unexpected labels: %#v", st.labelCalls)
	}
}

func TestManageUnreleasedPR_missingPR_creates(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &stubChangelogREST{listResult: nil}

	res, err := prmgmt.ManageUnreleasedPR(ctx, prmgmt.ManageUnreleasedOptions{
		Owner:        "org",
		Repo:         "repo",
		CompareRange: "v0.9.0...HEAD",
		GitHub:       st,
		Now:          fixedClock(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != "created" || res.Number != 7 {
		t.Fatalf("unexpected %+v", res)
	}
	if len(st.createCalls) != 1 {
		t.Fatalf("expected one create %#v", st.createCalls)
	}
	c := st.createCalls[0]
	if c.head != "generated-changelog" || c.base != testPullRequestMainBase || c.title != "chore: update CHANGELOG.md [Unreleased] section" {
		t.Fatalf("unexpected create %+v", c)
	}
	if len(st.updateCalls) != 0 {
		t.Fatalf("unexpected updates %#v", st.updateCalls)
	}
	if len(st.labelCalls) != 1 || st.labelCalls[0].issue != 7 {
		t.Fatalf("labels %#v", st.labelCalls)
	}
}

func TestManageUnreleasedPR_labelsFailAfterCreate_warnsWithoutError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &stubChangelogREST{listResult: nil, labelErr: errors.New("boom")}

	res, err := prmgmt.ManageUnreleasedPR(ctx, prmgmt.ManageUnreleasedOptions{
		Owner:        "org",
		Repo:         "repo",
		CompareRange: "v0.9.0...HEAD",
		GitHub:       st,
		Now:          fixedClock(time.Date(2020, 5, 1, 15, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != "created" || res.Number != 7 {
		t.Fatalf("unexpected %+v", res)
	}
	if len(res.Warnings) != 1 {
		t.Fatalf("warnings: %#v", res.Warnings)
	}
	ok, rerr := regexp.MatchString(`Failed to apply no-changelog label to PR #7: boom`, res.Warnings[0])
	if rerr != nil || !ok {
		t.Fatalf("unexpected warning %q", res.Warnings[0])
	}
}

func TestManageUnreleasedPR_labelsFailAfterUpdate_warnsWithoutError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &stubChangelogREST{
		listResult: []prmgmt.PullRequestRef{{Number: 42, URL: "https://github.com/org/repo/pull/42"}},
		labelErr:   errors.New("boom"),
	}

	res, err := prmgmt.ManageUnreleasedPR(ctx, prmgmt.ManageUnreleasedOptions{
		Owner:        "org",
		Repo:         "repo",
		CompareRange: "v1.0.0...HEAD",
		GitHub:       st,
		Now:          fixedClock(time.Date(2020, 6, 1, 15, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != "updated" {
		t.Fatalf("got %+v", res)
	}
	if len(st.updateCalls) != 1 {
		t.Fatalf("update calls %#v", st.updateCalls)
	}
	if len(res.Warnings) != 1 {
		t.Fatalf("warnings %#v", res.Warnings)
	}
	ok, rerr := regexp.MatchString(`Failed to apply no-changelog label to PR #42: boom`, res.Warnings[0])
	if rerr != nil || !ok {
		t.Fatalf("unexpected warning %q", res.Warnings[0])
	}
}

func fixedClock(ts time.Time) func() time.Time {
	return func() time.Time { return ts }
}

func TestManageUnreleasedPR_listError_returnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	listErr := errors.New("list failed")
	st := &stubChangelogREST{listErr: listErr}

	_, err := prmgmt.ManageUnreleasedPR(ctx, prmgmt.ManageUnreleasedOptions{
		Owner: "o", Repo: "r", GitHub: st, Now: fixedClock(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
	})
	if err == nil || !strings.Contains(err.Error(), "list open pull requests") {
		t.Fatalf("got err=%v", err)
	}
	if !errors.Is(err, listErr) {
		t.Fatalf("expected wrapped listErr, got %v", err)
	}
	if len(st.createCalls) != 0 || len(st.updateCalls) != 0 {
		t.Fatalf("unexpected side effects %#v %#v", st.createCalls, st.updateCalls)
	}
}

func TestManageUnreleasedPR_updateExistingError_returnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	updErr := errors.New("update failed")
	st := &stubChangelogREST{
		listResult: []prmgmt.PullRequestRef{{Number: 1, URL: "u"}},
		updateErr:  updErr,
	}

	_, err := prmgmt.ManageUnreleasedPR(ctx, prmgmt.ManageUnreleasedOptions{
		Owner: "o", Repo: "r", GitHub: st, Now: fixedClock(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)),
	})
	if err == nil || !strings.Contains(err.Error(), "update pull request body") {
		t.Fatalf("got err=%v", err)
	}
	if !errors.Is(err, updErr) {
		t.Fatalf("expected wrapped updErr, got %v", err)
	}
}

func TestManageUnreleasedPR_createError_returnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	crErr := errors.New("create failed")
	st := &stubChangelogREST{listResult: nil, createErr: crErr}

	_, err := prmgmt.ManageUnreleasedPR(ctx, prmgmt.ManageUnreleasedOptions{
		Owner: "o", Repo: "r", GitHub: st, Now: fixedClock(time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)),
	})
	if err == nil || !strings.Contains(err.Error(), "create pull request") {
		t.Fatalf("got err=%v", err)
	}
	if !errors.Is(err, crErr) {
		t.Fatalf("expected wrapped crErr, got %v", err)
	}
	if len(st.createCalls) != 0 {
		t.Fatalf("create should fail before recording success path")
	}
}
