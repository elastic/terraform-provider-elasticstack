package detection_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityDetectionRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityDetectionRuleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Implement actual API call to update security detection rule
	// For now, just set the updated state
	
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}