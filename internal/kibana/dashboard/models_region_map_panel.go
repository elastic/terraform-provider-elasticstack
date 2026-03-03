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

func newRegionMapPanelConfigConverter() regionMapPanelConfigConverter {
	return regionMapPanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.RegionMapNoESQLTypeRegionMap),
		},
	}
}

type regionMapPanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c regionMapPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.RegionMapConfig != nil
}

func (c regionMapPanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	cfgMap, err := config.AsDashboardPanelItemConfig8()
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

	var regionMap kbapi.RegionMapChart
	if err := json.Unmarshal(attrsJSON, &regionMap); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.RegionMapConfig = &regionMapConfigModel{}

	regionMapNoESQL, err := regionMap.AsRegionMapNoESQL()
	if err == nil {
		return pm.RegionMapConfig.fromAPINoESQL(ctx, regionMapNoESQL)
	}

	regionMapESQL, err := regionMap.AsRegionMapESQL()
	if err == nil {
		return pm.RegionMapConfig.fromAPIESQL(ctx, regionMapESQL)
	}

	return diagutil.FrameworkDiagFromError(err)
}

func (c regionMapPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	configModel := *pm.RegionMapConfig

	regionMap, regionDiags := configModel.toAPI()
	diags.Append(regionDiags...)
	if diags.HasError() {
		return diags
	}

	var attrs0 kbapi.DashboardPanelItemConfig70Attributes0
	if err := attrs0.FromRegionMapChart(regionMap); err != nil {
		diags.AddError("Failed to create region map attributes", err.Error())
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
		diags.AddError("Failed to marshal region map config", err.Error())
	}

	return diags
}

type regionMapConfigModel struct {
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DatasetJSON         jsontypes.Normalized                              `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []searchFilterModel                               `tfsdk:"filters"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
	RegionJSON          jsontypes.Normalized                              `tfsdk:"region_json"`
}

func (m *regionMapConfigModel) fromAPINoESQL(ctx context.Context, api kbapi.RegionMapNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

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

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	metricBytes, err := api.Metric.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal metric", err.Error())
		return diags
	}
	m.MetricJSON = customtypes.NewJSONWithDefaultsValue[map[string]any](
		string(metricBytes),
		populateRegionMapMetricDefaults,
	)

	regionBytes, err := api.Region.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal region", err.Error())
		return diags
	}
	m.RegionJSON = jsontypes.NewNormalizedValue(string(regionBytes))

	return diags
}

func (m *regionMapConfigModel) fromAPIESQL(ctx context.Context, api kbapi.RegionMapESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

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

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	metricBytes, err := json.Marshal(api.Metric)
	if err != nil {
		diags.AddError("Failed to marshal metric", err.Error())
		return diags
	}
	m.MetricJSON = customtypes.NewJSONWithDefaultsValue[map[string]any](
		string(metricBytes),
		populateRegionMapMetricDefaults,
	)

	regionBytes, err := json.Marshal(api.Region)
	if err != nil {
		diags.AddError("Failed to marshal region", err.Error())
		return diags
	}
	m.RegionJSON = jsontypes.NewNormalizedValue(string(regionBytes))

	return diags
}

func (m *regionMapConfigModel) toAPI() (kbapi.RegionMapChart, diag.Diagnostics) {
	var diags diag.Diagnostics

	if m == nil {
		return kbapi.RegionMapChart{}, diags
	}

	if m.Query != nil && typeutils.IsKnown(m.Query.Query) {
		api := kbapi.RegionMapNoESQL{
			Type: kbapi.RegionMapNoESQLTypeRegionMap,
		}

		if typeutils.IsKnown(m.Title) {
			api.Title = m.Title.ValueStringPointer()
		}
		if typeutils.IsKnown(m.Description) {
			api.Description = m.Description.ValueStringPointer()
		}
		if typeutils.IsKnown(m.DatasetJSON) {
			if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &api.Dataset); err != nil {
				diags.AddError("Failed to unmarshal dataset", err.Error())
				return kbapi.RegionMapChart{}, diags
			}
		}
		if typeutils.IsKnown(m.IgnoreGlobalFilters) {
			api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
		}
		if typeutils.IsKnown(m.Sampling) {
			sampling := float32(m.Sampling.ValueFloat64())
			api.Sampling = &sampling
		}
		api.Query = m.Query.toAPI()

		if len(m.Filters) > 0 {
			filters := make([]kbapi.SearchFilter, len(m.Filters))
			for i, filterModel := range m.Filters {
				filter, filterDiags := filterModel.toAPI()
				diags.Append(filterDiags...)
				filters[i] = filter
			}
			api.Filters = &filters
		}

		if typeutils.IsKnown(m.MetricJSON) {
			if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
				diags.AddError("Failed to unmarshal metric", err.Error())
				return kbapi.RegionMapChart{}, diags
			}
		}
		if typeutils.IsKnown(m.RegionJSON) {
			if err := json.Unmarshal([]byte(m.RegionJSON.ValueString()), &api.Region); err != nil {
				diags.AddError("Failed to unmarshal region", err.Error())
				return kbapi.RegionMapChart{}, diags
			}
		}

		var schema kbapi.RegionMapChart
		if err := schema.FromRegionMapNoESQL(api); err != nil {
			diags.AddError("Failed to create region map schema", err.Error())
		}
		return schema, diags
	}

	api := kbapi.RegionMapESQL{
		Type: kbapi.RegionMapESQLTypeRegionMap,
	}

	if typeutils.IsKnown(m.Title) {
		api.Title = m.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = m.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(m.DatasetJSON) {
		if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &api.Dataset); err != nil {
			diags.AddError("Failed to unmarshal dataset", err.Error())
			return kbapi.RegionMapChart{}, diags
		}
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.Sampling) {
		sampling := float32(m.Sampling.ValueFloat64())
		api.Sampling = &sampling
	}

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilter, len(m.Filters))
		for i, filterModel := range m.Filters {
			filter, filterDiags := filterModel.toAPI()
			diags.Append(filterDiags...)
			filters[i] = filter
		}
		api.Filters = &filters
	}

	if typeutils.IsKnown(m.MetricJSON) {
		if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
			diags.AddError("Failed to unmarshal metric", err.Error())
			return kbapi.RegionMapChart{}, diags
		}
	}
	if typeutils.IsKnown(m.RegionJSON) {
		if err := json.Unmarshal([]byte(m.RegionJSON.ValueString()), &api.Region); err != nil {
			diags.AddError("Failed to unmarshal region", err.Error())
			return kbapi.RegionMapChart{}, diags
		}
	}

	var schema kbapi.RegionMapChart
	if err := schema.FromRegionMapESQL(api); err != nil {
		diags.AddError("Failed to create region map schema", err.Error())
	}
	return schema, diags
}
