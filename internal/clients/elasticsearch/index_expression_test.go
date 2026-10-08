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

func TestResolveIndexExpression(t *testing.T) {
	tests := []struct {
		name              string
		expression        string
		assertRequest     func(*testing.T, *http.Request)
		response          string
		wantKind          IndexTargetKind
		wantNames         []string
		wantDiagnosticFor string
		excludedAlias     string
	}{
		{
			name:       "resolves wildcard targets",
			expression: "traces-apm*",
			assertRequest: func(t *testing.T, r *http.Request) {
				require.Equal(t, "/_resolve/index/traces-apm*", r.URL.Path)
				require.Equal(t, "all", r.URL.Query().Get("expand_wildcards"))
				require.Equal(t, "true", r.URL.Query().Get("allow_no_indices"))
				require.Equal(t, "true", r.URL.Query().Get("ignore_unavailable"))
			},
			response:  `{"indices":[{"name":"traces-apm-default"},{"name":"traces-apm.rum-default"}],"aliases":[],"data_streams":[]}`,
			wantKind:  RegularIndexTarget,
			wantNames: []string{"traces-apm-default", "traces-apm.rum-default"},
		},
		{
			name:       "resolves comma separated targets",
			expression: "logs-a,logs-b",
			assertRequest: func(t *testing.T, r *http.Request) {
				require.Equal(t, "/_resolve/index/logs-a,logs-b", r.URL.Path)
			},
			response:  `{"indices":[{"name":"logs-a"},{"name":"logs-b"}],"aliases":[],"data_streams":[]}`,
			wantKind:  RegularIndexTarget,
			wantNames: []string{"logs-a", "logs-b"},
		},
		{
			name:       "resolves exclusion expression",
			expression: "logs-*,-logs-skip",
			assertRequest: func(t *testing.T, r *http.Request) {
				require.Equal(t, "/_resolve/index/logs-*,-logs-skip", r.URL.Path)
			},
			response:  `{"indices":[{"name":"logs-keep"}],"aliases":[],"data_streams":[]}`,
			wantKind:  RegularIndexTarget,
			wantNames: []string{"logs-keep"},
		},
		{
			name:       "resolves all expression",
			expression: "_all",
			assertRequest: func(t *testing.T, r *http.Request) {
				require.Equal(t, "/_resolve/index/_all", r.URL.Path)
			},
			response:  `{"indices":[{"name":"logs-1"},{"name":"metrics-1"}],"aliases":[],"data_streams":[]}`,
			wantKind:  RegularIndexTarget,
			wantNames: []string{"logs-1", "metrics-1"},
		},
		{
			name:       "allows no matches",
			expression: "missing-*",
			assertRequest: func(t *testing.T, r *http.Request) {
				require.Equal(t, "true", r.URL.Query().Get("allow_no_indices"))
			},
			response: `{"indices":[],"aliases":[],"data_streams":[]}`,
			wantKind: RegularIndexTarget,
		},
		{
			name:              "rejects alias targets",
			expression:        "*",
			response:          `{"indices":[],"aliases":[{"name":"existing-alias","indices":["logs-1"]}],"data_streams":[]}`,
			wantDiagnosticFor: "existing-alias",
		},
		{
			name:              "rejects unrelated alias after excluding managed alias",
			expression:        "*",
			response:          `{"indices":[{"name":"logs-1"}],"aliases":[{"name":"managed-alias","indices":["logs-1"]},{"name":"unrelated-alias","indices":["logs-1"]}],"data_streams":[]}`,
			wantDiagnosticFor: "unrelated-alias",
			excludedAlias:     "managed-alias",
		},
		{
			name:              "rejects remote index targets",
			expression:        "remote-cluster:logs-*",
			response:          `{"indices":[{"name":"remote-cluster:logs-1"}],"aliases":[],"data_streams":[]}`,
			wantDiagnosticFor: "remote-cluster:logs-1",
		},
		{
			name:              "rejects remote data stream targets",
			expression:        "remote-cluster:logs-*",
			response:          `{"indices":[],"aliases":[],"data_streams":[{"name":"remote-cluster:logs-default","backing_indices":[]}]}`,
			wantDiagnosticFor: "remote-cluster:logs-default",
		},
		{
			name:       "resolves data streams without backing indices",
			expression: "logs-*",
			response:   `{"indices":[{"name":".ds-logs-default-000001","data_stream":"logs-default"}],"aliases":[],"data_streams":[{"name":"logs-default","backing_indices":[".ds-logs-default-000001"]}]}`,
			wantKind:   DataStreamTarget,
			wantNames:  []string{"logs-default"},
		},
		{
			name:       "keeps explicitly resolved backing index",
			expression: ".ds-logs-default-000001",
			response:   `{"indices":[{"name":".ds-logs-default-000001","data_stream":"logs-default"}],"aliases":[],"data_streams":[]}`,
			wantKind:   RegularIndexTarget,
			wantNames:  []string{".ds-logs-default-000001"},
		},
		{
			name:       "rejects mixed target kinds",
			expression: "*",
			response: `{"indices":[{"name":"logs-1"},` +
				`{"name":".ds-metrics-default-000001","data_stream":"metrics-default"}],` +
				`"aliases":[],"data_streams":[{"name":"metrics-default","backing_indices":[".ds-metrics-default-000001"]}]}`,
			wantDiagnosticFor: "regular indices and data streams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newMockElasticsearchServer(t, func(w http.ResponseWriter, r *http.Request) {
				if tt.assertRequest != nil {
					tt.assertRequest(t, r)
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tt.response)
			})
			defer server.Close()

			targets, diags := ResolveIndexExpression(context.Background(), newMockScopedClient(t, server), tt.expression, tt.excludedAlias)

			if tt.wantDiagnosticFor != "" {
				require.True(t, diags.HasError())
				require.Contains(t, diags.Errors()[0].Detail(), tt.wantDiagnosticFor)
				return
			}

			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
			require.Equal(t, tt.wantKind, targets.Kind)
			require.ElementsMatch(t, tt.wantNames, targets.Names)
		})
	}
}
