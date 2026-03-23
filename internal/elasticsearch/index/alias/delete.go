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
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *aliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var stateModel tfModel

	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasName := stateModel.Name.ValueString()

	// Get current configuration from state
	currentConfigs, diags := stateModel.toAliasConfigs(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build remove actions for all indices
	var actions []elasticsearch.AliasAction
	for _, config := range currentConfigs {
		actions = append(actions, elasticsearch.AliasAction{
			Type:  "remove",
			Index: config.Name,
			Alias: aliasName,
		})
	}

	// Remove the alias from all indices
	if len(actions) > 0 {
		resp.Diagnostics.Append(elasticsearch.UpdateAliasesAtomic(ctx, r.client, actions)...)
	}
}
