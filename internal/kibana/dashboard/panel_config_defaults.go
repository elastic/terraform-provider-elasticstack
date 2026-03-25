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

package dashboard

// populatePanelConfigJSONDefaults normalizes panel config_json for semantic equality.
// It dispatches to type-specific defaulting based on config structure (markdown vs lens)
// so that user config and API read-back compare equal despite key reordering and
// server-injected defaults (GitHub issue #1789).
func populatePanelConfigJSONDefaults(config map[string]any) map[string]any {
	if config == nil {
		return config
	}

	// Lens panels have an "attributes" object with type-specific structure
	if attrs, ok := config["attributes"].(map[string]any); ok {
		config["attributes"] = populateLensAttributesDefaults(attrs)
	}

	// Markdown panels have title, content - no additional defaults needed
	return config
}

// populateLensAttributesDefaults applies type-specific defaults to lens attributes.
// Dispatches on attributes.type to reuse existing populate functions from schema.go.
func populateLensAttributesDefaults(attrs map[string]any) map[string]any {
	if attrs == nil {
		return attrs
	}

	vizType, _ := attrs["type"].(string)
	switch vizType {
	case "legacy_metric":
		populateLegacyMetricAttributes(attrs)
	case "tagcloud":
		populateTagcloudAttributes(attrs)
	case "gauge":
		populateGaugeAttributes(attrs)
	case "metric":
		populateMetricChartAttributes(attrs)
	case "pie":
		populatePieChartAttributes(attrs)
	case "region_map":
		populateRegionMapAttributes(attrs)
	case "heatmap":
		populateHeatmapAttributes(attrs)
	case "treemap":
		populateTreemapAttributes(attrs)
	case "waffle":
		populateWaffleAttributes(attrs)
	case "xy":
		populateXYChartAttributes(attrs)
	case "datatable":
		populateDatatableAttributes(attrs)
	default:
		// Unknown type: no defaults applied
	}

	return attrs
}

func populateLegacyMetricAttributes(attrs map[string]any) {
	// Ensure filters: [] when absent (API may omit when empty)
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}

	// Apply metric defaults (show_array_values, empty_as_null, format)
	if metric, ok := attrs["metric"].(map[string]any); ok {
		attrs["metric"] = populateLegacyMetricMetricDefaults(metric)
	}
}

func populateTagcloudAttributes(attrs map[string]any) {
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}
	if metric, ok := attrs["metric"].(map[string]any); ok {
		attrs["metric"] = populateTagcloudMetricDefaults(metric)
	}
	if tagBy, ok := attrs["tag_by"].(map[string]any); ok {
		attrs["tag_by"] = populateTagcloudTagByDefaults(tagBy)
	}
}

func populateGaugeAttributes(attrs map[string]any) {
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}
	if metric, ok := attrs["metric"].(map[string]any); ok {
		attrs["metric"] = populateGaugeMetricDefaults(metric)
	}
}

func populateMetricChartAttributes(attrs map[string]any) {
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}
	if metrics, ok := attrs["metrics"].([]any); ok {
		for i, m := range metrics {
			if metricMap, ok := m.(map[string]any); ok {
				metrics[i] = populateLensMetricDefaults(metricMap)
			}
		}
	}
}

func populatePieChartAttributes(attrs map[string]any) {
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}
	if metrics, ok := attrs["metrics"].([]any); ok {
		for i, m := range metrics {
			if metricMap, ok := m.(map[string]any); ok {
				metrics[i] = populatePieChartMetricDefaults(metricMap)
			}
		}
	}
	if groupBy, ok := attrs["group_by"].([]any); ok {
		for i, g := range groupBy {
			if groupMap, ok := g.(map[string]any); ok {
				groupBy[i] = populateLensGroupByDefaults(groupMap)
			}
		}
	}
}

// populateWaffleAttributes mirrors pie chart defaulting for metrics and group_by (see getWaffleSchema
// and models_waffle_panel: same populatePieChartMetricDefaults / populateLensGroupByDefaults).
func populateWaffleAttributes(attrs map[string]any) {
	populatePieChartAttributes(attrs)
}

func populateRegionMapAttributes(attrs map[string]any) {
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}
	if metric, ok := attrs["metric"].(map[string]any); ok {
		attrs["metric"] = populateRegionMapMetricDefaults(metric)
	}
}

func populateHeatmapAttributes(attrs map[string]any) {
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}
	if metric, ok := attrs["metric"].(map[string]any); ok {
		attrs["metric"] = populateTagcloudMetricDefaults(metric)
	}
}

func populateTreemapAttributes(attrs map[string]any) {
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

		populated := populatePartitionGroupByDefaults(groupByMaps)
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
		populated := populatePartitionMetricsDefaults(metricsMaps)
		for i := range metrics {
			if i < len(populated) {
				metrics[i] = populated[i]
			}
		}
	}
}

func populateXYChartAttributes(attrs map[string]any) {
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}
	// XY chart has layers: each layer can be a data layer (has "y" array) or reference line
	if layers, ok := attrs["layers"].([]any); ok {
		for _, layer := range layers {
			layerMap, ok := layer.(map[string]any)
			if !ok {
				continue
			}
			// Data layers have "y" array of metric configs
			if yArr, ok := layerMap["y"].([]any); ok {
				for i, y := range yArr {
					if yMap, ok := y.(map[string]any); ok {
						yArr[i] = populateLensMetricDefaults(yMap)
					}
				}
			}
		}
	}
}

func populateDatatableAttributes(attrs map[string]any) {
	if _, exists := attrs["filters"]; !exists {
		attrs["filters"] = []any{}
	}
	// Datatable metrics: each item is a metric config (operation, format, etc.)
	if metrics, ok := attrs["metrics"].([]any); ok {
		for i, m := range metrics {
			if metricMap, ok := m.(map[string]any); ok {
				metrics[i] = populateLensMetricDefaults(metricMap)
			}
		}
	}
	// Datatable rows: terms, date_histogram, etc. - apply group_by defaults for terms
	if rows, ok := attrs["rows"].([]any); ok {
		for i, r := range rows {
			if rowMap, ok := r.(map[string]any); ok {
				rows[i] = populateLensGroupByDefaults(rowMap)
			}
		}
	}
	// Datatable split_metrics_by: similar to group_by
	if splitBy, ok := attrs["split_metrics_by"].([]any); ok {
		for i, s := range splitBy {
			if splitMap, ok := s.(map[string]any); ok {
				splitBy[i] = populateLensGroupByDefaults(splitMap)
			}
		}
	}
}
