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
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRole(t *testing.T) {
	t.Parallel()

	const roleName = "test-role"

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		wantNil      bool
		wantErr      bool
		assertResult func(t *testing.T, result *Role)
	}{
		{
			name: "not found returns nil without error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, `{}`)
			},
			wantNil: true,
		},
		{
			name: "server error returns diagnostic",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error":"boom"}`)
			},
			wantErr: true,
		},
		{
			name: "absent role key returns nil without error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"some-other-role":{"cluster":["all"]}}`)
			},
			wantNil: true,
		},
		{
			name: "global with array-typed category decodes as raw JSON",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/_security/role/"+roleName {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintf(w, `{"error":"unexpected path: %s"}`, r.URL.Path)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"test-role":{
					"cluster":["all"],
					"global":{"data_source":[],"application":{}}
				}}`)
			},
			assertResult: func(t *testing.T, result *Role) {
				require.NotNil(t, result)
				require.JSONEq(t, `{"data_source":[],"application":{}}`, string(result.Global))
				require.Len(t, result.Cluster, 1)
				require.Equal(t, "all", result.Cluster[0].Name)
			},
		},
		{
			name: "explicit null global is treated as absent",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"test-role":{"cluster":["all"],"global":null}}`)
			},
			assertResult: func(t *testing.T, result *Role) {
				require.NotNil(t, result)
				require.Nil(t, result.Global)
			},
		},
		{
			name: "role without global returns nil raw global",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"test-role":{"cluster":["all"]}}`)
			},
			assertResult: func(t *testing.T, result *Role) {
				require.NotNil(t, result)
				require.Nil(t, result.Global)
				require.Len(t, result.Cluster, 1)
				require.Equal(t, "all", result.Cluster[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := newMockElasticsearchServer(t, tt.handler)
			defer srv.Close()

			apiClient := newMockScopedClient(t, srv)

			result, diags := GetRole(context.Background(), apiClient, roleName)

			if tt.wantErr {
				require.True(t, diags.HasError(), "expected error diagnostics")
				return
			}
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			if tt.wantNil {
				require.Nil(t, result)
				return
			}

			if tt.assertResult != nil {
				tt.assertResult(t, result)
			}
		})
	}
}
