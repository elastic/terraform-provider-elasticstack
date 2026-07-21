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

package managedintegration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// createAgentlessPolicy compiles the plan into a managed-integration create
// request body and calls POST /api/fleet/managed_integrations. Per spec: a
// non-2xx response surfaces diagnostics and no state is saved (the entitycore
// envelope never calls resp.State.Set when this callback returns an error --
// see kibana_resource_envelope.go's runKibanaWrite).
//
// The MinVersion 9.5.0 gate for managed_integrations is wired via
// GetVersionRequirements (models.go) and enforced by the entitycore envelope
// (entitycore.EnforceVersionRequirements, called from
// kibana_resource_envelope.go's Create) before this function is ever invoked
// -- see TestAgentlessPolicyModel_versionGate_firesBeforeAPICall in
// entitycore_contract_test.go.
//
// Task 6.2 adds the deployment-topology preflight check below (self-managed
// stacks are rejected before the POST call runs; see checkDeploymentTopology
// and its tests in topology.go/topology_test.go, design.md Decision 7, and
// specs/fleet-agentless-policy/spec.md's "Deployment topology preflight
// check" requirement).
//
// design.md's Open Question 6 asked whether the fail-open heuristic should
// additionally offer an explicit opt-out for legitimate Elastic Cloud
// Hosted/Serverless deployments whose networking (e.g. PrivateLink) never
// emits the cloud-proxy headers checkDeploymentTopology looks for and so get
// permanently, incorrectly classified as self-managed with no
// Terraform-native workaround. That question is now resolved: the
// `skip_topology_check` schema attribute (schema.go) is the escape hatch.
// When it is true, checkDeploymentTopology is not called at all -- the check
// itself makes a live HTTP call to Kibana's status endpoint, so there is no
// reason to pay for that call when the user has explicitly opted out of its
// result. Version gating (GetVersionRequirements, enforced by the envelope
// before this function even runs) is unaffected either way.
//
// Per coding-standards.md, persisted state after create comes from the
// entitycore envelope read-after-write (Read callback), not from the POST
// response body. This callback only copies the server-assigned policy id into
// the returned model so the envelope can invoke Read.
func createAgentlessPolicy(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[agentlessPolicyModel],
) (entitycore.KibanaWriteResult[agentlessPolicyModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	if !plan.SkipTopologyCheck.ValueBool() {
		diags.Append(checkDeploymentTopology(ctx, client)...)
		if diags.HasError() {
			return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
		}
	}

	fleetClient := client.GetFleetClient()

	body, bodyDiags := plan.toCreateBody(ctx)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	item, createDiags := fleetclient.CreateManagedIntegration(ctx, fleetClient, req.SpaceID, body)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}
	if item == nil || item.Id == "" {
		diags.AddError(
			"Managed integration create returned no identifier",
			"POST /api/fleet/managed_integrations succeeded but did not return a policy id.",
		)
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	plan.PolicyID = types.StringValue(item.Id)
	plan.ID = types.StringValue((&clients.CompositeID{ClusterID: req.SpaceID, ResourceID: item.Id}).String())

	return entitycore.KibanaWriteResult[agentlessPolicyModel]{Model: plan}, diags
}
