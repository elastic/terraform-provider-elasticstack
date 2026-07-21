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
	"github.com/stretchr/testify/require"
)

func TestCreateAgentlessPolicy(t *testing.T) {
	t.Run("success_returns_item", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"item":{"id":"pp-1","created_at":"2026-01-01T00:00:00.000Z","created_by":"elastic"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.CreateAgentlessPolicy(context.Background(), client, "", kbapi.PostFleetAgentlessPoliciesJSONRequestBody{
			Name:    "test-agentless",
			Package: kbapi.KibanaHTTPAPIsPackagePolicyPackage{Name: "cloud_security_posture", Version: "1.14.0"},
		})

		require.False(t, diags.HasError())
		require.NotNil(t, item)
		require.Equal(t, "pp-1", item.Id)
	})
}

func TestReadAgentlessPolicyViaPackagePolicy(t *testing.T) {
	t.Run("404_returns_nil_no_error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"package policy not found"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.ReadAgentlessPolicyViaPackagePolicy(context.Background(), client, "", "missing")

		require.False(t, diags.HasError())
		require.Nil(t, item)
	})
}

func TestUpdateAgentlessPolicyViaPackagePolicy(t *testing.T) {
	t.Run("success_returns_item", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"item":{"id":"pp-1","created_at":"2026-01-01T00:00:00.000Z","created_by":"elastic"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.UpdateAgentlessPolicyViaPackagePolicy(context.Background(), client, "", "pp-1", kbapi.PackagePolicyRequest{})

		require.False(t, diags.HasError())
		require.NotNil(t, item)
		require.Equal(t, "pp-1", item.Id)
	})
}

func TestDeleteAgentlessPolicy(t *testing.T) {
	t.Run("404_is_noop", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"policy not found"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		isConflict, diags := fleet.DeleteAgentlessPolicy(context.Background(), client, "", "missing", false)

		require.False(t, diags.HasError())
		require.False(t, isConflict)
	})

	t.Run("non_2xx_returns_error", func(t *testing.T) {
		var calls atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"statusCode":400,"error":"Bad Request","message":"invalid request"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		isConflict, diags := fleet.DeleteAgentlessPolicy(context.Background(), client, "", "pp-1", false)

		require.True(t, diags.HasError())
		require.False(t, isConflict, "a non-409 error must not be reported as a conflict")
		require.Equal(t, int64(1), calls.Load(), "non-409 errors must not be retried")
	})

	t.Run("retries_on_409_then_succeeds", func(t *testing.T) {
		var calls atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			n := calls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			if n < 2 {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, `{"statusCode":409,"error":"Conflict","message":"write lock"}`)
				return
			}
			fmt.Fprint(w, `{"id":"pp-1"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		isConflict, diags := fleet.DeleteAgentlessPolicy(context.Background(), client, "", "pp-1", false)

		require.False(t, diags.HasError())
		require.False(t, isConflict, "a transient 409 resolved by retry must not be reported as a conflict")
		require.Equal(t, int64(2), calls.Load())
	})

	t.Run("force_delete_sets_force_query_param", func(t *testing.T) {
		var capturedQuery string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"pp-1"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		isConflict, diags := fleet.DeleteAgentlessPolicy(context.Background(), client, "", "pp-1", true)

		require.False(t, diags.HasError())
		require.False(t, isConflict)
		require.Contains(t, capturedQuery, "force=true")
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
		isConflict, diags := fleet.DeleteAgentlessPolicy(context.Background(), client, "", "pp-1", false)

		require.True(t, diags.HasError())
		require.True(t, isConflict, "exhausted 409 retries must be reported as a conflict")
		require.Equal(t, int64(kibanautil.ConflictMaxAttempts), calls.Load())
	})

	t.Run("transport_error_after_409_resets_is_conflict", func(t *testing.T) {
		var calls atomic.Int64
		client := newTestClientWithRoundTripper(t, roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			n := calls.Add(1)
			if n == 1 {
				rec := httptest.NewRecorder()
				rec.Header().Set("Content-Type", "application/json")
				rec.WriteHeader(http.StatusConflict)
				fmt.Fprint(rec, `{"statusCode":409,"error":"Conflict","message":"write lock"}`)
				return rec.Result(), nil
			}
			return nil, fmt.Errorf("connection reset by peer")
		}))

		isConflict, diags := fleet.DeleteAgentlessPolicy(context.Background(), client, "", "pp-1", false)

		require.True(t, diags.HasError())
		require.False(t, isConflict, "a final transport error must reset isConflict after an earlier 409")
		require.Equal(t, int64(2), calls.Load())
	})
}
