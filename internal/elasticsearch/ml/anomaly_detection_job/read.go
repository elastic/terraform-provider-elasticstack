package anomaly_detection_job

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *anomalyDetectionJobResource) read(ctx context.Context, job *AnomalyDetectionJobTFModel) (bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	if !r.resourceReady(&diags) {
		return false, diags
	}

	jobID := job.JobID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Reading ML anomaly detection job: %s", jobID))

	esClient, err := r.client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return false, diags
	}

	// Get the ML job
	res, err := esClient.ML.GetJobs(esClient.ML.GetJobs.WithJobID(jobID), esClient.ML.GetJobs.WithContext(ctx))
	if err != nil {
		diags.AddError("Failed to get ML anomaly detection job", err.Error())
		return false, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if d := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to get ML anomaly detection job: %s", jobID)); d.HasError() {
		diags.Append(d...)
		return false, diags
	}

	// Parse the response
	var response struct {
		Jobs []AnomalyDetectionJobAPIModel `json:"jobs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		diags.AddError("Failed to decode job response", err.Error())
		return false, diags
	}

	if len(response.Jobs) == 0 {
		return false, nil
	}

	// Convert API response back to TF model
	diags.Append(job.fromAPIModel(ctx, &response.Jobs[0])...)
	if diags.HasError() {
		return false, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read ML anomaly detection job: %s", jobID))
	return true, diags
}
