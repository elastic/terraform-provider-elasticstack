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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func gaugeConfigUsesESQL(m *models.GaugeConfigModel) bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Expression.IsNull() && m.Query.Language.IsNull()
}

func gaugeConfigFromAPI(ctx context.Context, m *models.GaugeConfigModel, dashboard *models.DashboardModel, prior *models.GaugeConfigModel, api kbapi.GaugeNoESQL) diag.Diagnostics {
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

	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric", populateGaugeMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = preservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)
	m.EsqlMetric = nil

	m.Styling = &models.GaugeStylingModel{}
	if api.Styling.Shape != nil {
		shapeBytes, err := api.Styling.Shape.MarshalJSON()
		sv, ok := marshalToNormalized(shapeBytes, err, "shape", &diags)
		if !ok {
			return diags
		}
		m.Styling.ShapeJSON = sv
	} else {
		m.Styling.ShapeJSON = jsontypes.NewNormalizedNull()
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

func gaugeConfigFromAPIESQL(ctx context.Context, m *models.GaugeConfigModel, dashboard *models.DashboardModel, prior *models.GaugeConfigModel, api kbapi.GaugeESQL) diag.Diagnostics {
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
	m.MetricJSON = customtypes.NewJSONWithDefaultsNull(populateGaugeMetricDefaults)

	em := &models.GaugeEsqlMetric{
		Column: types.StringValue(api.Metric.Column),
	}
	formatVal, ok := lensESQLNumberFormatJSONFromAPI(api.Metric.Format, "esql_metric.format_json", &diags)
	if !ok {
		return diags
	}
	em.FormatJSON = formatVal

	if api.Metric.Label != nil {
		em.Label = types.StringValue(*api.Metric.Label)
	} else {
		em.Label = types.StringNull()
	}

	if api.Metric.Color != nil {
		colorBytes, cErr := api.Metric.Color.MarshalJSON()
		if cErr != nil {
			diags.AddError("Failed to marshal esql metric color", cErr.Error())
			return diags
		}
		if len(colorBytes) > 0 && string(colorBytes) != jsonNullString {
			em.ColorJSON = jsontypes.NewNormalizedValue(string(colorBytes))
		} else {
			em.ColorJSON = jsontypes.NewNormalizedNull()
		}
	} else {
		em.ColorJSON = jsontypes.NewNormalizedNull()
	}

	if api.Metric.Subtitle != nil {
		em.Subtitle = types.StringValue(*api.Metric.Subtitle)
	} else {
		em.Subtitle = types.StringNull()
	}

	if api.Metric.Goal != nil {
		em.Goal = &models.GaugeEsqlColumnRef{Column: types.StringValue(api.Metric.Goal.Column)}
		if api.Metric.Goal.Label != nil {
			em.Goal.Label = types.StringValue(*api.Metric.Goal.Label)
		} else {
			em.Goal.Label = types.StringNull()
		}
	}
	if api.Metric.Max != nil {
		em.Max = &models.GaugeEsqlColumnRef{Column: types.StringValue(api.Metric.Max.Column)}
		if api.Metric.Max.Label != nil {
			em.Max.Label = types.StringValue(*api.Metric.Max.Label)
		} else {
			em.Max.Label = types.StringNull()
		}
	}
	if api.Metric.Min != nil {
		em.Min = &models.GaugeEsqlColumnRef{Column: types.StringValue(api.Metric.Min.Column)}
		if api.Metric.Min.Label != nil {
			em.Min.Label = types.StringValue(*api.Metric.Min.Label)
		} else {
			em.Min.Label = types.StringNull()
		}
	}
	if api.Metric.Ticks != nil {
		em.Ticks = &models.GaugeEsqlTicks{}
		if api.Metric.Ticks.Mode != nil {
			em.Ticks.Mode = types.StringValue(string(*api.Metric.Ticks.Mode))
		} else {
			em.Ticks.Mode = types.StringNull()
		}
		if api.Metric.Ticks.Visible != nil {
			em.Ticks.Visible = types.BoolValue(*api.Metric.Ticks.Visible)
		} else {
			em.Ticks.Visible = types.BoolNull()
		}
	}
	if api.Metric.Title != nil {
		em.Title = &models.GaugeEsqlTitle{}
		if api.Metric.Title.Text != nil {
			em.Title.Text = types.StringValue(*api.Metric.Title.Text)
		} else {
			em.Title.Text = types.StringNull()
		}
		if api.Metric.Title.Visible != nil {
			em.Title.Visible = types.BoolValue(*api.Metric.Title.Visible)
		} else {
			em.Title.Visible = types.BoolNull()
		}
	}
	m.EsqlMetric = em

	m.Styling = &models.GaugeStylingModel{}
	if api.Styling.Shape != nil {
		shapeBytes, err := api.Styling.Shape.MarshalJSON()
		sv, ok := marshalToNormalized(shapeBytes, err, "shape", &diags)
		if !ok {
			return diags
		}
		m.Styling.ShapeJSON = sv
	} else {
		m.Styling.ShapeJSON = jsontypes.NewNormalizedNull()
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

func gaugeConfigToAPI(m *models.GaugeConfigModel, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	if gaugeConfigUsesESQL(m) {
		esql, d := gaugeConfigToAPIESQL(m, dashboard)
		diags.Append(d...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromGaugeESQL(esql); err != nil {
			diags.AddError("Failed to create gauge ES|QL attributes", err.Error())
		}
		return attrs, diags
	}

	noESQL, d := gaugeConfigToAPINoESQL(m, dashboard)
	diags.Append(d...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromGaugeNoESQL(noESQL); err != nil {
		diags.AddError("Failed to create gauge attributes", err.Error())
	}
	return attrs, diags
}

func gaugeConfigToAPINoESQL(m *models.GaugeConfigModel, dashboard *models.DashboardModel) (kbapi.GaugeNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.GaugeNoESQL

	api.Type = kbapi.GaugeNoESQLTypeGauge

	if !m.Title.IsNull() {
		api.Title = m.Title.ValueStringPointer()
	}

	if !m.Description.IsNull() {
		api.Description = m.Description.ValueStringPointer()
	}

	if typeutils.IsKnown(m.DataSourceJSON) {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal gauge_config.data_source_json", err.Error())
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
		diags.AddError("Missing query", "gauge_config.query must be set for non-ES|QL gauges (or omit `query` entirely for ES|QL mode)")
		return api, diags
	}
	api.Query = filterSimpleToAPI(m.Query)

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if m.MetricJSON.IsNull() {
		diags.AddError("Missing metric_json", "gauge_config.metric_json must be set for non-ES|QL gauges (or use `esql_metric` in ES|QL mode)")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
		diags.AddError("Failed to unmarshal metric", err.Error())
		return api, diags
	}

	if m.Styling != nil && typeutils.IsKnown(m.Styling.ShapeJSON) {
		var shape kbapi.GaugeStyling_Shape
		shapeDiags := m.Styling.ShapeJSON.Unmarshal(&shape)
		diags.Append(shapeDiags...)
		if !shapeDiags.HasError() {
			api.Styling.Shape = &shape
		}
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
		items, ddDiags := decodeLensDrilldownSlice[kbapi.GaugeNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}

func gaugeConfigToAPIESQL(m *models.GaugeConfigModel, dashboard *models.DashboardModel) (kbapi.GaugeESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.GaugeESQL
	api.Type = kbapi.GaugeESQLTypeGauge

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
		diags.AddError("Missing data_source_json", "gauge_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal gauge_config.data_source_json", err.Error())
		return api, diags
	}

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if m.EsqlMetric == nil {
		diags.AddError("Missing esql_metric", "gauge_config.esql_metric must be set in ES|QL mode")
		return api, diags
	}
	api.Metric.Column = m.EsqlMetric.Column.ValueString()
	if err := json.Unmarshal([]byte(m.EsqlMetric.FormatJSON.ValueString()), &api.Metric.Format); err != nil {
		diags.AddError("Failed to unmarshal esql_metric.format_json", err.Error())
		return api, diags
	}
	if typeutils.IsKnown(m.EsqlMetric.Label) {
		l := m.EsqlMetric.Label.ValueString()
		api.Metric.Label = &l
	}
	if typeutils.IsKnown(m.EsqlMetric.ColorJSON) {
		var color kbapi.GaugeESQL_Metric_Color
		if err := json.Unmarshal([]byte(m.EsqlMetric.ColorJSON.ValueString()), &color); err != nil {
			diags.AddError("Failed to unmarshal esql_metric.color_json", err.Error())
			return api, diags
		}
		api.Metric.Color = &color
	}
	if typeutils.IsKnown(m.EsqlMetric.Subtitle) {
		s := m.EsqlMetric.Subtitle.ValueString()
		api.Metric.Subtitle = &s
	}
	if m.EsqlMetric.Goal != nil {
		api.Metric.Goal = &struct {
			Column string  `json:"column"`
			Label  *string `json:"label,omitempty"`
		}{Column: m.EsqlMetric.Goal.Column.ValueString()}
		if typeutils.IsKnown(m.EsqlMetric.Goal.Label) {
			l := m.EsqlMetric.Goal.Label.ValueString()
			api.Metric.Goal.Label = &l
		}
	}
	if m.EsqlMetric.Max != nil {
		api.Metric.Max = &struct {
			Column string  `json:"column"`
			Label  *string `json:"label,omitempty"`
		}{Column: m.EsqlMetric.Max.Column.ValueString()}
		if typeutils.IsKnown(m.EsqlMetric.Max.Label) {
			l := m.EsqlMetric.Max.Label.ValueString()
			api.Metric.Max.Label = &l
		}
	}
	if m.EsqlMetric.Min != nil {
		api.Metric.Min = &struct {
			Column string  `json:"column"`
			Label  *string `json:"label,omitempty"`
		}{Column: m.EsqlMetric.Min.Column.ValueString()}
		if typeutils.IsKnown(m.EsqlMetric.Min.Label) {
			l := m.EsqlMetric.Min.Label.ValueString()
			api.Metric.Min.Label = &l
		}
	}
	if m.EsqlMetric.Ticks != nil {
		api.Metric.Ticks = &struct {
			Mode    *kbapi.GaugeESQLMetricTicksMode `json:"mode,omitempty"`
			Visible *bool                           `json:"visible,omitempty"`
		}{}
		if typeutils.IsKnown(m.EsqlMetric.Ticks.Mode) {
			mode := kbapi.GaugeESQLMetricTicksMode(m.EsqlMetric.Ticks.Mode.ValueString())
			api.Metric.Ticks.Mode = &mode
		}
		if typeutils.IsKnown(m.EsqlMetric.Ticks.Visible) {
			v := m.EsqlMetric.Ticks.Visible.ValueBool()
			api.Metric.Ticks.Visible = &v
		}
	}
	if m.EsqlMetric.Title != nil {
		api.Metric.Title = &struct {
			Text    *string `json:"text,omitempty"`
			Visible *bool   `json:"visible,omitempty"`
		}{}
		if typeutils.IsKnown(m.EsqlMetric.Title.Text) {
			t := m.EsqlMetric.Title.Text.ValueString()
			api.Metric.Title.Text = &t
		}
		if typeutils.IsKnown(m.EsqlMetric.Title.Visible) {
			v := m.EsqlMetric.Title.Visible.ValueBool()
			api.Metric.Title.Visible = &v
		}
	}

	if m.Styling != nil && typeutils.IsKnown(m.Styling.ShapeJSON) {
		var shape kbapi.GaugeStyling_Shape
		shapeDiags := m.Styling.ShapeJSON.Unmarshal(&shape)
		diags.Append(shapeDiags...)
		if !shapeDiags.HasError() {
			api.Styling.Shape = &shape
		}
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
		items, ddDiags := decodeLensDrilldownSlice[kbapi.GaugeESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}
