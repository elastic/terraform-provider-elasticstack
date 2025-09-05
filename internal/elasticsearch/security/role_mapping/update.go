package role_mapping

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *roleMappingResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var data RoleMappingData
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &data)...)
	if diags.HasError() {
		return diags
	}

	roleMappingName := data.Name.ValueString()

	client, frameworkDiags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(frameworkDiags...)
	if diags.HasError() {
		return diags
	}

	// Parse rules JSON
	var rules map[string]interface{}
	if err := json.Unmarshal([]byte(data.Rules.ValueString()), &rules); err != nil {
		diags.AddError("Failed to parse rules JSON", err.Error())
		return diags
	}

	// Parse metadata JSON
	metadata := json.RawMessage(data.Metadata.ValueString())

	// Prepare role mapping
	roleMapping := models.RoleMapping{
		Name:     roleMappingName,
		Enabled:  data.Enabled.ValueBool(),
		Rules:    rules,
		Metadata: metadata,
	}

	// Handle roles or role templates
	if utils.IsKnown(data.Roles) {
		var roles []string
		rolesElements := make([]types.String, 0, len(data.Roles.Elements()))
		diags.Append(data.Roles.ElementsAs(ctx, &rolesElements, false)...)
		if diags.HasError() {
			return diags
		}
		for _, role := range rolesElements {
			roles = append(roles, role.ValueString())
		}
		roleMapping.Roles = roles
	}

	if utils.IsKnown(data.RoleTemplates) {
		var roleTemplates []map[string]interface{}
		if err := json.Unmarshal([]byte(data.RoleTemplates.ValueString()), &roleTemplates); err != nil {
			diags.AddError("Failed to parse role templates JSON", err.Error())
			return diags
		}
		roleMapping.RoleTemplates = roleTemplates
	}

	// Put role mapping
	apiDiags := elasticsearch.PutRoleMapping(ctx, client, &roleMapping)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return diags
	}

	// Read the updated role mapping to ensure consistent result
	readData, readDiags := readRoleMapping(ctx, client, roleMappingName, data.ElasticsearchConnection)
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
