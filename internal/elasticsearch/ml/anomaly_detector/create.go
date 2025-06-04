package anomaly_detector

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

	data.ID = data.JobId // Use JobId as the Terraform resource ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
