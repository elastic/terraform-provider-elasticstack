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

package lensmosaic

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
)

func alignMosaicStateFromPlan(ctx context.Context, plan, state *models.LensByValueChartBlocks) {
	if plan == nil || state == nil {
		return
	}
	alignMosaicConfigStateFromPlan(ctx, plan.MosaicConfig, state.MosaicConfig)
}

func alignMosaicConfigStateFromPlan(ctx context.Context, plan, state *models.MosaicConfigModel) {
	if plan == nil || state == nil {
		return
	}
	lenscommon.AlignStandardPartitionChartStateFromPlan(
		ctx,
		plan.Title, plan.Description,
		plan.DataSourceJSON,
		plan.IgnoreGlobalFilters,
		plan.Sampling,
		plan.GroupBy, plan.Metrics,
		plan.Legend, plan.ValueDisplay,
		&state.Title, &state.Description,
		&state.DataSourceJSON,
		&state.IgnoreGlobalFilters,
		&state.Sampling,
		&state.GroupBy, &state.Metrics,
		state.Legend, &state.ValueDisplay,
	)
	// Mosaic also has group_breakdown_by, which is not part of the common partition fields.
	lenscommon.PreservePlanJSONWithDefaultsIfSemanticallyEqual(ctx, plan.GroupBreakdownBy, &state.GroupBreakdownBy)
}
