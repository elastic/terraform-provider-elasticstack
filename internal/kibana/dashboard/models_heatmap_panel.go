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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newHeatmapPanelConfigConverter() heatmapPanelConfigConverter {
	return heatmapPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.HeatmapNoESQLTypeHeatmap),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.HeatmapConfig != nil },
		},
	}
}

type heatmapPanelConfigConverter struct {
	lensVisualizationBase
}

func (c heatmapPanelConfigConverter) populateFromAttributes(ctx context.Context, pm *panelModel, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	pm.HeatmapConfig = &heatmapConfigModel{}
	if heatmapNoESQL, err := attrs.AsHeatmapNoESQL(); err == nil && !isHeatmapNoESQLCandidateActuallyESQL(heatmapNoESQL) {
		return pm.HeatmapConfig.fromAPINoESQL(ctx, heatmapNoESQL)
	}
	heatmapESQL, err := attrs.AsHeatmapESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.HeatmapConfig.fromAPIESQL(ctx, heatmapESQL)
}

func (c heatmapPanelConfigConverter) buildAttributes(pm panelModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *pm.HeatmapConfig

	attrs, heatmapDiags := configModel.toAPI()
	diags.Append(heatmapDiags...)
	if diags.HasError() {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	return attrs, diags
}

type heatmapConfigModel struct {
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized                              `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []chartFilterJSONModel                            `tfsdk:"filters"`
	Axes                *heatmapAxesModel                                 `tfsdk:"axes"`
	Cells               *heatmapCellsModel                                `tfsdk:"cells"`
	Legend              *heatmapLegendModel                               `tfsdk:"legend"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
	XAxisJSON           jsontypes.Normalized                              `tfsdk:"x_axis_json"`
	YAxisJSON           jsontypes.Normalized                              `tfsdk:"y_axis_json"`
}

func isHeatmapNoESQLCandidateActuallyESQL(apiChart kbapi.HeatmapNoESQL) bool {
	body, err := apiChart.DataSource.MarshalJSON()
	if err != nil {
		return false
	}

	var dataset struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &dataset); err != nil {
		return false
	}

	return dataset.Type == legacyMetricDatasetTypeESQL || dataset.Type == legacyMetricDatasetTypeTable
}

func inferHeatmapXAxisScale(xAxisJSON string) kbapi.HeatmapXAxisScale {
	var axis map[string]any
	if err := json.Unmarshal([]byte(xAxisJSON), &axis); err != nil {
		return kbapi.HeatmapXAxisScaleOrdinal
	}

	operation, _ := axis["operation"].(string)
	switch operation {
	case "date_histogram":
		return kbapi.HeatmapXAxisScaleTemporal
	case "histogram":
		return kbapi.HeatmapXAxisScaleLinear
	default:
		return kbapi.HeatmapXAxisScaleOrdinal
	}
}

func (m *heatmapConfigModel) populateCommonFields(
	title, description *string,
	ignoreGlobalFilters *bool,
	sampling *float32,
	datasetBytes []byte,
	datasetErr error,
	filters []kbapi.LensPanelFilters_Item,
	axes kbapi.HeatmapAxes,
	cells kbapi.HeatmapCells,
	legend kbapi.HeatmapLegend,
	diags *diag.Diagnostics,
) bool {
	m.Title = types.StringPointerValue(title)
	m.Description = types.StringPointerValue(description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(ignoreGlobalFilters)
	if sampling != nil {
		m.Sampling = types.Float64Value(float64(*sampling))
	} else {
		m.Sampling = types.Float64Null()
	}
	dv, ok := marshalToNormalized(datasetBytes, datasetErr, "data_source_json", diags)
	if !ok {
		return false
	}
	m.DataSourceJSON = dv
	m.Filters = populateFiltersFromAPI(filters, diags)
	m.Axes = &heatmapAxesModel{}
	axesDiags := m.Axes.fromAPI(axes)
	diags.Append(axesDiags...)
	m.Cells = &heatmapCellsModel{}
	m.Cells.fromAPI(cells)
	m.Legend = &heatmapLegendModel{}
	m.Legend.fromAPI(legend)
	return !diags.HasError()
}

func (m *heatmapConfigModel) fromAPINoESQL(ctx context.Context, api kbapi.HeatmapNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	if !m.populateCommonFields(api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling, datasetBytes, datasetErr, api.Filters, api.Axes, api.Cells, api.Legend, &diags) {
		return diags
	}

	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric_json", populateTagcloudMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = mv

	xAxisBytes, err := api.X.MarshalJSON()
	xv, ok := marshalToNormalized(xAxisBytes, err, "x_axis_json", &diags)
	if !ok {
		return diags
	}
	m.XAxisJSON = xv

	if api.Y != nil {
		yAxisBytes, err := api.Y.MarshalJSON()
		yv, ok := marshalToNormalized(yAxisBytes, err, "y_axis_json", &diags)
		if !ok {
			return diags
		}
		m.YAxisJSON = yv
	} else {
		m.YAxisJSON = jsontypes.NewNormalizedNull()
	}

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	return diags
}

func (m *heatmapConfigModel) fromAPIESQL(ctx context.Context, api kbapi.HeatmapESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	if !m.populateCommonFields(api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling, datasetBytes, datasetErr, api.Filters, api.Axes, api.Cells, api.Legend, &diags) {
		return diags
	}

	metricBytes, err := json.Marshal(api.Metric)
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric_json", populateTagcloudMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = mv

	xAxisBytes, err := json.Marshal(api.X)
	xv, ok := marshalToNormalized(xAxisBytes, err, "x_axis_json", &diags)
	if !ok {
		return diags
	}
	m.XAxisJSON = xv

	if api.Y != nil {
		yAxisBytes, err := json.Marshal(api.Y)
		yv, ok := marshalToNormalized(yAxisBytes, err, "y_axis_json", &diags)
		if !ok {
			return diags
		}
		m.YAxisJSON = yv
	} else {
		m.YAxisJSON = jsontypes.NewNormalizedNull()
	}

	return diags
}

func (m *heatmapConfigModel) toAPI() (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0

	if m == nil {
		return attrs, diags
	}

	if m.usesESQL() {
		esql, esqlDiags := m.toAPIESQL()
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromHeatmapESQL(esql); err != nil {
			diags.AddError("Failed to create heatmap ESQL schema", err.Error())
		}
		return attrs, diags
	}

	noESQL, noESQLDiags := m.toAPINoESQL()
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromHeatmapNoESQL(noESQL); err != nil {
		diags.AddError("Failed to create heatmap schema", err.Error())
	}

	return attrs, diags
}

func (m *heatmapConfigModel) usesESQL() bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Expression.IsNull() && m.Query.Language.IsNull()
}

func (m *heatmapConfigModel) toAPINoESQL() (kbapi.HeatmapNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.HeatmapNoESQL{
		Type: kbapi.HeatmapNoESQLTypeHeatmap,
	}

	if typeutils.IsKnown(m.Title) {
		api.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = new(m.Description.ValueString())
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = new(m.IgnoreGlobalFilters.ValueBool())
	}
	if typeutils.IsKnown(m.Sampling) {
		api.Sampling = new(float32(m.Sampling.ValueFloat64()))
	}

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing dataset", "heatmap_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal dataset", err.Error())
		return api, diags
	}

	if m.MetricJSON.IsNull() {
		diags.AddError("Missing metric", "heatmap_config.metric_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
		diags.AddError("Failed to unmarshal metric", err.Error())
		return api, diags
	}

	if m.XAxisJSON.IsNull() {
		diags.AddError("Missing x_axis", "heatmap_config.x_axis_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.XAxisJSON.ValueString()), &api.X); err != nil {
		diags.AddError("Failed to unmarshal x_axis_json", err.Error())
		return api, diags
	}

	if !m.YAxisJSON.IsNull() {
		var yAxis kbapi.HeatmapNoESQL_Y
		if err := json.Unmarshal([]byte(m.YAxisJSON.ValueString()), &yAxis); err != nil {
			diags.AddError("Failed to unmarshal y_axis_json", err.Error())
			return api, diags
		}
		api.Y = &yAxis
	}

	if m.Axes == nil {
		diags.AddError("Missing axes", "heatmap_config.axes must be provided")
		return api, diags
	}
	axes, axesDiags := m.Axes.toAPI()
	diags.Append(axesDiags...)
	axes.X.Scale = inferHeatmapXAxisScale(m.XAxisJSON.ValueString())
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

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	return api, diags
}

func (m *heatmapConfigModel) toAPIESQL() (kbapi.HeatmapESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.HeatmapESQL{
		Type: kbapi.HeatmapESQLTypeHeatmap,
	}

	if typeutils.IsKnown(m.Title) {
		api.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = new(m.Description.ValueString())
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = new(m.IgnoreGlobalFilters.ValueBool())
	}
	if typeutils.IsKnown(m.Sampling) {
		api.Sampling = new(float32(m.Sampling.ValueFloat64()))
	}

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing dataset", "heatmap_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal data_source_json", err.Error())
		return api, diags
	}

	if m.MetricJSON.IsNull() {
		diags.AddError("Missing metric", "heatmap_config.metric_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
		diags.AddError("Failed to unmarshal metric_json", err.Error())
		return api, diags
	}

	if m.XAxisJSON.IsNull() {
		diags.AddError("Missing x_axis", "heatmap_config.x_axis_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.XAxisJSON.ValueString()), &api.X); err != nil {
		diags.AddError("Failed to unmarshal x_axis_json", err.Error())
		return api, diags
	}

	if !m.YAxisJSON.IsNull() {
		partial, err := json.Marshal(api)
		if err != nil {
			diags.AddError("Failed to marshal heatmap for y_axis merge", err.Error())
			return api, diags
		}
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(partial, &envelope); err != nil {
			diags.AddError("Failed to prepare heatmap JSON for y_axis merge", err.Error())
			return api, diags
		}
		envelope["y"] = json.RawMessage([]byte(m.YAxisJSON.ValueString()))
		merged, err := json.Marshal(envelope)
		if err != nil {
			diags.AddError("Failed to marshal merged heatmap", err.Error())
			return api, diags
		}
		if err := json.Unmarshal(merged, &api); err != nil {
			diags.AddError("Failed to unmarshal heatmap after y_axis merge", err.Error())
			return api, diags
		}
	}

	if m.Axes == nil {
		diags.AddError("Missing axes", "heatmap_config.axes must be provided")
		return api, diags
	}
	axes, axesDiags := m.Axes.toAPI()
	diags.Append(axesDiags...)
	axes.X.Scale = inferHeatmapXAxisScale(m.XAxisJSON.ValueString())
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

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	return api, diags
}

type heatmapAxesModel struct {
	X *heatmapXAxisModel `tfsdk:"x"`
	Y *heatmapYAxisModel `tfsdk:"y"`
}

func (m *heatmapAxesModel) fromAPI(api kbapi.HeatmapAxes) diag.Diagnostics {
	diags := diag.Diagnostics{}

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
	Orientation kbapi.VisApiOrientation `json:"orientation"`
	Visible     *bool                   `json:"visible,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Orientation = types.StringValue(string(api.Orientation))
	m.Visible = types.BoolPointerValue(api.Visible)
}

func (m *heatmapXAxisLabelsModel) toAPI() *struct {
	Orientation kbapi.VisApiOrientation `json:"orientation"`
	Visible     *bool                   `json:"visible,omitempty"`
} {
	if m == nil {
		return nil
	}
	labels := &struct {
		Orientation kbapi.VisApiOrientation `json:"orientation"`
		Visible     *bool                   `json:"visible,omitempty"`
	}{}
	if typeutils.IsKnown(m.Orientation) {
		labels.Orientation = kbapi.VisApiOrientation(m.Orientation.ValueString())
	}
	if typeutils.IsKnown(m.Visible) {
		labels.Visible = new(m.Visible.ValueBool())
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
	if typeutils.IsKnown(m.Visible) {
		labels.Visible = new(m.Visible.ValueBool())
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
	if typeutils.IsKnown(m.Visible) {
		labels.Visible = new(m.Visible.ValueBool())
	}
	return labels
}

type heatmapLegendModel struct {
	Visibility         types.String `tfsdk:"visibility"`
	Size               types.String `tfsdk:"size"`
	TruncateAfterLines types.Int64  `tfsdk:"truncate_after_lines"`
}

func (m *heatmapLegendModel) fromAPI(api kbapi.HeatmapLegend) {
	if api.Visibility != nil {
		m.Visibility = types.StringValue(string(*api.Visibility))
	} else {
		m.Visibility = types.StringNull()
	}
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

	if typeutils.IsKnown(m.Visibility) {
		visibility := kbapi.HeatmapLegendVisibility(m.Visibility.ValueString())
		legend.Visibility = &visibility
	}
	if typeutils.IsKnown(m.Size) {
		legend.Size = kbapi.LegendSize(m.Size.ValueString())
	} else {
		diags.AddError("Missing legend size", "heatmap_config.legend.size must be provided")
	}
	if typeutils.IsKnown(m.TruncateAfterLines) {
		legend.TruncateAfterLines = new(float32(m.TruncateAfterLines.ValueInt64()))
	}

	return legend, diags
}
