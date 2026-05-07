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

package alias

import (
	"context"

	esTypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state tfModel) (tfModel, bool, diag.Diagnostics) {
	aliasName := resourceID

	indices, diags := elasticsearch.GetAlias(ctx, client, aliasName)
	if diags.HasError() {
		return state, false, diags
	}

	diags = readAliasIntoModel(ctx, aliasName, indices, &state)
	if diags.HasError() {
		return state, false, diags
	}

	// Check if the alias was found
	if state.WriteIndex.IsNull() && state.ReadIndices.IsNull() {
		return state, false, nil
	}

	return state, true, nil
}

// readAliasIntoModel populates the provided model from alias API response.
func readAliasIntoModel(ctx context.Context, aliasName string, indices map[string]esTypes.IndexAliases, model *tfModel) diag.Diagnostics {
	if len(indices) == 0 {
		model.WriteIndex = types.ObjectNull(getIndexAttrTypes(ctx))
		model.ReadIndices = types.SetNull(types.ObjectType{AttrTypes: getIndexAttrTypes(ctx)})
		return nil
	}

	aliasData := make(map[string]esTypes.AliasDefinition)
	for indexName, indexAliases := range indices {
		if alias, exists := indexAliases.Aliases[aliasName]; exists {
			aliasData[indexName] = alias
		}
	}

	if len(aliasData) == 0 {
		model.WriteIndex = types.ObjectNull(getIndexAttrTypes(ctx))
		model.ReadIndices = types.SetNull(types.ObjectType{AttrTypes: getIndexAttrTypes(ctx)})
		return nil
	}

	return model.populateFromAPI(ctx, aliasName, aliasData)
}
