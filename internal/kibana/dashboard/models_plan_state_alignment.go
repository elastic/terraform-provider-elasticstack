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

package dashboard

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// alignDashboardStateFromPlanPanels preserves practitioner intent that depends on
// the full top-level panel slice. Common per-panel alignment already happens
// inside mapPanelFromAPI.
func alignDashboardStateFromPlanPanels(planPanels, statePanels []models.PanelModel) {
	alignXYChartStateFromPlanPanels(planPanels, statePanels)
}

func alignDashboardStateFromPlanPinnedPanels(ctx context.Context, planPins, statePins []models.PinnedPanelModel) {
	n := min(len(planPins), len(statePins))
	for i := range n {
		plan := planPins[i].SyntheticPanel()
		state := statePins[i].SyntheticPanel()
		alignPanelStateFromPlan(ctx, &plan, &state)
	}
}

func alignPanelStateFromPlan(ctx context.Context, plan, state *models.PanelModel) {
	if plan == nil || state == nil {
		return
	}

	preservePlanJSONIfStateOmitsOptionalKeys(plan.ConfigJSON.Normalized, &state.ConfigJSON.Normalized, "filters", "query", "settings")
	planBlocks := lensByValueChartBlocksFromPanel(plan)
	stateBlocks := lensByValueChartBlocksFromPanel(state)
	if planBlocks != nil && stateBlocks != nil {
		alignDatatableStateFromPlan(planBlocks.DatatableConfig, stateBlocks.DatatableConfig)
		alignGaugeStateFromPlan(ctx, planBlocks.GaugeConfig, stateBlocks.GaugeConfig)
		alignHeatmapStateFromPlan(ctx, planBlocks.HeatmapConfig, stateBlocks.HeatmapConfig)
		alignLegacyMetricStateFromPlan(ctx, planBlocks.LegacyMetricConfig, stateBlocks.LegacyMetricConfig)
		alignMetricStateFromPlan(planBlocks.MetricChartConfig, stateBlocks.MetricChartConfig)
		alignMosaicStateFromPlan(planBlocks.MosaicConfig, stateBlocks.MosaicConfig)
		alignPieStateFromPlan(planBlocks.PieChartConfig, stateBlocks.PieChartConfig)
		alignRegionMapStateFromPlan(ctx, planBlocks.RegionMapConfig, stateBlocks.RegionMapConfig)
		alignTagcloudStateFromPlan(ctx, planBlocks.TagcloudConfig, stateBlocks.TagcloudConfig)
		alignTreemapStateFromPlan(planBlocks.TreemapConfig, stateBlocks.TreemapConfig)
		alignWaffleStateFromPlan(ctx, planBlocks.WaffleConfig, stateBlocks.WaffleConfig)
	}
	if h := LookupHandler(state.Type.ValueString()); h != nil {
		h.AlignStateFromPlan(ctx, plan, state)
	}
}

func alignDatatableStateFromPlan(plan, state *models.DatatableConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignDatatableNoESQLStateFromPlan(plan.NoESQL, state.NoESQL)
	alignDatatableESQLStateFromPlan(plan.ESQL, state.ESQL)
}

func alignDatatableNoESQLStateFromPlan(plan, state *models.DatatableNoESQLConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignDatatableESQLStateFromPlan(plan, state *models.DatatableESQLConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignGaugeStateFromPlan(ctx context.Context, plan, state *models.GaugeConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	preservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
	if plan.EsqlMetric != nil && state.EsqlMetric != nil {
		alignGaugeEsqlMetricStateFromPlan(plan.EsqlMetric, state.EsqlMetric)
	}
}

func alignHeatmapStateFromPlan(ctx context.Context, plan, state *models.HeatmapConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	preservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
}

func alignLegacyMetricStateFromPlan(ctx context.Context, plan, state *models.LegacyMetricConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	preservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
}

func alignMetricStateFromPlan(plan, state *models.MetricChartConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	preservePlanJSONIfStateAddsOptionalKeys(plan.DataSourceJSON, &state.DataSourceJSON, "time_field")
	preservePlanJSONIfStateAddsOptionalKeys(plan.BreakdownByJSON, &state.BreakdownByJSON, "rank_by")
	m := min(len(plan.Metrics), len(state.Metrics))
	for i := range m {
		preserveMetricChartMetricConfigFromPlan(plan.Metrics[i].ConfigJSON, &state.Metrics[i].ConfigJSON)
	}
}

func alignMosaicStateFromPlan(plan, state *models.MosaicConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	// Lens API commonly omits optional snapshot fields Kibana echoes as absent on read while the practitioner
	// set them explicitly; mosaic fromAPI maps unknown "current" to null unless we replay plan here.
	preserveKnownTfBoolIfStateNull(plan.IgnoreGlobalFilters, &state.IgnoreGlobalFilters)
	preserveKnownTfFloat64IfStateNull(plan.Sampling, &state.Sampling)
}

func alignPieStateFromPlan(plan, state *models.PieChartConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignRegionMapStateFromPlan(ctx context.Context, plan, state *models.RegionMapConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	preservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
}

func alignTagcloudStateFromPlan(ctx context.Context, plan, state *models.TagcloudConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	preservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
	preservePlanJSONIfStateAddsOptionalKeys(plan.TagByJSON.Normalized, &state.TagByJSON.Normalized, "rank_by", "color")
	preservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.TagByJSON, &state.TagByJSON)
	if plan.EsqlMetric != nil && state.EsqlMetric != nil {
		preserveNormalizedJSONSemanticEquality(plan.EsqlMetric.FormatJSON, &state.EsqlMetric.FormatJSON)
	}
	if plan.EsqlTagBy != nil && state.EsqlTagBy != nil {
		preserveNormalizedJSONSemanticEquality(plan.EsqlTagBy.FormatJSON, &state.EsqlTagBy.FormatJSON)
		preserveNormalizedJSONSemanticEquality(plan.EsqlTagBy.ColorJSON, &state.EsqlTagBy.ColorJSON)
	}
}

func alignTreemapStateFromPlan(plan, state *models.TreemapConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	preserveKnownTfBoolIfStateNull(plan.IgnoreGlobalFilters, &state.IgnoreGlobalFilters)
	preserveKnownTfFloat64IfStateNull(plan.Sampling, &state.Sampling)
}

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

func alignTitleAndDescriptionFromPlan(planTitle, planDescription types.String, stateTitle, stateDescription *types.String) {
	preserveKnownStringIfStateBlank(planTitle, stateTitle)
	preserveKnownStringIfStateBlank(planDescription, stateDescription)
}

func preserveKnownTfBoolIfStateNull(plan types.Bool, state *types.Bool) {
	if typeutils.IsKnown(plan) && !plan.IsNull() && (!typeutils.IsKnown(*state) || state.IsNull()) {
		*state = plan
	}
}

func preserveKnownTfFloat64IfStateNull(plan types.Float64, state *types.Float64) {
	if typeutils.IsKnown(plan) && !plan.IsNull() && (!typeutils.IsKnown(*state) || state.IsNull()) {
		*state = plan
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

func metricChartMetricConfigsEquivalent(plan, state customtypes.JSONWithDefaultsValue[map[string]any]) bool {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(state) {
		return false
	}

	var planObj map[string]any
	if err := json.Unmarshal([]byte(plan.ValueString()), &planObj); err != nil {
		return false
	}
	var stateObj map[string]any
	if err := json.Unmarshal([]byte(state.ValueString()), &stateObj); err != nil {
		return false
	}

	planNormalized := normalizeXYPlanComparisonJSON(populateMetricChartMetricDefaults(planObj))
	stateNormalized := normalizeXYPlanComparisonJSON(populateMetricChartMetricDefaults(stateObj))
	return reflect.DeepEqual(planNormalized, stateNormalized)
}

func preserveMetricChartMetricConfigFromPlan(plan customtypes.JSONWithDefaultsValue[map[string]any], state *customtypes.JSONWithDefaultsValue[map[string]any]) {
	if metricChartMetricConfigsEquivalent(plan, *state) {
		*state = plan
	}
}

func preservePlanNormalizedJSONWithDefaultsIfSemanticallyEqual[T any](plan jsontypes.Normalized, state *jsontypes.Normalized, defaults func(T) T) {
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

	planNormalized := normalizeXYPlanComparisonJSON(defaults(planObj))
	stateNormalized := normalizeXYPlanComparisonJSON(defaults(stateObj))
	if reflect.DeepEqual(planNormalized, stateNormalized) {
		*state = plan
	}
}

func preservePlanJSONIfStateOmitsOptionalKeys(plan jsontypes.Normalized, state *jsontypes.Normalized, optionalKeys ...string) {
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

	stateNormalized := normalizeXYPlanComparisonJSON(stateObj)
	planNormalized := normalizeXYPlanComparisonJSON(planObj)
	if reflect.DeepEqual(stateNormalized, planNormalized) {
		*state = plan
	}
}

const gaugeEsqlTicksModeBandsDefault = "bands"

// gaugeEsqlAutoColorSentinel is the normalized form of Kibana's default
// `{"type":"auto"}` color payload, precomputed so `gaugeEsqlColorJSONIsAuto`
// does not allocate a fresh map on every alignment pass.
var gaugeEsqlAutoColorSentinel = normalizeXYPlanComparisonJSON(map[string]any{"type": "auto"})

func preserveNormalizedJSONSemanticEquality(plan jsontypes.Normalized, state *jsontypes.Normalized) {
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

	if reflect.DeepEqual(normalizeXYPlanComparisonJSON(planObj), normalizeXYPlanComparisonJSON(stateObj)) {
		*state = plan
	}
}

func alignGaugeEsqlMetricStateFromPlan(plan, state *models.GaugeEsqlMetric) {
	if plan == nil || state == nil {
		return
	}
	if plan.Title == nil && gaugeEsqlTitleMatchesKibanaDefaultVisible(state.Title) {
		state.Title = nil
	}
	if plan.Ticks == nil && gaugeEsqlTicksMatchesKibanaDefaultBands(state.Ticks) {
		state.Ticks = nil
	}
	if plan.ColorJSON.IsNull() && typeutils.IsKnown(state.ColorJSON) && gaugeEsqlColorJSONIsAuto(state.ColorJSON) {
		state.ColorJSON = jsontypes.NewNormalizedNull()
	}
}

func gaugeEsqlTitleMatchesKibanaDefaultVisible(title *models.GaugeEsqlTitle) bool {
	if title == nil {
		return false
	}
	textUnset := !typeutils.IsKnown(title.Text) || title.Text.IsNull()
	visibleTrue := typeutils.IsKnown(title.Visible) && !title.Visible.IsNull() && title.Visible.ValueBool()
	return textUnset && visibleTrue
}

func gaugeEsqlTicksMatchesKibanaDefaultBands(ticks *models.GaugeEsqlTicks) bool {
	if ticks == nil {
		return false
	}
	modeBands := typeutils.IsKnown(ticks.Mode) && ticks.Mode.ValueString() == gaugeEsqlTicksModeBandsDefault
	visibleTrue := typeutils.IsKnown(ticks.Visible) && !ticks.Visible.IsNull() && ticks.Visible.ValueBool()
	return modeBands && visibleTrue
}

func gaugeEsqlColorJSONIsAuto(color jsontypes.Normalized) bool {
	var m map[string]any
	if err := json.Unmarshal([]byte(color.ValueString()), &m); err != nil {
		return false
	}
	return reflect.DeepEqual(normalizeXYPlanComparisonJSON(m), gaugeEsqlAutoColorSentinel)
}
