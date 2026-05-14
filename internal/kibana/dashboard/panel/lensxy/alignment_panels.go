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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	if pm.LensDashboardAppConfig != nil && pm.LensDashboardAppConfig.ByValue != nil &&
		pm.LensDashboardAppConfig.ByValue.XYChartConfig != nil {
		return pm.LensDashboardAppConfig.ByValue.XYChartConfig
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
	alignXYLegendStateFromPlan(plan.Legend, state.Legend)
	alignXYLayerStateFromPlan(plan.Layers, state.Layers)
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

	preserveNullBoolIfStateEquals(plan.Grid, &state.Grid, true)
	preserveNullBoolIfStateEquals(plan.Ticks, &state.Ticks, true)
	preserveNullStringIfStateEquals(plan.LabelOrientation, &state.LabelOrientation, "horizontal")
	preserveNullStringIfStateEquals(plan.Scale, &state.Scale, string(kbapi.VisApiXyAxisConfigXScaleOrdinal))
	preserveKnownBoolIfStateNull(plan.Grid, &state.Grid)
	preserveKnownBoolIfStateNull(plan.Ticks, &state.Ticks)
	preserveKnownStringIfStateNull(plan.LabelOrientation, &state.LabelOrientation)
	preserveKnownStringIfStateNull(plan.Scale, &state.Scale)
	// When axis.title is omitted from config, suppress any server-filled defaults.
	if plan.Title == nil {
		state.Title = nil
	}
	preserveKnownAxisTitleIfStateBlank(plan.Title, &state.Title)
	preserveNullJSONIfStateMatches(plan.DomainJSON, &state.DomainJSON, `{"type":"fit","rounding":false}`)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.DomainJSON, &state.DomainJSON, "rounding")
}

func alignXYYAxisStateFromPlan(plan, state *models.YAxisConfigModel) {
	if plan == nil || state == nil {
		return
	}

	preserveNullBoolIfStateEquals(plan.Grid, &state.Grid, true)
	preserveNullBoolIfStateEquals(plan.Ticks, &state.Ticks, true)
	preserveNullStringIfStateEquals(plan.LabelOrientation, &state.LabelOrientation, "horizontal")
	preserveKnownBoolIfStateNull(plan.Grid, &state.Grid)
	preserveKnownBoolIfStateNull(plan.Ticks, &state.Ticks)
	preserveKnownStringIfStateNull(plan.LabelOrientation, &state.LabelOrientation)
	preserveKnownStringIfStateNull(plan.Scale, &state.Scale)
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

	preserveKnownBoolIfStateNull(plan.Grid, &state.Grid)
	preserveKnownBoolIfStateNull(plan.Ticks, &state.Ticks)
	preserveKnownStringIfStateNull(plan.LabelOrientation, &state.LabelOrientation)
	preserveKnownStringIfStateNull(plan.Scale, &state.Scale)
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

	preserveNullBoolIfStateEquals(plan.ShowEndZones, &state.ShowEndZones, false)
	preserveNullBoolIfStateEquals(plan.ShowCurrentTimeMarker, &state.ShowCurrentTimeMarker, false)
	preserveNullStringIfStateEquals(plan.PointVisibility, &state.PointVisibility, "auto")
	preserveNullStringIfStateEquals(plan.LineInterpolation, &state.LineInterpolation, "linear")
	preserveKnownBoolIfStateNull(plan.ShowEndZones, &state.ShowEndZones)
	preserveKnownBoolIfStateNull(plan.ShowCurrentTimeMarker, &state.ShowCurrentTimeMarker)
	preserveKnownStringIfStateNull(plan.PointVisibility, &state.PointVisibility)
	preserveKnownStringIfStateNull(plan.LineInterpolation, &state.LineInterpolation)
	preserveKnownInt64IfStateNull(plan.MinimumBarHeight, &state.MinimumBarHeight)
	preserveKnownBoolIfStateNull(plan.ShowValueLabels, &state.ShowValueLabels)
	preserveKnownFloat64IfStateNull(plan.FillOpacity, &state.FillOpacity)
}

func alignXYLegendStateFromPlan(plan, state *models.XYLegendModel) {
	if plan == nil || state == nil {
		return
	}

	preserveNullInt64IfStateEquals(plan.TruncateAfterLines, &state.TruncateAfterLines, 1)
	preserveKnownStringIfStateNull(plan.Visibility, &state.Visibility)
	preserveKnownBoolIfStateNull(plan.Inside, &state.Inside)
	preserveKnownStringIfStateNull(plan.Position, &state.Position)
	preserveKnownStringIfStateNull(plan.Size, &state.Size)
	preserveKnownInt64IfStateNull(plan.Columns, &state.Columns)
	preserveKnownStringIfStateNull(plan.Alignment, &state.Alignment)
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
	preserveKnownBoolIfStateNull(plan.Visible, &(*state).Visible)
}

func preserveKnownStringIfStateNull(plan types.String, state *types.String) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func preserveKnownBoolIfStateNull(plan types.Bool, state *types.Bool) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func preserveKnownInt64IfStateNull(plan types.Int64, state *types.Int64) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func preserveKnownFloat64IfStateNull(plan types.Float64, state *types.Float64) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func preserveNullStringIfStateEquals(plan types.String, state *types.String, expected string) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueString() == expected {
		*state = plan
	}
}

func preserveNullBoolIfStateEquals(plan types.Bool, state *types.Bool, expected bool) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueBool() == expected {
		*state = plan
	}
}

func preserveNullInt64IfStateEquals(plan types.Int64, state *types.Int64, expected int64) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueInt64() == expected {
		*state = plan
	}
}

func preserveNullJSONIfStateMatches(plan jsontypes.Normalized, state *jsontypes.Normalized, expected string) {
	if !plan.IsNull() || plan.IsUnknown() || !typeutils.IsKnown(*state) {
		return
	}
	expectedNormalized := jsontypes.NewNormalizedValue(expected)
	if state.ValueString() == expectedNormalized.ValueString() {
		*state = plan
	}
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
