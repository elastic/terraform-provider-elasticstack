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

package index

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, planModel.ElasticsearchConnection)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	apiModel, diags := planModel.toAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverFlavor, sdkDiags := client.ServerFlavor(ctx)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	params := planModel.toPutIndexParams(serverFlavor)

	concreteName, diags := elasticsearch.PutIndex(ctx, client, &apiModel, &params)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Compute id from the cluster UUID and the concrete index name returned by
	// Elasticsearch.  For static names concreteName equals the configured name.
	// For date math names concreteName is the resolved index (e.g. logs-2024.01.15).
	id, sdkDiags := client.ID(ctx, concreteName)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	planModel.ID = types.StringValue(id.String())
	planModel.ConcreteName = types.StringValue(concreteName)

	finalModel, diags := readIndex(ctx, planModel, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, finalModel)...)
}
