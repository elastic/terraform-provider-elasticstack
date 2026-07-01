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

// deleteAgentlessPolicy implements Task 5.4 of the fleet-agentless-policy
// OpenSpec change: calls fleetclient.DeleteAgentlessPolicy
// (DELETE /api/fleet/agentless_policies/{id}, space-aware) with
// force = force_delete. HTTP 404 is already treated as a no-op by
// fleetclient.DeleteAgentlessPolicy (see internal/clients/fleet/
// agentless_policy.go and internal/clients/fleet/responses.go's
// handleDeleteResponse, both Task 2 deliverables this task does not modify).
//
// When force_delete = false and the API returns a conflict, a helpful
// diagnostic pointing at force_delete is appended (see conflictHintDiagnostics).
// fleetclient.DeleteAgentlessPolicy reports whether the underlying response
// was an HTTP 409 directly (isConflict), so this no longer has to infer a
// conflict by pattern-matching diagnostic summary text.
func deleteAgentlessPolicy(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model agentlessPolicyModel) diag.Diagnostics {
	fleetClient := client.GetFleetClient()

	force := model.ForceDelete.ValueBool()
	isConflict, diags := fleetclient.DeleteAgentlessPolicy(ctx, fleetClient, spaceID, resourceID, force)
	if diags.HasError() && !force && isConflict {
		diags.Append(conflictHintDiagnostics()...)
	}
	return diags
}

// conflictHintDiagnostics returns a diagnostic explaining force_delete, for
// deleteAgentlessPolicy to append when fleetclient.DeleteAgentlessPolicy
// reports the delete failed with an HTTP 409 conflict.
func conflictHintDiagnostics() diag.Diagnostics {
	var hint diag.Diagnostics
	hint.AddError(
		"Agentless policy delete conflict",
		"Kibana refused to delete this agentless policy, likely because its underlying managed agent "+
			"policy is in a conflicting state (for example, still provisioning, or has associated agents). "+
			"Set force_delete = true on this resource and re-apply to force deletion "+
			"(sent to the API as the ?force=true query parameter).",
	)
	return hint
}
