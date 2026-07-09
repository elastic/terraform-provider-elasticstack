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

package lenstreemap

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
)

func alignTreemapStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	if plan == nil || state == nil {
		return
	}
	alignTreemapConfigStateFromPlan(ctx, plan.TreemapConfig, state.TreemapConfig)
}

func alignTreemapConfigStateFromPlan(ctx context.Context, plan, state *models.TreemapConfigModel) {
	if plan == nil || state == nil {
		return
	}
	lenscommon.AlignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.DataSourceJSON, &state.DataSourceJSON, "time_field")
	lenscommon.PreserveKnownTfValueIfStateNull(plan.IgnoreGlobalFilters, &state.IgnoreGlobalFilters)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.Sampling, &state.Sampling)
	// Partition group_by/metrics JSON gets re-emitted by Kibana with default keys
	// (color, rank_by, format defaults). Treat them as semantically equal.
	lenscommon.PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.GroupBy, &state.GroupBy)
	lenscommon.PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.Metrics, &state.Metrics)
	lenscommon.AlignPartitionLegendStateFromPlan(plan.Legend, state.Legend)
	if plan.ValueDisplay == nil && state.ValueDisplay != nil && lenscommon.PartitionValueDisplayMatchesKibanaDefault(state.ValueDisplay) {
		state.ValueDisplay = nil
	}
}
