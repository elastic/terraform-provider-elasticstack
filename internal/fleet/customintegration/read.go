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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *customIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibanaClient, diags := r.client.GetKibanaClient(ctx, state.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	fleetClient, err := kibanaClient.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to obtain Fleet client", err.Error())
		return
	}

	name := state.PackageName.ValueString()
	version := state.PackageVersion.ValueString()
	if name == "" || version == "" {
		// State is incomplete; can't verify. Remove and let a fresh
		// create re-establish everything.
		resp.State.RemoveResource(ctx)
		return
	}

	pkg, getDiags := fleet.GetPackage(ctx, fleetClient, name, version, state.SpaceID.ValueString())
	resp.Diagnostics.Append(getDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if pkg == nil || !installed(pkg) {
		resp.State.RemoveResource(ctx)
		return
	}

	// Nothing on the server side changes the values the user cares about
	// (name/version/checksum are all anchored to the uploaded archive).
	// Persist the state as-is.
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// installed reports whether Fleet considers the package fully installed,
// across both modern (`InstallationInfo.InstallStatus`) and legacy
// (`Status`) response shapes. Mirrors the helper in internal/fleet/integration.
func installed(pkg *kbapi.PackageInfo) bool {
	if pkg == nil {
		return false
	}
	if pkg.InstallationInfo != nil {
		switch pkg.InstallationInfo.InstallStatus {
		case kbapi.PackageInfoInstallationInfoInstallStatusInstalled:
			return true
		case kbapi.PackageInfoInstallationInfoInstallStatusInstallFailed:
			return false
		}
	}
	if pkg.Status != nil {
		return strings.EqualFold(*pkg.Status, "installed")
	}
	return false
}

// pickInstalledVersion returns the highest version string for a named package
// that reports `installed` status in the packages list. Used as a fallback
// when the upload response does not carry a version.
func pickInstalledVersion(pkgs []kbapi.PackageListItem, name string) string {
	var best string
	for _, p := range pkgs {
		if p.Name != name {
			continue
		}
		if p.Status != nil && !strings.EqualFold(string(*p.Status), "installed") {
			continue
		}
		if p.Version == "" {
			continue
		}
		if best == "" || p.Version > best {
			best = p.Version
		}
	}
	return best
}
