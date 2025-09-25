package script

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (r *scriptResource) read(ctx context.Context, scriptID string, client *clients.ApiClient) (ScriptData, diag.Diagnostics) {
	var data ScriptData
	var diags diag.Diagnostics

	script, frameworkDiags := elasticsearch.GetScript(ctx, client, scriptID)
	diags.Append(frameworkDiags...)
	if diags.HasError() {
		return data, diags
	}

	if script == nil {
		// Script not found - return empty data with null ScriptId to signal not found
		data.ScriptId = types.StringNull()
		return data, diags
	}

	data.ScriptId = types.StringValue(scriptID)
	data.Lang = types.StringValue(script.Language)
	data.Source = types.StringValue(script.Source)

	// Handle params if returned by the API
	if len(script.Params) > 0 {
		paramsBytes, err := json.Marshal(script.Params)
		if err != nil {
			diags.AddError("Error marshaling script params", err.Error())
			return data, diags
		}
		data.Params = jsontypes.NewNormalizedValue(string(paramsBytes))
	} else {
		data.Params = jsontypes.NewNormalizedNull()
	}
	// Note: If params were set but API doesn't return them, they are preserved from state
	// This maintains backwards compatibility

	// Note: context is not returned by the Elasticsearch API (json:"-" in model)
	// It's only used during script creation, so we preserve it from state
	// This is consistent with the SDKv2 implementation

	return data, diags
}
