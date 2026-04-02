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

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
)

// assertPanelConfigEquals asserts that result matches expected using semantic JSON equality.
// Use this for full-output tests to catch unintended changes to the normalized config.
func assertPanelConfigEquals(t *testing.T, expectedJSON string, result map[string]any) {
	t.Helper()
	var expected map[string]any
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expected))
	resultBytes, err := json.Marshal(result)
	require.NoError(t, err)
	expectedBytes, err := json.Marshal(expected)
	require.NoError(t, err)
	eq, err := schemautil.JSONBytesEqual(resultBytes, expectedBytes)
	require.NoError(t, err)
	assert.True(t, eq, "result %s != expected %s", string(resultBytes), string(expectedBytes))
}

func Test_populatePanelConfigJSONDefaults_legacyMetric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "adds show_array_values and filters",
			input: `{
				"attributes": {
					"type": "legacy_metric",
					"metric": {"field": "bytes", "operation": "last_value"},
					"dataset": {"type": "dataView", "id": "metrics-*"}
				}
			}`,
			expected: `{
				"attributes": {
					"type": "legacy_metric",
					"filters": [],
					"metric": {"field": "bytes", "operation": "last_value", "show_array_values": false, "empty_as_null": false},
					"dataset": {"type": "dataView", "id": "metrics-*"}
				}
			}`,
		},
		{
			name: "preserves existing filters and show_array_values",
			input: `{
				"attributes": {
					"type": "legacy_metric",
					"filters": [],
					"metric": {"field": "bytes", "operation": "last_value", "show_array_values": false}
				}
			}`,
			expected: `{
				"attributes": {
					"type": "legacy_metric",
					"filters": [],
					"metric": {"field": "bytes", "operation": "last_value", "show_array_values": false, "empty_as_null": false}
				}
			}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input map[string]any
			require.NoError(t, json.Unmarshal([]byte(tt.input), &input))

			result := populatePanelConfigJSONDefaults(input)

			assertPanelConfigEquals(t, tt.expected, result)
		})
	}
}

func Test_populatePanelConfigJSONDefaults_markdown(t *testing.T) {
	input := map[string]any{
		"title":   "My Panel",
		"content": "body",
	}
	result := populatePanelConfigJSONDefaults(input)
	assert.Equal(t, input, result)
	assert.Equal(t, "My Panel", result["title"])
	assert.Equal(t, "body", result["content"])
}

func Test_populatePanelConfigJSONDefaults_unknownLensType(t *testing.T) {
	input := map[string]any{
		"attributes": map[string]any{
			"type": "unknown_viz",
			"data": "something",
		},
	}
	result := populatePanelConfigJSONDefaults(input)
	assert.Equal(t, input, result)
}

func Test_populatePanelConfigJSONDefaults_tagcloud(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "adds filters and tagcloud metric and tag_by defaults",
			input: `{
				"attributes": {
					"type": "tagcloud",
					"metric": {"field": "bytes", "operation": "sum"},
					"tag_by": {"operation": "terms", "field": "host.name"}
				}
			}`,
			expected: `{
				"attributes": {
					"type": "tagcloud",
					"filters": [],
					"metric": {"field": "bytes", "operation": "sum", "empty_as_null": false, "show_metric_label": true},
					"tag_by": {
						"operation": "terms",
						"field": "host.name",
						"rank_by": {"type": "column", "metric": 0, "direction": "desc"}
					}
				}
			}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input map[string]any
			require.NoError(t, json.Unmarshal([]byte(tt.input), &input))
			result := populatePanelConfigJSONDefaults(input)
			assertPanelConfigEquals(t, tt.expected, result)
		})
	}
}

func Test_populatePanelConfigJSONDefaults_gauge(t *testing.T) {
	input := `{
		"attributes": {
			"type": "gauge",
			"metric": {"operation": "median", "field": "latency"}
		}
	}`
	expected := `{
		"attributes": {
			"type": "gauge",
			"filters": [],
			"metric": {
				"operation": "median",
				"field": "latency",
				"empty_as_null": false,
				"title": {"visible": true},
				"ticks": {"visible": true, "mode": "auto"}
			}
		}
	}`
	var config map[string]any
	require.NoError(t, json.Unmarshal([]byte(input), &config))
	result := populatePanelConfigJSONDefaults(config)
	assertPanelConfigEquals(t, expected, result)
}

func Test_populatePanelConfigJSONDefaults_metric(t *testing.T) {
	input := `{
		"attributes": {
			"type": "metric",
			"metrics": [{"operation": "count"}, {"operation": "sum", "field": "bytes"}]
		}
	}`
	expected := `{
		"attributes": {
			"type": "metric",
			"filters": [],
			"metrics": [
				{"operation": "count", "empty_as_null": false, "fit": false},
				{"operation": "sum", "field": "bytes", "empty_as_null": false, "fit": false}
			]
		}
	}`
	var config map[string]any
	require.NoError(t, json.Unmarshal([]byte(input), &config))
	result := populatePanelConfigJSONDefaults(config)
	assertPanelConfigEquals(t, expected, result)
}

func Test_populatePanelConfigJSONDefaults_pie(t *testing.T) {
	input := `{
		"attributes": {
			"type": "pie",
			"metrics": [{"operation": "count"}],
			"group_by": [{"operation": "terms", "field": "status"}]
		}
	}`
	expected := `{
		"attributes": {
			"type": "pie",
			"filters": [],
			"metrics": [{"operation": "count", "empty_as_null": false}],
			"group_by": [
				{
					"operation": "terms",
					"field": "status",
					"size": 5,
					"rank_by": {"type": "column", "metric": 0, "direction": "desc"}
				}
			]
		}
	}`
	var config map[string]any
	require.NoError(t, json.Unmarshal([]byte(input), &config))
	result := populatePanelConfigJSONDefaults(config)
	assertPanelConfigEquals(t, expected, result)
}

func Test_populatePanelConfigJSONDefaults_waffle(t *testing.T) {
	input := `{
		"attributes": {
			"type": "waffle",
			"metrics": [{"operation": "count"}],
			"group_by": [{"operation": "terms", "field": "status"}]
		}
	}`
	expected := `{
		"attributes": {
			"type": "waffle",
			"filters": [],
			"metrics": [{"operation": "count", "empty_as_null": false}],
			"group_by": [
				{
					"operation": "terms",
					"field": "status",
					"size": 5,
					"rank_by": {"type": "column", "metric": 0, "direction": "desc"}
				}
			]
		}
	}`
	var config map[string]any
	require.NoError(t, json.Unmarshal([]byte(input), &config))
	result := populatePanelConfigJSONDefaults(config)
	assertPanelConfigEquals(t, expected, result)
}

func Test_populatePanelConfigJSONDefaults_region_map(t *testing.T) {
	input := `{
		"attributes": {
			"type": "region_map",
			"metric": {"operation": "sum", "field": "count"}
		}
	}`
	expected := `{
		"attributes": {
			"type": "region_map",
			"filters": [],
			"metric": {
				"operation": "sum",
				"field": "count",
				"empty_as_null": false,
				"show_metric_label": true
			}
		}
	}`
	var config map[string]any
	require.NoError(t, json.Unmarshal([]byte(input), &config))
	result := populatePanelConfigJSONDefaults(config)
	assertPanelConfigEquals(t, expected, result)
}

func Test_populatePanelConfigJSONDefaults_heatmap(t *testing.T) {
	input := `{
		"attributes": {
			"type": "heatmap",
			"metric": {"operation": "max", "field": "cpu"}
		}
	}`
	expected := `{
		"attributes": {
			"type": "heatmap",
			"filters": [],
			"metric": {
				"operation": "max",
				"field": "cpu",
				"empty_as_null": false,
				"show_metric_label": true
			}
		}
	}`
	var config map[string]any
	require.NoError(t, json.Unmarshal([]byte(input), &config))
	result := populatePanelConfigJSONDefaults(config)
	assertPanelConfigEquals(t, expected, result)
}

func Test_populatePanelConfigJSONDefaults_treemap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		check    func(t *testing.T, attrs map[string]any)
	}{
		{
			name: "terms group_by gets partition defaults and metrics get tagcloud metric defaults",
			input: `{
				"attributes": {
					"type": "treemap",
					"group_by": [{"operation": "terms", "field": "org"}],
					"metrics": [{"operation": "count"}]
				}
			}`,
			expected: `{
				"attributes": {
					"type": "treemap",
					"filters": [],
					"group_by": [
						{
							"operation": "terms",
							"field": "org",
							"collapse_by": "avg",
							"format": {"type": "number", "decimals": 2},
							"size": 5,
							"rank_by": {"type": "column", "metric": 0, "direction": "desc"}
						}
					],
					"metrics": [{"operation": "count", "empty_as_null": false, "show_metric_label": true}]
				}
			}`,
		},
		{
			name: "value group_by gets color null and value metric gets format null",
			input: `{
				"attributes": {
					"type": "treemap",
					"group_by": [{"operation": "value"}],
					"metrics": [{"operation": "value"}]
				}
			}`,
			check: func(t *testing.T, attrs map[string]any) {
				assert.Equal(t, []any{}, attrs["filters"])
				gb := attrs["group_by"].([]any)
				require.Len(t, gb, 1)
				g0 := gb[0].(map[string]any)
				assert.Equal(t, "value", g0["operation"])
				assert.Nil(t, g0["color"])
				metrics := attrs["metrics"].([]any)
				require.Len(t, metrics, 1)
				m0 := metrics[0].(map[string]any)
				assert.Equal(t, "value", m0["operation"])
				assert.Nil(t, m0["format"])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config map[string]any
			require.NoError(t, json.Unmarshal([]byte(tt.input), &config))
			result := populatePanelConfigJSONDefaults(config)
			if tt.expected != "" {
				assertPanelConfigEquals(t, tt.expected, result)
			} else if tt.check != nil {
				attrs := result["attributes"].(map[string]any)
				tt.check(t, attrs)
			}
		})
	}
}

func Test_populatePanelConfigJSONDefaults_mosaic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		check    func(t *testing.T, attrs map[string]any)
	}{
		{
			name: "terms groupings get partition defaults and metrics get partition metric defaults",
			input: `{
				"attributes": {
					"type": "mosaic",
					"group_by": [{"operation": "terms", "field": "org"}],
					"group_breakdown_by": [{"operation": "terms", "field": "service"}],
					"metrics": [{"operation": "count"}]
				}
			}`,
			expected: `{
				"attributes": {
					"type": "mosaic",
					"filters": [],
					"group_by": [
						{
							"operation": "terms",
							"field": "org",
							"collapse_by": "avg",
							"format": {"type": "number", "decimals": 2},
							"size": 5,
							"rank_by": {"type": "column", "metric": 0, "direction": "desc"}
						}
					],
					"group_breakdown_by": [
						{
							"operation": "terms",
							"field": "service",
							"collapse_by": "avg",
							"format": {"type": "number", "decimals": 2},
							"size": 5,
							"rank_by": {"type": "column", "metric": 0, "direction": "desc"}
						}
					],
					"metrics": [{"operation": "count", "empty_as_null": false, "show_metric_label": true}]
				}
			}`,
		},
		{
			name: "value groupings and metrics normalize null fields",
			input: `{
				"attributes": {
					"type": "mosaic",
					"group_by": [{"operation": "value"}],
					"group_breakdown_by": [{"operation": "value"}],
					"metrics": [{"operation": "value"}]
				}
			}`,
			check: func(t *testing.T, attrs map[string]any) {
				assert.Equal(t, []any{}, attrs["filters"])
				groupBy := attrs["group_by"].([]any)
				require.Len(t, groupBy, 1)
				assert.Equal(t, "value", groupBy[0].(map[string]any)["operation"])

				groupBreakdownBy := attrs["group_breakdown_by"].([]any)
				require.Len(t, groupBreakdownBy, 1)
				assert.Equal(t, "value", groupBreakdownBy[0].(map[string]any)["operation"])

				metrics := attrs["metrics"].([]any)
				require.Len(t, metrics, 1)
				m0 := metrics[0].(map[string]any)
				assert.Equal(t, "value", m0["operation"])
				assert.Nil(t, m0["format"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config map[string]any
			require.NoError(t, json.Unmarshal([]byte(tt.input), &config))
			result := populatePanelConfigJSONDefaults(config)
			if tt.expected != "" {
				assertPanelConfigEquals(t, tt.expected, result)
			} else if tt.check != nil {
				attrs := result["attributes"].(map[string]any)
				tt.check(t, attrs)
			}
		})
	}
}

func Test_populatePanelConfigJSONDefaults_xy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // if set, use full semantic equality; else check is used
		check    func(t *testing.T, attrs map[string]any)
	}{
		{
			name: "adds filters and y metric defaults",
			input: `{
				"attributes": {
					"type": "xy",
					"layers": [
						{
							"type": "line",
							"y": [{"operation": "count", "axis": "left"}]
						}
					]
				}
			}`,
			expected: `{
				"attributes": {
					"type": "xy",
					"filters": [],
					"layers": [
						{
							"type": "line",
							"y": [{"operation": "count", "axis": "left", "empty_as_null": false, "fit": false}]
						}
					]
				}
			}`,
		},
		{
			name: "multiple layers and y metrics",
			input: `{
				"attributes": {
					"type": "xy",
					"layers": [
						{
							"type": "line",
							"y": [
								{"operation": "count"},
								{"operation": "sum", "field": "bytes"}
							]
						},
						{
							"type": "bar",
							"y": [{"operation": "avg", "field": "cpu"}]
						}
					]
				}
			}`,
			check: func(t *testing.T, attrs map[string]any) {
				layers := attrs["layers"].([]any)
				require.Len(t, layers, 2)
				// First layer: two y metrics
				y0 := layers[0].(map[string]any)["y"].([]any)
				require.Len(t, y0, 2)
				assert.Equal(t, false, y0[0].(map[string]any)["empty_as_null"])
				assert.Equal(t, false, y0[1].(map[string]any)["empty_as_null"])
				// Second layer: one y metric
				y1 := layers[1].(map[string]any)["y"].([]any)
				require.Len(t, y1, 1)
				assert.Equal(t, false, y1[0].(map[string]any)["fit"])
			},
		},
		{
			name: "layer without y (reference line) is skipped",
			input: `{
				"attributes": {
					"type": "xy",
					"layers": [
						{
							"type": "referenceLine",
							"thresholds": [{"value": 100}]
						}
					]
				}
			}`,
			check: func(t *testing.T, attrs map[string]any) {
				assert.Equal(t, []any{}, attrs["filters"])
				layers := attrs["layers"].([]any)
				require.Len(t, layers, 1)
				layer0 := layers[0].(map[string]any)
				assert.NotContains(t, layer0, "y")
				thresholds := layer0["thresholds"].([]any)
				require.Len(t, thresholds, 1)
				assert.InDelta(t, float64(100), thresholds[0].(map[string]any)["value"], 1e-9)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config map[string]any
			require.NoError(t, json.Unmarshal([]byte(tt.input), &config))
			result := populatePanelConfigJSONDefaults(config)
			if tt.expected != "" {
				assertPanelConfigEquals(t, tt.expected, result)
			} else if tt.check != nil {
				attrs := result["attributes"].(map[string]any)
				tt.check(t, attrs)
			}
		})
	}
}

func Test_populatePanelConfigJSONDefaults_datatable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // if set, use full semantic equality; else check is used
		check    func(t *testing.T, attrs map[string]any)
	}{
		{
			name: "adds filters and metric defaults",
			input: `{
				"attributes": {
					"type": "datatable",
					"metrics": [{"operation": "count"}]
				}
			}`,
			expected: `{
				"attributes": {
					"type": "datatable",
					"filters": [],
					"metrics": [{"operation": "count", "empty_as_null": false, "fit": false}]
				}
			}`,
		},
		{
			name: "rows with terms get group_by defaults",
			input: `{
				"attributes": {
					"type": "datatable",
					"metrics": [{"operation": "count"}],
					"rows": [
						{"operation": "terms", "field": "host.name"}
					]
				}
			}`,
			check: func(t *testing.T, attrs map[string]any) {
				rows := attrs["rows"].([]any)
				require.Len(t, rows, 1)
				row0 := rows[0].(map[string]any)
				assert.InDelta(t, float64(5), row0["size"], 1e-9)
				assert.Contains(t, row0, "rank_by")
				rankBy := row0["rank_by"].(map[string]any)
				assert.Equal(t, "desc", rankBy["direction"])
			},
		},
		{
			name: "split_metrics_by gets group_by defaults",
			input: `{
				"attributes": {
					"type": "datatable",
					"metrics": [{"operation": "count"}],
					"split_metrics_by": [
						{"operation": "terms", "field": "service.name"}
					]
				}
			}`,
			check: func(t *testing.T, attrs map[string]any) {
				splitBy := attrs["split_metrics_by"].([]any)
				require.Len(t, splitBy, 1)
				s0 := splitBy[0].(map[string]any)
				assert.InDelta(t, float64(5), s0["size"], 1e-9)
				assert.Contains(t, s0, "rank_by")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config map[string]any
			require.NoError(t, json.Unmarshal([]byte(tt.input), &config))
			result := populatePanelConfigJSONDefaults(config)
			if tt.expected != "" {
				assertPanelConfigEquals(t, tt.expected, result)
			} else if tt.check != nil {
				attrs := result["attributes"].(map[string]any)
				tt.check(t, attrs)
			}
		})
	}
}
