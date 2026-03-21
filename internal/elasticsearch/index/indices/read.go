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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read refreshes the Terraform state with the latest data.
func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var stateModel tfModel

	diags := req.Config.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve target attribute — use types.String to handle the null case
	// (when the user omits the optional "target" attribute).
	var targetAttr types.String
	diags = req.Config.GetAttribute(ctx, path.Root("target"), &targetAttr)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, stateModel.ElasticsearchConnection, &d.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default to "*" (all indices) when target is null or empty.
	target := targetAttr.ValueString()
	if target == "" {
		target = "*"
	}

	// Call client API
	indexAPIModels, diags := elasticsearch.GetIndices(ctx, client, target)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to model
	indices := []indexTfModel{}
	for indexName, indexAPIModel := range indexAPIModels {
		indexStateModel := indexTfModel{}

		diags := indexStateModel.populateFromAPI(ctx, indexName, indexAPIModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		indices = append(indices, indexStateModel)
	}

	indicesList, diags := types.ListValueFrom(ctx, indicesElementType(), indices)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateModel.ID = types.StringValue(target)
	stateModel.Target = targetAttr
	stateModel.Indices = indicesList

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, stateModel)...)
}
