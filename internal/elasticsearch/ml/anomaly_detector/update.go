package anomaly_detector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, client *clients.ApiClient) {
	if !resourceReady(client, &resp.Diagnostics) {
		return
	}

	var plan AnomalyDetectorJobTFModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AnomalyDetectorJobTFModel
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
	// results_retention_days, custom_settings, and background_persist_interval
	// can be updated.

	// Create update body with only updatable fields
	updateBody := map[string]interface{}{}

	if !plan.Description.Equal(state.Description) {
		updateBody["description"] = plan.Description.ValueString()
	}

	if !plan.Groups.Equal(state.Groups) {
		var groups []string
		plan.Groups.ElementsAs(ctx, &groups, false)
		updateBody["groups"] = groups
	}

	if !plan.ModelPlotConfig.Equal(state.ModelPlotConfig) {
		var modelPlotConfig ModelPlotConfigTFModel
		plan.ModelPlotConfig.As(ctx, &modelPlotConfig, basetypes.ObjectAsOptions{})
		apiModelPlotConfig := &ModelPlotConfigAPIModel{
			Enabled:            modelPlotConfig.Enabled.ValueBool(),
			AnnotationsEnabled: utils.Pointer(modelPlotConfig.AnnotationsEnabled.ValueBool()),
			Terms:              modelPlotConfig.Terms.ValueString(),
		}
		updateBody["model_plot_config"] = apiModelPlotConfig
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
		updateBody["analysis_limits"] = apiAnalysisLimits
	}

	if !plan.RenormalizationWindowDays.Equal(state.RenormalizationWindowDays) && !plan.RenormalizationWindowDays.IsNull() {
		updateBody["renormalization_window_days"] = plan.RenormalizationWindowDays.ValueInt64()
	}

	if !plan.ResultsRetentionDays.Equal(state.ResultsRetentionDays) && !plan.ResultsRetentionDays.IsNull() {
		updateBody["results_retention_days"] = plan.ResultsRetentionDays.ValueInt64()
	}

	if !plan.BackgroundPersistInterval.Equal(state.BackgroundPersistInterval) && !plan.BackgroundPersistInterval.IsNull() {
		updateBody["background_persist_interval"] = plan.BackgroundPersistInterval.ValueString()
	}

	if !plan.CustomSettings.Equal(state.CustomSettings) && !plan.CustomSettings.IsNull() {
		var customSettings map[string]interface{}
		if err := json.Unmarshal([]byte(plan.CustomSettings.ValueString()), &customSettings); err != nil {
			resp.Diagnostics.AddError("Failed to parse custom_settings", err.Error())
			return
		}
		updateBody["custom_settings"] = customSettings
	}

	// Only proceed with update if there are changes
	if len(updateBody) == 0 {
		tflog.Debug(ctx, fmt.Sprintf("No updates needed for ML anomaly detection job: %s", jobID))
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		return
	}

	esClient, err := client.GetESClient()
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

	if diags := utils.CheckErrorFromFW(res, fmt.Sprintf("Unable to update ML anomaly detection job: %s", jobID)); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Read the updated job to get the current state
	readReq := resource.ReadRequest{State: req.State}
	readResp := resource.ReadResponse{
		State:       resp.State,
		Diagnostics: resp.Diagnostics,
	}
	read(ctx, readReq, &readResp, client)
	resp.Diagnostics = readResp.Diagnostics
	resp.State = readResp.State

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML anomaly detection job: %s", jobID))
}
