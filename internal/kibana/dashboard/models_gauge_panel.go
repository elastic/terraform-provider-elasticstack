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

func newGaugePanelConfigConverter() gaugePanelConfigConverter {
	return gaugePanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.GaugeNoESQLTypeGauge),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.GaugeConfig != nil },
		},
	}
}

type gaugePanelConfigConverter struct {
	lensVisualizationBase
}

func (c gaugePanelConfigConverter) populateFromAttributes(ctx context.Context, pm *panelModel, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	gaugeNoESQL, err := attrs.AsGaugeNoESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.GaugeConfig = &gaugeConfigModel{}
	return pm.GaugeConfig.fromAPI(ctx, gaugeNoESQL)
}

func (c gaugePanelConfigConverter) buildAttributes(pm panelModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *pm.GaugeConfig

	gaugeNoESQL, gaugeDiags := configModel.toAPI()
	diags.Append(gaugeDiags...)
	if diags.HasError() {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	if err := attrs.FromGaugeNoESQL(gaugeNoESQL); err != nil {
		diags.AddError("Failed to create gauge attributes", err.Error())
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	return attrs, diags
}

type gaugeConfigModel struct {
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized                              `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []chartFilterJSONModel                            `tfsdk:"filters"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
	ShapeJSON           jsontypes.Normalized                              `tfsdk:"shape_json"`
}

func (m *gaugeConfigModel) fromAPI(ctx context.Context, api kbapi.GaugeNoESQL) diag.Diagnostics {
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

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	m.Filters = populateFiltersFromAPI(api.Filters, &diags)

	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric", populateGaugeMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = mv

	if api.Shape != nil {
		shapeBytes, err := api.Shape.MarshalJSON()
		sv, ok := marshalToNormalized(shapeBytes, err, "shape", &diags)
		if !ok {
			return diags
		}
		m.ShapeJSON = sv
	} else {
		m.ShapeJSON = jsontypes.NewNormalizedNull()
	}

	return diags
}

func (m *gaugeConfigModel) toAPI() (kbapi.GaugeNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.GaugeNoESQL

	api.Type = kbapi.GaugeNoESQLTypeGauge
	api.TimeRange = lensPanelTimeRange()

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

	if m.Query != nil {
		api.Query = m.Query.toAPI()
	}

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if typeutils.IsKnown(m.MetricJSON) {
		if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
			diags.AddError("Failed to unmarshal metric", err.Error())
			return api, diags
		}
	}

	if typeutils.IsKnown(m.ShapeJSON) {
		var shape kbapi.GaugeNoESQL_Shape
		shapeDiags := m.ShapeJSON.Unmarshal(&shape)
		diags.Append(shapeDiags...)
		if !shapeDiags.HasError() {
			api.Shape = &shape
		}
	}

	return api, diags
}
