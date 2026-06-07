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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func createTrainedModelDeployment(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[TrainedModelDeploymentData],
) (entitycore.WriteResult[TrainedModelDeploymentData], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	modelID := plan.ModelID.ValueString()
	deploymentID := req.WriteID
	if deploymentID == "" {
		deploymentID = modelID
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating trained model deployment: model_id=%s deployment_id=%s", modelID, deploymentID))

	opts := elasticsearch.StartTrainedModelDeploymentOptions{
		DeploymentID:        &deploymentID,
		Priority:            typeutils.OptionalString(plan.Priority),
		WaitFor:             typeutils.OptionalString(plan.WaitFor),
		AdaptiveAllocations: toAdaptiveAllocationsSettings(plan.AdaptiveAllocations),
	}

	if !plan.NumberOfAllocations.IsUnknown() && !plan.NumberOfAllocations.IsNull() {
		opts.NumberOfAllocations = typeutils.OptionalInt(plan.NumberOfAllocations)
	}
	if !plan.ThreadsPerAllocation.IsUnknown() && !plan.ThreadsPerAllocation.IsNull() {
		opts.ThreadsPerAllocation = typeutils.OptionalInt(plan.ThreadsPerAllocation)
	}
	if !plan.QueueCapacity.IsUnknown() && !plan.QueueCapacity.IsNull() {
		opts.QueueCapacity = typeutils.OptionalInt(plan.QueueCapacity)
	}
	if !plan.APITimeout.IsUnknown() && !plan.APITimeout.IsNull() {
		v := plan.APITimeout.ValueString()
		opts.Timeout = &v
	}

	_, startDiags := elasticsearch.StartTrainedModelDeployment(ctx, client, modelID, opts)
	diags.Append(startDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[TrainedModelDeploymentData]{Model: plan}, diags
	}

	compID, idDiags := client.ID(ctx, deploymentID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[TrainedModelDeploymentData]{Model: plan}, diags
	}

	plan.ID = types.StringValue(compID.String())
	plan.DeploymentID = types.StringValue(deploymentID)

	return entitycore.WriteResult[TrainedModelDeploymentData]{Model: plan}, diags
}
