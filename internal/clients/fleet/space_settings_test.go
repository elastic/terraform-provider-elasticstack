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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/stretchr/testify/require"
)

func TestGetSpaceSettings_RoutesToSpacePath(t *testing.T) {
	var capturedMethod, capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod, capturedPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"item":{"allowed_namespace_prefixes":["team_a"],"managed_by":"kibana"}}`)
	}))
	defer srv.Close()

	settings, diags := fleet.GetSpaceSettings(context.Background(), newTestClient(t, srv), "team-a")

	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, http.MethodGet, capturedMethod)
	require.Equal(t, "/s/team-a/api/fleet/space_settings", capturedPath)
	require.NotNil(t, settings)
	require.Equal(t, []string{"team_a"}, settings.AllowedNamespacePrefixes)
	require.Equal(t, "kibana", *settings.ManagedBy)
}

func TestGetSpaceSettings_NotFoundReturnsNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"space not found"}`)
	}))
	defer srv.Close()

	settings, diags := fleet.GetSpaceSettings(context.Background(), newTestClient(t, srv), "missing")

	require.False(t, diags.HasError(), "%v", diags)
	require.Nil(t, settings)
}

func TestPutSpaceSettings_SendsPrefixesToSpacePath(t *testing.T) {
	var capturedMethod, capturedPath string
	var capturedBody map[string][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod, capturedPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"item":{"allowed_namespace_prefixes":["team_a","shared"]}}`)
	}))
	defer srv.Close()

	diags := fleet.PutSpaceSettings(context.Background(), newTestClient(t, srv), "team-a", []string{"team_a", "shared"})

	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, http.MethodPut, capturedMethod)
	require.Equal(t, "/s/team-a/api/fleet/space_settings", capturedPath)
	require.Equal(t, map[string][]string{"allowed_namespace_prefixes": {"team_a", "shared"}}, capturedBody)
}

func TestPutSpaceSettings_APIErrorIsReported(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"statusCode":403,"error":"Forbidden","message":"missing fleet-settings-all"}`)
	}))
	defer srv.Close()

	diags := fleet.PutSpaceSettings(context.Background(), newTestClient(t, srv), "team-a", []string{})

	require.True(t, diags.HasError())
}

func TestResetSpaceSettings(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		wantError bool
	}{
		{name: "success", status: http.StatusOK},
		{name: "not found counts as success", status: http.StatusNotFound},
		{name: "other errors are reported", status: http.StatusForbidden, wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedPath string
			var capturedBody map[string][]string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				_ = json.NewDecoder(r.Body).Decode(&capturedBody)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, `{"item":{"allowed_namespace_prefixes":[]}}`)
			}))
			defer srv.Close()

			diags := fleet.ResetSpaceSettings(context.Background(), newTestClient(t, srv), "team-a")

			require.Equal(t, tc.wantError, diags.HasError(), "%v", diags)
			require.Equal(t, "/s/team-a/api/fleet/space_settings", capturedPath)
			require.Equal(t, map[string][]string{"allowed_namespace_prefixes": {}}, capturedBody)
		})
	}
}
