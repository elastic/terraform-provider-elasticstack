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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[tfModel]) (entitycore.WriteResult[tfModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan
	aliasName := req.WriteID

	diags.Append(plan.Validate(ctx)...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	// Read current alias state from API to find which indices to remove
	currentIndices, readDiags := elasticsearch.GetAlias(ctx, client, aliasName)
	diags.Append(readDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	// Build current index map from API response
	currentIndexMap := make(map[string]IndexConfig)
	for indexName, indexAliases := range currentIndices {
		if aliasDef, exists := indexAliases.Aliases[aliasName]; exists {
			config, configDiags := aliasDefinitionToConfig(indexName, aliasDef)
			diags.Append(configDiags...)
			if diags.HasError() {
				return entitycore.WriteResult[tfModel]{Model: plan}, diags
			}
			currentIndexMap[indexName] = config
		}
	}

	// Get planned configuration
	plannedConfigs, configDiags := plan.toAliasConfigs(ctx)
	diags.Append(configDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[tfModel]{Model: plan}, diags
	}

	plannedIndexMap := make(map[string]IndexConfig)
	for _, config := range plannedConfigs {
		plannedIndexMap[config.Name] = config
	}

	// Build atomic actions
	var actions []elasticsearch.AliasAction

	// Remove indices that are no longer in the plan
	for indexName := range currentIndexMap {
		if _, exists := plannedIndexMap[indexName]; !exists {
			actions = append(actions, elasticsearch.AliasAction{
				Type:  "remove",
				Index: indexName,
				Alias: aliasName,
			})
		}
	}

	// Add or update indices in the plan
	for _, config := range plannedConfigs {
		currentAlias, ok := currentIndexMap[config.Name]
		if ok && currentAlias.Equals(config) {
			continue
		}

		actions = append(actions, elasticsearch.AliasAction{
			Type:          "add",
			Index:         config.Name,
			Alias:         aliasName,
			IsWriteIndex:  config.IsWriteIndex,
			Filter:        config.Filter,
			IndexRouting:  config.IndexRouting,
			IsHidden:      config.IsHidden,
			Routing:       config.Routing,
			SearchRouting: config.SearchRouting,
		})
	}

	if len(actions) > 0 {
		diags.Append(elasticsearch.UpdateAliasesAtomic(ctx, client, actions)...)
		if diags.HasError() {
			return entitycore.WriteResult[tfModel]{Model: plan}, diags
		}
	}

	return entitycore.WriteResult[tfModel]{Model: plan}, diags
}
