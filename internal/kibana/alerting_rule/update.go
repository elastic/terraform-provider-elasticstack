package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planModel alertingRuleModel
	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateModel alertingRuleModel
	diags = req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	ruleID, spaceID := stateModel.getRuleIDAndSpaceID()

	// Check if enabled status changed - handle separately if needed
	enabledChanged := !planModel.Enabled.Equal(stateModel.Enabled)

	// Convert model to API request
	updateReq, diags := planModel.toAPIUpdateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the rule
	_, diags = kibana_oapi.UpdateAlertingRule(ctx, kibanaClient, spaceID, ruleID, updateReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle enable/disable separately if status changed
	if enabledChanged {
		if planModel.Enabled.ValueBool() {
			diags = kibana_oapi.EnableAlertingRule(ctx, kibanaClient, spaceID, ruleID)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			diags = kibana_oapi.DisableAlertingRule(ctx, kibanaClient, spaceID, ruleID)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	// Read back to get complete state
	diags = read(ctx, kibanaClient, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, planModel)
	resp.Diagnostics.Append(diags...)
}
