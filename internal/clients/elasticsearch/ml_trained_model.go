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

	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/gettrainedmodels"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetTrainedModel retrieves a single trained model by ID.
// Returns (config, found, diagnostics). When the model is not found, returns (nil, false, nil).
func GetTrainedModel(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, modelID string) (*types.TrainedModelConfig, bool, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, diags := CallOrNotFound(func() (*gettrainedmodels.Response, error) {
		return typedClient.Ml.GetTrainedModels().ModelId(modelID).Do(ctx)
	})
	if diags.HasError() || res == nil {
		return nil, false, diags
	}

	if len(res.TrainedModelConfigs) == 0 {
		return nil, false, nil
	}

	return &res.TrainedModelConfigs[0], true, nil
}
