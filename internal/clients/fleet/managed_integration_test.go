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
	"sync/atomic"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/stretchr/testify/require"
)

func TestCreateManagedIntegration(t *testing.T) {
	t.Run("success_returns_item", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"item":{"id":"mi-1","created_at":"2026-01-01T00:00:00.000Z","created_by":"elastic"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.CreateManagedIntegration(context.Background(), client, "", kbapi.PostFleetManagedIntegrationsJSONRequestBody{
			Name:    "test-managed-integration",
			Package: kbapi.KibanaHTTPAPIsPackagePolicyPackage{Name: "cloud_security_posture", Version: "1.14.0"},
		})

		require.False(t, diags.HasError())
		require.NotNil(t, item)
		require.Equal(t, "mi-1", item.Id)
	})

	t.Run("non_2xx_returns_error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"statusCode":400,"error":"Bad Request","message":"invalid body"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.CreateManagedIntegration(context.Background(), client, "", kbapi.PostFleetManagedIntegrationsJSONRequestBody{
			Name:    "test-managed-integration",
			Package: kbapi.KibanaHTTPAPIsPackagePolicyPackage{Name: "cloud_security_posture", Version: "1.14.0"},
		})

		require.Nil(t, item)
		require.True(t, diags.HasError())
	})

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
			fmt.Fprint(w, `{"item":{"id":"mi-1","name":"test-managed-integration"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.CreateManagedIntegration(context.Background(), client, "", kbapi.PostFleetManagedIntegrationsJSONRequestBody{
			Name:      "test-managed-integration",
			Namespace: new("default"),
		})

		require.False(t, diags.HasError())
		require.NotNil(t, item)
		require.Equal(t, "mi-1", item.Id)
		require.Equal(t, int64(3), calls.Load())
	})
}

func TestReadManagedIntegration(t *testing.T) {
	t.Run("success_returns_item", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"item":{"id":"mi-1","created_at":"2026-01-01T00:00:00.000Z","created_by":"elastic"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.ReadManagedIntegration(context.Background(), client, "", "mi-1")

		require.False(t, diags.HasError())
		require.NotNil(t, item)
		require.NotNil(t, item.Id)
		require.Equal(t, "mi-1", item.Id)
	})

	t.Run("404_returns_nil_no_error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"managed integration not found"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.ReadManagedIntegration(context.Background(), client, "", "missing")

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
		item, diags := fleet.ReadManagedIntegration(context.Background(), client, "", "mi-1")

		require.Nil(t, item)
		require.True(t, diags.HasError())
	})
}

func TestUpdateManagedIntegration(t *testing.T) {
	t.Run("success_returns_item", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"item":{"id":"mi-1","created_at":"2026-01-01T00:00:00.000Z","created_by":"elastic"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.UpdateManagedIntegration(context.Background(), client, "", "mi-1", kbapi.PutFleetManagedIntegrationsPolicyidJSONRequestBody{})

		require.False(t, diags.HasError())
		require.NotNil(t, item)
		require.NotNil(t, item.Id)
		require.Equal(t, "mi-1", item.Id)
	})

	t.Run("non_2xx_returns_error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"statusCode":400,"error":"Bad Request","message":"invalid body"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.UpdateManagedIntegration(context.Background(), client, "", "mi-1", kbapi.PutFleetManagedIntegrationsPolicyidJSONRequestBody{})

		require.Nil(t, item)
		require.True(t, diags.HasError())
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
			fmt.Fprint(w, `{"item":{"id":"mi-1","name":"updated"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		item, diags := fleet.UpdateManagedIntegration(context.Background(), client, "", "mi-1", kbapi.PutFleetManagedIntegrationsPolicyidJSONRequestBody{
			Name: "updated",
		})

		require.False(t, diags.HasError())
		require.NotNil(t, item)
		require.Equal(t, int64(2), calls.Load())
	})
}

func TestDeleteManagedIntegration(t *testing.T) {
	t.Run("success_no_error", func(t *testing.T) {
		var capturedQuery string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"mi-1"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		isConflict, diags := fleet.DeleteManagedIntegration(context.Background(), client, "", "mi-1", false)

		require.False(t, diags.HasError())
		require.False(t, isConflict)
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
		isConflict, diags := fleet.DeleteManagedIntegration(context.Background(), client, "", "missing", false)

		require.False(t, diags.HasError())
		require.False(t, isConflict)
	})

	t.Run("non_2xx_returns_error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"statusCode":400,"error":"Bad Request","message":"invalid request"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		isConflict, diags := fleet.DeleteManagedIntegration(context.Background(), client, "", "mi-1", false)

		require.True(t, diags.HasError())
		require.False(t, isConflict, "a non-409 error must not be reported as a conflict")
	})

	t.Run("409_reports_conflict", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `{"statusCode":409,"error":"Conflict","message":"agent policy is provisioning"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		isConflict, diags := fleet.DeleteManagedIntegration(ctx, client, "", "mi-1", false)

		require.True(t, diags.HasError())
		require.True(t, isConflict, "a 409 response must be reported as a conflict so delete.go can offer the force_delete hint")
	})

	t.Run("force_delete_sets_force_query_param", func(t *testing.T) {
		var capturedQuery string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"mi-1"}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		isConflict, diags := fleet.DeleteManagedIntegration(context.Background(), client, "", "mi-1", true)

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
		_, diags := fleet.DeleteManagedIntegration(context.Background(), client, "", "mi-1", false)

		require.True(t, diags.HasError())
		require.Equal(t, int64(kibanautil.ConflictMaxAttempts), calls.Load())
	})
}

func TestManagedIntegration_SpaceAwarePath(t *testing.T) {
	tests := []struct {
		name        string
		spaceID     string
		wantPathPfx string
	}{
		{
			name:        "no space id uses default path",
			spaceID:     "",
			wantPathPfx: "/api/fleet/managed_integrations/mi-1",
		},
		{
			name:        "default space uses default path",
			spaceID:     "default",
			wantPathPfx: "/api/fleet/managed_integrations/mi-1",
		},
		{
			name:        "custom space id prefixes path with /s/{space_id}",
			spaceID:     "my-space",
			wantPathPfx: "/s/my-space/api/fleet/managed_integrations/mi-1",
		},
	}

	for _, tc := range tests {
		t.Run("read_"+tc.name, func(t *testing.T) {
			var capturedPath string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"item":{"id":"mi-1"}}`)
			}))
			defer srv.Close()

			client := newTestClient(t, srv)
			_, diags := fleet.ReadManagedIntegration(context.Background(), client, tc.spaceID, "mi-1")
			require.False(t, diags.HasError())
			require.True(t, strings.HasPrefix(capturedPath, tc.wantPathPfx), "request path = %q, want prefix %q", capturedPath, tc.wantPathPfx)
		})

		t.Run("delete_"+tc.name, func(t *testing.T) {
			var capturedPath string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"id":"mi-1"}`)
			}))
			defer srv.Close()

			client := newTestClient(t, srv)
			_, diags := fleet.DeleteManagedIntegration(context.Background(), client, tc.spaceID, "mi-1", false)
			require.False(t, diags.HasError())
			require.True(t, strings.HasPrefix(capturedPath, tc.wantPathPfx), "request path = %q, want prefix %q", capturedPath, tc.wantPathPfx)
		})
	}

	t.Run("create_custom_space", func(t *testing.T) {
		var capturedPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"item":{"id":"mi-1"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		_, diags := fleet.CreateManagedIntegration(context.Background(), client, "my-space", kbapi.PostFleetManagedIntegrationsJSONRequestBody{
			Name: "test",
		})
		require.False(t, diags.HasError())
		require.True(t, strings.HasPrefix(capturedPath, "/s/my-space/api/fleet/managed_integrations"))
	})

	t.Run("update_custom_space", func(t *testing.T) {
		var capturedPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"item":{"id":"mi-1"}}`)
		}))
		defer srv.Close()

		client := newTestClient(t, srv)
		_, diags := fleet.UpdateManagedIntegration(context.Background(), client, "my-space", "mi-1", kbapi.PutFleetManagedIntegrationsPolicyidJSONRequestBody{})
		require.False(t, diags.HasError())
		require.True(t, strings.HasPrefix(capturedPath, "/s/my-space/api/fleet/managed_integrations/mi-1"))
	})
}
