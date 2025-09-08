package detection_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityDetectionRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityDetectionRuleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Implement actual API call to create security detection rule
	// For now, just set a placeholder ID
	data.Id = data.RuleId
	if data.Id.IsNull() || data.Id.IsUnknown() {
		// Generate a UUID if rule_id is not provided
		data.Id = data.Name // Placeholder - should generate proper UUID
		data.RuleId = data.Id
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}