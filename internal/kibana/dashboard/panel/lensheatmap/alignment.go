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
)

func alignHeatmapStateFromPlan(ctx context.Context, plan, state *models.HeatmapConfigModel) {
	if plan == nil || state == nil {
		return
	}
	lenscommon.AlignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.DataSourceJSON, &state.DataSourceJSON, "time_field")
	lenscommon.PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
	alignHeatmapAxisStateFromPlan(plan.Axis, state.Axis)
	alignHeatmapStylingStateFromPlan(plan.Styling, state.Styling)
	alignHeatmapLegendStateFromPlan(plan.Legend, &state.Legend)
}

// Kibana materializes axis label/title visibility flags as true and cells
// labels visibility as false when omitted by the practitioner. Preserve the
// null plan values when the read-back matches those defaults.
func alignHeatmapAxisStateFromPlan(plan, state *models.HeatmapAxesModel) {
	if plan == nil || state == nil {
		return
	}
	if plan.X != nil && state.X != nil {
		if plan.X.Labels != nil && state.X.Labels != nil {
			lenscommon.PreserveNullBoolIfStateEquals(plan.X.Labels.Visible, &state.X.Labels.Visible, true)
		}
		// Kibana reports axis.title.visible=false when the practitioner omits it.
		if plan.X.Title != nil && state.X.Title != nil {
			lenscommon.PreserveNullBoolIfStateEquals(plan.X.Title.Visible, &state.X.Title.Visible, false)
		}
	}
	if plan.Y != nil && state.Y != nil {
		if plan.Y.Labels != nil && state.Y.Labels != nil {
			lenscommon.PreserveNullBoolIfStateEquals(plan.Y.Labels.Visible, &state.Y.Labels.Visible, true)
		}
		if plan.Y.Title != nil && state.Y.Title != nil {
			lenscommon.PreserveNullBoolIfStateEquals(plan.Y.Title.Visible, &state.Y.Title.Visible, false)
		}
	}
}

func alignHeatmapStylingStateFromPlan(plan, state *models.HeatmapStylingModel) {
	if plan == nil || state == nil {
		return
	}
	if plan.Cells != nil && state.Cells != nil && plan.Cells.Labels != nil && state.Cells.Labels != nil {
		lenscommon.PreserveNullBoolIfStateEquals(plan.Cells.Labels.Visible, &state.Cells.Labels.Visible, false)
	}
}

func alignHeatmapLegendStateFromPlan(plan *models.HeatmapLegendModel, state **models.HeatmapLegendModel) {
	if plan == nil {
		return
	}
	if *state == nil || heatmapLegendEffectivelyUnset(*state) {
		*state = cloneHeatmapLegendModel(plan)
		return
	}
	// Kibana renders the legend by default; preserve the null plan when the
	// API read-back returns the default "visible" value.
	lenscommon.PreserveNullStringIfStateEquals(plan.Visibility, &(*state).Visibility, "visible")
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Visibility, &(*state).Visibility)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Size, &(*state).Size)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.TruncateAfterLines, &(*state).TruncateAfterLines)
}

func heatmapLegendEffectivelyUnset(m *models.HeatmapLegendModel) bool {
	if m == nil {
		return true
	}
	return !typeutils.IsKnown(m.Visibility) && !typeutils.IsKnown(m.Size) && !typeutils.IsKnown(m.TruncateAfterLines)
}

func cloneHeatmapLegendModel(model *models.HeatmapLegendModel) *models.HeatmapLegendModel {
	if model == nil {
		return nil
	}
	cloned := *model
	return &cloned
}
