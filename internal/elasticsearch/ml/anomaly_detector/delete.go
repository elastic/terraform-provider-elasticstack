package anomaly_detector

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *anomalyDetectorJobResource) delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var jobIDValue basetypes.StringValue
	diags := req.State.GetAttribute(ctx, path.Root("job_id"), &jobIDValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := jobIDValue.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Deleting ML anomaly detection job: %s", jobID))

	esClient, err := r.client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	// First, close the job if it's open
	closeRes, err := esClient.ML.CloseJob(jobID, esClient.ML.CloseJob.WithContext(ctx))
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("Failed to close ML job %s before deletion: %s", jobID, err.Error()))
		// Continue with deletion even if close fails, as the job might already be closed
	} else {
		defer closeRes.Body.Close()
		if closeRes.StatusCode != 200 && closeRes.StatusCode != 409 { // 409 means already closed
			tflog.Warn(ctx, fmt.Sprintf("Failed to close ML job %s: status %d", jobID, closeRes.StatusCode))
		}
	}

	// Delete the ML job
	res, err := esClient.ML.DeleteJob(jobID, esClient.ML.DeleteJob.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete ML anomaly detection job", err.Error())
		return
	}
	defer res.Body.Close()

	if diags := diagutil.CheckErrorFromFW(res, fmt.Sprintf("Unable to delete ML anomaly detection job: %s", jobID)); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully deleted ML anomaly detection job: %s", jobID))
}
