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

func (r *datafeedResource) update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var plan Datafeed
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Datafeed
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	datafeedID := plan.DatafeedID.ValueString()
	if datafeedID == "" {
		resp.Diagnostics.AddError("Invalid Configuration", "datafeed_id cannot be empty")
		return
	}

	// Convert to API update model
	updateRequest, diags := plan.toAPIUpdateModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	needsRestart, diags := r.maybeStopDatafeed(ctx, datafeedID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the datafeed
	updateDiags := elasticsearch.UpdateDatafeed(ctx, r.client, datafeedID, *updateRequest)
	resp.Diagnostics.Append(updateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restart the datafeed if it was running
	if needsRestart {
		startDiags := elasticsearch.StartDatafeed(ctx, r.client, datafeedID, "", "", 0)
		resp.Diagnostics.Append(startDiags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Wait for the datafeed to reach started state
		_, waitDiags := WaitForDatafeedState(ctx, r.client, datafeedID, StateStarted)
		resp.Diagnostics.Append(waitDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Read the updated datafeed to get the full state
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
		resp.Diagnostics.AddError("Failed to read updated datafeed", "Datafeed not found after update")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
