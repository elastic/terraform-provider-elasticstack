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

package lenscommon

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AlignPartitionLegendStateFromPlan handles the legend.visible = "auto" default
// Kibana injects for partition-style Lens charts (pie, treemap, mosaic) when
// the practitioner omits visible.
func AlignPartitionLegendStateFromPlan(plan, state *models.PartitionLegendModel) {
	if plan == nil || state == nil {
		return
	}
	PreserveNullStringIfStateEquals(plan.Visible, &state.Visible, "auto")
}

// PartitionValueDisplayMatchesKibanaDefault reports whether state.value_display
// looks like the Kibana-injected default block ({mode="percentage",
// percent_decimals=null}). The treemap, mosaic, pie, and waffle converters all
// emit this default block when the practitioner omits value_display.
func PartitionValueDisplayMatchesKibanaDefault(state *models.PartitionValueDisplay) bool {
	if state == nil {
		return false
	}
	modeIsPercentage := typeutils.IsKnown(state.Mode) && state.Mode.ValueString() == "percentage"
	percentDecimalsUnset := !typeutils.IsKnown(state.PercentDecimals)
	return modeIsPercentage && percentDecimalsUnset
}

// AlignStandardPartitionChartStateFromPlan aligns the common partition-chart state
// fields (title/description, data_source_json, ignore_global_filters, sampling,
// group_by, metrics, legend, value_display) from plan into state. Callers with
// additional fields (e.g. group_breakdown_by in mosaic) should handle those after
// calling this function.
func AlignStandardPartitionChartStateFromPlan(
	ctx context.Context,
	planTitle, planDescription types.String,
	planDataSourceJSON jsontypes.Normalized,
	planIgnoreGlobalFilters types.Bool,
	planSampling types.Float64,
	planGroupBy customtypes.JSONWithDefaultsValue[[]map[string]any],
	planMetrics customtypes.JSONWithDefaultsValue[[]map[string]any],
	planLegend *models.PartitionLegendModel,
	planValueDisplay *models.PartitionValueDisplay,
	stateTitle, stateDescription *types.String,
	stateDataSourceJSON *jsontypes.Normalized,
	stateIgnoreGlobalFilters *types.Bool,
	stateSampling *types.Float64,
	stateGroupBy *customtypes.JSONWithDefaultsValue[[]map[string]any],
	stateMetrics *customtypes.JSONWithDefaultsValue[[]map[string]any],
	stateLegend *models.PartitionLegendModel,
	stateValueDisplay **models.PartitionValueDisplay,
) {
	AlignTitleAndDescriptionFromPlan(planTitle, planDescription, stateTitle, stateDescription)
	PreservePlanJSONIfStateAddsOptionalKeys(planDataSourceJSON, stateDataSourceJSON, "time_field")
	PreserveKnownTfValueIfStateNull(planIgnoreGlobalFilters, stateIgnoreGlobalFilters)
	PreserveKnownTfValueIfStateNull(planSampling, stateSampling)
	PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, planGroupBy, stateGroupBy)
	PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, planMetrics, stateMetrics)
	AlignPartitionLegendStateFromPlan(planLegend, stateLegend)
	if planValueDisplay == nil && *stateValueDisplay != nil && PartitionValueDisplayMatchesKibanaDefault(*stateValueDisplay) {
		*stateValueDisplay = nil
	}
}
