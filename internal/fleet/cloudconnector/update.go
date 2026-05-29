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

package cloudconnector

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateCloudConnector(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[cloudConnectorModel],
) (entitycore.KibanaWriteResult[cloudConnectorModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	if req.Prior == nil {
		diags.AddError("Internal provider error", "Update callback received nil prior state.")
		return entitycore.KibanaWriteResult[cloudConnectorModel]{}, diags
	}

	fleetClient := client.GetFleetClient()

	body, bodyDiags := plan.toAPIUpdateBody(req.Config, *req.Prior)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[cloudConnectorModel]{}, diags
	}

	updated, updateDiags := fleetclient.UpdateCloudConnector(ctx, fleetClient, req.SpaceID, req.WriteID, body)
	diags.Append(updateDiags...)
	if diags.HasError() || updated == nil {
		return entitycore.KibanaWriteResult[cloudConnectorModel]{}, diags
	}

	forceDelete := plan.ForceDelete
	diags.Append(plan.populateFromAPI(req.SpaceID, *updated)...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[cloudConnectorModel]{}, diags
	}
	plan.ForceDelete = forceDelete

	return entitycore.KibanaWriteResult[cloudConnectorModel]{Model: plan}, diags
}
