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

package visconfig

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLensConfigJSONSeparatesChartAndPresentation(t *testing.T) {
	t.Parallel()

	const raw = `{
		"type": "metric",
		"title": "Metric",
		"hide_title": true,
		"hide_border": false,
		"time_range": {"from": "now-1h", "to": "now"},
		"metrics": []
	}`

	chart, presentation, err := lensConfigJSON([]byte(raw))
	require.NoError(t, err)

	chartJSON, err := json.Marshal(chart)
	require.NoError(t, err)
	assert.JSONEq(t, raw, string(chartJSON))
	metricChart, err := chart.AsKibanaHTTPAPIsMetricChart()
	require.NoError(t, err)
	metric, err := metricChart.AsKibanaHTTPAPIsMetricNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "metric", string(metric.Type))

	assert.Equal(t, "Metric", *presentation.Title)
	assert.True(t, *presentation.HideTitle)
	assert.False(t, *presentation.HideBorder)
	require.NotNil(t, presentation.TimeRange)
	assert.Equal(t, "now-1h", presentation.TimeRange.From)
}

func TestComposeLensVisConfigMergesChartAndPresentation(t *testing.T) {
	t.Parallel()

	var chart kbapi.KibanaHTTPAPIsLensApiConfig
	require.NoError(t, json.Unmarshal([]byte(`{"type":"metric","title":"Metric","metrics":[]}`), &chart))
	hideTitle := true
	presentation := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0{
		HideTitle: &hideTitle,
		TimeRange: &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
			From: "now-1h",
			To:   "now",
		},
	}

	config, err := composeLensVisConfig(chart, presentation)
	require.NoError(t, err)
	raw, err := json.Marshal(config)
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"type":"metric",
		"title":"Metric",
		"metrics":[],
		"hide_title":true,
		"time_range":{"from":"now-1h","to":"now"}
	}`, string(raw))
}
