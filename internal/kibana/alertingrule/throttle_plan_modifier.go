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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// planThrottleForActionFrequency implements the decision logic for
// SetUnknownIfActionsFrequencyConfigured: when the configuration is
// transitioning the rule into per-action frequency mode (config has an
// actions[*].frequency block AND prior state did not), the planned rule-level
// throttle value (already resolved by an earlier UseStateForUnknown to the
// API-preserved value) is flipped back to unknown so that:
//
//  1. The stale rule-level throttle is not sent alongside per-action frequency,
//     which Kibana documents as an invalid combination.
//  2. The plan does not assert a specific final throttle value, letting state
//     resolve to whatever Kibana returns without an inconsistent-result error.
//
// On subsequent plans where the rule is already in frequency mode (state also
// has an actions[*].frequency block), the planned value is left untouched so
// the modifier remains idempotent and does not produce a perpetual
// "value -> (known after apply)" diff against the Kibana-preserved throttle.
func planThrottleForActionFrequency(ctx context.Context, planValue types.String, stateActions, configActions types.List, diags *diag.Diagnostics) types.String {
	configHasFrequency := actionsListIncludesKnownFrequencyBlock(ctx, configActions, diags)
	if !configHasFrequency {
		return planValue
	}
	stateHasFrequency := actionsListIncludesKnownFrequencyBlock(ctx, stateActions, diags)
	if stateHasFrequency {
		return planValue
	}
	return types.StringUnknown()
}

// actionsListIncludesKnownFrequencyBlock is a thin alias over
// configActionsIncludeKnownFrequencyBlock that also accepts the state's
// actions list (the body of the function does not depend on the list having
// come from config). Kept as a separate name so call sites read naturally.
func actionsListIncludesKnownFrequencyBlock(ctx context.Context, actions types.List, diags *diag.Diagnostics) bool {
	return configActionsIncludeKnownFrequencyBlock(ctx, actions, diags)
}

// SetUnknownIfActionsFrequencyConfigured returns a plan modifier intended to
// be registered on rule-level throttle AFTER stringplanmodifier.UseStateForUnknown.
//
// USFU is the primary safeguard against the Kibana PUT behaviour that
// preserves deprecated rule-level throttle even when omitted or sent as null
// (preventing a provider inconsistency error when the practitioner removes the
// field from configuration). This trailing modifier resets the plan back to
// unknown when actions[*].frequency is configured, so the preserved value is
// not silently carried into the API request.
func SetUnknownIfActionsFrequencyConfigured() planmodifier.String {
	return setUnknownIfActionsFrequencyConfigured{}
}

type setUnknownIfActionsFrequencyConfigured struct{}

func (setUnknownIfActionsFrequencyConfigured) Description(_ context.Context) string {
	return "Resets the planned rule-level throttle to unknown when actions[*].frequency is configured, " +
		"so the value preserved from prior state by UseStateForUnknown is not sent alongside per-action frequency."
}

func (m setUnknownIfActionsFrequencyConfigured) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m setUnknownIfActionsFrequencyConfigured) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Nothing to do on create (no prior state) or destroy (no config).
	if req.State.Raw.IsNull() || req.Config.Raw.IsNull() {
		return
	}

	var configActions types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("actions"), &configActions)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateActions types.List
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("actions"), &stateActions)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = planThrottleForActionFrequency(ctx, req.PlanValue, stateActions, configActions, &resp.Diagnostics)
}
