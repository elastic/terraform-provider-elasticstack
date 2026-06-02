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

package acctest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/trainedmodeltype"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// DefaultPyTorchModelID is a built-in PyTorch model available in 9.x stacks.
const DefaultPyTorchModelID = ".elser_model_2"

// EnsurePyTorchModelDeployment ensures the given PyTorch model can be deployed
// and starts a transient deployment to verify cluster capacity.
// It returns the model ID. If the model is not available or cannot be deployed
// the test is skipped.
// This helper does NOT leave a deployment running; it cleans up immediately.
func EnsurePyTorchModelDeployment(t *testing.T, modelID string) string {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("failed to create elasticsearch client: %v", err)
	}

	es := client.GetESClient()
	ctx := context.Background()

	// Verify the model exists and is fully defined.
	modelRes, err := es.Ml.GetTrainedModels().ModelId(modelID).Do(ctx)
	if err != nil {
		t.Skipf("skipping test: model %q is not available: %v", modelID, err)
	}
	if len(modelRes.TrainedModelConfigs) == 0 {
		t.Skipf("skipping test: model %q not found", modelID)
	}
	modelType := modelRes.TrainedModelConfigs[0].ModelType
	if modelType == nil || *modelType != trainedmodeltype.Pytorch {
		t.Skipf("skipping test: model %q is not a PyTorch model", modelID)
	}

	// Stop any stale deployment from a previous interrupted test.
	_, _ = es.Ml.StopTrainedModelDeployment(modelID).Do(ctx)
	// Allow the stop to take effect.
	time.Sleep(2 * time.Second)

	// Attempt a transient start to verify capacity.
	_, startErr := es.Ml.StartTrainedModelDeployment(modelID).
		Timeout("30s").
		Do(ctx)
	if startErr != nil {
		errStr := strings.ToLower(startErr.Error())
		if strings.Contains(errStr, "already exist") {
			// Deployment was not stopped in time; still means it's deployable.
			_, _ = es.Ml.StopTrainedModelDeployment(modelID).Do(ctx)
			return modelID
		}
		if strings.Contains(errStr, "429") ||
			strings.Contains(errStr, "too_many_requests") ||
			strings.Contains(errStr, "no ml nodes") ||
			strings.Contains(errStr, "insufficient memory") ||
			strings.Contains(errStr, "insufficient capacity") ||
			strings.Contains(errStr, "status_exception") {
			t.Skipf("skipping test: ML cluster cannot deploy model %q: %v", modelID, startErr)
		}
		t.Fatalf("failed to start deployment for model %q: %v", modelID, startErr)
	}

	// Immediately stop the transient deployment so the real Terraform test
	// can manage the deployment lifecycle.
	_, stopErr := es.Ml.StopTrainedModelDeployment(modelID).Do(ctx)
	if stopErr != nil {
		tflog.Warn(ctx, "failed to stop transient PyTorch model deployment", map[string]any{
			"model_id": modelID,
			"error":    stopErr.Error(),
		})
	} else {
		tflog.Info(ctx, "stopped transient PyTorch model deployment", map[string]any{
			"model_id": modelID,
		})
	}
	// Wait briefly for the stop to fully take effect.
	time.Sleep(2 * time.Second)

	return modelID
}
