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
	"net/url"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// RoleWithRawGlobal carries the typed role fields alongside the raw `global`
// JSON. The `global` field is kept as raw JSON because the go-elasticsearch
// typed client declares Role.Global as map[string]map[string]map[string][]string,
// which cannot decode heterogeneous per-category shapes such as the
// "data_source": [] array introduced in Elasticsearch 9.5.
// Upstream tracking: elasticsearch-specification#6377.
type RoleWithRawGlobal struct {
	Role   *types.Role
	Global json.RawMessage
}

func PutRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, role *types.Role) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	req := typedClient.Security.PutRole(name)

	req.Applications(role.Applications...)
	req.Cluster(role.Cluster...)
	if role.Description != nil {
		req.Description(*role.Description)
	}
	if role.Global != nil {
		globalJSON, err := json.Marshal(role.Global)
		if err != nil {
			return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to marshal global privileges", err.Error())}
		}
		var global map[string]json.RawMessage
		if err := json.Unmarshal(globalJSON, &global); err != nil {
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
// rather than the typed Security.GetRole client. This bypasses the typed
// client's Role.Global decode, which fails on heterogeneous per-category
// shapes such as the "data_source": [] array introduced in Elasticsearch 9.5
// (upstream: elasticsearch-specification#6377). The raw `global` JSON is
// extracted before the remaining fields are decoded into the typed types.Role,
// and is carried back to the caller out-of-band via RoleWithRawGlobal.
func GetRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rolename string) (*RoleWithRawGlobal, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	path := fmt.Sprintf("/_security/role/%s", url.PathEscape(rolename))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
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
			"Unable to get a role",
			fmt.Sprintf("status: %d, body: %s", httpRes.StatusCode, string(body)),
		)}
	}

	var roles map[string]json.RawMessage
	if err := json.NewDecoder(httpRes.Body).Decode(&roles); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	rawRole, ok := roles[rolename]
	if !ok {
		return nil, nil
	}

	// Split the per-role entry into its fields so we can extract `global` as
	// raw JSON and decode everything else into the typed struct.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rawRole, &fields); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	rawGlobal := fields["global"]
	delete(fields, "global")

	remainder, err := json.Marshal(fields)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	var role types.Role
	if err := json.Unmarshal(remainder, &role); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return &RoleWithRawGlobal{Role: &role, Global: rawGlobal}, nil
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
