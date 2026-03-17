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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *securityDetectionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data Data

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse ID to get space_id and rule_id
	compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the extracted read method
	readData, diags := r.read(ctx, compID.ResourceID, compID.ClusterID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the rule was found (nil data indicates 404)
	if readData == nil {
		// Rule was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	// Set the composite ID and state
	readData.ID = data.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, readData)...)
}

// read extracts the core functionality of reading a security detection rule
func (r *securityDetectionRuleResource) read(ctx context.Context, resourceID, spaceID string) (*Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	data := &Data{}
	data.initializeAllFieldsToDefaults()

	// Get the rule using kbapi client
	kbClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError(
			"Error getting Kibana client",
			"Could not get Kibana OAPI client: "+err.Error(),
		)
		return nil, diags
	}

	// Read the rule
	uid, err := uuid.Parse(resourceID)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return nil, diags
	}
	params := &kbapi.ReadRuleParams{
		Id: &uid,
	}

	response, err := kbClient.API.ReadRuleWithResponse(ctx, spaceID, params)
	if err != nil {
		diags.AddError(
			"Error reading security detection rule",
			"Could not read security detection rule: "+err.Error(),
		)
		return nil, diags
	}

	if response.StatusCode() == 404 {
		// Rule was deleted - return nil to indicate this
		return nil, diags
	}

	if response.StatusCode() != 200 {
		diags.AddError(
			"Error reading security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return nil, diags
	}

	// Parse the response
	updateDiags := data.updateFromRule(ctx, response.JSON200)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return nil, diags
	}

	// Ensure space_id is set correctly
	data.SpaceID = types.StringValue(spaceID)

	compID := clients.CompositeID{
		ResourceID: resourceID,
		ClusterID:  spaceID,
	}

	data.ID = types.StringValue(compID.String())

	return data, diags
}
