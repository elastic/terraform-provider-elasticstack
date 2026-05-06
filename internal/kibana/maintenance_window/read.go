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

func readMaintenanceWindow(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model Model) (Model, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Unable to get Kibana client", err.Error())
		return model, false, diags
	}

	maintenanceWindow, getDiags := kibanaoapi.GetMaintenanceWindow(ctx, oapiClient, spaceID, resourceID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	if maintenanceWindow == nil {
		return model, false, diags
	}

	diags.Append(model.fromAPIReadResponse(ctx, maintenanceWindow)...)
	if diags.HasError() {
		return model, false, diags
	}

	model.ID = types.StringValue(resourceID)
	model.SpaceID = types.StringValue(spaceID)

	return model, true, diags
}
