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

package lensheatmap

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func inferHeatmapXAxisScale(xAxisJSON string) kbapi.KibanaHTTPAPIsHeatmapXAxisScale {
	var axis map[string]any
	if err := json.Unmarshal([]byte(xAxisJSON), &axis); err != nil {
		return kbapi.KibanaHTTPAPIsHeatmapXAxisScaleOrdinal
	}

	operation, _ := axis["operation"].(string)
	switch operation {
	case "date_histogram":
		return kbapi.KibanaHTTPAPIsHeatmapXAxisScaleTemporal
	case "histogram":
		return kbapi.KibanaHTTPAPIsHeatmapXAxisScaleLinear
	default:
		return kbapi.KibanaHTTPAPIsHeatmapXAxisScaleOrdinal
	}
}

func heatmapConfigPopulateCommonFields(m *models.HeatmapConfigModel,
	title, description *string,
	ignoreGlobalFilters *bool,
	sampling *float32,
	datasetBytes []byte,
	datasetErr error,
	filters *kbapi.KibanaHTTPAPIsLensPanelFilters,
	axis *kbapi.KibanaHTTPAPIsHeatmapAxes,
	styling *kbapi.KibanaHTTPAPIsHeatmapStyling,
	legend *kbapi.KibanaHTTPAPIsHeatmapLegend,
	prior *models.HeatmapConfigModel,
	diags *diag.Diagnostics,
) bool {
	base, ok := lenscommon.PopulateLensChartBaseFromAPI(
		title, description, ignoreGlobalFilters, sampling,
		datasetBytes, datasetErr, "data_source_json", filters, diags,
	)
	if !ok {
		return false
	}
	m.LensChartBaseTFModel = base
	m.Axis = &models.HeatmapAxesModel{}
	var priorAxis *models.HeatmapAxesModel
	if prior != nil {
		priorAxis = prior.Axis
	}
	axisDiags := heatmapAxesFromAPI(m.Axis, axis, priorAxis)
	diags.Append(axisDiags...)
	m.Styling = &models.HeatmapStylingModel{}
	heatmapStylingFromAPI(m.Styling, styling)
	if legend == nil {
		m.Legend = nil
	} else {
		m.Legend = &models.HeatmapLegendModel{}
		heatmapLegendFromAPI(m.Legend, legend)
	}
	return !diags.HasError()
}

func heatmapConfigFromAPINoESQL(
	ctx context.Context,
	m *models.HeatmapConfigModel,
	prior *models.HeatmapConfigModel,
	api kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	if !heatmapConfigPopulateCommonFields(m, api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling, datasetBytes, datasetErr, api.Filters, api.Axis, api.Styling, api.Legend, prior, &diags) {
		return diags
	}

	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := lenscommon.MarshalToJSONWithDefaults(metricBytes, err, "metric_json", lenscommon.PopulateTagcloudMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

	xAxisBytes, err := api.X.MarshalJSON()
	xv, ok := lenscommon.WrapNormalizedJSON(xAxisBytes, err, "x_axis", &diags)
	if !ok {
		return diags
	}
	m.XAxisJSON = xv

	if api.Y != nil {
		yAxisBytes, err := api.Y.MarshalJSON()
		yv, ok := lenscommon.WrapNormalizedJSON(yAxisBytes, err, "y_axis", &diags)
		if !ok {
			return diags
		}
		m.YAxisJSON = yv
	} else {
		m.YAxisJSON = jsontypes.NewNormalizedNull()
	}

	m.Query = &models.FilterSimpleModel{}
	lenscommon.FilterSimpleFromAPI(m.Query, api.Query)

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}

	return diags
}

func heatmapConfigFromAPIESQL(ctx context.Context, m *models.HeatmapConfigModel, prior *models.HeatmapConfigModel, api kbapi.KibanaHTTPAPIsHeatmapESQLByValuePanel) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	if !heatmapConfigPopulateCommonFields(m, api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling, datasetBytes, datasetErr, api.Filters, api.Axis, api.Styling, api.Legend, prior, &diags) {
		return diags
	}

	metricBytes, err := json.Marshal(api.Metric)
	mv, ok := lenscommon.MarshalToJSONWithDefaults(metricBytes, err, "metric_json", lenscommon.PopulateTagcloudMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

	xAxisBytes, err := json.Marshal(api.X)
	if err != nil {
		diags.AddError("Failed to marshal x_axis", err.Error())
		return diags
	}
	m.XAxisJSON = jsontypes.NewNormalizedValue(string(xAxisBytes))

	if api.Y != nil {
		yAxisBytes, err := json.Marshal(api.Y)
		if err != nil {
			diags.AddError("Failed to marshal y_axis", err.Error())
			return diags
		}
		m.YAxisJSON = jsontypes.NewNormalizedValue(string(yAxisBytes))
	} else {
		m.YAxisJSON = jsontypes.NewNormalizedNull()
	}

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}

	return diags
}

func heatmapConfigToAPI(m *models.HeatmapConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs lenscommon.VisByValueConfig0

	if m == nil {
		return attrs, diags
	}

	if lenscommon.ConfigUsesESQL(m.Query) {
		esql, esqlDiags := heatmapConfigToAPIESQL(m)
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromKibanaHTTPAPIsHeatmapESQLByValuePanel(esql); err != nil {
			diags.AddError("Failed to create heatmap ESQL schema", err.Error())
		}
		return attrs, diags
	}

	noESQL, noESQLDiags := heatmapConfigToAPINoESQL(m)
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromKibanaHTTPAPIsHeatmapNoESQLByValuePanel(noESQL); err != nil {
		diags.AddError("Failed to create heatmap schema", err.Error())
	}

	return attrs, diags
}

func heatmapConfigToAPINoESQL(m *models.HeatmapConfigModel) (kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanel{
		Type: kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanelTypeHeatmap,
	}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing dataset", "heatmap_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal heatmap_config.data_source_json", err.Error())
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
		var yAxis kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanel_Y
		if err := json.Unmarshal([]byte(m.YAxisJSON.ValueString()), &yAxis); err != nil {
			diags.AddError("Failed to unmarshal y_axis_json", err.Error())
			return api, diags
		}
		api.Y = &yAxis
	}

	if m.Axis == nil {
		diags.AddError("Missing axis", "heatmap_config.axis must be provided")
		return api, diags
	}
	axis, axisDiags := heatmapAxesToAPI(m.Axis)
	diags.Append(axisDiags...)
	if axis != nil && axis.X != nil {
		axis.X.Scale = typeutils.NonZero(axis.X.Scale, inferHeatmapXAxisScale(m.XAxisJSON.ValueString()))
	}
	api.Axis = axis

	if m.Styling == nil || m.Styling.Cells == nil {
		diags.AddError("Missing styling.cells", "heatmap_config.styling.cells must be provided")
		return api, diags
	}
	api.Styling = heatmapStylingToAPI(m.Styling)

	if m.Legend == nil {
		diags.AddError("Missing legend", "heatmap_config.legend must be provided")
		return api, diags
	}
	legend, legendDiags := heatmapLegendToAPI(m.Legend)
	diags.Append(legendDiags...)
	api.Legend = &legend

	if m.Query == nil {
		diags.AddError("Missing query", "heatmap_config.query must be provided for non-ES|QL heatmaps")
		return api, diags
	}
	api.Query = lenscommon.FilterSimpleToAPI(m.Query)

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsHeatmapNoESQLByValuePanel_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	return api, diags
}

func heatmapConfigToAPIESQL(m *models.HeatmapConfigModel) (kbapi.KibanaHTTPAPIsHeatmapESQLByValuePanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KibanaHTTPAPIsHeatmapESQLByValuePanel{
		Type: kbapi.KibanaHTTPAPIsHeatmapESQLByValuePanelTypeHeatmap,
	}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

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
		yAxis := new(struct {
			Column string                          `json:"column"`
			Format *kbapi.KibanaHTTPAPIsFormatType `json:"format,omitempty"`
			Label  *string                         `json:"label,omitempty"`
		})
		if err := json.Unmarshal([]byte(m.YAxisJSON.ValueString()), yAxis); err != nil {
			diags.AddError("Failed to unmarshal y_axis_json", err.Error())
			return api, diags
		}
		api.Y = yAxis
	}

	if m.Axis == nil {
		diags.AddError("Missing axis", "heatmap_config.axis must be provided")
		return api, diags
	}
	axis, axisDiags := heatmapAxesToAPI(m.Axis)
	diags.Append(axisDiags...)
	if axis != nil && axis.X != nil {
		axis.X.Scale = typeutils.NonZero(axis.X.Scale, inferHeatmapXAxisScale(m.XAxisJSON.ValueString()))
	}
	api.Axis = axis

	if m.Styling == nil || m.Styling.Cells == nil {
		diags.AddError("Missing styling.cells", "heatmap_config.styling.cells must be provided")
		return api, diags
	}
	api.Styling = heatmapStylingToAPI(m.Styling)

	if m.Legend == nil {
		diags.AddError("Missing legend", "heatmap_config.legend must be provided")
		return api, diags
	}
	legend, legendDiags := heatmapLegendToAPI(m.Legend)
	diags.Append(legendDiags...)
	api.Legend = &legend

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsHeatmapESQLByValuePanel_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	return api, diags
}

func heatmapAxesFromAPI(m *models.HeatmapAxesModel, api *kbapi.KibanaHTTPAPIsHeatmapAxes, prior *models.HeatmapAxesModel) diag.Diagnostics {
	diags := diag.Diagnostics{}
	// Kibana may omit `axis`, `axis.x`, or `axis.y` from a GET response when the
	// chart has no corresponding dimension. The kbapi spec types each as a
	// pointer with `omitempty`, so a nil here is "field absent" rather than
	// "field empty". Preserve the prior model in that case to avoid drift like
	// "axis.y was {...} but now null" after a write/read round trip.
	if api == nil {
		if prior != nil {
			m.X = prior.X
			m.Y = prior.Y
		}
		return diags
	}

	if api.X != nil {
		m.X = &models.HeatmapXAxisModel{}
		var priorX *models.HeatmapXAxisModel
		if prior != nil {
			priorX = prior.X
		}
		heatmapXAxisFromAPI(m.X, api.X, priorX)
	} else if prior != nil && prior.X != nil {
		m.X = prior.X
	}

	if api.Y != nil {
		m.Y = &models.HeatmapYAxisModel{}
		var priorY *models.HeatmapYAxisModel
		if prior != nil {
			priorY = prior.Y
		}
		heatmapYAxisFromAPI(m.Y, api.Y, priorY)
	} else if prior != nil && prior.Y != nil {
		m.Y = prior.Y
	}

	return diags
}

func heatmapAxesToAPI(m *models.HeatmapAxesModel) (*kbapi.KibanaHTTPAPIsHeatmapAxes, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m == nil {
		diags.AddError("Missing axis", "heatmap_config.axis must be provided")
		return nil, diags
	}

	axis := &kbapi.KibanaHTTPAPIsHeatmapAxes{}

	if m.X != nil {
		xAxis := heatmapXAxisToAPI(m.X)
		axis.X = &xAxis
	}
	if m.Y != nil {
		yAxis := heatmapYAxisToAPI(m.Y)
		axis.Y = &yAxis
	}

	return axis, diags
}

func heatmapXAxisFromAPI(m *models.HeatmapXAxisModel, api *kbapi.KibanaHTTPAPIsHeatmapXAxis, _ *models.HeatmapXAxisModel) {
	if api == nil {
		return
	}
	if api.Labels != nil {
		m.Labels = &models.HeatmapXAxisLabelsModel{}
		heatmapXAxisLabelsFromAPI(m.Labels, api.Labels)
	}
	if api.Title != nil {
		m.Title = &models.AxisTitleModel{}
		lenscommon.AxisTitleFromAPI(m.Title, api.Title)
	}
}

func heatmapXAxisToAPI(m *models.HeatmapXAxisModel) kbapi.KibanaHTTPAPIsHeatmapXAxis {
	axis := kbapi.KibanaHTTPAPIsHeatmapXAxis{}
	if m == nil {
		return axis
	}
	if m.Labels != nil {
		axis.Labels = heatmapXAxisLabelsToAPI(m.Labels)
	}
	if m.Title != nil {
		axis.Title = lenscommon.AxisTitleToAPI(m.Title)
	}
	return axis
}

func heatmapXAxisLabelsFromAPI(m *models.HeatmapXAxisLabelsModel, api *struct {
	Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	Visible     *bool                                  `json:"visible,omitempty"`
}) {
	if api == nil {
		return
	}
	if api.Orientation != nil {
		m.Orientation = types.StringValue(string(*api.Orientation))
	} else {
		m.Orientation = types.StringNull()
	}
	m.Visible = types.BoolPointerValue(api.Visible)
}

func heatmapXAxisLabelsToAPI(m *models.HeatmapXAxisLabelsModel) *struct {
	Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	Visible     *bool                                  `json:"visible,omitempty"`
} {
	if m == nil {
		return nil
	}
	labels := &struct {
		Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
		Visible     *bool                                  `json:"visible,omitempty"`
	}{}
	if typeutils.IsKnown(m.Orientation) {
		orientation := kbapi.KibanaHTTPAPIsVisApiOrientation(m.Orientation.ValueString())
		labels.Orientation = &orientation
	}
	if typeutils.IsKnown(m.Visible) {
		labels.Visible = new(m.Visible.ValueBool())
	}
	return labels
}

func heatmapYAxisFromAPI(m *models.HeatmapYAxisModel, api *kbapi.KibanaHTTPAPIsHeatmapYAxis, prior *models.HeatmapYAxisModel) {
	if api == nil {
		return
	}
	if api.Labels != nil {
		m.Labels = &models.HeatmapYAxisLabelsModel{}
		heatmapYAxisLabelsFromAPI(m.Labels, api.Labels)
	} else if prior != nil && prior.Labels != nil {
		// Kibana may omit Y-axis labels when there is no Y breakdown dimension.
		// Preserve the prior state to avoid a false drift.
		m.Labels = prior.Labels
	}
	if api.Title != nil {
		m.Title = &models.AxisTitleModel{}
		lenscommon.AxisTitleFromAPI(m.Title, api.Title)
	} else if prior != nil && prior.Title != nil {
		// Kibana may omit Y-axis title when there is no Y breakdown dimension.
		// Preserve the prior state to avoid a false drift.
		m.Title = prior.Title
	}
}

func heatmapYAxisToAPI(m *models.HeatmapYAxisModel) kbapi.KibanaHTTPAPIsHeatmapYAxis {
	axis := kbapi.KibanaHTTPAPIsHeatmapYAxis{}
	if m == nil {
		return axis
	}
	if m.Labels != nil {
		axis.Labels = heatmapYAxisLabelsToAPI(m.Labels)
	}
	if m.Title != nil {
		axis.Title = lenscommon.AxisTitleToAPI(m.Title)
	}
	return axis
}

func heatmapYAxisLabelsFromAPI(m *models.HeatmapYAxisLabelsModel, api *struct {
	Visible *bool `json:"visible,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Visible = types.BoolPointerValue(api.Visible)
}

func heatmapYAxisLabelsToAPI(m *models.HeatmapYAxisLabelsModel) *struct {
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

func heatmapCellsFromAPI(m *models.HeatmapCellsModel, api *kbapi.KibanaHTTPAPIsHeatmapCells) {
	if api == nil {
		return
	}
	if api.Labels != nil {
		m.Labels = &models.HeatmapCellsLabelsModel{}
		heatmapCellsLabelsFromAPI(m.Labels, api.Labels)
	}
}

func heatmapCellsToAPI(m *models.HeatmapCellsModel) kbapi.KibanaHTTPAPIsHeatmapCells {
	cells := kbapi.KibanaHTTPAPIsHeatmapCells{}
	if m == nil {
		return cells
	}
	if m.Labels != nil {
		cells.Labels = heatmapCellsLabelsToAPI(m.Labels)
	}
	return cells
}

func heatmapStylingFromAPI(m *models.HeatmapStylingModel, api *kbapi.KibanaHTTPAPIsHeatmapStyling) {
	if api == nil || api.Cells == nil {
		return
	}
	m.Cells = &models.HeatmapCellsModel{}
	heatmapCellsFromAPI(m.Cells, api.Cells)
}

func heatmapStylingToAPI(m *models.HeatmapStylingModel) *kbapi.KibanaHTTPAPIsHeatmapStyling {
	if m == nil || m.Cells == nil {
		return nil
	}
	cells := heatmapCellsToAPI(m.Cells)
	return &kbapi.KibanaHTTPAPIsHeatmapStyling{Cells: &cells}
}

func heatmapCellsLabelsFromAPI(m *models.HeatmapCellsLabelsModel, api *struct {
	Visible *bool `json:"visible,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Visible = types.BoolPointerValue(api.Visible)
}

func heatmapCellsLabelsToAPI(m *models.HeatmapCellsLabelsModel) *struct {
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

func heatmapLegendFromAPI(m *models.HeatmapLegendModel, api *kbapi.KibanaHTTPAPIsHeatmapLegend) {
	if api == nil {
		return
	}
	if api.Visibility != nil {
		m.Visibility = types.StringValue(string(*api.Visibility))
	} else {
		m.Visibility = types.StringNull()
	}
	if api.Size != nil {
		m.Size = types.StringValue(string(*api.Size))
	} else {
		m.Size = types.StringNull()
	}

	if api.TruncateAfterLines != nil {
		m.TruncateAfterLines = types.Int64Value(int64(*api.TruncateAfterLines))
	} else {
		m.TruncateAfterLines = types.Int64Null()
	}
}

func heatmapLegendToAPI(m *models.HeatmapLegendModel) (kbapi.KibanaHTTPAPIsHeatmapLegend, diag.Diagnostics) {
	var diags diag.Diagnostics
	legend := kbapi.KibanaHTTPAPIsHeatmapLegend{}

	if m == nil {
		diags.AddError("Missing legend", "heatmap_config.legend must be provided")
		return legend, diags
	}

	if typeutils.IsKnown(m.Visibility) {
		visibility := kbapi.KibanaHTTPAPIsHeatmapLegendVisibility(m.Visibility.ValueString())
		legend.Visibility = &visibility
	}
	if typeutils.IsKnown(m.Size) {
		size := kbapi.KibanaHTTPAPIsLegendSize(m.Size.ValueString())
		legend.Size = &size
	} else {
		diags.AddError("Missing legend size", "heatmap_config.legend.size must be provided")
	}
	if typeutils.IsKnown(m.TruncateAfterLines) {
		legend.TruncateAfterLines = new(float32(m.TruncateAfterLines.ValueInt64()))
	}

	return legend, diags
}
