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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// readDetectionRule reads a security detection rule and returns the model,
// a bool indicating whether the rule was found, and any diagnostics.
func readDetectionRule(ctx context.Context, apiClient *clients.KibanaScopedClient, resourceID, spaceID string, model Data) (Data, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	data := model
	data.initializeAllFieldsToDefaults()

	// Get the rule using kbapi client
	kbClient := apiClient.GetKibanaOapiClient()

	// Read the rule
	uid, err := uuid.Parse(resourceID)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return model, false, diags
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
		return model, false, diags
	}

	if response.StatusCode() == 404 {
		// Rule was deleted - return false to indicate this
		return model, false, diags
	}

	if response.StatusCode() != 200 {
		diags.AddError(
			"Error reading security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return model, false, diags
	}

	// Parse the response
	updateDiags := data.updateFromRule(ctx, response.JSON200)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	// Reconcile empty lists from the reference model to preserve explicit [] configurations
	reconcileEmptyListsFromPlan(ctx, &model, &data)

	// Ensure space_id is set correctly
	data.SpaceID = types.StringValue(spaceID)

	compID := clients.CompositeID{
		ResourceID: resourceID,
		ClusterID:  spaceID,
	}

	data.ID = types.StringValue(compID.String())

	return data, true, diags
}
