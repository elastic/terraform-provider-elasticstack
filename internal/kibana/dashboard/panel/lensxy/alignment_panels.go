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
	"encoding/json"
	"reflect"

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
		state.Legend = lenscommon.CloneModel(plan.Legend)
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
	alignXYYAxisStateFromPlan(plan.Y2, state.Y2)
}

func alignXYXAxisStateFromPlan(plan, state *models.XYAxisConfigModel) {
	if plan == nil || state == nil {
		return
	}

	preserveXYAxisCommonFieldsFromPlan(
		plan.Grid, &state.Grid,
		plan.Ticks, &state.Ticks,
		plan.LabelOrientation, &state.LabelOrientation,
		plan.Title, &state.Title,
		plan.DomainJSON, &state.DomainJSON,
	)
	lenscommon.PreserveNullIfStateEquals(plan.Scale, &state.Scale, types.StringValue(string(kbapi.KibanaHTTPAPIsVisApiXyAxisConfigXScaleOrdinal)))
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Scale, &state.Scale)
	lenscommon.PreserveNullJSONIfStateMatchesDefault(plan.DomainJSON, &state.DomainJSON, `{"type":"fit","rounding":false}`)
}

func alignXYYAxisStateFromPlan(plan, state *models.YAxisConfigModel) {
	if plan == nil || state == nil {
		return
	}

	preserveXYAxisCommonFieldsFromPlan(
		plan.Grid, &state.Grid,
		plan.Ticks, &state.Ticks,
		plan.LabelOrientation, &state.LabelOrientation,
		plan.Title, &state.Title,
		plan.DomainJSON, &state.DomainJSON,
	)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Scale, &state.Scale)
}

// preserveXYAxisCommonFieldsFromPlan applies the field-preservation defaults shared by the X and
// Y axis alignment functions: grid/ticks/label-orientation defaults, title nil-suppression, and
// the DomainJSON "rounding" key preservation. Axis-specific defaults (e.g. X's ordinal scale
// default, X's DomainJSON fit/rounding default) remain in the respective callers.
func preserveXYAxisCommonFieldsFromPlan(
	planGrid types.Bool, stateGrid *types.Bool,
	planTicks types.Bool, stateTicks *types.Bool,
	planLabelOrientation types.String, stateLabelOrientation *types.String,
	planTitle *models.AxisTitleModel, stateTitle **models.AxisTitleModel,
	planDomainJSON jsontypes.Normalized, stateDomainJSON *jsontypes.Normalized,
) {
	lenscommon.PreserveNullIfStateEquals(planGrid, stateGrid, types.BoolValue(true))
	lenscommon.PreserveNullIfStateEquals(planTicks, stateTicks, types.BoolValue(true))
	lenscommon.PreserveNullIfStateEquals(planLabelOrientation, stateLabelOrientation, types.StringValue("horizontal"))
	lenscommon.PreserveKnownTfValueIfStateNull(planGrid, stateGrid)
	lenscommon.PreserveKnownTfValueIfStateNull(planTicks, stateTicks)
	lenscommon.PreserveKnownTfValueIfStateNull(planLabelOrientation, stateLabelOrientation)
	// When axis.title is omitted from config, suppress any server-filled defaults.
	if planTitle == nil {
		*stateTitle = nil
	}
	preserveKnownAxisTitleIfStateBlank(planTitle, stateTitle)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planDomainJSON, stateDomainJSON, "rounding")
}

func alignXYDecorationsStateFromPlan(plan, state *models.XYDecorationsModel) {
	if plan == nil || state == nil {
		return
	}

	lenscommon.PreserveNullIfStateEquals(plan.ShowEndZones, &state.ShowEndZones, types.BoolValue(false))
	lenscommon.PreserveNullIfStateEquals(plan.ShowCurrentTimeMarker, &state.ShowCurrentTimeMarker, types.BoolValue(false))
	lenscommon.PreserveNullIfStateEquals(plan.PointVisibility, &state.PointVisibility, types.StringValue("auto"))
	lenscommon.PreserveNullIfStateEquals(plan.LineInterpolation, &state.LineInterpolation, types.StringValue("linear"))
	// Kibana injects bar-styling defaults (show_value_labels=false,
	// minimum_bar_height=1) for bar/bar_stacked layers even when the
	// practitioner omits decorations. Preserve the null plan so the apply
	// matches and no spurious drift appears on subsequent plans.
	lenscommon.PreserveNullIfStateEquals(plan.ShowValueLabels, &state.ShowValueLabels, types.BoolValue(false))
	lenscommon.PreserveNullIfStateEquals(plan.MinimumBarHeight, &state.MinimumBarHeight, types.Int64Value(1))
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

	lenscommon.PreserveNullIfStateEquals(plan.TruncateAfterLines, &state.TruncateAfterLines, types.Int64Value(1))
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
			lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.DataSourceJSON, &stateLayer.DataLayer.DataSourceJSON, "time_field", "name")
			lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.XJSON, &stateLayer.DataLayer.XJSON)
			lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.BreakdownByJSON, &stateLayer.DataLayer.BreakdownByJSON)
			lenscommon.PreservePlanNormalizedJSONWithDefaultsIfSemanticallyEqual(planLayer.DataLayer.BreakdownByJSON, &stateLayer.DataLayer.BreakdownByJSON, lenscommon.PopulateLensGroupByDefaults)

			m := min(len(stateLayer.DataLayer.Y), len(planLayer.DataLayer.Y))
			for j := range m {
				lenscommon.PreservePlanJSONIfStateOmitsOptionalKeys(planLayer.DataLayer.Y[j].ConfigJSON, &stateLayer.DataLayer.Y[j].ConfigJSON, "color")
				lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.Y[j].ConfigJSON, &stateLayer.DataLayer.Y[j].ConfigJSON, "axis_id")
				lenscommon.PreservePlanNormalizedJSONWithDefaultsIfSemanticallyEqual(planLayer.DataLayer.Y[j].ConfigJSON, &stateLayer.DataLayer.Y[j].ConfigJSON, lenscommon.PopulateXYMetricDefaults)
			}
		}

		if planLayer.ReferenceLineLayer == nil || stateLayer.ReferenceLineLayer == nil {
			continue
		}

		lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planLayer.ReferenceLineLayer.DataSourceJSON, &stateLayer.ReferenceLineLayer.DataSourceJSON, "time_field", "name")
		m := min(len(stateLayer.ReferenceLineLayer.Thresholds), len(planLayer.ReferenceLineLayer.Thresholds))
		for j := range m {
			planThreshold := planLayer.ReferenceLineLayer.Thresholds[j]
			stateThreshold := &stateLayer.ReferenceLineLayer.Thresholds[j]
			lenscommon.PreserveNullIfStateEquals(planThreshold.Axis, &stateThreshold.Axis, types.StringValue("y"))
			lenscommon.PreserveNullIfStateEquals(planThreshold.Operation, &stateThreshold.Operation, types.StringValue("static_value"))
			lenscommon.PreserveNullJSONIfStateMatchesDefault(planThreshold.ColorJSON, &stateThreshold.ColorJSON, `{"type":"auto"}`)
			preserveThresholdValueJSONIfStateIsPlanValue(planThreshold.ValueJSON, &stateThreshold.ValueJSON, planThreshold.Operation)
			lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(planThreshold.ValueJSON, &stateThreshold.ValueJSON, "axis_id", "color")
		}
	}
}

func preserveKnownAxisTitleIfStateBlank(plan *models.AxisTitleModel, state **models.AxisTitleModel) {
	if plan == nil {
		return
	}
	if *state == nil {
		*state = lenscommon.CloneModel(plan)
		return
	}

	lenscommon.PreserveKnownStringIfStateBlank(plan.Value, &(*state).Value)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Visible, &(*state).Visible)
}

func cloneYAxisConfigModel(model *models.YAxisConfigModel) *models.YAxisConfigModel {
	if model == nil {
		return nil
	}
	cloned := *lenscommon.CloneModel(model)
	cloned.Title = lenscommon.CloneModel(model.Title)
	return &cloned
}

func preserveThresholdValueJSONIfStateIsPlanValue(plan jsontypes.Normalized, state *jsontypes.Normalized, siblingOperation types.String) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	var planObj map[string]any
	if err := json.Unmarshal([]byte(plan.ValueString()), &planObj); err != nil {
		return
	}
	var stateVal any
	if err := json.Unmarshal([]byte(state.ValueString()), &stateVal); err != nil {
		return
	}
	if _, isMap := stateVal.(map[string]any); isMap {
		return
	}
	if !thresholdValueJSONIsStaticValue(planObj, siblingOperation) {
		return
	}
	planValue, hasValue := planObj["value"]
	if !hasValue {
		return
	}
	if reflect.DeepEqual(planValue, stateVal) {
		*state = plan
	}
}

func thresholdValueJSONIsStaticValue(planObj map[string]any, siblingOperation types.String) bool {
	if operation, hasOp := planObj["operation"]; hasOp {
		return operation == "static_value"
	}
	return typeutils.IsKnown(siblingOperation) && siblingOperation.ValueString() == "static_value"
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
