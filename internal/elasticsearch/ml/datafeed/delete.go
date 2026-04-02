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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *datafeedResource) delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.resourceReady(&resp.Diagnostics) {
		return
	}

	var state Datafeed
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	datafeedID := state.DatafeedID.ValueString()
	if datafeedID == "" {
		resp.Diagnostics.AddError("Invalid Configuration", "datafeed_id cannot be empty")
		return
	}

	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, state.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Before deleting, we need to stop the datafeed if it's running
	_, stopDiags := r.maybeStopDatafeed(ctx, state, r.client, datafeedID)
	resp.Diagnostics.Append(stopDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the datafeed
	deleteDiags := elasticsearch.DeleteDatafeed(ctx, client, datafeedID, false)
	resp.Diagnostics.Append(deleteDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The resource is automatically removed from state on successful delete
}

func (r *datafeedResource) maybeStopDatafeed(
	ctx context.Context,
	model Datafeed,
	defaultClient *clients.APIClient,
	datafeedID string,
) (bool, diag.Diagnostics) {
	client, diags := clients.MaybeNewAPIClientFromFrameworkResource(ctx, model.ElasticsearchConnection, defaultClient)
	if diags.HasError() {
		return false, diags
	}

	// Check current state
	currentState, diags := GetDatafeedState(ctx, model, defaultClient, datafeedID)
	if diags.HasError() {
		return false, diags
	}

	if currentState == nil {
		return false, nil
	}

	// If the datafeed is not running, nothing to stop
	if *currentState != StateStarted && *currentState != StateStarting {
		return false, diags
	}

	// Stop the datafeed
	stopDiags := elasticsearch.StopDatafeed(ctx, client, datafeedID, false, 0)
	diags.Append(stopDiags...)
	if diags.HasError() {
		return true, diags
	}

	// Wait for the datafeed to reach stopped state
	_, waitDiags := WaitForDatafeedState(ctx, model, defaultClient, datafeedID, StateStopped)
	diags.Append(waitDiags...)
	if diags.HasError() {
		return true, diags
	}

	return true, diags
}
