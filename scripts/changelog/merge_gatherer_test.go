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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v89/github"
)

type rewriteTransport struct {
	base *url.URL
	rt   http.RoundTripper
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = t.base.Scheme
	req.URL.Host = t.base.Host
	return t.rt.RoundTrip(req)
}

type staticSHAGit struct{}

func (staticSHAGit) Run(string, ...string) ([]byte, error) {
	return []byte("aaa\nbbb\n"), nil
}

func Test_gitMergedPRGatherer_dedupesByPRNumberAcrossCommits(t *testing.T) {
	t.Parallel()
	var callCount int

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/commits/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method %s", r.Method)
		}
		callCount++

		prJSON := []map[string]any{
			{
				"number":           10,
				"state":            "closed",
				"merged_at":        "2025-01-01T00:00:00Z",
				"title":            "T",
				"html_url":         "https://example/pull/10",
				"merge_commit_sha": "abcmerge001",
				"user":             map[string]string{"login": "alice"},
				"labels":           []map[string]string{{"name": "bug"}},
				"body":             "## Changelog\nCustomer impact: fix\nSummary: fix it\n",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(prJSON); err != nil {
			t.Fatalf("encode: %v", err)
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	apiURL, err := url.Parse(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	client, err := github.NewClient(github.WithTransport(&rewriteTransport{base: apiURL, rt: http.DefaultTransport}))
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	g := &gitMergedPRGatherer{client: client, execer: staticSHAGit{}}
	got, warns, err := g.GatherMergedPRs(ctx, "o", "r", "v1..HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if len(warns) != 0 {
		t.Fatalf("unexpected warnings: %#v", warns)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 commit PR API calls, got %d", callCount)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 unique PR, got %d", len(got))
	}
	if got[0].Number != 10 || got[0].Labels[0] != "bug" {
		t.Fatalf("unexpected record: %+v", got[0])
	}
	if got[0].Title != "T" {
		t.Fatalf("title: got %q want T", got[0].Title)
	}
	if got[0].MergeCommitSHA != "abcmerge001" {
		t.Fatalf("merge_commit_sha: got %q want abcmerge001", got[0].MergeCommitSHA)
	}
	if got[0].AuthorLogin != "alice" {
		t.Fatalf("AuthorLogin: got %q want alice", got[0].AuthorLogin)
	}
}
