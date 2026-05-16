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

package elasticsearch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/stretchr/testify/require"
)

// newMockElasticsearchServerForILM returns an httptest.Server that responds to
// the index settings APIs used by GetIndicesWithILMPolicy and ClearILMPolicyFromIndices.
func newMockElasticsearchServerForILM(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"cluster_uuid":"test-cluster","version":{"number":"8.19.0","build_flavor":"default"}}`)
			return
		}
		handler(w, r)
	}))
}

func newMockScopedClient(t *testing.T, srv *httptest.Server) *clients.ElasticsearchScopedClient {
	t.Helper()
	esClient, err := elasticsearch8.NewTypedClient(elasticsearch8.Config{
		Addresses: []string{srv.URL},
		Username:  "elastic",
		Password:  "changeme",
	})
	require.NoError(t, err)
	return clients.NewElasticsearchScopedClientForTest(esClient, []string{srv.URL})
}

func TestGetIndicesWithILMPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		policyName   string
		handler      http.HandlerFunc
		wantIndices  []string
		wantHasError bool
	}{
		{
			name:       "no indices in use",
			policyName: "my-policy",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/_ilm/policy/my-policy" {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintf(w, `{"error":"unexpected path: %s"}`, r.URL.Path)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"my-policy":{
					"version":1,
					"modified_date":1700000000000,
					"policy":{"phases":{}},
					"in_use_by":{"indices":[],"data_streams":[],"composable_templates":[]}
				}}`)
			},
			wantIndices: nil,
		},
		{
			name:       "some indices match",
			policyName: "my-policy",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/_ilm/policy/my-policy" {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintf(w, `{"error":"unexpected path: %s"}`, r.URL.Path)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"my-policy":{
					"version":1,
					"modified_date":1700000000000,
					"policy":{"phases":{}},
					"in_use_by":{
						"indices":[".ds-logs-test-default-2026.01.01-000001",".ds-logs-test-default-2026.01.02-000001"],
						"data_streams":["logs-test-default"],
						"composable_templates":[]
					}
				}}`)
			},
			wantIndices: []string{".ds-logs-test-default-2026.01.01-000001", ".ds-logs-test-default-2026.01.02-000001"},
		},
		{
			name:       "policy missing from response body",
			policyName: "my-policy",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{}`)
			},
			wantIndices: nil,
		},
		{
			name:       "http 404 returns empty",
			policyName: "my-policy",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, `{"error":{"root_cause":[{"type":"resource_not_found_exception","reason":"policy not found"}],"type":"resource_not_found_exception","status":404}}`)
			},
			wantIndices:  nil,
			wantHasError: false,
		},
		{
			name:       "http 500 returns error",
			policyName: "my-policy",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error":{"type":"exception","reason":"something went wrong"},"status":500}`)
			},
			wantIndices:  nil,
			wantHasError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := newMockElasticsearchServerForILM(t, tc.handler)
			defer srv.Close()

			apiClient := newMockScopedClient(t, srv)
			got, diags := GetIndicesWithILMPolicy(context.Background(), apiClient, tc.policyName)

			if tc.wantHasError {
				require.True(t, diags.HasError(), "expected error diagnostics")
				return
			}
			require.False(t, diags.HasError(), "unexpected error diagnostics: %v", diags.Errors())

			if tc.wantIndices == nil {
				require.Empty(t, got)
			} else {
				require.ElementsMatch(t, tc.wantIndices, got)
			}
		})
	}
}

func TestClearILMPolicyFromIndices(t *testing.T) {
	t.Parallel()

	t.Run("empty indices slice is a no-op", func(t *testing.T) {
		t.Parallel()
		callCount := 0
		srv := newMockElasticsearchServerForILM(t, func(w http.ResponseWriter, _ *http.Request) {
			callCount++
			w.WriteHeader(http.StatusOK)
		})
		defer srv.Close()

		apiClient := newMockScopedClient(t, srv)
		diags := ClearILMPolicyFromIndices(context.Background(), apiClient, []string{})
		require.False(t, diags.HasError(), "unexpected error diagnostics: %v", diags.Errors())
		require.Equal(t, 0, callCount, "expected no HTTP requests for empty indices")
	})

	t.Run("non-empty indices makes correct request", func(t *testing.T) {
		t.Parallel()
		var capturedPath string
		var capturedBody string
		srv := newMockElasticsearchServerForILM(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "_settings") {
				capturedPath = r.URL.Path
				body, _ := io.ReadAll(r.Body)
				capturedBody = string(body)
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"acknowledged":true}`)
		})
		defer srv.Close()

		apiClient := newMockScopedClient(t, srv)
		diags := ClearILMPolicyFromIndices(context.Background(), apiClient, []string{"index-1", "index-2"})
		require.False(t, diags.HasError(), "unexpected error diagnostics: %v", diags.Errors())
		require.Equal(t, "/index-1,index-2/_settings", capturedPath)
		require.Contains(t, capturedBody, `"index.lifecycle.name":null`)
	})

	t.Run("http 500 returns error", func(t *testing.T) {
		t.Parallel()
		srv := newMockElasticsearchServerForILM(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "_settings") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error":{"type":"exception","reason":"something went wrong"},"status":500}`)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"acknowledged":true}`)
		})
		defer srv.Close()

		apiClient := newMockScopedClient(t, srv)
		diags := ClearILMPolicyFromIndices(context.Background(), apiClient, []string{"index-1"})
		require.True(t, diags.HasError(), "expected error diagnostics")
	})
}
