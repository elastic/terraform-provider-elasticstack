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
	"fmt"
	"net/http"
	"net/http/httptest"
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
