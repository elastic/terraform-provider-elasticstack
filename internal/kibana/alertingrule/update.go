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

package alertingrule

import (
	"context"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertingRuleModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for rule ID
	var state alertingRuleModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Expected configured API client")
		return
	}

	// Get server version to validate version-specific features
	serverVersion, versionDiags := r.client.ServerVersion(ctx)
	if versionDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(versionDiags)...)
		return
	}

	// Convert to API model (includes version-specific validation)
	rule, d := plan.toAPIModel(ctx, serverVersion)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure rule ID is set from state
	ruleID, spaceID := state.getRuleIDAndSpaceID()
	rule.RuleID = ruleID
	rule.SpaceID = spaceID

	oapiClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana client", err.Error())
		return
	}

	// Update the rule (enable/disable logic is handled inside)
	_, updateDiags := kibanaoapi.UpdateAlertingRule(ctx, oapiClient, rule.SpaceID, rule)
	resp.Diagnostics.Append(updateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set plan's ID to the state's ID so we can re-read from API
	plan.ID = state.ID
	plan.RuleID = state.RuleID
	plan.SpaceID = state.SpaceID

	// Re-read rule from API to get the authoritative state
	// (sometimes update response differs from what's actually stored)
	exists, readDiags := r.readRuleFromAPI(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !exists {
		resp.Diagnostics.AddError("Rule not found after update", "The alerting rule was updated but could not be read back from the API")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
