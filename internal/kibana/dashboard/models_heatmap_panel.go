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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newHeatmapPanelConfigConverter() heatmapPanelConfigConverter {
	return heatmapPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.HeatmapNoESQLTypeHeatmap),
			hasTFChartBlock: func(blocks *models.LensByValueChartBlocks) bool {
				return blocks != nil && blocks.HeatmapConfig != nil
			},
		},
	}
}

type heatmapPanelConfigConverter struct {
	lensVisualizationBase
}

func (c heatmapPanelConfigConverter) populateFromAttributes(
	ctx context.Context,
	dashboard *models.DashboardModel,
	tfPanel *models.PanelModel,
	blocks *models.LensByValueChartBlocks,
	attrs kbapi.KbnDashboardPanelTypeVisConfig0,
) diag.Diagnostics {
	var prior *models.HeatmapConfigModel
	if b := lensByValueChartBlocksFromPanel(tfPanel); b != nil && b.HeatmapConfig != nil {
		cpy := *b.HeatmapConfig
		prior = &cpy
	}
	blocks.HeatmapConfig = &models.HeatmapConfigModel{}
	if heatmapNoESQL, err := attrs.AsHeatmapNoESQL(); err == nil && !isHeatmapNoESQLCandidateActuallyESQL(heatmapNoESQL) {
		return heatmapConfigFromAPINoESQL(ctx, blocks.HeatmapConfig, dashboard, prior, heatmapNoESQL)
	}
	heatmapESQL, err := attrs.AsHeatmapESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return heatmapConfigFromAPIESQL(ctx, blocks.HeatmapConfig, dashboard, prior, heatmapESQL)
}

func (c heatmapPanelConfigConverter) buildAttributes(blocks *models.LensByValueChartBlocks, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *blocks.HeatmapConfig

	attrs, heatmapDiags := heatmapConfigToAPI(&configModel, dashboard)
	diags.Append(heatmapDiags...)
	if diags.HasError() {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	return attrs, diags
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

func heatmapConfigPopulateCommonFields(m *models.HeatmapConfigModel,
	title, description *string,
	ignoreGlobalFilters *bool,
	sampling *float32,
	datasetBytes []byte,
	datasetErr error,
	filters []kbapi.LensPanelFilters_Item,
	axis kbapi.HeatmapAxes,
	styling kbapi.HeatmapStyling,
	legend kbapi.HeatmapLegend,
	prior *models.HeatmapConfigModel,
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
	m.Axis = &models.HeatmapAxesModel{}
	var priorAxis *models.HeatmapAxesModel
	if prior != nil {
		priorAxis = prior.Axis
	}
	axisDiags := heatmapAxesFromAPI(m.Axis, axis, priorAxis)
	diags.Append(axisDiags...)
	m.Styling = &models.HeatmapStylingModel{}
	heatmapStylingFromAPI(m.Styling, styling)
	m.Legend = &models.HeatmapLegendModel{}
	heatmapLegendFromAPI(m.Legend, legend)
	return !diags.HasError()
}

func heatmapConfigFromAPINoESQL(ctx context.Context, m *models.HeatmapConfigModel, dashboard *models.DashboardModel, prior *models.HeatmapConfigModel, api kbapi.HeatmapNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	if !heatmapConfigPopulateCommonFields(m, api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling, datasetBytes, datasetErr, api.Filters, api.Axis, api.Styling, api.Legend, prior, &diags) {
		return diags
	}

	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric_json", populateTagcloudMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = preservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

	xAxisBytes, err := api.X.MarshalJSON()
	xv, ok := marshalToNormalized(xAxisBytes, err, "x_axis", &diags)
	if !ok {
		return diags
	}
	m.XAxisJSON = xv

	if api.Y != nil {
		yAxisBytes, err := api.Y.MarshalJSON()
		yv, ok := marshalToNormalized(yAxisBytes, err, "y_axis", &diags)
		if !ok {
			return diags
		}
		m.YAxisJSON = yv
	} else {
		m.YAxisJSON = jsontypes.NewNormalizedNull()
	}

	m.Query = &models.FilterSimpleModel{}
	filterSimpleFromAPI(m.Query, api.Query)

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lensDrilldownsAPIToWire(api.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lensChartPresentationReadsFor(ctx, dashboard, priorLens, api.TimeRange, api.HideTitle, api.HideBorder, api.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres

	return diags
}

func heatmapConfigFromAPIESQL(ctx context.Context, m *models.HeatmapConfigModel, dashboard *models.DashboardModel, prior *models.HeatmapConfigModel, api kbapi.HeatmapESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	if !heatmapConfigPopulateCommonFields(m, api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling, datasetBytes, datasetErr, api.Filters, api.Axis, api.Styling, api.Legend, prior, &diags) {
		return diags
	}

	metricBytes, err := json.Marshal(api.Metric)
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric_json", populateTagcloudMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = preservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

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

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lensDrilldownsAPIToWire(api.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lensChartPresentationReadsFor(ctx, dashboard, priorLens, api.TimeRange, api.HideTitle, api.HideBorder, api.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres

	return diags
}

func heatmapConfigToAPI(m *models.HeatmapConfigModel, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0

	if m == nil {
		return attrs, diags
	}

	if heatmapConfigUsesESQL(m) {
		esql, esqlDiags := heatmapConfigToAPIESQL(m, dashboard)
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromHeatmapESQL(esql); err != nil {
			diags.AddError("Failed to create heatmap ESQL schema", err.Error())
		}
		return attrs, diags
	}

	noESQL, noESQLDiags := heatmapConfigToAPINoESQL(m, dashboard)
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromHeatmapNoESQL(noESQL); err != nil {
		diags.AddError("Failed to create heatmap schema", err.Error())
	}

	return attrs, diags
}

func heatmapConfigUsesESQL(m *models.HeatmapConfigModel) bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Expression.IsNull() && m.Query.Language.IsNull()
}

func heatmapConfigToAPINoESQL(m *models.HeatmapConfigModel, dashboard *models.DashboardModel) (kbapi.HeatmapNoESQL, diag.Diagnostics) {
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
		var yAxis kbapi.HeatmapNoESQL_Y
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
	if axis.X.Scale == "" {
		axis.X.Scale = inferHeatmapXAxisScale(m.XAxisJSON.ValueString())
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
	api.Legend = legend

	if m.Query == nil {
		diags.AddError("Missing query", "heatmap_config.query must be provided for non-ES|QL heatmaps")
		return api, diags
	}
	api.Query = filterSimpleToAPI(m.Query)

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	api.TimeRange = writes.TimeRange
	if writes.HideTitle != nil {
		api.HideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		api.HideBorder = writes.HideBorder
	}
	if writes.References != nil {
		api.References = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := decodeLensDrilldownSlice[kbapi.HeatmapNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}

func heatmapConfigToAPIESQL(m *models.HeatmapConfigModel, dashboard *models.DashboardModel) (kbapi.HeatmapESQL, diag.Diagnostics) {
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
		yAxis := new(struct {
			Column string           `json:"column"`
			Format kbapi.FormatType `json:"format"`
			Label  *string          `json:"label,omitempty"`
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
	if axis.X.Scale == "" {
		axis.X.Scale = inferHeatmapXAxisScale(m.XAxisJSON.ValueString())
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
	api.Legend = legend

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	api.TimeRange = writes.TimeRange
	if writes.HideTitle != nil {
		api.HideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		api.HideBorder = writes.HideBorder
	}
	if writes.References != nil {
		api.References = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := decodeLensDrilldownSlice[kbapi.HeatmapESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}

func heatmapAxesFromAPI(m *models.HeatmapAxesModel, api kbapi.HeatmapAxes, prior *models.HeatmapAxesModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	m.X = &models.HeatmapXAxisModel{}
	var priorX *models.HeatmapXAxisModel
	if prior != nil {
		priorX = prior.X
	}
	heatmapXAxisFromAPI(m.X, api.X, priorX)

	m.Y = &models.HeatmapYAxisModel{}
	var priorY *models.HeatmapYAxisModel
	if prior != nil {
		priorY = prior.Y
	}
	heatmapYAxisFromAPI(m.Y, api.Y, priorY)

	return diags
}

func heatmapAxesToAPI(m *models.HeatmapAxesModel) (kbapi.HeatmapAxes, diag.Diagnostics) {
	var diags diag.Diagnostics
	axis := kbapi.HeatmapAxes{}

	if m == nil {
		diags.AddError("Missing axis", "heatmap_config.axis must be provided")
		return axis, diags
	}

	if m.X != nil {
		axis.X = heatmapXAxisToAPI(m.X)
	}
	if m.Y != nil {
		axis.Y = heatmapYAxisToAPI(m.Y)
	}

	return axis, diags
}

func heatmapXAxisFromAPI(m *models.HeatmapXAxisModel, api kbapi.HeatmapXAxis, _ *models.HeatmapXAxisModel) {
	if api.Labels != nil {
		m.Labels = &models.HeatmapXAxisLabelsModel{}
		heatmapXAxisLabelsFromAPI(m.Labels, api.Labels)
	}
	if api.Title != nil {
		m.Title = &models.AxisTitleModel{}
		axisTitleFromAPI(m.Title, api.Title)
	}
}

func heatmapXAxisToAPI(m *models.HeatmapXAxisModel) kbapi.HeatmapXAxis {
	axis := kbapi.HeatmapXAxis{}
	if m == nil {
		return axis
	}
	if m.Labels != nil {
		axis.Labels = heatmapXAxisLabelsToAPI(m.Labels)
	}
	if m.Title != nil {
		axis.Title = axisTitleToAPI(m.Title)
	}
	return axis
}

func heatmapXAxisLabelsFromAPI(m *models.HeatmapXAxisLabelsModel, api *struct {
	Orientation kbapi.VisApiOrientation `json:"orientation"`
	Visible     *bool                   `json:"visible,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Orientation = types.StringValue(string(api.Orientation))
	m.Visible = types.BoolPointerValue(api.Visible)
}

func heatmapXAxisLabelsToAPI(m *models.HeatmapXAxisLabelsModel) *struct {
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

func heatmapYAxisFromAPI(m *models.HeatmapYAxisModel, api kbapi.HeatmapYAxis, prior *models.HeatmapYAxisModel) {
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
		axisTitleFromAPI(m.Title, api.Title)
	} else if prior != nil && prior.Title != nil {
		// Kibana may omit Y-axis title when there is no Y breakdown dimension.
		// Preserve the prior state to avoid a false drift.
		m.Title = prior.Title
	}
}

func heatmapYAxisToAPI(m *models.HeatmapYAxisModel) kbapi.HeatmapYAxis {
	axis := kbapi.HeatmapYAxis{}
	if m == nil {
		return axis
	}
	if m.Labels != nil {
		axis.Labels = heatmapYAxisLabelsToAPI(m.Labels)
	}
	if m.Title != nil {
		axis.Title = axisTitleToAPI(m.Title)
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

func heatmapCellsFromAPI(m *models.HeatmapCellsModel, api kbapi.HeatmapCells) {
	if api.Labels != nil {
		m.Labels = &models.HeatmapCellsLabelsModel{}
		heatmapCellsLabelsFromAPI(m.Labels, api.Labels)
	}
}

func heatmapCellsToAPI(m *models.HeatmapCellsModel) kbapi.HeatmapCells {
	cells := kbapi.HeatmapCells{}
	if m == nil {
		return cells
	}
	if m.Labels != nil {
		cells.Labels = heatmapCellsLabelsToAPI(m.Labels)
	}
	return cells
}

func heatmapStylingFromAPI(m *models.HeatmapStylingModel, api kbapi.HeatmapStyling) {
	m.Cells = &models.HeatmapCellsModel{}
	heatmapCellsFromAPI(m.Cells, api.Cells)
}

func heatmapStylingToAPI(m *models.HeatmapStylingModel) kbapi.HeatmapStyling {
	styling := kbapi.HeatmapStyling{}
	if m == nil || m.Cells == nil {
		return styling
	}
	styling.Cells = heatmapCellsToAPI(m.Cells)
	return styling
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

func heatmapLegendFromAPI(m *models.HeatmapLegendModel, api kbapi.HeatmapLegend) {
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

func heatmapLegendToAPI(m *models.HeatmapLegendModel) (kbapi.HeatmapLegend, diag.Diagnostics) {
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
