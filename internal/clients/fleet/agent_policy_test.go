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

package fleet_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
)

func TestCreateAgentPolicy(t *testing.T) {
	t.Run("retries_on_409_then_succeeds", func(t *testing.T) {
		var calls atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			n := calls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			if n < 3 {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, `{"statusCode":409,"error":"Conflict","message":"write lock"}`)
				return
			}
			fmt.Fprint(w, `{"item":{"id":"test-id","name":"test-policy","namespace":"default","status":"active"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		policy, diags := fleet.CreateAgentPolicy(context.Background(), client, kbapi.PostFleetAgentPoliciesJSONRequestBody{
			Name:      "test-policy",
			Namespace: "default",
		}, false, "")

		if diags.HasError() {
			t.Fatalf("unexpected error: %s", diags[0].Summary())
		}
		if policy == nil || policy.Id != "test-id" {
			t.Fatalf("got policy %+v, want id=test-id", policy)
		}
		if got := calls.Load(); got != 3 {
			t.Fatalf("server received %d requests, want 3", got)
		}
	})

	t.Run("max_retries_exhausted_returns_error", func(t *testing.T) {
		var calls atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `{"statusCode":409,"error":"Conflict","message":"write lock"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		_, diags := fleet.CreateAgentPolicy(context.Background(), client, kbapi.PostFleetAgentPoliciesJSONRequestBody{
			Name:      "test-policy",
			Namespace: "default",
		}, false, "")

		if !diags.HasError() {
			t.Fatal("expected error diagnostics")
		}
		if got := calls.Load(); got != int64(kibanautil.ConflictMaxAttempts) {
			t.Fatalf("server received %d requests, want %d (maxAttempts)", got, kibanautil.ConflictMaxAttempts)
		}
	})

	t.Run("non_409_error_is_not_retried", func(t *testing.T) {
		var calls atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"statusCode":400,"error":"Bad Request","message":"invalid body"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		_, diags := fleet.CreateAgentPolicy(context.Background(), client, kbapi.PostFleetAgentPoliciesJSONRequestBody{
			Name:      "test-policy",
			Namespace: "default",
		}, false, "")

		if !diags.HasError() {
			t.Fatal("expected error diagnostics")
		}
		if got := calls.Load(); got != 1 {
			t.Fatalf("server received %d requests, want 1 (no retry on non-409)", got)
		}
	})
}

func TestGetAgentPolicyByName_SendsKueryFilterOnTheWire(t *testing.T) {
	const policyName = "k8s-metrics-pernode"

	var gotKuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKuery = r.URL.Query().Get("kuery")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[{"id":"policy-1","name":"k8s-metrics-pernode"}],"page":1,"perPage":1,"total":1}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, diags := fleet.GetAgentPolicyByName(t.Context(), client, policyName, "")
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary())
	}
	const wantKuery = `ingest-agent-policies.name:"k8s-metrics-pernode"`
	if gotKuery != wantKuery {
		t.Fatalf("kuery = %q, want %q", gotKuery, wantKuery)
	}
}

func TestGetAgentPolicyByName_ReturnsSingleExactMatchOnFirstPage(t *testing.T) {
	const policyName = "k8s-metrics-pernode"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[{"id":"policy-1","name":"k8s-metrics-pernode"}],"page":1,"perPage":1,"total":1}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicyByName(t.Context(), client, policyName, "")
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary())
	}
	if len(got) != 1 {
		t.Fatalf("got %d policies, want 1", len(got))
	}
	if got[0].Id != "policy-1" || got[0].Name != policyName {
		t.Fatalf("got id=%q name=%q, want id=%q name=%q", got[0].Id, got[0].Name, "policy-1", policyName)
	}
}

func TestGetAgentPolicyByName_EscapesQuotesAndBackslashesInKuery(t *testing.T) {
	const policyName = "has \"quotes\" and \\backslash"
	const wantKuery = `ingest-agent-policies.name:"has \"quotes\" and \\backslash"`

	var gotKuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKuery = r.URL.Query().Get("kuery")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[],"page":1,"perPage":1,"total":0}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, diags := fleet.GetAgentPolicyByName(t.Context(), client, policyName, "")
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary())
	}
	if gotKuery != wantKuery {
		t.Fatalf("kuery = %q, want %q", gotKuery, wantKuery)
	}
}

func TestGetAgentPolicyByName_DropsSimilarButNonExactNames(t *testing.T) {
	const policyName = "k8s-metrics-pernode"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[`+
			`{"id":"similar-suffix","name":"k8s-metrics-pernode-extra"},`+
			`{"id":"exact","name":"k8s-metrics-pernode"},`+
			`{"id":"similar-prefix","name":"prefix-k8s-metrics-pernode"}`+
			`],"page":1,"perPage":3,"total":3}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicyByName(t.Context(), client, policyName, "")
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary())
	}
	if len(got) != 1 {
		t.Fatalf("got %d policies, want 1 exact match", len(got))
	}
	if got[0].Id != "exact" || got[0].Name != policyName {
		t.Fatalf("got id=%q name=%q, want id=%q name=%q", got[0].Id, got[0].Name, "exact", policyName)
	}
}

func TestGetAgentPolicyByName_ScopesRequestToSpace(t *testing.T) {
	const spaceID = "my-space"

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[],"page":1,"perPage":100,"total":0}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, diags := fleet.GetAgentPolicyByName(t.Context(), client, "k8s-metrics-pernode", spaceID)
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary())
	}
	const wantPath = "/s/my-space/api/fleet/agent_policies"
	if gotPath != wantPath {
		t.Fatalf("path = %q, want %q", gotPath, wantPath)
	}
}

func TestGetAgentPolicyByName_ReturnsEmptySliceWhenNoPoliciesMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[],"page":1,"perPage":20,"total":0}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicyByName(t.Context(), client, "does-not-exist", "")
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary())
	}
	if got == nil {
		t.Fatal("got nil, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("got %d policies, want 0", len(got))
	}
}

func TestGetAgentPolicyByName_PaginatesFilteredResultsBeforeMatching(t *testing.T) {
	const policyName = "k8s-metrics-pernode"
	const wantKuery = `ingest-agent-policies.name:"k8s-metrics-pernode"`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		perPage, perPageErr := strconv.Atoi(q.Get("perPage"))
		page, pageErr := strconv.Atoi(q.Get("page"))
		paged := perPageErr == nil && pageErr == nil && perPage > 0 && page > 0 && q.Get("sortField") == "name" && q.Get("kuery") == wantKuery

		w.Header().Set("Content-Type", "application/json")
		if !paged {
			fmt.Fprint(w, `{"items":[{"id":"exact-page-1","name":"k8s-metrics-pernode"}],"page":1,"perPage":1,"total":2}`)
			return
		}

		total := perPage + 1
		if page == 1 {
			items := make([]string, 0, perPage)
			items = append(items, `{"id":"exact-page-1","name":"k8s-metrics-pernode"}`)
			for i := 1; i < perPage; i++ {
				items = append(items, fmt.Sprintf(`{"id":"other-%d","name":"k8s-metrics-pernode-%d"}`, i, i))
			}
			fmt.Fprintf(w, `{"items":[%s],"page":%d,"perPage":%d,"total":%d}`, strings.Join(items, ","), page, perPage, total)
			return
		}
		fmt.Fprintf(w, `{"items":[{"id":"exact-page-2","name":"k8s-metrics-pernode"}],"page":%d,"perPage":%d,"total":%d}`, page, perPage, total)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicyByName(t.Context(), client, policyName, "")
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary())
	}
	if len(got) != 2 {
		t.Fatalf("got %d policies, want both exact matches", len(got))
	}
	if got[0].Id != "exact-page-1" || got[1].Id != "exact-page-2" {
		t.Fatalf("got ids %q and %q, want exact-page-1 and exact-page-2", got[0].Id, got[1].Id)
	}
}

func TestGetAgentPolicyByName_Non2xxReturnsDiagnostic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"statusCode":500,"error":"Internal Server Error","message":"boom"}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicyByName(t.Context(), client, "k8s-metrics-pernode", "")
	if !diags.HasError() {
		t.Fatal("expected a diagnostic for HTTP 500")
	}
	if got != nil {
		t.Fatalf("got %d policies, want none when the list request fails", len(got))
	}
}

func TestGetAgentPolicyByName_NonJSON200ReturnsDiagnostic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `not-json`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicyByName(t.Context(), client, "k8s-metrics-pernode", "")
	if !diags.HasError() {
		t.Fatal("expected a diagnostic when the list body is not JSON")
	}
	if got != nil {
		t.Fatalf("got %d policies, want none when the list body is not JSON", len(got))
	}
}

func TestGetAgentPolicyByName_NetworkErrorReturnsDiagnostic(t *testing.T) {
	client := newTestClientWithRoundTripper(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	}))

	got, diags := fleet.GetAgentPolicyByName(t.Context(), client, "k8s-metrics-pernode", "")
	if !diags.HasError() {
		t.Fatal("expected a diagnostic for a network error")
	}
	if got != nil {
		t.Fatalf("got %d policies, want none when the list request fails", len(got))
	}
}

func TestGetAgentPolicyByName_ErrorsWhenPageIsShortBeforeTotal(t *testing.T) {
	const policyName = "k8s-metrics-pernode"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || page < 1 {
			page = 1
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"items":[{"id":"policy-%d","name":"k8s-metrics-pernode"}],"page":%d,"perPage":1,"total":3}`, page, page)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicyByName(t.Context(), client, policyName, "")
	if !diags.HasError() {
		t.Fatalf("expected error for a short page, got %d policies", len(got))
	}
	if got != nil {
		t.Fatalf("got %d policies, want none when pagination fails", len(got))
	}
}

func TestGetAgentPolicy_NotFound(t *testing.T) {
	const policyID = "does-not-exist"

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"Agent policy not found"}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicy(t.Context(), client, policyID, "")
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags[0].Summary())
	}
	if got != nil {
		t.Fatalf("got policy %+v, want not found", got)
	}
	if gotPath != "/api/fleet/agent_policies/"+policyID {
		t.Fatalf("path = %q, want /api/fleet/agent_policies/%s", gotPath, policyID)
	}
}

func TestGetAgentPolicy_Non2xxReturnsDiagnostic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"statusCode":500,"error":"Internal Server Error","message":"boom"}`)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicy(t.Context(), client, "policy-1", "")
	if !diags.HasError() {
		t.Fatal("expected a diagnostic for HTTP 500")
	}
	if got != nil {
		t.Fatalf("got policy %+v, want none when the get request fails", got)
	}
}

func TestGetAgentPolicy_NetworkErrorReturnsDiagnostic(t *testing.T) {
	client := newTestClientWithRoundTripper(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	}))

	got, diags := fleet.GetAgentPolicy(t.Context(), client, "policy-1", "")
	if !diags.HasError() {
		t.Fatal("expected a diagnostic for a network error")
	}
	if got != nil {
		t.Fatalf("got policy %+v, want none when the get request fails", got)
	}
}

func TestGetAgentPolicyByName_ErrorsWhenTotalChangesBetweenPages(t *testing.T) {
	const policyName = "k8s-metrics-pernode"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		perPage, err := strconv.Atoi(q.Get("perPage"))
		if err != nil || perPage < 1 {
			perPage = 1
		}
		page, err := strconv.Atoi(q.Get("page"))
		if err != nil || page < 1 {
			page = 1
		}
		total := perPage * 2
		if page >= 2 {
			total = perPage
		}
		items := make([]string, 0, perPage)
		for i := range perPage {
			itemName := "other"
			if page == 1 && i == 0 {
				itemName = policyName
			}
			items = append(items, fmt.Sprintf(`{"id":"policy-%d","name":"%s"}`, (page-1)*perPage+i, itemName))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"items":[%s],"page":%d,"perPage":%d,"total":%d}`, strings.Join(items, ","), page, perPage, total)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicyByName(t.Context(), client, policyName, "")
	if !diags.HasError() {
		t.Fatalf("expected error when total changes, got %d policies", len(got))
	}
	if got != nil {
		t.Fatalf("got %d policies, want none when pagination fails", len(got))
	}
}

func TestGetAgentPolicyByName_ErrorsWhenPolicyIDRepeats(t *testing.T) {
	const policyName = "k8s-metrics-pernode"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		perPage, err := strconv.Atoi(q.Get("perPage"))
		if err != nil || perPage < 1 {
			perPage = 1
		}
		page, err := strconv.Atoi(q.Get("page"))
		if err != nil || page < 1 {
			page = 1
		}
		total := perPage + 1
		w.Header().Set("Content-Type", "application/json")
		if page > 1 {
			fmt.Fprintf(w, `{"items":[{"id":"policy-0","name":"%s"}],"page":%d,"perPage":%d,"total":%d}`, policyName, page, perPage, total)
			return
		}
		items := make([]string, 0, perPage)
		for i := range perPage {
			itemName := "other"
			if i == 0 {
				itemName = policyName
			}
			items = append(items, fmt.Sprintf(`{"id":"policy-%d","name":"%s"}`, i, itemName))
		}
		fmt.Fprintf(w, `{"items":[%s],"page":%d,"perPage":%d,"total":%d}`, strings.Join(items, ","), page, perPage, total)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, diags := fleet.GetAgentPolicyByName(t.Context(), client, policyName, "")
	if !diags.HasError() {
		t.Fatalf("expected error when a policy id repeats, got %d policies", len(got))
	}
	if got != nil {
		t.Fatalf("got %d policies, want none when pagination fails", len(got))
	}
}
