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

package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/ml/starttrainedmodeldeployment"
	"github.com/elastic/go-elasticsearch/v9/typedapi/ml/updatetrainedmodeldeployment"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/deploymentallocationstate"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/trainingpriority"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// StartTrainedModelDeploymentOptions holds optional parameters for starting a trained model deployment.
type StartTrainedModelDeploymentOptions struct {
	DeploymentID         *string
	NumberOfAllocations  *int
	ThreadsPerAllocation *int
	Priority             *string
	QueueCapacity        *int
	WaitFor              *string
	Timeout              *string
	AdaptiveAllocations  *types.AdaptiveAllocationsSettings
}

// StartTrainedModelDeployment starts a trained model deployment.
func StartTrainedModelDeployment(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	modelID string,
	opts StartTrainedModelDeploymentOptions,
) (*starttrainedmodeldeployment.Response, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()
	req := typedClient.Ml.StartTrainedModelDeployment(modelID)

	if opts.DeploymentID != nil {
		req.DeploymentId(*opts.DeploymentID)
	}
	if opts.NumberOfAllocations != nil {
		req.NumberOfAllocations(*opts.NumberOfAllocations)
	}
	if opts.ThreadsPerAllocation != nil {
		req.ThreadsPerAllocation(*opts.ThreadsPerAllocation)
	}
	if opts.Priority != nil {
		var priority trainingpriority.TrainingPriority
		switch strings.ToLower(*opts.Priority) {
		case "low":
			priority = trainingpriority.Low
		case "normal":
			priority = trainingpriority.Normal
		default:
			diags.AddError("Invalid priority", fmt.Sprintf("Invalid priority %q. Valid values are 'low' and 'normal'.", *opts.Priority))
			return nil, diags
		}
		req.Priority(priority)
	}
	if opts.QueueCapacity != nil {
		req.QueueCapacity(*opts.QueueCapacity)
	}
	if opts.WaitFor != nil {
		var waitFor deploymentallocationstate.DeploymentAllocationState
		switch strings.ToLower(*opts.WaitFor) {
		case "starting":
			waitFor = deploymentallocationstate.Starting
		case "started":
			waitFor = deploymentallocationstate.Started
		case "fully_allocated":
			waitFor = deploymentallocationstate.Fullyallocated
		default:
			diags.AddError("Invalid wait_for", fmt.Sprintf("Invalid wait_for %q. Valid values are 'starting', 'started', and 'fully_allocated'.", *opts.WaitFor))
			return nil, diags
		}
		req.WaitFor(waitFor)
	}
	if opts.Timeout != nil {
		req.Timeout(*opts.Timeout)
	}
	if opts.AdaptiveAllocations != nil {
		req.AdaptiveAllocations(opts.AdaptiveAllocations)
	}

	res, err := req.Do(ctx)
	if err != nil {
		diags.AddError("Failed to start trained model deployment", fmt.Sprintf("Unable to start trained model deployment: %s — %s", modelID, err.Error()))
		return nil, diags
	}

	return res, diags
}

// GetTrainedModelStats retrieves the stats for a specific trained model deployment,
// filtering the response to the deployment matching the given deployment_id.
func GetTrainedModelStats(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, modelID string, deploymentID string) (*types.TrainedModelStats, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	res, err := typedClient.Ml.GetTrainedModelsStats().ModelId(modelID).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, diags
		}
		diags.AddError("Failed to get trained model stats", fmt.Sprintf("Unable to get trained model stats: %s — %s", modelID, err.Error()))
		return nil, diags
	}

	for i := range res.TrainedModelStats {
		if res.TrainedModelStats[i].DeploymentStats != nil && res.TrainedModelStats[i].DeploymentStats.DeploymentId == deploymentID {
			return &res.TrainedModelStats[i], diags
		}
	}

	return nil, diags
}

// UpdateTrainedModelDeploymentOptions holds optional parameters for updating a trained model deployment.
type UpdateTrainedModelDeploymentOptions struct {
	NumberOfAllocations *int
	AdaptiveAllocations *types.AdaptiveAllocationsSettings
}

// UpdateTrainedModelDeployment updates a trained model deployment.
func UpdateTrainedModelDeployment(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, deploymentID string, opts UpdateTrainedModelDeploymentOptions) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()
	req := typedClient.Ml.UpdateTrainedModelDeployment(deploymentID)

	updateReq := updatetrainedmodeldeployment.NewRequest()
	if opts.NumberOfAllocations != nil {
		updateReq.NumberOfAllocations = opts.NumberOfAllocations
	}
	if opts.AdaptiveAllocations != nil {
		updateReq.AdaptiveAllocations = opts.AdaptiveAllocations
	}

	_, err := req.Request(updateReq).Do(ctx)
	if err != nil {
		diags.AddError("Failed to update trained model deployment", fmt.Sprintf("Unable to update trained model deployment: %s — %s", deploymentID, err.Error()))
		return diags
	}

	return diags
}

// StopTrainedModelDeployment stops a trained model deployment. Treats HTTP 404 as success.
func StopTrainedModelDeployment(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, deploymentID string, force bool) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()
	req := typedClient.Ml.StopTrainedModelDeployment(deploymentID)
	if force {
		req.Force(force)
	}

	_, err := req.Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Failed to stop trained model deployment", fmt.Sprintf("Unable to stop trained model deployment: %s — %s", deploymentID, err.Error()))
		return diags
	}

	return diags
}

// GetTrainedModelStatsJSON retrieves the raw JSON stats for a specific trained model deployment.
func GetTrainedModelStatsJSON(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, modelID string, deploymentID string) (string, *types.TrainedModelStats, fwdiag.Diagnostics) {
	stats, diags := GetTrainedModelStats(ctx, apiClient, modelID, deploymentID)
	if diags.HasError() || stats == nil {
		return "", stats, diags
	}

	statsJSON, marshalErr := json.Marshal(stats)
	if marshalErr != nil {
		diags.AddError("Failed to marshal trained model stats", marshalErr.Error())
		return "", stats, diags
	}

	return string(statsJSON), stats, diags
}
