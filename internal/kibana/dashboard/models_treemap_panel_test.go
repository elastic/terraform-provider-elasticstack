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
	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newTreemapPanelConfigConverter(t *testing.T) {
	converter := newTreemapPanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, "treemap", converter.visualizationType)
}

func Test_treemapPanelConfigConverter_roundTrip_populateFromAPIPanel_mapPanelToAPI_noESQL(t *testing.T) {
	t.Parallel()

	attrs := map[string]any{
		"type":                  "treemap",
		"title":                 "Test Treemap",
		"description":           "Treemap description",
		"ignore_global_filters": true,
		"sampling":              0.5,
		"dataset": map[string]any{
			"type": "dataView",
			"id":   "metrics-*",
		},
		"group_by": []any{
			map[string]any{
				"operation":   "terms",
				"collapse_by": "avg",
				"rank_by": map[string]any{
					"type":      "column",
					"metric":    float64(0),
					"direction": "desc",
				},
				"size": float64(5),
				"color": map[string]any{
					"mode":    "categorical",
					"palette": "default",
					"mapping": []any{},
					"unassignedColor": map[string]any{
						"type":  "colorCode",
						"value": "#D3DAE6",
					},
				},
				"fields": []any{"host.name"},
				"format": map[string]any{
					"type":     "number",
					"decimals": float64(2),
				},
			},
		},
		"metrics": []any{
			map[string]any{
				"operation": "count",
			},
		},
		"label_position": "visible",
		"legend": map[string]any{
			"nested":              true,
			"size":                "medium",
			"truncate_after_lines": 4.0,
			"visible":             "auto",
		},
		"value_display": map[string]any{
			"mode":             "percentage",
			"percent_decimals": 2.0,
		},
		"query": map[string]any{
			"query":    "status:200",
			"language": "kuery",
		},
		"filters": []any{
			map[string]any{
				"query":    "response:200",
				"language": "kuery",
				"meta": map[string]any{
					"disabled": false,
					"negate":   false,
					"alias":    nil,
				},
			},
		},
	}

	apiConfig := dashboardPanelItemConfigFromAttributes(t, attrs)

	converter := newTreemapPanelConfigConverter()
	var pm panelModel
	diags := converter.populateFromAPIPanel(context.Background(), &pm, apiConfig)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.TreemapConfig)

	assert.Equal(t, types.StringValue("Test Treemap"), pm.TreemapConfig.Title)
	assert.Equal(t, types.StringValue("Treemap description"), pm.TreemapConfig.Description)
	assert.Equal(t, types.BoolValue(true), pm.TreemapConfig.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(0.5), pm.TreemapConfig.Sampling)

	require.NotNil(t, pm.TreemapConfig.Query)
	assert.Equal(t, types.StringValue("status:200"), pm.TreemapConfig.Query.Query)
	assert.Equal(t, types.StringValue("kuery"), pm.TreemapConfig.Query.Language)

	require.False(t, pm.TreemapConfig.Dataset.IsNull())
	assert.Equal(t, normalizeAny(t, attrs["dataset"]), normalizeAny(t, mustUnmarshalJSON(t, pm.TreemapConfig.Dataset.ValueString())))

	require.False(t, pm.TreemapConfig.GroupBy.IsNull())
	assert.Equal(t, normalizeAny(t, attrs["group_by"]), normalizeAny(t, mustUnmarshalJSON(t, pm.TreemapConfig.GroupBy.ValueString())))

	require.False(t, pm.TreemapConfig.Metrics.IsNull())
	assert.Equal(t, normalizeAny(t, attrs["metrics"]), normalizeAny(t, mustUnmarshalJSON(t, pm.TreemapConfig.Metrics.ValueString())))

	assert.Equal(t, types.StringValue("visible"), pm.TreemapConfig.LabelPosition)

	require.NotNil(t, pm.TreemapConfig.Legend)
	assert.Equal(t, types.BoolValue(true), pm.TreemapConfig.Legend.Nested)
	assert.Equal(t, types.StringValue("medium"), pm.TreemapConfig.Legend.Size)
	assert.Equal(t, types.Float64Value(4), pm.TreemapConfig.Legend.TruncateAfterLine)
	assert.Equal(t, types.StringValue("auto"), pm.TreemapConfig.Legend.Visible)

	require.NotNil(t, pm.TreemapConfig.ValueDisplay)
	assert.Equal(t, types.StringValue("percentage"), pm.TreemapConfig.ValueDisplay.Mode)
	assert.Equal(t, types.Float64Value(2), pm.TreemapConfig.ValueDisplay.PercentDecimals)

	require.Len(t, pm.TreemapConfig.Filters, 1)
	assert.Equal(t, types.StringValue("response:200"), pm.TreemapConfig.Filters[0].Query)
	assert.Equal(t, types.StringValue("kuery"), pm.TreemapConfig.Filters[0].Language)
	require.False(t, pm.TreemapConfig.Filters[0].MetaJSON.IsNull())
	assert.Equal(t, normalizeAny(t, attrs["filters"].([]any)[0].(map[string]any)["meta"]), normalizeAny(t, mustUnmarshalJSON(t, pm.TreemapConfig.Filters[0].MetaJSON.ValueString())))

	var roundTripConfig kbapi.DashboardPanelItem_Config
	diags = converter.mapPanelToAPI(pm, &roundTripConfig)
	require.False(t, diags.HasError())

	roundTripAttrs := dashboardPanelItemAttributes(t, roundTripConfig)
	assert.Equal(t, normalizeAny(t, attrs), normalizeAny(t, roundTripAttrs))
}

func Test_treemapPanelConfigConverter_roundTrip_populateFromAPIPanel_mapPanelToAPI_esql(t *testing.T) {
	t.Parallel()

	attrs := map[string]any{
		"type":                  "treemap",
		"title":                 "ESQL Treemap",
		"description":           "ESQL description",
		"ignore_global_filters": false,
		"sampling":              1.0,
		"dataset": map[string]any{
			"type":  "esql",
			"query": "FROM metrics-* | KEEP host.name, bytes | LIMIT 10",
		},
		"group_by": []any{
			map[string]any{
				"operation":   "value",
				"collapse_by": "avg",
				"column":      "host.name",
				"color": map[string]any{
					"mode":    "categorical",
					"palette": "default",
					"mapping": []any{},
					"unassignedColor": map[string]any{
						"type":  "colorCode",
						"value": "#D3DAE6",
					},
				},
			},
		},
		"metrics": []any{
			map[string]any{
				"operation": "value",
				"column":    "bytes",
				"format": map[string]any{
					"type":     "number",
					"decimals": float64(2),
				},
				"color": map[string]any{
					"type":  "static",
					"color": "#54B399",
				},
			},
		},
		"label_position": "hidden",
		"legend": map[string]any{
			"nested":              false,
			"size":                "small",
		},
		"value_display": map[string]any{
			"mode":             "absolute",
			"percent_decimals": 1.0,
		},
		"filters": []any{
			map[string]any{
				"query":    "host.name: test-host",
				"language": "kuery",
				"meta": map[string]any{
					"disabled": false,
					"negate":   true,
					"alias":    "host filter",
				},
			},
		},
	}

	apiConfig := dashboardPanelItemConfigFromAttributes(t, attrs)

	converter := newTreemapPanelConfigConverter()
	var pm panelModel
	diags := converter.populateFromAPIPanel(context.Background(), &pm, apiConfig)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.TreemapConfig)

	assert.Equal(t, types.StringValue("ESQL Treemap"), pm.TreemapConfig.Title)
	assert.Equal(t, types.StringValue("ESQL description"), pm.TreemapConfig.Description)
	assert.Equal(t, types.BoolValue(false), pm.TreemapConfig.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(1), pm.TreemapConfig.Sampling)
	assert.Nil(t, pm.TreemapConfig.Query)

	require.False(t, pm.TreemapConfig.Dataset.IsNull())
	assert.Equal(t, normalizeAny(t, attrs["dataset"]), normalizeAny(t, mustUnmarshalJSON(t, pm.TreemapConfig.Dataset.ValueString())))

	require.False(t, pm.TreemapConfig.GroupBy.IsNull())
	assert.Equal(t, normalizeAny(t, attrs["group_by"]), normalizeAny(t, mustUnmarshalJSON(t, pm.TreemapConfig.GroupBy.ValueString())))

	require.False(t, pm.TreemapConfig.Metrics.IsNull())
	assert.Equal(t, normalizeAny(t, attrs["metrics"]), normalizeAny(t, mustUnmarshalJSON(t, pm.TreemapConfig.Metrics.ValueString())))

	assert.Equal(t, types.StringValue("hidden"), pm.TreemapConfig.LabelPosition)

	require.NotNil(t, pm.TreemapConfig.Legend)
	assert.Equal(t, types.BoolValue(false), pm.TreemapConfig.Legend.Nested)
	assert.Equal(t, types.StringValue("small"), pm.TreemapConfig.Legend.Size)
	assert.True(t, pm.TreemapConfig.Legend.TruncateAfterLine.IsNull())
	assert.True(t, pm.TreemapConfig.Legend.Visible.IsNull())

	require.NotNil(t, pm.TreemapConfig.ValueDisplay)
	assert.Equal(t, types.StringValue("absolute"), pm.TreemapConfig.ValueDisplay.Mode)
	assert.Equal(t, types.Float64Value(1), pm.TreemapConfig.ValueDisplay.PercentDecimals)

	require.Len(t, pm.TreemapConfig.Filters, 1)
	assert.Equal(t, types.StringValue("host.name: test-host"), pm.TreemapConfig.Filters[0].Query)
	assert.Equal(t, types.StringValue("kuery"), pm.TreemapConfig.Filters[0].Language)
	require.False(t, pm.TreemapConfig.Filters[0].MetaJSON.IsNull())
	assert.Equal(t, normalizeAny(t, attrs["filters"].([]any)[0].(map[string]any)["meta"]), normalizeAny(t, mustUnmarshalJSON(t, pm.TreemapConfig.Filters[0].MetaJSON.ValueString())))

	var roundTripConfig kbapi.DashboardPanelItem_Config
	diags = converter.mapPanelToAPI(pm, &roundTripConfig)
	require.False(t, diags.HasError())

	roundTripAttrs := dashboardPanelItemAttributes(t, roundTripConfig)
	assert.Equal(t, normalizeAny(t, attrs), normalizeAny(t, roundTripAttrs))
}

func dashboardPanelItemConfigFromAttributes(t *testing.T, attributes map[string]any) kbapi.DashboardPanelItem_Config {
	t.Helper()

	configMap := map[string]any{
		"attributes": attributes,
	}

	// The generated kbapi unions can be picky about the exact shape; populate the union
	// via the From* helper then also JSON-roundtrip the map into the union for parity
	// with the other panel converter tests in this package.
	configJSON, err := json.Marshal(configMap)
	require.NoError(t, err)

	var config kbapi.DashboardPanelItem_Config
	require.NoError(t, config.FromDashboardPanelItemConfig2(configMap))
	require.NoError(t, json.Unmarshal(configJSON, &config))
	return config
}

func dashboardPanelItemAttributes(t *testing.T, config kbapi.DashboardPanelItem_Config) map[string]any {
	t.Helper()

	cfgMap, err := config.AsDashboardPanelItemConfig2()
	require.NoError(t, err)
	attrs, ok := cfgMap["attributes"].(map[string]any)
	require.True(t, ok)
	return attrs
}

func mustUnmarshalJSON(t *testing.T, s string) any {
	t.Helper()
	var v any
	require.NoError(t, json.Unmarshal([]byte(s), &v))
	return v
}

func normalizeAny(t *testing.T, v any) any {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	var out any
	require.NoError(t, json.Unmarshal(b, &out))
	return out
}

func Test_treemapConfigModel_fromAPI_toAPI_noESQL(t *testing.T) {
	api := kbapi.TreemapNoESQL{
		Type:                kbapi.TreemapNoESQLTypeTreemap,
		Title:               schemautil.Pointer("Test Treemap"),
		Description:         schemautil.Pointer("Treemap description"),
		IgnoreGlobalFilters: schemautil.Pointer(true),
		Sampling:            schemautil.Pointer(float32(0.5)),
		Query: kbapi.FilterSimpleSchema{
			Query: "status:200",
			Language: func() *kbapi.FilterSimpleSchemaLanguage {
				lang := kbapi.FilterSimpleSchemaLanguage("kuery")
				return &lang
			}(),
		},
		Legend: kbapi.TreemapLegend{
			Size: kbapi.LegendSizeMedium,
			Nested: func() *bool {
				b := true
				return &b
			}(),
			TruncateAfterLines: schemautil.Pointer(float32(4)),
			Visible: func() *kbapi.TreemapLegendVisible {
				v := kbapi.TreemapLegendVisibleAuto
				return &v
			}(),
		},
		ValueDisplay: &struct {
			Mode            kbapi.TreemapNoESQLValueDisplayMode `json:"mode"`
			PercentDecimals *float32                            `json:"percent_decimals,omitempty"`
		}{
			Mode:            kbapi.TreemapNoESQLValueDisplayModePercentage,
			PercentDecimals: schemautil.Pointer(float32(2)),
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.Dataset))

	var groupByItem kbapi.TreemapNoESQL_GroupBy_Item
	require.NoError(t, json.Unmarshal([]byte(`{
		"operation":"terms",
		"collapse_by":"avg",
		"color":{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"colorCode","value":"#D3DAE6"}},
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
	require.NoError(t, json.Unmarshal([]byte(`{"mode":"categorical","palette":"default","mapping":[],"unassignedColor":{"type":"colorCode","value":"#D3DAE6"}}`), &colorMapping))

	staticColor := kbapi.StaticColor{}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"static","color":"#54B399"}`), &staticColor))

	format := kbapi.FormatTypeSchema{}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number","decimals":2}`), &format))

	groupBy := []struct {
		CollapseBy kbapi.CollapseBy                  `json:"collapse_by"`
		Color      kbapi.ColorMapping                `json:"color"`
		Column     string                            `json:"column"`
		Operation  kbapi.TreemapESQLGroupByOperation `json:"operation"`
	}{
		{
			CollapseBy: kbapi.CollapseByAvg,
			Color:      colorMapping,
			Column:     "host.name",
			Operation:  kbapi.TreemapESQLGroupByOperationValue,
		},
	}

	metrics := []struct {
		Color     kbapi.StaticColor                 `json:"color"`
		Column    string                            `json:"column"`
		Format    kbapi.FormatTypeSchema            `json:"format"`
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
		Title:               schemautil.Pointer("ESQL Treemap"),
		Description:         schemautil.Pointer("ESQL description"),
		IgnoreGlobalFilters: schemautil.Pointer(false),
		Sampling:            schemautil.Pointer(float32(1)),
		Legend:              kbapi.TreemapLegend{Size: kbapi.LegendSizeSmall},
		Metrics:             metrics,
		GroupBy:             &groupBy,
		ValueDisplay: &struct {
			Mode            kbapi.TreemapESQLValueDisplayMode `json:"mode"`
			PercentDecimals *float32                          `json:"percent_decimals,omitempty"`
		}{
			Mode: kbapi.TreemapESQLValueDisplayModeAbsolute,
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
