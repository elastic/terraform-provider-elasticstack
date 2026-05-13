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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
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
		"data_source": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kql","query":"status:200"},
		"legend": {"size": "medium"},
		"metrics": [{"operation":"count"}],
		"group_by": ` + groupBy + `
	}`
	var api kbapi.TreemapNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromTreemapNoESQL(api))

	converter := newTreemapPanelConfigConverter()
	visBv := models.VisByValueModel{}
	diags := converter.populateFromAttributes(ctx, nil, nil, &visBv.LensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, visBv.TreemapConfig)

	attrs2, diags := converter.buildAttributes(&visBv.LensByValueChartBlocks, nil)
	require.False(t, diags.HasError())

	noESQL2, err := attrs2.AsTreemapNoESQL()
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
		"data_source": {"type":"esql","query":"FROM metrics-* | LIMIT 10"},
		"legend": {"size": "small"},
		"metrics": [{"column":"bytes","operation":"value","color":{"type":"static","color":"#54B399"},"format":{"type":"number","decimals":2}}],
		"group_by": [{"collapse_by":"avg","column":"host.name","operation":"value","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]
	}`
	var api kbapi.TreemapESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromTreemapESQL(api))

	converter := newTreemapPanelConfigConverter()
	visBv := models.VisByValueModel{}
	diags := converter.populateFromAttributes(ctx, nil, nil, &visBv.LensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, visBv.TreemapConfig)

	attrs2, diags := converter.buildAttributes(&visBv.LensByValueChartBlocks, nil)
	require.False(t, diags.HasError())

	esql2, err := attrs2.AsTreemapESQL()
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
			Expression: "status:200",
			Language: func() *kbapi.FilterSimpleLanguage {
				lang := kbapi.FilterSimpleLanguage("kql")
				return &lang
			}(),
		},
		Legend: kbapi.TreemapLegend{
			Size: kbapi.LegendSizeM,
			Nested: func() *bool {
				b := true
				return &b
			}(),
			TruncateAfterLines: new(float32(4)),
			Visibility: func() *kbapi.TreemapLegendVisibility {
				v := kbapi.TreemapLegendVisibilityAuto
				return &v
			}(),
		},
		Styling: kbapi.TreemapStyling{
			Values: kbapi.ValueDisplay{
				Mode:            func() *kbapi.ValueDisplayMode { m := kbapi.ValueDisplayModePercentage; return &m }(),
				PercentDecimals: new(float32(2)),
			},
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.DataSource))

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

	model := &models.TreemapConfigModel{}
	diags := treemapConfigFromAPINoESQL(context.Background(), model, nil, nil, api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Test Treemap"), model.Title)
	assert.Equal(t, types.StringValue("Treemap description"), model.Description)
	assert.Equal(t, types.BoolValue(true), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(0.5), model.Sampling)
	require.NotNil(t, model.Query)
	assert.Equal(t, types.StringValue("status:200"), model.Query.Expression)
	assert.Equal(t, types.StringValue("kql"), model.Query.Language)
	assert.False(t, model.DataSourceJSON.IsNull())
	assert.False(t, model.GroupBy.IsNull())
	assert.False(t, model.Metrics.IsNull())
	require.NotNil(t, model.Legend)
	assert.Equal(t, types.StringValue("m"), model.Legend.Size)
	require.NotNil(t, model.ValueDisplay)
	assert.Equal(t, types.StringValue("percentage"), model.ValueDisplay.Mode)
	assert.Equal(t, types.Float64Value(2), model.ValueDisplay.PercentDecimals)

	lensAttrs, diags := treemapConfigToAPI(model, nil)
	require.False(t, diags.HasError())

	roundTrip, err := lensAttrs.AsTreemapNoESQL()
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
		CollapseBy kbapi.CollapseBy   `json:"collapse_by"`
		Color      kbapi.ColorMapping `json:"color"`
		Column     string             `json:"column"`
		Format     kbapi.FormatType   `json:"format"`
		Label      *string            `json:"label,omitempty"`
	}{
		{
			CollapseBy: kbapi.CollapseByAvg,
			Color:      colorMapping,
			Column:     "host.name",
			Format:     format,
		},
	}

	staticColorUnion := kbapi.TreemapESQL_Metrics_Color{}
	require.NoError(t, staticColorUnion.FromStaticColor(staticColor))

	metrics := []struct {
		Color  *kbapi.TreemapESQL_Metrics_Color `json:"color,omitempty"`
		Column string                           `json:"column"`
		Format kbapi.FormatType                 `json:"format"`
		Label  *string                          `json:"label,omitempty"`
	}{
		{
			Color:  &staticColorUnion,
			Column: "bytes",
			Format: format,
		},
	}

	api := kbapi.TreemapESQL{
		Type:                kbapi.TreemapESQLTypeTreemap,
		Title:               new("ESQL Treemap"),
		Description:         new("ESQL description"),
		IgnoreGlobalFilters: new(false),
		Sampling:            new(float32(1)),
		Legend:              kbapi.TreemapLegend{Size: kbapi.LegendSizeS},
		Metrics:             metrics,
		GroupBy:             &groupBy,
		Styling: kbapi.TreemapStyling{
			Values: kbapi.ValueDisplay{
				Mode: func() *kbapi.ValueDisplayMode { m := kbapi.ValueDisplayModeAbsolute; return &m }(),
			},
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM metrics-* | LIMIT 10"}`), &api.DataSource))

	model := &models.TreemapConfigModel{}
	diags := treemapConfigFromAPIESQL(context.Background(), model, nil, nil, api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("ESQL Treemap"), model.Title)
	assert.False(t, model.DataSourceJSON.IsNull())
	assert.True(t, model.GroupBy.IsNull())
	assert.True(t, model.Metrics.IsNull())
	assert.NotEmpty(t, model.EsqlMetrics)
	assert.NotEmpty(t, model.EsqlGroupBy)
	assert.Nil(t, model.Query)

	lensAttrs, diags := treemapConfigToAPI(model, nil)
	require.False(t, diags.HasError())

	b, err := json.Marshal(lensAttrs)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(b, &raw))
	assert.Equal(t, "treemap", raw["type"])

	ds, ok := raw["data_source"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "esql", ds["type"])

	groupByAny, ok := raw["group_by"].([]any)
	require.True(t, ok)
	assert.Len(t, groupByAny, 1)

	metricsAny, ok := raw["metrics"].([]any)
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
		Query:               kbapi.FilterSimple{Expression: "x", Language: func() *kbapi.FilterSimpleLanguage { l := kbapi.FilterSimpleLanguage("kql"); return &l }()},
		Legend:              kbapi.TreemapLegend{Size: kbapi.LegendSizeM},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"x"}`), &api.DataSource))
	var metricItem kbapi.TreemapNoESQL_Metrics_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metricItem))
	api.Metrics = []kbapi.TreemapNoESQL_Metrics_Item{metricItem}
	var groupByItem kbapi.TreemapNoESQL_GroupBy_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"terms","field":"host.name","collapse_by":"avg"}`), &groupByItem))
	groupBy := []kbapi.TreemapNoESQL_GroupBy_Item{groupByItem}
	api.GroupBy = &groupBy

	model := &models.TreemapConfigModel{
		IgnoreGlobalFilters: types.BoolValue(true),
		Sampling:            types.Float64Value(0.5),
	}
	diags := treemapConfigFromAPINoESQL(context.Background(), model, nil, nil, api)
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
		"data_source": {"type":"esql","query":"FROM metrics-* | LIMIT 10"},
		"legend": {"size": "small"},
		"metrics": [{"column":"bytes","operation":"value","color":{"type":"static","color":"#54B399"},"format":{"type":"number","decimals":2}}],
		"group_by": [{"collapse_by":"avg","column":"host.name","operation":"value","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]
	}`
	var api kbapi.TreemapESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromTreemapESQL(api))

	converter := newTreemapPanelConfigConverter()
	visBv := models.VisByValueModel{}
	ctx := context.Background()
	diags := converter.populateFromAttributes(ctx, nil, nil, &visBv.LensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, visBv.TreemapConfig)

	lensAttrs, lensDiags := treemapConfigToAPI(visBv.TreemapConfig, nil)
	require.False(t, lensDiags.HasError())

	b, err := json.Marshal(lensAttrs)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, "treemap", out["type"])
	assert.Equal(t, "ESQL Treemap", out["title"])
}

func Test_treemapConfigModel_esqlTypedMetricsRoundTrip(t *testing.T) {
	// Verifies esql_metrics typed nested attribute round-trips correctly.
	apiJSON := `{
		"type": "treemap",
		"title": "ESQL Treemap Typed Test",
		"description": "test",
		"ignore_global_filters": false,
		"sampling": 1,
		"data_source": {"type":"esql","query":"FROM metrics-* | LIMIT 10"},
		"legend": {"size": "small"},
		"metrics": [{"column":"bytes","operation":"value","color":{"type":"static","color":"#54B399"},"format":{"type":"number","decimals":2}}],
		"group_by": [{"collapse_by":"avg","column":"host.name","operation":"value","color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]
	}`
	var api kbapi.TreemapESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromTreemapESQL(api))

	converter := newTreemapPanelConfigConverter()
	visBv := models.VisByValueModel{}
	diags := converter.populateFromAttributes(context.Background(), nil, nil, &visBv.LensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, visBv.TreemapConfig)

	// ES|QL mode: EsqlMetrics populated, Metrics/GroupBy null, Query nil
	assert.Nil(t, visBv.TreemapConfig.Query)
	assert.True(t, visBv.TreemapConfig.Metrics.IsNull())
	assert.True(t, visBv.TreemapConfig.GroupBy.IsNull())
	require.NotEmpty(t, visBv.TreemapConfig.EsqlMetrics)
	assert.Equal(t, "bytes", visBv.TreemapConfig.EsqlMetrics[0].Column.ValueString())
	require.NotEmpty(t, visBv.TreemapConfig.EsqlGroupBy)
	assert.Equal(t, "host.name", visBv.TreemapConfig.EsqlGroupBy[0].Column.ValueString())
	assert.Equal(t, "avg", visBv.TreemapConfig.EsqlGroupBy[0].CollapseBy.ValueString())
}

func Test_treemapConfigModel_truncateAfterLinesIsInt64(t *testing.T) {
	// Verify truncate_after_lines is Int64 (not Float64) per REQ-043.
	api := kbapi.TreemapNoESQL{
		Type:  kbapi.TreemapNoESQLTypeTreemap,
		Query: kbapi.FilterSimple{Expression: "x", Language: func() *kbapi.FilterSimpleLanguage { l := kbapi.FilterSimpleLanguage("kql"); return &l }()},
		Legend: kbapi.TreemapLegend{
			Size:               kbapi.LegendSizeM,
			TruncateAfterLines: new(float32(5)),
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"x"}`), &api.DataSource))
	var metricItem kbapi.TreemapNoESQL_Metrics_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metricItem))
	api.Metrics = []kbapi.TreemapNoESQL_Metrics_Item{metricItem}
	var groupByItem kbapi.TreemapNoESQL_GroupBy_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"terms","field":"host.name","collapse_by":"avg"}`), &groupByItem))
	groupBy := []kbapi.TreemapNoESQL_GroupBy_Item{groupByItem}
	api.GroupBy = &groupBy

	model := &models.TreemapConfigModel{}
	diags := treemapConfigFromAPINoESQL(context.Background(), model, nil, nil, api)
	require.False(t, diags.HasError())
	require.NotNil(t, model.Legend)
	assert.Equal(t, int64(5), model.Legend.TruncateAfterLine.ValueInt64())
}

func Test_treemapConfig_lensChartPresentation_hideTitleRoundTrip(t *testing.T) {
	ctx := context.Background()
	dash := lensPresentationTestDashboard()
	pm := buildLensTreemapPanelForTest(t)

	require.NotNil(t, pm.VisConfig)
	require.NotNil(t, pm.VisConfig.ByValue)
	require.NotNil(t, pm.VisConfig.ByValue.TreemapConfig)
	m := *pm.VisConfig.ByValue.TreemapConfig
	m.HideTitle = types.BoolValue(true)

	attrs, diags := treemapConfigToAPI(&m, dash)
	require.False(t, diags.HasError())
	api, err := attrs.AsTreemapNoESQL()
	require.NoError(t, err)

	got := &models.TreemapConfigModel{}
	require.False(t, treemapConfigFromAPINoESQL(ctx, got, dash, &m, api).HasError())
	assert.Equal(t, types.BoolValue(true), got.HideTitle)
}
