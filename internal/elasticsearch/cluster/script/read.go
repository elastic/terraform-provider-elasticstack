package script

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *scriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ScriptData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	scriptId := compId.ResourceId

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the helper read function
	readData, readDiags := r.read(ctx, scriptId, client)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if script was found
	if readData.ScriptId.IsNull() {
		tflog.Warn(ctx, fmt.Sprintf(`Script "%s" not found, removing from state`, compId.ResourceId))
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve connection and ID from original state
	readData.ElasticsearchConnection = data.ElasticsearchConnection
	readData.Id = data.Id

	// Preserve context from state as it's not returned by the API
	readData.Context = data.Context

	// Preserve params from state if API didn't return them
	if readData.Params.IsNull() && !data.Params.IsNull() {
		readData.Params = data.Params
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &readData)...)
}
