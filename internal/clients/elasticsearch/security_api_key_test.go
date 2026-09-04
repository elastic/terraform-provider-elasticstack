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
	"io"
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
func newAPIKeyServer(t *testing.T, getBody, deleteBody string, lastGetQuery, lastDeleteQuery *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/_security/api_key":
			if lastGetQuery != nil {
				*lastGetQuery = r.URL.RawQuery
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

// newSequencedDeleteServer returns an httptest.Server that responds to
// successive DELETE /_security/api_key requests with the bodies in
// deleteResponses, in order, recording the request body seen on each request
// in deleteRequestBodies (the `owner` flag is sent in the Invalidate API Key
// request body, not the query string). Used to test DeleteAPIKey's
// owner=true-then-owner=false fallback, which issues up to two requests.
func newSequencedDeleteServer(t *testing.T, deleteResponses []string, deleteRequestBodies *[]string) *httptest.Server {
	t.Helper()
	var callCount int
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete && r.URL.Path == "/_security/api_key" {
			idx := callCount
			callCount++
			if deleteRequestBodies != nil {
				body, _ := io.ReadAll(r.Body)
				*deleteRequestBodies = append(*deleteRequestBodies, string(body))
			}
			if idx >= len(deleteResponses) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error":"unexpected extra delete request"}`)
				return
			}
			fmt.Fprint(w, deleteResponses[idx])
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"unexpected request: %s %s"}`, r.Method, r.URL.Path)
	}))
}

func TestGetAPIKey_RestrictToOwnedTrue_NotFoundTreatedAsNonExistent(t *testing.T) {
	t.Parallel()
	var lastQuery string
	srv := newAPIKeyServer(t, `{"api_keys":[]}`, "", &lastQuery, nil)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
	require.Nil(t, apiKey, "a key owned by a different user should be treated as not found")
	require.Contains(t, lastQuery, "owner=true")
}

func TestGetAPIKey_RestrictToOwnedFalse_DoesNotFilterByOwner(t *testing.T) {
	t.Parallel()
	var lastQuery string
	srv := newAPIKeyServer(t, `{"api_keys":[{"id":"some-id","name":"k","creation":0,"invalidated":false,"username":"other","realm":"default"}]}`, "", &lastQuery, nil)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", false)

	require.False(t, diags.HasError())
	require.NotNil(t, apiKey)
	require.Equal(t, "some-id", apiKey.Id)
	require.NotContains(t, lastQuery, "owner=true")
}

func TestGetAPIKey_Found(t *testing.T) {
	t.Parallel()
	srv := newAPIKeyServer(t, `{"api_keys":[{"id":"some-id","name":"k","creation":0,"invalidated":false,"username":"me","realm":"default"}]}`, "", nil, nil)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
	require.NotNil(t, apiKey)
	require.Equal(t, "some-id", apiKey.Id)
}

func TestDeleteAPIKey_RestrictToOwnedTrue_InvalidatedApiKeys_NoError(t *testing.T) {
	t.Parallel()
	srv := newAPIKeyServer(t, "", `{"invalidated_api_keys":["some-id"],"previously_invalidated_api_keys":[],"error_count":0}`, nil, nil)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
}

func TestDeleteAPIKey_RestrictToOwnedTrue_PreviouslyInvalidatedApiKeys_NoError(t *testing.T) {
	t.Parallel()
	srv := newAPIKeyServer(t, "", `{"invalidated_api_keys":[],"previously_invalidated_api_keys":["some-id"],"error_count":0}`, nil, nil)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
}

func TestDeleteAPIKey_RestrictToOwnedTrue_IDNotInResponse_ErrorsWithoutFallback(t *testing.T) {
	t.Parallel()
	// Elasticsearch can return a 200 with error_count == 0 while silently
	// omitting an id that didn't match the request's filters (e.g. an
	// `owner` mismatch). This must surface as an error rather than being
	// treated as a successful delete, and restrict_to_owned=true must not
	// fall back to an unscoped delete.
	var bodies []string
	srv := newSequencedDeleteServer(t, []string{
		`{"invalidated_api_keys":[],"previously_invalidated_api_keys":[],"error_count":0}`,
	}, &bodies)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.True(t, diags.HasError())
	require.Len(t, bodies, 1, "restrict_to_owned=true must not fall back to an unscoped delete")
	require.Contains(t, bodies[0], `"owner":true`)
}

func TestDeleteAPIKey_RestrictToOwnedFalse_FallsBackWhenOwnerTrueDoesNotInvalidate(t *testing.T) {
	t.Parallel()
	// The common case: the connection only holds `manage_own_api_key`, so
	// the owner=true attempt succeeds and no fallback is needed. This test
	// instead exercises the case where owner=true does not invalidate the
	// key (e.g. it's owned by someone else), and confirms DeleteAPIKey
	// falls back to an unscoped (owner=false) request rather than erroring
	// immediately.
	var bodies []string
	srv := newSequencedDeleteServer(t, []string{
		`{"invalidated_api_keys":[],"previously_invalidated_api_keys":[],"error_count":0}`,
		`{"invalidated_api_keys":["some-id"],"previously_invalidated_api_keys":[],"error_count":0}`,
	}, &bodies)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", false)

	require.False(t, diags.HasError())
	require.Len(t, bodies, 2, "expected an owner=true attempt followed by an owner=false fallback")
	require.Contains(t, bodies[0], `"owner":true`)
	require.Contains(t, bodies[1], `"owner":false`)
}

func TestDeleteAPIKey_RestrictToOwnedFalse_ErrorsWhenBothAttemptsFail(t *testing.T) {
	t.Parallel()
	var bodies []string
	srv := newSequencedDeleteServer(t, []string{
		`{"invalidated_api_keys":[],"previously_invalidated_api_keys":[],"error_count":0}`,
		`{"invalidated_api_keys":[],"previously_invalidated_api_keys":[],"error_count":0}`,
	}, &bodies)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", false)

	require.True(t, diags.HasError())
	require.Len(t, bodies, 2, "expected both the owner=true attempt and the owner=false fallback to have been made")
}
