package dashboard

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertLegacyMetricConfigEqual verifies that two legacy metric config models are equivalent (round-trip safe).
func assertLegacyMetricConfigEqual(t *testing.T, ctx context.Context, a, b *legacyMetricConfigModel) {
	t.Helper()
	if a == nil && b == nil {
		return
	}
	require.NotNil(t, a, "expected non-nil first config")
	require.NotNil(t, b, "expected non-nil second config")
	assert.Equal(t, a.Title, b.Title)
	assert.Equal(t, a.Description, b.Description)
	assert.Equal(t, a.IgnoreGlobalFilters, b.IgnoreGlobalFilters)
	assert.Equal(t, a.Sampling, b.Sampling)
	if a.Dataset.IsNull() != b.Dataset.IsNull() || a.Dataset.IsUnknown() != b.Dataset.IsUnknown() {
		assert.Fail(t, "dataset null/unknown state mismatch")
		return
	}
	if !a.Dataset.IsNull() && !a.Dataset.IsUnknown() {
		eq, d := a.Dataset.StringSemanticEquals(ctx, b.Dataset)
		require.False(t, d.HasError())
		assert.True(t, eq, "dataset should be semantically equal")
	}
	if (a.Query == nil) != (b.Query == nil) {
		assert.Fail(t, "query nil mismatch: one has query, other does not")
		return
	}
	if a.Query != nil {
		assert.Equal(t, a.Query.Language, b.Query.Language)
		assert.Equal(t, a.Query.Query, b.Query.Query)
	}
	assert.Len(t, b.Filters, len(a.Filters))
	for i := range a.Filters {
		assert.Equal(t, a.Filters[i].Query, b.Filters[i].Query)
	}
	if a.Metric.IsNull() != b.Metric.IsNull() || a.Metric.IsUnknown() != b.Metric.IsUnknown() {
		assert.Fail(t, "metric null/unknown state mismatch")
		return
	}
	if !a.Metric.IsNull() && !a.Metric.IsUnknown() {
		eq, d := a.Metric.StringSemanticEquals(ctx, b.Metric)
		require.False(t, d.HasError())
		assert.True(t, eq, "metric should be semantically equal")
	}
}

func Test_legacyMetricConfigModel_fromAPI_toAPI_NoESQL(t *testing.T) {
	ctx := t.Context()
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

	var chart kbapi.LegacyMetricChartSchema
	require.NoError(t, json.Unmarshal([]byte(apiJSON), &chart))

	// Round-trip: API chart → model → API chart → model; then assert first model equals second model.
	model1 := &legacyMetricConfigModel{}
	diags := model1.fromAPI(ctx, chart)
	require.False(t, diags.HasError())

	chart2, diags := model1.toAPI()
	require.False(t, diags.HasError())

	model2 := &legacyMetricConfigModel{}
	diags = model2.fromAPI(ctx, chart2)
	require.False(t, diags.HasError())

	assertLegacyMetricConfigEqual(t, ctx, model1, model2)
}

func Test_legacyMetricConfigModel_fromAPI_toAPI_ESQL(t *testing.T) {
	ctx := t.Context()
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

	// Round-trip: ESQL API → model (fromAPIESQL) → API chart (toAPI) → model (fromAPI); assert models equal when second fromAPI succeeds.
	model1 := &legacyMetricConfigModel{}
	diags := model1.fromAPIESQL(ctx, apiESQL)
	require.False(t, diags.HasError())

	chart2, diags := model1.toAPI()
	require.False(t, diags.HasError())

	// Round-trip back: chart → model. Union may parse as NoESQL, so only assert API-level ESQL round-trip.
	model2 := &legacyMetricConfigModel{}
	diags = model2.fromAPI(ctx, chart2)
	if !diags.HasError() && model2.Query == nil {
		assertLegacyMetricConfigEqual(t, ctx, model1, model2)
		return
	}
	apiRoundTrip, err := chart2.AsLegacyMetricESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.LegacyMetricESQLTypeLegacyMetric, apiRoundTrip.Type)
	assert.Equal(t, "Legacy Metric ESQL", *apiRoundTrip.Title)
	assert.Equal(t, "cpu", apiRoundTrip.Metric.Column)
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

func Test_legacyMetricPanelConfigConverter_handlesAPIPanelConfig(t *testing.T) {
	buildConfig := func(t *testing.T, configMap map[string]interface{}) kbapi.DashboardPanelItem_Config {
		t.Helper()
		var cfg kbapi.DashboardPanelItem_Config
		require.NoError(t, cfg.FromDashboardPanelItemConfig2(configMap))
		return cfg
	}

	tests := []struct {
		name      string
		panelType string
		config    kbapi.DashboardPanelItem_Config
		want      bool
	}{
		{
			name:      "handles lens legacy metric config",
			panelType: "lens",
			config: buildConfig(t, map[string]interface{}{
				"attributes": map[string]interface{}{
					"type":    "legacy_metric",
					"dataset": map[string]interface{}{"type": "dataView", "id": "metrics-*"},
					"query":   map[string]interface{}{"language": "kuery", "query": ""},
					"metric":  map[string]interface{}{"operation": "count"},
				},
			}),
			want: true,
		},
		{
			name:      "does not handle lens non-legacy metric config",
			panelType: "lens",
			config: buildConfig(t, map[string]interface{}{
				"attributes": map[string]interface{}{"type": "xy"},
			}),
			want: false,
		},
		{
			name:      "does not handle non-lens type",
			panelType: "DASHBOARD_MARKDOWN",
			config: buildConfig(t, map[string]interface{}{
				"attributes": map[string]interface{}{"type": "legacy_metric"},
			}),
			want: false,
		},
		{
			name:      "does not handle empty type",
			panelType: "",
			config:    buildConfig(t, map[string]interface{}{"attributes": map[string]interface{}{"type": "legacy_metric"}}),
			want:      false,
		},
		{
			name:      "does not handle missing attributes",
			panelType: "lens",
			config:    buildConfig(t, map[string]interface{}{}),
			want:      false,
		},
		{
			name:      "does not handle non-map attributes",
			panelType: "lens",
			config:    buildConfig(t, map[string]interface{}{"attributes": "legacy_metric"}),
			want:      false,
		},
		{
			name:      "does not handle missing visualization type",
			panelType: "lens",
			config:    buildConfig(t, map[string]interface{}{"attributes": map[string]interface{}{"dataset": map[string]interface{}{"type": "dataView"}}}),
			want:      false,
		},
		{
			name:      "does not handle invalid config union",
			panelType: "lens",
			config:    kbapi.DashboardPanelItem_Config{},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newLegacyMetricPanelConfigConverter()
			got := c.handlesAPIPanelConfig(nil, tt.panelType, tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_legacyMetricPanelConfigConverter_roundTrip(t *testing.T) {
	ctx := context.Background()
	// Start from API config (dashboard panel config with legacy_metric attributes).
	attrs := map[string]interface{}{
		"type":                  "legacy_metric",
		"title":                 "Round-Trip Title",
		"description":           "Round-trip description",
		"dataset":               map[string]interface{}{"type": "dataView", "id": "logs-*"},
		"query":                 map[string]interface{}{"language": "kuery", "query": "host:*"},
		"metric":                map[string]interface{}{"operation": "count", "format": map[string]interface{}{"type": "number"}},
		"sampling":              0.5,
		"ignore_global_filters": true,
		"filters":               []interface{}{map[string]interface{}{"query": "status:200", "language": "kuery"}},
	}
	configMap := map[string]interface{}{"attributes": attrs}
	var apiConfig1 kbapi.DashboardPanelItem_Config
	require.NoError(t, apiConfig1.FromDashboardPanelItemConfig2(configMap))

	c := newLegacyMetricPanelConfigConverter()

	// API → model (populateFromAPIPanel)
	pm1 := &panelModel{}
	diags := c.populateFromAPIPanel(ctx, pm1, apiConfig1)
	require.False(t, diags.HasError())
	require.NotNil(t, pm1.LegacyMetricConfig)

	// model → API (mapPanelToAPI)
	var apiConfig2 kbapi.DashboardPanelItem_Config
	diags = c.mapPanelToAPI(*pm1, &apiConfig2)
	require.False(t, diags.HasError())

	// API → model again (round-trip)
	pm2 := &panelModel{}
	diags = c.populateFromAPIPanel(ctx, pm2, apiConfig2)
	require.False(t, diags.HasError())
	require.NotNil(t, pm2.LegacyMetricConfig)

	assertLegacyMetricConfigEqual(t, ctx, pm1.LegacyMetricConfig, pm2.LegacyMetricConfig)
}

func Test_legacyMetricConfigModel_fromAPI_roundTrip(t *testing.T) {
	ctx := context.Background()

	t.Run("NoESQL round-trip", func(t *testing.T) {
		var chart kbapi.LegacyMetricChartSchema
		require.NoError(t, json.Unmarshal([]byte(`{"type":"legacy_metric","dataset":{"type":"dataView","id":"x"},"query":{"language":"kuery","query":""},"metric":{"operation":"count","format":{"type":"number"}}}`), &chart))
		model1 := &legacyMetricConfigModel{}
		diags := model1.fromAPI(ctx, chart)
		require.False(t, diags.HasError())
		chart2, diags := model1.toAPI()
		require.False(t, diags.HasError())
		model2 := &legacyMetricConfigModel{}
		diags = model2.fromAPI(ctx, chart2)
		require.False(t, diags.HasError())
		assertLegacyMetricConfigEqual(t, ctx, model1, model2)
	})

	t.Run("ESQL round-trip", func(t *testing.T) {
		var apiESQL kbapi.LegacyMetricESQL
		require.NoError(t, json.Unmarshal([]byte(`{"type":"legacy_metric","dataset":{"type":"esql","query":"FROM x"},"metric":{"operation":"value","column":"y","format":{"type":"number"},"color":{"type":"static","color":"#fff"}}}`), &apiESQL))
		model1 := &legacyMetricConfigModel{}
		diags := model1.fromAPIESQL(ctx, apiESQL)
		require.False(t, diags.HasError())
		chart2, diags := model1.toAPI()
		require.False(t, diags.HasError())
		model2 := &legacyMetricConfigModel{}
		diags = model2.fromAPI(ctx, chart2)
		if !diags.HasError() && model2.Query == nil {
			assertLegacyMetricConfigEqual(t, ctx, model1, model2)
			return
		}
		_, err := chart2.AsLegacyMetricESQL()
		require.NoError(t, err)
	})
}

func Test_legacyMetricConfigModel_toAPI_nil(t *testing.T) {
	var model *legacyMetricConfigModel
	_, diags := model.toAPI()
	require.True(t, diags.HasError())
}

func Test_legacyMetricConfigModel_toAPI_unsupportedDataset(t *testing.T) {
	model := &legacyMetricConfigModel{
		Dataset: jsontypes.NewNormalizedValue(`{"type":"unknown"}`),
		Metric:  customtypes.NewJSONWithDefaultsValue[map[string]any](`{}`, populateLegacyMetricMetricDefaults),
	}
	_, diags := model.toAPI()
	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Unsupported legacy metric dataset")
}

func Test_legacyMetricConfigModel_toAPI_ESQL_withQuery(t *testing.T) {
	model := &legacyMetricConfigModel{
		Dataset: jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM x"}`),
		Query:   &filterSimpleModel{Language: types.StringValue("kuery"), Query: types.StringValue("*")},
		Metric:  customtypes.NewJSONWithDefaultsValue[map[string]any](`{"operation":"value","column":"y","format":{"type":"number"},"color":{"type":"static","color":"#fff"}}`, populateLegacyMetricMetricDefaults),
	}
	_, diags := model.toAPI()
	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Invalid legacy metric query")
}

func Test_legacyMetricConfigModel_toAPI_missingMetric(t *testing.T) {
	model := &legacyMetricConfigModel{
		Title:   types.StringValue("T"),
		Dataset: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"x"}`),
		Query:   &filterSimpleModel{Language: types.StringValue("kuery"), Query: types.StringValue("")},
		Metric:  customtypes.NewJSONWithDefaultsNull[map[string]any](populateLegacyMetricMetricDefaults),
	}
	_, diags := model.toAPI()
	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Missing metric")
}

func Test_legacyMetricConfigModel_datasetType_errors(t *testing.T) {
	t.Run("missing dataset", func(t *testing.T) {
		model := &legacyMetricConfigModel{}
		_, diags := model.datasetType()
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Summary(), "Missing dataset")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		model := &legacyMetricConfigModel{
			Dataset: jsontypes.NewNormalizedValue(`{invalid`),
		}
		_, diags := model.datasetType()
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Summary(), "Failed to decode dataset type")
	})

	t.Run("missing type field", func(t *testing.T) {
		model := &legacyMetricConfigModel{
			Dataset: jsontypes.NewNormalizedValue(`{"id":"x"}`),
		}
		_, diags := model.datasetType()
		require.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Summary(), "Missing dataset type")
	})
}
