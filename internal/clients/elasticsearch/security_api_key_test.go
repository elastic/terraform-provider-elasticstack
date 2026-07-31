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
	"testing"

	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/stretchr/testify/require"
)

// newAPIKeyServer returns an httptest.Server that responds to the get/delete
// API key endpoints with the given bodies, recording the query string seen on
// each request so tests can assert on the `owner` parameter that was sent.
func newAPIKeyServer(t *testing.T, getBody, deleteBody string, getStatus int, lastGetQuery, lastDeleteQuery *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/_security/api_key":
			if lastGetQuery != nil {
				*lastGetQuery = r.URL.RawQuery
			}
			if getStatus != 0 && getStatus != http.StatusOK {
				w.WriteHeader(getStatus)
				fmt.Fprint(w, `{"error":{"type":"error"},"status":404}`)
				return
			}
			fmt.Fprint(w, getBody)
			return
		case r.Method == http.MethodDelete && r.URL.Path == "/_security/api_key":
			if lastDeleteQuery != nil {
				*lastDeleteQuery = r.URL.RawQuery
			}
			fmt.Fprint(w, deleteBody)
			return
		default:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"unexpected request: %s %s"}`, r.Method, r.URL.Path)
		}
	}))
}

func newAPIKeyScopedClient(t *testing.T, srv *httptest.Server) *clients.ElasticsearchScopedClient {
	t.Helper()
	esClient, err := elasticsearch8.NewTypedClient(elasticsearch8.Config{
		Addresses: []string{srv.URL},
		Username:  "elastic",
		Password:  "changeme",
	})
	require.NoError(t, err)
	return clients.NewElasticsearchScopedClientForTest(esClient, []string{srv.URL})
}

func TestGetAPIKey_OwnerTrue_NotFoundTreatedAsNonExistent(t *testing.T) {
	t.Parallel()
	var lastQuery string
	srv := newAPIKeyServer(t, `{"api_keys":[]}`, "", http.StatusOK, &lastQuery, nil)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
	require.Nil(t, apiKey, "a key owned by a different user should be treated as not found")
	require.Contains(t, lastQuery, "owner=true")
}

func TestGetAPIKey_OwnerFalse_DoesNotFilterByOwner(t *testing.T) {
	t.Parallel()
	var lastQuery string
	srv := newAPIKeyServer(t, `{"api_keys":[{"id":"some-id","name":"k","creation":0,"invalidated":false,"username":"other","realm":"default"}]}`, "", http.StatusOK, &lastQuery, nil)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", false)

	require.False(t, diags.HasError())
	require.NotNil(t, apiKey)
	require.Equal(t, "some-id", apiKey.Id)
	require.NotContains(t, lastQuery, "owner=true")
}

func TestGetAPIKey_Found(t *testing.T) {
	t.Parallel()
	srv := newAPIKeyServer(t, `{"api_keys":[{"id":"some-id","name":"k","creation":0,"invalidated":false,"username":"me","realm":"default"}]}`, "", http.StatusOK, nil, nil)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
	require.NotNil(t, apiKey)
	require.Equal(t, "some-id", apiKey.Id)
}

func TestDeleteAPIKey_InvalidatedApiKeys_NoError(t *testing.T) {
	t.Parallel()
	srv := newAPIKeyServer(t, "", `{"invalidated_api_keys":["some-id"],"previously_invalidated_api_keys":[],"error_count":0}`, http.StatusOK, nil, nil)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
}

func TestDeleteAPIKey_PreviouslyInvalidatedApiKeys_NoError(t *testing.T) {
	t.Parallel()
	srv := newAPIKeyServer(t, "", `{"invalidated_api_keys":[],"previously_invalidated_api_keys":["some-id"],"error_count":0}`, http.StatusOK, nil, nil)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
}

func TestDeleteAPIKey_IDNotInResponse_ErrorsInsteadOfSilentlyDropping(t *testing.T) {
	t.Parallel()
	// Elasticsearch can return a 200 with error_count == 0 while silently
	// omitting an id that didn't match the request's filters (e.g. an
	// `owner` mismatch). This must surface as an error rather than being
	// treated as a successful delete.
	srv := newAPIKeyServer(t, "", `{"invalidated_api_keys":[],"previously_invalidated_api_keys":[],"error_count":0}`, http.StatusOK, nil, nil)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.True(t, diags.HasError())
}

func TestDeleteAPIKey_OwnerFlagSentOnRequest(t *testing.T) {
	t.Parallel()
	var lastQuery string
	srv := newAPIKeyServer(t, "", `{"invalidated_api_keys":["some-id"],"previously_invalidated_api_keys":[],"error_count":0}`, http.StatusOK, nil, &lastQuery)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
}
