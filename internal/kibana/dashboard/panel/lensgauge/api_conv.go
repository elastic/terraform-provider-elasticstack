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

package lensgauge

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const jsonNullString = "null"

func gaugeConfigFromAPI(ctx context.Context, m *models.GaugeConfigModel, prior *models.GaugeConfigModel, api kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanel) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

	datasetBytes, err := api.DataSource.MarshalJSON()
	v, ok := lenscommon.WrapNormalizedJSON(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = v

	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	m.Sampling = typeutils.Float32PointerToFloat64Value(api.Sampling)

	m.Query = &models.FilterSimpleModel{}
	lenscommon.FilterSimpleFromAPI(m.Query, api.Query)

	m.Filters = lenscommon.PopulateFiltersFromAPI(api.Filters, &diags)

	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := lenscommon.MarshalToJSONWithDefaults(metricBytes, err, "metric", lenscommon.PopulateGaugeMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)
	m.EsqlMetric = nil

	m.Styling = &models.GaugeStylingModel{}
	if api.Styling != nil && api.Styling.Shape != nil {
		shapeBytes, err := api.Styling.Shape.MarshalJSON()
		sv, ok := lenscommon.WrapNormalizedJSON(shapeBytes, err, "shape", &diags)
		if !ok {
			return diags
		}
		m.Styling.ShapeJSON = sv
	} else {
		m.Styling.ShapeJSON = jsontypes.NewNormalizedNull()
	}

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}

	return diags
}

func gaugeConfigFromAPIESQL(ctx context.Context, m *models.GaugeConfigModel, prior *models.GaugeConfigModel, api kbapi.KibanaHTTPAPIsGaugeESQLByValuePanel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	m.Sampling = typeutils.Float32PointerToFloat64Value(api.Sampling)

	datasetBytes, err := json.Marshal(api.DataSource)
	dv, ok := lenscommon.WrapNormalizedJSON(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = dv

	m.Query = nil
	m.Filters = lenscommon.PopulateFiltersFromAPI(api.Filters, &diags)
	m.MetricJSON = customtypes.NewJSONWithDefaultsNull(lenscommon.PopulateGaugeMetricDefaults)

	em := &models.GaugeEsqlMetric{
		Column: types.StringValue(api.Metric.Column),
	}
	formatVal, ok := lenscommon.LensESQLNumberFormatJSONFromAPI(api.Metric.Format, "esql_metric.format_json", &diags)
	if !ok {
		return diags
	}
	em.FormatJSON = formatVal

	em.Label = typeutils.StringishPointerValue(api.Metric.Label)

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

	em.Subtitle = typeutils.StringishPointerValue(api.Metric.Subtitle)

	if api.Metric.Goal != nil {
		em.Goal = &models.GaugeEsqlColumnRef{Column: types.StringValue(api.Metric.Goal.Column)}
		em.Goal.Label = typeutils.StringishPointerValue(api.Metric.Goal.Label)
	}
	if api.Metric.Max != nil {
		em.Max = &models.GaugeEsqlColumnRef{Column: types.StringValue(api.Metric.Max.Column)}
		em.Max.Label = typeutils.StringishPointerValue(api.Metric.Max.Label)
	}
	if api.Metric.Min != nil {
		em.Min = &models.GaugeEsqlColumnRef{Column: types.StringValue(api.Metric.Min.Column)}
		em.Min.Label = typeutils.StringishPointerValue(api.Metric.Min.Label)
	}
	if api.Metric.Ticks != nil {
		em.Ticks = &models.GaugeEsqlTicks{}
		if api.Metric.Ticks.Mode != nil {
			em.Ticks.Mode = types.StringValue(string(*api.Metric.Ticks.Mode))
		} else {
			em.Ticks.Mode = types.StringNull()
		}
		em.Ticks.Visible = types.BoolPointerValue(api.Metric.Ticks.Visible)
	}
	if api.Metric.Title != nil {
		em.Title = &models.GaugeEsqlTitle{}
		em.Title.Text = typeutils.StringishPointerValue(api.Metric.Title.Text)
		em.Title.Visible = types.BoolPointerValue(api.Metric.Title.Visible)
	}
	m.EsqlMetric = em

	m.Styling = &models.GaugeStylingModel{}
	if api.Styling != nil && api.Styling.Shape != nil {
		shapeBytes, err := api.Styling.Shape.MarshalJSON()
		sv, ok := lenscommon.WrapNormalizedJSON(shapeBytes, err, "shape", &diags)
		if !ok {
			return diags
		}
		m.Styling.ShapeJSON = sv
	} else {
		m.Styling.ShapeJSON = jsontypes.NewNormalizedNull()
	}

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}

	return diags
}

func gaugeConfigToAPI(m *models.GaugeConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var attrs lenscommon.VisByValueConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	if lenscommon.ConfigUsesESQL(m.Query) {
		esql, d := gaugeConfigToAPIESQL(m)
		diags.Append(d...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromKibanaHTTPAPIsGaugeESQLByValuePanel(esql); err != nil {
			diags.AddError("Failed to create gauge ES|QL attributes", err.Error())
		}
		return attrs, diags
	}

	noESQL, d := gaugeConfigToAPINoESQL(m)
	diags.Append(d...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromKibanaHTTPAPIsGaugeNoESQLByValuePanel(noESQL); err != nil {
		diags.AddError("Failed to create gauge attributes", err.Error())
	}
	return attrs, diags
}

func gaugeConfigToAPINoESQL(m *models.GaugeConfigModel) (kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanel

	api.Type = kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanelTypeGauge

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	if typeutils.IsKnown(m.DataSourceJSON) {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal gauge_config.data_source_json", err.Error())
			return api, diags
		}
	}

	if m.Query == nil {
		diags.AddError("Missing query", "gauge_config.query must be set for non-ES|QL gauges (or omit `query` entirely for ES|QL mode)")
		return api, diags
	}
	api.Query = lenscommon.FilterSimpleToAPI(m.Query)

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if m.MetricJSON.IsNull() {
		diags.AddError("Missing metric_json", "gauge_config.metric_json must be set for non-ES|QL gauges (or use `esql_metric` in ES|QL mode)")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
		diags.AddError("Failed to unmarshal metric", err.Error())
		return api, diags
	}

	if m.Styling != nil && typeutils.IsKnown(m.Styling.ShapeJSON) {
		var shape kbapi.KibanaHTTPAPIsGaugeStyling_Shape
		shapeDiags := m.Styling.ShapeJSON.Unmarshal(&shape)
		diags.Append(shapeDiags...)
		if !shapeDiags.HasError() {
			if api.Styling == nil {
				api.Styling = &kbapi.KibanaHTTPAPIsGaugeStyling{}
			}
			api.Styling.Shape = &shape
		}
	}

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsGaugeNoESQLByValuePanel_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	return api, diags
}

func gaugeConfigToAPIESQL(m *models.GaugeConfigModel) (kbapi.KibanaHTTPAPIsGaugeESQLByValuePanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.KibanaHTTPAPIsGaugeESQLByValuePanel
	api.Type = kbapi.KibanaHTTPAPIsGaugeESQLByValuePanelTypeGauge

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing data_source_json", "gauge_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal gauge_config.data_source_json", err.Error())
		return api, diags
	}

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

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
		var color kbapi.KibanaHTTPAPIsGaugeESQLByValuePanel_Metric_Color
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
			Mode    *kbapi.KibanaHTTPAPIsGaugeESQLByValuePanelMetricTicksMode `json:"mode,omitempty"`
			Visible *bool                                                     `json:"visible,omitempty"`
		}{}
		if typeutils.IsKnown(m.EsqlMetric.Ticks.Mode) {
			mode := kbapi.KibanaHTTPAPIsGaugeESQLByValuePanelMetricTicksMode(m.EsqlMetric.Ticks.Mode.ValueString())
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
		var shape kbapi.KibanaHTTPAPIsGaugeStyling_Shape
		shapeDiags := m.Styling.ShapeJSON.Unmarshal(&shape)
		diags.Append(shapeDiags...)
		if !shapeDiags.HasError() {
			if api.Styling == nil {
				api.Styling = &kbapi.KibanaHTTPAPIsGaugeStyling{}
			}
			api.Styling.Shape = &shape
		}
	}

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsGaugeESQLByValuePanel_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	return api, diags
}
