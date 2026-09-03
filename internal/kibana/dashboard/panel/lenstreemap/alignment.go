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
}
