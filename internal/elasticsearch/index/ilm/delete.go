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

package ilm

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tfModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID, diags := clients.CompositeIDFromStrFw(state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetElasticsearchClient(ctx, state.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.ForceDestroy.IsNull() && state.ForceDestroy.ValueBool() {
		indices, diags := elasticsearch.GetIndicesWithILMPolicy(ctx, client, compID.ResourceID)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(indices) > 0 {
			resp.Diagnostics.Append(elasticsearch.ClearILMPolicyFromIndices(ctx, client, indices)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	resp.Diagnostics.Append(elasticsearch.DeleteIlm(ctx, client, compID.ResourceID)...)
}
