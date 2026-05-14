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

package lenswaffle

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func alignWaffleStateFromPlan(ctx context.Context, plan, state *models.WaffleConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	m := min(len(plan.Metrics), len(state.Metrics))
	for i := range m {
		preservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.Metrics[i].Config, &state.Metrics[i].Config)
	}
	g := min(len(plan.GroupBy), len(state.GroupBy))
	for i := range g {
		preservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.GroupBy[i].Config, &state.GroupBy[i].Config)
	}
}

func preservePlanJSONWithDefaultsIfSemanticallyEqual[T any](ctx context.Context, plan customtypes.JSONWithDefaultsValue[T], state *customtypes.JSONWithDefaultsValue[T]) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	eq, diags := plan.StringSemanticEquals(ctx, *state)
	if !diags.HasError() && eq {
		*state = plan
	}
}

func alignTitleAndDescriptionFromPlan(planTitle, planDescription types.String, stateTitle, stateDescription *types.String) {
	preserveKnownStringIfStateBlank(planTitle, stateTitle)
	preserveKnownStringIfStateBlank(planDescription, stateDescription)
}

func preserveKnownStringIfStateBlank(plan types.String, state *types.String) {
	if !typeutils.IsKnown(plan) {
		return
	}
	if state.IsNull() || state.IsUnknown() || state.ValueString() == "" {
		*state = plan
	}
}
