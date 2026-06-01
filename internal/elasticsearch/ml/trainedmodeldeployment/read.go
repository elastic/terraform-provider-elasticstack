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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readTrainedModelDeployment(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	resourceID string,
	state TrainedModelDeploymentData,
) (TrainedModelDeploymentData, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	modelID := state.ModelID.ValueString()
	if modelID == "" {
		// During import, resourceID is the composite id <cluster_uuid>/<deployment_id>.
		// We need to extract deployment_id and set model_id from it if possible.
		// However, model_id is part of the resource ID... Actually, for import we
		// need to handle this differently. The readFunc gets the prior state.
		// After import, state will have id set but model_id will be empty.
		// We need to use the deployment_id as the model_id fallback for import.
		modelID = resourceID
	}

	// Determine deployment_id
	deploymentID := state.DeploymentID.ValueString()
	if deploymentID == "" {
		deploymentID = resourceID
	}

	statsJSON, stats, getDiags := elasticsearch.GetTrainedModelStatsJSON(ctx, client, modelID, deploymentID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	if stats == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Trained model deployment "%s" not found, removing from state`, deploymentID))
		return state, false, diags
	}

	if stats.DeploymentStats == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Trained model deployment "%s" has no deployment stats, removing from state`, deploymentID))
		return state, false, diags
	}

	// Populate computed attributes
	state.ModelID = types.StringValue(stats.ModelId)
	state.DeploymentID = types.StringValue(stats.DeploymentStats.DeploymentId)
	if stats.DeploymentStats.State != nil {
		state.State = types.StringValue(stats.DeploymentStats.State.String())
	} else {
		state.State = types.StringNull()
	}
	if stats.DeploymentStats.AllocationStatus != nil {
		state.AllocationStatus = types.StringValue(stats.DeploymentStats.AllocationStatus.State.String())
	} else {
		state.AllocationStatus = types.StringNull()
	}
	state.StatsJSON = types.StringValue(statsJSON)

	// Update number_of_allocations from API only when adaptive_allocations is NOT configured
	if len(state.AdaptiveAllocations) == 0 {
		if stats.DeploymentStats.NumberOfAllocations != nil {
			state.NumberOfAllocations = types.Int64Value(int64(*stats.DeploymentStats.NumberOfAllocations))
		} else {
			state.NumberOfAllocations = types.Int64Null()
		}
	}

	// Set defaults for computed attributes if not already set (e.g. during import)
	if state.ForceStop.IsNull() {
		state.ForceStop = types.BoolValue(false)
	}
	if state.WaitFor.IsNull() {
		state.WaitFor = types.StringValue("fully_allocated")
	}

	return state, true, diags
}
