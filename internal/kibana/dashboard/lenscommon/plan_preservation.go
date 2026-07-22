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

package lenscommon

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PreserveKnownTfValueIfStateNull copies plan into *state when plan is known but
// state is null or unknown. Used to preserve practitioner intent across chart config round-trips.
func PreserveKnownTfValueIfStateNull[T attr.Value](plan T, state *T) {
	if typeutils.IsKnown(plan) && !typeutils.IsKnown(*state) {
		*state = plan
	}
}

// PreserveKnownStringIfStateBlank copies plan into *state when plan is known and state is null,
// unknown, or empty. Used to preserve practitioner intent for chart titles and descriptions
// that the API normalizes to empty values.
func PreserveKnownStringIfStateBlank(plan types.String, state *types.String) {
	if !typeutils.IsKnown(plan) {
		return
	}
	if state.IsNull() || state.IsUnknown() || state.ValueString() == "" {
		*state = plan
	}
}

// AlignTitleAndDescriptionFromPlan applies PreserveKnownStringIfStateBlank to a chart's
// title and description fields in one call.
func AlignTitleAndDescriptionFromPlan(planTitle, planDescription types.String, stateTitle, stateDescription *types.String) {
	PreserveKnownStringIfStateBlank(planTitle, stateTitle)
	PreserveKnownStringIfStateBlank(planDescription, stateDescription)
}

// PreservePlanJSONWithDefaultsIfSemanticallyEqual replaces *state with plan when both are known
// and StringSemanticEquals reports them equal. Lets practitioners keep their plan formatting
// when only whitespace or default-key ordering differs from the API response.
func PreservePlanJSONWithDefaultsIfSemanticallyEqual[T any](ctx context.Context, plan customtypes.JSONWithDefaultsValue[T], state *customtypes.JSONWithDefaultsValue[T]) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}
	eq, diags := plan.StringSemanticEquals(ctx, *state)
	if !diags.HasError() && eq {
		*state = plan
	}
}

// PreserveNormalizedJSONSemanticEquality replaces state with plan when normalized structures match semantically.
func PreserveNormalizedJSONSemanticEquality(plan jsontypes.Normalized, state *jsontypes.Normalized) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	var planObj map[string]any
	if err := json.Unmarshal([]byte(plan.ValueString()), &planObj); err != nil {
		return
	}
	var stateObj map[string]any
	if err := json.Unmarshal([]byte(state.ValueString()), &stateObj); err != nil {
		return
	}

	if reflect.DeepEqual(NormalizeXYPlanComparisonJSON(planObj), NormalizeXYPlanComparisonJSON(stateObj)) {
		*state = plan
	}
}

// PreservePlanJSONIfStateAddsOptionalKeys treats optional keys Kibana adds on read as semantically absent when the practitioner omitted them.
func PreservePlanJSONIfStateAddsOptionalKeys(plan jsontypes.Normalized, state *jsontypes.Normalized, optionalKeys ...string) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	var planObj map[string]any
	if err := json.Unmarshal([]byte(plan.ValueString()), &planObj); err != nil {
		return
	}
	var stateObj map[string]any
	if err := json.Unmarshal([]byte(state.ValueString()), &stateObj); err != nil {
		return
	}

	for _, key := range optionalKeys {
		if _, hasPlan := planObj[key]; hasPlan {
			continue
		}
		delete(stateObj, key)
	}

	stateNormalized := NormalizeXYPlanComparisonJSON(stateObj)
	planNormalized := NormalizeXYPlanComparisonJSON(planObj)
	if reflect.DeepEqual(stateNormalized, planNormalized) {
		*state = plan
	}
}

// PreserveNullJSONIfStateMatchesDefault preserves a null plan value when the API read-back
// returned the supplied default JSON payload. Use this for typed JSON attributes (e.g.
// gauge `styling.shape_json`) that Kibana auto-populates with a fixed default object when the
// practitioner omitted the field. The default is provided as a raw JSON string and is
// compared against state using NormalizeXYPlanComparisonJSON so that key order and Lens
// number-format defaults do not cause false negatives.
func PreserveNullJSONIfStateMatchesDefault(plan jsontypes.Normalized, state *jsontypes.Normalized, defaultJSON string) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if !typeutils.IsKnown(*state) {
		return
	}

	var defaultObj any
	if err := json.Unmarshal([]byte(defaultJSON), &defaultObj); err != nil {
		return
	}
	var stateObj any
	if err := json.Unmarshal([]byte(state.ValueString()), &stateObj); err != nil {
		return
	}

	if reflect.DeepEqual(NormalizeXYPlanComparisonJSON(defaultObj), NormalizeXYPlanComparisonJSON(stateObj)) {
		*state = plan
	}
}

// PreservePlanJSONIfStateOmitsOptionalKeys mirrors PreservePlanJSONIfStateAddsOptionalKeys when state drops optional keys.
func PreservePlanJSONIfStateOmitsOptionalKeys(plan jsontypes.Normalized, state *jsontypes.Normalized, optionalKeys ...string) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	var planObj map[string]any
	if err := json.Unmarshal([]byte(plan.ValueString()), &planObj); err != nil {
		return
	}
	var stateObj map[string]any
	if err := json.Unmarshal([]byte(state.ValueString()), &stateObj); err != nil {
		return
	}

	for _, key := range optionalKeys {
		if _, hasState := stateObj[key]; hasState {
			continue
		}
		delete(planObj, key)
	}

	stateNormalized := NormalizeXYPlanComparisonJSON(stateObj)
	planNormalized := NormalizeXYPlanComparisonJSON(planObj)
	if reflect.DeepEqual(stateNormalized, planNormalized) {
		*state = plan
	}
}

// PreservePlanNormalizedJSONWithDefaultsIfSemanticallyEqual preserves normalized JSON from plan when state only adds structurally
// different defaults that normalize to the same comparison shape (used by XY layer alignment).
func PreservePlanNormalizedJSONWithDefaultsIfSemanticallyEqual[T any](plan jsontypes.Normalized, state *jsontypes.Normalized, defaults func(T) T) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	var planObj T
	if err := json.Unmarshal([]byte(plan.ValueString()), &planObj); err != nil {
		return
	}
	var stateObj T
	if err := json.Unmarshal([]byte(state.ValueString()), &stateObj); err != nil {
		return
	}

	planNormalized := NormalizeXYPlanComparisonJSON(defaults(planObj))
	stateNormalized := NormalizeXYPlanComparisonJSON(defaults(stateObj))
	if reflect.DeepEqual(planNormalized, stateNormalized) {
		*state = plan
	}
}

// PreserveNullStringIfStateEquals copies a null plan value back into state when the API
// read-back returned the supplied default. Use this for optional typed string attributes
// (e.g. `tagcloud.orientation`, `pie.label_position`) that Kibana auto-populates with a
// hard-coded default when the practitioner omitted the field. Without this, the
// inconsistent plan/state values would surface as "Provider produced inconsistent result
// after apply" diagnostics.
func PreserveNullStringIfStateEquals(plan types.String, state *types.String, expected string) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueString() == expected {
		*state = plan
	}
}

// PreserveNullBoolIfStateEquals mirrors PreserveNullStringIfStateEquals for bool attributes.
// See PreserveNullStringIfStateEquals.
func PreserveNullBoolIfStateEquals(plan types.Bool, state *types.Bool, expected bool) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueBool() == expected {
		*state = plan
	}
}

// PreserveNullInt64IfStateEquals mirrors PreserveNullStringIfStateEquals for int64 attributes.
// See PreserveNullStringIfStateEquals.
func PreserveNullInt64IfStateEquals(plan types.Int64, state *types.Int64, expected int64) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueInt64() == expected {
		*state = plan
	}
}

// PreserveNullFloat64IfStateEquals mirrors PreserveNullStringIfStateEquals for float64 attributes.
// See PreserveNullStringIfStateEquals.
func PreserveNullFloat64IfStateEquals(plan types.Float64, state *types.Float64, expected float64) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueFloat64() == expected {
		*state = plan
	}
}
