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
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *aliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel tfModel

	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasName := stateModel.Name.ValueString()

	// Read the alias and update the model
	diags := readAliasIntoModel(ctx, r.client, aliasName, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the alias was found
	if stateModel.WriteIndex.IsNull() && stateModel.ReadIndices.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, stateModel)...)
}

// readAliasIntoModel reads an alias from Elasticsearch and populates the provided model
func readAliasIntoModel(ctx context.Context, client *clients.APIClient, aliasName string, model *tfModel) diag.Diagnostics {
	// Get the alias
	indices, diags := elasticsearch.GetAlias(ctx, client, aliasName)
	if diags.HasError() {
		return diags
	}

	// If no indices returned, the alias doesn't exist
	if len(indices) == 0 {
		// Set both to null to indicate the alias doesn't exist
		model.WriteIndex = types.ObjectNull(getIndexAttrTypes())
		model.ReadIndices = types.SetNull(types.ObjectType{AttrTypes: getIndexAttrTypes()})
		return nil
	}

	// Extract alias data from the response
	aliasData := make(map[string]models.IndexAlias)
	for indexName, index := range indices {
		if alias, exists := index.Aliases[aliasName]; exists {
			aliasData[indexName] = alias
		}
	}

	if len(aliasData) == 0 {
		// Set both to null to indicate the alias doesn't exist
		model.WriteIndex = types.ObjectNull(getIndexAttrTypes())
		model.ReadIndices = types.SetNull(types.ObjectType{AttrTypes: getIndexAttrTypes()})
		return nil
	}

	// Update the model with API data
	return model.populateFromAPI(ctx, aliasName, aliasData)
}
