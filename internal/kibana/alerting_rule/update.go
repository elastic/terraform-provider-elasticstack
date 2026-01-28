package alerting_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertingRuleModel
	var state alertingRuleModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	compositeID, diags := clients.CompositeIdFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to API request
	apiReq, diags := plan.toAPIUpdateRequest(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the rule
	ruleResp, diags := kibana_oapi.UpdateAlertingRule(ctx, client, compositeID.ResourceId, apiReq)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle enabled state changes
	currentEnabled := ruleResp.Enabled
	desiredEnabled := plan.Enabled.ValueBool()

	if desiredEnabled && !currentEnabled {
		diags := kibana_oapi.EnableAlertingRule(ctx, client, compositeID.ResourceId)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else if !desiredEnabled && currentEnabled {
		diags := kibana_oapi.DisableAlertingRule(ctx, client, compositeID.ResourceId)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Re-read to get the final state
	rule, diags := r.readRule(ctx, compositeID.ClusterId, compositeID.ResourceId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, rule)
	resp.Diagnostics.Append(diags...)
}
