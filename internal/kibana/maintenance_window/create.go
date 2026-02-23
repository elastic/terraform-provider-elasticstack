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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var planMaintenanceWindow Model

	diags := req.Plan.Get(ctx, &planMaintenanceWindow)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	body, diags := planMaintenanceWindow.toAPICreateRequest(ctx)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isSupported, sdkDiags := r.client.EnforceMinVersion(ctx, version.Must(version.NewVersion("9.1.0")))
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !isSupported {
		resp.Diagnostics.AddError("Unsupported server version", "Maintenance windows are not supported until Elastic Stack v9.0. Upgrade the target server to use this resource")
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	spaceID := planMaintenanceWindow.SpaceID.ValueString()
	createMaintenanceWindowResponse, diags := kibanaoapi.CreateMaintenanceWindow(ctx, client, spaceID, body)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	/*
	* In create/update paths we typically follow the write operation with a read, and then set the state from the read.
	* We want to avoid a dirty plan immediately after an apply.
	 */
	maintenanceWindowID := createMaintenanceWindowResponse.JSON200.Id
	readMaintenanceWindowResponse, diags := kibanaoapi.GetMaintenanceWindow(ctx, client, spaceID, maintenanceWindowID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readMaintenanceWindowResponse == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = planMaintenanceWindow.fromAPIReadResponse(ctx, readMaintenanceWindowResponse)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	planMaintenanceWindow.ID = types.StringValue(maintenanceWindowID)
	planMaintenanceWindow.SpaceID = types.StringValue(spaceID)

	diags = resp.State.Set(ctx, planMaintenanceWindow)
	resp.Diagnostics.Append(diags...)
}
