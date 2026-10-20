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
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func updateServerHost(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[serverHostModel]) (entitycore.KibanaWriteResult[serverHostModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	fleetClient := client.GetFleetClient()

	hostID := req.Plan.HostID.ValueString()
	var prior serverHostModel
	if req.Prior != nil {
		prior = *req.Prior
	}
	body, d := req.Plan.toAPIUpdateModel(ctx, prior)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[serverHostModel]{}, diags
	}

	spaceID := req.SpaceID
	if req.Prior != nil {
		spaceID = req.Prior.GetSpaceID().ValueString()
	}

	if body.Ssl != nil {
		diags.Append(checkNoDeferredSSLFields(ctx, fleetClient, hostID, spaceID)...)
		if diags.HasError() {
			return entitycore.KibanaWriteResult[serverHostModel]{}, diags
		}
	}

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

// checkNoDeferredSSLFields rejects an ssl update when the stored host holds
// agent_*/es_* ssl fields, because the PUT replaces the whole ssl object.
func checkNoDeferredSSLFields(ctx context.Context, fleetClient *fleet.Client, hostID, spaceID string) (diags diag.Diagnostics) {
	current, d := fleet.GetFleetServerHost(ctx, fleetClient, hostID, spaceID)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	if fields := deferredSSLFields(current); len(fields) > 0 {
		diags.AddAttributeError(
			path.Root("ssl"),
			"Cannot update ssl without erasing unmanaged fields",
			fmt.Sprintf(
				"The Fleet server host stores ssl fields that this resource does not manage (%s). "+
					"Updating ssl replaces the whole ssl object and would erase them, so the update was not sent. "+
					"Remove these fields from the host outside of Terraform before changing ssl.",
				strings.Join(fields, ", "),
			),
		)
	}
	return diags
}
