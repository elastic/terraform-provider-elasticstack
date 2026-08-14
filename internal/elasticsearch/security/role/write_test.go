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

package role

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testLiveClusterUUID  = "live-cluster-uuid-aaaa-bbbb-cccc-dddddddddddd"
	testStaleClusterUUID = "stale-cluster-uuid-1111-2222-3333-444444444444"
	testRoleName         = "test-role"
)

func TestWriteRole_UpdatePreservesPriorID(t *testing.T) {
	t.Parallel()

	staleID := types.StringValue(testStaleClusterUUID + "/" + testRoleName)
	prior := Data{
		ID:   staleID,
		Name: types.StringValue(testRoleName),
	}
	plan := Data{
		ID:   types.StringUnknown(),
		Name: types.StringValue(testRoleName),
	}

	client := newWriteRoleTestClient(t, testLiveClusterUUID)
	result, diags := writeRole(context.Background(), client, entitycore.WriteRequest[Data]{
		Plan:    plan,
		Prior:   &prior,
		WriteID: testRoleName,
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %s", diags)
	assert.Equal(t, staleID, result.Model.ID)
	assert.NotEqual(t, testLiveClusterUUID+"/"+testRoleName, result.Model.ID.ValueString())
}

func newWriteRoleTestClient(t *testing.T, liveUUID string) *clients.ElasticsearchScopedClient {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"cluster_uuid": liveUUID,
				"version": map[string]any{
					"number":       "8.19.0",
					"build_flavor": "default",
				},
			})
			return
		}
		if r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/_security/role/") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"role": map[string]any{"created": false},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "unexpected request: " + r.Method + " " + r.URL.Path})
	}))
	t.Cleanup(srv.Close)

	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{srv.URL},
		Username:  "elastic",
		Password:  "changeme",
	})
	require.NoError(t, err)
	return clients.NewElasticsearchScopedClientForTest(esClient, []string{srv.URL})
}
