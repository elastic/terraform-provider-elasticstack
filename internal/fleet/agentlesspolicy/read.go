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
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// readAgentlessPolicy implements Task 5.2 of the fleet-agentless-policy
// OpenSpec change: reads the current agentless policy via
// GET /api/fleet/package_policies/{id} (space-aware) -- Decision 4, there is
// no dedicated agentless GET endpoint. A nil response (HTTP 404) signals the
// resource was removed out of band; the caller (entitycore's
// baseResourceEnvelope.Read, or the write-then-read refresh in
// runKibanaWrite) is responsible for removing it from state.
//
// force, force_delete, and create_dataset_templates are preserved from the
// incoming model -- which is either the prior state (plain Read) or the
// plan/written model (post-write refresh) -- because
// populateFromPackagePolicy deliberately never touches them: none of the
// three round-trip through this API response (see specs/
// fleet-agentless-policy/spec.md's "Read preserves force_delete" and
// "Create-only flags are not round-tripped from the API" scenarios).
func readAgentlessPolicy(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model agentlessPolicyModel) (agentlessPolicyModel, bool, diag.Diagnostics) {
	fleetClient := client.GetFleetClient()

	data, diags := fleetclient.ReadAgentlessPolicyViaPackagePolicy(ctx, fleetClient, spaceID, resourceID)
	if diags.HasError() {
		return model, false, diags
	}

	if data == nil {
		return model, false, diags
	}

	diags.Append(model.populateFromPackagePolicy(ctx, spaceID, data)...)
	if diags.HasError() {
		return model, false, diags
	}

	return model, true, diags
}
