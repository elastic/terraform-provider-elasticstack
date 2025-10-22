package anomaly_detection_job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *anomalyDetectionJobResource) update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var plan AnomalyDetectionJobTFModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AnomalyDetectionJobTFModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := state.JobID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Updating ML anomaly detection job: %s", jobID))

	// Note: Many ML job properties cannot be updated after creation.
	// Only certain properties like description, groups, model_plot_config,
	// analysis_limits.model_memory_limit, renormalization_window_days,
	// results_retention_days, custom_settings, background_persist_interval,
	// allow_lazy_open, daily_model_snapshot_retention_after_days,
	// and model_snapshot_retention_days can be updated.

	// Create update body with only updatable fields
	updateBody := &AnomalyDetectionJobUpdateAPIModel{}
	hasChanges, diags := updateBody.BuildFromPlan(ctx, &plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only proceed with update if there are changes
	if !hasChanges {
		tflog.Debug(ctx, fmt.Sprintf("No updates needed for ML anomaly detection job: %s", jobID))
		diags.AddWarning("No changed detected to updateble fields during an update operation", `
Changes to non-updateable fields should force a recreation of the anomaly detection job. 
Please report this warning to the provider developers.`)
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		return
	}

	esClient, err := r.client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	// Marshal the update body to JSON
	body, err := json.Marshal(updateBody)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal job update", err.Error())
		return
	}

	// Update the ML job
	res, err := esClient.ML.UpdateJob(jobID, bytes.NewReader(body), esClient.ML.UpdateJob.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update ML anomaly detection job", err.Error())
		return
	}
	defer res.Body.Close()

	diags = diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to update ML anomaly detection job: %s", jobID))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the updated job to get the current state
	found, diags := r.read(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Set the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML anomaly detection job: %s", jobID))
}
