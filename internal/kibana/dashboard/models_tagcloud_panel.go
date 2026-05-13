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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newTagcloudPanelConfigConverter() tagcloudPanelConfigConverter {
	return tagcloudPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.TagcloudNoESQLTypeTagCloud),
			hasTFChartBlock: func(blocks *models.LensByValueChartBlocks) bool {
				return blocks != nil && blocks.TagcloudConfig != nil
			},
		},
	}
}

type tagcloudPanelConfigConverter struct {
	lensVisualizationBase
}

func (c tagcloudPanelConfigConverter) populateFromAttributes(
	ctx context.Context,
	dashboard *models.DashboardModel,
	tfPanel *models.PanelModel,
	blocks *models.LensByValueChartBlocks,
	attrs kbapi.KbnDashboardPanelTypeVisConfig0,
) diag.Diagnostics {
	var prior *models.TagcloudConfigModel
	if b := lensByValueChartBlocksFromPanel(tfPanel); b != nil && b.TagcloudConfig != nil {
		cpy := *b.TagcloudConfig
		prior = &cpy
	}
	blocks.TagcloudConfig = &models.TagcloudConfigModel{}

	if noESQL, err := attrs.AsTagcloudNoESQL(); err == nil && !isTagcloudNoESQLCandidateActuallyESQL(noESQL) {
		return tagcloudConfigFromAPI(ctx, blocks.TagcloudConfig, dashboard, prior, noESQL)
	}

	tagcloudESQL, err := attrs.AsTagcloudESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return tagcloudConfigFromAPIESQL(ctx, blocks.TagcloudConfig, dashboard, prior, tagcloudESQL)
}

func (c tagcloudPanelConfigConverter) buildAttributes(blocks *models.LensByValueChartBlocks, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	configModel := *blocks.TagcloudConfig
	return tagcloudConfigToAPI(&configModel, dashboard)
}

func isTagcloudNoESQLCandidateActuallyESQL(api kbapi.TagcloudNoESQL) bool {
	return lensDataSourceIsESQLOrTable(api.DataSource.MarshalJSON())
}

// applyStylingFromAPI populates the typed `orientation` and `font_size`
// attributes from a TagcloudStyling payload. Used by both NoESQL and ES|QL
// reads so the two paths stay in lockstep.
func tagcloudConfigApplyStylingFromAPI(m *models.TagcloudConfigModel, s kbapi.TagcloudStyling) {
	if s.Orientation != "" {
		m.Orientation = types.StringValue(string(s.Orientation))
	} else {
		m.Orientation = types.StringNull()
	}
	if s.FontSize == nil {
		m.FontSize = nil
		return
	}
	m.FontSize = &models.FontSizeModel{}
	if s.FontSize.Min != nil {
		m.FontSize.Min = types.Float64Value(float64(*s.FontSize.Min))
	} else {
		m.FontSize.Min = types.Float64Null()
	}
	if s.FontSize.Max != nil {
		m.FontSize.Max = types.Float64Value(float64(*s.FontSize.Max))
	} else {
		m.FontSize.Max = types.Float64Null()
	}
}

func tagcloudConfigUsesESQL(m *models.TagcloudConfigModel) bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Expression.IsNull() && m.Query.Language.IsNull()
}

func tagcloudConfigFromAPI(ctx context.Context, m *models.TagcloudConfigModel, dashboard *models.DashboardModel, prior *models.TagcloudConfigModel, api kbapi.TagcloudNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

	datasetBytes, err := api.DataSource.MarshalJSON()
	v, ok := marshalToNormalized(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = v

	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	m.Query = &models.FilterSimpleModel{}
	filterSimpleFromAPI(m.Query, api.Query)

	m.Filters = populateFiltersFromAPI(api.Filters, &diags)
	tagcloudConfigApplyStylingFromAPI(m, api.Styling)

	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric", populateTagcloudMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = preservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

	tagByBytes, err := api.TagBy.MarshalJSON()
	tv, ok := marshalToJSONWithDefaults(tagByBytes, err, "tag_by", populateTagcloudTagByDefaults, &diags)
	if !ok {
		return diags
	}
	m.TagByJSON = preservePriorJSONWithDefaultsIfEquivalent(ctx, m.TagByJSON, tv, &diags)

	m.EsqlMetric = nil
	m.EsqlTagBy = nil

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

func tagcloudConfigFromAPIESQL(ctx context.Context, m *models.TagcloudConfigModel, dashboard *models.DashboardModel, prior *models.TagcloudConfigModel, api kbapi.TagcloudESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	datasetBytes, err := json.Marshal(api.DataSource)
	dv, ok := marshalToNormalized(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = dv

	m.Query = nil
	m.Filters = populateFiltersFromAPI(api.Filters, &diags)
	tagcloudConfigApplyStylingFromAPI(m, api.Styling)

	m.MetricJSON = customtypes.NewJSONWithDefaultsNull(populateTagcloudMetricDefaults)
	m.TagByJSON = customtypes.NewJSONWithDefaultsNull(populateTagcloudTagByDefaults)

	em := &models.TagcloudEsqlMetric{
		Column: types.StringValue(api.Metric.Column),
	}
	metricFormat, ok := lensESQLNumberFormatJSONFromAPI(api.Metric.Format, "esql_metric.format_json", &diags)
	if !ok {
		return diags
	}
	em.FormatJSON = metricFormat
	if api.Metric.Label != nil {
		em.Label = types.StringValue(*api.Metric.Label)
	} else {
		em.Label = types.StringNull()
	}
	m.EsqlMetric = em

	tb := &models.TagcloudEsqlTagBy{
		Column: types.StringValue(api.TagBy.Column),
	}
	tagByFormat, ok := lensESQLNumberFormatJSONFromAPI(api.TagBy.Format, "esql_tag_by.format_json", &diags)
	if !ok {
		return diags
	}
	tb.FormatJSON = tagByFormat
	colorBytes, cErr := json.Marshal(api.TagBy.Color)
	if cErr != nil {
		diags.AddError("Failed to marshal esql tag_by color", cErr.Error())
		return diags
	}
	tb.ColorJSON = jsontypes.NewNormalizedValue(string(colorBytes))
	if api.TagBy.Label != nil {
		tb.Label = types.StringValue(*api.TagBy.Label)
	} else {
		tb.Label = types.StringNull()
	}
	m.EsqlTagBy = tb

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

func tagcloudConfigToAPI(m *models.TagcloudConfigModel, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	if tagcloudConfigUsesESQL(m) {
		esql, d := tagcloudConfigToAPIESQL(m, dashboard)
		diags.Append(d...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromTagcloudESQL(esql); err != nil {
			diags.AddError("Failed to create tagcloud ES|QL attributes", err.Error())
		}
		return attrs, diags
	}

	noESQL, d := tagcloudConfigToAPINoESQL(m, dashboard)
	diags.Append(d...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromTagcloudNoESQL(noESQL); err != nil {
		diags.AddError("Failed to create tagcloud attributes", err.Error())
	}
	return attrs, diags
}

func tagcloudConfigToAPINoESQL(m *models.TagcloudConfigModel, dashboard *models.DashboardModel) (kbapi.TagcloudNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.TagcloudNoESQL

	api.Type = kbapi.TagcloudNoESQLTypeTagCloud

	if !m.Title.IsNull() {
		api.Title = m.Title.ValueStringPointer()
	}

	if !m.Description.IsNull() {
		api.Description = m.Description.ValueStringPointer()
	}

	if !m.DataSourceJSON.IsNull() {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal tagcloud_config.data_source_json", err.Error())
			return api, diags
		}
	}

	if !m.IgnoreGlobalFilters.IsNull() {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}

	if !m.Sampling.IsNull() {
		sampling := float32(m.Sampling.ValueFloat64())
		api.Sampling = &sampling
	}

	if m.Query == nil {
		diags.AddError("Missing query", "tagcloud_config.query must be set for non-ES|QL tagclouds (or omit `query` entirely for ES|QL mode)")
		return api, diags
	}
	api.Query = filterSimpleToAPI(m.Query)

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if !m.Orientation.IsNull() {
		api.Styling.Orientation = kbapi.VisApiOrientation(m.Orientation.ValueString())
	}

	if m.FontSize != nil {
		fontSize := struct {
			Max *float32 `json:"max,omitempty"`
			Min *float32 `json:"min,omitempty"`
		}{}
		if !m.FontSize.Min.IsNull() {
			minValue := float32(m.FontSize.Min.ValueFloat64())
			fontSize.Min = &minValue
		}
		if !m.FontSize.Max.IsNull() {
			maxValue := float32(m.FontSize.Max.ValueFloat64())
			fontSize.Max = &maxValue
		}
		api.Styling.FontSize = &fontSize
	}

	if m.MetricJSON.IsNull() {
		diags.AddError("Missing metric_json", "tagcloud_config.metric_json must be set for non-ES|QL tagclouds (or use `esql_metric` in ES|QL mode)")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
		diags.AddError("Failed to unmarshal metric", err.Error())
		return api, diags
	}

	if m.TagByJSON.IsNull() {
		diags.AddError("Missing tag_by_json", "tagcloud_config.tag_by_json must be set for non-ES|QL tagclouds (or use `esql_tag_by` in ES|QL mode)")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.TagByJSON.ValueString()), &api.TagBy); err != nil {
		diags.AddError("Failed to unmarshal tag_by", err.Error())
		return api, diags
	}

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
		items, ddDiags := decodeLensDrilldownSlice[kbapi.TagcloudNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}

func tagcloudConfigToAPIESQL(m *models.TagcloudConfigModel, dashboard *models.DashboardModel) (kbapi.TagcloudESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.TagcloudESQL
	api.Type = kbapi.TagcloudESQLTypeTagCloud

	if typeutils.IsKnown(m.Title) {
		api.Title = m.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = m.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.Sampling) {
		s := float32(m.Sampling.ValueFloat64())
		api.Sampling = &s
	}

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing data_source_json", "tagcloud_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal tagcloud_config.data_source_json", err.Error())
		return api, diags
	}

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if !m.Orientation.IsNull() {
		api.Styling.Orientation = kbapi.VisApiOrientation(m.Orientation.ValueString())
	}
	if m.FontSize != nil {
		fontSize := struct {
			Max *float32 `json:"max,omitempty"`
			Min *float32 `json:"min,omitempty"`
		}{}
		if !m.FontSize.Min.IsNull() {
			minValue := float32(m.FontSize.Min.ValueFloat64())
			fontSize.Min = &minValue
		}
		if !m.FontSize.Max.IsNull() {
			maxValue := float32(m.FontSize.Max.ValueFloat64())
			fontSize.Max = &maxValue
		}
		api.Styling.FontSize = &fontSize
	}

	if m.EsqlMetric == nil {
		diags.AddError("Missing esql_metric", "tagcloud_config.esql_metric must be set in ES|QL mode")
		return api, diags
	}
	api.Metric.Column = m.EsqlMetric.Column.ValueString()
	if err := json.Unmarshal([]byte(m.EsqlMetric.FormatJSON.ValueString()), &api.Metric.Format); err != nil {
		diags.AddError("Failed to unmarshal esql_metric.format_json", err.Error())
		return api, diags
	}
	if typeutils.IsKnown(m.EsqlMetric.Label) {
		s := m.EsqlMetric.Label.ValueString()
		api.Metric.Label = &s
	}

	if m.EsqlTagBy == nil {
		diags.AddError("Missing esql_tag_by", "tagcloud_config.esql_tag_by must be set in ES|QL mode")
		return api, diags
	}
	api.TagBy.Column = m.EsqlTagBy.Column.ValueString()
	if err := json.Unmarshal([]byte(m.EsqlTagBy.FormatJSON.ValueString()), &api.TagBy.Format); err != nil {
		diags.AddError("Failed to unmarshal esql_tag_by.format_json", err.Error())
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.EsqlTagBy.ColorJSON.ValueString()), &api.TagBy.Color); err != nil {
		diags.AddError("Failed to unmarshal esql_tag_by.color_json", err.Error())
		return api, diags
	}
	if typeutils.IsKnown(m.EsqlTagBy.Label) {
		s := m.EsqlTagBy.Label.ValueString()
		api.TagBy.Label = &s
	}

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
		items, ddDiags := decodeLensDrilldownSlice[kbapi.TagcloudESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}
