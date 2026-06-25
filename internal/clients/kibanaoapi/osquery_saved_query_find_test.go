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

package kibanaoapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newOsqueryTestClient(t *testing.T, srv *httptest.Server) *kibanaoapi.Client {
	t.Helper()

	client, err := kibanaoapi.NewClient(kibanaoapi.Config{URL: srv.URL})
	require.NoError(t, err)
	return client
}

func encodeJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(payload))
}

func TestFindOsquerySavedObjectID(t *testing.T) {
	t.Run("returns saved object id on first page match", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/osquery/saved_queries", r.URL.Path)
			encodeJSON(t, w, map[string]any{
				"data": []map[string]any{
					{"id": "list_processes", "saved_object_id": "uuid-list-processes"},
				},
				"page":     1,
				"per_page": 100,
				"total":    1,
			})
		}))
		t.Cleanup(server.Close)

		savedObjectID, found, diags := kibanaoapi.FindOsquerySavedObjectID(
			context.Background(),
			newOsqueryTestClient(t, server),
			"default",
			"list_processes",
		)
		require.False(t, diags.HasError(), "diags: %v", diags)
		assert.True(t, found)
		assert.Equal(t, "uuid-list-processes", savedObjectID)
	})

	t.Run("paginates until match", func(t *testing.T) {
		t.Parallel()

		page := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/osquery/saved_queries", r.URL.Path)
			page++
			switch page {
			case 1:
				encodeJSON(t, w, map[string]any{
					"data": []map[string]any{
						{"id": "other", "saved_object_id": "uuid-other"},
					},
					"page":     1,
					"per_page": 1,
					"total":    2,
				})
			case 2:
				encodeJSON(t, w, map[string]any{
					"data": []map[string]any{
						{"id": "target_query", "saved_object_id": "uuid-target"},
					},
					"page":     2,
					"per_page": 1,
					"total":    2,
				})
			default:
				t.Fatalf("unexpected page %d", page)
			}
		}))
		t.Cleanup(server.Close)

		savedObjectID, found, diags := kibanaoapi.FindOsquerySavedObjectID(
			context.Background(),
			newOsqueryTestClient(t, server),
			"default",
			"target_query",
		)
		require.False(t, diags.HasError())
		assert.True(t, found)
		assert.Equal(t, "uuid-target", savedObjectID)
		assert.Equal(t, 2, page)
	})

	t.Run("returns not found when absent", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			encodeJSON(t, w, map[string]any{
				"data":     []map[string]any{},
				"page":     1,
				"per_page": 100,
				"total":    0,
			})
		}))
		t.Cleanup(server.Close)

		savedObjectID, found, diags := kibanaoapi.FindOsquerySavedObjectID(
			context.Background(),
			newOsqueryTestClient(t, server),
			"default",
			"missing",
		)
		require.False(t, diags.HasError())
		assert.False(t, found)
		assert.Empty(t, savedObjectID)
	})
}

func TestGetOsquerySavedQuery_resolvesSavedObjectID(t *testing.T) {
	t.Parallel()

	var detailPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/osquery/saved_queries":
			encodeJSON(t, w, map[string]any{
				"data": []map[string]any{
					{"id": "list_processes", "saved_object_id": "uuid-list-processes"},
				},
				"page":     1,
				"per_page": 100,
				"total":    1,
			})
		case "/api/osquery/saved_queries/uuid-list-processes":
			detailPath = r.URL.Path
			encodeJSON(t, w, map[string]any{
				"data": map[string]any{
					"id":              "list_processes",
					"saved_object_id": "uuid-list-processes",
					"query":           "SELECT 1",
				},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	entity, diags := kibanaoapi.GetOsquerySavedQuery(
		context.Background(),
		newOsqueryTestClient(t, server),
		"default",
		"list_processes",
	)
	require.False(t, diags.HasError())
	require.NotNil(t, entity)
	assert.Equal(t, "/api/osquery/saved_queries/uuid-list-processes", detailPath)
	assert.Equal(t, kbapi.SecurityOsqueryAPISavedQueryId("list_processes"), entity.ID)
}
