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
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func PutRoleMapping(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, roleMapping *types.SecurityRoleMapping) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	req := typedClient.Security.PutRoleMapping(name).
		Enabled(roleMapping.Enabled).
		Rules(&roleMapping.Rules)

	if len(roleMapping.Roles) > 0 {
		req.Roles(roleMapping.Roles...)
	}
	if len(roleMapping.RoleTemplates) > 0 {
		req.RoleTemplates(roleMapping.RoleTemplates...)
	}
	if roleMapping.Metadata != nil {
		req.Metadata(roleMapping.Metadata)
	}

	// The typed client's Script type marshals template strings as objects
	// {"source":"..."}. The ES role mapping API accepts and returns
	// templates as plain strings by default, and sending an object may
	// cause ES to wrap the value. Override the request body to preserve
	// the template as a plain string, matching the old provider behaviour.
	if len(roleMapping.RoleTemplates) > 0 {
		body := map[string]any{
			"enabled": roleMapping.Enabled,
			"rules":   &roleMapping.Rules,
		}
		if len(roleMapping.Roles) > 0 {
			body["roles"] = roleMapping.Roles
		}
		if roleMapping.Metadata != nil {
			body["metadata"] = roleMapping.Metadata
		}
		templates := make([]map[string]any, len(roleMapping.RoleTemplates))
		for i, rt := range roleMapping.RoleTemplates {
			t := map[string]any{}
			if rt.Format != nil {
				t["format"] = rt.Format.String()
			}
			if rt.Template.Source != nil {
				t["template"] = *rt.Template.Source
			}
			templates[i] = t
		}
		body["role_templates"] = templates
		data, err := json.Marshal(body)
		if err != nil {
			diags.AddError("Unable to marshal role mapping request", err.Error())
			return diags
		}
		req.Raw(bytes.NewReader(data))
	}

	_, err := req.Do(ctx)
	if err != nil {
		diags.AddError("Unable to create or update a role mapping", err.Error())
		return diags
	}

	return diags
}

func GetRoleMapping(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, roleMappingName string) (*types.SecurityRoleMapping, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	res, err := typedClient.Security.GetRoleMapping().Name(roleMappingName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, diags
		}
		diags.AddError("Unable to get role mapping", err.Error())
		return nil, diags
	}

	if roleMapping, ok := res[roleMappingName]; ok {
		return &roleMapping, diags
	}

	diags.AddError("Role mapping not found", fmt.Sprintf("unable to find role mapping '%s' in the cluster", roleMappingName))
	return nil, diags
}

func DeleteRoleMapping(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, roleMappingName string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	_, err := typedClient.Security.DeleteRoleMapping(roleMappingName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Unable to delete role mapping", err.Error())
		return diags
	}

	return diags
}
