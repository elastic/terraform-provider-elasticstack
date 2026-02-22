package datafeed

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *datafeedResource) delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var state Datafeed
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	datafeedID := state.DatafeedID.ValueString()
	if datafeedID == "" {
		resp.Diagnostics.AddError("Invalid Configuration", "datafeed_id cannot be empty")
		return
	}

	// Before deleting, we need to stop the datafeed if it's running
	_, stopDiags := r.maybeStopDatafeed(ctx, datafeedID)
	resp.Diagnostics.Append(stopDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the datafeed
	deleteDiags := elasticsearch.DeleteDatafeed(ctx, r.client, datafeedID, false)
	resp.Diagnostics.Append(deleteDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The resource is automatically removed from state on successful delete
}

func (r *datafeedResource) maybeStopDatafeed(ctx context.Context, datafeedID string) (bool, diag.Diagnostics) {
	// Check current state
	currentState, diags := GetDatafeedState(ctx, r.client, datafeedID)
	if diags.HasError() {
		return false, diags
	}

	if currentState == nil {
		return false, nil
	}

	// If the datafeed is not running, nothing to stop
	if *currentState != StateStarted && *currentState != StateStarting {
		return false, diags
	}

	// Stop the datafeed
	stopDiags := elasticsearch.StopDatafeed(ctx, r.client, datafeedID, false, 0)
	diags.Append(stopDiags...)
	if diags.HasError() {
		return true, diags
	}

	// Wait for the datafeed to reach stopped state
	_, waitDiags := WaitForDatafeedState(ctx, r.client, datafeedID, StateStopped)
	diags.Append(waitDiags...)
	if diags.HasError() {
		return true, diags
	}

	return true, diags
}
