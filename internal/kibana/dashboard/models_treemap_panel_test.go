package dashboard

import (
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
