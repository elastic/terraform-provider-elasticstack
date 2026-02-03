package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// readRuleFromAPI reads the alerting rule from the API and populates the model.
// Returns (exists, diagnostics).
func (r *Resource) readRuleFromAPI(ctx context.Context, model *alertingRuleModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	ruleID, spaceID := model.getRuleIDAndSpaceID()

	rule, readDiags := kibana.GetAlertingRule(ctx, r.client, ruleID, spaceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(readDiags)...)
	if diags.HasError() {
		return false, diags
	}

	if rule == nil {
		return false, diags
	}

	diags.Append(model.populateFromAPI(ctx, rule)...)
	return true, diags
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertingRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Expected configured API client")
		return
	}

	exists, diags := r.readRuleFromAPI(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
