package role

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RoleData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	roleId := compId.ResourceId

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, sdkDiags := elasticsearch.GetRole(ctx, client, roleId)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if role == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Role "%s" not found, removing from state`, roleId))
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert from API model
	resp.Diagnostics.Append(data.fromAPIModel(ctx, role)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the name to the roleId we extracted to ensure consistency
	data.Name = types.StringValue(roleId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
