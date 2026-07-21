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
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// readAgentlessPolicy reads the current managed integration via the temporary
// GET /api/fleet/package_policies/{id} compat path (see
// agentless_policy_compat.go) until task 8 switches to
// ReadManagedIntegration. A nil response (HTTP 404) signals the resource was
// removed out of band; the caller is responsible for removing it from state.
//
// force, force_delete, and create_dataset_templates are preserved from the
// incoming model because populateFromManagedIntegration deliberately never
// touches them: none round-trip through the read response.
func readAgentlessPolicy(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model agentlessPolicyModel) (agentlessPolicyModel, bool, diag.Diagnostics) {
	fleetClient := client.GetFleetClient()

	data, diags := fleetclient.ReadAgentlessPolicyViaPackagePolicy(ctx, fleetClient, spaceID, resourceID)
	if diags.HasError() {
		return model, false, diags
	}

	if data == nil {
		return model, false, diags
	}

	item, bridgeDiags := managedIntegrationFromPackagePolicyReadResponse(data)
	diags.Append(bridgeDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	var spaceIDs *[]string
	if data.SpaceIds != nil {
		spaceIDs = data.SpaceIds
	}

	diags.Append(model.populateFromManagedIntegration(ctx, spaceID, &item, spaceIDs)...)
	if diags.HasError() {
		return model, false, diags
	}

	return model, true, diags
}
