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

package role_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	mockClusterUUID = "mock-cluster-uuid-aaaa-bbbb-cccc-dddddddddddd"
	mockRoleName    = "tf-test-allow-restricted-indices"
)

// TestUnitResourceSecurityRoleAllowRestrictedIndicesOmittedOnAppend reproduces
// GitHub issue #4759: appending a new indices Set element that omits
// allow_restricted_indices during Update. The mock Elasticsearch normalizes an
// omitted field to false, matching Put Role API behavior.
func TestUnitResourceSecurityRoleAllowRestrictedIndicesOmittedOnAppend(t *testing.T) {
	srv := newMockSecurityRoleServer(t)
	t.Setenv("ELASTICSEARCH_ENDPOINTS", srv.URL)

	resourceName := "elasticstack_elasticsearch_security_role.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: mockRoleAllowRestrictedIndicesConfig(srv.URL, `[{
      names      = ["logs-*"]
      privileges = ["read"]
    }]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", mockRoleName),
					resource.TestCheckResourceAttr(resourceName, "indices.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "indices.0.allow_restricted_indices", "false"),
					resource.TestCheckTypeSetElemAttr(resourceName, "indices.*.names.*", "logs-*"),
				),
			},
			{
				Config: mockRoleAllowRestrictedIndicesConfig(srv.URL, `[{
      names      = ["logs-*"]
      privileges = ["read"]
    },
    {
      names      = ["metrics-*"]
      privileges = ["read"]
    }]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", mockRoleName),
					resource.TestCheckResourceAttr(resourceName, "indices.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "indices.0.allow_restricted_indices", "false"),
					resource.TestCheckResourceAttr(resourceName, "indices.1.allow_restricted_indices", "false"),
					resource.TestCheckTypeSetElemAttr(resourceName, "indices.*.names.*", "logs-*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "indices.*.names.*", "metrics-*"),
				),
			},
		},
	})
}

func mockRoleAllowRestrictedIndicesConfig(endpoint, indexPermissions string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {
    endpoints = ["%s"]
    username  = "elastic"
    password  = "changeme"
  }
}

locals {
  index_permissions = %s
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name = %q

  dynamic "indices" {
    for_each = local.index_permissions
    content {
      names      = indices.value.names
      privileges = indices.value.privileges
    }
  }
}
`, endpoint, indexPermissions, mockRoleName)
}

// newMockSecurityRoleServer persists Put Role bodies and echoes omitted
// allow_restricted_indices as false, matching Elasticsearch Put Role
// normalization.
func newMockSecurityRoleServer(t *testing.T) *httptest.Server {
	t.Helper()

	var mu sync.Mutex
	roles := map[string]map[string]any{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"cluster_uuid": mockClusterUUID,
				"version": map[string]any{
					"number":       "8.19.0",
					"build_flavor": "default",
				},
			})
			return
		}

		roleName, ok := strings.CutPrefix(r.URL.Path, "/_security/role/")
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "unexpected request: " + r.Method + " " + r.URL.Path})
			return
		}

		mu.Lock()
		defer mu.Unlock()

		switch r.Method {
		case http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
				return
			}
			var role map[string]any
			if err := json.Unmarshal(body, &role); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
				return
			}
			normalizeOmittedAllowRestrictedIndices(role)
			roles[roleName] = role
			_ = json.NewEncoder(w).Encode(map[string]any{
				"role": map[string]any{"created": true},
			})
		case http.MethodGet:
			role, exists := roles[roleName]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "role not found"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{roleName: role})
		case http.MethodDelete:
			delete(roles, roleName)
			_ = json.NewEncoder(w).Encode(map[string]any{"found": true})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "unexpected method: " + r.Method + " " + r.URL.Path})
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func normalizeOmittedAllowRestrictedIndices(role map[string]any) {
	for _, key := range []string{"indices", "remote_indices"} {
		entries, ok := role[key].([]any)
		if !ok {
			continue
		}
		for _, entry := range entries {
			entryMap, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			if _, set := entryMap["allow_restricted_indices"]; !set {
				entryMap["allow_restricted_indices"] = false
			}
		}
	}
}
