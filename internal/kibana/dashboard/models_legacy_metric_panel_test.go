package dashboard

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_legacyMetricConfigModel_fromAPI_toAPI_NoESQL(t *testing.T) {
	apiJSON := `{
		"type": "legacy_metric",
		"title": "Legacy Metric",
		"description": "Legacy metric description",
		"dataset": {"type": "dataView", "id": "metrics-*"},
		"query": {"language": "kuery", "query": ""},
		"sampling": 0.5,
		"ignore_global_filters": true,
		"filters": [{"query": "status:200", "language": "kuery"}],
		"metric": {"operation": "count", "format": {"type": "number"}}
	}`

	var apiNoESQL kbapi.LegacyMetricNoESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &apiNoESQL))

	model := &legacyMetricConfigModel{}
	diags := model.fromAPINoESQL(t.Context(), apiNoESQL)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Legacy Metric"), model.Title)
	assert.Equal(t, types.StringValue("Legacy metric description"), model.Description)
	expectedDataset := jsontypes.NewNormalizedValue(`{"type":"dataView","id":"metrics-*"}`)
	semanticEqual, diags := model.Dataset.StringSemanticEquals(t.Context(), expectedDataset)
	require.False(t, diags.HasError())
	require.True(t, semanticEqual)
	assert.Equal(t, types.BoolValue(true), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(0.5), model.Sampling)
	require.NotNil(t, model.Query)
	assert.Equal(t, types.StringValue("kuery"), model.Query.Language)
	assert.Equal(t, types.StringValue(""), model.Query.Query)
	require.Len(t, model.Filters, 1)
	assert.Equal(t, types.StringValue("status:200"), model.Filters[0].Query)
	expectedMetric := customtypes.NewJSONWithDefaultsValue[map[string]any](
		`{"format":{"type":"number"},"operation":"count"}`,
		populateLegacyMetricMetricDefaults,
	)
	metricEqual, metricDiags := model.Metric.StringSemanticEquals(t.Context(), expectedMetric)
	require.False(t, metricDiags.HasError())
	require.True(t, metricEqual)

	legacyMetricChart, toDiags := model.toAPI()
	require.False(t, toDiags.HasError())

	apiRoundTrip, err := legacyMetricChart.AsLegacyMetricNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.LegacyMetricNoESQLTypeLegacyMetric, apiRoundTrip.Type)
	assert.Equal(t, "Legacy Metric", *apiRoundTrip.Title)
	assert.Equal(t, "Legacy metric description", *apiRoundTrip.Description)
	assert.NotNil(t, apiRoundTrip.Query)
	assert.NotNil(t, apiRoundTrip.Filters)
}

func Test_legacyMetricConfigModel_fromAPI_toAPI_ESQL(t *testing.T) {
	apiJSON := `{
		"type": "legacy_metric",
		"title": "Legacy Metric ESQL",
		"description": "Legacy metric esql description",
		"dataset": {"type": "esql", "query": "FROM metrics-* | LIMIT 10"},
		"sampling": 1,
		"ignore_global_filters": false,
		"filters": [{"query": "service.name:api", "language": "kuery"}],
		"metric": {
			"format": {"type": "number"},
			"label": "CPU",
			"operation": "value",
			"column": "cpu",
			"size": "m",
			"alignments": {"labels": "top", "value": "center"},
			"apply_color_to": "value",
			"color": {
				"type": "dynamic",
				"range": "absolute",
				"steps": [{"type": "from", "from": 0, "color": "#00ff00"}]
			}
		}
	}`

	var apiESQL kbapi.LegacyMetricESQL
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &apiESQL))

	model := &legacyMetricConfigModel{}
	diags := model.fromAPIESQL(t.Context(), apiESQL)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Legacy Metric ESQL"), model.Title)
	assert.Equal(t, types.StringValue("Legacy metric esql description"), model.Description)
	expectedESQLDataset := jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM metrics-* | LIMIT 10"}`)
	esqlDatasetEqual, datasetDiags := model.Dataset.StringSemanticEquals(t.Context(), expectedESQLDataset)
	require.False(t, datasetDiags.HasError())
	require.True(t, esqlDatasetEqual)
	assert.Equal(t, types.BoolValue(false), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(1), model.Sampling)
	assert.Nil(t, model.Query)
	require.Len(t, model.Filters, 1)
	assert.Equal(t, types.StringValue("service.name:api"), model.Filters[0].Query)
	expectedESQLMetric := customtypes.NewJSONWithDefaultsValue[map[string]any](
		`{"alignments":{"labels":"top","value":"center"},"apply_color_to":"value","color":{"range":"absolute","steps":[{"color":"#00ff00","from":0,"type":"from"}],"type":"dynamic"},"column":"cpu","format":{"type":"number"},"label":"CPU","operation":"value","size":"m"}`,
		populateLegacyMetricMetricDefaults,
	)
	esqlMetricEqual, esqlMetricDiags := model.Metric.StringSemanticEquals(t.Context(), expectedESQLMetric)
	require.False(t, esqlMetricDiags.HasError())
	require.True(t, esqlMetricEqual)

	legacyMetricChart, toDiags := model.toAPI()
	require.False(t, toDiags.HasError())

	apiRoundTrip, err := legacyMetricChart.AsLegacyMetricESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.LegacyMetricESQLTypeLegacyMetric, apiRoundTrip.Type)
	assert.Equal(t, "Legacy Metric ESQL", *apiRoundTrip.Title)
	assert.Equal(t, "Legacy metric esql description", *apiRoundTrip.Description)
	assert.Equal(t, "cpu", apiRoundTrip.Metric.Column)
	assert.Equal(t, kbapi.LegacyMetricESQLMetricOperationValue, apiRoundTrip.Metric.Operation)
}

func Test_legacyMetricConfigModel_toAPI_requiresQueryForNoESQL(t *testing.T) {
	model := &legacyMetricConfigModel{
		Title:   types.StringValue("Missing Query"),
		Dataset: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"metrics-*"}`),
		Metric: customtypes.NewJSONWithDefaultsValue[map[string]any](
			`{"operation":"count","format":{"type":"number"}}`,
			populateLegacyMetricMetricDefaults,
		),
	}

	_, diags := model.toAPI()
	require.True(t, diags.HasError())
}
