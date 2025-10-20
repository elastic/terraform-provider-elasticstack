package role_mapping

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// readRoleMapping reads role mapping data from Elasticsearch and returns RoleMappingData
func readRoleMapping(ctx context.Context, client *clients.ApiClient, roleMappingName string, elasticsearchConnection types.List) (*RoleMappingData, diag.Diagnostics) {
	var diags diag.Diagnostics

	roleMapping, apiDiags := elasticsearch.GetRoleMapping(ctx, client, roleMappingName)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return nil, diags
	}

	if roleMapping == nil {
		return nil, diags
	}

	data := &RoleMappingData{}

	// Set basic fields
	compId, compDiags := client.ID(ctx, roleMappingName)
	diags.Append(diagutil.FrameworkDiagsFromSDK(compDiags)...)
	if diags.HasError() {
		return nil, diags
	}
	data.Id = types.StringValue(compId.String())
	data.ElasticsearchConnection = elasticsearchConnection
	data.Name = types.StringValue(roleMapping.Name)
	data.Enabled = types.BoolValue(roleMapping.Enabled)

	// Handle rules
	rulesJSON, err := json.Marshal(roleMapping.Rules)
	if err != nil {
		diags.AddError("Failed to marshal rules", err.Error())
		return nil, diags
	}
	data.Rules = jsontypes.NewNormalizedValue(string(rulesJSON))

	// Handle roles
	data.Roles = utils.SetValueFrom(ctx, roleMapping.Roles, types.StringType, path.Root("roles"), &diags)
	if diags.HasError() {
		return nil, diags
	}

	// Handle role templates
	if len(roleMapping.RoleTemplates) > 0 {
		roleTemplatesJSON, err := json.Marshal(roleMapping.RoleTemplates)
		if err != nil {
			diags.AddError("Failed to marshal role templates", err.Error())
			return nil, diags
		}
		data.RoleTemplates = jsontypes.NewNormalizedValue(string(roleTemplatesJSON))
	} else {
		data.RoleTemplates = jsontypes.NewNormalizedNull()
	}

	// Handle metadata
	metadataJSON, err := json.Marshal(roleMapping.Metadata)
	if err != nil {
		diags.AddError("Failed to marshal metadata", err.Error())
		return nil, diags
	}
	data.Metadata = jsontypes.NewNormalizedValue(string(metadataJSON))

	return data, diags
}

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

	readData, diags := readRoleMapping(ctx, client, roleMappingName, data.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readData == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Role mapping "%s" not found, removing from state`, roleMappingName))
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}
