package systemuser

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

func (r *systemUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data Data
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	usernameID := compID.ResourceID

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, sdkDiags := elasticsearch.GetUser(ctx, client, usernameID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if user == nil || !user.IsSystemUser() {
		tflog.Warn(ctx, fmt.Sprintf(`System user "%s" not found, removing from state`, compID.ResourceID))
		resp.State.RemoveResource(ctx)
		return
	}

	data.Username = types.StringValue(usernameID)
	data.Enabled = types.BoolValue(user.Enabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
