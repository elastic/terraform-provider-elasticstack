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

package prcheck_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/prcheck"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/section"
)

const validateOptsRepoOwnerGuardMsg = "owner and repo are required"

type stubFetcher struct {
	pr  prcheck.PullRequest
	err error
}

func (s stubFetcher) GetPullRequest(_ context.Context, _, _ string, _ int) (*prcheck.PullRequest, error) {
	if s.err != nil {
		return nil, s.err
	}
	p := s.pr
	return &p, nil
}

func TestValidate_options_and_fetch_errors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseOpts := func() prcheck.ValidateOptions {
		return prcheck.ValidateOptions{
			Owner:            "elastic",
			Repo:             "terraform-provider-elasticstack",
			Number:           1,
			Fetcher:          stubFetcher{pr: prcheck.PullRequest{Body: "## Changelog\nCustomer impact: none\n"}},
			NoChangelogLabel: "",
		}
	}

	t.Run("nil fetcher", func(t *testing.T) {
		t.Parallel()
		o := baseOpts()
		o.Fetcher = nil
		_, err := prcheck.Validate(ctx, o)
		if err == nil || !strings.Contains(err.Error(), "fetcher required") {
			t.Fatalf("expected fetcher required error, got %#v", err)
		}
	})

	t.Run("empty owner", func(t *testing.T) {
		t.Parallel()
		o := baseOpts()
		o.Owner = ""
		_, err := prcheck.Validate(ctx, o)
		if err == nil || err.Error() != validateOptsRepoOwnerGuardMsg {
			t.Fatalf("expected %q, got %#v", validateOptsRepoOwnerGuardMsg, err)
		}
	})

	t.Run("empty repo", func(t *testing.T) {
		t.Parallel()
		o := baseOpts()
		o.Repo = ""
		_, err := prcheck.Validate(ctx, o)
		if err == nil || err.Error() != validateOptsRepoOwnerGuardMsg {
			t.Fatalf("expected %q, got %#v", validateOptsRepoOwnerGuardMsg, err)
		}
	})

	t.Run("non-positive number", func(t *testing.T) {
		t.Parallel()
		o := baseOpts()
		o.Number = 0
		_, err := prcheck.Validate(ctx, o)
		if err == nil || !strings.Contains(err.Error(), "number") {
			t.Fatalf("expected number error, got %#v", err)
		}
	})

	t.Run("fetcher error", func(t *testing.T) {
		t.Parallel()
		o := baseOpts()
		sentinel := errors.New("network down")
		o.Fetcher = stubFetcher{err: sentinel}
		_, err := prcheck.Validate(ctx, o)
		if err == nil || !strings.Contains(err.Error(), "fetch pull request #1") || !strings.Contains(err.Error(), "network down") {
			t.Fatalf("expected wrapped fetch error, got %#v", err)
		}
		if !errors.Is(err, sentinel) {
			t.Fatalf("errors.Is sentinel: expected true, got err=%v", err)
		}
	})
}

func TestValidate_customer_impact_variants_ok(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	owner, repo := "elastic", "example"
	num := 77

	tests := []struct {
		name string
		body string
	}{
		{name: "none without summary",
			body: "## Changelog\nCustomer impact: none\n"},
		{name: "fix with summary",
			body: "## Changelog\nCustomer impact: fix\nSummary: fixed a bug\n"},
		{name: "enhancement with summary",
			body: "## Changelog\nCustomer impact: enhancement\nSummary: nicer UX\n"},
		{name: "breaking with breaking subsection",
			body: "## Changelog\nCustomer impact: breaking\nSummary: remove field X\n\n### Breaking changes\nUsers must...\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := prcheck.Validate(ctx, prcheck.ValidateOptions{
				Owner:   owner,
				Repo:    repo,
				Number:  num,
				Fetcher: stubFetcher{pr: prcheck.PullRequest{Number: num, Body: tc.body}},
			})
			if err != nil {
				t.Fatalf("Validate: unexpected error %v", err)
			}
			if got.Status != prcheck.StatusPass {
				t.Fatalf("expected pass, got %+v", got)
			}
			if len(got.Errors) != 0 {
				t.Fatalf("expected no errors, got %v", got.Errors)
			}
		})
	}
}

func TestValidate_failures_and_bypass(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	owner, repo := "elastic", "example"
	baseNum := 100

	tests := []struct {
		name         string
		body         string
		wantStatus   prcheck.Status
		wantSkip     bool
		wantErrSubs  []string
		customLabels []string
		skipNoChLbl  bool
	}{
		{
			name:       "missing changelog heading",
			body:       "## Description\nnothing here",
			wantStatus: prcheck.StatusFail,
			wantErrSubs: []string{
				"No ## Changelog section found in PR body",
			},
		},
		{
			name:       "empty body",
			body:       "",
			wantStatus: prcheck.StatusFail,
			wantErrSubs: []string{
				"No ## Changelog section found in PR body",
			},
		},
		{
			name:       "missing customer impact field",
			body:       "## Changelog\nSummary: something\n",
			wantStatus: prcheck.StatusFail,
			wantErrSubs: []string{
				"Missing required field: Customer impact",
			},
		},
		{
			name:       "invalid impact",
			body:       "## Changelog\nCustomer impact: maybe\nSummary: hmm\n",
			wantStatus: prcheck.StatusFail,
			wantErrSubs: []string{
				`Invalid Customer impact value: "maybe"`,
				`Must be one of: none, fix, enhancement, breaking`,
			},
		},
		{
			name:       "missing summary when fix",
			body:       "## Changelog\nCustomer impact: fix\n",
			wantStatus: prcheck.StatusFail,
			wantErrSubs: []string{
				`Missing required field: Summary (required when Customer impact is not "none")`,
			},
		},
		{
			name:       "empty summary value",
			body:       "## Changelog\nCustomer impact: fix\nSummary:\n",
			wantStatus: prcheck.StatusFail,
			wantErrSubs: []string{
				`Missing required field: Summary (required when Customer impact is not "none")`,
			},
		},
		{
			name:       "breaking without breaking subsection heading",
			body:       "## Changelog\nCustomer impact: breaking\nSummary: major change\n",
			wantStatus: prcheck.StatusFail,
			wantErrSubs: []string{
				"Customer impact: breaking requires a ### Breaking changes subsection",
			},
		},
		{
			name:       "empty breaking subsection content",
			body:       "## Changelog\nCustomer impact: breaking\nSummary: majors\n\n### Breaking changes\n",
			wantStatus: prcheck.StatusFail,
			wantErrSubs: []string{
				"### Breaking changes section is present but contains no content",
			},
		},
		{
			name:        "breaking heading forbidden when fix",
			body:        "## Changelog\nCustomer impact: fix\nSummary: small fix\n\n### Breaking changes\noops\n",
			wantStatus:  prcheck.StatusFail,
			wantErrSubs: []string{section.RuleCBreakingOnlyWhenBreakingImpactMsg},
		},
		{
			name:        "breaking heading forbidden when enhancement",
			body:        "## Changelog\nCustomer impact: enhancement\nSummary: improved UI\n\n### Breaking changes\nshould not appear\n",
			wantStatus:  prcheck.StatusFail,
			wantErrSubs: []string{section.RuleCBreakingOnlyWhenBreakingImpactMsg},
		},
		{
			name:        "breaking heading forbidden when none",
			body:        "## Changelog\nCustomer impact: none\n\n### Breaking changes\nstray subsection\n",
			wantStatus:  prcheck.StatusFail,
			wantErrSubs: []string{section.RuleCBreakingOnlyWhenBreakingImpactMsg},
		},
		{
			name:         "no-changelog label bypass ignores bad body",
			body:         "",
			customLabels: []string{"no-changelog"},
			wantStatus:   prcheck.StatusPass,
			wantSkip:     true,
		},
		{
			name:         "custom no-changelog label",
			body:         "",
			customLabels: []string{"changelog-not-needed"},
			wantStatus:   prcheck.StatusPass,
			wantSkip:     true,
			skipNoChLbl:  true,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			num := baseNum + i
			opts := prcheck.ValidateOptions{
				Owner:   owner,
				Repo:    repo,
				Number:  num,
				Fetcher: stubFetcher{pr: prcheck.PullRequest{Number: num, Body: tc.body, Labels: tc.customLabels}},
			}
			if tc.skipNoChLbl {
				opts.NoChangelogLabel = "changelog-not-needed"
			}

			got, err := prcheck.Validate(ctx, opts)
			if err != nil {
				t.Fatalf("Validate: unexpected error %v", err)
			}
			if got.Status != tc.wantStatus {
				t.Fatalf("status got %s want %s (verdict %+v)", got.Status, tc.wantStatus, got)
			}
			if got.NoChangelogSkip != tc.wantSkip {
				t.Fatalf("no_changelog_skip got %v want %v", got.NoChangelogSkip, tc.wantSkip)
			}
			joined := strings.Join(got.Errors, "\n")
			for _, sub := range tc.wantErrSubs {
				if !strings.Contains(joined, sub) {
					t.Fatalf("expected error containing %q, got %#v", sub, got.Errors)
				}
			}
		})
	}
}

func TestStatus_UnmarshalJSON_edges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload string
		wantErr string
	}{
		{name: "unknown literal", payload: `"unknown"`, wantErr: "prcheck: invalid status JSON value"},
		{name: "number", payload: `123`, wantErr: "prcheck: status must be a JSON string"},
		{name: "array", payload: `[]`, wantErr: "prcheck: status must be a JSON string"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var st prcheck.Status
			err := json.Unmarshal([]byte(tc.payload), &st)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("got %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestVerdictMarshalJSON_roundTrip(t *testing.T) {
	t.Parallel()

	v := prcheck.Verdict{
		Status:          prcheck.StatusFail,
		Errors:          []string{"a", "b"},
		NoChangelogSkip: false,
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	wantSub := `"status":"fail"`
	if !strings.Contains(string(b), wantSub) {
		t.Fatalf("expected %s in %s", wantSub, b)
	}
	var back prcheck.Verdict
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Status != v.Status || len(back.Errors) != 2 || back.NoChangelogSkip {
		t.Fatalf("round-trip mismatch %+v vs %+v", back, v)
	}
}
