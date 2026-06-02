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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func updateTrainedModelDeployment(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[TrainedModelDeploymentData],
) (entitycore.WriteResult[TrainedModelDeploymentData], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	deploymentID := req.WriteID
	modelID := plan.ModelID.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("Updating trained model deployment: model_id=%s deployment_id=%s", modelID, deploymentID))

	adaptiveAllocations := toAdaptiveAllocationsSettings(plan.AdaptiveAllocations)

	var numberOfAllocations *int
	if !plan.NumberOfAllocations.IsUnknown() && !plan.NumberOfAllocations.IsNull() {
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
		return entitycore.WriteResult[TrainedModelDeploymentData]{Model: plan}, diags
	}

	return entitycore.WriteResult[TrainedModelDeploymentData]{Model: plan}, diags
}
