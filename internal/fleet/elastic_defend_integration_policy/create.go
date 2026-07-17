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

package elasticdefendintegrationpolicy

import (
	"context"
	"fmt"

	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *elasticDefendIntegrationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel elasticDefendIntegrationPolicyModel

	diags := req.Plan.Get(ctx, &planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetKibanaClient(ctx, planModel.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !planModel.AgentPolicyIDs.IsNull() && !planModel.AgentPolicyIDs.IsUnknown() {
		supported, d := client.EnforceMinVersion(ctx, MinVersionPolicyIDs)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !supported {
			resp.Diagnostics.AddError(
				"Unsupported Elasticsearch version",
				fmt.Sprintf("agent_policy_ids requires Elastic Stack >= %s", MinVersionPolicyIDs.String()),
			)
			return
		}
	}

	fleetClient := client.GetFleetClient()

	// Determine space context for creating the package policy
	spaceID, d := fleetutils.SpaceIDFromSet(ctx, planModel.SpaceIDs)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 1: Bootstrap create using ENDPOINT_INTEGRATION_CONFIG input type
	bootstrapReq, d := buildBootstrapRequest(ctx, &planModel)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	bootstrapPolicy, d := fleetclient.CreateDefendPackagePolicy(ctx, fleetClient, spaceID, bootstrapReq)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Capture the server-managed private state from bootstrap response
	ps := extractPrivateStateFromResponse(bootstrapPolicy)

	// Save policy_id and private state immediately after bootstrap so the
	// resource can be recovered if finalize fails. Populate basic fields from
	// the bootstrap response to ensure no unknown values remain in state
	// (the framework rejects unknown values after apply).
	bootstrapID := bootstrapPolicy.Id
	planModel.PolicyID = types.StringValue(bootstrapID)
	if spaceID != "" {
		planModel.ID = types.StringValue(spaceID + "/" + bootstrapID)
	} else {
		planModel.ID = types.StringValue(bootstrapID)
	}
	// Normalize space_ids from bootstrap response to avoid unknown state values
	if bootstrapPolicy.SpaceIds != nil && len(*bootstrapPolicy.SpaceIds) > 0 {
		spaceIDs, d := types.SetValueFrom(ctx, types.StringType, *bootstrapPolicy.SpaceIds)
		resp.Diagnostics.Append(d...)
		planModel.SpaceIDs = spaceIDs
	} else if planModel.SpaceIDs.IsNull() || planModel.SpaceIDs.IsUnknown() {
		planModel.SpaceIDs = types.SetNull(types.StringType)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &planModel)...)
	resp.Diagnostics.Append(savePrivateState(ctx, resp.Private, ps)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 2: Finalize with the user-configured typed policy
	finalizeReq, d := buildFinalizeRequest(ctx, &planModel, nil, ps)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	// ID is passed as the URL path parameter to UpdateDefendPackagePolicy

	_, d = fleetclient.UpdateDefendPackagePolicy(ctx, fleetClient, bootstrapID, spaceID, finalizeReq)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The PUT response does not include spaceIds, so do a GET to retrieve the
	// full policy state (including spaceIds and the server-managed artifact_manifest).
	finalPolicy, d := fleetclient.GetDefendPackagePolicy(ctx, fleetClient, bootstrapID, spaceID)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if finalPolicy == nil {
		resp.Diagnostics.AddError(
			"Defend policy not found after create",
			"The policy was created but could not be retrieved (HTTP 404). This is unexpected; the policy may have been deleted externally.",
		)
		return
	}

	// Refresh private state from final GET response
	ps = extractPrivateStateFromResponse(finalPolicy)
	resp.Diagnostics.Append(savePrivateState(ctx, resp.Private, ps)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state from the final GET response
	d = populateModelFromAPI(ctx, &planModel, finalPolicy)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, planModel)...)
}
