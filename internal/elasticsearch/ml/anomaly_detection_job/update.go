package anomaly_detection_job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	hasChanges := false

	if !plan.Description.Equal(state.Description) {
		updateBody.Description = utils.Pointer(plan.Description.ValueString())
		hasChanges = true
	}

	if !plan.Groups.Equal(state.Groups) {
		var groups []string
		plan.Groups.ElementsAs(ctx, &groups, false)
		updateBody.Groups = groups
		hasChanges = true
	}

	if !plan.ModelPlotConfig.Equal(state.ModelPlotConfig) {
		var modelPlotConfig ModelPlotConfigTFModel
		plan.ModelPlotConfig.As(ctx, &modelPlotConfig, basetypes.ObjectAsOptions{})
		apiModelPlotConfig := &ModelPlotConfigAPIModel{
			Enabled:            modelPlotConfig.Enabled.ValueBool(),
			AnnotationsEnabled: utils.Pointer(modelPlotConfig.AnnotationsEnabled.ValueBool()),
			Terms:              modelPlotConfig.Terms.ValueString(),
		}
		updateBody.ModelPlotConfig = apiModelPlotConfig
		hasChanges = true
	}

	if !plan.AnalysisLimits.Equal(state.AnalysisLimits) {
		var analysisLimits AnalysisLimitsTFModel
		plan.AnalysisLimits.As(ctx, &analysisLimits, basetypes.ObjectAsOptions{})
		apiAnalysisLimits := &AnalysisLimitsAPIModel{
			ModelMemoryLimit: analysisLimits.ModelMemoryLimit.ValueString(),
		}
		if !analysisLimits.CategorizationExamplesLimit.IsNull() {
			apiAnalysisLimits.CategorizationExamplesLimit = utils.Pointer(analysisLimits.CategorizationExamplesLimit.ValueInt64())
		}
		updateBody.AnalysisLimits = apiAnalysisLimits
		hasChanges = true
	}

	if !plan.AllowLazyOpen.Equal(state.AllowLazyOpen) {
		updateBody.AllowLazyOpen = utils.Pointer(plan.AllowLazyOpen.ValueBool())
		hasChanges = true
	}

	if !plan.BackgroundPersistInterval.Equal(state.BackgroundPersistInterval) && !plan.BackgroundPersistInterval.IsNull() {
		updateBody.BackgroundPersistInterval = utils.Pointer(plan.BackgroundPersistInterval.ValueString())
		hasChanges = true
	}

	if !plan.CustomSettings.Equal(state.CustomSettings) && !plan.CustomSettings.IsNull() {
		var customSettings map[string]interface{}
		if err := json.Unmarshal([]byte(plan.CustomSettings.ValueString()), &customSettings); err != nil {
			resp.Diagnostics.AddError("Failed to parse custom_settings", err.Error())
			return
		}
		updateBody.CustomSettings = customSettings
		hasChanges = true
	}

	if !plan.DailyModelSnapshotRetentionAfterDays.Equal(state.DailyModelSnapshotRetentionAfterDays) && !plan.DailyModelSnapshotRetentionAfterDays.IsNull() {
		updateBody.DailyModelSnapshotRetentionAfterDays = utils.Pointer(plan.DailyModelSnapshotRetentionAfterDays.ValueInt64())
		hasChanges = true
	}

	if !plan.ModelSnapshotRetentionDays.Equal(state.ModelSnapshotRetentionDays) && !plan.ModelSnapshotRetentionDays.IsNull() {
		updateBody.ModelSnapshotRetentionDays = utils.Pointer(plan.ModelSnapshotRetentionDays.ValueInt64())
		hasChanges = true
	}

	if !plan.RenormalizationWindowDays.Equal(state.RenormalizationWindowDays) && !plan.RenormalizationWindowDays.IsNull() {
		updateBody.RenormalizationWindowDays = utils.Pointer(plan.RenormalizationWindowDays.ValueInt64())
		hasChanges = true
	}

	if !plan.ResultsRetentionDays.Equal(state.ResultsRetentionDays) && !plan.ResultsRetentionDays.IsNull() {
		updateBody.ResultsRetentionDays = utils.Pointer(plan.ResultsRetentionDays.ValueInt64())
		hasChanges = true
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
