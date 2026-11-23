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

package indexsettings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	elasticsearchclient "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/stretchr/testify/require"
)

func writeJSON(w http.ResponseWriter, body any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(body)
}

func writeJSONStatus(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = writeJSON(w, body)
}

func newIndexSettingsServer(t *testing.T, indices map[string]map[string]any, putFails bool) (*clients.ElasticsearchScopedClient, *[]map[string]any) {
	t.Helper()

	var putBodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet && r.URL.Path == "/" {
			_ = writeJSON(w, map[string]any{
				"name":         "test-node",
				"cluster_name": "test-cluster",
				"cluster_uuid": "test-cluster-uuid",
				"version":      map[string]any{"number": "8.15.0"},
				"tagline":      "You Know, for Search",
			})
			return
		}

		indexName := strings.TrimPrefix(r.URL.Path, "/")
		isSettingsUpdate := r.Method == http.MethodPut && strings.HasSuffix(indexName, "/_settings")
		if isSettingsUpdate {
			indexName = strings.TrimSuffix(indexName, "/_settings")
		} else if r.Method != http.MethodGet {
			http.Error(w, "unexpected request: "+r.Method+" "+r.URL.Path, http.StatusNotFound)
			return
		}

		settings, exists := indices[indexName]
		if !exists {
			writeJSONStatus(w, http.StatusNotFound, map[string]any{
				"error": map[string]any{
					"type":   "index_not_found_exception",
					"reason": "no such index [" + indexName + "]",
				},
				"status": http.StatusNotFound,
			})
			return
		}

		if !isSettingsUpdate {
			_ = writeJSON(w, map[string]any{
				indexName: map[string]any{
					"aliases":  map[string]any{},
					"mappings": map[string]any{},
					"settings": settings,
				},
			})
			return
		}

		// Preserve numeric tokens so payload assertions catch precision loss.
		decoder := json.NewDecoder(r.Body)
		decoder.UseNumber()
		var body map[string]any
		if err := decoder.Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		putBodies = append(putBodies, body)
		if putFails {
			writeJSONStatus(w, http.StatusInternalServerError, map[string]any{
				"error": map[string]any{
					"type":   "illegal_argument_exception",
					"reason": "closed index [" + indexName + "]",
				},
				"status": http.StatusInternalServerError,
			})
			return
		}

		if settings == nil {
			settings = make(map[string]any)
			indices[indexName] = settings
		}
		for key, value := range body {
			if value == nil {
				delete(settings, key)
			} else {
				settings[key] = value
			}
		}
		_ = writeJSON(w, map[string]any{"acknowledged": true})
	}))
	t.Cleanup(server.Close)

	typedClient, err := elasticsearchclient.NewTypedClient(elasticsearchclient.Config{Addresses: []string{server.URL}})
	require.NoError(t, err)
	return clients.NewElasticsearchScopedClientForTest(typedClient, []string{server.URL}), &putBodies
}
