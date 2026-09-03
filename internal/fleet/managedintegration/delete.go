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

// deleteManagedIntegration calls DELETE /api/fleet/managed_integrations/{id}
// (space-aware) with force = force_delete. HTTP 404 is treated as a no-op by
// the Fleet client (see handleDeleteResponse in responses.go).
//
// When force_delete = false and the API returns a conflict, a helpful diagnostic
// pointing at force_delete is appended (see conflictHintDiagnostics).
// fleetclient.DeleteManagedIntegration reports whether the final observed HTTP status
// was 409 directly (isConflict), so this no longer pattern-matches diagnostic text.
func deleteManagedIntegration(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model managedIntegrationModel) diag.Diagnostics {
	fleetClient := client.GetFleetClient()

	force := model.ForceDelete.ValueBool()
	isConflict, diags := fleetclient.DeleteManagedIntegration(ctx, fleetClient, spaceID, resourceID, force)
	if diags.HasError() && !force && isConflict {
		diags.Append(conflictHintDiagnostics()...)
	}
	return diags
}

// conflictHintDiagnostics returns a diagnostic explaining force_delete, for
// deleteManagedIntegration to append when fleetclient.DeleteManagedIntegration
// reports the delete failed with an HTTP 409 conflict.
func conflictHintDiagnostics() diag.Diagnostics {
	var hint diag.Diagnostics
	hint.AddError(
		"Managed integration delete conflict",
		"Kibana refused to delete this managed integration, likely because its underlying managed agent "+
			"policy is in a conflicting state (for example, still provisioning, or has associated agents). "+
			"Set force_delete = true on this resource and re-apply to force deletion "+
			"(sent to the API as the ?force=true query parameter).",
	)
	return hint
}
