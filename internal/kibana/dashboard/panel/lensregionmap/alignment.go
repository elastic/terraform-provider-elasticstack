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

package lensregionmap

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// alignRegionMapStateFromPlan is the canonical region_map alignment invoked via VizConverter.AlignStateFromPlan.
func alignRegionMapStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	if plan == nil || state == nil {
		return
	}
	alignRegionMapConfigStateFromPlan(ctx, plan.RegionMapConfig, state.RegionMapConfig)
}

func alignRegionMapConfigStateFromPlan(ctx context.Context, plan, state *models.RegionMapConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	preservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
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

func preservePlanJSONWithDefaultsIfSemanticallyEqual(ctx context.Context, plan customtypes.JSONWithDefaultsValue[map[string]any], state *customtypes.JSONWithDefaultsValue[map[string]any]) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	eq, diags := plan.StringSemanticEquals(ctx, *state)
	if !diags.HasError() && eq {
		*state = plan
	}
}
