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

const (
	lensPieChartTypeNumber  = "number"
	lensPieChartTypePercent = "percent"
	lensDashboardAverage    = "average"
)

func isLegacyMetricFieldMetricOperation(operation string) bool {
	switch operation {
	case "count", "unique_count", "min", "max", lensDashboardAverage, "median", "standard_deviation", "sum", "last_value", "percentile", "percentile_rank":
		return true
	default:
		return false
	}
}

// PopulateLegacyMetricMetricDefaults populates default values for legacy metric metric_json maps.
func PopulateLegacyMetricMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	if operation, ok := model["operation"].(string); ok && isLegacyMetricFieldMetricOperation(operation) {
		if _, exists := model["show_array_values"]; !exists {
			model["show_array_values"] = false
		}
		if _, exists := model["empty_as_null"]; !exists {
			model["empty_as_null"] = false
		}
	}

	format, ok := model["format"].(map[string]any)
	if ok {
		if formatType, ok := format["type"].(string); ok {
			switch formatType {
			case lensPieChartTypeNumber, lensPieChartTypePercent:
				if _, exists := format["decimals"]; !exists {
					format["decimals"] = float64(2)
				}
				if _, exists := format["compact"]; !exists {
					format["compact"] = false
				}
			case "bytes", "bits":
				if _, exists := format["decimals"]; !exists {
					format["decimals"] = float64(2)
				}
			}
		}
		model["format"] = format
	}

	return model
}
