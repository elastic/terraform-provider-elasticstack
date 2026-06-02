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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *trainedModelDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TrainedModelDeploymentData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	adaptiveAllocations := toAdaptiveAllocationsSettings(plan.AdaptiveAllocations)

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

	statsJSON, stats, readDiags := elasticsearch.GetTrainedModelStatsJSON(ctx, client, modelID, deploymentID)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	populateComputedFromStats(&plan, stats, statsJSON)

	tflog.Info(ctx, fmt.Sprintf("Trained model deployment %s updated successfully", deploymentID))

	persistDiags := state.Set(ctx, &plan)
	diags.Append(persistDiags...)
	return diags
}
