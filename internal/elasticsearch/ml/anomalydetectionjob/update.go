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

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *anomalyDetectionJobResource) update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var plan TFModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state TFModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := state.JobID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Updating ML anomaly detection job: %s", jobID))

	// Create update body with only updatable fields
	updateBody := &UpdateAPIModel{}
	hasChanges, diags := updateBody.BuildFromPlan(ctx, &plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only proceed with update if there are changes
	if !hasChanges {
		tflog.Debug(ctx, fmt.Sprintf("No updates needed for ML anomaly detection job: %s", jobID))
		resp.Diagnostics.AddWarning("No changes detected to updatable fields during an update operation", `
Changes to non-updateable fields should force a recreation of the anomaly detection job.
Please report this warning to the provider developers.`)
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, plan.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	typedClient, err := client.GetESTypedClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Elasticsearch client", err.Error())
		return
	}

	// Send the update as raw JSON so that all fields including
	// categorization_examples_limit are included. The typed updatejob.Request
	// uses types.AnalysisMemoryLimit which only models model_memory_limit,
	// dropping categorization_examples_limit.
	updateJSON, err := json.Marshal(updateBody)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal ML anomaly detection job update", err.Error())
		return
	}
	_, err = typedClient.Ml.UpdateJob(jobID).Raw(bytes.NewReader(updateJSON)).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update ML anomaly detection job", fmt.Sprintf("Unable to update ML anomaly detection job: %s — %s", jobID, err.Error()))
		return
	}

	// Read the updated job to get the current state
	found, diags := r.read(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Set the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML anomaly detection job: %s", jobID))
}
