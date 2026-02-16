package dashboard

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newMetricChartPanelConfigConverter(t *testing.T) {
	converter := newMetricChartPanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, string(kbapi.MetricChartSchema0TypeMetric), converter.visualizationType)
}

func Test_metricChartConfigModel_fromAPI_toAPI_variant0(t *testing.T) {
	tests := []struct {
		name             string
		apiChart         kbapi.MetricChartSchema0
		expectedTitle    string
		expectedDesc     string
		expectedSampling float64
	}{
		{
			name: "basic metric chart with all fields",
			apiChart: kbapi.MetricChartSchema0{
				Type:                kbapi.MetricChartSchema0TypeMetric,
				Title:               utils.Pointer("Test Metric"),
				Description:         utils.Pointer("Test Description"),
				IgnoreGlobalFilters: utils.Pointer(false),
				Sampling:            utils.Pointer(float32(1.0)),
				Query: kbapi.FilterSimpleSchema{
					Language: utils.Pointer(kbapi.FilterSimpleSchemaLanguage("kuery")),
					Query:    "",
				},
				Metrics: []kbapi.MetricChartSchema_0_Metrics_Item{},
			},
			expectedTitle:    "Test Metric",
			expectedDesc:     "Test Description",
			expectedSampling: 1.0,
		},
		{
			name: "minimal metric chart",
			apiChart: kbapi.MetricChartSchema0{
				Type: kbapi.MetricChartSchema0TypeMetric,
				Query: kbapi.FilterSimpleSchema{
					Query: "",
				},
				Metrics: []kbapi.MetricChartSchema_0_Metrics_Item{},
			},
			expectedTitle:    "",
			expectedDesc:     "",
			expectedSampling: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Convert API to metric chart schema
			var apiSchema kbapi.MetricChartSchema
			err := apiSchema.FromMetricChartSchema0(tt.apiChart)
			require.NoError(t, err)

			// Test fromAPI
			model := &metricChartConfigModel{}
			diags := model.fromAPI(ctx, apiSchema)
			require.False(t, diags.HasError(), "fromAPI should not have errors")

			// Verify the fields
			if tt.expectedTitle != "" {
				assert.Equal(t, tt.expectedTitle, model.Title.ValueString())
			} else {
				assert.True(t, model.Title.IsNull())
			}

			if tt.expectedDesc != "" {
				assert.Equal(t, tt.expectedDesc, model.Description.ValueString())
			} else {
				assert.True(t, model.Description.IsNull())
			}

			if tt.expectedSampling > 0 {
				assert.Equal(t, tt.expectedSampling, model.Sampling.ValueFloat64())
			} else {
				assert.True(t, model.Sampling.IsNull())
			}

			// Test toAPI round-trip
			resultSchema, diags := model.toAPI()
			require.False(t, diags.HasError(), "toAPI should not have errors")

			// Verify we can convert back to variant 0
			resultVariant0, err := resultSchema.AsMetricChartSchema0()
			require.NoError(t, err)

			assert.Equal(t, tt.apiChart.Type, resultVariant0.Type)
			if tt.apiChart.Title != nil {
				require.NotNil(t, resultVariant0.Title)
				assert.Equal(t, *tt.apiChart.Title, *resultVariant0.Title)
			}
			if tt.apiChart.Description != nil {
				require.NotNil(t, resultVariant0.Description)
				assert.Equal(t, *tt.apiChart.Description, *resultVariant0.Description)
			}
		})
	}
}

func Test_metricChartConfigModel_fromAPI_toAPI_variant1(t *testing.T) {
	tests := []struct {
		name             string
		apiChart         kbapi.MetricChartSchema1
		expectedTitle    string
		expectedDesc     string
		expectedSampling float64
	}{
		{
			name: "ESQL metric chart with all fields",
			apiChart: kbapi.MetricChartSchema1{
				Type:                kbapi.MetricChartSchema1TypeMetric,
				Title:               utils.Pointer("ESQL Metric"),
				Description:         utils.Pointer("ESQL Description"),
				IgnoreGlobalFilters: utils.Pointer(true),
				Sampling:            utils.Pointer(float32(0.5)),
				Dataset: func() kbapi.MetricChartSchema_1_Dataset {
					var ds kbapi.MetricChartSchema_1_Dataset
					_ = ds.FromEsqlDatasetTypeSchema(kbapi.EsqlDatasetTypeSchema{
						Type:  kbapi.EsqlDatasetTypeSchemaTypeEsql,
						Query: "FROM logs-*",
					})
					return ds
				}(),
				Metrics: []kbapi.MetricChartSchema_1_Metrics_Item{},
			},
			expectedTitle:    "ESQL Metric",
			expectedDesc:     "ESQL Description",
			expectedSampling: 0.5,
		},
		{
			name: "minimal ESQL metric chart",
			apiChart: kbapi.MetricChartSchema1{
				Type: kbapi.MetricChartSchema1TypeMetric,
				Dataset: func() kbapi.MetricChartSchema_1_Dataset {
					var ds kbapi.MetricChartSchema_1_Dataset
					_ = ds.FromEsqlDatasetTypeSchema(kbapi.EsqlDatasetTypeSchema{
						Type:  kbapi.EsqlDatasetTypeSchemaTypeEsql,
						Query: "FROM *",
					})
					return ds
				}(),
				Metrics: []kbapi.MetricChartSchema_1_Metrics_Item{},
			},
			expectedTitle:    "",
			expectedDesc:     "",
			expectedSampling: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Convert API to metric chart schema
			var apiSchema kbapi.MetricChartSchema
			err := apiSchema.FromMetricChartSchema1(tt.apiChart)
			require.NoError(t, err)

			// Test fromAPI
			model := &metricChartConfigModel{}
			diags := model.fromAPI(ctx, apiSchema)
			require.False(t, diags.HasError(), "fromAPI should not have errors")

			// Verify the fields
			if tt.expectedTitle != "" {
				assert.Equal(t, tt.expectedTitle, model.Title.ValueString())
			} else {
				assert.True(t, model.Title.IsNull())
			}

			if tt.expectedDesc != "" {
				assert.Equal(t, tt.expectedDesc, model.Description.ValueString())
			} else {
				assert.True(t, model.Description.IsNull())
			}

			if tt.expectedSampling > 0 {
				assert.Equal(t, tt.expectedSampling, model.Sampling.ValueFloat64())
			} else {
				assert.True(t, model.Sampling.IsNull())
			}

			// Verify query is nil for variant 1
			assert.Nil(t, model.Query)

			// Test toAPI round-trip
			resultSchema, diags := model.toAPI()
			require.False(t, diags.HasError(), "toAPI should not have errors")

			// Verify we can convert back to variant 1
			resultVariant1, err := resultSchema.AsMetricChartSchema1()
			require.NoError(t, err)

			assert.Equal(t, tt.apiChart.Type, resultVariant1.Type)
			if tt.apiChart.Title != nil {
				require.NotNil(t, resultVariant1.Title)
				assert.Equal(t, *tt.apiChart.Title, *resultVariant1.Title)
			}
			if tt.apiChart.Description != nil {
				require.NotNil(t, resultVariant1.Description)
				assert.Equal(t, *tt.apiChart.Description, *resultVariant1.Description)
			}
		})
	}
}

func Test_metricChartConfigModel_withMetrics(t *testing.T) {
	ctx := context.Background()

	// Create a metric with primary metric configuration
	metricJSON := `{
		"type": "primary",
		"operation": "count",
		"format": {"id": "number"},
		"alignments": {"value": "center"},
		"icon": {"name": "empty"}
	}`

	var metricItem kbapi.MetricChartSchema_1_Metrics_Item
	err := json.Unmarshal([]byte(metricJSON), &metricItem)
	require.NoError(t, err)

	apiChart := kbapi.MetricChartSchema1{
		Type:    kbapi.MetricChartSchema1TypeMetric,
		Title:   utils.Pointer("Test with Metrics"),
		Metrics: []kbapi.MetricChartSchema_1_Metrics_Item{metricItem},
	}

	var apiSchema kbapi.MetricChartSchema
	err = apiSchema.FromMetricChartSchema1(apiChart)
	require.NoError(t, err)

	// Test fromAPI
	model := &metricChartConfigModel{}
	diags := model.fromAPI(ctx, apiSchema)
	require.False(t, diags.HasError())

	// Verify metrics were populated
	assert.Len(t, model.Metrics, 1)
	assert.True(t, utils.IsKnown(model.Metrics[0].ConfigJSON))

	// Verify the metric config contains expected data
	var parsedMetric map[string]interface{}
	diags = model.Metrics[0].ConfigJSON.Unmarshal(&parsedMetric)
	require.False(t, diags.HasError())
	assert.Equal(t, "primary", parsedMetric["type"])
	assert.Equal(t, "count", parsedMetric["operation"])

	// Test toAPI round-trip
	resultSchema, diags := model.toAPI()
	require.False(t, diags.HasError())

	resultVariant1, err := resultSchema.AsMetricChartSchema1()
	require.NoError(t, err)
	assert.Len(t, resultVariant1.Metrics, 1)
}

func Test_metricChartPanelConfigConverter_handlesTFPanelConfig(t *testing.T) {
	converter := newMetricChartPanelConfigConverter()

	tests := []struct {
		name     string
		panel    panelModel
		expected bool
	}{
		{
			name: "handles metric chart config",
			panel: panelModel{
				MetricChartConfig: &metricChartConfigModel{},
			},
			expected: true,
		},
		{
			name: "does not handle xy chart config",
			panel: panelModel{
				XYChartConfig: &xyChartConfigModel{},
			},
			expected: false,
		},
		{
			name: "does not handle markdown config",
			panel: panelModel{
				MarkdownConfig: &markdownConfigModel{},
			},
			expected: false,
		},
		{
			name:     "does not handle empty panel",
			panel:    panelModel{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.handlesTFPanelConfig(tt.panel)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_metricChartConfigModel_withDataset(t *testing.T) {
	ctx := context.Background()

	// Create a dataset
	datasetJSON := `{"type": "dataview", "id": "test-dataview"}`
	var dataset kbapi.MetricChartSchema_0_Dataset
	err := json.Unmarshal([]byte(datasetJSON), &dataset)
	require.NoError(t, err)

	apiChart := kbapi.MetricChartSchema0{
		Type:    kbapi.MetricChartSchema0TypeMetric,
		Dataset: dataset,
		Query: kbapi.FilterSimpleSchema{
			Query: "",
		},
		Metrics: []kbapi.MetricChartSchema_0_Metrics_Item{},
	}

	var apiSchema kbapi.MetricChartSchema
	err = apiSchema.FromMetricChartSchema0(apiChart)
	require.NoError(t, err)

	// Test fromAPI
	model := &metricChartConfigModel{}
	diags := model.fromAPI(ctx, apiSchema)
	require.False(t, diags.HasError())

	// Verify dataset was populated
	assert.True(t, utils.IsKnown(model.DatasetJSON))

	var parsedDataset map[string]interface{}
	diags = model.DatasetJSON.Unmarshal(&parsedDataset)
	require.False(t, diags.HasError())
	assert.Equal(t, "dataview", parsedDataset["type"])
	assert.Equal(t, "test-dataview", parsedDataset["id"])

	// Round-trip: toAPI should preserve dataset
	resultSchema, diags := model.toAPI()
	require.False(t, diags.HasError())
	resultVariant0, err := resultSchema.AsMetricChartSchema0()
	require.NoError(t, err)
	resultDatasetJSON, err := json.Marshal(resultVariant0.Dataset)
	require.NoError(t, err)
	var resultDataset map[string]interface{}
	require.NoError(t, json.Unmarshal(resultDatasetJSON, &resultDataset))
	assert.Equal(t, "dataview", resultDataset["type"])
	assert.Equal(t, "test-dataview", resultDataset["id"])
}

func Test_metricChartConfigModel_withFilters(t *testing.T) {
	ctx := context.Background()

	filters := []kbapi.SearchFilterSchema{
		func() kbapi.SearchFilterSchema {
			var filter kbapi.SearchFilterSchema
			_ = filter.FromSearchFilterSchema0(kbapi.SearchFilterSchema0{
				Language: utils.Pointer(kbapi.SearchFilterSchema0Language("kuery")),
				Query: func() kbapi.SearchFilterSchema_0_Query {
					var q kbapi.SearchFilterSchema_0_Query
					_ = q.FromSearchFilterSchema0Query0("status:active")
					return q
				}(),
			})
			return filter
		}(),
	}

	apiChart := kbapi.MetricChartSchema0{
		Type:    kbapi.MetricChartSchema0TypeMetric,
		Filters: &filters,
		Query: kbapi.FilterSimpleSchema{
			Query: "",
		},
		Metrics: []kbapi.MetricChartSchema_0_Metrics_Item{},
	}

	var apiSchema kbapi.MetricChartSchema
	err := apiSchema.FromMetricChartSchema0(apiChart)
	require.NoError(t, err)

	// Test fromAPI
	model := &metricChartConfigModel{}
	diags := model.fromAPI(ctx, apiSchema)
	require.False(t, diags.HasError())

	// Verify filters were populated
	assert.Len(t, model.Filters, 1)
	assert.Equal(t, "status:active", model.Filters[0].Query.ValueString())
	assert.Equal(t, "kuery", model.Filters[0].Language.ValueString())

	// Test toAPI round-trip
	resultSchema, diags := model.toAPI()
	require.False(t, diags.HasError())

	resultVariant0, err := resultSchema.AsMetricChartSchema0()
	require.NoError(t, err)
	require.NotNil(t, resultVariant0.Filters)
	assert.Len(t, *resultVariant0.Filters, 1)
}

func Test_metricChartConfigModel_withBreakdownBy(t *testing.T) {
	ctx := context.Background()

	breakdownByJSON := `{"operation": "terms", "field": "category", "columns": 3}`
	var breakdownBy kbapi.MetricChartSchema_0_BreakdownBy
	err := json.Unmarshal([]byte(breakdownByJSON), &breakdownBy)
	require.NoError(t, err)

	apiChart := kbapi.MetricChartSchema0{
		Type:        kbapi.MetricChartSchema0TypeMetric,
		BreakdownBy: &breakdownBy,
		Query: kbapi.FilterSimpleSchema{
			Language: utils.Pointer(kbapi.FilterSimpleSchemaLanguage("kuery")),
			Query:    "status:active",
		},
		Metrics: []kbapi.MetricChartSchema_0_Metrics_Item{},
	}

	var apiSchema kbapi.MetricChartSchema
	err = apiSchema.FromMetricChartSchema0(apiChart)
	require.NoError(t, err)

	// Test fromAPI
	model := &metricChartConfigModel{}
	diags := model.fromAPI(ctx, apiSchema)
	require.False(t, diags.HasError())

	// Verify breakdown_by was populated
	assert.True(t, utils.IsKnown(model.BreakdownByJSON))

	var parsedBreakdown map[string]interface{}
	diags = model.BreakdownByJSON.Unmarshal(&parsedBreakdown)
	require.False(t, diags.HasError())
	assert.Equal(t, "terms", parsedBreakdown["operation"])
	assert.Equal(t, "category", parsedBreakdown["field"])

	// Test toAPI round-trip
	resultSchema, diags := model.toAPI()
	require.False(t, diags.HasError())

	resultVariant0, err := resultSchema.AsMetricChartSchema0()
	require.NoError(t, err)
	assert.NotNil(t, resultVariant0.BreakdownBy)
}

func Test_metricItemModel_jsonRoundTrip(t *testing.T) {
	// Test that we can round-trip complex metric configurations
	metricConfigs := []string{
		`{"type": "primary", "operation": "count", "format": {"id": "number"}, "alignments": {"value": "center"}, "icon": {"name": "empty"}}`,
		`{"type": "secondary", "operation": "average", "column": "price", "format": {"id": "currency"}, "prefix": "$"}`,
		`{"type": "primary", "operation": "value", "column": "total", "format": {"id": "number"}, "alignments": {"value": "left"}, "icon": {"name": "star"}, "fit": true}`,
	}

	for i, configJSON := range metricConfigs {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			// Create a metric item with the config
			item := metricItemModel{
				ConfigJSON: customtypes.NewJSONWithDefaultsValue[map[string]any](
					configJSON,
					populateMetricChartMetricDefaults,
				),
			}

			// Unmarshal and re-marshal to verify it's valid
			var parsed map[string]interface{}
			diags := item.ConfigJSON.Unmarshal(&parsed)
			require.False(t, diags.HasError())

			// Verify we can marshal it back
			remarshaled, err := json.Marshal(parsed)
			require.NoError(t, err)
			assert.NotEmpty(t, remarshaled)
		})
	}
}
