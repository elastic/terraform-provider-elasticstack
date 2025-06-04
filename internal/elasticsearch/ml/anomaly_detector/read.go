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
		analysisConfigObj, diags := types.ObjectValue(analysisConfigAttrTypes, map[string]attr.Value{
			"bucket_span": types.StringValue(jobDetails.AnalysisConfig.BucketSpan),
			"detectors":   detectorsList,
			"influencers": influencersList,
		})
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
		dataDescriptionObj, diags := types.ObjectValue(dataDescriptionAttrTypes, map[string]attr.Value{
			"time_field": types.StringValue(jobDetails.DataDescription.TimeField),
		})
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
