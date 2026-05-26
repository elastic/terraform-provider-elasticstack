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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// newScopedClientForIDTest creates an ElasticsearchScopedClient backed by srv.
func newScopedClientForIDTest(t *testing.T, srv *httptest.Server) *clients.ElasticsearchScopedClient {
	t.Helper()
	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{srv.URL},
		Username:  "elastic",
		Password:  "changeme",
	})
	require.NoError(t, err)
	return clients.NewElasticsearchScopedClientForTest(esClient, []string{srv.URL})
}

// TestResolveDataSourceID_success verifies that the target is populated with the
// composite ID (<cluster_uuid>/<resourceID>) when the cluster UUID is available.
func TestResolveDataSourceID_success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	srv := newMockElasticsearchStatusServer("8.0.0")
	defer srv.Close()

	client := newScopedClientForIDTest(t, srv)

	var target types.String
	diags := ResolveDataSourceID(ctx, client, "my-resource", &target)

	require.False(t, diags.HasError(), "must not return errors: %v", diags)
	require.Equal(t, "test-cluster/my-resource", target.ValueString())
}

// TestResolveDataSourceID_clusterIDError verifies that target is not modified
// and that the returned diagnostics contain the error when the cluster UUID
// cannot be resolved.
func TestResolveDataSourceID_clusterIDError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Server returns a "_na_" cluster UUID, which ElasticsearchScopedClient.ClusterID
	// treats as unavailable and converts to an error diagnostic.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			fmt.Fprint(w, `{"cluster_uuid":"_na_","version":{"number":"8.0.0","build_flavor":"default"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := newScopedClientForIDTest(t, srv)

	var target types.String
	diags := ResolveDataSourceID(ctx, client, "my-resource", &target)

	require.True(t, diags.HasError(), "must return an error diagnostic when cluster UUID is unavailable")
	require.True(t, target.IsNull(), "target must remain null when an error occurs")
}
