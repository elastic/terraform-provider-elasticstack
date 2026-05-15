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
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mosaicConfigModel_fromAPI_toAPI_noESQL(t *testing.T) {
	api := kbapi.MosaicNoESQL{
		Type:                kbapi.MosaicNoESQLTypeMosaic,
		Title:               new("Test Mosaic"),
		Description:         new("Mosaic description"),
		IgnoreGlobalFilters: new(true),
		Sampling:            new(float32(0.5)),
		Query: kbapi.FilterSimple{
			Expression: "status:200",
			Language: func() *kbapi.FilterSimpleLanguage {
				lang := kbapi.FilterSimpleLanguage("kql")
				return &lang
			}(),
		},
		Legend: kbapi.MosaicLegend{
			Size: kbapi.LegendSizeM,
			Nested: func() *bool {
				b := true
				return &b
			}(),
			TruncateAfterLines: new(float32(4)),
			Visibility: func() *kbapi.MosaicLegendVisibility {
				v := kbapi.MosaicLegendVisibilityAuto
				return &v
			}(),
		},
		Styling: kbapi.MosaicStyling{
			Values: kbapi.ValueDisplay{
				Mode:            func() *kbapi.ValueDisplayMode { m := kbapi.ValueDisplayModePercentage; return &m }(),
				PercentDecimals: new(float32(2)),
			},
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.DataSource))

	var groupByItem kbapi.MosaicNoESQL_GroupBy_Item
	require.NoError(t, json.Unmarshal([]byte(`{
		"operation":"terms",
		"collapse_by":"avg",
		"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}},
		"fields":["host.name"],
		"format":{"type":"number","decimals":2}
	}`), &groupByItem))
	groupBy := []kbapi.MosaicNoESQL_GroupBy_Item{groupByItem}
	api.GroupBy = &groupBy

	var groupBreakdownByItem kbapi.MosaicNoESQL_GroupBreakdownBy_Item
	require.NoError(t, json.Unmarshal([]byte(`{
		"operation":"terms",
		"collapse_by":"avg",
		"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}},
		"fields":["service.name"],
		"format":{"type":"number","decimals":2}
	}`), &groupBreakdownByItem))
	groupBreakdownBy := []kbapi.MosaicNoESQL_GroupBreakdownBy_Item{groupBreakdownByItem}
	api.GroupBreakdownBy = &groupBreakdownBy

	var metricUnion kbapi.MosaicNoESQL_Metric
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metricUnion))
	api.Metric = metricUnion

	model := &models.MosaicConfigModel{}
	diags := mosaicConfigFromAPINoESQL(context.Background(), model, nil, nil, api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Test Mosaic"), model.Title)
	assert.Equal(t, types.StringValue("Mosaic description"), model.Description)
	assert.Equal(t, types.BoolValue(true), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(0.5), model.Sampling)
	require.NotNil(t, model.Query)
	assert.Equal(t, types.StringValue("status:200"), model.Query.Expression)
	assert.Equal(t, types.StringValue("kql"), model.Query.Language)
	assert.False(t, model.DataSourceJSON.IsNull())
	assert.False(t, model.GroupBy.IsNull())
	assert.False(t, model.GroupBreakdownBy.IsNull())
	assert.False(t, model.Metrics.IsNull())
	require.NotNil(t, model.Legend)
	assert.Equal(t, types.StringValue("m"), model.Legend.Size)
	require.NotNil(t, model.ValueDisplay)
	assert.Equal(t, types.StringValue("percentage"), model.ValueDisplay.Mode)
	assert.Equal(t, types.Float64Value(2), model.ValueDisplay.PercentDecimals)

	attrs, diags := mosaicConfigToAPI(model, nil)
	require.False(t, diags.HasError())

	roundTrip, err := attrs.AsMosaicNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.MosaicNoESQLTypeMosaic, roundTrip.Type)
	assert.NotNil(t, roundTrip.GroupBy)
	assert.Len(t, *roundTrip.GroupBy, 1)
	assert.NotNil(t, roundTrip.GroupBreakdownBy)
	assert.Len(t, *roundTrip.GroupBreakdownBy, 1)
	opBytes, err := roundTrip.Metric.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, string(opBytes), "count")
}

func Test_mosaicConfigModel_fromAPI_toAPI_esql(t *testing.T) {
	colorMapping := kbapi.ColorMapping{}
	require.NoError(t, json.Unmarshal([]byte(`{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}`), &colorMapping))

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

	groupBreakdownBy := []struct {
		CollapseBy kbapi.CollapseBy   `json:"collapse_by"`
		Color      kbapi.ColorMapping `json:"color"`
		Column     string             `json:"column"`
		Format     kbapi.FormatType   `json:"format"`
		Label      *string            `json:"label,omitempty"`
	}{
		{
			CollapseBy: kbapi.CollapseByAvg,
			Color:      colorMapping,
			Column:     "service.name",
			Format:     format,
		},
	}

	api := kbapi.MosaicESQL{
		Type:                kbapi.MosaicESQLTypeMosaic,
		Title:               new("ESQL Mosaic"),
		Description:         new("ESQL description"),
		IgnoreGlobalFilters: new(false),
		Sampling:            new(float32(1)),
		Legend:              kbapi.MosaicLegend{Size: kbapi.LegendSizeS},
		Metric: struct {
			Column string           `json:"column"`
			Format kbapi.FormatType `json:"format"`
			Label  *string          `json:"label,omitempty"`
		}{
			Column: "bytes",
			Format: format,
		},
		GroupBy:          &groupBy,
		GroupBreakdownBy: &groupBreakdownBy,
		Styling: kbapi.MosaicStyling{
			Values: kbapi.ValueDisplay{
				Mode:            func() *kbapi.ValueDisplayMode { m := kbapi.ValueDisplayModeAbsolute; return &m }(),
				PercentDecimals: new(float32(1)),
			},
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM metrics-* | LIMIT 10"}`), &api.DataSource))

	model := &models.MosaicConfigModel{}
	diags := mosaicConfigFromAPIESQL(context.Background(), model, nil, nil, api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("ESQL Mosaic"), model.Title)
	assert.Nil(t, model.Query)
	assert.True(t, model.GroupBy.IsNull())
	assert.True(t, model.Metrics.IsNull())
	assert.NotEmpty(t, model.EsqlMetrics)
	assert.NotEmpty(t, model.EsqlGroupBy)
	assert.False(t, model.GroupBreakdownBy.IsNull())
	require.NotNil(t, model.ValueDisplay)
	assert.Equal(t, types.StringValue("absolute"), model.ValueDisplay.Mode)

	attrs, diags := mosaicConfigToAPI(model, nil)
	require.False(t, diags.HasError())

	roundTrip, err := attrs.AsMosaicESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.MosaicESQLTypeMosaic, roundTrip.Type)
	assert.NotNil(t, roundTrip.GroupBy)
	assert.NotNil(t, roundTrip.GroupBreakdownBy)
	assert.Equal(t, "bytes", roundTrip.Metric.Column)
}

func newTestMosaicNoESQLModel(t *testing.T) *models.MosaicConfigModel {
	t.Helper()
	groupBy := `[{"operation":"terms","collapse_by":"avg","fields":["host.name"],` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	groupBreakdownBy := `[{"operation":"terms","collapse_by":"avg","fields":["service.name"],` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	apiJSON := `{
		"type": "mosaic",
		"legend": {"size": "medium"},
		"metric": {"operation":"count"},
		"data_source": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kql","query":"status:200"},
		"group_by": ` + groupBy + `,
		"group_breakdown_by": ` + groupBreakdownBy + `
	}`
	var api kbapi.MosaicNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
	model := &models.MosaicConfigModel{}
	require.False(t, mosaicConfigFromAPINoESQL(context.Background(), model, nil, nil, api).HasError())
	return model
}

func newTestMosaicESQLModel(t *testing.T) *models.MosaicConfigModel {
	t.Helper()
	groupBy := `[{"collapse_by":"avg","column":"host.name","operation":"value",` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	groupBreakdownBy := `[{"collapse_by":"avg","column":"service.name","operation":"value",` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	apiJSON := `{
		"type": "mosaic",
		"legend": {"size": "small"},
		"metric": {"column":"bytes","operation":"value","format":{"type":"number","decimals":2}},
		"data_source": {"type":"esql","query":"FROM metrics-* | LIMIT 10"},
		"group_by": ` + groupBy + `,
		"group_breakdown_by": ` + groupBreakdownBy + `
	}`
	var api kbapi.MosaicESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
	model := &models.MosaicConfigModel{}
	require.False(t, mosaicConfigFromAPIESQL(context.Background(), model, nil, nil, api).HasError())
	return model
}

func requireMosaicMetricsExactlyOneDiagnostic(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	require.True(t, diags.HasError())
	var found bool
	for _, d := range diags.Errors() {
		if d.Summary() == "Invalid metrics_json" && strings.Contains(d.Detail(), "exactly one") {
			found = true
			break
		}
	}
	require.True(t, found, "expected Invalid metrics_json with 'exactly one' in detail, got %#v", diags)
}

func Test_mosaicConfigModel_toAPI_metrics_json_exactly_one(t *testing.T) {
	t.Run("noESQL_empty_array", func(t *testing.T) {
		model := newTestMosaicNoESQLModel(t)
		model.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](`[]`, populatePartitionMetricsDefaults)
		_, diags := mosaicConfigToAPI(model, nil)
		requireMosaicMetricsExactlyOneDiagnostic(t, diags)
	})
	t.Run("noESQL_two_items", func(t *testing.T) {
		model := newTestMosaicNoESQLModel(t)
		model.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](
			`[{"operation":"count"},{"operation":"sum","field":"bytes"}]`,
			populatePartitionMetricsDefaults,
		)
		_, diags := mosaicConfigToAPI(model, nil)
		requireMosaicMetricsExactlyOneDiagnostic(t, diags)
	})
	t.Run("esql_empty_array", func(t *testing.T) {
		model := newTestMosaicESQLModel(t)
		model.EsqlMetrics = nil
		_, diags := mosaicConfigToAPI(model, nil)
		require.True(t, diags.HasError())
		var found bool
		for _, d := range diags.Errors() {
			if d.Summary() == "Invalid esql_metrics" && strings.Contains(d.Detail(), "exactly one") {
				found = true
				break
			}
		}
		require.True(t, found, "expected Invalid esql_metrics with 'exactly one' in detail, got %#v", diags)
	})
	t.Run("esql_two_items", func(t *testing.T) {
		model := newTestMosaicESQLModel(t)
		model.EsqlMetrics = append(model.EsqlMetrics, models.MosaicEsqlMetric{Column: types.StringValue("extra")})
		_, diags := mosaicConfigToAPI(model, nil)
		require.True(t, diags.HasError())
		var found bool
		for _, d := range diags.Errors() {
			if d.Summary() == "Invalid esql_metrics" && strings.Contains(d.Detail(), "exactly one") {
				found = true
				break
			}
		}
		require.True(t, found, "expected Invalid esql_metrics with 'exactly one' in detail, got %#v", diags)
	})
}

func Test_mosaicConfigModel_esqlTypedMetricsRoundTrip(t *testing.T) {
	// Verifies esql_metrics typed nested attribute round-trips correctly for mosaic.
	groupBy := `[{"collapse_by":"avg","column":"host.name","operation":"value",` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	groupBreakdownBy := `[{"collapse_by":"avg","column":"service.name","operation":"value",` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	apiJSON := `{
		"type": "mosaic",
		"title": "ESQL Mosaic Typed Test",
		"description": "test",
		"ignore_global_filters": false,
		"sampling": 1,
		"data_source": {"type":"esql","query":"FROM metrics-* | LIMIT 10"},
		"legend": {"size": "small"},
		"metric": {"column":"bytes","operation":"value","format":{"type":"number","decimals":2}},
		"group_by": ` + groupBy + `,
		"group_breakdown_by": ` + groupBreakdownBy + `
	}`
	var api kbapi.MosaicESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromMosaicESQL(api))

	c := lenscommon.ForType(string(kbapi.MosaicNoESQLTypeMosaic))
	require.NotNil(t, c)
	visBv := models.VisByValueModel{}
	diags := c.PopulateFromAttributes(context.Background(), lensChartResolver(nil), &visBv.LensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, visBv.MosaicConfig)

	// ES|QL mode: EsqlMetrics populated, Metrics/GroupBy null, Query nil
	assert.Nil(t, visBv.MosaicConfig.Query)
	assert.True(t, visBv.MosaicConfig.Metrics.IsNull())
	assert.True(t, visBv.MosaicConfig.GroupBy.IsNull())
	require.NotEmpty(t, visBv.MosaicConfig.EsqlMetrics)
	assert.Equal(t, "bytes", visBv.MosaicConfig.EsqlMetrics[0].Column.ValueString())
	require.NotEmpty(t, visBv.MosaicConfig.EsqlGroupBy)
	assert.Equal(t, "host.name", visBv.MosaicConfig.EsqlGroupBy[0].Column.ValueString())
}

func Test_mosaicConfigModel_truncateAfterLinesIsInt64(t *testing.T) {
	// Verify truncate_after_lines is Int64 (not Float64) per REQ-043.
	api := kbapi.MosaicNoESQL{
		Type:  kbapi.MosaicNoESQLTypeMosaic,
		Query: kbapi.FilterSimple{Expression: "x", Language: func() *kbapi.FilterSimpleLanguage { l := kbapi.FilterSimpleLanguage("kql"); return &l }()},
		Legend: kbapi.MosaicLegend{
			Size:               kbapi.LegendSizeM,
			TruncateAfterLines: new(float32(5)),
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"x"}`), &api.DataSource))
	var metricUnion kbapi.MosaicNoESQL_Metric
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metricUnion))
	api.Metric = metricUnion
	var groupByItem kbapi.MosaicNoESQL_GroupBy_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"terms","field":"host.name","collapse_by":"avg"}`), &groupByItem))
	groupBy := []kbapi.MosaicNoESQL_GroupBy_Item{groupByItem}
	api.GroupBy = &groupBy
	var groupBreakdownByItem kbapi.MosaicNoESQL_GroupBreakdownBy_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"terms","field":"service.name","collapse_by":"avg"}`), &groupBreakdownByItem))
	groupBreakdownBy := []kbapi.MosaicNoESQL_GroupBreakdownBy_Item{groupBreakdownByItem}
	api.GroupBreakdownBy = &groupBreakdownBy

	model := &models.MosaicConfigModel{}
	diags := mosaicConfigFromAPINoESQL(context.Background(), model, nil, nil, api)
	require.False(t, diags.HasError())
	require.NotNil(t, model.Legend)
	assert.Equal(t, int64(5), model.Legend.TruncateAfterLine.ValueInt64())
}

func Test_mosaicConfig_lensChartPresentation_hideTitleRoundTrip(t *testing.T) {
	ctx := context.Background()
	dash := lensPresentationTestDashboard()
	pm := buildLensMosaicPanelForTest(t)

	require.NotNil(t, pm.VisConfig)
	require.NotNil(t, pm.VisConfig.ByValue)
	require.NotNil(t, pm.VisConfig.ByValue.MosaicConfig)
	m := *pm.VisConfig.ByValue.MosaicConfig
	m.HideTitle = types.BoolValue(true)

	attrs, diags := mosaicConfigToAPI(&m, dash)
	require.False(t, diags.HasError())
	api, err := attrs.AsMosaicNoESQL()
	require.NoError(t, err)

	got := &models.MosaicConfigModel{}
	require.False(t, mosaicConfigFromAPINoESQL(ctx, got, dash, &m, api).HasError())
	assert.Equal(t, types.BoolValue(true), got.HideTitle)
}
