package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel alertingRuleModel
	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	diags = read(ctx, kibanaClient, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the rule was not found, mark it as removed
	if stateModel.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}

func read(ctx context.Context, client *kibana_oapi.Client, model *alertingRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	ruleID, spaceID := model.getRuleIDAndSpaceID()

	// Get the alerting rule
	getResp, getDiags := kibana_oapi.GetAlertingRule(ctx, client, spaceID, ruleID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return diags
	}

	// If rule not found, mark for removal
	if getResp == nil {
		model.ID = types.StringNull()
		return diags
	}

	// Populate model from API response
	diags.Append(model.populateFromAPI(ctx, getResp)...)

	return diags
}
