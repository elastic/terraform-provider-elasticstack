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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/planmodifiers"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const frequencyExclusivityDetail = "Rule-level notify_when and throttle cannot be combined with actions[*].frequency " +
	"(per-action notification). Use either rule-level notify_when/throttle or per-action frequency blocks, not both. " +
	"Kibana does not allow these parameters when notify_when or throttle are defined at the rule level."

func validateNotifyWhenThrottleFrequencyExclusivity(ctx context.Context, data *alertingRuleModel, diags *diag.Diagnostics) {
	if !configActionsIncludeKnownFrequencyBlock(ctx, data.Actions, diags) {
		return
	}
	ruleNotify := ruleLevelNotifyWhenExclusive(data.NotifyWhen)
	ruleThrottle := ruleLevelThrottleExclusive(data.Throttle.StringValue)
	if !ruleNotify && !ruleThrottle {
		return
	}
	if ruleNotify {
		diags.AddAttributeError(path.Root("notify_when"), "Cannot combine rule-level notify_when with actions.frequency", frequencyExclusivityDetail)
		return
	}
	diags.AddAttributeError(path.Root("throttle"), "Cannot combine rule-level throttle with actions.frequency", frequencyExclusivityDetail)
}

func ruleLevelNotifyWhenExclusive(v types.String) bool {
	return typeutils.IsKnown(v) && !v.IsNull() && strings.TrimSpace(v.ValueString()) != ""
}

func ruleLevelThrottleExclusive(v basetypes.StringValue) bool {
	return typeutils.IsKnown(v) && !v.IsNull() && strings.TrimSpace(v.ValueString()) != ""
}

func configActionsIncludeKnownFrequencyBlock(ctx context.Context, actions types.List, diags *diag.Diagnostics) bool {
	if !typeutils.IsKnown(actions) || actions.IsNull() {
		return false
	}
	var elems []actionModel
	diags.Append(actions.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return false
	}
	for _, a := range elems {
		if typeutils.IsKnown(a.Frequency) && !a.Frequency.IsNull() {
			return true
		}
	}
	return false
}

// planNotifyWhenForActionFrequency implements REQ-041: when planned top-level notify_when is unknown
// and configuration includes a non-null action frequency block, the planned value becomes null so
// later plan modifiers (e.g. UseStateForUnknown) do not reintroduce rule-level notify_when from state alone.
func planNotifyWhenForActionFrequency(ctx context.Context, planValue types.String, actions types.List, diags *diag.Diagnostics) types.String {
	if !planValue.IsUnknown() {
		return planValue
	}
	if !configActionsIncludeKnownFrequencyBlock(ctx, actions, diags) {
		return planValue
	}
	return types.StringNull()
}

// notifyWhenNullIfUnknownWithActionFrequency is registered on notify_when before UseStateForUnknown (REQ-041).
func notifyWhenNullIfUnknownWithActionFrequency() planmodifier.String {
	return notifyWhenActionFrequencyModifier{}
}

type notifyWhenActionFrequencyModifier struct{}

func (notifyWhenActionFrequencyModifier) Description(context.Context) string {
	return "Sets top-level notify_when to null when it is unknown in the plan and an action frequency block is present in config (REQ-041, before UseStateForUnknown)"
}

func (m notifyWhenActionFrequencyModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m notifyWhenActionFrequencyModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var actions types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("actions"), &actions)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.PlanValue = planNotifyWhenForActionFrequency(ctx, req.PlanValue, actions, &resp.Diagnostics)
}

// throttleShouldResetForActionFrequency is the predicate driving the
// planmodifiers.StringSetUnknownIf modifier registered on rule-level throttle
// after UseStateForUnknown. It is the StringRequest/StringResponse adapter
// over actionFrequencyNewlyIntroduced.
//
// USFU restores the Kibana-preserved throttle when the practitioner removes
// it from configuration (Kibana's PUT cannot clear deprecated rule-level
// throttle). The predicate returns true — i.e. resets the plan back to
// Unknown — only when the practitioner is newly introducing an
// actions[*].frequency block (config has one, prior state does not). On
// subsequent plans where the rule is already in frequency mode (state also
// has actions[*].frequency), the predicate returns false so the modifier is
// idempotent and does not produce a perpetual "value → (known after apply)"
// diff against the Kibana-preserved throttle.
//
// When true, the surrounding StringSetUnknownIf wrapper omits throttle from
// the API request and lets state resolve to whatever Kibana returns; combined
// with the config-time REQ-042 exclusivity check, the stale throttle is never
// silently sent alongside per-action frequency.
func throttleShouldResetForActionFrequency(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) bool {
	var configActions types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("actions"), &configActions)...)
	if resp.Diagnostics.HasError() {
		return false
	}

	var stateActions types.List
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("actions"), &stateActions)...)
	if resp.Diagnostics.HasError() {
		return false
	}

	return actionFrequencyNewlyIntroduced(ctx, stateActions, configActions, &resp.Diagnostics)
}

// actionFrequencyNewlyIntroduced returns true when the configuration includes
// an actions[*].frequency block AND the prior state did not. Extracted from
// throttleShouldResetForActionFrequency to allow direct unit testing without
// fabricating tfsdk.Config / tfsdk.State values, mirroring the existing
// planNotifyWhenForActionFrequency pattern in this file.
func actionFrequencyNewlyIntroduced(ctx context.Context, stateActions, configActions types.List, diags *diag.Diagnostics) bool {
	if !configActionsIncludeKnownFrequencyBlock(ctx, configActions, diags) {
		return false
	}
	if configActionsIncludeKnownFrequencyBlock(ctx, stateActions, diags) {
		return false
	}
	return true
}

// throttleSetUnknownIfActionsFrequencyConfigured returns the throttle plan
// modifier described by throttleShouldResetForActionFrequency, wired through
// the reusable planmodifiers.StringSetUnknownIf helper.
func throttleSetUnknownIfActionsFrequencyConfigured() planmodifier.String {
	return planmodifiers.StringSetUnknownIf(
		"Resets the planned rule-level throttle to unknown when actions[*].frequency is being newly introduced, "+
			"so the value preserved from prior state by UseStateForUnknown is not sent alongside per-action frequency. "+
			"No-op once the rule is already in frequency mode to keep the modifier idempotent.",
		throttleShouldResetForActionFrequency,
	)
}
