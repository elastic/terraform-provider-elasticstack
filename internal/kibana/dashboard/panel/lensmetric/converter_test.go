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

package lensmetric

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_VizType(t *testing.T) {
	var c converter
	require.Equal(t, string(kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric), c.VizType())
}

func TestConverter_HandlesBlocks(t *testing.T) {
	var c converter
	require.False(t, c.HandlesBlocks(nil))
	require.False(t, c.HandlesBlocks(&models.LensByValueChartBlocks{}))
	require.True(t, c.HandlesBlocks(&models.LensByValueChartBlocks{
		MetricChartConfig: &models.MetricChartConfigModel{},
	}))
}

func TestConverter_roundTrip_NoESQL(t *testing.T) {
	ctx := t.Context()
	var c converter
	query := kbapi.KibanaHTTPAPIsFilterSimple{
		Language:   new(kbapi.KibanaHTTPAPIsFilterSimpleLanguage("kql")),
		Expression: "*",
	}
	apiChart := kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel{
		Type:                kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric,
		Title:               new("Metric Round-Trip"),
		Description:         new("Converter test"),
		IgnoreGlobalFilters: new(false),
		Sampling:            new(float32(1.0)),
		Query:               &query,
		Metrics:             []kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel_Metrics_Item{},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &apiChart.DataSource))
	metric := kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel_Metrics_Item{}
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metric))
	apiChart.Metrics = []kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel_Metrics_Item{metric}

	var attrs lenscommon.VisByValueConfig0
	require.NoError(t, attrs.FromKibanaHTTPAPIsMetricNoESQLByValuePanel(apiChart))

	blocks := &models.LensByValueChartBlocks{}
	diags := c.PopulateFromAttributes(ctx, blocks, attrs)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, blocks.MetricChartConfig)

	attrs2, diags := c.BuildAttributes(blocks)
	require.False(t, diags.HasError(), "%v", diags)

	variant0, err := attrs2.AsKibanaHTTPAPIsMetricNoESQLByValuePanel()
	require.NoError(t, err)
	assert.Equal(t, "Metric Round-Trip", *variant0.Title)
	assert.Equal(t, kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric, variant0.Type)
}

func TestConverter_roundTrip_ESQL_metric(t *testing.T) {
	ctx := t.Context()
	var c converter
	var metricItem kbapi.KibanaHTTPAPIsMetricESQLByValuePanel_Metrics_Item
	require.NoError(t, json.Unmarshal([]byte(`{
		"type": "primary",
		"operation": "count",
		"format": {"id": "number"},
		"alignments": {"value": "center"},
		"icon": {"name": "empty"}
	}`), &metricItem))

	title := "Metric ESQL RT"
	apiChart := kbapi.KibanaHTTPAPIsMetricESQLByValuePanel{
		Type:    kbapi.KibanaHTTPAPIsMetricESQLByValuePanelTypeMetric,
		Title:   &title,
		Metrics: []kbapi.KibanaHTTPAPIsMetricESQLByValuePanel_Metrics_Item{metricItem},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM logs-* | STATS c = COUNT(*) | LIMIT 1"}`), &apiChart.DataSource))

	var attrs lenscommon.VisByValueConfig0
	require.NoError(t, attrs.FromKibanaHTTPAPIsMetricESQLByValuePanel(apiChart))

	blocks := &models.LensByValueChartBlocks{}
	diags := c.PopulateFromAttributes(ctx, blocks, attrs)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, blocks.MetricChartConfig)
	require.Nil(t, blocks.MetricChartConfig.Query)
	assert.Contains(t, blocks.MetricChartConfig.DataSourceJSON.ValueString(), "FROM logs-*")
	require.Len(t, blocks.MetricChartConfig.Metrics, 1)

	attrs2, diags := c.BuildAttributes(blocks)
	require.False(t, diags.HasError(), "%v", diags)

	out, err := attrs2.AsKibanaHTTPAPIsMetricESQLByValuePanel()
	require.NoError(t, err)
	assert.Equal(t, kbapi.KibanaHTTPAPIsMetricESQLByValuePanelTypeMetric, out.Type)
	require.NotNil(t, out.Title)
	assert.Equal(t, "Metric ESQL RT", *out.Title)
	dsBytes, err := json.Marshal(out.DataSource)
	require.NoError(t, err)
	assert.Contains(t, string(dsBytes), "FROM logs-*")
}
