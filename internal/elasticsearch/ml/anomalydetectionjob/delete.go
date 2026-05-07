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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	elasticsearch "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// deleteAnomalyDetectionJob closes and deletes the ML job. It satisfies the
// entitycore elasticsearchDeleteFunc[TFModel] signature.
func deleteAnomalyDetectionJob(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state TFModel) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	jobID := resourceID
	if jobID == "" {
		diags.AddError("Invalid resource ID", "job_id cannot be empty")
		return diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting ML anomaly detection job: %s", jobID))

	deleteTimeout, fwDiags := state.Timeouts.Delete(ctx, 20*time.Minute)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return diags
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
			diags.AddError(
				"Timeout waiting for ML job to close",
				fmt.Sprintf("ML job %s did not close within the allotted time: %s", jobID, err.Error()),
			)
			return diags
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
			return diags
		}
		tflog.Warn(ctx, fmt.Sprintf("Initial delete of ML job %s failed, retrying with force=true: %s", jobID, err.Error()))
		_, retryErr := typedClient.Ml.DeleteJob(jobID).Force(true).Do(ctx)
		if retryErr != nil {
			if errors.As(retryErr, &esErr) && esErr.Status == 404 {
				tflog.Debug(ctx, fmt.Sprintf("ML job %s not found on force-delete retry, treating as success", jobID))
				return diags
			}
			diags.AddError("Failed to delete ML anomaly detection job", fmt.Sprintf("Unable to delete ML anomaly detection job: %s — %s", jobID, retryErr.Error()))
			return diags
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully deleted ML anomaly detection job: %s", jobID))
	return diags
}
