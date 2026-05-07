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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func createAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, plan tfModel) (tfModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	diags.Append(plan.Validate(ctx)...)
	if diags.HasError() {
		return plan, diags
	}

	aliasName := resourceID

	// Set the ID using client.ID
	id, sdkDiags := client.ID(ctx, aliasName)
	if sdkDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return plan, diags
	}
	plan.ID = basetypes.NewStringValue(id.String())

	// Get alias configurations from the plan
	configs, configDiags := plan.toAliasConfigs(ctx)
	diags.Append(configDiags...)
	if diags.HasError() {
		return plan, diags
	}

	// Convert to alias actions
	var actions []elasticsearch.AliasAction
	for _, config := range configs {
		action := elasticsearch.AliasAction{
			Type:          "add",
			Index:         config.Name,
			Alias:         aliasName,
			IsWriteIndex:  config.IsWriteIndex,
			Filter:        config.Filter,
			IndexRouting:  config.IndexRouting,
			IsHidden:      config.IsHidden,
			Routing:       config.Routing,
			SearchRouting: config.SearchRouting,
		}
		actions = append(actions, action)
	}

	// Create the alias atomically
	diags.Append(elasticsearch.UpdateAliasesAtomic(ctx, client, actions)...)
	if diags.HasError() {
		return plan, diags
	}

	return plan, diags
}
