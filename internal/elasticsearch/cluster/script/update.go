package script

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *scriptResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var data ScriptData
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &data)...)
	if diags.HasError() {
		return diags
	}

	scriptId := data.ScriptId.ValueString()
	id, sdkDiags := r.client.ID(ctx, scriptId)
	diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(diags...)
	if diags.HasError() {
		return diags
	}

	script := models.Script{
		ID:       scriptId,
		Language: data.Lang.ValueString(),
		Source:   data.Source.ValueString(),
	}

	if utils.IsKnown(data.Params) && !data.Params.IsNull() {
		var params map[string]interface{}
		err := json.Unmarshal([]byte(data.Params.ValueString()), &params)
		if err != nil {
			diags.AddError("Error unmarshaling script params", err.Error())
			return diags
		}
		script.Params = params
	}

	if utils.IsKnown(data.Context) && !data.Context.IsNull() {
		script.Context = data.Context.ValueString()
	}

	sdkDiags = elasticsearch.PutScript(ctx, client, &script)
	diags.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	data.Id = types.StringValue(id.String())
	diags.Append(state.Set(ctx, &data)...)
	return diags
}

func (r *scriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := r.update(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}