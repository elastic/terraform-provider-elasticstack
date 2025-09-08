package anomaly_detector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, client *clients.ApiClient) {
	if !resourceReady(client, &resp.Diagnostics) {
		return
	}

	var state AnomalyDetectorJobTFModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := state.JobID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Reading ML anomaly detection job: %s", jobID))

	esClient, err := client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	// Get the ML job
	res, err := esClient.ML.GetJobs(esClient.ML.GetJobs.WithJobID(jobID), esClient.ML.GetJobs.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get ML anomaly detection job", err.Error())
		return
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if diags := utils.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML anomaly detection job: %s", jobID)); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Parse the response
	var response struct {
		Jobs []AnomalyDetectorJobAPIModel `json:"jobs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		resp.Diagnostics.AddError("Failed to decode job response", err.Error())
		return
	}

	if len(response.Jobs) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert API response back to TF model
	newState, diags := apiModelToTFModel(ctx, response.Jobs[0])
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the ID and any plan-only values
	newState.ID = state.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)

	tflog.Debug(ctx, fmt.Sprintf("Successfully read ML anomaly detection job: %s", jobID))
}
