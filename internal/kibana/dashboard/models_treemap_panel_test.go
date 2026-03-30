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
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_treemapPanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip_NoESQL(t *testing.T) {
	ctx := context.Background()

	groupBy := `[{"operation":"terms","collapse_by":"avg","fields":["host.name"],` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	apiJSON := `{
		"type": "treemap",
		"title": "Treemap NoESQL Round-Trip",
		"description": "Converter test",
		"ignore_global_filters": true,
		"sampling": 0.5,
		"dataset": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kuery","query":"status:200"},
		"legend": {"size": "medium"},
		"metrics": [{"operation":"count"}],
		"group_by": ` + groupBy + `
	}`
	var api kbapi.TreemapNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var treemapChart kbapi.TreemapChart
	require.NoError(t, treemapChart.FromTreemapNoESQL(api))

	var attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	require.NoError(t, attrs.FromTreemapChart(treemapChart))

	converter := newTreemapPanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.TreemapConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	chart2, err := attrs2.AsTreemapChart()
	require.NoError(t, err)
	noESQL2, err := chart2.AsTreemapNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "Treemap NoESQL Round-Trip", *noESQL2.Title)
	assert.Equal(t, kbapi.TreemapNoESQLTypeTreemap, noESQL2.Type)
}

func Test_treemapPanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip_ESQL(t *testing.T) {
	ctx := context.Background()

	apiJSON := `{
		"type": "treemap",
		"title": "Treemap ESQL Round-Trip",
		"description": "Converter test",
		"ignore_global_filters": false,
		"sampling": 1,
		"dataset": {"type":"esql","query":"FROM metrics-* | LIMIT 10"},
		"legend": {"size": "small"},
		"metrics": [{"column":"bytes","operation":"value","color":{"type":"static","color":"#54B399"},"format":{"type":"number","decimals":2}}],
		"group_by": [{"collapse_by":"avg","column":"host.name","operation":"value","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]
	}`
	var api kbapi.TreemapESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var treemapChart kbapi.TreemapChart
	require.NoError(t, treemapChart.FromTreemapESQL(api))

	var attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	require.NoError(t, attrs.FromTreemapChart(treemapChart))

	converter := newTreemapPanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.TreemapConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	chart2, err := attrs2.AsTreemapChart()
	require.NoError(t, err)
	esql2, err := chart2.AsTreemapESQL()
	require.NoError(t, err)
	assert.Equal(t, "Treemap ESQL Round-Trip", *esql2.Title)
	assert.Equal(t, kbapi.TreemapESQLTypeTreemap, esql2.Type)
}

func Test_newTreemapPanelConfigConverter(t *testing.T) {
	converter := newTreemapPanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, "treemap", converter.visualizationType)
}

func Test_treemapConfigModel_fromAPI_toAPI_noESQL(t *testing.T) {
	api := kbapi.TreemapNoESQL{
		Type:                kbapi.TreemapNoESQLTypeTreemap,
		Title:               new("Test Treemap"),
		Description:         new("Treemap description"),
		IgnoreGlobalFilters: new(true),
		Sampling:            new(float32(0.5)),
		Query: kbapi.FilterSimple{
			Query: "status:200",
			Language: func() *kbapi.FilterSimpleLanguage {
				lang := kbapi.FilterSimpleLanguage("kuery")
				return &lang
			}(),
		},
		Legend: kbapi.TreemapLegend{
			Size: kbapi.LegendSizeMedium,
			Nested: func() *bool {
				b := true
				return &b
			}(),
			TruncateAfterLines: new(float32(4)),
			Visible: func() *kbapi.TreemapLegendVisible {
				v := kbapi.TreemapLegendVisibleAuto
				return &v
			}(),
		},
		ValueDisplay: kbapi.ValueDisplay{
			Mode:            kbapi.ValueDisplayModePercentage,
			PercentDecimals: new(float32(2)),
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.Dataset))

	var groupByItem kbapi.TreemapNoESQL_GroupBy_Item
	require.NoError(t, json.Unmarshal([]byte(`{
		"operation":"terms",
		"collapse_by":"avg",
		"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}},
		"fields":["host.name"],
		"format":{"type":"number","decimals":2}
	}`), &groupByItem))
	groupBy := []kbapi.TreemapNoESQL_GroupBy_Item{groupByItem}
	api.GroupBy = &groupBy

	var metricItem kbapi.TreemapNoESQL_Metrics_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metricItem))
	api.Metrics = []kbapi.TreemapNoESQL_Metrics_Item{metricItem}

	lp := kbapi.TreemapNoESQLLabelPositionVisible
	api.LabelPosition = &lp

	model := &treemapConfigModel{}
	diags := model.fromAPINoESQL(api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Test Treemap"), model.Title)
	assert.Equal(t, types.StringValue("Treemap description"), model.Description)
	assert.Equal(t, types.BoolValue(true), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(0.5), model.Sampling)
	require.NotNil(t, model.Query)
	assert.Equal(t, types.StringValue("status:200"), model.Query.Query)
	assert.Equal(t, types.StringValue("kuery"), model.Query.Language)
	assert.False(t, model.Dataset.IsNull())
	assert.False(t, model.GroupBy.IsNull())
	assert.False(t, model.Metrics.IsNull())
	assert.Equal(t, types.StringValue("visible"), model.LabelPosition)
	require.NotNil(t, model.Legend)
	assert.Equal(t, types.StringValue("medium"), model.Legend.Size)
	require.NotNil(t, model.ValueDisplay)
	assert.Equal(t, types.StringValue("percentage"), model.ValueDisplay.Mode)
	assert.Equal(t, types.Float64Value(2), model.ValueDisplay.PercentDecimals)

	schema, diags := model.toAPI()
	require.False(t, diags.HasError())

	roundTrip, err := schema.AsTreemapNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.TreemapNoESQLTypeTreemap, roundTrip.Type)
	assert.NotNil(t, roundTrip.GroupBy)
	assert.Len(t, *roundTrip.GroupBy, 1)
	assert.Len(t, roundTrip.Metrics, 1)
}

func Test_treemapConfigModel_fromAPI_toAPI_esql(t *testing.T) {
	colorMapping := kbapi.ColorMapping{}
	require.NoError(t, json.Unmarshal([]byte(`{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}`), &colorMapping))

	staticColor := kbapi.StaticColor{}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"static","color":"#54B399"}`), &staticColor))

	format := kbapi.FormatType{}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number","decimals":2}`), &format))

	groupBy := []struct {
		CollapseBy kbapi.CollapseBy                  `json:"collapse_by"`
		Color      kbapi.ColorMapping                `json:"color"`
		Column     string                            `json:"column"`
		Format     kbapi.FormatType                  `json:"format"`
		Label      *string                           `json:"label,omitempty"`
		Operation  kbapi.TreemapESQLGroupByOperation `json:"operation"`
	}{
		{
			CollapseBy: kbapi.CollapseByAvg,
			Color:      colorMapping,
			Column:     "host.name",
			Format:     format,
			Operation:  kbapi.TreemapESQLGroupByOperationValue,
		},
	}

	metrics := []struct {
		Color     kbapi.StaticColor                 `json:"color"`
		Column    string                            `json:"column"`
		Format    kbapi.FormatType                  `json:"format"`
		Label     *string                           `json:"label,omitempty"`
		Operation kbapi.TreemapESQLMetricsOperation `json:"operation"`
	}{
		{
			Color:     staticColor,
			Column:    "bytes",
			Format:    format,
			Operation: kbapi.TreemapESQLMetricsOperationValue,
		},
	}

	api := kbapi.TreemapESQL{
		Type:                kbapi.TreemapESQLTypeTreemap,
		Title:               new("ESQL Treemap"),
		Description:         new("ESQL description"),
		IgnoreGlobalFilters: new(false),
		Sampling:            new(float32(1)),
		Legend:              kbapi.TreemapLegend{Size: kbapi.LegendSizeSmall},
		Metrics:             metrics,
		GroupBy:             &groupBy,
		ValueDisplay: kbapi.ValueDisplay{
			Mode: kbapi.ValueDisplayModeAbsolute,
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM metrics-* | LIMIT 10"}`), &api.Dataset))

	lp := kbapi.TreemapESQLLabelPositionHidden
	api.LabelPosition = &lp

	model := &treemapConfigModel{}
	diags := model.fromAPIESQL(api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("ESQL Treemap"), model.Title)
	assert.False(t, model.Dataset.IsNull())
	assert.False(t, model.GroupBy.IsNull())
	assert.False(t, model.Metrics.IsNull())
	assert.Equal(t, types.StringValue("hidden"), model.LabelPosition)
	assert.Nil(t, model.Query)

	schema, diags := model.toAPI()
	require.False(t, diags.HasError())

	// The ES|QL treemap attributes are marshalled from a map for maximum compatibility
	// with Kibana validation behavior. Validate the resulting JSON contains the key
	// shape rather than requiring it to decode into the generated schema.
	b, err := json.Marshal(schema)
	require.NoError(t, err)

	var attrs map[string]any
	require.NoError(t, json.Unmarshal(b, &attrs))
	assert.Equal(t, "treemap", attrs["type"])

	dataset, ok := attrs["dataset"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "esql", dataset["type"])

	groupByAny, ok := attrs["group_by"].([]any)
	require.True(t, ok)
	assert.Len(t, groupByAny, 1)

	metricsAny, ok := attrs["metrics"].([]any)
	require.True(t, ok)
	assert.Len(t, metricsAny, 1)
}

func Test_treemapConfigModel_fromAPINoESQL_preservesKnownWhenAPIIsDefault(t *testing.T) {
	// Exercises mapOptionalBoolWithSnapshotDefault and mapOptionalFloatWithSnapshotDefault:
	// when API returns snapshot defaults (false for IgnoreGlobalFilters, 1 for Sampling),
	// we preserve the existing model values if they differ.
	api := kbapi.TreemapNoESQL{
		Type:                kbapi.TreemapNoESQLTypeTreemap,
		IgnoreGlobalFilters: new(false),      // snapshot default
		Sampling:            new(float32(1)), // snapshot default
		Query:               kbapi.FilterSimple{Query: "x", Language: func() *kbapi.FilterSimpleLanguage { l := kbapi.FilterSimpleLanguage("kuery"); return &l }()},
		Legend:              kbapi.TreemapLegend{Size: kbapi.LegendSizeMedium},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"x"}`), &api.Dataset))
	var metricItem kbapi.TreemapNoESQL_Metrics_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metricItem))
	api.Metrics = []kbapi.TreemapNoESQL_Metrics_Item{metricItem}
	var groupByItem kbapi.TreemapNoESQL_GroupBy_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"terms","field":"host.name","collapse_by":"avg"}`), &groupByItem))
	groupBy := []kbapi.TreemapNoESQL_GroupBy_Item{groupByItem}
	api.GroupBy = &groupBy

	model := &treemapConfigModel{
		IgnoreGlobalFilters: types.BoolValue(true),
		Sampling:            types.Float64Value(0.5),
	}
	diags := model.fromAPINoESQL(api)
	require.False(t, diags.HasError())

	// Should preserve existing values when API has defaults
	assert.True(t, model.IgnoreGlobalFilters.ValueBool())
	assert.InDelta(t, 0.5, model.Sampling.ValueFloat64(), 0.001)
}

func Test_treemapConfigModel_toAPIESQLChartSchema(t *testing.T) {
	// Build model via round-trip path to ensure compatible structure
	apiJSON := `{
		"type": "treemap",
		"title": "ESQL Treemap",
		"description": "Test",
		"ignore_global_filters": false,
		"sampling": 1,
		"dataset": {"type":"esql","query":"FROM metrics-* | LIMIT 10"},
		"legend": {"size": "small"},
		"metrics": [{"column":"bytes","operation":"value","color":{"type":"static","color":"#54B399"},"format":{"type":"number","decimals":2}}],
		"group_by": [{"collapse_by":"avg","column":"host.name","operation":"value","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]
	}`
	var api kbapi.TreemapESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var treemapChart kbapi.TreemapChart
	require.NoError(t, treemapChart.FromTreemapESQL(api))

	var attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	require.NoError(t, attrs.FromTreemapChart(treemapChart))

	converter := newTreemapPanelConfigConverter()
	pm := &panelModel{}
	ctx := context.Background()
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.TreemapConfig)

	chart, diags := pm.TreemapConfig.toAPIESQLChartSchema()
	require.False(t, diags.HasError())

	b, err := json.Marshal(chart)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, "treemap", out["type"])
	assert.Equal(t, "ESQL Treemap", out["title"])
}
