package role

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	MinSupportedRemoteIndicesVersion = version.Must(version.NewVersion("8.10.0"))
	MinSupportedDescriptionVersion   = version.Must(version.NewVersion("8.15.0"))
)

func (r *roleResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var data RoleData
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &data)...)
	if diags.HasError() {
		return diags
	}

	roleId := data.Name.ValueString()
	id, sdkDiags := r.client.ID(ctx, roleId)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	client, clientDiags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(clientDiags...)
	if diags.HasError() {
		return diags
	}

	serverVersion, sdkDiags := client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	// Check version requirements
	if utils.IsKnown(data.Description) {
		if serverVersion.LessThan(MinSupportedDescriptionVersion) {
			diags.AddError("Unsupported Feature", fmt.Sprintf("'description' is supported only for Elasticsearch v%s and above", MinSupportedDescriptionVersion.String()))
			return diags
		}
	}

	if utils.IsKnown(data.RemoteIndices) {
		var remoteIndicesList []RemoteIndexPermsData
		diags.Append(data.RemoteIndices.ElementsAs(ctx, &remoteIndicesList, false)...)
		if len(remoteIndicesList) > 0 && serverVersion.LessThan(MinSupportedRemoteIndicesVersion) {
			diags.AddError("Unsupported Feature", fmt.Sprintf("'remote_indices' is supported only for Elasticsearch v%s and above", MinSupportedRemoteIndicesVersion.String()))
			return diags
		}
	}

	// Convert to API model
	role, modelDiags := data.toAPIModel(ctx)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return diags
	}

	// Put the role
	sdkDiags = elasticsearch.PutRole(ctx, client, role)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	data.Id = types.StringValue(id.String())
	readData, readDiags := r.read(ctx, data)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	if readData == nil {
		diags.AddError("Not Found", fmt.Sprintf("Role %q was not found after update", roleId))
		return diags
	}

	diags.Append(state.Set(ctx, readData)...)
	return diags
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := r.update(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
}
