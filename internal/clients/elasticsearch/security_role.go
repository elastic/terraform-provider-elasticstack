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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/clusterprivilege"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// Role is the provider's own representation of an Elasticsearch role
// document, used instead of the go-elasticsearch typed client's types.Role.
// It mirrors types.Role's JSON shape field-for-field, except `Global` is
// declared as raw JSON instead of the typed client's
// map[string]map[string]map[string][]string. types.Role has a hand-written
// UnmarshalJSON that decodes `global` eagerly into that fixed shape and fails
// on heterogeneous per-category values such as the "data_source": [] array
// introduced in Elasticsearch 9.5 (upstream: elasticsearch-specification#6377).
// Role has no custom UnmarshalJSON/MarshalJSON of its own, so the ordinary
// reflection-based encoder/decoder applies to `global` regardless of shape.
//
// Keep this in sync with types.Role if a go-elasticsearch upgrade adds,
// removes, or renames fields.
type Role struct {
	Applications      []types.ApplicationPrivileges       `json:"applications"`
	Cluster           []clusterprivilege.ClusterPrivilege `json:"cluster"`
	Description       *string                             `json:"description,omitempty"`
	Global            json.RawMessage                     `json:"global,omitempty"`
	Indices           []types.IndicesPrivileges           `json:"indices"`
	Metadata          types.Metadata                      `json:"metadata"`
	RemoteCluster     []types.RemoteClusterPrivileges     `json:"remote_cluster,omitempty"`
	RemoteIndices     []types.RemoteIndicesPrivileges     `json:"remote_indices,omitempty"`
	RoleTemplates     []types.RoleTemplate                `json:"role_templates,omitempty"`
	RunAs             []string                            `json:"run_as,omitempty"`
	TransientMetadata map[string]json.RawMessage          `json:"transient_metadata,omitempty"`
}

func PutRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, role *Role) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	req := typedClient.Security.PutRole(name)

	req.Applications(role.Applications...)
	req.Cluster(role.Cluster...)
	if role.Description != nil {
		req.Description(*role.Description)
	}
	if role.Global != nil {
		var global map[string]json.RawMessage
		if err := json.Unmarshal(role.Global, &global); err != nil {
			return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to convert global privileges", err.Error())}
		}
		req.Global(global)
	}
	req.Indices(role.Indices...)
	req.Metadata(role.Metadata)
	req.RemoteIndices(role.RemoteIndices...)
	req.RunAs(role.RunAs...)

	_, err := req.Do(ctx)
	if err != nil {
		return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to create or update a role", err.Error())}
	}

	return nil
}

// GetRole fetches a role via a raw GET /_security/role/<name> transport call
// rather than the typed Security.GetRole client, decoding the response
// directly into Role. This bypasses the typed client's Role.Global decode,
// which fails on heterogeneous per-category shapes such as the
// "data_source": [] array introduced in Elasticsearch 9.5
// (upstream: elasticsearch-specification#6377).
func GetRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rolename string) (*Role, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	req, err := typedClient.Security.GetRole().Name(rolename).HttpRequest(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	httpRes, err := typedClient.Transport.Perform(req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if httpRes.StatusCode >= 400 {
		body, _ := io.ReadAll(httpRes.Body)
		return nil, fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(
			fmt.Sprintf("Unable to get role %q: unexpected status code from server: got HTTP %d", rolename, httpRes.StatusCode),
			string(body),
		)}
	}

	var roles map[string]json.RawMessage
	if err := json.NewDecoder(httpRes.Body).Decode(&roles); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	rawRole, ok := roles[rolename]
	if !ok {
		return nil, fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(
			"Unable to find a role in the cluster",
			fmt.Sprintf(`Unable to find "%s" role in the cluster`, rolename),
		)}
	}

	var role Role
	if err := json.Unmarshal(rawRole, &role); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	// Treat an explicit JSON null the same as an absent global so state is
	// stored as null rather than the literal string "null".
	if strings.TrimSpace(string(role.Global)) == "null" {
		role.Global = nil
	}

	return &role, nil
}

func DeleteRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rolename string) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Security.DeleteRole(rolename).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to delete a role", err.Error())}
	}

	return nil
}
