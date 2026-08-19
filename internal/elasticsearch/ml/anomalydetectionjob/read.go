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
	"fmt"

	esv8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/getjobs"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// readAnomalyDetectionJob fetches the job from Elasticsearch and populates the
// TF model. It satisfies the entitycore ElasticsearchReadFunc[TFModel] signature.
func readAnomalyDetectionJob(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state TFModel) (TFModel, bool, fwdiags.Diagnostics) {
	jobID := resourceID
	if jobID == "" {
		var diags fwdiags.Diagnostics
		diags.AddError("Invalid resource ID", "job_id cannot be empty")
		return state, false, diags
	}

	return entitycore.SimpleElasticsearchRead[TFModel, getjobs.Response](
		"ML anomaly detection job",
		func(ctx context.Context, typedClient *esv8.TypedClient, jobID string) (*getjobs.Response, error) {
			return typedClient.Ml.GetJobs().JobId(jobID).AllowNoMatch(true).Do(ctx)
		},
		func(model *TFModel, ctx context.Context, res *getjobs.Response) (bool, fwdiags.Diagnostics) {
			var diags fwdiags.Diagnostics

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

			apiModel := fromTypedJob(&res.Jobs[0])
			diags.Append(model.fromAPIModel(ctx, apiModel)...)
			if diags.HasError() {
				return false, diags
			}

			return true, diags
		},
	)(ctx, client, jobID, state)
}
