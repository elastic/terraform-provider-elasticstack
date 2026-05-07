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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createMaintenanceWindow(ctx context.Context, client *clients.KibanaScopedClient, spaceID string, plan Model) (Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	body, bodyDiags := plan.toAPICreateRequest(ctx)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return Model{}, diags
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Unable to get Kibana client", err.Error())
		return Model{}, diags
	}

	createMaintenanceWindowResponse, createDiags := kibanaoapi.CreateMaintenanceWindow(ctx, oapiClient, spaceID, body)
	diags.Append(createDiags...)
	if diags.HasError() {
		return Model{}, diags
	}

	/*
	* In create/update paths we typically follow the write operation with a read, and then set the state from the read.
	* We want to avoid a dirty plan immediately after an apply.
	 */
	maintenanceWindowID := createMaintenanceWindowResponse.JSON200.Id
	readMaintenanceWindowResponse, readDiags := kibanaoapi.GetMaintenanceWindow(ctx, oapiClient, spaceID, maintenanceWindowID)
	diags.Append(readDiags...)
	if diags.HasError() {
		return Model{}, diags
	}

	diags.Append(plan.fromAPIReadResponse(ctx, readMaintenanceWindowResponse)...)
	if diags.HasError() {
		return Model{}, diags
	}

	plan.ID = types.StringValue(maintenanceWindowID)
	plan.SpaceID = types.StringValue(spaceID)

	return plan, diags
}
