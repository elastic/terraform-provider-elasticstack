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

func (r *anomalyDetectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data anomalyDetectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	esClient, err := r.client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Elasticsearch client", err.Error())
		return
	}

	jobId := data.JobId.ValueString()

	var apiReq apiCreateAnomalyDetectorRequest

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		descVal := data.Description.ValueString()
		apiReq.Description = &descVal
	}

	if !data.ResultsIndexName.IsNull() && !data.ResultsIndexName.IsUnknown() {
		rinVal := data.ResultsIndexName.ValueString()
		apiReq.ResultsIndexName = &rinVal
	}

	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		resp.Diagnostics.Append(data.Groups.ElementsAs(ctx, &apiReq.Groups, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Analysis Config
	var analysisConfigData analysisConfigModel
	resp.Diagnostics.Append(data.AnalysisConfig.As(ctx, &analysisConfigData, basetypes.ObjectAsOptions{})...)
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

	// Detectors
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

	// Data Description
	var dataDescriptionData dataDescriptionModel
	resp.Diagnostics.Append(data.DataDescription.As(ctx, &dataDescriptionData, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiReq.DataDescription = &apiDataDescription{
		TimeField: dataDescriptionData.TimeField.ValueString(),
	}

	// Model Plot Config
	if !data.ModelPlotConfig.IsNull() && !data.ModelPlotConfig.IsUnknown() {
		var modelPlotConfigData modelPlotConfigModel
		resp.Diagnostics.Append(data.ModelPlotConfig.As(ctx, &modelPlotConfigData, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiReq.ModelPlotConfig = &apiModelPlotConfig{
			Enabled: modelPlotConfigData.Enabled.ValueBool(),
		}
	}

	// Analysis Limits
	if !data.AnalysisLimits.IsNull() && !data.AnalysisLimits.IsUnknown() {
		var analysisLimitsData analysisLimitsModel
		resp.Diagnostics.Append(data.AnalysisLimits.As(ctx, &analysisLimitsData, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiReq.AnalysisLimits = &apiAnalysisLimits{}
		if !analysisLimitsData.ModelMemoryLimit.IsNull() && !analysisLimitsData.ModelMemoryLimit.IsUnknown() {
			mlVal := analysisLimitsData.ModelMemoryLimit.ValueString()
			apiReq.AnalysisLimits.ModelMemoryLimit = &mlVal
		}
		if !analysisLimitsData.CategorizationExamplesLimit.IsNull() && !analysisLimitsData.CategorizationExamplesLimit.IsUnknown() {
			celVal := analysisLimitsData.CategorizationExamplesLimit.ValueInt64()
			apiReq.AnalysisLimits.CategorizationExamplesLimit = &celVal
		}
	}

	// ModelSnapshotRetentionDays
	if !data.ModelSnapshotRetentionDays.IsNull() && !data.ModelSnapshotRetentionDays.IsUnknown() {
		msrdVal := data.ModelSnapshotRetentionDays.ValueInt64()
		apiReq.ModelSnapshotRetentionDays = &msrdVal
	}

	// ResultsRetentionDays
	if !data.ResultsRetentionDays.IsNull() && !data.ResultsRetentionDays.IsUnknown() {
		rrdVal := data.ResultsRetentionDays.ValueInt64()
		apiReq.ResultsRetentionDays = &rrdVal
	}

	// AllowLazyOpen
	if !data.AllowLazyOpen.IsNull() && !data.AllowLazyOpen.IsUnknown() {
		aloVal := data.AllowLazyOpen.ValueBool()
		apiReq.AllowLazyOpen = &aloVal
	}

	// CategorizationFieldName and SummaryCountFieldName in AnalysisConfig
	if apiReq.AnalysisConfig != nil { // Ensure AnalysisConfig was initialized
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
	if apiReq.DataDescription != nil { // Ensure DataDescription was initialized
		if !dataDescriptionData.TimeFormat.IsNull() && !dataDescriptionData.TimeFormat.IsUnknown() {
			tfVal := dataDescriptionData.TimeFormat.ValueString()
			apiReq.DataDescription.TimeFormat = &tfVal
		}
	}

	// DailyModelSnapshotRetentionAfterDays
	if !data.DailyModelSnapshotRetentionAfterDays.IsNull() && !data.DailyModelSnapshotRetentionAfterDays.IsUnknown() {
		dmsradVal := data.DailyModelSnapshotRetentionAfterDays.ValueInt64()
		apiReq.DailyModelSnapshotRetentionAfterDays = &dmsradVal
	}

	// CustomSettings
	if !data.CustomSettings.IsNull() && !data.CustomSettings.IsUnknown() {
		var tfCustomSettings map[string]string
		resp.Diagnostics.Append(data.CustomSettings.ElementsAs(ctx, &tfCustomSettings, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Convert map[string]string to map[string]interface{} for the API request
		apiCustomSettings := make(map[string]interface{}, len(tfCustomSettings))
		for k, v := range tfCustomSettings {
			apiCustomSettings[k] = v
		}
		apiReq.CustomSettings = apiCustomSettings
	}

	// Latency in AnalysisConfig
	if apiReq.AnalysisConfig != nil { // Ensure AnalysisConfig was initialized
		// analysisConfigData should be populated from plan.AnalysisConfig.As(...) earlier in the function
		if !analysisConfigData.Latency.IsNull() && !analysisConfigData.Latency.IsUnknown() {
			latVal := analysisConfigData.Latency.ValueString()
			apiReq.AnalysisConfig.Latency = &latVal
		}
	}

	// DetectorDescription and UseNull in Detectors
	if apiReq.AnalysisConfig != nil && len(apiReq.AnalysisConfig.Detectors) > 0 && len(detectorsData) == len(apiReq.AnalysisConfig.Detectors) {
		for i, detData := range detectorsData {
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

	bodyBytes, err := json.Marshal(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to serialize ML job request", err.Error())
		return
	}

	mlClient := esClient.ML
	apiEsResp, err := mlClient.PutJob(jobId, bytes.NewReader(bodyBytes), mlClient.PutJob.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create ML job", err.Error())
		return
	}
	defer apiEsResp.Body.Close()
	if apiEsResp.IsError() {
		sdkDiags := utils.CheckError(apiEsResp, "Failed to create ML job")
		resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	resourceId, sdkDiags := r.client.ID(ctx, jobId)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}
	data.ID = types.StringValue(resourceId.String())
	if data.ResultsIndexName.IsUnknown() {
		data.ResultsIndexName = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
