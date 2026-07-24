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

package serverhost

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateServerHost(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[serverHostModel]) (entitycore.KibanaWriteResult[serverHostModel], diag.Diagnostics) {
	plan := req.Plan
	fleetClient := client.GetFleetClient()

	hostID := plan.HostID.ValueString()
	spaceID := req.SpaceID
	if req.Prior != nil {
		spaceID = req.Prior.GetSpaceID().ValueString()
	}

	return entitycore.WriteEntity(
		func() (kbapi.PutFleetFleetServerHostsItemidJSONRequestBody, diag.Diagnostics) {
			return plan.toAPIUpdateModel(ctx)
		},
		func(body kbapi.PutFleetFleetServerHostsItemidJSONRequestBody) (*kbapi.ServerHost, diag.Diagnostics) {
			return fleet.UpdateFleetServerHost(ctx, fleetClient, hostID, spaceID, body)
		},
		func(host *kbapi.ServerHost) (serverHostModel, diag.Diagnostics) {
			diags := plan.populateFromAPI(ctx, host)
			return plan, diags
		},
	)
}
