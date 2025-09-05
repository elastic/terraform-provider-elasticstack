package role_mapping

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *roleMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RoleMappingData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	roleMappingName := compId.ResourceId

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleMapping, sdkDiags := elasticsearch.GetRoleMapping(ctx, client, roleMappingName)
	resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if roleMapping == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Role mapping "%s" not found, removing from state`, roleMappingName))
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(roleMapping.Name)
	data.Enabled = types.BoolValue(roleMapping.Enabled)

	// Handle rules
	rulesJSON, err := json.Marshal(roleMapping.Rules)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal rules", err.Error())
		return
	}
	data.Rules = types.StringValue(string(rulesJSON))

	// Handle roles
	if len(roleMapping.Roles) > 0 {
		rolesValues := make([]attr.Value, len(roleMapping.Roles))
		for i, role := range roleMapping.Roles {
			rolesValues[i] = types.StringValue(role)
		}
		rolesSet, diags := types.SetValue(types.StringType, rolesValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Roles = rolesSet
	} else {
		data.Roles = types.SetNull(types.StringType)
	}

	// Handle role templates
	if len(roleMapping.RoleTemplates) > 0 {
		roleTemplatesJSON, err := json.Marshal(roleMapping.RoleTemplates)
		if err != nil {
			resp.Diagnostics.AddError("Failed to marshal role templates", err.Error())
			return
		}
		data.RoleTemplates = types.StringValue(string(roleTemplatesJSON))
	} else {
		data.RoleTemplates = types.StringNull()
	}

	// Handle metadata
	metadataJSON, err := json.Marshal(roleMapping.Metadata)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal metadata", err.Error())
		return
	}
	data.Metadata = types.StringValue(string(metadataJSON))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}