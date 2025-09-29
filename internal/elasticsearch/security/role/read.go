package role

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

	readData, diags := r.read(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readData == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(req.State.Set(ctx, readData)...)
}

func (r *roleResource) read(ctx context.Context, data RoleData) (*RoleData, diag.Diagnostics) {
	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	diags.Append(diags...)
	if diags.HasError() {
		return nil, diags
	}
	roleId := compId.ResourceId

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(diags...)
	if diags.HasError() {
		return nil, diags
	}

	role, sdkDiags := elasticsearch.GetRole(ctx, client, roleId)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return nil, diags
	}

	if role == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Role "%s" not found`, roleId))
		return nil, diags
	}

	// Convert from API model
	diags.Append(data.fromAPIModel(ctx, role)...)
	if diags.HasError() {
		return nil, diags
	}

	// Set the name to the roleId we extracted to ensure consistency
	data.Name = types.StringValue(roleId)

	return &data, diags
}
