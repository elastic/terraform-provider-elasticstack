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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r *aliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel tfModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(planModel.Validate(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasName := planModel.Name.ValueString()

	// Set the ID using client.ID
	id, sdkDiags := r.client.ID(ctx, aliasName)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}
	planModel.ID = basetypes.NewStringValue(id.String())

	// Get alias configurations from the plan
	configs, diags := planModel.toAliasConfigs(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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
	resp.Diagnostics.Append(elasticsearch.UpdateAliasesAtomic(ctx, r.client, actions)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read back the alias to ensure state consistency, updating the current model
	diags = readAliasIntoModel(ctx, r.client, aliasName, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, planModel)...)
}
