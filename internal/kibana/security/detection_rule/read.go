package detection_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityDetectionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityDetectionRuleData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Implement actual API call to read security detection rule
	// For now, just keep the current state
	
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}