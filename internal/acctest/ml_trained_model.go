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
	"sync"
	"testing"

	"github.com/elastic/go-elasticsearch/v9/typedapi/ml/puttrainedmodel"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/trainedmodeltype"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
)

// AccTestTrainedModelID is the stable ID for the first acceptance-test bootstrap trained model.
const AccTestTrainedModelID = "terraform-acc-test-model"

// AccTestTrainedModelID2 is the stable ID for the second acceptance-test bootstrap trained model.
const AccTestTrainedModelID2 = "terraform-acc-test-model-2"

var ensureTrainedModelOnces sync.Map // key: modelID, value: *sync.Once

// EnsureTrainedModelByID creates a minimal trained model with the given ID in the
// acceptance-test cluster if it does not already exist. The model is a tiny
// single-leaf tree_ensemble that uses negligible memory and requires no deployment.
// Creation is executed once per model ID across the test process.
func EnsureTrainedModelByID(t *testing.T, modelID string) {
	t.Helper()

	once, _ := ensureTrainedModelOnces.LoadOrStore(modelID, &sync.Once{})
	once.(*sync.Once).Do(func() {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			t.Fatal(err)
		}

		_, found, _ := esclient.GetTrainedModel(ctx, client, modelID)
		if found {
			return
		}

		typedClient := client.GetESClient()

		leafValue := estypes.Float64(1.0)
		targetType := "regression"
		description := "Terraform acceptance test trained model"
		modelType := trainedmodeltype.Treeensemble

		_, err = typedClient.Ml.PutTrainedModel(modelID).
			Request(&puttrainedmodel.Request{
				Description: &description,
				ModelType:   &modelType,
				Input: &estypes.Input{
					FieldNames: []string{"foo"},
				},
				InferenceConfig: &estypes.InferenceConfigCreateContainer{
					Regression: &estypes.RegressionInferenceOptions{},
				},
				Definition: &estypes.Definition{
					TrainedModel: estypes.TrainedModel{
						Ensemble: &estypes.Ensemble{
							TargetType: &targetType,
							TrainedModels: []estypes.TrainedModel{
								{
									Tree: &estypes.TrainedModelTree{
										FeatureNames: []string{"foo"},
										TargetType:   &targetType,
										TreeStructure: []estypes.TrainedModelTreeNode{
											{
												NodeIndex: 0,
												LeafValue: &leafValue,
											},
										},
									},
								},
							},
						},
					},
				},
			}).
			Do(ctx)
		if err != nil {
			t.Fatalf("failed to create acceptance test trained model %q: %v", modelID, err)
		}
	})
}

// EnsureTrainedModel creates the first acceptance-test trained model if it does not exist.
// This is a convenience wrapper around EnsureTrainedModelByID.
func EnsureTrainedModel(t *testing.T) {
	t.Helper()
	EnsureTrainedModelByID(t, AccTestTrainedModelID)
}
