package datafeed_state

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *mlDatafeedStateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MLDatafeedStateData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get create timeout
	createTimeout, fwDiags := data.Timeouts.Create(ctx, 5*time.Minute) // Default 5 minutes
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.update(ctx, req.Plan, &resp.State, createTimeout)
	resp.Diagnostics.Append(diags...)
}
