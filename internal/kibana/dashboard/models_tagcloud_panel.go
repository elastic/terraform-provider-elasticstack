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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newTagcloudPanelConfigConverter() tagcloudPanelConfigConverter {
	return tagcloudPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.TagcloudNoESQLTypeTagCloud),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.TagcloudConfig != nil },
		},
	}
}

type tagcloudPanelConfigConverter struct {
	lensVisualizationBase
}

func (c tagcloudPanelConfigConverter) populateFromAttributes(ctx context.Context, pm *panelModel, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	tagcloudNoESQL, err := attrs.AsTagcloudNoESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	// Populate the model
	pm.TagcloudConfig = &tagcloudConfigModel{}
	return pm.TagcloudConfig.fromAPI(ctx, tagcloudNoESQL)
}

func (c tagcloudPanelConfigConverter) buildAttributes(pm panelModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *pm.TagcloudConfig

	// Convert the structured model to API schema
	tagcloudNoESQL, tagcloudDiags := configModel.toAPI()
	diags.Append(tagcloudDiags...)
	if diags.HasError() {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	if err := attrs.FromTagcloudNoESQL(tagcloudNoESQL); err != nil {
		diags.AddError("Failed to create tagcloud attributes", err.Error())
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	return attrs, diags
}

type tagcloudConfigModel struct {
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized                              `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []chartFilterJSONModel                            `tfsdk:"filters"`
	Orientation         types.String                                      `tfsdk:"orientation"`
	FontSize            *fontSizeModel                                    `tfsdk:"font_size"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
	TagByJSON           customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"tag_by_json"`
}

type fontSizeModel struct {
	Min types.Float64 `tfsdk:"min"`
	Max types.Float64 `tfsdk:"max"`
}

func (m *tagcloudConfigModel) fromAPI(ctx context.Context, api kbapi.TagcloudNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

	// Handle dataset
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

	// Handle query
	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	// Handle filters
	m.Filters = populateFiltersFromAPI(api.Filters, &diags)

	// Handle orientation
	if api.Orientation != "" {
		m.Orientation = types.StringValue(string(api.Orientation))
	} else {
		m.Orientation = types.StringNull()
	}

	// Handle font size
	if api.FontSize != nil {
		m.FontSize = &fontSizeModel{}
		if api.FontSize.Min != nil {
			m.FontSize.Min = types.Float64Value(float64(*api.FontSize.Min))
		} else {
			m.FontSize.Min = types.Float64Null()
		}
		if api.FontSize.Max != nil {
			m.FontSize.Max = types.Float64Value(float64(*api.FontSize.Max))
		} else {
			m.FontSize.Max = types.Float64Null()
		}
	}

	// Handle metric (as JSON) - union type
	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric", populateTagcloudMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = mv

	// Handle tagBy (as JSON) - union type
	tagByBytes, err := api.TagBy.MarshalJSON()
	tv, ok := marshalToJSONWithDefaults(tagByBytes, err, "tag_by", populateTagcloudTagByDefaults, &diags)
	if !ok {
		return diags
	}
	m.TagByJSON = tv

	return diags
}

func (m *tagcloudConfigModel) toAPI() (kbapi.TagcloudNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.TagcloudNoESQL

	// Set type to "tagcloud"
	api.Type = kbapi.TagcloudNoESQLTypeTagCloud
	api.TimeRange = lensPanelTimeRange()

	if !m.Title.IsNull() {
		api.Title = m.Title.ValueStringPointer()
	}

	if !m.Description.IsNull() {
		api.Description = m.Description.ValueStringPointer()
	}

	// Handle dataset
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

	// Handle query
	if m.Query != nil {
		api.Query = m.Query.toAPI()
	}

	// Handle filters
	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	// Handle orientation
	if !m.Orientation.IsNull() {
		api.Orientation = kbapi.VisApiOrientation(m.Orientation.ValueString())
	}

	// Handle font size
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
		api.FontSize = &fontSize
	}

	// Handle metric (as JSON)
	if !m.MetricJSON.IsNull() {
		if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
			diags.AddError("Failed to unmarshal metric", err.Error())
			return api, diags
		}
	}

	// Handle tagBy (as JSON)
	if !m.TagByJSON.IsNull() {
		if err := json.Unmarshal([]byte(m.TagByJSON.ValueString()), &api.TagBy); err != nil {
			diags.AddError("Failed to unmarshal tag_by", err.Error())
			return api, diags
		}
	}

	return api, diags
}
