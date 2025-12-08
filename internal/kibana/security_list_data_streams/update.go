package security_list_data_streams

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityListDataStreamsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This resource doesn't support updates - all attributes require replacement
	// This function should never be called, but we implement it to satisfy the interface
	var plan SecurityListDataStreamsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
