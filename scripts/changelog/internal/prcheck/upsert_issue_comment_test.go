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
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/prcheck"
)

const recListCall = "list"

type recorderREST struct {
	listReply []prcheck.Comment
	calls     []string
	listErr   error
}

func (r *recorderREST) ListIssueComments(context.Context, string, string, int) ([]prcheck.Comment, error) {
	r.calls = append(r.calls, recListCall)
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.listReply, nil
}

func (r *recorderREST) CreateIssueComment(_ context.Context, _, _ string, _ int, body string) error {
	r.calls = append(r.calls, "create:"+body)
	return nil
}

func (r *recorderREST) UpdateIssueComment(_ context.Context, _, _ string, id int64, body string) error {
	r.calls = append(r.calls, "update:"+strconv.FormatInt(id, 10)+":"+body)
	return nil
}

func TestUpsertVerdictIssueComment_noChangelog(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	wantBody := prcheck.BuildNoChangelogPassCommentBody(prcheck.MarkerForPRCheck)
	botExisting := []prcheck.Comment{
		{ID: 7, Body: prcheck.MarkerForPRCheck + "\nOLD", UserLogin: "github-actions[bot]"},
	}

	t.Run("existing updates marker comment", func(t *testing.T) {
		t.Parallel()
		rec := &recorderREST{listReply: botExisting}
		if err := prcheck.UpsertVerdictIssueComment(ctx, rec, "o", "r", 1, prcheck.Verdict{
			Status: prcheck.StatusPass, NoChangelogSkip: true,
		}); err != nil {
			t.Fatal(err)
		}
		if len(rec.calls) != 2 || rec.calls[0] != recListCall || !strings.HasPrefix(rec.calls[1], "update:7:") {
			t.Fatalf("calls=%q", rec.calls)
		}
		if body := strings.TrimPrefix(rec.calls[1], "update:7:"); body != wantBody {
			t.Fatalf("update body mismatch")
		}
	})

	t.Run("no existing no mutate", func(t *testing.T) {
		t.Parallel()
		rec := &recorderREST{listReply: nil}
		if err := prcheck.UpsertVerdictIssueComment(ctx, rec, "o", "r", 1, prcheck.Verdict{
			Status: prcheck.StatusPass, NoChangelogSkip: true,
		}); err != nil {
			t.Fatal(err)
		}
		if strings.Join(rec.calls, ",") != recListCall {
			t.Fatalf("calls=%q", rec.calls)
		}
	})
}

func TestUpsertVerdictIssueComment_pass(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	wantPass := prcheck.BuildPassCommentBody(prcheck.MarkerForPRCheck)
	bot := []prcheck.Comment{
		{ID: 9, Body: prcheck.MarkerForPRCheck, UserLogin: "github-actions[bot]"},
	}

	t.Run("existing updates pass body", func(t *testing.T) {
		t.Parallel()
		rec := &recorderREST{listReply: bot}
		err := prcheck.UpsertVerdictIssueComment(ctx, rec, "o", "r", 2, prcheck.Verdict{Status: prcheck.StatusPass})
		if err != nil {
			t.Fatal(err)
		}
		upd := ""
		for _, c := range rec.calls {
			if after, ok := strings.CutPrefix(c, "update:9:"); ok {
				upd = after
			}
		}
		if upd != wantPass {
			t.Fatalf("got upd body %q", upd)
		}
	})

	t.Run("silent when no marker comment yet", func(t *testing.T) {
		t.Parallel()
		rec := &recorderREST{listReply: nil}
		if err := prcheck.UpsertVerdictIssueComment(ctx, rec, "o", "r", 2, prcheck.Verdict{Status: prcheck.StatusPass}); err != nil {
			t.Fatal(err)
		}
		if strings.Join(rec.calls, ",") != recListCall {
			t.Fatalf("calls=%q", rec.calls)
		}
	})
}

func TestUpsertVerdictIssueComment_fail(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	failBody := prcheck.BuildFailureCommentBody(prcheck.MarkerForPRCheck, []string{"err-a"})
	bot := []prcheck.Comment{{ID: 3, Body: prcheck.MarkerForPRCheck, UserLogin: "github-actions[bot]"}}

	t.Run("existing updates failure comment", func(t *testing.T) {
		t.Parallel()
		rec := &recorderREST{listReply: bot}
		if err := prcheck.UpsertVerdictIssueComment(ctx, rec, "o", "r", 4, prcheck.Verdict{Status: prcheck.StatusFail, Errors: []string{"err-a"}}); err != nil {
			t.Fatal(err)
		}
		var gotBody string
		for _, c := range rec.calls {
			if after, ok := strings.CutPrefix(c, "update:3:"); ok {
				gotBody = after
			}
		}
		if gotBody != failBody {
			t.Fatalf("want failure comment body parity")
		}
	})

	t.Run("missing creates failure comment", func(t *testing.T) {
		t.Parallel()
		rec := &recorderREST{listReply: nil}
		if err := prcheck.UpsertVerdictIssueComment(ctx, rec, "o", "r", 4, prcheck.Verdict{Status: prcheck.StatusFail, Errors: []string{"err-a"}}); err != nil {
			t.Fatal(err)
		}
		var cre string
		for _, c := range rec.calls {
			if after, ok := strings.CutPrefix(c, "create:"); ok {
				cre = after
			}
		}
		if cre != failBody {
			t.Fatalf("create body mismatch")
		}
	})
}

func TestUpsertVerdictIssueComment_propagates_list_error(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	wantErr := errors.New("boom list")
	rec := &recorderREST{listErr: wantErr}
	err := prcheck.UpsertVerdictIssueComment(ctx, rec, "o", "r", 9, prcheck.Verdict{Status: prcheck.StatusPass})
	if !errors.Is(err, wantErr) {
		t.Fatalf("want wrapped list error got %v", err)
	}
}
