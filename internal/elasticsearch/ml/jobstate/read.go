package jobstate

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

	compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get job stats to check current state
	jobID := compID.ResourceID
	currentState, diags := r.getJobState(ctx, jobID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if currentState == nil {
		tflog.Warn(ctx, fmt.Sprintf(`ML job "%s" not found, removing from state`, jobID))
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state with current job information
	data.JobID = types.StringValue(jobID)
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
