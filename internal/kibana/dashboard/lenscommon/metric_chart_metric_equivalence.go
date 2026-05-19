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
	"encoding/json"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
)

// MetricChartMetricConfigsEquivalent compares metric chart metric TF configs ignoring structural defaults Kibana adds on read.
func MetricChartMetricConfigsEquivalent(plan, state customtypes.JSONWithDefaultsValue[map[string]any]) bool {
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

	planNormalized := NormalizeXYPlanComparisonJSON(PopulateMetricChartMetricDefaults(planObj))
	stateNormalized := NormalizeXYPlanComparisonJSON(PopulateMetricChartMetricDefaults(stateObj))
	return reflect.DeepEqual(planNormalized, stateNormalized)
}
