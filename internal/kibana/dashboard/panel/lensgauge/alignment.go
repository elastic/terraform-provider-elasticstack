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

package lensgauge

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

func alignGaugeStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	if plan == nil || state == nil {
		return
	}
	alignGaugeConfigStateFromPlan(ctx, plan.GaugeConfig, state.GaugeConfig)
}

func alignGaugeConfigStateFromPlan(ctx context.Context, plan, state *models.GaugeConfigModel) {
	if plan == nil || state == nil {
		return
	}
	lenscommon.AlignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	lenscommon.PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.MetricJSON, &state.MetricJSON)
	if plan.EsqlMetric != nil && state.EsqlMetric != nil {
		alignGaugeEsqlMetricStateFromPlan(plan.EsqlMetric, state.EsqlMetric)
	}
}

const gaugeEsqlTicksModeBandsDefault = "bands"

var gaugeEsqlAutoColorSentinel = lenscommon.NormalizeXYPlanComparisonJSON(map[string]any{"type": "auto"})

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
	return reflect.DeepEqual(lenscommon.NormalizeXYPlanComparisonJSON(m), gaugeEsqlAutoColorSentinel)
}
