package anomaly_detector

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *anomalyDetectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state anomalyDetectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	esClient, err := r.client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Elasticsearch client", err.Error())
		return
	}

	jobId := state.JobId.ValueString()
	mlClient := esClient.ML
	apiEsResp, err := mlClient.DeleteJob(jobId, mlClient.DeleteJob.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete ML job", err.Error())
		return
	}
	defer apiEsResp.Body.Close()

	if apiEsResp.StatusCode == http.StatusNotFound {
		return // Successfully deleted or already gone
	}

	if apiEsResp.IsError() {
		sdkDiags := utils.CheckError(apiEsResp, "Failed to delete ML job")
		resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}
}
