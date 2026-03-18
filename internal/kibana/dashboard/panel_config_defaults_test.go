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
)

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

			var expected map[string]any
			require.NoError(t, json.Unmarshal([]byte(tt.expected), &expected))

			// Compare attributes.metric and attributes.filters
			attrs, ok := result["attributes"].(map[string]any)
			require.True(t, ok)
			expAttrs := expected["attributes"].(map[string]any)

			if filters, ok := attrs["filters"]; ok {
				assert.Equal(t, expAttrs["filters"], filters)
			}
			if metric, ok := attrs["metric"].(map[string]any); ok {
				expMetric := expAttrs["metric"].(map[string]any)
				assert.Equal(t, expMetric["show_array_values"], metric["show_array_values"])
				assert.Equal(t, expMetric["empty_as_null"], metric["empty_as_null"])
			}
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
	input := `{
		"attributes": {
			"type": "xy",
			"layers": [
				{
					"type": "line",
					"y": [{"operation": "count", "axis": "left"}]
				}
			]
		}
	}`
	var config map[string]any
	require.NoError(t, json.Unmarshal([]byte(input), &config))

	result := populatePanelConfigJSONDefaults(config)

	attrs := result["attributes"].(map[string]any)
	layers := attrs["layers"].([]any)
	layer0 := layers[0].(map[string]any)
	yArr := layer0["y"].([]any)
	y0 := yArr[0].(map[string]any)
	assert.Equal(t, false, y0["empty_as_null"])
	assert.Equal(t, false, y0["fit"])
	assert.Contains(t, attrs, "filters")
}

func Test_populatePanelConfigJSONDefaults_datatable(t *testing.T) {
	input := `{
		"attributes": {
			"type": "datatable",
			"metrics": [
				{"operation": "count"}
			]
		}
	}`
	var config map[string]any
	require.NoError(t, json.Unmarshal([]byte(input), &config))

	result := populatePanelConfigJSONDefaults(config)

	attrs := result["attributes"].(map[string]any)
	metrics := attrs["metrics"].([]any)
	m0 := metrics[0].(map[string]any)
	assert.Equal(t, false, m0["empty_as_null"])
	assert.Equal(t, false, m0["fit"])
	assert.Contains(t, attrs, "filters")
}
