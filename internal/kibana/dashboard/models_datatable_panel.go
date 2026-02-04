package dashboard

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newDatatablePanelConfigConverter() datatablePanelConfigConverter {
	return datatablePanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.DatatableNoESQLTypeDatatable),
		},
	}
}

type datatablePanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c datatablePanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.DatatableConfig != nil
}

func (c datatablePanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	cfgMap, err := config.AsDashboardPanelItemConfig2()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	attrs, ok := cfgMap["attributes"]
	if !ok {
		return nil
	}

	attrsMap, ok := attrs.(map[string]interface{})
	if !ok {
		return nil
	}

	attrsJSON, err := json.Marshal(attrsMap)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var datatableChart kbapi.DatatableChartSchema
	if err := json.Unmarshal(attrsJSON, &datatableChart); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.DatatableConfig = &datatableConfigModel{}

	if _, ok := attrsMap["query"]; ok {
		datatableNoESQL, err := datatableChart.AsDatatableNoESQL()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}

		pm.DatatableConfig.NoESQL = &datatableNoESQLConfigModel{}
		return pm.DatatableConfig.NoESQL.fromAPI(ctx, datatableNoESQL)
	}

	datatableESQL, err := datatableChart.AsDatatableESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.DatatableConfig.ESQL = &datatableESQLConfigModel{}
	return pm.DatatableConfig.ESQL.fromAPI(ctx, datatableESQL)
}

func (c datatablePanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	if pm.DatatableConfig == nil {
		return diags
	}

	var datatableChart kbapi.DatatableChartSchema

	switch {
	case pm.DatatableConfig.NoESQL != nil:
		noESQL, noDiags := pm.DatatableConfig.NoESQL.toAPI()
		diags.Append(noDiags...)
		if diags.HasError() {
			return diags
		}

		if err := datatableChart.FromDatatableNoESQL(noESQL); err != nil {
			diags.AddError("Failed to convert datatable no-esql config", err.Error())
			return diags
		}
	case pm.DatatableConfig.ESQL != nil:
		esql, esqlDiags := pm.DatatableConfig.ESQL.toAPI()
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return diags
		}

		if err := datatableChart.FromDatatableESQL(esql); err != nil {
			diags.AddError("Failed to convert datatable esql config", err.Error())
			return diags
		}
	default:
		return diags
	}

	var attrs0 kbapi.DashboardPanelItemConfig10Attributes0
	if err := attrs0.FromDatatableChartSchema(datatableChart); err != nil {
		diags.AddError("Failed to create datatable attributes", err.Error())
		return diags
	}

	var configAttrs kbapi.DashboardPanelItem_Config_1_0_Attributes
	if err := configAttrs.FromDashboardPanelItemConfig10Attributes0(attrs0); err != nil {
		diags.AddError("Failed to create config attributes", err.Error())
		return diags
	}

	config10 := kbapi.DashboardPanelItemConfig10{
		Attributes: configAttrs,
	}

	var config1 kbapi.DashboardPanelItemConfig1
	if err := config1.FromDashboardPanelItemConfig10(config10); err != nil {
		diags.AddError("Failed to create config1", err.Error())
		return diags
	}

	if err := apiConfig.FromDashboardPanelItemConfig1(config1); err != nil {
		diags.AddError("Failed to marshal datatable config", err.Error())
	}

	return diags
}

type datatableConfigModel struct {
	NoESQL *datatableNoESQLConfigModel `tfsdk:"no_esql"`
	ESQL   *datatableESQLConfigModel   `tfsdk:"esql"`
}

type datatableNoESQLConfigModel struct {
	Title               types.String            `tfsdk:"title"`
	Description         types.String            `tfsdk:"description"`
	Dataset             jsontypes.Normalized    `tfsdk:"dataset"`
	Density             *datatableDensityModel  `tfsdk:"density"`
	IgnoreGlobalFilters types.Bool              `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64           `tfsdk:"sampling"`
	Query               *filterSimpleModel      `tfsdk:"query"`
	Filters             []searchFilterModel     `tfsdk:"filters"`
	Metrics             []datatableMetricModel  `tfsdk:"metrics"`
	Rows                []datatableRowModel     `tfsdk:"rows"`
	SplitMetricsBy      []datatableSplitByModel `tfsdk:"split_metrics_by"`
	SortBy              jsontypes.Normalized    `tfsdk:"sort_by"`
	Paging              types.Int64             `tfsdk:"paging"`
}

type datatableESQLConfigModel struct {
	Title               types.String            `tfsdk:"title"`
	Description         types.String            `tfsdk:"description"`
	Dataset             jsontypes.Normalized    `tfsdk:"dataset"`
	Density             *datatableDensityModel  `tfsdk:"density"`
	IgnoreGlobalFilters types.Bool              `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64           `tfsdk:"sampling"`
	Filters             []searchFilterModel     `tfsdk:"filters"`
	Metrics             []datatableMetricModel  `tfsdk:"metrics"`
	Rows                []datatableRowModel     `tfsdk:"rows"`
	SplitMetricsBy      []datatableSplitByModel `tfsdk:"split_metrics_by"`
	SortBy              jsontypes.Normalized    `tfsdk:"sort_by"`
	Paging              types.Int64             `tfsdk:"paging"`
}

type datatableMetricModel struct {
	Config jsontypes.Normalized `tfsdk:"config"`
}

type datatableRowModel struct {
	Config jsontypes.Normalized `tfsdk:"config"`
}

type datatableSplitByModel struct {
	Config jsontypes.Normalized `tfsdk:"config"`
}

type datatableDensityModel struct {
	Mode   types.String                 `tfsdk:"mode"`
	Height *datatableDensityHeightModel `tfsdk:"height"`
}

type datatableDensityHeightModel struct {
	Header *datatableDensityHeightHeaderModel `tfsdk:"header"`
	Value  *datatableDensityHeightValueModel  `tfsdk:"value"`
}

type datatableDensityHeightHeaderModel struct {
	Type     types.String  `tfsdk:"type"`
	MaxLines types.Float64 `tfsdk:"max_lines"`
}

type datatableDensityHeightValueModel struct {
	Type  types.String  `tfsdk:"type"`
	Lines types.Float64 `tfsdk:"lines"`
}

func (m *datatableNoESQLConfigModel) fromAPI(ctx context.Context, api kbapi.DatatableNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

	datasetBytes, err := json.Marshal(api.Dataset)
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.Dataset = jsontypes.NewNormalizedValue(string(datasetBytes))

	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	m.Density = &datatableDensityModel{}
	if densityDiags := m.Density.fromAPI(api.Density); densityDiags.HasError() {
		return densityDiags
	}

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	if len(api.Metrics) > 0 {
		m.Metrics = make([]datatableMetricModel, len(api.Metrics))
		for i, metric := range api.Metrics {
			metricBytes, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				return diags
			}
			m.Metrics[i].Config = jsontypes.NewNormalizedValue(string(metricBytes))
		}
	}

	if api.Rows != nil && len(*api.Rows) > 0 {
		m.Rows = make([]datatableRowModel, len(*api.Rows))
		for i, row := range *api.Rows {
			rowBytes, err := json.Marshal(row)
			if err != nil {
				diags.AddError("Failed to marshal row", err.Error())
				return diags
			}
			m.Rows[i].Config = jsontypes.NewNormalizedValue(string(rowBytes))
		}
	}

	if api.SplitMetricsBy != nil && len(*api.SplitMetricsBy) > 0 {
		m.SplitMetricsBy = make([]datatableSplitByModel, len(*api.SplitMetricsBy))
		for i, splitBy := range *api.SplitMetricsBy {
			splitBytes, err := json.Marshal(splitBy)
			if err != nil {
				diags.AddError("Failed to marshal split_metrics_by", err.Error())
				return diags
			}
			m.SplitMetricsBy[i].Config = jsontypes.NewNormalizedValue(string(splitBytes))
		}
	}

	if api.SortBy != nil {
		sortBytes, err := json.Marshal(api.SortBy)
		if err != nil {
			diags.AddError("Failed to marshal sort_by", err.Error())
			return diags
		}
		m.SortBy = jsontypes.NewNormalizedValue(string(sortBytes))
	} else {
		m.SortBy = jsontypes.NewNormalizedNull()
	}

	if api.Paging != nil {
		m.Paging = types.Int64Value(int64(*api.Paging))
	} else {
		m.Paging = types.Int64Null()
	}

	return diags
}

func (m *datatableNoESQLConfigModel) toAPI() (kbapi.DatatableNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.DatatableNoESQL{Type: kbapi.DatatableNoESQLTypeDatatable}

	if utils.IsKnown(m.Title) {
		api.Title = m.Title.ValueStringPointer()
	}

	if utils.IsKnown(m.Description) {
		api.Description = m.Description.ValueStringPointer()
	}

	if utils.IsKnown(m.Dataset) {
		if err := json.Unmarshal([]byte(m.Dataset.ValueString()), &api.Dataset); err != nil {
			diags.AddError("Failed to unmarshal dataset", err.Error())
			return api, diags
		}
	}

	if m.Density != nil {
		density, densityDiags := m.Density.toAPI()
		diags.Append(densityDiags...)
		if diags.HasError() {
			return api, diags
		}
		api.Density = density
	}

	if utils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}

	if utils.IsKnown(m.Sampling) {
		sampling := float32(m.Sampling.ValueFloat64())
		api.Sampling = &sampling
	}

	if m.Query != nil {
		api.Query = m.Query.toAPI()
	}

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
		for i, filterModel := range m.Filters {
			filter, filterDiags := filterModel.toAPI()
			diags.Append(filterDiags...)
			filters[i] = filter
		}
		api.Filters = &filters
	}

	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.DatatableNoESQL_Metrics_Item, len(m.Metrics))
		for i, metricModel := range m.Metrics {
			if utils.IsKnown(metricModel.Config) {
				if err := json.Unmarshal([]byte(metricModel.Config.ValueString()), &metrics[i]); err != nil {
					diags.AddError("Failed to unmarshal metric", err.Error())
					return api, diags
				}
			}
		}
		api.Metrics = metrics
	}

	if len(m.Rows) > 0 {
		rows := make([]kbapi.DatatableNoESQL_Rows_Item, len(m.Rows))
		for i, rowModel := range m.Rows {
			if utils.IsKnown(rowModel.Config) {
				if err := json.Unmarshal([]byte(rowModel.Config.ValueString()), &rows[i]); err != nil {
					diags.AddError("Failed to unmarshal row", err.Error())
					return api, diags
				}
			}
		}
		api.Rows = &rows
	}

	if len(m.SplitMetricsBy) > 0 {
		splits := make([]kbapi.DatatableNoESQL_SplitMetricsBy_Item, len(m.SplitMetricsBy))
		for i, splitModel := range m.SplitMetricsBy {
			if utils.IsKnown(splitModel.Config) {
				if err := json.Unmarshal([]byte(splitModel.Config.ValueString()), &splits[i]); err != nil {
					diags.AddError("Failed to unmarshal split_metrics_by", err.Error())
					return api, diags
				}
			}
		}
		api.SplitMetricsBy = &splits
	}

	if utils.IsKnown(m.SortBy) {
		var sortBy kbapi.DatatableNoESQL_SortBy
		if err := json.Unmarshal([]byte(m.SortBy.ValueString()), &sortBy); err != nil {
			diags.AddError("Failed to unmarshal sort_by", err.Error())
			return api, diags
		}
		api.SortBy = &sortBy
	}

	if utils.IsKnown(m.Paging) {
		paging := kbapi.DatatableNoESQLPaging(m.Paging.ValueInt64())
		api.Paging = &paging
	}

	return api, diags
}

func (m *datatableESQLConfigModel) fromAPI(ctx context.Context, api kbapi.DatatableESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

	datasetBytes, err := json.Marshal(api.Dataset)
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.Dataset = jsontypes.NewNormalizedValue(string(datasetBytes))

	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	m.Density = &datatableDensityModel{}
	if densityDiags := m.Density.fromAPI(api.Density); densityDiags.HasError() {
		return densityDiags
	}

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	if len(api.Metrics) > 0 {
		m.Metrics = make([]datatableMetricModel, len(api.Metrics))
		for i, metric := range api.Metrics {
			metricBytes, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				return diags
			}
			m.Metrics[i].Config = jsontypes.NewNormalizedValue(string(metricBytes))
		}
	}

	if api.Rows != nil && len(*api.Rows) > 0 {
		m.Rows = make([]datatableRowModel, len(*api.Rows))
		for i, row := range *api.Rows {
			rowBytes, err := json.Marshal(row)
			if err != nil {
				diags.AddError("Failed to marshal row", err.Error())
				return diags
			}
			m.Rows[i].Config = jsontypes.NewNormalizedValue(string(rowBytes))
		}
	}

	if api.SplitMetricsBy != nil && len(*api.SplitMetricsBy) > 0 {
		m.SplitMetricsBy = make([]datatableSplitByModel, len(*api.SplitMetricsBy))
		for i, splitBy := range *api.SplitMetricsBy {
			splitBytes, err := json.Marshal(splitBy)
			if err != nil {
				diags.AddError("Failed to marshal split_metrics_by", err.Error())
				return diags
			}
			m.SplitMetricsBy[i].Config = jsontypes.NewNormalizedValue(string(splitBytes))
		}
	}

	if api.SortBy != nil {
		sortBytes, err := json.Marshal(api.SortBy)
		if err != nil {
			diags.AddError("Failed to marshal sort_by", err.Error())
			return diags
		}
		m.SortBy = jsontypes.NewNormalizedValue(string(sortBytes))
	} else {
		m.SortBy = jsontypes.NewNormalizedNull()
	}

	if api.Paging != nil {
		m.Paging = types.Int64Value(int64(*api.Paging))
	} else {
		m.Paging = types.Int64Null()
	}

	return diags
}

func (m *datatableESQLConfigModel) toAPI() (kbapi.DatatableESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.DatatableESQL{Type: kbapi.DatatableESQLTypeDatatable}

	if utils.IsKnown(m.Title) {
		api.Title = m.Title.ValueStringPointer()
	}

	if utils.IsKnown(m.Description) {
		api.Description = m.Description.ValueStringPointer()
	}

	if utils.IsKnown(m.Dataset) {
		if err := json.Unmarshal([]byte(m.Dataset.ValueString()), &api.Dataset); err != nil {
			diags.AddError("Failed to unmarshal dataset", err.Error())
			return api, diags
		}
	}

	if m.Density != nil {
		density, densityDiags := m.Density.toAPI()
		diags.Append(densityDiags...)
		if diags.HasError() {
			return api, diags
		}
		api.Density = density
	}

	if utils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}

	if utils.IsKnown(m.Sampling) {
		sampling := float32(m.Sampling.ValueFloat64())
		api.Sampling = &sampling
	}

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
		for i, filterModel := range m.Filters {
			filter, filterDiags := filterModel.toAPI()
			diags.Append(filterDiags...)
			filters[i] = filter
		}
		api.Filters = &filters
	}

	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.DatatableESQLMetric, len(m.Metrics))
		for i, metricModel := range m.Metrics {
			if utils.IsKnown(metricModel.Config) {
				if err := json.Unmarshal([]byte(metricModel.Config.ValueString()), &metrics[i]); err != nil {
					diags.AddError("Failed to unmarshal metric", err.Error())
					return api, diags
				}
			}
		}
		api.Metrics = metrics
	}

	if len(m.Rows) > 0 {
		rows := make([]struct {
			Alignment    *kbapi.DatatableESQLRowsAlignment    `json:"alignment,omitempty"`
			ApplyColorTo *kbapi.DatatableESQLRowsApplyColorTo `json:"apply_color_to,omitempty"`
			ClickFilter  *bool                                `json:"click_filter,omitempty"`
			CollapseBy   kbapi.CollapseBy                     `json:"collapse_by"`
			Color        *kbapi.DatatableESQL_Rows_Color      `json:"color,omitempty"`
			Column       string                               `json:"column"`
			Operation    kbapi.DatatableESQLRowsOperation     `json:"operation"`
			Visible      *bool                                `json:"visible,omitempty"`
			Width        *float32                             `json:"width,omitempty"`
		}, len(m.Rows))
		for i, rowModel := range m.Rows {
			if utils.IsKnown(rowModel.Config) {
				if err := json.Unmarshal([]byte(rowModel.Config.ValueString()), &rows[i]); err != nil {
					diags.AddError("Failed to unmarshal row", err.Error())
					return api, diags
				}
			}
		}
		api.Rows = &rows
	}

	if len(m.SplitMetricsBy) > 0 {
		splits := make([]struct {
			Column    string                                     `json:"column"`
			Operation kbapi.DatatableESQLSplitMetricsByOperation `json:"operation"`
		}, len(m.SplitMetricsBy))
		for i, splitModel := range m.SplitMetricsBy {
			if utils.IsKnown(splitModel.Config) {
				if err := json.Unmarshal([]byte(splitModel.Config.ValueString()), &splits[i]); err != nil {
					diags.AddError("Failed to unmarshal split_metrics_by", err.Error())
					return api, diags
				}
			}
		}
		api.SplitMetricsBy = &splits
	}

	if utils.IsKnown(m.SortBy) {
		var sortBy kbapi.DatatableESQL_SortBy
		if err := json.Unmarshal([]byte(m.SortBy.ValueString()), &sortBy); err != nil {
			diags.AddError("Failed to unmarshal sort_by", err.Error())
			return api, diags
		}
		api.SortBy = &sortBy
	}

	if utils.IsKnown(m.Paging) {
		paging := kbapi.DatatableESQLPaging(m.Paging.ValueInt64())
		api.Paging = &paging
	}

	return api, diags
}

func (m *datatableDensityModel) fromAPI(api kbapi.DatatableDensity) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Mode = typeutils.StringishPointerValue(api.Mode)

	if api.Height != nil {
		m.Height = &datatableDensityHeightModel{}
		heightDiags := m.Height.fromAPI(api.Height)
		diags.Append(heightDiags...)
	}

	return diags
}

func (m *datatableDensityModel) toAPI() (kbapi.DatatableDensity, diag.Diagnostics) {
	if m == nil {
		return kbapi.DatatableDensity{}, nil
	}

	var diags diag.Diagnostics
	var density kbapi.DatatableDensity

	if utils.IsKnown(m.Mode) {
		mode := kbapi.DatatableDensityMode(m.Mode.ValueString())
		density.Mode = &mode
	}

	if m.Height != nil {
		height := &struct {
			Header *kbapi.DatatableDensity_Height_Header `json:"header,omitempty"`
			Value  *kbapi.DatatableDensity_Height_Value  `json:"value,omitempty"`
		}{}

		if m.Height.Header != nil {
			header, headerDiags := m.Height.Header.toAPI()
			diags.Append(headerDiags...)
			if diags.HasError() {
				return density, diags
			}
			height.Header = header
		}

		if m.Height.Value != nil {
			value, valueDiags := m.Height.Value.toAPI()
			diags.Append(valueDiags...)
			if diags.HasError() {
				return density, diags
			}
			height.Value = value
		}

		density.Height = height
	}

	return density, diags
}

func (m *datatableDensityHeightModel) fromAPI(api *struct {
	Header *kbapi.DatatableDensity_Height_Header `json:"header,omitempty"`
	Value  *kbapi.DatatableDensity_Height_Value  `json:"value,omitempty"`
}) diag.Diagnostics {
	var diags diag.Diagnostics
	if api == nil {
		return diags
	}

	if api.Header != nil {
		m.Header = &datatableDensityHeightHeaderModel{}
		headerDiags := m.Header.fromAPI(api.Header)
		diags.Append(headerDiags...)
	}

	if api.Value != nil {
		m.Value = &datatableDensityHeightValueModel{}
		valueDiags := m.Value.fromAPI(api.Value)
		diags.Append(valueDiags...)
	}

	return diags
}

func (m *datatableDensityHeightHeaderModel) fromAPI(api *kbapi.DatatableDensity_Height_Header) diag.Diagnostics {
	var diags diag.Diagnostics
	if api == nil {
		return diags
	}

	raw, err := api.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal header density", err.Error())
		return diags
	}

	var header struct {
		Type     string   `json:"type"`
		MaxLines *float32 `json:"max_lines,omitempty"`
	}

	if err := json.Unmarshal(raw, &header); err != nil {
		diags.AddError("Failed to unmarshal header density", err.Error())
		return diags
	}

	m.Type = types.StringValue(header.Type)
	if header.MaxLines != nil {
		m.MaxLines = types.Float64Value(float64(*header.MaxLines))
	} else {
		m.MaxLines = types.Float64Null()
	}

	return diags
}

func (m *datatableDensityHeightHeaderModel) toAPI() (*kbapi.DatatableDensity_Height_Header, diag.Diagnostics) {
	if m == nil || !utils.IsKnown(m.Type) {
		return nil, nil
	}

	var diags diag.Diagnostics
	var header kbapi.DatatableDensity_Height_Header

	switch m.Type.ValueString() {
	case "auto":
		auto := kbapi.DatatableDensityHeightHeader0{Type: kbapi.DatatableDensityHeightHeader0TypeAuto}
		if err := header.FromDatatableDensityHeightHeader0(auto); err != nil {
			diags.AddError("Failed to marshal header density", err.Error())
			return nil, diags
		}
	case "custom":
		custom := kbapi.DatatableDensityHeightHeader1{Type: kbapi.DatatableDensityHeightHeader1TypeCustom}
		if utils.IsKnown(m.MaxLines) {
			maxLines := float32(m.MaxLines.ValueFloat64())
			custom.MaxLines = &maxLines
		}
		if err := header.FromDatatableDensityHeightHeader1(custom); err != nil {
			diags.AddError("Failed to marshal header density", err.Error())
			return nil, diags
		}
	default:
		return nil, diags
	}

	return &header, diags
}

func (m *datatableDensityHeightValueModel) fromAPI(api *kbapi.DatatableDensity_Height_Value) diag.Diagnostics {
	var diags diag.Diagnostics
	if api == nil {
		return diags
	}

	raw, err := api.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal value density", err.Error())
		return diags
	}

	var value struct {
		Type  string   `json:"type"`
		Lines *float32 `json:"lines,omitempty"`
	}

	if err := json.Unmarshal(raw, &value); err != nil {
		diags.AddError("Failed to unmarshal value density", err.Error())
		return diags
	}

	m.Type = types.StringValue(value.Type)
	if value.Lines != nil {
		m.Lines = types.Float64Value(float64(*value.Lines))
	} else {
		m.Lines = types.Float64Null()
	}

	return diags
}

func (m *datatableDensityHeightValueModel) toAPI() (*kbapi.DatatableDensity_Height_Value, diag.Diagnostics) {
	if m == nil || !utils.IsKnown(m.Type) {
		return nil, nil
	}

	var diags diag.Diagnostics
	var value kbapi.DatatableDensity_Height_Value

	switch m.Type.ValueString() {
	case "auto":
		auto := kbapi.DatatableDensityHeightValue0{Type: kbapi.DatatableDensityHeightValue0TypeAuto}
		if err := value.FromDatatableDensityHeightValue0(auto); err != nil {
			diags.AddError("Failed to marshal value density", err.Error())
			return nil, diags
		}
	case "custom":
		custom := kbapi.DatatableDensityHeightValue1{Type: kbapi.DatatableDensityHeightValue1TypeCustom}
		if utils.IsKnown(m.Lines) {
			lines := float32(m.Lines.ValueFloat64())
			custom.Lines = &lines
		}
		if err := value.FromDatatableDensityHeightValue1(custom); err != nil {
			diags.AddError("Failed to marshal value density", err.Error())
			return nil, diags
		}
	default:
		return nil, diags
	}

	return &value, diags
}
