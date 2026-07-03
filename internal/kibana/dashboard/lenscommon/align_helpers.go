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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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

// PreserveKnownTfBoolIfStateNull copies plan into *state when plan is known but
// state is null or unknown. Used to preserve practitioner intent across chart config round-trips.
func PreserveKnownTfBoolIfStateNull(plan types.Bool, state *types.Bool) {
	PreserveKnownTfValueIfStateNull(plan, state)
}

// PreserveKnownTfFloat64IfStateNull copies plan into *state when plan is known but
// state is null or unknown. Used to preserve practitioner intent across chart config round-trips.
func PreserveKnownTfFloat64IfStateNull(plan types.Float64, state *types.Float64) {
	PreserveKnownTfValueIfStateNull(plan, state)
}

// PreserveKnownTfStringIfStateNull copies plan into *state when plan is known but
// state is null or unknown. Used to preserve practitioner intent across chart config round-trips.
func PreserveKnownTfStringIfStateNull(plan types.String, state *types.String) {
	PreserveKnownTfValueIfStateNull(plan, state)
}

// PreserveKnownTfInt64IfStateNull copies plan into *state when plan is known but
// state is null or unknown. Used to preserve practitioner intent across chart config round-trips.
func PreserveKnownTfInt64IfStateNull(plan types.Int64, state *types.Int64) {
	PreserveKnownTfValueIfStateNull(plan, state)
}

// PreserveKnownTfListIfStateNull copies plan into *state when plan is known but
// state is null or unknown. Used to preserve practitioner intent across chart config round-trips.
func PreserveKnownTfListIfStateNull(plan types.List, state *types.List) {
	PreserveKnownTfValueIfStateNull(plan, state)
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

// SamplingFromAPIWithDefault converts a nullable API *float32 sampling value to types.Float64.
// Returns Float64Value(def) when s is nil.
func SamplingFromAPIWithDefault(s *float32, def float32) types.Float64 {
	if s == nil {
		s = &def
	}
	return typeutils.Float32PointerToFloat64Value(s)
}
