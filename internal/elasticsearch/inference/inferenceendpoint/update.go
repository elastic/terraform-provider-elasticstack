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

package inferenceendpoint

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *inferenceEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var diags diag.Diagnostics

	var plan Data
	diags.Append(req.Plan.Get(ctx, &plan)...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var state Data
	diags.Append(req.State.Get(ctx, &state)...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	update, modelDiags := plan.toUpdateModel(ctx)
	diags.Append(modelDiags...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	updateDiags := elasticsearch.UpdateInferenceEndpoint(ctx, r.client, update)
	diags.Append(updateDiags...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	plan.ID = state.ID

	readData, readDiags := r.read(ctx, plan)
	diags.Append(readDiags...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if readData == nil {
		diags.AddError("Not Found", fmt.Sprintf("Inference endpoint %q was not found after update", plan.InferenceID.ValueString()))
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}
