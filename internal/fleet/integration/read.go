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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readIntegration(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	_ string,
	spaceID string,
	model integrationModel,
) (integrationModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	fleetClient := client.GetFleetClient()

	name := model.Name.ValueString()
	version := model.Version.ValueString()

	spaceAware := resolveSpaceAware(ctx, client, model.SpaceID, &diags)
	if diags.HasError() {
		return model, false, diags
	}

	pkg, getDiags := fleet.GetPackage(ctx, fleetClient, name, version, spaceID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return model, false, diags
	}
	if pkg == nil || !fleetPackageInstalled(pkg, spaceID, spaceAware) {
		return model, false, diags
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
	model.Version = types.StringValue(installedVersion)
	model.ID = types.StringValue(getPackageID(name, installedVersion))
	if model.SpaceID.IsNull() {
		model.SpaceID = installedKibanaSpaceID(pkg)
	}

	return model, true, diags
}
