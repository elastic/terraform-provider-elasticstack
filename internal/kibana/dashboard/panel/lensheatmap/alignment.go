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

package lensheatmap

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func alignHeatmapStateFromPlan(ctx context.Context, plan, state *models.HeatmapConfigModel) {
	if plan == nil || state == nil {
		return
	}
	lenscommon.AlignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	lenscommon.PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
	alignHeatmapLegendStateFromPlan(plan.Legend, &state.Legend)
}

func alignHeatmapLegendStateFromPlan(plan *models.HeatmapLegendModel, state **models.HeatmapLegendModel) {
	if plan == nil {
		return
	}
	if *state == nil || heatmapLegendEffectivelyUnset(*state) {
		*state = cloneHeatmapLegendModel(plan)
		return
	}
	preserveHeatmapLegendStringIfStateNull(plan.Visibility, &(*state).Visibility)
	preserveHeatmapLegendStringIfStateNull(plan.Size, &(*state).Size)
	preserveHeatmapLegendInt64IfStateNull(plan.TruncateAfterLines, &(*state).TruncateAfterLines)
}

func heatmapLegendEffectivelyUnset(m *models.HeatmapLegendModel) bool {
	if m == nil {
		return true
	}
	return !typeutils.IsKnown(m.Visibility) && !typeutils.IsKnown(m.Size) && !typeutils.IsKnown(m.TruncateAfterLines)
}

func preserveHeatmapLegendStringIfStateNull(plan types.String, state *types.String) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func preserveHeatmapLegendInt64IfStateNull(plan types.Int64, state *types.Int64) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func cloneHeatmapLegendModel(model *models.HeatmapLegendModel) *models.HeatmapLegendModel {
	if model == nil {
		return nil
	}
	cloned := *model
	return &cloned
}
