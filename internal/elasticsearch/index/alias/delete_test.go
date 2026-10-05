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

package alias

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	elasticsearchclient "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestDeleteAlias_VirtualStateSkipsAPICalls(t *testing.T) {
	t.Parallel()

	apiCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		apiCalls++
	}))
	t.Cleanup(server.Close)

	typedClient, err := elasticsearchclient.NewTypedClient(elasticsearchclient.Config{
		Addresses: []string{server.URL},
	})
	require.NoError(t, err)
	client := clients.NewElasticsearchScopedClientForTest(typedClient, []string{server.URL})

	state := tfModel{
		WriteIndex:  types.ObjectNull(getIndexAttrTypes(context.Background())),
		ReadIndices: mustReadIndexSet(context.Background(), t, virtualReadIndexModel("traces-apm*")),
	}

	diags := deleteAlias(context.Background(), client, "traces", state)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Zero(t, apiCalls)
}

func TestDeleteAlias_RemovesLiveConcreteMembersInsteadOfStateSelector(t *testing.T) {
	t.Parallel()

	var getAliasCalls int
	var removeIndices []string
	var requestDecodeErr error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/_alias/traces":
			getAliasCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"traces-apm-default": map[string]any{
					"aliases": map[string]any{"traces": map[string]any{"is_write_index": true}},
				},
				"traces-apm.rum-default": map[string]any{
					"aliases": map[string]any{"traces": map[string]any{}},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/_aliases":
			var request struct {
				Actions []map[string]map[string]string `json:"actions"`
			}
			requestDecodeErr = json.NewDecoder(r.Body).Decode(&request)
			if requestDecodeErr != nil {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
			for _, action := range request.Actions {
				removeIndices = append(removeIndices, action["remove"]["index"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"acknowledged": true})
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	typedClient, err := elasticsearchclient.NewTypedClient(elasticsearchclient.Config{
		Addresses: []string{server.URL},
	})
	require.NoError(t, err)
	client := clients.NewElasticsearchScopedClientForTest(typedClient, []string{server.URL})

	state := tfModel{
		ReadIndices: mustReadIndexSet(context.Background(), t, readIndexModelWithRouting("traces-apm*", "")),
	}

	diags := deleteAlias(context.Background(), client, "traces", state)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.NoError(t, requestDecodeErr)
	require.Equal(t, 1, getAliasCalls)
	require.ElementsMatch(t, []string{"traces-apm-default", "traces-apm.rum-default"}, removeIndices)
	require.NotContains(t, removeIndices, "traces-apm*")
}

func TestDeleteAlias_WithoutLiveAssociationsSkipsUpdate(t *testing.T) {
	t.Parallel()

	var getAliasCalls int
	var updateAliasCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/_alias/traces":
			getAliasCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{})
		case r.Method == http.MethodPost && r.URL.Path == "/_aliases":
			updateAliasCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{"acknowledged": true})
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	typedClient, err := elasticsearchclient.NewTypedClient(elasticsearchclient.Config{
		Addresses: []string{server.URL},
	})
	require.NoError(t, err)
	client := clients.NewElasticsearchScopedClientForTest(typedClient, []string{server.URL})

	state := tfModel{
		ReadIndices: mustReadIndexSet(context.Background(), t, readIndexModelWithRouting("traces-apm*", "")),
	}

	diags := deleteAlias(context.Background(), client, "traces", state)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Equal(t, 1, getAliasCalls)
	require.Zero(t, updateAliasCalls)
}
