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

package integration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *integrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel integrationModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := r.Client().GetKibanaClient(ctx, stateModel.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fleetClient, err := client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	name := stateModel.Name.ValueString()
	version := stateModel.Version.ValueString()
	spaceID := stateModel.SpaceID.ValueString()

	spaceAware := false
	if typeutils.IsKnown(stateModel.SpaceID) {
		supported, versionDiags := supportsSpaceAwareIntegration(ctx, client, spaceID)
		resp.Diagnostics.Append(versionDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		spaceAware = supported
	}

	pkg, diags := fleet.GetPackage(ctx, fleetClient, name, version, spaceID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pkg == nil || !fleetPackageInstalled(pkg, spaceID, spaceAware) {
		resp.State.RemoveResource(ctx)
		return
	}

	// Fleet's GET /epm/packages/{name}/{version} reports status "installed"
	// whenever the package is installed at *any* version, regardless of the
	// version path parameter. Use InstallationInfo.Version when present so
	// Terraform observes out-of-band upgrades and downgrades as drift.
	// See https://github.com/elastic/terraform-provider-elasticstack/issues/1585.
	installedVersion := version
	if pkg.InstallationInfo != nil && pkg.InstallationInfo.Version != "" {
		installedVersion = pkg.InstallationInfo.Version
	}
	stateModel.Version = types.StringValue(installedVersion)
	stateModel.ID = types.StringValue(getPackageID(name, installedVersion))
	if stateModel.SpaceID.IsNull() {
		stateModel.SpaceID = installedKibanaSpaceID(pkg)
	}

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
