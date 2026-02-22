package script

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *scriptResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var data Data
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &data)...)
	if diags.HasError() {
		return diags
	}

	scriptID := data.ScriptID.ValueString()
	id, sdkDiags := r.client.ID(ctx, scriptID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(diags...)
	if diags.HasError() {
		return diags
	}

	script := models.Script{
		ID:       scriptID,
		Language: data.Lang.ValueString(),
		Source:   data.Source.ValueString(),
	}

	if typeutils.IsKnown(data.Params) {
		paramsStr := data.Params.ValueString()
		if paramsStr != "" {
			var params map[string]any
			err := json.Unmarshal([]byte(paramsStr), &params)
			if err != nil {
				diags.AddError("Error unmarshaling script params", err.Error())
				return diags
			}
			script.Params = params
		}
	}

	if typeutils.IsKnown(data.Context) {
		script.Context = data.Context.ValueString()
	}

	diags.Append(elasticsearch.PutScript(ctx, client, &script)...)
	if diags.HasError() {
		return diags
	}

	// Read the script back from Elasticsearch to populate state
	readData, readDiags := r.read(ctx, scriptID, client)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	// Preserve connection and ID from the original data
	readData.ElasticsearchConnection = data.ElasticsearchConnection
	readData.ID = types.StringValue(id.String())

	// Preserve context from the original data as it's not returned by the API
	readData.Context = data.Context

	// Preserve params from original data if API didn't return them
	if readData.Params.IsNull() && !data.Params.IsNull() {
		readData.Params = data.Params
	}

	diags.Append(state.Set(ctx, &readData)...)
	return diags
}

func (r *scriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := r.update(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
}
