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

package customintegration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func deleteCustomIntegration(ctx context.Context, client *clients.KibanaScopedClient, _, spaceID string, model customIntegrationModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if model.SkipDestroy.ValueBool() {
		tflog.Debug(ctx, "Skipping uninstall of custom integration package", map[string]any{
			"name":    model.PackageName.ValueString(),
			"version": model.PackageVersion.ValueString(),
		})
		return diags
	}

	fleetClient := client.GetFleetClient()

	if model.PackageName.IsNull() || model.PackageName.IsUnknown() || model.PackageName.ValueString() == "" {
		diags.AddError(
			"Cannot uninstall custom integration package",
			"skip_destroy is false, but package_name is not set in state. The provider cannot determine which Fleet package to uninstall.",
		)
		return diags
	}

	if model.PackageVersion.IsNull() || model.PackageVersion.IsUnknown() || model.PackageVersion.ValueString() == "" {
		diags.AddError(
			"Cannot uninstall custom integration package",
			"skip_destroy is false, but package_version is not set in state. The provider cannot determine which Fleet package version to uninstall.",
		)
		return diags
	}

	diags = fleet.Uninstall(ctx, fleetClient, model.PackageName.ValueString(), model.PackageVersion.ValueString(), spaceID, false)
	return diags
}
