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

package indices

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, config tfModel) (tfModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Default to "*" (all indices) when target is null or empty.
	target := config.Target.ValueString()
	if target == "" {
		target = "*"
	}

	// Call client API
	indexAPIModels, idxDiags := elasticsearch.GetIndices(ctx, esClient, target)
	diags.Append(idxDiags...)
	if diags.HasError() {
		return config, diags
	}

	// Map response body to model
	indices := []indexTfModel{}
	for indexName, indexAPIModel := range indexAPIModels {
		indexStateModel := indexTfModel{}

		pDiags := indexStateModel.populateFromAPI(ctx, indexName, indexAPIModel)
		diags.Append(pDiags...)
		if diags.HasError() {
			return config, diags
		}

		indices = append(indices, indexStateModel)
	}

	indicesList, listDiags := types.ListValueFrom(ctx, indicesElementType(ctx), indices)
	diags.Append(listDiags...)
	if diags.HasError() {
		return config, diags
	}

	config.ID = types.StringValue(target)
	config.Indices = indicesList

	return config, diags
}
