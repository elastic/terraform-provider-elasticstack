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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// minVersionCustomPackageGet is the minimum Kibana version at which
// GET /api/fleet/epm/packages/{name}/{version} reliably returns status
// information for custom-uploaded packages. On older versions (7.17.x returns
// HTTP 400, 8.0.x–8.1.x returns HTTP 404) the endpoint does not support
// custom packages and cannot be used for drift detection.
var minVersionCustomPackageGet = goversion.Must(goversion.NewVersion("8.2.0"))

func (r *customIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customIntegrationModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiClient, diags := r.client.GetKibanaClient(ctx, state.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fleetClient, err := apiClient.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	// Determine whether GetPackage supports custom-uploaded packages on this
	// Kibana version. On Kibana <8.2, the individual package GET endpoint
	// returns HTTP 400 (7.17.x) or HTTP 404 (8.0.x–8.1.x) for custom packages
	// even when installed. On 8.2+ it returns 200 when installed and 404 when
	// genuinely absent, making it usable for drift detection.
	supportsCustomPackageGet, verDiags := apiClient.EnforceMinVersion(ctx, minVersionCustomPackageGet)
	if verDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(verDiags)...)
		return
	}

	pkg, pkgDiags := fleet.GetPackage(ctx, fleetClient, state.PackageName.ValueString(), state.PackageVersion.ValueString(), state.SpaceID.ValueString())
	if pkgDiags.HasError() {
		if supportsCustomPackageGet {
			// On a new-enough Kibana, unexpected errors from GetPackage should be
			// surfaced to the user.
			resp.Diagnostics.Append(pkgDiags...)
			return
		}
		// On older Kibana (7.17.x returns HTTP 400 "filePath"), GetPackage is not
		// supported for custom packages. Keep the existing state so the resource is
		// not inadvertently removed from state.
		diags = resp.State.Set(ctx, state)
		resp.Diagnostics.Append(diags...)
		return
	}

	if pkg == nil {
		if !supportsCustomPackageGet {
			// Older Kibana (8.0.x–8.1.x) returns HTTP 404 for custom-uploaded
			// packages via the individual GET endpoint even when they are installed.
			// Keep state rather than incorrectly removing the resource.
			diags = resp.State.Set(ctx, state)
			resp.Diagnostics.Append(diags...)
			return
		}
		// On Kibana 8.2+, HTTP 404 means the package is genuinely not installed.
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
