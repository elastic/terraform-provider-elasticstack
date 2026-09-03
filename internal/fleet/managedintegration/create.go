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

// createManagedIntegration POSTs /api/fleet/managed_integrations. Errors return
// diagnostics and no state (entitycore runKibanaWrite). MinVersion 9.5.0 is
// enforced by the envelope before this runs (GetVersionRequirements). Unless
// skip_topology_check is set, checkDeploymentTopology runs first (see
// openspec/specs/fleet-managed-integration/spec.md).
//
// Persisted state comes from the envelope read-after-write (Read callback),
// not the POST body; this callback only sets policy_id (and composite id) so
// Read can run.
func createManagedIntegration(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[managedIntegrationModel],
) (entitycore.KibanaWriteResult[managedIntegrationModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	if !plan.SkipTopologyCheck.ValueBool() {
		diags.Append(checkDeploymentTopology(ctx, client)...)
		if diags.HasError() {
			return entitycore.KibanaWriteResult[managedIntegrationModel]{}, diags
		}
	}

	fleetClient := client.GetFleetClient()

	body, bodyDiags := plan.toCreateBody(ctx)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[managedIntegrationModel]{}, diags
	}

	item, createDiags := fleetclient.CreateManagedIntegration(ctx, fleetClient, req.SpaceID, body)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[managedIntegrationModel]{}, diags
	}
	if item == nil || item.Id == "" {
		diags.AddError(
			"Managed integration create returned no identifier",
			"POST /api/fleet/managed_integrations succeeded but did not return a policy id.",
		)
		return entitycore.KibanaWriteResult[managedIntegrationModel]{}, diags
	}

	plan.PolicyID = types.StringValue(item.Id)
	plan.ID = types.StringValue((&clients.CompositeID{ClusterID: req.SpaceID, ResourceID: item.Id}).String())

	return entitycore.KibanaWriteResult[managedIntegrationModel]{Model: plan}, diags
}
