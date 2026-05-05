// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package anomalydetectionjob

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	elasticsearch "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *anomalyDetectionJobResource) delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

	var data TFModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, fwDiags := data.Timeouts.Delete(ctx, 20*time.Minute)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	client, diags := r.Client().GetElasticsearchClient(ctx, data.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	typedClient, err := client.GetESClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	// First, close the job if it's open. Force=true and AllowNoMatch=true to be safe.
	_, err = typedClient.Ml.CloseJob(jobID).Force(true).AllowNoMatch(true).Do(ctx)
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("Failed to close ML job %s before deletion: %s", jobID, err.Error()))
		// Continue with deletion even if close fails, as the job might already be closed
	}

	// Wait for the job to reach closed state before deleting. Elasticsearch uses
	// optimistic concurrency control on .ml-config: if CloseJob's state
	// transition commits after DeleteJob reads the seqNo but before it writes,
	// the delete fails with HTTP 409 version_conflict_engine_exception.
	if err := elasticsearch.WaitForMLJobClosed(ctx, client, jobID); err != nil {
		// If the Terraform operation context expired during the wait, surface the
		// timeout directly rather than letting it propagate as a confusing delete
		// error.
		if ctx.Err() != nil {
			resp.Diagnostics.AddError(
				"Timeout waiting for ML job to close",
				fmt.Sprintf("ML job %s did not close within the allotted time: %s", jobID, err.Error()),
			)
			return
		}
		tflog.Warn(ctx, fmt.Sprintf("Failed to wait for ML job %s to close before deletion: %s", jobID, err.Error()))
		// Continue with deletion even if the wait fails for non-timeout reasons.
	}

	// Delete the ML job. If the first attempt fails, retry once with force=true.
	// 404 means the job is already gone — treat it as success (idempotent delete).
	_, err = typedClient.Ml.DeleteJob(jobID).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			tflog.Debug(ctx, fmt.Sprintf("ML job %s not found on delete, treating as success", jobID))
			return
		}
		tflog.Warn(ctx, fmt.Sprintf("Initial delete of ML job %s failed, retrying with force=true: %s", jobID, err.Error()))
		_, retryErr := typedClient.Ml.DeleteJob(jobID).Force(true).Do(ctx)
		if retryErr != nil {
			if errors.As(retryErr, &esErr) && esErr.Status == 404 {
				tflog.Debug(ctx, fmt.Sprintf("ML job %s not found on force-delete retry, treating as success", jobID))
				return
			}
			resp.Diagnostics.AddError("Failed to delete ML anomaly detection job", fmt.Sprintf("Unable to delete ML anomaly detection job: %s — %s", jobID, retryErr.Error()))
			return
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully deleted ML anomaly detection job: %s", jobID))
}
