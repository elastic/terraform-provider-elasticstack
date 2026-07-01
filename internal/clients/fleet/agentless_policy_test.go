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
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
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

	t.Run("non_2xx_returns_error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"statusCode":400,"error":"Bad Request","message":"invalid body"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.CreateAgentlessPolicy(context.Background(), client, "", kbapi.PostFleetAgentlessPoliciesJSONRequestBody{
			Name:    "test-agentless",
			Package: kbapi.KibanaHTTPAPIsPackagePolicyPackage{Name: "cloud_security_posture", Version: "1.14.0"},
		})

		require.Nil(t, item)
		require.True(t, diags.HasError())
	})
}

func TestReadAgentlessPolicyViaPackagePolicy(t *testing.T) {
	t.Run("success_returns_item", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"item":{"id":"pp-1","created_at":"2026-01-01T00:00:00.000Z","created_by":"elastic"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.ReadAgentlessPolicyViaPackagePolicy(context.Background(), client, "", "pp-1")

		require.False(t, diags.HasError())
		require.NotNil(t, item)
		require.NotNil(t, item.Id)
		require.Equal(t, "pp-1", *item.Id)
	})

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

	t.Run("non_2xx_returns_error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"statusCode":500,"error":"Internal Server Error","message":"boom"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.ReadAgentlessPolicyViaPackagePolicy(context.Background(), client, "", "pp-1")

		require.Nil(t, item)
		require.True(t, diags.HasError())
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
		require.NotNil(t, item.Id)
		require.Equal(t, "pp-1", *item.Id)
	})

	t.Run("non_2xx_returns_error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"statusCode":400,"error":"Bad Request","message":"invalid body"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.UpdateAgentlessPolicyViaPackagePolicy(context.Background(), client, "", "pp-1", kbapi.PackagePolicyRequest{})

		require.Nil(t, item)
		require.True(t, diags.HasError())
	})
}

func TestDeleteAgentlessPolicy(t *testing.T) {
	t.Run("success_no_error", func(t *testing.T) {
		var capturedQuery string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"pp-1"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		diags := fleet.DeleteAgentlessPolicy(context.Background(), client, "", "pp-1", false)

		require.False(t, diags.HasError())
		require.NotContains(t, capturedQuery, "force")
	})

	t.Run("404_is_noop", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"policy not found"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		diags := fleet.DeleteAgentlessPolicy(context.Background(), client, "", "missing", false)

		require.False(t, diags.HasError())
	})

	t.Run("non_2xx_returns_error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"statusCode":400,"error":"Bad Request","message":"invalid request"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		diags := fleet.DeleteAgentlessPolicy(context.Background(), client, "", "pp-1", false)

		require.True(t, diags.HasError())
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
		diags := fleet.DeleteAgentlessPolicy(context.Background(), client, "", "pp-1", true)

		require.False(t, diags.HasError())
		require.Contains(t, capturedQuery, "force=true")
	})
}

func TestDeleteAgentlessPolicy_SpaceAwarePath(t *testing.T) {
	tests := []struct {
		name        string
		spaceID     string
		wantPathPfx string
	}{
		{
			name:        "no space id uses default path",
			spaceID:     "",
			wantPathPfx: "/api/fleet/agentless_policies/pp-1",
		},
		{
			name:        "default space uses default path",
			spaceID:     "default",
			wantPathPfx: "/api/fleet/agentless_policies/pp-1",
		},
		{
			name:        "custom space id prefixes path with /s/{space_id}",
			spaceID:     "my-space",
			wantPathPfx: "/s/my-space/api/fleet/agentless_policies/pp-1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedPath string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"id":"pp-1"}`)
			}))
			defer srv.Close()

			client := newTestClient(t, srv)
			diags := fleet.DeleteAgentlessPolicy(context.Background(), client, tc.spaceID, "pp-1", false)
			require.False(t, diags.HasError())

			require.True(t, strings.HasPrefix(capturedPath, tc.wantPathPfx), "request path = %q, want prefix %q", capturedPath, tc.wantPathPfx)
		})
	}
}
