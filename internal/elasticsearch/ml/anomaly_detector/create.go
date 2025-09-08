package anomaly_detector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *anomalyDetectorJobResource) create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
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
	apiModel, diags := plan.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating ML anomaly detection job: %s", jobID))

	esClient, err := r.client.GetESClient()
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

	if diags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to create ML anomaly detection job: %s", jobID)); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Read the created job to get the full state.
	compID, sdkDiags := r.client.ID(ctx, jobID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(compID.String())
	found, diags := r.read(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read created job", fmt.Sprintf("Job with ID %s not found after creation", jobID))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, fmt.Sprintf("Successfully created ML anomaly detection job: %s", jobID))
}
