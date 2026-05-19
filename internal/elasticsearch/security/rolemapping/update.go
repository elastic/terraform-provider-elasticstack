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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// writeRoleMapping handles both Create and Update; the role mapping PUT API
// is idempotent so the same callback serves both lifecycle methods.
func writeRoleMapping(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[Data]) (entitycore.WriteResult[Data], diag.Diagnostics) {
	var diags diag.Diagnostics
	data := req.Plan
	roleMappingName := req.WriteID

	// Parse rules JSON
	var rules types.RoleMappingRule
	if err := json.Unmarshal([]byte(data.Rules.ValueString()), &rules); err != nil {
		diags.AddError("Failed to parse rules JSON", err.Error())
		return entitycore.WriteResult[Data]{}, diags
	}

	// Parse metadata JSON
	var metadata types.Metadata
	if err := json.Unmarshal([]byte(data.Metadata.ValueString()), &metadata); err != nil {
		diags.AddError("Failed to parse metadata JSON", err.Error())
		return entitycore.WriteResult[Data]{}, diags
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
			return entitycore.WriteResult[Data]{}, diags
		}
	}

	if typeutils.IsKnown(data.RoleTemplates) {
		var roleTemplates []types.RoleTemplate
		if err := json.Unmarshal([]byte(data.RoleTemplates.ValueString()), &roleTemplates); err != nil {
			diags.AddError("Failed to parse role templates JSON", err.Error())
			return entitycore.WriteResult[Data]{}, diags
		}
		roleMapping.RoleTemplates = roleTemplates
	}

	// Put role mapping
	apiDiags := elasticsearch.PutRoleMapping(ctx, client, roleMappingName, &roleMapping)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{}, diags
	}

	return entitycore.WriteResult[Data]{Model: data}, diags
}
