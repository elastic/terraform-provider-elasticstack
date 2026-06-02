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
	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const createTimeoutErrorMessage = "Timed out while waiting for trained model deployment to reach desired state. The deployment may still be starting in the background. Timeout: %s"

func (r *trainedModelDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TrainedModelDeploymentData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get create timeout
	createTimeout, fwDiags := data.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	diags = r.create(ctx, req, &resp.State)
	if diagutil.ContainsContextDeadlineExceeded(ctx, diags) {
		diags.AddError("Operation timed out", fmt.Sprintf(createTimeoutErrorMessage, createTimeout))
	}

	resp.Diagnostics.Append(diags...)
}

func (r *trainedModelDeploymentResource) create(ctx context.Context, req resource.CreateRequest, state *tfsdk.State) diag.Diagnostics {
	var data TrainedModelDeploymentData
	diags := req.Plan.Get(ctx, &data)
	if diags.HasError() {
		return diags
	}

	client, fwDiags := r.Client().GetElasticsearchClient(ctx, data.ElasticsearchConnection)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return diags
	}

	modelID := data.ModelID.ValueString()
	deploymentID := data.DeploymentID.ValueString()
	if deploymentID == "" {
		deploymentID = modelID
	}

	// Build start options
	var adaptiveAllocations *types.AdaptiveAllocationsSettings
	if data.AdaptiveAllocations != nil && !data.AdaptiveAllocations.Enabled.IsNull() {
		aa := data.AdaptiveAllocations
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
	if !data.NumberOfAllocations.IsNull() {
		v := int(data.NumberOfAllocations.ValueInt64())
		numberOfAllocations = &v
	}

	var threadsPerAllocation *int
	if !data.ThreadsPerAllocation.IsNull() {
		v := int(data.ThreadsPerAllocation.ValueInt64())
		threadsPerAllocation = &v
	}

	var priority *string
	if !data.Priority.IsNull() {
		v := data.Priority.ValueString()
		priority = &v
	}

	var queueCapacity *int
	if !data.QueueCapacity.IsNull() {
		v := int(data.QueueCapacity.ValueInt64())
		queueCapacity = &v
	}

	var waitFor *string
	if !data.WaitFor.IsNull() {
		v := data.WaitFor.ValueString()
		waitFor = &v
	}

	var apiTimeout *string
	if !data.APITimeout.IsNull() {
		v := data.APITimeout.ValueString()
		apiTimeout = &v
	}

	startOpts := elasticsearch.StartTrainedModelDeploymentOptions{
		DeploymentID:         &deploymentID,
		NumberOfAllocations:  numberOfAllocations,
		ThreadsPerAllocation: threadsPerAllocation,
		Priority:             priority,
		QueueCapacity:        queueCapacity,
		WaitFor:              waitFor,
		Timeout:              apiTimeout,
		AdaptiveAllocations:  adaptiveAllocations,
	}

	_, startDiags := elasticsearch.StartTrainedModelDeployment(ctx, client, modelID, startOpts)
	diags.Append(startDiags...)
	if diags.HasError() {
		return diags
	}

	// Poll until deployment reaches desired allocation status
	pollErr := r.waitForDeploymentAllocationStatus(ctx, client, modelID, deploymentID, data.WaitFor.ValueString())
	if pollErr != nil {
		diags.AddError("Failed to wait for deployment allocation", pollErr.Error())
		return diags
	}

	// Read the deployment state to populate computed attributes
	statsJSON, stats, readDiags := elasticsearch.GetTrainedModelStatsJSON(ctx, client, modelID, deploymentID)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	if stats == nil || stats.DeploymentStats == nil {
		diags.AddError("Deployment not found after start", fmt.Sprintf("Trained model deployment %s not found after starting", deploymentID))
		return diags
	}

	// Set the composite ID
	compID, idDiags := client.ID(ctx, deploymentID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return diags
	}

	data.ID = fwtypes.StringValue(compID.String())
	data.DeploymentID = fwtypes.StringValue(deploymentID)
	if stats.DeploymentStats.State != nil {
		data.State = fwtypes.StringValue(stats.DeploymentStats.State.String())
	} else {
		data.State = fwtypes.StringNull()
	}
	if stats.DeploymentStats.AllocationStatus != nil {
		data.AllocationStatus = fwtypes.StringValue(stats.DeploymentStats.AllocationStatus.State.String())
	} else {
		data.AllocationStatus = fwtypes.StringNull()
	}
	data.StatsJSON = fwtypes.StringValue(statsJSON)

	// Update number_of_allocations from API only when adaptive_allocations is NOT configured
	if data.AdaptiveAllocations == nil || data.AdaptiveAllocations.Enabled.IsNull() {
		if stats.DeploymentStats.NumberOfAllocations != nil {
			data.NumberOfAllocations = fwtypes.Int64Value(int64(*stats.DeploymentStats.NumberOfAllocations))
		} else {
			data.NumberOfAllocations = fwtypes.Int64Null()
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Trained model deployment %s started successfully", deploymentID))

	// Persist state
	persistDiags := state.Set(ctx, &data)
	diags.Append(persistDiags...)
	return diags
}

func (r *trainedModelDeploymentResource) waitForDeploymentAllocationStatus(ctx context.Context, client *clients.ElasticsearchScopedClient, modelID, deploymentID, desiredStatus string) error {
	if desiredStatus == "" {
		desiredStatus = "fully_allocated"
	}

	checkState := func(ctx context.Context) (bool, error) {
		stats, diags := elasticsearch.GetTrainedModelStats(ctx, client, modelID, deploymentID)
		if diags.HasError() {
			return false, diagutil.FwDiagsAsError(diags)
		}
		if stats == nil || stats.DeploymentStats == nil || stats.DeploymentStats.AllocationStatus == nil {
			return false, nil
		}
		return stats.DeploymentStats.AllocationStatus.State.String() == desiredStatus, nil
	}

	// Check immediately before entering poll loop
	alreadyReady, err := checkState(ctx)
	if err != nil || alreadyReady {
		return err
	}

	return asyncutils.WaitForStateTransition(ctx, "ml_trained_model_deployment", deploymentID, checkState)
}
