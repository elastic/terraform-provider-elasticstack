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

func TestBuildReleasePRBody_generatedVersionCompare(t *testing.T) {
	t.Parallel()
	body := prmgmt.BuildReleasePRBody("2.0.0", "v1.0.0...v2.0.0", "2006-01-02")
	ok, err := regexp.MatchString(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2}`, body)
	if err != nil || !ok {
		t.Fatalf("generated header: %q", body)
	}
	if !strings.Contains(body, "**Version:** `2.0.0`") || !strings.Contains(body, "**Compare range:** `v1.0.0...v2.0.0`") {
		t.Fatalf("body: %q", body)
	}
}

type spyListREST struct {
	listHead string
	listBase string
	listFns  []func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error)

	updateBodies []updateStub
	labelCalls   []pullLabelCall
	labelErr     error
	updateErr    error
}

type pullLabelCall struct {
	issue int
	tags  []string
}

func (s *spyListREST) popList(ctx context.Context, owner, repo, headRef, base string) ([]prmgmt.PullRequestRef, error) {
	s.listHead = headRef
	s.listBase = base
	if len(s.listFns) == 0 {
		return nil, errors.New("listFn not configured")
	}
	fn := s.listFns[0]
	s.listFns = s.listFns[1:]
	return fn(ctx, owner, repo, headRef, base)
}

func (s *spyListREST) ListOpenPullRequestsByHead(ctx context.Context, owner, repo, headRef, baseBranch string) ([]prmgmt.PullRequestRef, error) {
	return s.popList(ctx, owner, repo, headRef, baseBranch)
}

func (s *spyListREST) CreatePullRequest(context.Context, string, string, string, string, string, string) (*prmgmt.PullRequestRef, error) {
	return nil, errors.New("unexpected create")
}

type updateStub struct {
	number int
	body   string
}

func (s *spyListREST) UpdatePullRequestBody(_ context.Context, _, _ string, number int, body string) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updateBodies = append(s.updateBodies, updateStub{number: number, body: body})
	return nil
}

func (s *spyListREST) AddIssueLabels(_ context.Context, _, _ string, issueNumber int, labels []string) error {
	s.labelCalls = append(s.labelCalls, pullLabelCall{issue: issueNumber, tags: append([]string(nil), labels...)})
	return s.labelErr
}

func TestRefreshReleasePR_prNumber_updatesBodyAndLabels(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &spyListREST{}

	res, err := prmgmt.RefreshReleasePR(ctx, prmgmt.RefreshReleaseOptions{
		Owner:         "org",
		Repo:          "repo",
		PRNumber:      55,
		CompareRange:  "v1.0.0...v2.0.0",
		TargetVersion: "2.0.0",
		GitHub:        st,
		Now:           fixedClock(time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Updated || res.Number != 55 || len(st.updateBodies) != 1 || st.updateBodies[0].number != 55 {
		t.Fatalf("update state: %+v %#v", res, st.updateBodies)
	}
	b := st.updateBodies[0].body
	if !strings.Contains(b, "**Version:** `2.0.0`") || !strings.Contains(b, "**Compare range:** `v1.0.0...v2.0.0`") {
		t.Fatalf("unexpected body %q", b)
	}
	ok, rerr := regexp.MatchString(`\*\*Generated:\*\* \d{4}-\d{2}-\d{2}`, b)
	if rerr != nil || !ok {
		t.Fatalf("missing generated metadata: %q", b)
	}
	if len(st.labelCalls) != 1 || st.labelCalls[0].issue != 55 || len(st.labelCalls[0].tags) != 1 {
		t.Fatalf("labels %#v", st.labelCalls)
	}
	if len(res.Infos) != 2 ||
		res.Infos[0] != "Refreshing release PR #55 metadata" ||
		res.Infos[1] != "Release PR #55 metadata refreshed" {
		t.Fatalf("unexpected infos %#v", res.Infos)
	}
}

func TestRefreshReleasePR_labelFailureWarns_afterUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &spyListREST{labelErr: errors.New("label failed")}

	res, err := prmgmt.RefreshReleasePR(ctx, prmgmt.RefreshReleaseOptions{
		Owner:         "org",
		Repo:          "repo",
		PRNumber:      55,
		CompareRange:  "v1.0.0...v2.0.0",
		TargetVersion: "2.0.0",
		GitHub:        st,
		Now:           fixedClock(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Updated || len(st.updateBodies) != 1 {
		t.Fatalf("expected update %+v %#v", res, st.updateBodies)
	}
	if len(res.Warnings) != 1 {
		t.Fatalf("warnings %#v", res.Warnings)
	}
	ok, rerr := regexp.MatchString(`Failed to apply no-changelog label to PR #55: label failed`, res.Warnings[0])
	if rerr != nil || !ok {
		t.Fatalf("unexpected warning %q", res.Warnings[0])
	}
	if len(res.Infos) != 2 ||
		res.Infos[0] != "Refreshing release PR #55 metadata" ||
		res.Infos[1] != "Release PR #55 metadata refreshed" {
		t.Fatalf("unexpected infos %#v", res.Infos)
	}
}

func TestFindOpenReleasePrepPRNumber_resolvesFirstOpen(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &spyListREST{
		listFns: []func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error){
			func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error) {
				return []prmgmt.PullRequestRef{{Number: 88}}, nil
			},
		},
	}

	num, err := prmgmt.FindOpenReleasePrepPRNumber(ctx, st, "org", "repo", testPullRequestMainBase, "1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	if num != 88 {
		t.Fatalf("got %d", num)
	}
	if st.listHead != "org:prep-release-1.2.3" || st.listBase != testPullRequestMainBase {
		t.Fatalf("unexpected list args head=%q base=%q", st.listHead, st.listBase)
	}
}

func TestFindOpenReleasePrepPRNumber_emptyVersionSkipsList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &spyListREST{
		listFns: []func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error){
			func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error) {
				t.Fatal("list should not be called")
				return nil, nil
			},
		},
	}

	num, err := prmgmt.FindOpenReleasePrepPRNumber(ctx, st, "org", "repo", testPullRequestMainBase, "")
	if err != nil || num != 0 {
		t.Fatalf("num=%d err=%v", num, err)
	}
}

func TestFindOpenReleasePrepPRNumber_noMatchReturnsZero(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &spyListREST{
		listFns: []func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error){
			func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error) {
				return nil, nil
			},
		},
	}

	num, err := prmgmt.FindOpenReleasePrepPRNumber(ctx, st, "org", "repo", testPullRequestMainBase, "9.0.0")
	if err != nil || num != 0 {
		t.Fatalf("num=%d err=%v", num, err)
	}
}

func TestRefreshReleasePR_lookupByPrepRelease_updates(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &spyListREST{
		listFns: []func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error){
			func(_ context.Context, _, _, head, base string) ([]prmgmt.PullRequestRef, error) {
				if head != "acme:prep-release-2.1.0" || base != testPullRequestMainBase {
					t.Fatalf("unexpected list head=%q base=%q", head, base)
				}
				return []prmgmt.PullRequestRef{{Number: 77}}, nil
			},
		},
	}

	res, err := prmgmt.RefreshReleasePR(ctx, prmgmt.RefreshReleaseOptions{
		Owner:         "acme",
		Repo:          "repo",
		CompareRange:  "v1.0.0...v2.1.0",
		TargetVersion: "2.1.0",
		GitHub:        st,
		Now:           fixedClock(time.Date(2020, 8, 1, 12, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Updated || res.Number != 77 {
		t.Fatalf("result %+v", res)
	}
	if len(st.updateBodies) != 1 || st.updateBodies[0].number != 77 {
		t.Fatalf("updates %#v", st.updateBodies)
	}
	if len(st.labelCalls) != 1 || st.labelCalls[0].issue != 77 {
		t.Fatalf("labels %#v", st.labelCalls)
	}
	wantInfos := []string{
		"Refreshing release PR #77 metadata",
		"Release PR #77 metadata refreshed",
	}
	if len(res.Infos) != 2 || res.Infos[0] != wantInfos[0] || res.Infos[1] != wantInfos[1] {
		t.Fatalf("infos %#v want %#v", res.Infos, wantInfos)
	}
}

func TestRefreshReleasePR_noPrepPR_warnsSkipped(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &spyListREST{
		listFns: []func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error){
			func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error) {
				return nil, nil
			},
		},
	}

	res, err := prmgmt.RefreshReleasePR(ctx, prmgmt.RefreshReleaseOptions{
		Owner:         "org",
		Repo:          "repo",
		CompareRange:  "v1.0.0...v2.0.0",
		TargetVersion: "2.0.0",
		GitHub:        st,
		Now:           fixedClock(time.Date(2020, 10, 1, 12, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Updated || len(st.updateBodies) != 0 {
		t.Fatalf("unexpected work done: %+v %#v", res, st.updateBodies)
	}
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], "Could not resolve") {
		t.Fatalf("unexpected warnings %#v", res.Warnings)
	}
	if len(res.Infos) != 0 {
		t.Fatalf("infos should be empty when skipped; got %#v", res.Infos)
	}
}

func TestRefreshReleasePR_lookupPrepReleaseListError_returnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	lookupErr := errors.New("github list exploded")
	st := &spyListREST{
		listFns: []func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error){
			func(context.Context, string, string, string, string) ([]prmgmt.PullRequestRef, error) {
				return nil, lookupErr
			},
		},
	}

	_, err := prmgmt.RefreshReleasePR(ctx, prmgmt.RefreshReleaseOptions{
		Owner:         "org",
		Repo:          "repo",
		CompareRange:  "x..HEAD",
		TargetVersion: "2.2.2",
		GitHub:        st,
		Now:           fixedClock(time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)),
	})
	if err == nil || !strings.Contains(err.Error(), "list prep-release pull requests") {
		t.Fatalf("got err=%v", err)
	}
	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected wrapped lookupErr, got %v", err)
	}
}

func TestRefreshReleasePR_updateBodyError_returnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	editingErr := errors.New("PATCH failed")
	st := &spyListREST{updateErr: editingErr}

	_, err := prmgmt.RefreshReleasePR(ctx, prmgmt.RefreshReleaseOptions{
		Owner:         "o",
		Repo:          "r",
		PRNumber:      33,
		CompareRange:  "a..b",
		TargetVersion: "1.1.1",
		GitHub:        st,
		Now:           fixedClock(time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)),
	})
	if err == nil || !strings.Contains(err.Error(), "update pull request body") {
		t.Fatalf("got err=%v", err)
	}
	if !errors.Is(err, editingErr) {
		t.Fatalf("expected wrapped editingErr, got %v", err)
	}
}
