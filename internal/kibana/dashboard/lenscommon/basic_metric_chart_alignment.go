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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AlignBasicMetricChartStateFromPlan aligns the common "basic metric chart" state
// fields (title/description, data_source_json, metric_json) from plan into state.
// Shared by Lens panel types whose config is just a metric over a data source
// (gauge, heatmap, legacy_metric, region_map, tagcloud). Callers with additional
// fields should handle those after calling this function.
func AlignBasicMetricChartStateFromPlan[T any](
	ctx context.Context,
	planTitle, planDescription types.String,
	planDataSourceJSON jsontypes.Normalized,
	planMetricJSON customtypes.JSONWithDefaultsValue[T],
	stateTitle, stateDescription *types.String,
	stateDataSourceJSON *jsontypes.Normalized,
	stateMetricJSON *customtypes.JSONWithDefaultsValue[T],
) {
	AlignTitleAndDescriptionFromPlan(planTitle, planDescription, stateTitle, stateDescription)
	PreservePlanJSONIfStateAddsOptionalKeys(planDataSourceJSON, stateDataSourceJSON, "time_field", "name")
	PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, planMetricJSON, stateMetricJSON)
}
