package anomaly_detector

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r *anomalyDetectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan anomalyDetectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	esClient, err := r.client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Elasticsearch client", err.Error())
		return
	}

	jobId := plan.JobId.ValueString()

	// Construct the API request from the plan data
	var apiReq apiCreateAnomalyDetectorRequest // Reusing the Create request struct

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		descVal := plan.Description.ValueString()
		apiReq.Description = &descVal
	}

	if !plan.ResultsIndexName.IsNull() && !plan.ResultsIndexName.IsUnknown() {
		rinVal := plan.ResultsIndexName.ValueString()
		apiReq.ResultsIndexName = &rinVal
	}

	if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
		resp.Diagnostics.Append(plan.Groups.ElementsAs(ctx, &apiReq.Groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !plan.AnalysisConfig.IsNull() && !plan.AnalysisConfig.IsUnknown() {
		var analysisConfigData analysisConfigModel
		resp.Diagnostics.Append(plan.AnalysisConfig.As(ctx, &analysisConfigData, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiReq.AnalysisConfig = &apiAnalysisConfig{
			BucketSpan: analysisConfigData.BucketSpan.ValueString(),
		}
		if !analysisConfigData.Influencers.IsNull() && !analysisConfigData.Influencers.IsUnknown() {
			resp.Diagnostics.Append(analysisConfigData.Influencers.ElementsAs(ctx, &apiReq.AnalysisConfig.Influencers, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		var detectorsData []detectorModel
		resp.Diagnostics.Append(analysisConfigData.Detectors.ElementsAs(ctx, &detectorsData, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiReq.AnalysisConfig.Detectors = make([]apiDetector, len(detectorsData))
		for i, detData := range detectorsData {
			apiDet := apiDetector{
				Function: detData.Function.ValueString(),
			}
			if !detData.FieldName.IsNull() && !detData.FieldName.IsUnknown() {
				fnVal := detData.FieldName.ValueString()
				apiDet.FieldName = &fnVal
			}
			if !detData.ByFieldName.IsNull() && !detData.ByFieldName.IsUnknown() {
				bfnVal := detData.ByFieldName.ValueString()
				apiDet.ByFieldName = &bfnVal
			}
			if !detData.PartitionFieldName.IsNull() && !detData.PartitionFieldName.IsUnknown() {
				pfnVal := detData.PartitionFieldName.ValueString()
				apiDet.PartitionFieldName = &pfnVal
			}
			apiReq.AnalysisConfig.Detectors[i] = apiDet
		}
	}

	if !plan.DataDescription.IsNull() && !plan.DataDescription.IsUnknown() {
		var dataDescriptionData dataDescriptionModel
		resp.Diagnostics.Append(plan.DataDescription.As(ctx, &dataDescriptionData, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiReq.DataDescription = &apiDataDescription{
			TimeField: dataDescriptionData.TimeField.ValueString(),
		}
	}

	if !plan.ModelPlotConfig.IsNull() && !plan.ModelPlotConfig.IsUnknown() {
		var modelPlotConfigData modelPlotConfigModel
		resp.Diagnostics.Append(plan.ModelPlotConfig.As(ctx, &modelPlotConfigData, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiReq.ModelPlotConfig = &apiModelPlotConfig{
			Enabled: modelPlotConfigData.Enabled.ValueBool(),
		}
	}

	// Analysis Limits
	if !plan.AnalysisLimits.IsNull() && !plan.AnalysisLimits.IsUnknown() {
		var analysisLimitsData analysisLimitsModel
		resp.Diagnostics.Append(plan.AnalysisLimits.As(ctx, &analysisLimitsData, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiAnalysisLimits := apiAnalysisLimits{}
		if !analysisLimitsData.ModelMemoryLimit.IsNull() && !analysisLimitsData.ModelMemoryLimit.IsUnknown() {
			mmlVal := analysisLimitsData.ModelMemoryLimit.ValueString()
			apiAnalysisLimits.ModelMemoryLimit = &mmlVal
		}
		if !analysisLimitsData.CategorizationExamplesLimit.IsNull() && !analysisLimitsData.CategorizationExamplesLimit.IsUnknown() {
			celVal := analysisLimitsData.CategorizationExamplesLimit.ValueInt64()
			apiAnalysisLimits.CategorizationExamplesLimit = &celVal
		}
		apiReq.AnalysisLimits = &apiAnalysisLimits
	} else if plan.AnalysisLimits.IsNull() { // Handle explicit null to clear analysis_limits
		apiReq.AnalysisLimits = &apiAnalysisLimits{} // Send empty struct to clear, or nil if API supports that for clearing
	}

	// ModelSnapshotRetentionDays
	if !plan.ModelSnapshotRetentionDays.IsNull() && !plan.ModelSnapshotRetentionDays.IsUnknown() {
		msrdVal := plan.ModelSnapshotRetentionDays.ValueInt64()
		apiReq.ModelSnapshotRetentionDays = &msrdVal
	}

	// ResultsRetentionDays
	if !plan.ResultsRetentionDays.IsNull() && !plan.ResultsRetentionDays.IsUnknown() {
		rrdVal := plan.ResultsRetentionDays.ValueInt64()
		apiReq.ResultsRetentionDays = &rrdVal
	}

	// AllowLazyOpen
	if !plan.AllowLazyOpen.IsNull() && !plan.AllowLazyOpen.IsUnknown() {
		aloVal := plan.AllowLazyOpen.ValueBool()
		apiReq.AllowLazyOpen = &aloVal
	}

	// CategorizationFieldName and SummaryCountFieldName in AnalysisConfig
	// Need to re-fetch analysisConfigData if it wasn't fetched before, or ensure it's available
	// For simplicity, assuming analysisConfigData is populated if AnalysisConfig is being updated.
	if apiReq.AnalysisConfig != nil { // Ensure AnalysisConfig was initialized if any part of it is updated
		var analysisConfigData analysisConfigModel // This might be redundant if already populated, ensure correct scope
		// If plan.AnalysisConfig is not null, it implies we are processing it.
		// We should have already called plan.AnalysisConfig.As(ctx, &analysisConfigData, ...)
		// The following logic assumes analysisConfigData is correctly populated from the plan.
		resp.Diagnostics.Append(plan.AnalysisConfig.As(ctx, &analysisConfigData, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() { // Check after attempting to populate
			return
		}

		if !analysisConfigData.CategorizationFieldName.IsNull() && !analysisConfigData.CategorizationFieldName.IsUnknown() {
			cfnVal := analysisConfigData.CategorizationFieldName.ValueString()
			apiReq.AnalysisConfig.CategorizationFieldName = &cfnVal
		}
		if !analysisConfigData.SummaryCountFieldName.IsNull() && !analysisConfigData.SummaryCountFieldName.IsUnknown() {
			scfnVal := analysisConfigData.SummaryCountFieldName.ValueString()
			apiReq.AnalysisConfig.SummaryCountFieldName = &scfnVal
		}
	}

	// TimeFormat in DataDescription
	// Similar to AnalysisConfig, ensure dataDescriptionData is populated if DataDescription is being updated.
	if apiReq.DataDescription != nil { // Ensure DataDescription was initialized
		var dataDescriptionData dataDescriptionModel // Ensure correct scope
		resp.Diagnostics.Append(plan.DataDescription.As(ctx, &dataDescriptionData, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() { // Check after attempting to populate
			return
		}
		if !dataDescriptionData.TimeFormat.IsNull() && !dataDescriptionData.TimeFormat.IsUnknown() {
			tfVal := dataDescriptionData.TimeFormat.ValueString()
			apiReq.DataDescription.TimeFormat = &tfVal
		}
	}

	// DailyModelSnapshotRetentionAfterDays
	if !plan.DailyModelSnapshotRetentionAfterDays.IsNull() && !plan.DailyModelSnapshotRetentionAfterDays.IsUnknown() {
		dmsradVal := plan.DailyModelSnapshotRetentionAfterDays.ValueInt64()
		apiReq.DailyModelSnapshotRetentionAfterDays = &dmsradVal
	}

	// CustomSettings
	if !plan.CustomSettings.IsNull() && !plan.CustomSettings.IsUnknown() {
		var tfCustomSettings map[string]string
		resp.Diagnostics.Append(plan.CustomSettings.ElementsAs(ctx, &tfCustomSettings, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Convert map[string]string to map[string]interface{} for the API request
		apiCustomSettings := make(map[string]interface{}, len(tfCustomSettings))
		for k, v := range tfCustomSettings {
			apiCustomSettings[k] = v
		}
		apiReq.CustomSettings = apiCustomSettings
	} else if plan.CustomSettings.IsNull() { // Handle explicit null to clear custom_settings
		apiReq.CustomSettings = make(map[string]interface{}) // Send empty map to clear, or nil if API supports that for clearing
	}

	// Latency in AnalysisConfig
	if apiReq.AnalysisConfig != nil { // Ensure AnalysisConfig was initialized
		// analysisConfigDataForDetectors is populated from plan.AnalysisConfig.As(...) below for detectors
		// We can reuse it here if it's already populated, or ensure it's populated before this block.
		// For safety, let's ensure analysisConfigData is populated from the plan if we are setting latency.
		var analysisConfigDataForLatency analysisConfigModel
		if !plan.AnalysisConfig.IsNull() && !plan.AnalysisConfig.IsUnknown() { // only try to get if plan has it
			resp.Diagnostics.Append(plan.AnalysisConfig.As(ctx, &analysisConfigDataForLatency, basetypes.ObjectAsOptions{})...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !analysisConfigDataForLatency.Latency.IsNull() && !analysisConfigDataForLatency.Latency.IsUnknown() {
				latVal := analysisConfigDataForLatency.Latency.ValueString()
				apiReq.AnalysisConfig.Latency = &latVal
			}
		}
	}

	// DetectorDescription and UseNull in Detectors
	// This assumes 'analysisConfigData' and 'detectorsData' are populated from the plan if 'plan.AnalysisConfig' is not null.
	if apiReq.AnalysisConfig != nil && len(apiReq.AnalysisConfig.Detectors) > 0 {
		// We need to ensure detectorsData is populated from the plan to access DetectorDescription and UseNull
		var analysisConfigDataForDetectors analysisConfigModel
		resp.Diagnostics.Append(plan.AnalysisConfig.As(ctx, &analysisConfigDataForDetectors, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		var detectorsDataFromPlan []detectorModel
		resp.Diagnostics.Append(analysisConfigDataForDetectors.Detectors.ElementsAs(ctx, &detectorsDataFromPlan, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(detectorsDataFromPlan) == len(apiReq.AnalysisConfig.Detectors) {
			for i, detData := range detectorsDataFromPlan {
				if !detData.DetectorDescription.IsNull() && !detData.DetectorDescription.IsUnknown() {
					ddVal := detData.DetectorDescription.ValueString()
					apiReq.AnalysisConfig.Detectors[i].DetectorDescription = &ddVal
				}
				if !detData.UseNull.IsNull() && !detData.UseNull.IsUnknown() {
					unVal := detData.UseNull.ValueBool()
					apiReq.AnalysisConfig.Detectors[i].UseNull = &unVal
				}
				if !detData.ExcludeFrequent.IsNull() && !detData.ExcludeFrequent.IsUnknown() {
					efVal := detData.ExcludeFrequent.ValueString()
					apiReq.AnalysisConfig.Detectors[i].ExcludeFrequent = &efVal
				}

				// Custom Rules for this detector
				if !detData.CustomRules.IsNull() && !detData.CustomRules.IsUnknown() {
					var customRulesData []customRuleModel
					resp.Diagnostics.Append(detData.CustomRules.ElementsAs(ctx, &customRulesData, false)...)
					if resp.Diagnostics.HasError() {
						return
					}
					apiReq.AnalysisConfig.Detectors[i].CustomRules = make([]apiCustomRule, len(customRulesData))
					for j, ruleData := range customRulesData {
						apiRule := apiCustomRule{}
						resp.Diagnostics.Append(ruleData.Actions.ElementsAs(ctx, &apiRule.Actions, false)...)
						if resp.Diagnostics.HasError() {
							return
						}

						// Set the Scope for the apiCustomRule directly from ruleData
						if !ruleData.Scope.IsNull() && !ruleData.Scope.IsUnknown() {
							scopeVal := ruleData.Scope.ValueString()
							apiRule.Scope = &scopeVal
						}

						if !ruleData.Conditions.IsNull() && !ruleData.Conditions.IsUnknown() {
							var conditionsData []ruleConditionModel
							resp.Diagnostics.Append(ruleData.Conditions.ElementsAs(ctx, &conditionsData, false)...)
							if resp.Diagnostics.HasError() {
								return
							}
							apiRule.Conditions = make([]apiRuleCondition, len(conditionsData))
							for k, condData := range conditionsData {
								// Scope is now part of apiCustomRule, not apiRuleCondition
								apiRule.Conditions[k] = apiRuleCondition{
									Operator: condData.Operator.ValueString(),
									Value:    condData.Value.ValueFloat64(),
								}
							}
						}
						apiReq.AnalysisConfig.Detectors[i].CustomRules[j] = apiRule
					}
				}
			}
		}
	}

	bodyBytes, err := json.Marshal(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to serialize ML job update request", err.Error())
		return
	}

	mlClient := esClient.ML
	apiEsResp, err := mlClient.UpdateJob(jobId, bytes.NewReader(bodyBytes), mlClient.UpdateJob.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update ML job", err.Error())
		return
	}
	defer apiEsResp.Body.Close()
	if apiEsResp.IsError() {
		sdkDiags := utils.CheckError(apiEsResp, "Failed to update ML job")
		resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	plan.ID = types.StringValue(jobId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
