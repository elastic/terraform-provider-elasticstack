package anomaly_detector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *anomalyDetectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data anomalyDetectorResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	esClient, err := r.client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Elasticsearch client", err.Error())
		return
	}

	jobId := data.JobId.ValueString()
	mlClient := esClient.ML
	apiEsResp, err := mlClient.GetJobs(mlClient.GetJobs.WithJobID(jobId), mlClient.GetJobs.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get ML job", err.Error())
		return
	}
	defer apiEsResp.Body.Close()

	if apiEsResp.StatusCode == http.StatusNotFound {
		resp.Diagnostics.AddWarning("ML Job not found", fmt.Sprintf("Anomaly detector job '%s' not found, removing from state.", jobId))
		resp.State.RemoveResource(ctx)
		return
	}

	if apiEsResp.IsError() {
		sdkDiags := utils.CheckError(apiEsResp, "Failed to get ML job")
		resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	var getResp apiGetAnomalyDetectorResponse
	if err := json.NewDecoder(apiEsResp.Body).Decode(&getResp); err != nil {
		resp.Diagnostics.AddError("Failed to parse ML job response", err.Error())
		return
	}

	if getResp.Count == 0 || len(getResp.Jobs) == 0 {
		resp.Diagnostics.AddWarning("ML Job not found", fmt.Sprintf("Anomaly detector job '%s' not found (empty response), removing from state.", jobId))
		resp.State.RemoveResource(ctx)
		return
	}

	jobDetails := getResp.Jobs[0]

	// Update the model from jobDetails
	data.ID = types.StringValue(jobDetails.JobID)
	data.JobId = types.StringValue(jobDetails.JobID)

	if jobDetails.Description != nil {
		data.Description = types.StringValue(*jobDetails.Description)
	} else {
		data.Description = types.StringNull()
	}

	if jobDetails.ResultsIndexName != nil {
		data.ResultsIndexName = types.StringValue(*jobDetails.ResultsIndexName)
	} else {
		data.ResultsIndexName = types.StringNull()
	}

	groupsList, diags := types.ListValueFrom(ctx, types.StringType, jobDetails.Groups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Groups = groupsList

	// Top-level optional fields
	if jobDetails.ModelSnapshotRetentionDays != nil {
		data.ModelSnapshotRetentionDays = types.Int64Value(*jobDetails.ModelSnapshotRetentionDays)
	} else {
		data.ModelSnapshotRetentionDays = types.Int64Null()
	}
	if jobDetails.ResultsRetentionDays != nil {
		data.ResultsRetentionDays = types.Int64Value(*jobDetails.ResultsRetentionDays)
	} else {
		data.ResultsRetentionDays = types.Int64Null()
	}
	if jobDetails.AllowLazyOpen != nil {
		data.AllowLazyOpen = types.BoolValue(*jobDetails.AllowLazyOpen)
	} else {
		data.AllowLazyOpen = types.BoolNull()
	}

	// DailyModelSnapshotRetentionAfterDays
	if jobDetails.DailyModelSnapshotRetentionAfterDays != nil {
		data.DailyModelSnapshotRetentionAfterDays = types.Int64Value(*jobDetails.DailyModelSnapshotRetentionAfterDays)
	} else {
		data.DailyModelSnapshotRetentionAfterDays = types.Int64Null()
	}

	// Analysis Limits
	if jobDetails.AnalysisLimits != nil {
		analysisLimitsValues := map[string]attr.Value{}
		if jobDetails.AnalysisLimits.ModelMemoryLimit != nil {
			analysisLimitsValues["model_memory_limit"] = types.StringValue(*jobDetails.AnalysisLimits.ModelMemoryLimit)
		} else {
			analysisLimitsValues["model_memory_limit"] = types.StringNull()
		}
		mml := ""
		if jobDetails.AnalysisLimits.ModelMemoryLimit != nil {
			mml = *jobDetails.AnalysisLimits.ModelMemoryLimit
		}
		// CategorizationExamplesLimit
		var celVal types.Int64
		if jobDetails.AnalysisLimits.CategorizationExamplesLimit != nil {
			celVal = types.Int64Value(*jobDetails.AnalysisLimits.CategorizationExamplesLimit)
		} else {
			celVal = types.Int64Null()
		}

		data.AnalysisLimits, diags = types.ObjectValueFrom(ctx, analysisLimitsModel{}.attrTypes(ctx, resp.Diagnostics), analysisLimitsModel{
			ModelMemoryLimit:            types.StringValue(mml),
			CategorizationExamplesLimit: celVal,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		data.AnalysisLimits = types.ObjectNull(analysisLimitsModel{}.attrTypes(ctx, resp.Diagnostics))
	}

	// CustomSettings
	if jobDetails.CustomSettings != nil {
		// The API returns map[string]interface{}, but TF schema is map[string]string.
		// We need to convert. This might lose information if non-string values are present.
		csMap := make(map[string]string)
		for k, v := range jobDetails.CustomSettings {
			if vStr, ok := v.(string); ok {
				csMap[k] = vStr
			} else {
				// Optionally, add a diagnostic warning about non-string custom setting ignored
				resp.Diagnostics.AddWarning("Non-string custom setting ignored", fmt.Sprintf("Custom setting '%s' has a non-string value and will be ignored.", k))
			}
		}
		customSettingsVal, diags := types.MapValueFrom(ctx, types.StringType, csMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.CustomSettings = customSettingsVal
	} else {
		data.CustomSettings = types.MapNull(types.StringType)
	}

	// Analysis Config
	if jobDetails.AnalysisConfig != nil {
		var analysisConfigDetectors []detectorModel
		for _, apiDet := range jobDetails.AnalysisConfig.Detectors {
			detector := detectorModel{
				Function: types.StringValue(apiDet.Function),
			}
			if apiDet.FieldName != nil {
				detector.FieldName = types.StringValue(*apiDet.FieldName)
			} else {
				detector.FieldName = types.StringNull()
			}
			if apiDet.ByFieldName != nil {
				detector.ByFieldName = types.StringValue(*apiDet.ByFieldName)
			} else {
				detector.ByFieldName = types.StringNull()
			}
			if apiDet.PartitionFieldName != nil {
				detector.PartitionFieldName = types.StringValue(*apiDet.PartitionFieldName)
			} else {
				detector.PartitionFieldName = types.StringNull()
			}
			if apiDet.DetectorDescription != nil {
				detector.DetectorDescription = types.StringValue(*apiDet.DetectorDescription)
			} else {
				detector.DetectorDescription = types.StringNull()
			}
			if apiDet.UseNull != nil {
				detector.UseNull = types.BoolValue(*apiDet.UseNull)
			} else {
				detector.UseNull = types.BoolNull()
			}
			if apiDet.ExcludeFrequent != nil {
				detector.ExcludeFrequent = types.StringValue(*apiDet.ExcludeFrequent)
			} else {
				detector.ExcludeFrequent = types.StringNull()
			}
			if apiDet.CustomRules != nil {
				var customRulesModelList []customRuleModel
				for _, apiRule := range apiDet.CustomRules {
					crm := customRuleModel{}
					actionsVal, actDiags := types.ListValueFrom(ctx, types.StringType, apiRule.Actions)
					resp.Diagnostics.Append(actDiags...)
					if resp.Diagnostics.HasError() {
						return
					}
					crm.Actions = actionsVal

					// Populate rule-level scope from apiRule.Scope
					if apiRule.Scope != nil {
						crm.Scope = types.StringValue(*apiRule.Scope)
					} else {
						crm.Scope = types.StringNull()
					}

					if apiRule.Conditions != nil {
						var ruleConditionsModelList []ruleConditionModel
						for _, apiCond := range apiRule.Conditions {
							rcm := ruleConditionModel{
								// Scope is now part of apiCustomRule, not apiRuleCondition from API
								Operator: types.StringValue(apiCond.Operator),
								Value:    types.Float64Value(apiCond.Value),
							}
							ruleConditionsModelList = append(ruleConditionsModelList, rcm)
						}
						conditionObjectType := types.ObjectType{AttrTypes: ruleConditionModel{}.attrTypes(ctx, resp.Diagnostics)}
						conditionsVal, condDiags := types.ListValueFrom(ctx, conditionObjectType, ruleConditionsModelList)
						resp.Diagnostics.Append(condDiags...)
						if resp.Diagnostics.HasError() {
							return
						}
						crm.Conditions = conditionsVal
					} else {
						crm.Conditions = types.ListNull(types.ObjectType{AttrTypes: ruleConditionModel{}.attrTypes(ctx, resp.Diagnostics)})
					}
					customRulesModelList = append(customRulesModelList, crm)
				}
				customRuleObjectType := types.ObjectType{AttrTypes: customRuleModel{}.attrTypes(ctx, resp.Diagnostics)}
				customRulesVal, crDiags := types.ListValueFrom(ctx, customRuleObjectType, customRulesModelList)
				resp.Diagnostics.Append(crDiags...)
				if resp.Diagnostics.HasError() {
					return
				}
				detector.CustomRules = customRulesVal
			} else {
				detector.CustomRules = types.ListNull(types.ObjectType{AttrTypes: customRuleModel{}.attrTypes(ctx, resp.Diagnostics)})
			}
			analysisConfigDetectors = append(analysisConfigDetectors, detector)
		}

		detectorObjectType := types.ObjectType{AttrTypes: detectorModel{}.attrTypes(ctx, resp.Diagnostics)}
		detectorsList, diags := types.ListValueFrom(ctx, detectorObjectType, analysisConfigDetectors)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		influencersList, diags := types.ListValueFrom(ctx, types.StringType, jobDetails.AnalysisConfig.Influencers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		analysisConfigAttrTypes := analysisConfigModel{}.attrTypes(ctx, resp.Diagnostics)
		acValues := map[string]attr.Value{
			"bucket_span": types.StringValue(jobDetails.AnalysisConfig.BucketSpan),
			"detectors":   detectorsList,
			"influencers": influencersList,
		}
		if jobDetails.AnalysisConfig.CategorizationFieldName != nil {
			acValues["categorization_field_name"] = types.StringValue(*jobDetails.AnalysisConfig.CategorizationFieldName)
		} else {
			acValues["categorization_field_name"] = types.StringNull()
		}
		if jobDetails.AnalysisConfig.SummaryCountFieldName != nil {
			acValues["summary_count_field_name"] = types.StringValue(*jobDetails.AnalysisConfig.SummaryCountFieldName)
		} else {
			acValues["summary_count_field_name"] = types.StringNull()
		}
		if jobDetails.AnalysisConfig.Latency != nil {
			acValues["latency"] = types.StringValue(*jobDetails.AnalysisConfig.Latency)
		} else {
			acValues["latency"] = types.StringNull()
		}

		analysisConfigObj, diags := types.ObjectValue(analysisConfigAttrTypes, acValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.AnalysisConfig = analysisConfigObj
	} else {
		data.AnalysisConfig = types.ObjectNull(analysisConfigModel{}.attrTypes(ctx, resp.Diagnostics))
	}

	// Data Description
	if jobDetails.DataDescription != nil {
		dataDescriptionAttrTypes := dataDescriptionModel{}.attrTypes(ctx, resp.Diagnostics)
		ddValues := map[string]attr.Value{
			"time_field": types.StringValue(jobDetails.DataDescription.TimeField),
		}
		if jobDetails.DataDescription.TimeFormat != nil {
			ddValues["time_format"] = types.StringValue(*jobDetails.DataDescription.TimeFormat)
		} else {
			ddValues["time_format"] = types.StringNull()
		}
		dataDescriptionObj, diags := types.ObjectValue(dataDescriptionAttrTypes, ddValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.DataDescription = dataDescriptionObj
	} else {
		data.DataDescription = types.ObjectNull(dataDescriptionModel{}.attrTypes(ctx, resp.Diagnostics))
	}

	// Model Plot Config
	if jobDetails.ModelPlotConfig != nil {
		modelPlotConfigAttrTypes := modelPlotConfigModel{}.attrTypes(ctx, resp.Diagnostics)
		modelPlotObj, diags := types.ObjectValue(modelPlotConfigAttrTypes, map[string]attr.Value{
			"enabled": types.BoolValue(jobDetails.ModelPlotConfig.Enabled),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.ModelPlotConfig = modelPlotObj
	} else {
		data.ModelPlotConfig = types.ObjectNull(modelPlotConfigModel{}.attrTypes(ctx, resp.Diagnostics))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
