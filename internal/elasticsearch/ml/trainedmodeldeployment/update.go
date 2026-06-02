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

package trainedmodeldeployment

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const updateTimeoutErrorMessage = "Timed out while waiting for trained model deployment update. Timeout: %s"

func (r *trainedModelDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TrainedModelDeploymentData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get update timeout
	updateTimeout, fwDiags := data.Timeouts.Update(ctx, 5*time.Minute)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	diags = r.update(ctx, req, &resp.State)
	if diagutil.ContainsContextDeadlineExceeded(ctx, diags) {
		diags.AddError("Operation timed out", fmt.Sprintf(updateTimeoutErrorMessage, updateTimeout))
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *trainedModelDeploymentResource) update(ctx context.Context, req resource.UpdateRequest, state *tfsdk.State) diag.Diagnostics {
	var plan TrainedModelDeploymentData
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return diags
	}

	var prior TrainedModelDeploymentData
	priorDiags := req.State.Get(ctx, &prior)
	diags.Append(priorDiags...)
	if diags.HasError() {
		return diags
	}

	client, fwDiags := r.Client().GetElasticsearchClient(ctx, plan.ElasticsearchConnection)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	deploymentID := prior.DeploymentID.ValueString()
	if deploymentID == "" {
		deploymentID = plan.DeploymentID.ValueString()
	}

	modelID := prior.ModelID.ValueString()
	if modelID == "" {
		modelID = plan.ModelID.ValueString()
	}

	// Build update options
	var adaptiveAllocations *types.AdaptiveAllocationsSettings
	if !plan.AdaptiveAllocations.Enabled.IsNull() {
		aa := plan.AdaptiveAllocations
		adaptiveAllocations = &types.AdaptiveAllocationsSettings{
			Enabled: aa.Enabled.ValueBool(),
		}
		if !aa.MinNumberOfAllocations.IsNull() {
			v := int(aa.MinNumberOfAllocations.ValueInt64())
			adaptiveAllocations.MinNumberOfAllocations = &v
		}
		if !aa.MaxNumberOfAllocations.IsNull() {
			v := int(aa.MaxNumberOfAllocations.ValueInt64())
			adaptiveAllocations.MaxNumberOfAllocations = &v
		}
	}

	var numberOfAllocations *int
	if !plan.NumberOfAllocations.IsNull() {
		v := int(plan.NumberOfAllocations.ValueInt64())
		numberOfAllocations = &v
	}

	updateOpts := elasticsearch.UpdateTrainedModelDeploymentOptions{
		NumberOfAllocations: numberOfAllocations,
		AdaptiveAllocations: adaptiveAllocations,
	}

	updateDiags := elasticsearch.UpdateTrainedModelDeployment(ctx, client, deploymentID, updateOpts)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return diags
	}

	// Re-read stats to update state
	statsJSON, stats, readDiags := elasticsearch.GetTrainedModelStatsJSON(ctx, client, modelID, deploymentID)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	if stats == nil || stats.DeploymentStats == nil {
		diags.AddError("Deployment not found after update", fmt.Sprintf("Trained model deployment %s not found after updating", deploymentID))
		return diags
	}

	// Update state with current values
	if stats.DeploymentStats.State != nil {
		plan.State = fwtypes.StringValue(stats.DeploymentStats.State.String())
	} else {
		plan.State = fwtypes.StringNull()
	}
	if stats.DeploymentStats.AllocationStatus != nil {
		plan.AllocationStatus = fwtypes.StringValue(stats.DeploymentStats.AllocationStatus.State.String())
	} else {
		plan.AllocationStatus = fwtypes.StringNull()
	}
	plan.StatsJSON = fwtypes.StringValue(statsJSON)

	// Update number_of_allocations from API only when adaptive_allocations is NOT configured
	if plan.AdaptiveAllocations.Enabled.IsNull() {
		if stats.DeploymentStats.NumberOfAllocations != nil {
			plan.NumberOfAllocations = fwtypes.Int64Value(int64(*stats.DeploymentStats.NumberOfAllocations))
		} else {
			plan.NumberOfAllocations = fwtypes.Int64Null()
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Trained model deployment %s updated successfully", deploymentID))

	persistDiags := state.Set(ctx, &plan)
	diags.Append(persistDiags...)
	return diags
}
