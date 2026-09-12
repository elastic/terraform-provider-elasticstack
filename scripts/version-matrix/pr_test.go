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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

type stubStandingREST struct {
	listResult []PullRequestRef
	listErr    error
	createErr  error
	updateErr  error
	labelErr   error

	listCalls   int
	createCalls []pullCreateCall
	updateCalls []pullUpdateCall
	labelCalls  []labelCall
}

type pullUpdateCall struct {
	number int
	body   string
}

func (s *stubStandingREST) ListOpenPullRequestsByHead(context.Context, string, string, string, string) ([]PullRequestRef, error) {
	s.listCalls++
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.listResult, nil
}

func (s *stubStandingREST) CreatePullRequest(_ context.Context, _, _ string, title, body, head, base string) (*PullRequestRef, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	s.createCalls = append(s.createCalls, pullCreateCall{title: title, body: body, head: head, base: base})
	return &PullRequestRef{Number: 7, URL: "https://github.com/org/repo/pull/7"}, nil
}

func (s *stubStandingREST) UpdatePullRequestBody(_ context.Context, _, _ string, number int, body string) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updateCalls = append(s.updateCalls, pullUpdateCall{number: number, body: body})
	return nil
}

func (s *stubStandingREST) AddIssueLabels(_ context.Context, _, _ string, issueNumber int, labels []string) error {
	s.labelCalls = append(s.labelCalls, labelCall{issue: issueNumber, tags: append([]string(nil), labels...)})
	return s.labelErr
}

func TestBuildStandingPRBody_statesAllOrNothingMergeIntent(t *testing.T) {
	t.Parallel()

	body := BuildStandingPRBody("2026-09-11")
	assert.Contains(t, body, "**Generated:** 2026-09-11")
	assert.Contains(t, body, "Merging this pull request makes the default branch's pinned acceptance-test version list")
	assert.Contains(t, body, "equal the full computed list")
	assert.Contains(t, body, "all-or-nothing")
	assert.Contains(t, body, "a newly-red GA version holds the")
	assert.Contains(t, body, "whole delta rather than being silently dropped")
	assert.Contains(t, body, "Do not make manual edits to the `acceptance-test-version-matrix` branch.")
}

func TestManageStandingPR_missingPR_creates(t *testing.T) {
	t.Parallel()

	st := &stubStandingREST{}
	res, err := ManageStandingPR(context.Background(), ManageStandingOptions{
		Changed: true,
		Owner:   "org",
		Repo:    "repo",
		GitHub:  st,
		Now:     fixedClock(time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)
	assert.Equal(t, standingPRActionCreated, res.Action)
	assert.Equal(t, 7, res.Number)
	assert.Equal(t, "https://github.com/org/repo/pull/7", res.URL)
	require.Len(t, st.createCalls, 1)
	assert.Equal(t, standingPRTitle, st.createCalls[0].title)
	assert.Equal(t, standingPRHeadBranch, st.createCalls[0].head)
	assert.Equal(t, defaultPullRequestBaseBranch, st.createCalls[0].base)
	assert.Contains(t, st.createCalls[0].body, "all-or-nothing")
	assert.Empty(t, st.updateCalls)
	require.Len(t, st.labelCalls, 1)
	assert.Equal(t, 7, st.labelCalls[0].issue)
	assert.Equal(t, []string{noChangelogLabel}, st.labelCalls[0].tags)
}

func TestManageStandingPR_labelsFailAfterCreate_warnsWithoutError(t *testing.T) {
	t.Parallel()

	st := &stubStandingREST{labelErr: assert.AnError}
	res, err := ManageStandingPR(context.Background(), ManageStandingOptions{
		Changed: true,
		Owner:   "org",
		Repo:    "repo",
		GitHub:  st,
		Now:     fixedClock(time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)
	assert.Equal(t, standingPRActionCreated, res.Action)
	require.Len(t, res.Warnings, 1)
	assert.Contains(t, res.Warnings[0], "Failed to apply no-changelog label to PR #7")
}

func TestManageStandingPR_unchanged_doesNotTouchExistingPR(t *testing.T) {
	t.Parallel()

	st := &stubStandingREST{
		listResult: []PullRequestRef{{Number: 42, URL: "https://github.com/org/repo/pull/42"}},
	}
	res, err := ManageStandingPR(context.Background(), ManageStandingOptions{
		Changed: false,
		Owner:   "org",
		Repo:    "repo",
		GitHub:  st,
		Now:     fixedClock(time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)
	assert.Equal(t, standingPRActionNoop, res.Action)
	assert.Zero(t, res.Number)
	assert.Empty(t, res.URL)
	assert.Zero(t, st.listCalls)
	assert.Empty(t, st.createCalls)
	assert.Empty(t, st.updateCalls)
	assert.Empty(t, st.labelCalls)
}

func TestManageStandingPR_existingOpen_updatesBody(t *testing.T) {
	t.Parallel()

	st := &stubStandingREST{
		listResult: []PullRequestRef{{Number: 42, URL: "https://github.com/org/repo/pull/42"}},
	}
	res, err := ManageStandingPR(context.Background(), ManageStandingOptions{
		Changed: true,
		Owner:   "org",
		Repo:    "repo",
		GitHub:  st,
		Now:     fixedClock(time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)
	assert.Equal(t, standingPRActionUpdated, res.Action)
	assert.Equal(t, 42, res.Number)
	assert.Equal(t, "https://github.com/org/repo/pull/42", res.URL)
	require.Len(t, st.updateCalls, 1)
	assert.Equal(t, 42, st.updateCalls[0].number)
	assert.Contains(t, st.updateCalls[0].body, "all-or-nothing")
	assert.Empty(t, st.createCalls)
	require.Len(t, st.labelCalls, 1)
	assert.Equal(t, 42, st.labelCalls[0].issue)
	assert.Equal(t, []string{noChangelogLabel}, st.labelCalls[0].tags)
}

func fixedClock(ts time.Time) func() time.Time {
	return func() time.Time { return ts }
}
