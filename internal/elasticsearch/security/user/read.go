package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
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
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if user == nil {
		tflog.Warn(ctx, fmt.Sprintf(`User "%s" not found, removing from state`, compId.ResourceId))
		resp.State.RemoveResource(ctx)
		return
	}

	// Set the fields
	data.Username = types.StringValue(usernameId)
	data.Email = types.StringValue(user.Email)
	data.FullName = types.StringValue(user.FullName)
	data.Enabled = types.BoolValue(user.Enabled)

	// Handle metadata
	if user.Metadata != nil && len(user.Metadata) > 0 {
		metadata, err := json.Marshal(user.Metadata)
		if err != nil {
			resp.Diagnostics.AddError("Failed to marshal metadata", err.Error())
			return
		}
		data.Metadata = jsontypes.NewNormalizedValue(string(metadata))
	} else {
		data.Metadata = jsontypes.NewNormalizedNull()
	}

	// Convert roles slice to set
	rolesSet, diags := types.SetValueFrom(ctx, types.StringType, user.Roles)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Roles = rolesSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
