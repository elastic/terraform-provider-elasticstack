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

func newTagcloudPanelConfigConverter() tagcloudPanelConfigConverter {
	return tagcloudPanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.TagcloudNoESQLTypeTagcloud),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.TagcloudConfig != nil },
		},
	}
}

type tagcloudPanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c tagcloudPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.TagcloudConfig != nil
}

func (c tagcloudPanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	// Try to extract the tagcloud config from the panel config
	cfgMap, err := config.AsDashboardPanelItemConfig8()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	// Extract the attributes
	attrs, ok := cfgMap["attributes"]
	if !ok {
		return nil
	}

	attrsMap, ok := attrs.(map[string]any)
	if !ok {
		return nil
	}

	// Marshal and unmarshal to get the TagcloudChart
	attrsJSON, err := json.Marshal(attrsMap)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var tagcloudChart kbapi.TagcloudChart
	if err := json.Unmarshal(attrsJSON, &tagcloudChart); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	// Try to extract as TagcloudNoESQL (standard non-ES|QL tagcloud)
	tagcloudNoESQL, err := tagcloudChart.AsTagcloudNoESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	// Populate the model
	pm.TagcloudConfig = &tagcloudConfigModel{}
	return pm.TagcloudConfig.fromAPI(ctx, tagcloudNoESQL)
}

func (c tagcloudPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	configModel := *pm.TagcloudConfig

	// Convert the structured model to API schema
	tagcloudNoESQL, tagcloudDiags := configModel.toAPI()
	diags.Append(tagcloudDiags...)
	if diags.HasError() {
		return diags
	}

	// Convert TagcloudNoESQL to TagcloudChart
	var tagcloudChart kbapi.TagcloudChart
	if err := tagcloudChart.FromTagcloudNoESQL(tagcloudNoESQL); err != nil {
		diags.AddError("Failed to convert tagcloud to schema", err.Error())
		return diags
	}

	// Create the nested dashboard panel config structure for the tagcloud attributes
	var attrs0 kbapi.DashboardPanelItemConfig70Attributes0
	if err := attrs0.FromTagcloudChart(tagcloudChart); err != nil {
		diags.AddError("Failed to create tagcloud attributes", err.Error())
		return diags
	}

	var configAttrs kbapi.DashboardPanelItem_Config_7_0_Attributes
	if err := configAttrs.FromDashboardPanelItemConfig70Attributes0(attrs0); err != nil {
		diags.AddError("Failed to create config attributes", err.Error())
		return diags
	}

	config10 := kbapi.DashboardPanelItemConfig70{
		Attributes: configAttrs,
	}

	var config1 kbapi.DashboardPanelItemConfig7
	if err := config1.FromDashboardPanelItemConfig70(config10); err != nil {
		diags.AddError("Failed to create config1", err.Error())
		return diags
	}

	if err := apiConfig.FromDashboardPanelItemConfig7(config1); err != nil {
		diags.AddError("Failed to marshal tagcloud config", err.Error())
	}

	return diags
}

type tagcloudConfigModel struct {
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DatasetJSON         jsontypes.Normalized                              `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []searchFilterModel                               `tfsdk:"filters"`
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
	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.DatasetJSON = jsontypes.NewNormalizedValue(string(datasetBytes))

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
	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	// Handle orientation
	m.Orientation = typeutils.StringishPointerValue(api.Orientation)

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
	if err != nil {
		diags.AddError("Failed to marshal metric", err.Error())
		return diags
	}
	m.MetricJSON = customtypes.NewJSONWithDefaultsValue[map[string]any](
		string(metricBytes),
		populateTagcloudMetricDefaults,
	)

	// Handle tagBy (as JSON) - union type
	tagByBytes, err := api.TagBy.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal tag_by", err.Error())
		return diags
	}
	m.TagByJSON = customtypes.NewJSONWithDefaultsValue[map[string]any](
		string(tagByBytes),
		populateTagcloudTagByDefaults,
	)

	return diags
}

func (m *tagcloudConfigModel) toAPI() (kbapi.TagcloudNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.TagcloudNoESQL

	// Set type to "tagcloud"
	api.Type = kbapi.TagcloudNoESQLTypeTagcloud

	if !m.Title.IsNull() {
		api.Title = m.Title.ValueStringPointer()
	}

	if !m.Description.IsNull() {
		api.Description = m.Description.ValueStringPointer()
	}

	// Handle dataset
	if !m.DatasetJSON.IsNull() {
		if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &api.Dataset); err != nil {
			diags.AddError("Failed to unmarshal dataset", err.Error())
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
	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilter, len(m.Filters))
		for i, filterModel := range m.Filters {
			filter, filterDiags := filterModel.toAPI()
			diags.Append(filterDiags...)
			filters[i] = filter
		}
		api.Filters = &filters
	}

	// Handle orientation
	if !m.Orientation.IsNull() {
		orientation := kbapi.TagcloudNoESQLOrientation(m.Orientation.ValueString())
		api.Orientation = &orientation
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
