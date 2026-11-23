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
// the get request so tests can assert on the `owner` parameter that was sent.
func newAPIKeyServer(t *testing.T, getBody, deleteBody string, lastGetQuery *string) *httptest.Server {
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

type sequencedResponse struct {
	status int
	body   string
}

// newSequencedGetServer returns an httptest.Server that responds to
// successive GET /_security/api_key requests with responses, in order,
// recording the query string of each request in queries (the `owner` flag is
// sent as a query parameter on the Get API key request). Used to test
// GetAPIKey's owner=true-then-unscoped fallback, which issues up to two
// requests.
func newSequencedGetServer(t *testing.T, responses []sequencedResponse, queries *[]string) *httptest.Server {
	t.Helper()
	var callCount int
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/_security/api_key" {
			idx := callCount
			callCount++
			if queries != nil {
				*queries = append(*queries, r.URL.RawQuery)
			}
			if idx >= len(responses) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error":"unexpected extra get request"}`)
				return
			}
			w.WriteHeader(responses[idx].status)
			fmt.Fprint(w, responses[idx].body)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"unexpected request: %s %s"}`, r.Method, r.URL.Path)
	}))
}

func TestGetAPIKey_RestrictToOwnedTrue_NotFoundTreatedAsNonExistent(t *testing.T) {
	t.Parallel()
	var lastQuery string
	srv := newAPIKeyServer(t, `{"api_keys":[]}`, "", &lastQuery)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
	require.Nil(t, apiKey, "a key owned by a different user should be treated as not found")
	require.Contains(t, lastQuery, "owner=true")
}

func TestGetAPIKey_RestrictToOwnedTrue_404TreatedAsNonExistentWithoutFallback(t *testing.T) {
	t.Parallel()
	// Elasticsearch answers an id-scoped Get API key request that matches no
	// key with HTTP 404 and an empty `api_keys` list.
	var queries []string
	srv := newSequencedGetServer(t, []sequencedResponse{
		{status: http.StatusNotFound, body: `{"api_keys":[]}`},
	}, &queries)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
	require.Nil(t, apiKey)
	require.Len(t, queries, 1, "restrict_to_owned=true must not fall back to an unscoped lookup")
	require.Contains(t, queries[0], "owner=true")
}

func TestGetAPIKey_RestrictToOwnedFalse_OwnedKeyFoundWithoutFallback(t *testing.T) {
	t.Parallel()
	// The common case: the key is owned by the connection's user, so the
	// owner=true lookup (authorized by `manage_own_api_key` alone) finds it
	// and no unscoped lookup is made.
	var queries []string
	srv := newSequencedGetServer(t, []sequencedResponse{
		{status: http.StatusOK, body: `{"api_keys":[{"id":"some-id","name":"k","creation":0,"invalidated":false,"username":"me","realm":"default"}]}`},
	}, &queries)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", false)

	require.False(t, diags.HasError())
	require.NotNil(t, apiKey)
	require.Equal(t, "some-id", apiKey.Id)
	require.Len(t, queries, 1)
	require.Contains(t, queries[0], "owner=true")
}

func TestGetAPIKey_RestrictToOwnedFalse_FallsBackWhenNotOwned(t *testing.T) {
	t.Parallel()
	var queries []string
	srv := newSequencedGetServer(t, []sequencedResponse{
		{status: http.StatusNotFound, body: `{"api_keys":[]}`},
		{status: http.StatusOK, body: `{"api_keys":[{"id":"some-id","name":"k","creation":0,"invalidated":false,"username":"other","realm":"default"}]}`},
	}, &queries)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", false)

	require.False(t, diags.HasError())
	require.NotNil(t, apiKey)
	require.Equal(t, "some-id", apiKey.Id)
	require.Len(t, queries, 2, "expected an owner=true lookup followed by an unscoped fallback")
	require.Contains(t, queries[0], "owner=true")
	require.NotContains(t, queries[1], "owner=true")
}

func TestGetAPIKey_RestrictToOwnedFalse_NotFoundAnywhere(t *testing.T) {
	t.Parallel()
	var queries []string
	srv := newSequencedGetServer(t, []sequencedResponse{
		{status: http.StatusNotFound, body: `{"api_keys":[]}`},
		{status: http.StatusNotFound, body: `{"api_keys":[]}`},
	}, &queries)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", false)

	require.False(t, diags.HasError())
	require.Nil(t, apiKey)
	require.Len(t, queries, 2)
}

func TestGetAPIKey_RestrictToOwnedFalse_FallbackForbiddenExplainsPrivileges(t *testing.T) {
	t.Parallel()
	// A connection holding only `manage_own_api_key` is not allowed to look
	// up keys regardless of owner. The resulting error should point at the
	// missing privilege and the restrict_to_owned escape hatch.
	var queries []string
	srv := newSequencedGetServer(t, []sequencedResponse{
		{status: http.StatusNotFound, body: `{"api_keys":[]}`},
		{status: http.StatusForbidden, body: `{"error":{"type":"security_exception","reason":"action [cluster:admin/xpack/security/api_key/get] is unauthorized"},"status":403}`},
	}, &queries)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", false)

	require.True(t, diags.HasError())
	require.Nil(t, apiKey)
	require.Len(t, queries, 2)
	require.Contains(t, diags[0].Detail(), "manage_api_key")
	require.Contains(t, diags[0].Detail(), "restrict_to_owned")
}

func TestGetAPIKey_Found(t *testing.T) {
	t.Parallel()
	srv := newAPIKeyServer(t, `{"api_keys":[{"id":"some-id","name":"k","creation":0,"invalidated":false,"username":"me","realm":"default"}]}`, "", nil)
	defer srv.Close()

	apiKey, diags := GetAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
	require.NotNil(t, apiKey)
	require.Equal(t, "some-id", apiKey.Id)
}

func TestDeleteAPIKey_RestrictToOwnedTrue_InvalidatedApiKeys_NoError(t *testing.T) {
	t.Parallel()
	srv := newAPIKeyServer(t, "", `{"invalidated_api_keys":["some-id"],"previously_invalidated_api_keys":[],"error_count":0}`, nil)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
}

func TestDeleteAPIKey_RestrictToOwnedTrue_PreviouslyInvalidatedApiKeys_NoError(t *testing.T) {
	t.Parallel()
	srv := newAPIKeyServer(t, "", `{"invalidated_api_keys":[],"previously_invalidated_api_keys":["some-id"],"error_count":0}`, nil)
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

// newNotFoundThenDeleteServer returns an httptest.Server whose first
// DELETE /_security/api_key request fails with HTTP 404 and whose later
// requests (if any) report id as invalidated. Request bodies are recorded in
// bodies.
func newNotFoundThenDeleteServer(t *testing.T, id string, bodies *[]string) *httptest.Server {
	t.Helper()
	var callCount int
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodDelete || r.URL.Path != "/_security/api_key" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"unexpected request: %s %s"}`, r.Method, r.URL.Path)
			return
		}
		body, _ := io.ReadAll(r.Body)
		*bodies = append(*bodies, string(body))
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":{"type":"resource_not_found_exception","reason":"no API key owned by requesting user found"},"status":404}`)
			return
		}
		fmt.Fprintf(w, `{"invalidated_api_keys":["%s"],"previously_invalidated_api_keys":[],"error_count":0}`, id)
	}))
}

func TestDeleteAPIKey_RestrictToOwnedFalse_FallsBackWhenOwnerTrueReturnsNotFound(t *testing.T) {
	t.Parallel()
	// A 404 on the owner=true attempt only means no key owned by the current
	// user matched; it must not be treated as a successful delete without
	// trying the unscoped fallback.
	var bodies []string
	srv := newNotFoundThenDeleteServer(t, "some-id", &bodies)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", false)

	require.False(t, diags.HasError())
	require.Len(t, bodies, 2)
	require.Contains(t, bodies[0], `"owner":true`)
	require.Contains(t, bodies[1], `"owner":false`)
}

func TestDeleteAPIKey_RestrictToOwnedTrue_NotFoundIsSuccess(t *testing.T) {
	t.Parallel()
	var bodies []string
	srv := newNotFoundThenDeleteServer(t, "some-id", &bodies)
	defer srv.Close()

	diags := DeleteAPIKey(context.Background(), newAPIKeyScopedClient(t, srv), "some-id", true)

	require.False(t, diags.HasError())
	require.Len(t, bodies, 1, "restrict_to_owned=true must not fall back to an unscoped delete")
}
