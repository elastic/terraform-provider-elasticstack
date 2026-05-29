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

package cloudconnector

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const forceDeleteHint = "To force-delete a cloud connector that is still referenced by package policies, set force_delete = true. " +
	"Note: this is destructive and will leave the package policies broken."

func deleteCloudConnector(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model cloudConnectorModel) diag.Diagnostics {
	fleetClient := client.GetFleetClient()

	force := model.ForceDelete.ValueBool()
	diags := fleetclient.DeleteCloudConnector(ctx, fleetClient, spaceID, resourceID, force)
	if diags.HasError() && !force {
		return augmentInUseConflictDiagnostic(diags)
	}
	return diags
}

// augmentInUseConflictDiagnostic appends a force_delete hint when the API error
// body mentions package_policy_count (option 1: pattern-match diagnostic detail).
func augmentInUseConflictDiagnostic(diags diag.Diagnostics) diag.Diagnostics {
	for _, d := range diags {
		if d.Severity() == diag.SeverityError && strings.Contains(d.Detail(), "package_policy_count") {
			diags.AddError("Cloud connector in use", forceDeleteHint)
			break
		}
	}
	return diags
}
