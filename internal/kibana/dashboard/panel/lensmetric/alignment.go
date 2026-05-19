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

package lensmetric

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
)

func alignMetricStateFromPlan(plan, state *models.MetricChartConfigModel) {
	if plan == nil || state == nil {
		return
	}
	lenscommon.AlignTitleAndDescriptionFromPlan(plan.Title, plan.Description, &state.Title, &state.Description)
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.DataSourceJSON, &state.DataSourceJSON, "time_field")
	lenscommon.PreservePlanJSONIfStateAddsOptionalKeys(plan.BreakdownByJSON, &state.BreakdownByJSON, "rank_by")
	m := min(len(plan.Metrics), len(state.Metrics))
	for i := range m {
		preserveMetricChartMetricConfigFromPlan(plan.Metrics[i].ConfigJSON, &state.Metrics[i].ConfigJSON)
	}
}

func preserveMetricChartMetricConfigFromPlan(plan customtypes.JSONWithDefaultsValue[map[string]any], state *customtypes.JSONWithDefaultsValue[map[string]any]) {
	if lenscommon.MetricChartMetricConfigsEquivalent(plan, *state) {
		*state = plan
	}
}
