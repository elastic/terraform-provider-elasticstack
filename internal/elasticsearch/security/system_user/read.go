package system_user

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *systemUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SystemUserData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Keep this one
	if resp.Diagnostics.HasError() {
		return
	}

	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	usernameId := compId.ResourceId

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, sdkDiags := elasticsearch.GetUser(ctx, client, usernameId)
	resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...) // Keep this one
	if resp.Diagnostics.HasError() {
		return
	}

	if user == nil || !user.IsSystemUser() {
		tflog.Warn(ctx, fmt.Sprintf(`System user "%s" not found, removing from state`, compId.ResourceId))
		resp.State.RemoveResource(ctx)
		return
	}

	data.Username = types.StringValue(usernameId)
	data.Enabled = types.BoolValue(user.Enabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Keep this one
}
