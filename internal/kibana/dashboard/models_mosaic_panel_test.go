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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mosaicPanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip_NoESQL(t *testing.T) {
	ctx := context.Background()

	groupBy := `[{"operation":"terms","collapse_by":"avg","fields":["host.name"],` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	groupBreakdownBy := `[{"operation":"terms","collapse_by":"avg","fields":["service.name"],` +
		`"format":{"type":"number","decimals":2},` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	apiJSON := `{
		"type": "mosaic",
		"title": "Mosaic NoESQL Round-Trip",
		"description": "Converter test",
		"ignore_global_filters": true,
		"sampling": 0.5,
		"dataset": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kuery","query":"status:200"},
		"legend": {"size": "medium"},
		"metrics": [{"operation":"count"}],
		"group_by": ` + groupBy + `,
		"group_breakdown_by": ` + groupBreakdownBy + `
	}`
	var api kbapi.MosaicNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var mosaicChart kbapi.MosaicChart
	require.NoError(t, mosaicChart.FromMosaicNoESQL(api))

	var attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	require.NoError(t, attrs.FromMosaicChart(mosaicChart))

	converter := newMosaicPanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.MosaicConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	chart2, err := attrs2.AsMosaicChart()
	require.NoError(t, err)
	noESQL2, err := chart2.AsMosaicNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "Mosaic NoESQL Round-Trip", *noESQL2.Title)
	assert.Equal(t, kbapi.MosaicNoESQLTypeMosaic, noESQL2.Type)
}

func Test_mosaicPanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip_ESQL(t *testing.T) {
	ctx := context.Background()

	groupBy := `[{"collapse_by":"avg","column":"host.name","operation":"value",` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	groupBreakdownBy := `[{"collapse_by":"avg","column":"service.name","operation":"value",` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	apiJSON := `{
		"type": "mosaic",
		"title": "Mosaic ESQL Round-Trip",
		"description": "Converter test",
		"ignore_global_filters": false,
		"sampling": 1,
		"dataset": {"type":"esql","query":"FROM metrics-* | LIMIT 10"},
		"legend": {"size": "small"},
		"metrics": [{"column":"bytes","operation":"value","format":{"type":"number","decimals":2}}],
		"group_by": ` + groupBy + `,
		"group_breakdown_by": ` + groupBreakdownBy + `
	}`
	var api kbapi.MosaicESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))

	var mosaicChart kbapi.MosaicChart
	require.NoError(t, mosaicChart.FromMosaicESQL(api))

	var attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	require.NoError(t, attrs.FromMosaicChart(mosaicChart))

	converter := newMosaicPanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.MosaicConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	chart2, err := attrs2.AsMosaicChart()
	require.NoError(t, err)
	esql2, err := chart2.AsMosaicESQL()
	require.NoError(t, err)
	assert.Equal(t, "Mosaic ESQL Round-Trip", *esql2.Title)
	assert.Equal(t, kbapi.MosaicESQLTypeMosaic, esql2.Type)
}

func Test_newMosaicPanelConfigConverter(t *testing.T) {
	converter := newMosaicPanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, "mosaic", converter.visualizationType)
}

func Test_mosaicConfigModel_fromAPI_toAPI_noESQL(t *testing.T) {
	api := kbapi.MosaicNoESQL{
		Type:                kbapi.MosaicNoESQLTypeMosaic,
		Title:               new("Test Mosaic"),
		Description:         new("Mosaic description"),
		IgnoreGlobalFilters: new(true),
		Sampling:            new(float32(0.5)),
		Query: kbapi.FilterSimple{
			Query: "status:200",
			Language: func() *kbapi.FilterSimpleLanguage {
				lang := kbapi.FilterSimpleLanguage("kuery")
				return &lang
			}(),
		},
		Legend: kbapi.MosaicLegend{
			Size: kbapi.LegendSizeMedium,
			Nested: func() *bool {
				b := true
				return &b
			}(),
			TruncateAfterLines: new(float32(4)),
			Visible: func() *kbapi.MosaicLegendVisible {
				v := kbapi.MosaicLegendVisibleAuto
				return &v
			}(),
		},
		ValueDisplay: &struct {
			Mode            kbapi.MosaicNoESQLValueDisplayMode `json:"mode"`
			PercentDecimals *float32                           `json:"percent_decimals,omitempty"`
		}{
			Mode:            kbapi.MosaicNoESQLValueDisplayModePercentage,
			PercentDecimals: new(float32(2)),
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.Dataset))

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

	var metricItem kbapi.MosaicNoESQL_Metrics_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metricItem))
	api.Metrics = []kbapi.MosaicNoESQL_Metrics_Item{metricItem}

	model := &mosaicConfigModel{}
	diags := model.fromAPINoESQL(api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Test Mosaic"), model.Title)
	assert.Equal(t, types.StringValue("Mosaic description"), model.Description)
	assert.Equal(t, types.BoolValue(true), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(0.5), model.Sampling)
	require.NotNil(t, model.Query)
	assert.Equal(t, types.StringValue("status:200"), model.Query.Query)
	assert.Equal(t, types.StringValue("kuery"), model.Query.Language)
	assert.False(t, model.Dataset.IsNull())
	assert.False(t, model.GroupBy.IsNull())
	assert.False(t, model.GroupBreakdownBy.IsNull())
	assert.False(t, model.Metrics.IsNull())
	require.NotNil(t, model.Legend)
	assert.Equal(t, types.StringValue("medium"), model.Legend.Size)
	require.NotNil(t, model.ValueDisplay)
	assert.Equal(t, types.StringValue("percentage"), model.ValueDisplay.Mode)
	assert.Equal(t, types.Float64Value(2), model.ValueDisplay.PercentDecimals)

	schema, diags := model.toAPI()
	require.False(t, diags.HasError())

	roundTrip, err := schema.AsMosaicNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.MosaicNoESQLTypeMosaic, roundTrip.Type)
	assert.NotNil(t, roundTrip.GroupBy)
	assert.Len(t, *roundTrip.GroupBy, 1)
	assert.NotNil(t, roundTrip.GroupBreakdownBy)
	assert.Len(t, *roundTrip.GroupBreakdownBy, 1)
	assert.Len(t, roundTrip.Metrics, 1)
}

func Test_mosaicConfigModel_fromAPI_toAPI_esql(t *testing.T) {
	colorMapping := kbapi.ColorMapping{}
	require.NoError(t, json.Unmarshal([]byte(`{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}`), &colorMapping))

	format := kbapi.FormatType{}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number","decimals":2}`), &format))

	groupBy := []struct {
		CollapseBy kbapi.CollapseBy                 `json:"collapse_by"`
		Color      kbapi.ColorMapping               `json:"color"`
		Column     string                           `json:"column"`
		Operation  kbapi.MosaicESQLGroupByOperation `json:"operation"`
	}{
		{
			CollapseBy: kbapi.CollapseByAvg,
			Color:      colorMapping,
			Column:     "host.name",
			Operation:  kbapi.MosaicESQLGroupByOperationValue,
		},
	}

	groupBreakdownBy := []struct {
		CollapseBy kbapi.CollapseBy                          `json:"collapse_by"`
		Color      kbapi.ColorMapping                        `json:"color"`
		Column     string                                    `json:"column"`
		Operation  kbapi.MosaicESQLGroupBreakdownByOperation `json:"operation"`
	}{
		{
			CollapseBy: kbapi.CollapseByAvg,
			Color:      colorMapping,
			Column:     "service.name",
			Operation:  kbapi.MosaicESQLGroupBreakdownByOperationValue,
		},
	}

	metrics := []struct {
		Column    string                           `json:"column"`
		Format    kbapi.FormatType                 `json:"format"`
		Label     *string                          `json:"label,omitempty"`
		Operation kbapi.MosaicESQLMetricsOperation `json:"operation"`
	}{
		{
			Column:    "bytes",
			Format:    format,
			Operation: kbapi.MosaicESQLMetricsOperationValue,
		},
	}

	api := kbapi.MosaicESQL{
		Type:                kbapi.MosaicESQLTypeMosaic,
		Title:               new("ESQL Mosaic"),
		Description:         new("ESQL description"),
		IgnoreGlobalFilters: new(false),
		Sampling:            new(float32(1)),
		Legend:              kbapi.MosaicLegend{Size: kbapi.LegendSizeSmall},
		Metrics:             metrics,
		GroupBy:             &groupBy,
		GroupBreakdownBy:    &groupBreakdownBy,
		ValueDisplay: &struct {
			Mode            kbapi.MosaicESQLValueDisplayMode `json:"mode"`
			PercentDecimals *float32                         `json:"percent_decimals,omitempty"`
		}{
			Mode:            kbapi.MosaicESQLValueDisplayModeAbsolute,
			PercentDecimals: new(float32(1)),
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM metrics-* | LIMIT 10"}`), &api.Dataset))

	model := &mosaicConfigModel{}
	diags := model.fromAPIESQL(api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("ESQL Mosaic"), model.Title)
	assert.Nil(t, model.Query)
	assert.False(t, model.GroupBy.IsNull())
	assert.False(t, model.GroupBreakdownBy.IsNull())
	require.NotNil(t, model.ValueDisplay)
	assert.Equal(t, types.StringValue("absolute"), model.ValueDisplay.Mode)

	schema, diags := model.toAPI()
	require.False(t, diags.HasError())

	roundTrip, err := schema.AsMosaicESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.MosaicESQLTypeMosaic, roundTrip.Type)
	assert.NotNil(t, roundTrip.GroupBy)
	assert.NotNil(t, roundTrip.GroupBreakdownBy)
	assert.Len(t, roundTrip.Metrics, 1)
}

func newTestMosaicNoESQLModel(t *testing.T) *mosaicConfigModel {
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
		"metrics": [{"operation":"count"}],
		"dataset": {"type":"dataView","id":"metrics-*"},
		"query": {"language":"kuery","query":"status:200"},
		"group_by": ` + groupBy + `,
		"group_breakdown_by": ` + groupBreakdownBy + `
	}`
	var api kbapi.MosaicNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
	model := &mosaicConfigModel{}
	require.False(t, model.fromAPINoESQL(api).HasError())
	return model
}

func newTestMosaicESQLModel(t *testing.T) *mosaicConfigModel {
	t.Helper()
	groupBy := `[{"collapse_by":"avg","column":"host.name","operation":"value",` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	groupBreakdownBy := `[{"collapse_by":"avg","column":"service.name","operation":"value",` +
		`"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"color_code","value":"#D3DAE6"}}}]`
	apiJSON := `{
		"type": "mosaic",
		"legend": {"size": "small"},
		"metrics": [{"column":"bytes","operation":"value","format":{"type":"number","decimals":2}}],
		"dataset": {"type":"esql","query":"FROM metrics-* | LIMIT 10"},
		"group_by": ` + groupBy + `,
		"group_breakdown_by": ` + groupBreakdownBy + `
	}`
	var api kbapi.MosaicESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &api))
	model := &mosaicConfigModel{}
	require.False(t, model.fromAPIESQL(api).HasError())
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
		_, diags := model.toAPI()
		requireMosaicMetricsExactlyOneDiagnostic(t, diags)
	})
	t.Run("noESQL_two_items", func(t *testing.T) {
		model := newTestMosaicNoESQLModel(t)
		model.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](
			`[{"operation":"count"},{"operation":"sum","field":"bytes"}]`,
			populatePartitionMetricsDefaults,
		)
		_, diags := model.toAPI()
		requireMosaicMetricsExactlyOneDiagnostic(t, diags)
	})
	t.Run("esql_empty_array", func(t *testing.T) {
		model := newTestMosaicESQLModel(t)
		model.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](`[]`, populatePartitionMetricsDefaults)
		_, diags := model.toAPI()
		requireMosaicMetricsExactlyOneDiagnostic(t, diags)
	})
	t.Run("esql_two_items", func(t *testing.T) {
		model := newTestMosaicESQLModel(t)
		model.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](
			`[{"column":"bytes","operation":"value","format":{"type":"number","decimals":2}},`+
				`{"column":"events","operation":"value","format":{"type":"number","decimals":2}}]`,
			populatePartitionMetricsDefaults,
		)
		_, diags := model.toAPI()
		requireMosaicMetricsExactlyOneDiagnostic(t, diags)
	})
}
