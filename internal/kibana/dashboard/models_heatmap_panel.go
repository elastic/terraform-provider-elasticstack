package dashboard

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newHeatmapPanelConfigConverter() heatmapPanelConfigConverter {
	return heatmapPanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.HeatmapNoESQLTypeHeatmap),
		},
	}
}

type heatmapPanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c heatmapPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.HeatmapConfig != nil
}

func (c heatmapPanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	cfgMap, err := config.AsDashboardPanelItemConfig2()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	attrs, ok := cfgMap["attributes"]
	if !ok {
		return nil
	}

	attrsMap, ok := attrs.(map[string]any)
	if !ok {
		return nil
	}

	attrsJSON, err := json.Marshal(attrsMap)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var heatmapChart kbapi.HeatmapChartSchema
	if err := json.Unmarshal(attrsJSON, &heatmapChart); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	_, hasQuery := attrsMap["query"]

	pm.HeatmapConfig = &heatmapConfigModel{}
	if hasQuery {
		heatmapNoESQL, err := heatmapChart.AsHeatmapNoESQL()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		return pm.HeatmapConfig.fromAPINoESQL(ctx, heatmapNoESQL)
	}

	heatmapESQL, err := heatmapChart.AsHeatmapESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.HeatmapConfig.fromAPIESQL(ctx, heatmapESQL)
}

func (c heatmapPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	configModel := *pm.HeatmapConfig

	heatmapChart, heatmapDiags := configModel.toAPI()
	diags.Append(heatmapDiags...)
	if diags.HasError() {
		return diags
	}

	var attrs0 kbapi.DashboardPanelItemConfig10Attributes0
	if err := attrs0.FromHeatmapChartSchema(heatmapChart); err != nil {
		diags.AddError("Failed to create heatmap attributes", err.Error())
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
		diags.AddError("Failed to marshal heatmap config", err.Error())
	}

	return diags
}

type heatmapConfigModel struct {
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	Dataset             jsontypes.Normalized                              `tfsdk:"dataset"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []searchFilterModel                               `tfsdk:"filters"`
	Axes                *heatmapAxesModel                                 `tfsdk:"axes"`
	Cells               *heatmapCellsModel                                `tfsdk:"cells"`
	Legend              *heatmapLegendModel                               `tfsdk:"legend"`
	Metric              customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric"`
	XAxis               jsontypes.Normalized                              `tfsdk:"x_axis"`
	YAxis               jsontypes.Normalized                              `tfsdk:"y_axis"`
}

func (m *heatmapConfigModel) fromAPINoESQL(ctx context.Context, api kbapi.HeatmapNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)

	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.Dataset = jsontypes.NewNormalizedValue(string(datasetBytes))

	metricBytes, err := api.Metric.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal metric", err.Error())
		return diags
	}
	m.Metric = customtypes.NewJSONWithDefaultsValue[map[string]any](
		string(metricBytes),
		populateTagcloudMetricDefaults,
	)

	xAxisBytes, err := api.XAxis.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal x_axis", err.Error())
		return diags
	}
	m.XAxis = jsontypes.NewNormalizedValue(string(xAxisBytes))

	if api.YAxis != nil {
		yAxisBytes, err := api.YAxis.MarshalJSON()
		if err != nil {
			diags.AddError("Failed to marshal y_axis", err.Error())
			return diags
		}
		m.YAxis = jsontypes.NewNormalizedValue(string(yAxisBytes))
	} else {
		m.YAxis = jsontypes.NewNormalizedNull()
	}

	m.Axes = &heatmapAxesModel{}
	axesDiags := m.Axes.fromAPI(api.Axes)
	diags.Append(axesDiags...)

	m.Cells = &heatmapCellsModel{}
	m.Cells.fromAPI(api.Cells)

	m.Legend = &heatmapLegendModel{}
	m.Legend.fromAPI(api.Legend)

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, 0, len(*api.Filters))
		for _, filter := range *api.Filters {
			filterModel := searchFilterModel{}
			filterDiags := filterModel.fromAPI(filter)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				m.Filters = append(m.Filters, filterModel)
			}
		}
	}

	return diags
}

func (m *heatmapConfigModel) fromAPIESQL(ctx context.Context, api kbapi.HeatmapESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)

	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.Dataset = jsontypes.NewNormalizedValue(string(datasetBytes))

	metricBytes, err := json.Marshal(api.Metric)
	if err != nil {
		diags.AddError("Failed to marshal metric", err.Error())
		return diags
	}
	m.Metric = customtypes.NewJSONWithDefaultsValue[map[string]any](
		string(metricBytes),
		populateTagcloudMetricDefaults,
	)

	xAxisBytes, err := json.Marshal(api.XAxis)
	if err != nil {
		diags.AddError("Failed to marshal x_axis", err.Error())
		return diags
	}
	m.XAxis = jsontypes.NewNormalizedValue(string(xAxisBytes))

	if api.YAxis != nil {
		yAxisBytes, err := json.Marshal(api.YAxis)
		if err != nil {
			diags.AddError("Failed to marshal y_axis", err.Error())
			return diags
		}
		m.YAxis = jsontypes.NewNormalizedValue(string(yAxisBytes))
	} else {
		m.YAxis = jsontypes.NewNormalizedNull()
	}

	m.Axes = &heatmapAxesModel{}
	axesDiags := m.Axes.fromAPI(api.Axes)
	diags.Append(axesDiags...)

	m.Cells = &heatmapCellsModel{}
	m.Cells.fromAPI(api.Cells)

	m.Legend = &heatmapLegendModel{}
	m.Legend.fromAPI(api.Legend)

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, 0, len(*api.Filters))
		for _, filter := range *api.Filters {
			filterModel := searchFilterModel{}
			filterDiags := filterModel.fromAPI(filter)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				m.Filters = append(m.Filters, filterModel)
			}
		}
	}

	return diags
}

func (m *heatmapConfigModel) toAPI() (kbapi.HeatmapChartSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	var heatmapChart kbapi.HeatmapChartSchema

	if m == nil {
		return heatmapChart, diags
	}

	if m.usesESQL() {
		esql, esqlDiags := m.toAPIESQL()
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return heatmapChart, diags
		}
		if err := heatmapChart.FromHeatmapESQL(esql); err != nil {
			diags.AddError("Failed to create heatmap ESQL schema", err.Error())
		}
		return heatmapChart, diags
	}

	noESQL, noESQLDiags := m.toAPINoESQL()
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return heatmapChart, diags
	}
	if err := heatmapChart.FromHeatmapNoESQL(noESQL); err != nil {
		diags.AddError("Failed to create heatmap schema", err.Error())
	}

	return heatmapChart, diags
}

func (m *heatmapConfigModel) usesESQL() bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Query.IsNull() && m.Query.Language.IsNull()
}

func (m *heatmapConfigModel) toAPINoESQL() (kbapi.HeatmapNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.HeatmapNoESQL{
		Type: kbapi.HeatmapNoESQLTypeHeatmap,
	}

	if utils.IsKnown(m.Title) {
		api.Title = utils.Pointer(m.Title.ValueString())
	}
	if utils.IsKnown(m.Description) {
		api.Description = utils.Pointer(m.Description.ValueString())
	}
	if utils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = utils.Pointer(m.IgnoreGlobalFilters.ValueBool())
	}
	if utils.IsKnown(m.Sampling) {
		api.Sampling = utils.Pointer(float32(m.Sampling.ValueFloat64()))
	}

	if m.Dataset.IsNull() {
		diags.AddError("Missing dataset", "heatmap_config.dataset must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.Dataset.ValueString()), &api.Dataset); err != nil {
		diags.AddError("Failed to unmarshal dataset", err.Error())
		return api, diags
	}

	if m.Metric.IsNull() {
		diags.AddError("Missing metric", "heatmap_config.metric must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.Metric.ValueString()), &api.Metric); err != nil {
		diags.AddError("Failed to unmarshal metric", err.Error())
		return api, diags
	}

	if m.XAxis.IsNull() {
		diags.AddError("Missing x_axis", "heatmap_config.x_axis must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.XAxis.ValueString()), &api.XAxis); err != nil {
		diags.AddError("Failed to unmarshal x_axis", err.Error())
		return api, diags
	}

	if !m.YAxis.IsNull() {
		var yAxis kbapi.HeatmapNoESQL_YAxis
		if err := json.Unmarshal([]byte(m.YAxis.ValueString()), &yAxis); err != nil {
			diags.AddError("Failed to unmarshal y_axis", err.Error())
			return api, diags
		}
		api.YAxis = &yAxis
	}

	if m.Axes == nil {
		diags.AddError("Missing axes", "heatmap_config.axes must be provided")
		return api, diags
	}
	axes, axesDiags := m.Axes.toAPI()
	diags.Append(axesDiags...)
	api.Axes = axes

	if m.Cells == nil {
		diags.AddError("Missing cells", "heatmap_config.cells must be provided")
		return api, diags
	}
	api.Cells = m.Cells.toAPI()

	if m.Legend == nil {
		diags.AddError("Missing legend", "heatmap_config.legend must be provided")
		return api, diags
	}
	legend, legendDiags := m.Legend.toAPI()
	diags.Append(legendDiags...)
	api.Legend = legend

	if m.Query == nil {
		diags.AddError("Missing query", "heatmap_config.query must be provided for non-ES|QL heatmaps")
		return api, diags
	}
	api.Query = m.Query.toAPI()

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilterSchema, 0, len(m.Filters))
		for _, filter := range m.Filters {
			apiFilter, filterDiags := filter.toAPI()
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				filters = append(filters, apiFilter)
			}
		}
		if len(filters) > 0 {
			api.Filters = &filters
		}
	}

	return api, diags
}

func (m *heatmapConfigModel) toAPIESQL() (kbapi.HeatmapESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.HeatmapESQL{
		Type: kbapi.HeatmapESQLTypeHeatmap,
	}

	if utils.IsKnown(m.Title) {
		api.Title = utils.Pointer(m.Title.ValueString())
	}
	if utils.IsKnown(m.Description) {
		api.Description = utils.Pointer(m.Description.ValueString())
	}
	if utils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = utils.Pointer(m.IgnoreGlobalFilters.ValueBool())
	}
	if utils.IsKnown(m.Sampling) {
		api.Sampling = utils.Pointer(float32(m.Sampling.ValueFloat64()))
	}

	if m.Dataset.IsNull() {
		diags.AddError("Missing dataset", "heatmap_config.dataset must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.Dataset.ValueString()), &api.Dataset); err != nil {
		diags.AddError("Failed to unmarshal dataset", err.Error())
		return api, diags
	}

	if m.Metric.IsNull() {
		diags.AddError("Missing metric", "heatmap_config.metric must be provided")
		return api, diags
	}
	var metric struct {
		Color     kbapi.ColorByValue               `json:"color"`
		Column    string                           `json:"column"`
		Operation kbapi.HeatmapESQLMetricOperation `json:"operation"`
	}
	if err := json.Unmarshal([]byte(m.Metric.ValueString()), &metric); err != nil {
		diags.AddError("Failed to unmarshal metric", err.Error())
		return api, diags
	}
	api.Metric = metric

	if m.XAxis.IsNull() {
		diags.AddError("Missing x_axis", "heatmap_config.x_axis must be provided")
		return api, diags
	}
	var xAxis struct {
		Column    string                          `json:"column"`
		Operation kbapi.HeatmapESQLXAxisOperation `json:"operation"`
	}
	if err := json.Unmarshal([]byte(m.XAxis.ValueString()), &xAxis); err != nil {
		diags.AddError("Failed to unmarshal x_axis", err.Error())
		return api, diags
	}
	api.XAxis = xAxis

	if !m.YAxis.IsNull() {
		var yAxis struct {
			Column    string                          `json:"column"`
			Operation kbapi.HeatmapESQLYAxisOperation `json:"operation"`
		}
		if err := json.Unmarshal([]byte(m.YAxis.ValueString()), &yAxis); err != nil {
			diags.AddError("Failed to unmarshal y_axis", err.Error())
			return api, diags
		}
		api.YAxis = &yAxis
	}

	if m.Axes == nil {
		diags.AddError("Missing axes", "heatmap_config.axes must be provided")
		return api, diags
	}
	axes, axesDiags := m.Axes.toAPI()
	diags.Append(axesDiags...)
	api.Axes = axes

	if m.Cells == nil {
		diags.AddError("Missing cells", "heatmap_config.cells must be provided")
		return api, diags
	}
	api.Cells = m.Cells.toAPI()

	if m.Legend == nil {
		diags.AddError("Missing legend", "heatmap_config.legend must be provided")
		return api, diags
	}
	legend, legendDiags := m.Legend.toAPI()
	diags.Append(legendDiags...)
	api.Legend = legend

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilterSchema, 0, len(m.Filters))
		for _, filter := range m.Filters {
			apiFilter, filterDiags := filter.toAPI()
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				filters = append(filters, apiFilter)
			}
		}
		if len(filters) > 0 {
			api.Filters = &filters
		}
	}

	return api, diags
}

type heatmapAxesModel struct {
	X *heatmapXAxisModel `tfsdk:"x"`
	Y *heatmapYAxisModel `tfsdk:"y"`
}

func (m *heatmapAxesModel) fromAPI(api kbapi.HeatmapAxes) diag.Diagnostics {
	var diags diag.Diagnostics

	m.X = &heatmapXAxisModel{}
	m.X.fromAPI(api.X)

	m.Y = &heatmapYAxisModel{}
	m.Y.fromAPI(api.Y)

	return diags
}

func (m *heatmapAxesModel) toAPI() (kbapi.HeatmapAxes, diag.Diagnostics) {
	var diags diag.Diagnostics
	axes := kbapi.HeatmapAxes{}

	if m == nil {
		diags.AddError("Missing axes", "heatmap_config.axes must be provided")
		return axes, diags
	}

	if m.X != nil {
		axes.X = m.X.toAPI()
	}
	if m.Y != nil {
		axes.Y = m.Y.toAPI()
	}

	return axes, diags
}

type heatmapXAxisModel struct {
	Labels *heatmapXAxisLabelsModel `tfsdk:"labels"`
	Title  *axisTitleModel          `tfsdk:"title"`
}

func (m *heatmapXAxisModel) fromAPI(api kbapi.HeatmapXAxis) {
	if api.Labels != nil {
		m.Labels = &heatmapXAxisLabelsModel{}
		m.Labels.fromAPI(api.Labels)
	}
	if api.Title != nil {
		m.Title = &axisTitleModel{}
		m.Title.fromAPI(api.Title)
	}
}

func (m *heatmapXAxisModel) toAPI() kbapi.HeatmapXAxis {
	axis := kbapi.HeatmapXAxis{}
	if m == nil {
		return axis
	}
	if m.Labels != nil {
		axis.Labels = m.Labels.toAPI()
	}
	if m.Title != nil {
		axis.Title = m.Title.toAPI()
	}
	return axis
}

type heatmapXAxisLabelsModel struct {
	Orientation types.String `tfsdk:"orientation"`
	Visible     types.Bool   `tfsdk:"visible"`
}

func (m *heatmapXAxisLabelsModel) fromAPI(api *struct {
	Orientation *kbapi.HeatmapXAxisLabelsOrientation `json:"orientation,omitempty"`
	Visible     *bool                                `json:"visible,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Orientation = typeutils.StringishPointerValue(api.Orientation)
	m.Visible = types.BoolPointerValue(api.Visible)
}

func (m *heatmapXAxisLabelsModel) toAPI() *struct {
	Orientation *kbapi.HeatmapXAxisLabelsOrientation `json:"orientation,omitempty"`
	Visible     *bool                                `json:"visible,omitempty"`
} {
	if m == nil {
		return nil
	}
	labels := &struct {
		Orientation *kbapi.HeatmapXAxisLabelsOrientation `json:"orientation,omitempty"`
		Visible     *bool                                `json:"visible,omitempty"`
	}{}
	if utils.IsKnown(m.Orientation) {
		orientation := kbapi.HeatmapXAxisLabelsOrientation(m.Orientation.ValueString())
		labels.Orientation = &orientation
	}
	if utils.IsKnown(m.Visible) {
		labels.Visible = utils.Pointer(m.Visible.ValueBool())
	}
	return labels
}

type heatmapYAxisModel struct {
	Labels *heatmapYAxisLabelsModel `tfsdk:"labels"`
	Title  *axisTitleModel          `tfsdk:"title"`
}

func (m *heatmapYAxisModel) fromAPI(api kbapi.HeatmapYAxis) {
	if api.Labels != nil {
		m.Labels = &heatmapYAxisLabelsModel{}
		m.Labels.fromAPI(api.Labels)
	}
	if api.Title != nil {
		m.Title = &axisTitleModel{}
		m.Title.fromAPI(api.Title)
	}
}

func (m *heatmapYAxisModel) toAPI() kbapi.HeatmapYAxis {
	axis := kbapi.HeatmapYAxis{}
	if m == nil {
		return axis
	}
	if m.Labels != nil {
		axis.Labels = m.Labels.toAPI()
	}
	if m.Title != nil {
		axis.Title = m.Title.toAPI()
	}
	return axis
}

type heatmapYAxisLabelsModel struct {
	Visible types.Bool `tfsdk:"visible"`
}

func (m *heatmapYAxisLabelsModel) fromAPI(api *struct {
	Visible *bool `json:"visible,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Visible = types.BoolPointerValue(api.Visible)
}

func (m *heatmapYAxisLabelsModel) toAPI() *struct {
	Visible *bool `json:"visible,omitempty"`
} {
	if m == nil {
		return nil
	}
	labels := &struct {
		Visible *bool `json:"visible,omitempty"`
	}{}
	if utils.IsKnown(m.Visible) {
		labels.Visible = utils.Pointer(m.Visible.ValueBool())
	}
	return labels
}

type heatmapCellsModel struct {
	Labels *heatmapCellsLabelsModel `tfsdk:"labels"`
}

func (m *heatmapCellsModel) fromAPI(api kbapi.HeatmapCells) {
	if api.Labels != nil {
		m.Labels = &heatmapCellsLabelsModel{}
		m.Labels.fromAPI(api.Labels)
	}
}

func (m *heatmapCellsModel) toAPI() kbapi.HeatmapCells {
	cells := kbapi.HeatmapCells{}
	if m == nil {
		return cells
	}
	if m.Labels != nil {
		cells.Labels = m.Labels.toAPI()
	}
	return cells
}

type heatmapCellsLabelsModel struct {
	Visible types.Bool `tfsdk:"visible"`
}

func (m *heatmapCellsLabelsModel) fromAPI(api *struct {
	Visible *bool `json:"visible,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Visible = types.BoolPointerValue(api.Visible)
}

func (m *heatmapCellsLabelsModel) toAPI() *struct {
	Visible *bool `json:"visible,omitempty"`
} {
	if m == nil {
		return nil
	}
	labels := &struct {
		Visible *bool `json:"visible,omitempty"`
	}{}
	if utils.IsKnown(m.Visible) {
		labels.Visible = utils.Pointer(m.Visible.ValueBool())
	}
	return labels
}

type heatmapLegendModel struct {
	Visible            types.Bool   `tfsdk:"visible"`
	Position           types.String `tfsdk:"position"`
	Size               types.String `tfsdk:"size"`
	TruncateAfterLines types.Int64  `tfsdk:"truncate_after_lines"`
}

func (m *heatmapLegendModel) fromAPI(api kbapi.HeatmapLegend) {
	m.Visible = types.BoolPointerValue(api.Visible)
	m.Position = typeutils.StringishPointerValue(api.Position)
	m.Size = types.StringValue(string(api.Size))

	if api.TruncateAfterLines != nil {
		m.TruncateAfterLines = types.Int64Value(int64(*api.TruncateAfterLines))
	} else {
		m.TruncateAfterLines = types.Int64Null()
	}
}

func (m *heatmapLegendModel) toAPI() (kbapi.HeatmapLegend, diag.Diagnostics) {
	var diags diag.Diagnostics
	legend := kbapi.HeatmapLegend{}

	if m == nil {
		diags.AddError("Missing legend", "heatmap_config.legend must be provided")
		return legend, diags
	}

	if utils.IsKnown(m.Visible) {
		legend.Visible = utils.Pointer(m.Visible.ValueBool())
	}
	if utils.IsKnown(m.Position) {
		pos := kbapi.HeatmapLegendPosition(m.Position.ValueString())
		legend.Position = &pos
	}
	if utils.IsKnown(m.Size) {
		legend.Size = kbapi.LegendSize(m.Size.ValueString())
	} else {
		diags.AddError("Missing legend size", "heatmap_config.legend.size must be provided")
	}
	if utils.IsKnown(m.TruncateAfterLines) {
		legend.TruncateAfterLines = utils.Pointer(float32(m.TruncateAfterLines.ValueInt64()))
	}

	return legend, diags
}
