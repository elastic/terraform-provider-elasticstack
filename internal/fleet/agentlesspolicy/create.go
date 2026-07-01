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

package agentlesspolicy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// createAgentlessPolicy implements Task 5.1 of the fleet-agentless-policy
// OpenSpec change: compiles the plan into
// PostFleetAgentlessPoliciesJSONRequestBody, calls
// fleetclient.CreateAgentlessPolicy (POST /api/fleet/agentless_policies,
// space-aware), and decodes the response into state. Per spec: a non-2xx
// response surfaces diagnostics and no state is saved (the entitycore
// envelope never calls resp.State.Set when this callback returns an error --
// see kibana_resource_envelope.go's runKibanaWrite).
//
// TODO: Task 6 adds the deployment-topology preflight check here (self-managed
// stacks must be rejected before the POST call below runs; see design.md
// Decision 7 and specs/fleet-agentless-policy/spec.md's "Deployment topology
// preflight check" requirement). Task 6 also wires the MinVersion 9.3.0 gate
// via GetVersionRequirements (already implemented in models.go), which the
// entitycore envelope enforces before this function is ever invoked.
func createAgentlessPolicy(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[agentlessPolicyModel],
) (entitycore.KibanaWriteResult[agentlessPolicyModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	fleetClient := client.GetFleetClient()

	body, bodyDiags := plan.toCreateBody(ctx)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	item, createDiags := fleetclient.CreateAgentlessPolicy(ctx, fleetClient, req.SpaceID, body)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	diags.Append(plan.populateFromCreateResponse(ctx, req.SpaceID, *item)...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	return entitycore.KibanaWriteResult[agentlessPolicyModel]{Model: plan}, diags
}
