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
