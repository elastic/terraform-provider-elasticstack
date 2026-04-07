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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model

	diags := req.Plan.Get(ctx, &plan)
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

	// Determine space from plan (first space_ids entry) for CREATE.
	var spaceID string
	if !plan.SpaceIDs.IsNull() && !plan.SpaceIDs.IsUnknown() {
		var tempDiags diag.Diagnostics
		spaceIDs := typeutils.SetTypeAs[types.String](ctx, plan.SpaceIDs, path.Root("space_ids"), &tempDiags)
		resp.Diagnostics.Append(tempDiags...)
		if !tempDiags.HasError() && len(spaceIDs) > 0 {
			spaceID = spaceIDs[0].ValueString()
		}
	}

	body := plan.toAPICreateModel(ctx)

	createResp, diags := fleet.CreateAgentDownloadSource(ctx, client, spaceID, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if createResp.JSON200 == nil {
		resp.Diagnostics.AddError("Unexpected API response", "Create agent download source response missing JSON200 body")
		return
	}

	item := createResp.JSON200.Item

	// Ensure we keep the operational space information consistent with how Read/Update/Delete will resolve it.
	if plan.SpaceIDs.IsUnknown() {
		plan.SpaceIDs = types.SetNull(types.StringType)
	}

	readState, found, diags := r.readAndHydrateState(ctx, client, item.Id, spaceID, plan.SpaceIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Unexpected API response", "Created agent download source could not be read back by source_id")
		return
	}

	diags = resp.State.Set(ctx, readState)
	resp.Diagnostics.Append(diags...)
}
