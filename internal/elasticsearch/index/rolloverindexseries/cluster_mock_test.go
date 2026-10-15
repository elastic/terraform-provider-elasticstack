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

package rolloverindexseries

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	elasticsearchclient "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/stretchr/testify/require"
)

const testClusterUUID = "abc123"

type aliasTarget struct {
	write      *bool
	dataStream bool
}

type clusterMock struct {
	mu sync.Mutex

	indices      map[string]struct{}
	aliasObjects map[string]struct{}
	dataStreams  map[string]struct{}
	membership   map[string]map[string]aliasTarget

	requests       []string
	putBodies      []map[string]json.RawMessage
	aliasMutations int
	failAliasCode  int
}

func newClusterMock(t *testing.T) (*clusterMock, *clients.ElasticsearchScopedClient) {
	t.Helper()
	mock := &clusterMock{
		indices:      map[string]struct{}{},
		aliasObjects: map[string]struct{}{},
		dataStreams:  map[string]struct{}{},
		membership:   map[string]map[string]aliasTarget{},
	}
	server := httptest.NewServer(mock)
	t.Cleanup(server.Close)

	typed, err := elasticsearchclient.NewTypedClient(elasticsearchclient.Config{Addresses: []string{server.URL}})
	require.NoError(t, err)
	return mock, clients.NewElasticsearchScopedClientForTest(typed, []string{server.URL})
}

func (m *clusterMock) addIndex(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.indices[name] = struct{}{}
}

func (m *clusterMock) addAliasObject(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.aliasObjects[name] = struct{}{}
}

func (m *clusterMock) addMember(aliasName, target string, write *bool, dataStream bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.putMember(aliasName, target, aliasTarget{write: write, dataStream: dataStream})
	if dataStream {
		m.dataStreams[target] = struct{}{}
	} else {
		m.indices[target] = struct{}{}
	}
}

func (m *clusterMock) putMember(aliasName, target string, member aliasTarget) {
	if m.membership[aliasName] == nil {
		m.membership[aliasName] = map[string]aliasTarget{}
	}
	m.membership[aliasName][target] = member
}

func (m *clusterMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, r.Method+" "+r.URL.Path)

	switch {
	case r.URL.Path == "/":
		writeJSON(w, http.StatusOK, map[string]any{
			"name":         "test-node",
			"cluster_name": "test-cluster",
			"cluster_uuid": testClusterUUID,
			"version": map[string]any{
				"number":       "8.19.0",
				"build_flavor": "default",
			},
			"tagline": "You Know, for Search",
		})
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/_alias/"):
		m.getAlias(w, strings.TrimPrefix(r.URL.Path, "/_alias/"))
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/_resolve/index/"):
		target, err := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/_resolve/index/"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		m.resolve(w, target)
	case r.Method == http.MethodPut && !strings.Contains(r.URL.Path, "/_"):
		m.createIndex(w, r, strings.TrimPrefix(r.URL.Path, "/"))
	case r.Method == http.MethodDelete && !strings.Contains(r.URL.Path, "/_"):
		m.deleteIndex(w, strings.TrimPrefix(r.URL.Path, "/"))
	default:
		if strings.Contains(r.URL.Path, "_alias") {
			m.aliasMutations++
		}
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": map[string]any{
				"type":   "invalid_request",
				"reason": "unexpected " + r.Method + " " + r.URL.Path,
			},
			"status": http.StatusNotFound,
		})
	}
}

func (m *clusterMock) getAlias(w http.ResponseWriter, aliasName string) {
	if m.failAliasCode != 0 {
		writeJSON(w, m.failAliasCode, map[string]any{
			"error": map[string]any{
				"type":   "security_exception",
				"reason": "request failed",
			},
			"status": m.failAliasCode,
		})
		return
	}
	members := m.membership[aliasName]
	if len(members) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": map[string]any{
				"type":   "alias_not_found_exception",
				"reason": "alias [" + aliasName + "] missing",
			},
			"status": http.StatusNotFound,
		})
		return
	}
	body := map[string]any{}
	for target, member := range members {
		aliasBody := map[string]any{}
		if member.write != nil {
			aliasBody["is_write_index"] = *member.write
		}
		body[target] = map[string]any{"aliases": map[string]any{aliasName: aliasBody}}
	}
	writeJSON(w, http.StatusOK, body)
}

func (m *clusterMock) resolve(w http.ResponseWriter, target string) {
	if _, ok := m.dataStreams[target]; ok {
		backing := ".ds-" + target + "-000001"
		writeJSON(w, http.StatusOK, map[string]any{
			"indices": []any{map[string]any{
				"name":        backing,
				"attributes":  []string{"open"},
				"data_stream": target,
			}},
			"aliases": []any{},
			"data_streams": []any{map[string]any{
				"name":            target,
				"timestamp_field": "@timestamp",
				"backing_indices": []string{backing},
			}},
		})
		return
	}
	if _, ok := m.indices[target]; ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"indices": []any{map[string]any{
				"name":       target,
				"attributes": []string{"open"},
			}},
			"aliases":      []any{},
			"data_streams": []any{},
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"indices":      []any{},
		"aliases":      []any{},
		"data_streams": []any{},
	})
}

func (m *clusterMock) createIndex(w http.ResponseWriter, r *http.Request, name string) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	body := map[string]json.RawMessage{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
	}
	m.putBodies = append(m.putBodies, body)
	if m.nameTaken(name) {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"type":   "resource_already_exists_exception",
				"reason": "index [" + name + "] already exists",
			},
			"status": http.StatusBadRequest,
		})
		return
	}
	m.indices[name] = struct{}{}
	var aliases map[string]struct {
		IsWriteIndex *bool `json:"is_write_index"`
	}
	if rawAliases, ok := body["aliases"]; ok {
		if err := json.Unmarshal(rawAliases, &aliases); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
	}
	for aliasName, alias := range aliases {
		m.putMember(aliasName, name, aliasTarget{write: alias.IsWriteIndex})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"acknowledged":        true,
		"shards_acknowledged": true,
		"index":               name,
	})
}

func (m *clusterMock) nameTaken(name string) bool {
	if _, ok := m.indices[name]; ok {
		return true
	}
	if _, ok := m.aliasObjects[name]; ok {
		return true
	}
	if _, ok := m.dataStreams[name]; ok {
		return true
	}
	return false
}

func (m *clusterMock) deleteIndex(w http.ResponseWriter, name string) {
	if _, ok := m.indices[name]; !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": map[string]any{
				"type":   "index_not_found_exception",
				"reason": "no such index [" + name + "]",
			},
			"status": http.StatusNotFound,
		})
		return
	}
	delete(m.indices, name)
	for aliasName, members := range m.membership {
		delete(members, name)
		if len(members) == 0 {
			delete(m.membership, aliasName)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"acknowledged": true})
}

func (m *clusterMock) requestPaths() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.requests))
	copy(out, m.requests)
	return out
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("X-Elastic-Product", "Elasticsearch")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
