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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// createAgentlessPolicy is a stub for Task 3 of the fleet-agentless-policy
// OpenSpec change. Full implementation -- calling
// fleetclient.CreateAgentlessPolicy (POST /api/fleet/agentless_policies)
// and the deployment-topology preflight check -- lands in Tasks 5 and 6.
func createAgentlessPolicy(
	_ context.Context,
	_ *clients.KibanaScopedClient,
	_ entitycore.KibanaWriteRequest[agentlessPolicyModel],
) (entitycore.KibanaWriteResult[agentlessPolicyModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	diags.AddError(
		"Not yet implemented",
		"The elasticstack_fleet_agentless_policy resource's Create operation is not yet implemented "+
			"(see openspec/changes/fleet-agentless-policy, Task 5).",
	)
	return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
}
