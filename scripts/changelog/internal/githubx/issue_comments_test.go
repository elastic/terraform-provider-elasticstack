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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
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

func testGitHubRESTClient(tb testing.TB, srv *httptest.Server) *github.Client {
	tb.Helper()
	u, err := url.Parse(srv.URL + "/")
	if err != nil {
		tb.Fatal(err)
	}
	c, err := github.NewClient(github.WithTransport(&rewriteTransport{base: u, rt: http.DefaultTransport}))
	if err != nil {
		tb.Fatal(err)
	}
	return c
}

func TestListIssueComments_pages(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	path := "/repos/o/r/issues/42/comments"
	var srvURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if per := r.URL.Query().Get("per_page"); per != "" && per != "100" {
			http.Error(w, "bad per_page", http.StatusBadRequest)
			return
		}
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		if r.URL.Path != path || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch page {
		case "1":
			link := fmt.Sprintf("<%s%s?page=2&per_page=100>; rel=\"next\"", srvURL, path)
			w.Header().Set("Link", link)
			w.Header().Set("Content-Type", "application/json")
			_, err := fmt.Fprintf(w, `[{"id":1,"body":"one","user":{"login":"a"}},{"id":2,"body":"two"}]`)
			if err != nil {
				panic(err)
			}
			return
		case "2":
			w.Header().Set("Content-Type", "application/json")
			if _, err := fmt.Fprintf(w, `[{"id":3,"body":"three"}]`); err != nil {
				panic(err)
			}
			return
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	srvURL = ts.URL
	client := testGitHubRESTClient(t, ts)

	comments, err := ListIssueComments(ctx, client, "o", "r", 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 3 || comments[0].ID != 1 || comments[1].Body != "two" || comments[2].UserLogin != "" {
		t.Fatalf("unexpected comments: %+v", comments)
	}
}

func TestListIssueComments_server_error(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/issues/7/comments" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := testGitHubRESTClient(t, ts)
	if _, err := ListIssueComments(ctx, client, "o", "r", 7); err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateIssueComment_request(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var gotMethod, rawBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/ab/cd/issues/88/comments" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		gotMethod = r.Method
		b, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		rawBody = string(b)

		w.Header().Set("Content-Type", "application/json")
		payload := json.RawMessage(`{"id":501}`)
		if _, err := w.Write(payload); err != nil {
			panic(err)
		}
	}))
	defer ts.Close()

	client := testGitHubRESTClient(t, ts)
	body := "## hi"
	if err := CreateIssueComment(ctx, client, "ab", "cd", 88, body); err != nil {
		t.Fatal(err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method=%q want POST", gotMethod)
	}

	var decoded struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal([]byte(rawBody), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Body != body {
		t.Fatalf("body=%q want %q", decoded.Body, body)
	}
}

func TestUpdateIssueComment_request(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	const wantID int64 = 404
	var gotMethod, rawBody string

	path := fmt.Sprintf("/repos/ab/cd/issues/comments/%d", wantID)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		gotMethod = r.Method
		b, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		rawBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprintf(w, `{"id":%s}`, strconv.FormatInt(wantID, 10)); err != nil {
			panic(err)
		}
	}))
	defer ts.Close()

	client := testGitHubRESTClient(t, ts)
	newBody := "fixed"
	if err := UpdateIssueComment(ctx, client, "ab", "cd", wantID, newBody); err != nil {
		t.Fatal(err)
	}

	if gotMethod != http.MethodPatch {
		t.Fatalf("method=%q want PATCH", gotMethod)
	}
	var decoded struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal([]byte(rawBody), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Body != newBody {
		t.Fatalf("body=%q want %q", decoded.Body, newBody)
	}
}
