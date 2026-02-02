package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertingRuleModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Expected configured API client")
		return
	}

	ruleID, spaceID := state.getRuleIDAndSpaceID()

	deleteDiags := kibana.DeleteAlertingRule(ctx, r.client, ruleID, spaceID)
	if deleteDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(deleteDiags)...)
		return
	}
}
