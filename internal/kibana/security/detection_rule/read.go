package detection_rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *securityDetectionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityDetectionRuleData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the composite ID
	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceId := compId.ClusterId
	ruleId := compId.ResourceId

	// Get the rule from the API
	result, diags := GetSecurityDetectionRule(ctx, r.client, spaceId, ruleId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If rule not found, remove from state
	if result == nil {
		tflog.Warn(ctx, "Security detection rule not found, removing from state", map[string]interface{}{
			"rule_id":  ruleId,
			"space_id": spaceId,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the data with the response
	diags = apiResponseToData(ctx, result, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set space_id from the composite ID (keep existing value from config)
	// data.SpaceId remains unchanged

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
