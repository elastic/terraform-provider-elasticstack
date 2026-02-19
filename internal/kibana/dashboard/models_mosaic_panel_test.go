package dashboard

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newMosaicPanelConfigConverter(t *testing.T) {
	converter := newMosaicPanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, "mosaic", converter.visualizationType)
}

func Test_mosaicConfigModel_fromAPI_toAPI_ESQL(t *testing.T) {
	var color kbapi.ColorMapping
	require.NoError(t, json.Unmarshal([]byte(`{"mode":"gradient","palette":"default","unassignedColor":"#000000"}`), &color))

	metricFormat := kbapi.FormatTypeSchema{}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number","decimals":2}`), &metricFormat))

	api := kbapi.MosaicESQL{
		Type: kbapi.MosaicESQLTypeMosaic,
		Legend: kbapi.MosaicLegend{
			Size: kbapi.LegendSizeAuto,
		},
		Title:               utils.Pointer("Mosaic ES|QL"),
		Description:         utils.Pointer("ES|QL description"),
		IgnoreGlobalFilters: utils.Pointer(true),
		Sampling:            utils.Pointer(float32(0.75)),
		Metrics: []struct {
			Column    string                           `json:"column"`
			Format    kbapi.FormatTypeSchema           `json:"format"`
			Label     *string                          `json:"label,omitempty"`
			Operation kbapi.MosaicESQLMetricsOperation `json:"operation"`
		}{
			{
				Column:    "count",
				Format:    metricFormat,
				Label:     utils.Pointer("Count"),
				Operation: kbapi.MosaicESQLMetricsOperationValue,
			},
		},
		ValueDisplay: &struct {
			Mode            kbapi.MosaicESQLValueDisplayMode `json:"mode"`
			PercentDecimals *float32                         `json:"percent_decimals,omitempty"`
		}{
			Mode:            kbapi.MosaicESQLValueDisplayModePercentage,
			PercentDecimals: utils.Pointer(float32(2)),
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM metrics-* | STATS count() as count BY host.name"}`), &api.Dataset))

	groupBy := []struct {
		CollapseBy kbapi.CollapseBy                 `json:"collapse_by"`
		Color      kbapi.ColorMapping               `json:"color"`
		Column     string                           `json:"column"`
		Operation  kbapi.MosaicESQLGroupByOperation `json:"operation"`
	}{
		{
			CollapseBy: kbapi.CollapseByAvg,
			Color:      color,
			Column:     "host.name",
			Operation:  kbapi.MosaicESQLGroupByOperationValue,
		},
	}
	api.GroupBy = &groupBy

	groupBreakdownBy := []struct {
		CollapseBy kbapi.CollapseBy                          `json:"collapse_by"`
		Color      kbapi.ColorMapping                        `json:"color"`
		Column     string                                    `json:"column"`
		Operation  kbapi.MosaicESQLGroupBreakdownByOperation `json:"operation"`
	}{
		{
			CollapseBy: kbapi.CollapseByAvg,
			Color:      color,
			Column:     "service.name",
			Operation:  kbapi.MosaicESQLGroupBreakdownByOperationValue,
		},
	}
	api.GroupBreakdownBy = &groupBreakdownBy

	model := &mosaicConfigModel{}
	diags := model.fromAPIESQL(context.Background(), api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Mosaic ES|QL"), model.Title)
	assert.Equal(t, types.StringValue("ES|QL description"), model.Description)
	assert.Equal(t, types.BoolValue(true), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(0.75), model.Sampling)
	require.NotNil(t, model.Legend)
	assert.Equal(t, types.StringValue("auto"), model.Legend.Size)
	require.NotNil(t, model.ValueDisplay)
	assert.Equal(t, types.StringValue("percentage"), model.ValueDisplay.Mode)
	assert.Equal(t, types.Float64Value(2), model.ValueDisplay.PercentDecimals)

	require.NotNil(t, model.Esql)
	require.False(t, model.Esql.Dataset.IsNull())
	assert.Len(t, model.Esql.GroupBy, 1)
	assert.Len(t, model.Esql.GroupBreakdownBy, 1)
	assert.Len(t, model.Esql.Metrics, 1)

	mosaicChart, diags := model.toAPI()
	require.False(t, diags.HasError())

	roundTrip, err := mosaicChart.AsMosaicESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.MosaicESQLTypeMosaic, roundTrip.Type)
	assert.NotNil(t, roundTrip.Title)
	assert.NotNil(t, roundTrip.Description)
	assert.NotNil(t, roundTrip.IgnoreGlobalFilters)
	assert.NotNil(t, roundTrip.Sampling)
	assert.Len(t, roundTrip.Metrics, 1)
}

func Test_mosaicConfigModel_fromAPI_toAPI_NoESQL(t *testing.T) {
	api := kbapi.MosaicNoESQL{
		Type: kbapi.MosaicNoESQLTypeMosaic,
		Legend: kbapi.MosaicLegend{
			Size: kbapi.LegendSizeAuto,
		},
		Title:               utils.Pointer("Mosaic Standard"),
		Description:         utils.Pointer("Standard description"),
		IgnoreGlobalFilters: utils.Pointer(false),
		Sampling:            utils.Pointer(float32(1)),
		Query:               kbapi.FilterSimpleSchema{},
		Metrics:             []kbapi.MosaicNoESQL_Metrics_Item{},
		ValueDisplay: &struct {
			Mode            kbapi.MosaicNoESQLValueDisplayMode `json:"mode"`
			PercentDecimals *float32                           `json:"percent_decimals,omitempty"`
		}{
			Mode:            kbapi.MosaicNoESQLValueDisplayModeAbsolute,
			PercentDecimals: utils.Pointer(float32(0)),
		},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.Dataset))
	require.NoError(t, json.Unmarshal([]byte(`{"language":"kuery","query":"*"}`), &api.Query))

	metric := kbapi.MosaicNoESQL_Metrics_Item{}
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metric))
	api.Metrics = append(api.Metrics, metric)

	groupByItem := kbapi.MosaicNoESQL_GroupBy_Item{}
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"terms","field":"host.name"}`), &groupByItem))
	groupBy := []kbapi.MosaicNoESQL_GroupBy_Item{groupByItem}
	api.GroupBy = &groupBy

	groupBreakdownItem := kbapi.MosaicNoESQL_GroupBreakdownBy_Item{}
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"terms","field":"service.name"}`), &groupBreakdownItem))
	groupBreakdown := []kbapi.MosaicNoESQL_GroupBreakdownBy_Item{groupBreakdownItem}
	api.GroupBreakdownBy = &groupBreakdown

	model := &mosaicConfigModel{}
	diags := model.fromAPINoESQL(context.Background(), api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Mosaic Standard"), model.Title)
	assert.Equal(t, types.StringValue("Standard description"), model.Description)
	assert.Equal(t, types.BoolValue(false), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(1), model.Sampling)
	require.NotNil(t, model.Legend)
	assert.Equal(t, types.StringValue("auto"), model.Legend.Size)
	require.NotNil(t, model.ValueDisplay)
	assert.Equal(t, types.StringValue("absolute"), model.ValueDisplay.Mode)

	require.NotNil(t, model.Standard)
	require.NotNil(t, model.Standard.Query)
	assert.Equal(t, types.StringValue("*"), model.Standard.Query.Query)
	assert.Len(t, model.Standard.GroupBy, 1)
	assert.Len(t, model.Standard.GroupBreakdownBy, 1)
	assert.Len(t, model.Standard.Metrics, 1)

	mosaicChart, diags := model.toAPI()
	require.False(t, diags.HasError())

	roundTrip, err := mosaicChart.AsMosaicNoESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.MosaicNoESQLTypeMosaic, roundTrip.Type)
	assert.NotNil(t, roundTrip.Title)
	assert.NotNil(t, roundTrip.Description)
	assert.NotNil(t, roundTrip.Sampling)
	assert.Len(t, roundTrip.Metrics, 1)
}
