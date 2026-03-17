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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityDetectionRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data Data

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the rule using kbapi client
	kbClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Kibana client",
			"Could not get Kibana OAPI client: "+err.Error(),
		)
		return
	}

	// Build the create request
	createProps, diags := data.toCreateProps(ctx, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the rule
	response, err := kbClient.API.CreateRuleWithResponse(ctx, data.SpaceID.ValueString(), createProps)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating security detection rule",
			"Could not create security detection rule: "+err.Error(),
		)
		return
	}

	if response.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			"Error creating security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return
	}

	// Parse the response to get the ID, then use Read logic for consistency
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the ID based on the created rule
	id, diags := extractID(response.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	compID := clients.CompositeID{
		ClusterID:  data.SpaceID.ValueString(),
		ResourceID: id,
	}
	data.ID = types.StringValue(compID.String())
	readData, diags := r.read(ctx, id, data.SpaceID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if readData == nil {
		resp.Diagnostics.AddError(
			"Error reading created security detection rule",
			"Could not read security detection rule after creation",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}
