package dashboard

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newDatatablePanelConfigConverter(t *testing.T) {
	converter := newDatatablePanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, "datatable", converter.visualizationType)
}

func Test_datatableDensityModel_fromAPI_toAPI(t *testing.T) {
	header := kbapi.DatatableDensity_Height_Header{}
	require.NoError(t, header.FromDatatableDensityHeightHeader1(kbapi.DatatableDensityHeightHeader1{
		Type:     kbapi.DatatableDensityHeightHeader1TypeCustom,
		MaxLines: utils.Pointer(float32(2)),
	}))

	value := kbapi.DatatableDensity_Height_Value{}
	require.NoError(t, value.FromDatatableDensityHeightValue1(kbapi.DatatableDensityHeightValue1{
		Type:  kbapi.DatatableDensityHeightValue1TypeCustom,
		Lines: utils.Pointer(float32(3)),
	}))

	api := kbapi.DatatableDensity{
		Mode: utils.Pointer(kbapi.DatatableDensityModeCompact),
		Height: &struct {
			Header *kbapi.DatatableDensity_Height_Header `json:"header,omitempty"`
			Value  *kbapi.DatatableDensity_Height_Value  `json:"value,omitempty"`
		}{
			Header: &header,
			Value:  &value,
		},
	}

	model := &datatableDensityModel{}
	diags := model.fromAPI(api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("compact"), model.Mode)
	require.NotNil(t, model.Height)
	require.NotNil(t, model.Height.Header)
	assert.Equal(t, types.StringValue("custom"), model.Height.Header.Type)
	assert.Equal(t, types.Float64Value(2), model.Height.Header.MaxLines)
	require.NotNil(t, model.Height.Value)
	assert.Equal(t, types.StringValue("custom"), model.Height.Value.Type)
	assert.Equal(t, types.Float64Value(3), model.Height.Value.Lines)

	roundTrip, diags := model.toAPI()
	require.False(t, diags.HasError())
	assert.NotNil(t, roundTrip.Height)
	assert.NotNil(t, roundTrip.Mode)
}

func Test_datatableNoESQLConfigModel_fromAPI_toAPI(t *testing.T) {
	header := kbapi.DatatableDensity_Height_Header{}
	require.NoError(t, header.FromDatatableDensityHeightHeader0(kbapi.DatatableDensityHeightHeader0{
		Type: kbapi.DatatableDensityHeightHeader0TypeAuto,
	}))

	value := kbapi.DatatableDensity_Height_Value{}
	require.NoError(t, value.FromDatatableDensityHeightValue0(kbapi.DatatableDensityHeightValue0{
		Type: kbapi.DatatableDensityHeightValue0TypeAuto,
	}))

	density := kbapi.DatatableDensity{
		Mode: utils.Pointer(kbapi.DatatableDensityModeDefault),
		Height: &struct {
			Header *kbapi.DatatableDensity_Height_Header `json:"header,omitempty"`
			Value  *kbapi.DatatableDensity_Height_Value  `json:"value,omitempty"`
		}{
			Header: &header,
			Value:  &value,
		},
	}

	api := kbapi.DatatableNoESQL{
		Type:                kbapi.DatatableNoESQLTypeDatatable,
		Title:               utils.Pointer("Datatable NoESQL"),
		Description:         utils.Pointer("NoESQL description"),
		IgnoreGlobalFilters: utils.Pointer(true),
		Sampling:            utils.Pointer(float32(0.5)),
		Density:             density,
		Query:               kbapi.FilterSimpleSchema{},
		Metrics:             []kbapi.DatatableNoESQL_Metrics_Item{},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.Dataset))
	require.NoError(t, json.Unmarshal([]byte(`{"language":"kuery","query":"*"}`), &api.Query))

	metric := kbapi.DatatableNoESQL_Metrics_Item{}
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &metric))
	api.Metrics = append(api.Metrics, metric)

	row := kbapi.DatatableNoESQL_Rows_Item{}
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"terms","field":"host.name"}`), &row))
	rows := []kbapi.DatatableNoESQL_Rows_Item{row}
	api.Rows = &rows

	split := kbapi.DatatableNoESQL_SplitMetricsBy_Item{}
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"terms","field":"host.name"}`), &split))
	splits := []kbapi.DatatableNoESQL_SplitMetricsBy_Item{split}
	api.SplitMetricsBy = &splits

	sortBy := kbapi.DatatableNoESQL_SortBy{}
	require.NoError(t, json.Unmarshal([]byte(`{"column_type":"metric","direction":"asc","index":0}`), &sortBy))
	api.SortBy = &sortBy

	paging := kbapi.DatatableNoESQLPaging(10)
	api.Paging = &paging

	model := &datatableNoESQLConfigModel{}
	diags := model.fromAPI(context.Background(), api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Datatable NoESQL"), model.Title)
	assert.Equal(t, types.StringValue("NoESQL description"), model.Description)
	assert.False(t, model.DatasetJSON.IsNull())
	assert.Equal(t, types.BoolValue(true), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(0.5), model.Sampling)
	require.NotNil(t, model.Query)
	assert.Equal(t, types.StringValue("*"), model.Query.Query)
	assert.Len(t, model.Metrics, 1)
	assert.Len(t, model.Rows, 1)
	assert.Len(t, model.SplitMetricsBy, 1)
	assert.Equal(t, types.Int64Value(10), model.Paging)

	apiRoundTrip, diags := model.toAPI()
	require.False(t, diags.HasError())
	assert.Equal(t, kbapi.DatatableNoESQLTypeDatatable, apiRoundTrip.Type)
	assert.NotNil(t, apiRoundTrip.Paging)
}

func Test_datatableESQLConfigModel_fromAPI_toAPI(t *testing.T) {
	density := kbapi.DatatableDensity{
		Mode: utils.Pointer(kbapi.DatatableDensityModeExpanded),
	}

	metric := kbapi.DatatableESQLMetric{
		Column:    "system.cpu.user.pct",
		Operation: kbapi.DatatableESQLMetricOperationValue,
		Format:    kbapi.FormatTypeSchema{},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number","decimals":2}`), &metric.Format))

	row := struct {
		Alignment    *kbapi.DatatableESQLRowsAlignment    `json:"alignment,omitempty"`
		ApplyColorTo *kbapi.DatatableESQLRowsApplyColorTo `json:"apply_color_to,omitempty"`
		ClickFilter  *bool                                `json:"click_filter,omitempty"`
		CollapseBy   kbapi.CollapseBy                     `json:"collapse_by"`
		Color        *kbapi.DatatableESQL_Rows_Color      `json:"color,omitempty"`
		Column       string                               `json:"column"`
		Operation    kbapi.DatatableESQLRowsOperation     `json:"operation"`
		Visible      *bool                                `json:"visible,omitempty"`
		Width        *float32                             `json:"width,omitempty"`
	}{
		Column:     "host.name",
		Operation:  kbapi.DatatableESQLRowsOperationValue,
		CollapseBy: kbapi.CollapseByAvg,
	}

	split := struct {
		Column    string                                     `json:"column"`
		Operation kbapi.DatatableESQLSplitMetricsByOperation `json:"operation"`
	}{
		Column:    "host.name",
		Operation: kbapi.DatatableESQLSplitMetricsByOperationValue,
	}

	api := kbapi.DatatableESQL{
		Type:                kbapi.DatatableESQLTypeDatatable,
		Title:               utils.Pointer("Datatable ESQL"),
		Description:         utils.Pointer("ESQL description"),
		IgnoreGlobalFilters: utils.Pointer(false),
		Sampling:            utils.Pointer(float32(1)),
		Density:             density,
		Metrics:             []kbapi.DatatableESQLMetric{metric},
		Rows: &[]struct {
			Alignment    *kbapi.DatatableESQLRowsAlignment    `json:"alignment,omitempty"`
			ApplyColorTo *kbapi.DatatableESQLRowsApplyColorTo `json:"apply_color_to,omitempty"`
			ClickFilter  *bool                                `json:"click_filter,omitempty"`
			CollapseBy   kbapi.CollapseBy                     `json:"collapse_by"`
			Color        *kbapi.DatatableESQL_Rows_Color      `json:"color,omitempty"`
			Column       string                               `json:"column"`
			Operation    kbapi.DatatableESQLRowsOperation     `json:"operation"`
			Visible      *bool                                `json:"visible,omitempty"`
			Width        *float32                             `json:"width,omitempty"`
		}{row},
		SplitMetricsBy: &[]struct {
			Column    string                                     `json:"column"`
			Operation kbapi.DatatableESQLSplitMetricsByOperation `json:"operation"`
		}{split},
	}

	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM metrics-* | KEEP host.name, system.cpu.user.pct | LIMIT 10"}`), &api.Dataset))

	sortBy := kbapi.DatatableESQL_SortBy{}
	require.NoError(t, json.Unmarshal([]byte(`{"column_type":"metric","direction":"desc","index":0}`), &sortBy))
	api.SortBy = &sortBy

	paging := kbapi.DatatableESQLPaging(20)
	api.Paging = &paging

	model := &datatableESQLConfigModel{}
	diags := model.fromAPI(context.Background(), api)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("Datatable ESQL"), model.Title)
	assert.Equal(t, types.StringValue("ESQL description"), model.Description)
	assert.False(t, model.DatasetJSON.IsNull())
	assert.Equal(t, types.BoolValue(false), model.IgnoreGlobalFilters)
	assert.Equal(t, types.Float64Value(1), model.Sampling)
	assert.Len(t, model.Metrics, 1)
	assert.Len(t, model.Rows, 1)
	assert.Len(t, model.SplitMetricsBy, 1)
	assert.Equal(t, types.Int64Value(20), model.Paging)

	apiRoundTrip, diags := model.toAPI()
	require.False(t, diags.HasError())
	assert.Equal(t, kbapi.DatatableESQLTypeDatatable, apiRoundTrip.Type)
	assert.NotNil(t, apiRoundTrip.Paging)
	assert.NotNil(t, apiRoundTrip.Rows)
}

func Test_datatablePanelConfigConverter_roundTrip(t *testing.T) {
	converter := newDatatablePanelConfigConverter()
	configModel := &datatableNoESQLConfigModel{
		Title:       types.StringValue("Round Trip"),
		DatasetJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"metrics-*"}`),
		Density: &datatableDensityModel{
			Mode: types.StringValue("default"),
		},
		Query: &filterSimpleModel{
			Language: types.StringValue("kuery"),
			Query:    types.StringValue(""),
		},
		Metrics: []datatableMetricModel{
			{ConfigJSON: jsontypes.NewNormalizedValue(`{"operation":"count"}`)},
		},
	}

	panel := panelModel{
		Type:            types.StringValue("lens"),
		DatatableConfig: &datatableConfigModel{NoESQL: configModel},
	}

	var apiConfig kbapi.DashboardPanelItem_Config
	diags := converter.mapPanelToAPI(panel, &apiConfig)
	require.False(t, diags.HasError())

	newPanel := panelModel{Type: types.StringValue("lens")}
	diags = converter.populateFromAPIPanel(context.Background(), &newPanel, apiConfig)
	require.False(t, diags.HasError())
	require.NotNil(t, newPanel.DatatableConfig)
	require.NotNil(t, newPanel.DatatableConfig.NoESQL)
	assert.Equal(t, types.StringValue("Round Trip"), newPanel.DatatableConfig.NoESQL.Title)
}

func Test_datatablePanelConfigConverter_roundTrip_ESQL(t *testing.T) {
	converter := newDatatablePanelConfigConverter()
	esqlConfigModel := &datatableESQLConfigModel{
		Title:               types.StringValue("Round Trip ESQL"),
		Description:         types.StringValue("ESQL round-trip test"),
		DatasetJSON:         jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM metrics-* | KEEP host.name, system.cpu.user.pct | LIMIT 10"}`),
		Density:             &datatableDensityModel{Mode: types.StringValue("expanded")},
		IgnoreGlobalFilters: types.BoolValue(false),
		Sampling:            types.Float64Value(1),
		Metrics: []datatableMetricModel{
			{ConfigJSON: jsontypes.NewNormalizedValue(`{"column":"system.cpu.user.pct","operation":"value","format":{"type":"number","decimals":2}}`)},
		},
		Rows: []datatableRowModel{
			{ConfigJSON: jsontypes.NewNormalizedValue(`{"column":"host.name","operation":"value","collapse_by":"avg"}`)},
		},
		SplitMetricsBy: []datatableSplitByModel{
			{ConfigJSON: jsontypes.NewNormalizedValue(`{"column":"host.name","operation":"value"}`)},
		},
		SortByJSON: jsontypes.NewNormalizedValue(`{"column_type":"metric","direction":"desc","index":0}`),
		Paging:     types.Int64Value(20),
	}
	panel := panelModel{
		Type:            types.StringValue("lens"),
		DatatableConfig: &datatableConfigModel{ESQL: esqlConfigModel},
	}

	var apiConfig kbapi.DashboardPanelItem_Config
	diags := converter.mapPanelToAPI(panel, &apiConfig)
	require.False(t, diags.HasError())

	newPanel := panelModel{Type: types.StringValue("lens")}
	diags = converter.populateFromAPIPanel(context.Background(), &newPanel, apiConfig)
	require.False(t, diags.HasError())
	require.NotNil(t, newPanel.DatatableConfig)
	require.NotNil(t, newPanel.DatatableConfig.ESQL)
	assert.Equal(t, types.StringValue("Round Trip ESQL"), newPanel.DatatableConfig.ESQL.Title)
	assert.Equal(t, types.StringValue("ESQL round-trip test"), newPanel.DatatableConfig.ESQL.Description)
	assert.False(t, newPanel.DatatableConfig.ESQL.DatasetJSON.IsNull())
	assert.Equal(t, types.Int64Value(20), newPanel.DatatableConfig.ESQL.Paging)
	assert.Len(t, newPanel.DatatableConfig.ESQL.Metrics, 1)
	assert.Len(t, newPanel.DatatableConfig.ESQL.Rows, 1)
	assert.Len(t, newPanel.DatatableConfig.ESQL.SplitMetricsBy, 1)
}
