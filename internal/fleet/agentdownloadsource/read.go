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

package agentdownloadsource

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}
	resp.Diagnostics.Append(r.assertVersionSupported(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceID := state.SourceID.ValueString()

	// Read the existing spaces from state to determine where to query.
	spaceID, diags := fleetutils.GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readState, found, diags := r.readAndHydrateState(ctx, client, sourceID, spaceID, state.SpaceIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		// Resource no longer exists.
		resp.State.RemoveResource(ctx)
		return
	}
	diags = resp.State.Set(ctx, readState)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) readAndHydrateState(
	ctx context.Context,
	client *fleet.Client,
	sourceID string,
	spaceID string,
	preservedSpaceIDs types.Set,
) (model, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	var state model

	getResp, readDiags := fleet.GetAgentDownloadSource(ctx, client, sourceID, spaceID)
	diags.Append(readDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	if getResp == nil || getResp.JSON200 == nil {
		// Ensure we also exercise the list endpoint before deciding the resource is gone.
		listResp, listDiags := fleet.ListAgentDownloadSources(ctx, client, spaceID)
		diags.Append(listDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		if listResp != nil && listResp.JSON200 != nil {
			for _, item := range listResp.JSON200.Items {
				if item.Id == sourceID {
					diags.AddError("Unexpected API response", "Read by source_id returned not found but list endpoint still includes this source.")
					return state, false, diags
				}
			}
		}
		return state, false, diags
	}

	item := getResp.JSON200.Item
	state.ID = types.StringValue(item.Id)
	state.SourceID = types.StringValue(item.Id)
	state.Name = types.StringValue(item.Name)
	state.Host = types.StringValue(item.Host)
	state.Default = types.BoolPointerValue(item.IsDefault)
	if item.ProxyId != nil {
		state.ProxyID = types.StringValue(*item.ProxyId)
	} else {
		state.ProxyID = types.StringNull()
	}
	state.SpaceIDs = preservedSpaceIDs
	return state, true, diags
}
