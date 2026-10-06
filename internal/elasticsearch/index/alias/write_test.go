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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// newAliasTestServer sets up an httptest server that answers the GET
// /_alias/<name> and POST /_aliases endpoints used by applyResolvedAliasConfig,
// and optionally the cluster Info endpoint used by client.ID.
func newAliasTestServer(t *testing.T, aliasName string, currentMembers map[string]bool, includeInfoEndpoint bool) (*httptest.Server, *int, *int, *[]string) {
	t.Helper()

	getAliasCalls := 0
	updateAliasCalls := 0
	var removeIndices []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")

		switch {
		case includeInfoEndpoint && r.Method == http.MethodGet && r.URL.Path == "/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":         "test-node",
				"cluster_name": "test-cluster",
				"cluster_uuid": "test-cluster-uuid",
				"version": map[string]any{
					"number":                              "8.15.0",
					"build_flavor":                        "default",
					"minimum_wire_compatibility_version":  "7.17.0",
					"minimum_index_compatibility_version": "7.0.0",
				},
				"tagline": "You Know, for Search",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/_alias/"+aliasName:
			getAliasCalls++
			response := map[string]any{}
			for indexName := range currentMembers {
				response[indexName] = map[string]any{
					"aliases": map[string]any{aliasName: map[string]any{}},
				}
			}
			_ = json.NewEncoder(w).Encode(response)
		case r.Method == http.MethodPost && r.URL.Path == "/_aliases":
			updateAliasCalls++
			var request struct {
				Actions []map[string]map[string]string `json:"actions"`
			}
			_ = json.NewDecoder(r.Body).Decode(&request)
			for _, action := range request.Actions {
				if idx, ok := action["remove"]; ok {
					removeIndices = append(removeIndices, idx["index"])
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"acknowledged": true})
		default:
			http.Error(w, "unexpected request: "+r.Method+" "+r.URL.Path, http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	return server, &getAliasCalls, &updateAliasCalls, &removeIndices
}

func newAliasTestClient(t *testing.T, server *httptest.Server) *clients.ElasticsearchScopedClient {
	t.Helper()

	typedClient, err := elasticsearchclient.NewTypedClient(elasticsearchclient.Config{
		Addresses: []string{server.URL},
	})
	require.NoError(t, err)
	return clients.NewElasticsearchScopedClientForTest(typedClient, []string{server.URL})
}

func TestApplyResolvedAliasConfig_RemovesExistingMembersWhenDesiredEmpty(t *testing.T) {
	t.Parallel()

	const aliasName = "traces"
	server, getAliasCalls, updateAliasCalls, removeIndices := newAliasTestServer(t, aliasName, map[string]bool{
		"traces-apm-default": true,
	}, false)
	client := newAliasTestClient(t, server)

	plan := tfModel{
		WriteIndex:  types.ObjectNull(getIndexAttrTypes(context.Background())),
		ReadIndices: types.SetNull(types.ObjectType{AttrTypes: getReadIndexAttrTypes(context.Background())}),
	}

	result, diags := applyResolvedAliasConfig(context.Background(), client, plan, aliasName)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Equal(t, 1, *getAliasCalls)
	require.Equal(t, 1, *updateAliasCalls)
	require.Equal(t, []string{"traces-apm-default"}, *removeIndices)
	require.True(t, result.desiredEmptyAfterWrite)
}

func TestCreateAlias_GeneratesIDBeforeDelegatingToSharedApply(t *testing.T) {
	t.Parallel()

	const aliasName = "traces"
	server, getAliasCalls, _, _ := newAliasTestServer(t, aliasName, nil, true)
	client := newAliasTestClient(t, server)

	plan := tfModel{
		WriteIndex:  types.ObjectNull(getIndexAttrTypes(context.Background())),
		ReadIndices: types.SetNull(types.ObjectType{AttrTypes: getReadIndexAttrTypes(context.Background())}),
	}

	result, diags := createAlias(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:    plan,
		WriteID: aliasName,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Equal(t, 1, *getAliasCalls)
	require.NotEmpty(t, result.Model.ID.ValueString())
	require.Contains(t, result.Model.ID.ValueString(), aliasName)
}

func TestUpdateAlias_DelegatesToSharedApplyWithoutGeneratingID(t *testing.T) {
	t.Parallel()

	const aliasName = "traces"
	server, getAliasCalls, _, _ := newAliasTestServer(t, aliasName, nil, false)
	client := newAliasTestClient(t, server)

	plan := tfModel{
		WriteIndex:  types.ObjectNull(getIndexAttrTypes(context.Background())),
		ReadIndices: types.SetNull(types.ObjectType{AttrTypes: getReadIndexAttrTypes(context.Background())}),
	}

	result, diags := updateAlias(context.Background(), client, entitycore.WriteRequest[tfModel]{
		Plan:    plan,
		WriteID: aliasName,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
	require.Equal(t, 1, *getAliasCalls)
	require.Empty(t, result.Model.ID.ValueString())
}
