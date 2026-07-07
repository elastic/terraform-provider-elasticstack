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

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetTrainedModel retrieves a single trained model by ID.
// Returns (config, found, diagnostics). When the model is not found, returns (nil, false, nil).
func GetTrainedModel(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, modelID string) (*types.TrainedModelConfig, bool, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, err := typedClient.Ml.GetTrainedModels().ModelId(modelID).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, false, nil
		}
		return nil, false, diagutil.FrameworkDiagFromError(err)
	}

	if len(res.TrainedModelConfigs) == 0 {
		return nil, false, nil
	}

	return &res.TrainedModelConfigs[0], true, nil
}
