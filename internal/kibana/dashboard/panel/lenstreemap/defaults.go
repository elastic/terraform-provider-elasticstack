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

import "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"

func populateTreemapLensAttributes(attrs map[string]any) map[string]any {
	if attrs == nil {
		return attrs
	}
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}
	if groupBy, ok := attrs["group_by"].([]any); ok {
		groupByMaps := make([]map[string]any, 0, len(groupBy))
		for _, g := range groupBy {
			if m, ok := g.(map[string]any); ok {
				groupByMaps = append(groupByMaps, m)
			}
		}
		populated := lenscommon.PopulatePartitionGroupByDefaults(groupByMaps)
		for i := range groupBy {
			if i < len(populated) {
				groupBy[i] = populated[i]
			}
		}
	}
	if metrics, ok := attrs["metrics"].([]any); ok {
		metricsMaps := make([]map[string]any, 0, len(metrics))
		for _, m := range metrics {
			if mp, ok := m.(map[string]any); ok {
				metricsMaps = append(metricsMaps, mp)
			}
		}
		populated := lenscommon.PopulatePartitionMetricsDefaults(metricsMaps)
		for i := range metrics {
			if i < len(populated) {
				metrics[i] = populated[i]
			}
		}
	}
	return attrs
}
