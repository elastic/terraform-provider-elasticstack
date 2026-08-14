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

package entitycore

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testLiveClusterUUID  = "live-cluster-uuid-aaaa-bbbb-cccc-dddddddddddd"
	testStaleClusterUUID = "stale-cluster-uuid-1111-2222-3333-444444444444"
	testResourceName     = "my-resource"
)

func TestCompositeIDForWrite_CreateComputesFromLiveCluster(t *testing.T) {
	t.Parallel()

	client := newCompositeIDTestClient(t, testLiveClusterUUID)
	req := WriteRequest[testResourceModel]{
		WriteID: testResourceName,
	}

	id, diags := CompositeIDForWrite(context.Background(), client, req)
	require.False(t, diags.HasError(), "unexpected diagnostics: %s", diags)
	assert.Equal(t, testLiveClusterUUID+"/"+testResourceName, id.ValueString())
}

func TestCompositeIDForWrite_UpdatePreservesPriorID(t *testing.T) {
	t.Parallel()

	staleID := types.StringValue(testStaleClusterUUID + "/" + testResourceName)
	prior := testResourceModel{
		ID:   staleID,
		Name: types.StringValue(testResourceName),
	}
	// The live-cluster client would compute a different UUID prefix; Update must
	// ignore it and carry the prior id forward.
	client := newCompositeIDTestClient(t, testLiveClusterUUID)
	req := WriteRequest[testResourceModel]{
		Plan: testResourceModel{
			ID:   types.StringUnknown(),
			Name: types.StringValue(testResourceName),
		},
		Prior:   &prior,
		WriteID: testResourceName,
	}

	id, diags := CompositeIDForWrite(context.Background(), client, req)
	require.False(t, diags.HasError(), "unexpected diagnostics: %s", diags)
	assert.Equal(t, staleID, id)
	assert.NotEqual(t, testLiveClusterUUID+"/"+testResourceName, id.ValueString())
}

func newCompositeIDTestClient(t *testing.T, clusterUUID string) *clients.ElasticsearchScopedClient {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"cluster_uuid": clusterUUID,
				"version": map[string]any{
					"number":       "8.19.0",
					"build_flavor": "default",
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
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
