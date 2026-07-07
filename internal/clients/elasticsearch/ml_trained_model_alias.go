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
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/ml/gettrainedmodels"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// PutMLTrainedModelAlias creates or updates an ML trained model alias.
func PutMLTrainedModelAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, modelID, alias string, reassign bool) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient := client.GetESClient()

	_, err := typedClient.Ml.PutTrainedModelAlias(modelID, alias).Reassign(reassign).Do(ctx)
	if err != nil {
		diags.AddError("Failed to create or update ML trained model alias", fmt.Sprintf("Unable to create or update ML trained model alias: %s -> %s — %s", alias, modelID, err.Error()))
		return diags
	}

	return diags
}

// GetMLTrainedModelAlias resolves an alias to its current model ID.
// It calls GetTrainedModels with the alias as the model_id parameter.
// Returns not-found when the result is empty or the API returns 404.
// Retries a few times to handle brief eventual-consistency windows.
func GetMLTrainedModelAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, alias string) (modelID string, found bool, diags diag.Diagnostics) {
	typedClient := client.GetESClient()

	var res *gettrainedmodels.Response
	var err error

	for attempt := range 3 {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", false, diags
			case <-time.After(500 * time.Millisecond):
			}
		}

		res, err = typedClient.Ml.GetTrainedModels().ModelId(alias).AllowNoMatch(true).Do(ctx)
		if err != nil {
			if IsNotFoundElasticsearchError(err) {
				return "", false, diags
			}
			diags.AddError("Failed to get ML trained model alias", fmt.Sprintf("Unable to get ML trained model alias: %s — %s", alias, err.Error()))
			return "", false, diags
		}

		if res != nil && len(res.TrainedModelConfigs) > 0 {
			return res.TrainedModelConfigs[0].ModelId, true, diags
		}
	}

	return "", false, diags
}

// DeleteMLTrainedModelAlias deletes an ML trained model alias.
// It first resolves the alias to its current model ID, then calls DELETE.
// Treats GET 404/empty and DELETE 404 as idempotent success.
func DeleteMLTrainedModelAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, alias string) diag.Diagnostics {
	var diags diag.Diagnostics

	modelID, found, getDiags := GetMLTrainedModelAlias(ctx, client, alias)
	diags.Append(getDiags...)
	if diags.HasError() {
		return diags
	}
	if !found {
		return diags
	}

	typedClient := client.GetESClient()

	_, err := typedClient.Ml.DeleteTrainedModelAlias(modelID, alias).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Failed to delete ML trained model alias", fmt.Sprintf("Unable to delete ML trained model alias: %s — %s", alias, err.Error()))
		return diags
	}

	return diags
}
