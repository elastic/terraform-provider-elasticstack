package datafeed_state

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
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
	if diags.Contains(diagutil.FrameworkDiagFromError(context.DeadlineExceeded)[0]) {
		diags.AddError("Operation timed out", fmt.Sprintf("The operation to create the ML datafeed state timed out after %s. You may need to allocate more free memory within ML nodes by either closing other jobs, or increasing the overall ML memory. You may retry the operation.", createTimeout))
	}

	resp.Diagnostics.Append(diags...)
}
