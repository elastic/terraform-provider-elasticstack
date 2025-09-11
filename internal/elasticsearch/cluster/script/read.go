package script

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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

	script, sdkDiags := elasticsearch.GetScript(ctx, client, scriptId)
	resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if script == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Script "%s" not found, removing from state`, compId.ResourceId))
		resp.State.RemoveResource(ctx)
		return
	}

	data.ScriptId = types.StringValue(scriptId)
	data.Lang = types.StringValue(script.Language)
	data.Source = types.StringValue(script.Source)

	// Handle params if returned by the API
	if len(script.Params) > 0 {
		paramsBytes, err := json.Marshal(script.Params)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling script params", err.Error())
			return
		}
		data.Params = types.StringValue(string(paramsBytes))
	}
	// Note: If params were set but API doesn't return them, they are preserved from state
	// This maintains backwards compatibility

	// Note: context is not returned by the Elasticsearch API (json:"-" in model)
	// It's only used during script creation, so we preserve it from state
	// This is consistent with the SDKv2 implementation

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
