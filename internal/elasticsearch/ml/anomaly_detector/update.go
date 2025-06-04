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
