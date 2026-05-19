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
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func updateAnomalyDetectionJob(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[TFModel]) (entitycore.WriteResult[TFModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	jobID := req.WriteID

	tflog.Debug(ctx, fmt.Sprintf("Updating ML anomaly detection job: %s", jobID))

	updateBody := &UpdateAPIModel{}
	hasChanges, buildDiags := updateBody.BuildFromPlan(ctx, &plan, req.Prior)
	diags.Append(buildDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	if !hasChanges {
		tflog.Debug(ctx, fmt.Sprintf("No updates needed for ML anomaly detection job: %s", jobID))
		diags.AddWarning(
			"No changes detected to updatable fields during an update operation",
			`
Changes to non-updateable fields should force a recreation of the anomaly detection job.
Please report this warning to the provider developers.`,
		)
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	// Send the update as raw JSON so that all fields including
	// categorization_examples_limit are included. The typed updatejob.Request
	// uses types.AnalysisMemoryLimit which only models model_memory_limit,
	// dropping categorization_examples_limit.
	updateJSON, err := json.Marshal(updateBody)
	if err != nil {
		diags.AddError("Failed to marshal ML anomaly detection job update", err.Error())
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}
	_, err = typedClient.Ml.UpdateJob(jobID).Raw(bytes.NewReader(updateJSON)).Do(ctx)
	if err != nil {
		diags.AddError(
			"Failed to update ML anomaly detection job",
			fmt.Sprintf("Unable to update ML anomaly detection job: %s — %s", jobID, err.Error()),
		)
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML anomaly detection job: %s", jobID))
	return entitycore.WriteResult[TFModel]{Model: plan}, diags
}
