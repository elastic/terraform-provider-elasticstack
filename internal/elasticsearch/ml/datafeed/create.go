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

package datafeed

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *datafeedResource) create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var plan Datafeed
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	datafeedID := plan.DatafeedID.ValueString()
	if datafeedID == "" {
		resp.Diagnostics.AddError("Invalid Configuration", "datafeed_id cannot be empty")
		return
	}

	// Convert to API create model
	createRequest, diags := plan.toAPICreateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createDiags := elasticsearch.PutDatafeed(ctx, r.client, datafeedID, *createRequest)
	resp.Diagnostics.Append(createDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the created datafeed to get the full state.
	compID, sdkDiags := r.client.ID(ctx, datafeedID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(compID.String())
	found, readDiags := r.read(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read created datafeed", "Datafeed not found after creation")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
