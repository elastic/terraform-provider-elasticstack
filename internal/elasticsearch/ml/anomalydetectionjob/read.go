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

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *anomalyDetectionJobResource) read(ctx context.Context, job *TFModel) (bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	if !r.resourceReady(&diags) {
		return false, diags
	}

	jobID := job.JobID.ValueString()
	tflog.Debug(ctx, fmt.Sprintf("Reading ML anomaly detection job: %s", jobID))

	client, connDiags := r.Client().GetElasticsearchClient(ctx, job.ElasticsearchConnection)
	diags.Append(connDiags...)
	if diags.HasError() {
		return false, diags
	}

	typedClient, err := client.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return false, diags
	}

	// Get the ML job using the typed client
	res, err := typedClient.Ml.GetJobs().JobId(jobID).AllowNoMatch(true).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return false, nil
		}
		diags.AddError("Failed to get ML anomaly detection job", fmt.Sprintf("Unable to get ML anomaly detection job: %s — %s", jobID, err.Error()))
		return false, diags
	}

	if len(res.Jobs) == 0 {
		return false, nil
	}

	if len(res.Jobs) > 1 {
		jobIDs := make([]string, len(res.Jobs))
		for i, j := range res.Jobs {
			jobIDs[i] = j.JobId
		}
		diags.AddWarning(
			"Getting jobs by ID returned multiple results",
			fmt.Sprintf(
				"Expected a single result when getting anomaly detection jobs by ID. However the API returned %d jobs with IDs %v",
				len(res.Jobs),
				jobIDs,
			),
		)
	}

	// Convert the typed response to APIModel, then populate TF model
	apiModel := fromTypedJob(&res.Jobs[0])
	diags.Append(job.fromAPIModel(ctx, apiModel)...)
	if diags.HasError() {
		return false, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read ML anomaly detection job: %s", jobID))
	return true, diags
}
