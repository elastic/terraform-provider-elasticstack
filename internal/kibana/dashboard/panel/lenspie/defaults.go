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

package lenspie

import "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"

// populatePieLensAttributes is the canonical opaque-attribute JSON defaulting for VizConverter.PopulateJSONDefaults.
func populatePieLensAttributes(attrs map[string]any) map[string]any {
	if attrs == nil {
		return attrs
	}
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}
	if metrics, ok := attrs["metrics"].([]any); ok {
		for i, m := range metrics {
			if metricMap, ok := m.(map[string]any); ok {
				metrics[i] = lenscommon.PopulatePieChartMetricDefaults(metricMap)
			}
		}
	}
	if groupBy, ok := attrs["group_by"].([]any); ok {
		for i, g := range groupBy {
			if groupMap, ok := g.(map[string]any); ok {
				groupBy[i] = lenscommon.PopulateLensGroupByDefaults(groupMap)
			}
		}
	}
	return attrs
}
