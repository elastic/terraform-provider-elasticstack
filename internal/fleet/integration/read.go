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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func installedInSpace(pkg *kbapi.PackageInfo, desiredSpaceID string) bool {
	if pkg == nil || pkg.Status == nil || *pkg.Status != "installed" {
		return false
	}

	if pkg.InstallationInfo == nil {
		return false
	}

	if pkg.InstallationInfo.InstalledKibanaSpaceId != nil && *pkg.InstallationInfo.InstalledKibanaSpaceId == desiredSpaceID {
		return true
	}

	if pkg.InstallationInfo.AdditionalSpacesInstalledKibana != nil {
		if _, ok := (*pkg.InstallationInfo.AdditionalSpacesInstalledKibana)[desiredSpaceID]; ok {
			return true
		}
	}

	return false
}

func (r *integrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel integrationModel

	diags := req.State.Get(ctx, &stateModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	name := stateModel.Name.ValueString()
	version := stateModel.Version.ValueString()
	pkg, diags := fleet.GetPackage(ctx, client, name, version)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if pkg == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	desiredSpaceID := "default"
	if typeutils.IsKnown(stateModel.SpaceID) {
		desiredSpaceID = stateModel.SpaceID.ValueString()
	}

	if !installedInSpace(pkg, desiredSpaceID) {
		resp.State.RemoveResource(ctx)
		return
	}

	stateModel.ID = types.StringValue(getPackageID(name, version))

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
