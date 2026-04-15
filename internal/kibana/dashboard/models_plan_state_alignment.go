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
	"encoding/json"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// alignDashboardStateFromPlanPanels preserves practitioner intent for panel configs
// when Kibana injects defaults on read or omits configured fields from responses.
// Per-panel alignment is applied inside mapPanelFromAPI; this handles XY-specific
// state that requires the full panel slice for cross-panel context.
func alignDashboardStateFromPlanPanels(planPanels, statePanels []panelModel) {
	alignXYChartStateFromPlanPanels(planPanels, statePanels)
}

func alignPanelStateFromPlan(plan, state *panelModel) {
	if plan == nil || state == nil {
		return
	}

	preservePlanJSONIfStateOmitsOptionalKeys(plan.ConfigJSON.Normalized, &state.ConfigJSON.Normalized, "filters", "query")
	alignDatatableStateFromPlan(plan.DatatableConfig, state.DatatableConfig)
	alignGaugeStateFromPlan(plan.GaugeConfig, state.GaugeConfig)
	alignHeatmapStateFromPlan(plan.HeatmapConfig, state.HeatmapConfig)
	alignLegacyMetricStateFromPlan(plan.LegacyMetricConfig, state.LegacyMetricConfig)
	alignMetricStateFromPlan(plan.MetricChartConfig, state.MetricChartConfig)
	alignMosaicStateFromPlan(plan.MosaicConfig, state.MosaicConfig)
	alignPieStateFromPlan(plan.PieChartConfig, state.PieChartConfig)
	alignRegionMapStateFromPlan(plan.RegionMapConfig, state.RegionMapConfig)
	alignTagcloudStateFromPlan(plan.TagcloudConfig, state.TagcloudConfig)
	alignTreemapStateFromPlan(plan.TreemapConfig, state.TreemapConfig)
	alignWaffleStateFromPlan(plan.WaffleConfig, state.WaffleConfig)
	alignEsqlControlStateFromPlan(plan.EsqlControlConfig, state.EsqlControlConfig)
}

func alignDatatableStateFromPlan(plan, state *datatableConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignDatatableNoESQLStateFromPlan(plan.NoESQL, state.NoESQL)
	alignDatatableESQLStateFromPlan(plan.ESQL, state.ESQL)
}

func alignDatatableNoESQLStateFromPlan(plan, state *datatableNoESQLConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignDatatableESQLStateFromPlan(plan, state *datatableESQLConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignGaugeStateFromPlan(plan, state *gaugeConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignHeatmapStateFromPlan(plan, state *heatmapConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignLegacyMetricStateFromPlan(plan, state *legacyMetricConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignMetricStateFromPlan(plan, state *metricChartConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	preservePlanJSONIfStateAddsOptionalKeys(plan.BreakdownByJSON, &state.BreakdownByJSON, "rank_by")
}

func alignMosaicStateFromPlan(plan, state *mosaicConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignPieStateFromPlan(plan, state *pieChartConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignRegionMapStateFromPlan(plan, state *regionMapConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignTagcloudStateFromPlan(plan, state *tagcloudConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	preservePlanJSONIfStateAddsOptionalKeys(plan.TagByJSON.Normalized, &state.TagByJSON.Normalized, "rank_by")
}

func alignTreemapStateFromPlan(plan, state *treemapConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignWaffleStateFromPlan(plan, state *waffleConfigModel) {
	if plan == nil || state == nil {
		return
	}
	alignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
}

func alignEsqlControlStateFromPlan(plan, state *esqlControlConfigModel) {
	if plan == nil || state == nil {
		return
	}
	preserveKnownStringIfStateBlank(plan.EsqlQuery, &state.EsqlQuery)
	preserveKnownStringIfStateBlank(plan.Title, &state.Title)
	preserveKnownListIfStateNull(plan.AvailableOptions, &state.AvailableOptions)
}

func alignTitleAndDescriptionFromPlan(planTitle, planDescription types.String, stateTitle, stateDescription *types.String) {
	preserveKnownStringIfStateBlank(planTitle, stateTitle)
	preserveKnownStringIfStateBlank(planDescription, stateDescription)
}

func preserveKnownListIfStateNull(plan types.List, state *types.List) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
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
