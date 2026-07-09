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

package lensxy

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
)

// alignXYChartStateFromPlanPanels preserves practitioner intent for XY charts when Kibana
// injects implicit defaults on read or omits configured fields from the response.
func AlignXYChartStateFromPlanPanels(planPanels, statePanels []models.PanelModel) {
	n := min(len(statePanels), len(planPanels))
	for i := range n {
		pp, sp := xyChartConfigFromLensOrVisPlanPanel(planPanels[i]), xyChartConfigFromLensOrVisPlanPanel(statePanels[i])
		if pp == nil || sp == nil {
			continue
		}
		alignXYChartStateFromPlan(pp, sp)
	}
}

func xyChartConfigFromLensOrVisPlanPanel(pm models.PanelModel) *models.XYChartConfigModel {
	if pm.VisConfig != nil && pm.VisConfig.ByValue != nil && pm.VisConfig.ByValue.XYChartConfig != nil {
		return pm.VisConfig.ByValue.XYChartConfig
	}
	return nil
}

func alignXYChartStateFromPlan(plan, state *models.XYChartConfigModel) {
	if plan == nil || state == nil {
		return
	}

	lenscommon.AlignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)

	alignXYAxisStateFromPlan(plan.Axis, state.Axis)
	alignXYDecorationsStateFromPlan(plan.Decorations, state.Decorations)
	alignXYFittingStateFromPlan(plan.Fitting, state.Fitting)
	if plan.Legend != nil && (state.Legend == nil || xyLegendEffectivelyUnset(state.Legend)) {
		state.Legend = cloneXYLegendModel(plan.Legend)
	} else {
		alignXYLegendStateFromPlan(plan.Legend, state.Legend)
	}
	alignXYLayerStateFromPlan(plan.Layers, state.Layers)
}

func alignXYFittingStateFromPlan(plan, state *models.XYFittingModel) {
	if plan == nil || state == nil {
		return
	}

	// Kibana omits fitting for some XY chart kinds (e.g. bar_horizontal with terms).
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Type, &state.Type)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Dotted, &state.Dotted)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.EndValue, &state.EndValue)
}

func alignXYAxisStateFromPlan(plan, state *models.XYAxisModel) {
	if plan == nil || state == nil {
		return
	}

	// When the user omits axis.x entirely, suppress any server-filled defaults.
	if plan.X == nil {
		state.X = nil
	} else {
		alignXYXAxisStateFromPlan(plan.X, state.X)
	}
	alignXYYAxisStateFromPlan(plan.Y, state.Y)

	if plan.Y2 != nil && state.Y2 == nil {
		state.Y2 = cloneYAxisConfigModel(plan.Y2)
		return
	}
	alignXYY2AxisStateFromPlan(plan.Y2, state.Y2)
}

func alignXYXAxisStateFromPlan(plan, state *models.XYAxisConfigModel) {
	if plan == nil || state == nil {
		return
	}

	lenscommon.PreserveNullBoolIfStateEquals(plan.Grid, &state.Grid, true)
	lenscommon.PreserveNullBoolIfStateEquals(plan.Ticks, &state.Ticks, true)
	lenscommon.PreserveNullStringIfStateEquals(plan.LabelOrientation, &state.LabelOrientation, "horizontal")
	lenscommon.PreserveNullStringIfStateEquals(plan.Scale, &state.Scale, string(kbapi.KibanaHTTPAPIsVisApiXyAxisConfigXScaleOrdinal))
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Grid, &state.Grid)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Ticks, &state.Ticks)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.LabelOrientation, &state.LabelOrientation)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Scale, &state.Scale)
	// When axis.title is omitted from config, suppress any server-filled defaults.
	if plan.Title == nil {
		state.Title = nil
	}
	preserveKnownAxisTitleIfStateBlank(plan.Title, &state.Title)
	lenscommon.PreserveNullJSONIfStateMatchesDefault(plan.DomainJSON, &state.DomainJSON, `{"type":"fit","rounding":false}`)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.DomainJSON, &state.DomainJSON, "rounding")
}

func alignXYYAxisStateFromPlan(plan, state *models.YAxisConfigModel) {
	if plan == nil || state == nil {
		return
	}

	lenscommon.PreserveNullBoolIfStateEquals(plan.Grid, &state.Grid, true)
	lenscommon.PreserveNullBoolIfStateEquals(plan.Ticks, &state.Ticks, true)
	lenscommon.PreserveNullStringIfStateEquals(plan.LabelOrientation, &state.LabelOrientation, "horizontal")
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Grid, &state.Grid)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Ticks, &state.Ticks)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.LabelOrientation, &state.LabelOrientation)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Scale, &state.Scale)
	// When axis.title is omitted from config, suppress any server-filled defaults.
	if plan.Title == nil {
		state.Title = nil
	}
	preserveKnownAxisTitleIfStateBlank(plan.Title, &state.Title)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.DomainJSON, &state.DomainJSON, "rounding")
}

func alignXYY2AxisStateFromPlan(plan, state *models.YAxisConfigModel) {
	if plan == nil || state == nil {
		return
	}

	lenscommon.PreserveNullBoolIfStateEquals(plan.Grid, &state.Grid, true)
	lenscommon.PreserveNullBoolIfStateEquals(plan.Ticks, &state.Ticks, true)
	lenscommon.PreserveNullStringIfStateEquals(plan.LabelOrientation, &state.LabelOrientation, "horizontal")
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Grid, &state.Grid)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Ticks, &state.Ticks)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.LabelOrientation, &state.LabelOrientation)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Scale, &state.Scale)
	// When axis.title is omitted from config, suppress any server-filled defaults.
	if plan.Title == nil {
		state.Title = nil
	}
	preserveKnownAxisTitleIfStateBlank(plan.Title, &state.Title)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.DomainJSON, &state.DomainJSON, "rounding")
}

func alignXYDecorationsStateFromPlan(plan, state *models.XYDecorationsModel) {
	if plan == nil || state == nil {
		return
	}

	lenscommon.PreserveNullBoolIfStateEquals(plan.ShowEndZones, &state.ShowEndZones, false)
	lenscommon.PreserveNullBoolIfStateEquals(plan.ShowCurrentTimeMarker, &state.ShowCurrentTimeMarker, false)
	lenscommon.PreserveNullStringIfStateEquals(plan.PointVisibility, &state.PointVisibility, "auto")
	lenscommon.PreserveNullStringIfStateEquals(plan.LineInterpolation, &state.LineInterpolation, "linear")
	// Kibana injects bar-styling defaults (show_value_labels=false,
	// minimum_bar_height=1) for bar/bar_stacked layers even when the
	// practitioner omits decorations. Preserve the null plan so the apply
	// matches and no spurious drift appears on subsequent plans.
	lenscommon.PreserveNullBoolIfStateEquals(plan.ShowValueLabels, &state.ShowValueLabels, false)
	lenscommon.PreserveNullInt64IfStateEquals(plan.MinimumBarHeight, &state.MinimumBarHeight, 1)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.ShowEndZones, &state.ShowEndZones)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.ShowCurrentTimeMarker, &state.ShowCurrentTimeMarker)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.PointVisibility, &state.PointVisibility)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.LineInterpolation, &state.LineInterpolation)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.MinimumBarHeight, &state.MinimumBarHeight)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.ShowValueLabels, &state.ShowValueLabels)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.FillOpacity, &state.FillOpacity)
}

func alignXYLegendStateFromPlan(plan, state *models.XYLegendModel) {
	if plan == nil || state == nil {
		return
	}

	lenscommon.PreserveNullInt64IfStateEquals(plan.TruncateAfterLines, &state.TruncateAfterLines, 1)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Visibility, &state.Visibility)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Inside, &state.Inside)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Position, &state.Position)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Size, &state.Size)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Columns, &state.Columns)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Alignment, &state.Alignment)
}

func alignXYLayerStateFromPlan(planLayers, stateLayers []models.XYLayerModel) {
	n := min(len(stateLayers), len(planLayers))
	for i := range n {
		planLayer, stateLayer := planLayers[i], &stateLayers[i]
		if planLayer.DataLayer != nil && stateLayer.DataLayer != nil {
			lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.DataSourceJSON, &stateLayer.DataLayer.DataSourceJSON, "time_field")
			lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.XJSON, &stateLayer.DataLayer.XJSON)
			lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.BreakdownByJSON, &stateLayer.DataLayer.BreakdownByJSON)
			lenscommon.PreservePlanNormalizedJSONWithDefaultsIfSemanticallyEqual(planLayer.DataLayer.BreakdownByJSON, &stateLayer.DataLayer.BreakdownByJSON, lenscommon.PopulateLensGroupByDefaults)

			m := min(len(stateLayer.DataLayer.Y), len(planLayer.DataLayer.Y))
			for j := range m {
				lenscommon.PreservePlanJSONIfStateOmitsOptionalKeys(planLayer.DataLayer.Y[j].ConfigJSON, &stateLayer.DataLayer.Y[j].ConfigJSON, "color")
				lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.Y[j].ConfigJSON, &stateLayer.DataLayer.Y[j].ConfigJSON, "axis_id")
				lenscommon.PreservePlanNormalizedJSONWithDefaultsIfSemanticallyEqual(planLayer.DataLayer.Y[j].ConfigJSON, &stateLayer.DataLayer.Y[j].ConfigJSON, lenscommon.PopulateLensMetricDefaults)
			}
		}

		if planLayer.ReferenceLineLayer == nil || stateLayer.ReferenceLineLayer == nil {
			continue
		}

		lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.ReferenceLineLayer.DataSourceJSON, &stateLayer.ReferenceLineLayer.DataSourceJSON, "time_field")
		m := min(len(stateLayer.ReferenceLineLayer.Thresholds), len(planLayer.ReferenceLineLayer.Thresholds))
		for j := range m {
			lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.ReferenceLineLayer.Thresholds[j].ValueJSON, &stateLayer.ReferenceLineLayer.Thresholds[j].ValueJSON, "axis_id", "color")
		}
	}
}

func preserveKnownAxisTitleIfStateBlank(plan *models.AxisTitleModel, state **models.AxisTitleModel) {
	if plan == nil {
		return
	}
	if *state == nil {
		*state = cloneAxisTitleModel(plan)
		return
	}

	lenscommon.PreserveKnownStringIfStateBlank(plan.Value, &(*state).Value)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Visible, &(*state).Visible)
}

func cloneAxisTitleModel(model *models.AxisTitleModel) *models.AxisTitleModel {
	if model == nil {
		return nil
	}
	cloned := *model
	return &cloned
}

func cloneYAxisConfigModel(model *models.YAxisConfigModel) *models.YAxisConfigModel {
	if model == nil {
		return nil
	}
	cloned := *model
	cloned.Title = cloneAxisTitleModel(model.Title)
	return &cloned
}

func xyLegendEffectivelyUnset(m *models.XYLegendModel) bool {
	if m == nil {
		return true
	}
	return !typeutils.IsKnown(m.Visibility) &&
		!typeutils.IsKnown(m.Position) &&
		!typeutils.IsKnown(m.Size) &&
		!typeutils.IsKnown(m.Inside) &&
		!typeutils.IsKnown(m.Alignment) &&
		!typeutils.IsKnown(m.Columns) &&
		!typeutils.IsKnown(m.TruncateAfterLines) &&
		(m.Statistics.IsNull() || m.Statistics.IsUnknown())
}

func cloneXYLegendModel(model *models.XYLegendModel) *models.XYLegendModel {
	if model == nil {
		return nil
	}
	cloned := *model
	return &cloned
}
