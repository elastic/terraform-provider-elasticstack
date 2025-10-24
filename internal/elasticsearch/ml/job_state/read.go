package job_state

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *mlJobStateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MLJobStateData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get job stats to check current state
	jobId := compId.ResourceId
	currentState, diags := r.getJobState(ctx, jobId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if currentState == nil {
		tflog.Warn(ctx, fmt.Sprintf(`ML job "%s" not found, removing from state`, jobId))
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state with current job information
	data.JobId = types.StringValue(jobId)
	data.State = types.StringValue(*currentState)

	// Set defaults for computed attributes if they're not already set (e.g., during import)
	if data.Force.IsNull() {
		data.Force = types.BoolValue(false)
	}
	if data.Timeout.IsNull() {
		data.Timeout = customtypes.NewDurationValue("30s")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
