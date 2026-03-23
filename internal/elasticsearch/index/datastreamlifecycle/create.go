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

package datastreamlifecycle

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.create(ctx, req.Plan, &resp.State)...)
}

func (r Resource) create(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var planModel tfModel
	diags := plan.Get(ctx, &planModel)
	if diags.HasError() {
		return diags
	}

	client, d := clients.MaybeNewAPIClientFromFrameworkResource(ctx, planModel.ElasticsearchConnection, r.client)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	name := planModel.Name.ValueString()
	id, sdkDiags := client.ID(ctx, name)
	if sdkDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return diags
	}

	planModel.ID = types.StringValue(id.String())

	apiModel, d := planModel.toAPIModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	diags.Append(elasticsearch.PutDataStreamLifecycle(ctx, client, name, planModel.ExpandWildcards.ValueString(), apiModel)...)
	if diags.HasError() {
		return diags
	}

	finalModel, d := r.read(ctx, client, planModel)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	diags.Append(state.Set(ctx, finalModel)...)
	return diags
}
