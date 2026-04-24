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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// minVersionCustomPackageGet is the minimum Kibana version at which
// GET /api/fleet/epm/packages/{name}/{version} reliably supports
// custom-uploaded packages.
var minVersionCustomPackageGet = goversion.Must(goversion.NewVersion("8.2.0"))

func (r *customIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customIntegrationModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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

	meetsMinVersion, verDiags := apiClient.EnforceMinVersion(ctx, minVersionCustomPackageGet)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(verDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !meetsMinVersion {
		resp.Diagnostics.AddError(
			"Kibana version not supported",
			"elasticstack_fleet_custom_integration requires Kibana 8.2.0 or later.",
		)
		return
	}

	pkg, pkgDiags := fleet.GetPackage(ctx, fleetClient, state.PackageName.ValueString(), state.PackageVersion.ValueString(), state.SpaceID.ValueString())
	if pkgDiags.HasError() {
		resp.Diagnostics.Append(pkgDiags...)
		return
	}

	if pkg == nil {
		packages, listDiags := fleet.GetPackages(ctx, fleetClient, true, state.SpaceID.ValueString())
		if listDiags.HasError() {
			resp.Diagnostics.Append(listDiags...)
			return
		}
		for _, candidate := range packages {
			if candidate.Name != state.PackageName.ValueString() || candidate.Version != state.PackageVersion.ValueString() {
				continue
			}
			if candidate.Status != nil && strings.EqualFold(*candidate.Status, "installed") {
				diags = resp.State.Set(ctx, state)
				resp.Diagnostics.Append(diags...)
				return
			}
		}
		resp.State.RemoveResource(ctx)
		return
	}

	if pkg.Status == nil || *pkg.Status != "installed" {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
