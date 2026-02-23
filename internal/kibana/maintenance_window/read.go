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

package maintenancewindow

import (
	"context"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateModel Model

	req.State.GetAttribute(ctx, path.Root("id"), &stateModel.ID)
	req.State.GetAttribute(ctx, path.Root("space_id"), &stateModel.SpaceID)

	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	if sdkDiags.HasError() {
		return
	}

	serverFlavor, sdkDiags := r.client.ServerFlavor(ctx)
	if sdkDiags.HasError() {
		return
	}

	diags := validateMaintenanceWindowServer(serverVersion, serverFlavor)
	if diags.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	maintenanceWindowID, spaceID := stateModel.getMaintenanceWindowIDAndSpaceID()
	maintenanceWindow, diags := kibanaoapi.GetMaintenanceWindow(ctx, client, spaceID, maintenanceWindowID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if maintenanceWindow == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = stateModel.fromAPIReadResponse(ctx, maintenanceWindow)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateModel.ID = types.StringValue(maintenanceWindowID)
	stateModel.SpaceID = types.StringValue(spaceID)

	diags = resp.State.Set(ctx, stateModel)
	resp.Diagnostics.Append(diags...)
}
