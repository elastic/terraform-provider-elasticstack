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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateServerHost(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[serverHostModel]) (entitycore.KibanaWriteResult[serverHostModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	fleetClient := client.GetFleetClient()

	hostID := req.Plan.HostID.ValueString()
	body, d := req.Plan.toAPIUpdateModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[serverHostModel]{}, diags
	}

	spaceID := req.Prior.GetSpaceID().ValueString()

	host, d := fleet.UpdateFleetServerHost(ctx, fleetClient, hostID, spaceID, body)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[serverHostModel]{}, diags
	}

	d = req.Plan.populateFromAPI(ctx, host)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[serverHostModel]{}, diags
	}

	return entitycore.KibanaWriteResult[serverHostModel]{Model: req.Plan}, diags
}
