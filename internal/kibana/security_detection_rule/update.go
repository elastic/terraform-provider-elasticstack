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

package securitydetectionrule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *securityDetectionRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data Data

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule using kbapi client
	kbClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Kibana client",
			"Could not get Kibana OAPI client: "+err.Error(),
		)
		return
	}

	// Build the update request
	updateProps, diags := data.toUpdateProps(ctx, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the rule
	response, err := kbClient.API.UpdateRuleWithResponse(ctx, data.SpaceID.ValueString(), updateProps)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating security detection rule",
			"Could not update security detection rule: "+err.Error(),
		)
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error updating security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return
	}

	// Parse ID to get space_id and rule_id
	compID, resourceIDDiags := clients.CompositeIDFromStrFw(data.ID.ValueString())
	resp.Diagnostics.Append(resourceIDDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uid, err := uuid.Parse(compID.ResourceID)
	if err != nil {
		resp.Diagnostics.AddError("ID was not a valid UUID", err.Error())
		return
	}

	readData, diags := r.read(ctx, uid.String(), compID.ClusterID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readData == nil {
		resp.Diagnostics.AddError(
			"Error reading updated security detection rule",
			"Could not read security detection rule after update",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}
