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

package fleet

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PolicyFetchFunc fetches a Fleet policy of type P by ID within the given
// operational space.
type PolicyFetchFunc[P any] func(ctx context.Context, client *fleetclient.Client, policyID string, spaceID string) (*P, diag.Diagnostics)

// ReadPolicyEnvelope implements the get-state → resolve-client →
// resolve-space → fetch → nil-check sequence shared by Fleet policy
// resources' Read methods (integration_policy, elastic_defend_integration_policy,
// agentpolicy).
//
// It loads stateModel from req.State, resolves the Kibana/Fleet client via
// providerClient, resolves the operational space from state (not plan, so an
// in-flight space_ids change still queries the space the resource currently
// lives in - see [GetOperationalSpaceFromState]), and invokes fetch.
//
// On any failure, diagnostics are appended to resp and ok is false; the
// caller must return immediately. When the policy no longer exists, this
// removes it from state and returns ok=false. On success, ok is true and the
// caller proceeds with its package-specific populate-from-API step followed
// by its own resp.State.Set call.
func ReadPolicyEnvelope[T any, P any](
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
	providerClient *clients.ProviderClientFactory,
	stateModel *T,
	kibanaConnection func(*T) types.List,
	policyID func(*T) string,
	fetch PolicyFetchFunc[P],
) (fleetClient *fleetclient.Client, resolvedPolicyID string, spaceID string, policy *P, ok bool) {
	diags := req.State.Get(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return nil, "", "", nil, false
	}

	client, diags := providerClient.GetKibanaClient(ctx, kibanaConnection(stateModel))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return nil, "", "", nil, false
	}

	fleetClient = client.GetFleetClient()
	resolvedPolicyID = policyID(stateModel)

	// Read the existing spaces from state to determine where to query
	spaceID, diags = GetOperationalSpaceFromState(ctx, req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return nil, "", "", nil, false
	}

	// Query using the operational space from STATE
	policy, diags = fetch(ctx, fleetClient, resolvedPolicyID, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return nil, "", "", nil, false
	}

	if policy == nil {
		resp.State.RemoveResource(ctx)
		return nil, "", "", nil, false
	}

	return fleetClient, resolvedPolicyID, spaceID, policy, true
}
