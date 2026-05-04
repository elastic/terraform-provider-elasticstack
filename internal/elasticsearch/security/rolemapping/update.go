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

package rolemapping

import (
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (r *roleMappingResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var data Data
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &data)...)
	if diags.HasError() {
		return diags
	}

	roleMappingName := data.Name.ValueString()

	client, frameworkDiags := r.Client().GetElasticsearchClient(ctx, data.ElasticsearchConnection)
	diags.Append(frameworkDiags...)
	if diags.HasError() {
		return diags
	}

	// Parse rules JSON
	var rules types.RoleMappingRule
	if err := json.Unmarshal([]byte(data.Rules.ValueString()), &rules); err != nil {
		diags.AddError("Failed to parse rules JSON", err.Error())
		return diags
	}

	// Parse metadata JSON
	var metadata types.Metadata
	if err := json.Unmarshal([]byte(data.Metadata.ValueString()), &metadata); err != nil {
		diags.AddError("Failed to parse metadata JSON", err.Error())
		return diags
	}

	// Prepare role mapping
	roleMapping := types.SecurityRoleMapping{
		Enabled:  data.Enabled.ValueBool(),
		Rules:    rules,
		Metadata: metadata,
	}

	// Handle roles or role templates
	if typeutils.IsKnown(data.Roles) {
		roleMapping.Roles = typeutils.SetTypeAs[string](ctx, data.Roles, path.Root("roles"), &diags)
		if diags.HasError() {
			return diags
		}
	}

	if typeutils.IsKnown(data.RoleTemplates) {
		var roleTemplates []types.RoleTemplate
		if err := json.Unmarshal([]byte(data.RoleTemplates.ValueString()), &roleTemplates); err != nil {
			diags.AddError("Failed to parse role templates JSON", err.Error())
			return diags
		}
		roleMapping.RoleTemplates = roleTemplates
	}

	// Put role mapping
	apiDiags := elasticsearch.PutRoleMapping(ctx, client, roleMappingName, &roleMapping)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return diags
	}

	// Read the updated role mapping to ensure consistent result
	readData, readDiags := readRoleMapping(ctx, data, roleMappingName, client)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	if readData != nil {
		diags.Append(state.Set(ctx, readData)...)
	}

	return diags
}

func (r *roleMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := r.update(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
}
