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
	// OperationTerms is the Lens terms breakdown discriminator used by several charts.
	OperationTerms      = "terms"
	pieChartTypeNumber  = "number"
	pieChartTypePercent = "percent"
	dashboardValueAvg   = "average"
)

// IsFieldMetricOperation reports whether operation names a standard Lens field metric.
func IsFieldMetricOperation(operation string) bool {
	switch operation {
	case "count", "unique_count", "min", "max", dashboardValueAvg, "median", "standard_deviation", "sum", "last_value", "percentile", "percentile_rank":
		return true
	default:
		return false
	}
}

// PopulateLensMetricDefaults populates default values for Lens metric configuration (shared across XY, metric, pie, treemap, datatable, etc.).
func PopulateLensMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}

	if format, ok := model["format"].(map[string]any); ok {
		formatType, _ := format["type"].(string)
		formatID, _ := format["id"].(string)
		isNumberish := formatType == pieChartTypeNumber || formatType == pieChartTypePercent || formatID == pieChartTypeNumber || formatID == pieChartTypePercent

		if isNumberish {
			if params, ok := format["params"].(map[string]any); ok {
				if _, exists := params["compact"]; !exists {
					params["compact"] = false
				}
				if _, exists := params["decimals"]; !exists {
					params["decimals"] = float64(2)
				}
				format["params"] = params
			} else {
				if _, exists := format["compact"]; !exists {
					format["compact"] = false
				}
				if _, exists := format["decimals"]; !exists {
					format["decimals"] = float64(2)
				}
			}
		}
	}

	if _, exists := model["empty_as_null"]; !exists {
		model["empty_as_null"] = false
	}
	if _, exists := model["fit"]; !exists {
		model["fit"] = false
	}
	if _, exists := model["color"]; !exists {
		model["color"] = map[string]any{"type": "auto"}
	}

	metricType, _ := model["type"].(string)

	if metricType == "primary" {
		if _, exists := model["value"]; !exists {
			model["value"] = map[string]any{"alignment": "right"}
		} else if v, ok := model["value"].(map[string]any); ok {
			if _, exists := v["alignment"]; !exists {
				v["alignment"] = "right"
			}
		}
		if _, exists := model["labels"]; !exists {
			model["labels"] = map[string]any{"alignment": "left"}
		} else if l, ok := model["labels"].(map[string]any); ok {
			if _, exists := l["alignment"]; !exists {
				l["alignment"] = "left"
			}
		}
	}

	if metricType == "secondary" {
		if _, exists := model["placement"]; !exists {
			model["placement"] = "before"
		}
		if _, exists := model["value"]; !exists {
			model["value"] = map[string]any{"alignment": "right"}
		} else if v, ok := model["value"].(map[string]any); ok {
			if _, exists := v["alignment"]; !exists {
				v["alignment"] = "right"
			}
		}
	}

	return model
}

// PopulateMetricChartMetricDefaults applies metric-chart defaults on top of PopulateLensMetricDefaults.
func PopulateMetricChartMetricDefaults(model map[string]any) map[string]any {
	_, hadColor := model["color"]
	model = PopulateLensMetricDefaults(model)
	if model == nil {
		return model
	}

	if metricType, _ := model["type"].(string); metricType == "secondary" && !hadColor {
		model["color"] = map[string]any{"type": "none"}
	}

	return model
}

// PopulatePartitionGroupByDefaults populates defaults for partition chart group-by slices (treemap, mosaic).
func PopulatePartitionGroupByDefaults(model []map[string]any) []map[string]any {
	if model == nil {
		return model
	}

	for _, item := range model {
		if item == nil {
			continue
		}
		operation, _ := item["operation"].(string)
		if operation == "value" {
			continue
		}
		if operation != OperationTerms {
			continue
		}
		if _, exists := item["collapse_by"]; !exists {
			item["collapse_by"] = "avg"
		}
		if _, exists := item["format"]; !exists {
			item["format"] = map[string]any{
				"type":     "number",
				"decimals": float64(2),
			}
		}
		if _, exists := item["rank_by"]; !exists {
			item["rank_by"] = map[string]any{
				"type":      "column",
				"metric":    float64(0),
				"direction": "desc",
			}
		}
		if _, exists := item["size"]; !exists {
			item["size"] = float64(5)
		}
	}

	return model
}

// PopulatePartitionMetricsDefaults normalizes partition metric slices using tagcloud metric defaults.
func PopulatePartitionMetricsDefaults(model []map[string]any) []map[string]any {
	if model == nil {
		return model
	}

	for i := range model {
		model[i] = PopulateTagcloudMetricDefaults(model[i])

		if model[i] == nil {
			continue
		}
		if operation, ok := model[i]["operation"].(string); ok && operation == "value" {
			if _, exists := model[i]["format"]; !exists {
				model[i]["format"] = nil
			}
		}
	}

	return model
}

// PopulateGaugeMetricDefaults populates gauge metric_json defaults.
func PopulateGaugeMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}

	if _, exists := model["empty_as_null"]; !exists {
		model["empty_as_null"] = false
	}
	if _, exists := model["title"]; !exists {
		model["title"] = map[string]any{"visible": true}
	}
	if _, exists := model["ticks"]; !exists {
		model["ticks"] = map[string]any{"visible": true, "mode": "bands"}
	}
	if _, exists := model["color"]; !exists {
		model["color"] = map[string]any{"type": "auto"}
	}

	return model
}

// PopulateRegionMapMetricDefaults populates region_map metric_json defaults.
// Behaviorally identical to PopulateTagcloudMetricDefaults; kept as a distinct symbol so the
// JSONWithDefaultsType identity stays per-chart in case the defaults diverge.
func PopulateRegionMapMetricDefaults(model map[string]any) map[string]any {
	return populateFieldMetricLensDefaults(model)
}

// populateFieldMetricLensDefaults applies the standard "field metric" Lens defaults
// (empty_as_null=false, show_metric_label=true, color=auto) when the model represents a field metric.
func populateFieldMetricLensDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	if operation, ok := model["operation"].(string); ok && IsFieldMetricOperation(operation) {
		if _, exists := model["empty_as_null"]; !exists {
			model["empty_as_null"] = false
		}
		if _, exists := model["show_metric_label"]; !exists {
			model["show_metric_label"] = true
		}
		if _, exists := model["color"]; !exists {
			model["color"] = map[string]any{"type": "auto"}
		}
	}
	return model
}

// PopulatePieChartMetricDefaults populates pie metric defaults inside metrics[].config_json.
func PopulatePieChartMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}

	if _, exists := model["empty_as_null"]; !exists {
		model["empty_as_null"] = false
	}
	if _, exists := model["color"]; !exists {
		model["color"] = map[string]any{"type": "auto"}
	}

	if format, ok := model["format"].(map[string]any); ok {
		if format["type"] == pieChartTypeNumber {
			if _, exists := format["compact"]; !exists {
				format["compact"] = false
			}
			if _, exists := format["decimals"]; !exists {
				format["decimals"] = float64(2)
			}
		}
	}

	return model
}

// PopulateLensGroupByDefaults populates Lens dimension/group-by JSON defaults (pie group_by, etc.).
func PopulateLensGroupByDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}

	if operation, ok := model["operation"].(string); ok && operation == OperationTerms {
		if _, exists := model["size"]; !exists {
			model["size"] = float64(5)
		}
		if _, exists := model["rank_by"]; !exists {
			model["rank_by"] = map[string]any{
				"direction": "desc",
				"metric":    float64(0),
				"type":      "column",
			}
		}
	}

	return model
}

// PopulateTagcloudMetricDefaults populates tagcloud metric defaults.
// Behaviorally identical to PopulateRegionMapMetricDefaults; kept as a distinct symbol so the
// JSONWithDefaultsType identity stays per-chart in case the defaults diverge.
func PopulateTagcloudMetricDefaults(model map[string]any) map[string]any {
	return populateFieldMetricLensDefaults(model)
}

// PopulateLegacyMetricMetricDefaults populates default values for legacy metric metric_json maps.
func PopulateLegacyMetricMetricDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	if operation, ok := model["operation"].(string); ok && IsFieldMetricOperation(operation) {
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
			case pieChartTypeNumber, pieChartTypePercent:
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

// PopulatePartitionLensAttributes populates shared defaults for partition-type Lens charts
// (treemap, mosaic, and future variants): nil-guard, default empty filters, group_by defaults,
// and metrics defaults. Returns attrs so callers can chain or return it directly.
func PopulatePartitionLensAttributes(attrs map[string]any) map[string]any {
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
		populated := PopulatePartitionGroupByDefaults(groupByMaps)
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
		populated := PopulatePartitionMetricsDefaults(metricsMaps)
		for i := range metrics {
			if i < len(populated) {
				metrics[i] = populated[i]
			}
		}
	}
	return attrs
}

// PopulateTagcloudTagByDefaults populates tagcloud tag_by defaults.
func PopulateTagcloudTagByDefaults(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	if operation, ok := model["operation"].(string); ok && operation == OperationTerms {
		if _, exists := model["rank_by"]; !exists {
			model["rank_by"] = map[string]any{
				"type":         "metric",
				"metric_index": float64(0),
				"direction":    "desc",
			}
		}
		if _, exists := model["color"]; !exists {
			model["color"] = map[string]any{
				"mode":    "categorical",
				"palette": "default",
				"mapping": []any{},
			}
		}
	}
	return model
}
