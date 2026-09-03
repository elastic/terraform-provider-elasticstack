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
	"net/http/httptest"
	"strings"
	"testing"

	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/stretchr/testify/require"
)

// newDataStreamLifecycleServer returns an httptest.Server that advertises the
// given build_flavor on the `/` info endpoint (used by IsServerless) and
// records whether a DELETE to a data stream lifecycle was attempted.
func newDataStreamLifecycleServer(t *testing.T, buildFlavor string, infoStatus int, deleted *bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			if infoStatus != http.StatusOK {
				w.WriteHeader(infoStatus)
				fmt.Fprintf(w, `{"error":"info endpoint failure"}`)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"cluster_uuid":"test-cluster","version":{"number":"8.19.0","build_flavor":%q}}`, buildFlavor)
			return
		}
		if r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/_lifecycle") {
			*deleted = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"acknowledged":true}`)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"unexpected request: %s %s"}`, r.Method, r.URL.Path)
	}))
}

func newDataStreamLifecycleScopedClient(t *testing.T, srv *httptest.Server) *clients.ElasticsearchScopedClient {
	t.Helper()
	esClient, err := elasticsearch8.NewTypedClient(elasticsearch8.Config{
		Addresses: []string{srv.URL},
		Username:  "elastic",
		Password:  "changeme",
	})
	require.NoError(t, err)
	return clients.NewElasticsearchScopedClientForTest(esClient, []string{srv.URL})
}

func TestDeleteDataStreamLifecycle_Serverless_SkipsDelete(t *testing.T) {
	t.Parallel()
	deleted := false
	srv := newDataStreamLifecycleServer(t, "serverless", http.StatusOK, &deleted)
	defer srv.Close()

	diags := DeleteDataStreamLifecycle(context.Background(), newDataStreamLifecycleScopedClient(t, srv), "my-stream", "")

	require.False(t, diags.HasError(), "serverless delete must not error")
	require.False(t, deleted, "the DELETE _lifecycle request must not be sent on serverless")
	require.Len(t, diags, 1, "a single warning diagnostic is expected")
	require.Contains(t, diags[0].Summary(), "skipped on serverless")
}

func TestDeleteDataStreamLifecycle_Stateful_PerformsDelete(t *testing.T) {
	t.Parallel()
	deleted := false
	srv := newDataStreamLifecycleServer(t, "default", http.StatusOK, &deleted)
	defer srv.Close()

	diags := DeleteDataStreamLifecycle(context.Background(), newDataStreamLifecycleScopedClient(t, srv), "my-stream", "")

	require.False(t, diags.HasError(), "stateful delete must not error")
	require.True(t, deleted, "the DELETE _lifecycle request must be sent on a stateful cluster")
	require.Empty(t, diags, "no diagnostics expected on a successful stateful delete")
}

func TestDeleteDataStreamLifecycle_ServerlessCheckError_DoesNotDelete(t *testing.T) {
	t.Parallel()
	deleted := false
	srv := newDataStreamLifecycleServer(t, "default", http.StatusInternalServerError, &deleted)
	defer srv.Close()

	diags := DeleteDataStreamLifecycle(context.Background(), newDataStreamLifecycleScopedClient(t, srv), "my-stream", "")

	require.True(t, diags.HasError(), "a failed serverless check must surface an error diagnostic")
	require.False(t, deleted, "the DELETE _lifecycle request must not be sent when the serverless check fails")
}
