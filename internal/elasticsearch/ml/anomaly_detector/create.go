package anomaly_detector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, client *clients.ApiClient) {
	if !resourceReady(client, &resp.Diagnostics) {
		return
	}

	var plan AnomalyDetectorJobTFModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := plan.JobID.ValueString()

	// Convert TF model to API model
	apiModel, diags := tfModelToAPIModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating ML anomaly detection job: %s", jobID))

	esClient, err := client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	// Marshal the API model to JSON
	body, err := json.Marshal(apiModel)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal job configuration", err.Error())
		return
	}

	// Create the ML job
	res, err := esClient.ML.PutJob(jobID, bytes.NewReader(body), esClient.ML.PutJob.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create ML anomaly detection job", err.Error())
		return
	}
	defer res.Body.Close()

	if diags := utils.CheckErrorFromFW(res, fmt.Sprintf("Unable to create ML anomaly detection job: %s", jobID)); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Parse the response to get created job details
	var response AnomalyDetectorJobAPIModel
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		resp.Diagnostics.AddError("Failed to decode job response", err.Error())
		return
	}

	// Convert API response back to TF model
	state, diags := apiModelToTFModel(ctx, response)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ID for the resource
	state.ID = types.StringValue(jobID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	tflog.Debug(ctx, fmt.Sprintf("Successfully created ML anomaly detection job: %s", jobID))
}
