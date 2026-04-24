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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *customIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customIntegrationModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SkipDestroy.ValueBool() {
		tflog.Debug(ctx, "Skipping uninstall of custom integration package", map[string]any{
			"name":    state.PackageName.ValueString(),
			"version": state.PackageVersion.ValueString(),
		})
		return
	}

	apiClient, diags := r.Client().GetKibanaClient(ctx, state.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fleetClient, err := apiClient.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	if state.PackageName.IsNull() || state.PackageName.IsUnknown() || state.PackageName.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Cannot uninstall custom integration package",
			"skip_destroy is false, but package_name is not set in state. The provider cannot determine which Fleet package to uninstall.",
		)
		return
	}

	if state.PackageVersion.IsNull() || state.PackageVersion.IsUnknown() || state.PackageVersion.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Cannot uninstall custom integration package",
			"skip_destroy is false, but package_version is not set in state. The provider cannot determine which Fleet package version to uninstall.",
		)
		return
	}

	diags = fleet.Uninstall(ctx, fleetClient, state.PackageName.ValueString(), state.PackageVersion.ValueString(), state.SpaceID.ValueString(), false)
	resp.Diagnostics.Append(diags...)
}
