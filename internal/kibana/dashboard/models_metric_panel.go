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

func newMetricChartPanelConfigConverter() metricChartPanelConfigConverter {
	return metricChartPanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.MetricChartSchema0TypeMetric),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.MetricChartConfig != nil },
		},
	}
}

type metricChartPanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c metricChartPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.MetricChartConfig != nil
}

func (c metricChartPanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	// Try to extract the metric chart config from the panel config
	cfgMap, err := config.AsDashboardPanelItemConfig2()
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

	// Marshal and unmarshal to get the MetricChartSchema
	attrsJSON, err := json.Marshal(attrsMap)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var metricChart kbapi.MetricChartSchema
	if err := json.Unmarshal(attrsJSON, &metricChart); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	// Populate the model.
	//
	// Disambiguate variant 0 vs 1 using the presence of the `query` key. The generated union types can
	// successfully unmarshal into both variants, so relying on decoded field contents is brittle.
	pm.MetricChartConfig = &metricChartConfigModel{}
	if _, ok := attrsMap["query"]; ok {
		variant0, err := metricChart.AsMetricChartSchema0()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		return pm.MetricChartConfig.fromAPIVariant0(ctx, variant0)
	}

	variant1, err := metricChart.AsMetricChartSchema1()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.MetricChartConfig.fromAPIVariant1(ctx, variant1)
}

func (c metricChartPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	configModel := *pm.MetricChartConfig

	// Convert the structured model to API schema
	metricChart, metricDiags := configModel.toAPI()
	diags.Append(metricDiags...)
	if diags.HasError() {
		return diags
	}

	// Create the nested Config1 structure
	var attrs0 kbapi.DashboardPanelItemConfig10Attributes0
	if err := attrs0.FromMetricChartSchema(metricChart); err != nil {
		diags.AddError("Failed to create metric chart attributes", err.Error())
		return diags
	}

	var configAttrs kbapi.DashboardPanelItem_Config_1_0_Attributes
	if err := configAttrs.FromDashboardPanelItemConfig10Attributes0(attrs0); err != nil {
		diags.AddError("Failed to create config attributes", err.Error())
		return diags
	}

	config10 := kbapi.DashboardPanelItemConfig10{
		Attributes: configAttrs,
	}

	var config1 kbapi.DashboardPanelItemConfig1
	if err := config1.FromDashboardPanelItemConfig10(config10); err != nil {
		diags.AddError("Failed to create config1", err.Error())
		return diags
	}

	if err := apiConfig.FromDashboardPanelItemConfig1(config1); err != nil {
		diags.AddError("Failed to marshal metric chart config", err.Error())
		return diags
	}

	return diags
}

type metricChartConfigModel struct {
	Title               types.String         `tfsdk:"title"`
	Description         types.String         `tfsdk:"description"`
	DatasetJSON         jsontypes.Normalized `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool           `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64        `tfsdk:"sampling"`
	Query               *filterSimpleModel   `tfsdk:"query"`
	Filters             []searchFilterModel  `tfsdk:"filters"`
	Metrics             []metricItemModel    `tfsdk:"metrics"`
	BreakdownByJSON     jsontypes.Normalized `tfsdk:"breakdown_by_json"`
}

type metricItemModel struct {
	ConfigJSON customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config_json"`
}

func (m *metricChartConfigModel) fromAPI(ctx context.Context, apiChart kbapi.MetricChartSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try to get the metric chart variant 0 (non-ESQL) or 1 (ESQL)
	// Both variants share the same "type" field value ("metric"), so we can't use that to distinguish them.
	// The key difference is that variant 0 requires a Query field, while variant 1 doesn't have one.
	// We try variant 0 first, but if the Query field is empty (which happens when decoding variant 1 JSON),
	// we know it's actually variant 1.
	variant0, err := apiChart.AsMetricChartSchema0()
	if err == nil {
		// Check if this is actually variant 1 by looking at whether the Query field is empty
		if variant0.Query.Query == "" && variant0.Query.Language == nil {
			// This is likely variant 1 (ESQL), try decoding as that
			variant1, err1 := apiChart.AsMetricChartSchema1()
			if err1 == nil {
				return m.fromAPIVariant1(ctx, variant1)
			}
			// If variant 1 also fails, fall back to variant 0 anyway
		}
		return m.fromAPIVariant0(ctx, variant0)
	}

	variant1, err := apiChart.AsMetricChartSchema1()
	if err == nil {
		return m.fromAPIVariant1(ctx, variant1)
	}

	diags.AddError("Failed to parse metric chart schema", "Could not parse as either variant 0 or 1")
	return diags
}

func (m *metricChartConfigModel) fromAPIVariant0(ctx context.Context, apiChart kbapi.MetricChartSchema0) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	// Set simple fields
	m.Title = types.StringPointerValue(apiChart.Title)
	m.Description = types.StringPointerValue(apiChart.Description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(apiChart.IgnoreGlobalFilters)
	if apiChart.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*apiChart.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	// Set dataset
	datasetJSON, err := json.Marshal(apiChart.Dataset)
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.DatasetJSON = jsontypes.NewNormalizedValue(string(datasetJSON))

	// Set query
	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(apiChart.Query)

	// Set filters
	if apiChart.Filters != nil && len(*apiChart.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*apiChart.Filters))
		for i, filter := range *apiChart.Filters {
			filterDiags := m.Filters[i].fromAPI(filter)
			diags.Append(filterDiags...)
		}
	}

	// Set metrics - MetricChartSchema0 has a slice of metrics
	if len(apiChart.Metrics) > 0 {
		m.Metrics = make([]metricItemModel, len(apiChart.Metrics))
		for i, metric := range apiChart.Metrics {
			metricJSON, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				continue
			}
			m.Metrics[i].ConfigJSON = customtypes.NewJSONWithDefaultsValue(
				string(metricJSON),
				populateMetricChartMetricDefaults,
			)
		}
	}

	// Set breakdown_by
	if apiChart.BreakdownBy != nil {
		breakdownJSON, err := json.Marshal(apiChart.BreakdownBy)
		if err != nil {
			diags.AddError("Failed to marshal breakdown_by", err.Error())
		} else {
			m.BreakdownByJSON = jsontypes.NewNormalizedValue(string(breakdownJSON))
		}
	} else {
		m.BreakdownByJSON = jsontypes.NewNormalizedNull()
	}

	return diags
}

func (m *metricChartConfigModel) fromAPIVariant1(ctx context.Context, apiChart kbapi.MetricChartSchema1) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	// Set simple fields
	m.Title = types.StringPointerValue(apiChart.Title)
	m.Description = types.StringPointerValue(apiChart.Description)
	m.IgnoreGlobalFilters = types.BoolPointerValue(apiChart.IgnoreGlobalFilters)
	if apiChart.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*apiChart.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	// Set dataset
	datasetJSON, err := json.Marshal(apiChart.Dataset)
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.DatasetJSON = jsontypes.NewNormalizedValue(string(datasetJSON))

	// Variant 1 doesn't always have a query (ES|QL case)
	m.Query = nil

	// Set filters
	if apiChart.Filters != nil && len(*apiChart.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*apiChart.Filters))
		for i, filter := range *apiChart.Filters {
			filterDiags := m.Filters[i].fromAPI(filter)
			diags.Append(filterDiags...)
		}
	}

	// Set metrics - MetricChartSchema1 has a slice of metrics
	if len(apiChart.Metrics) > 0 {
		m.Metrics = make([]metricItemModel, len(apiChart.Metrics))
		for i, metric := range apiChart.Metrics {
			metricJSON, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				continue
			}
			m.Metrics[i].ConfigJSON = customtypes.NewJSONWithDefaultsValue(
				string(metricJSON),
				populateMetricChartMetricDefaults,
			)
		}
	}

	// Set breakdown_by
	if apiChart.BreakdownBy != nil {
		breakdownJSON, err := json.Marshal(apiChart.BreakdownBy)
		if err != nil {
			diags.AddError("Failed to marshal breakdown_by", err.Error())
		} else {
			m.BreakdownByJSON = jsontypes.NewNormalizedValue(string(breakdownJSON))
		}
	} else {
		m.BreakdownByJSON = jsontypes.NewNormalizedNull()
	}

	return diags
}

func (m *metricChartConfigModel) toAPI() (kbapi.MetricChartSchema, diag.Diagnostics) {
	// Determine which variant to use based on whether we have a query
	// Variant 0 (non-ESQL) requires a query
	// Variant 1 (ESQL) doesn't require a query
	if m.Query != nil {
		return m.toAPIVariant0()
	}
	return m.toAPIVariant1()
}

func (m *metricChartConfigModel) toAPIVariant0() (kbapi.MetricChartSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	var metricChart kbapi.MetricChartSchema

	variant0 := kbapi.MetricChartSchema0{
		Type: kbapi.MetricChartSchema0TypeMetric,
	}

	// Set simple fields
	if typeutils.IsKnown(m.Title) {
		variant0.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		variant0.Description = new(m.Description.ValueString())
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		variant0.IgnoreGlobalFilters = new(m.IgnoreGlobalFilters.ValueBool())
	}
	if typeutils.IsKnown(m.Sampling) {
		sampling := float32(m.Sampling.ValueFloat64())
		variant0.Sampling = &sampling
	}

	// Set dataset
	if typeutils.IsKnown(m.DatasetJSON) {
		var dataset kbapi.MetricChartSchema_0_Dataset
		datasetDiags := m.DatasetJSON.Unmarshal(&dataset)
		diags.Append(datasetDiags...)
		if !datasetDiags.HasError() {
			variant0.Dataset = dataset
		}
	}

	// Set query
	if m.Query != nil {
		variant0.Query = m.Query.toAPI()
	}

	// Set filters
	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
		for i, filter := range m.Filters {
			apiFilter, filterDiags := filter.toAPI()
			diags.Append(filterDiags...)
			filters[i] = apiFilter
		}
		variant0.Filters = &filters
	}

	// Set metrics
	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.MetricChartSchema_0_Metrics_Item, len(m.Metrics))
		for i, metric := range m.Metrics {
			if typeutils.IsKnown(metric.ConfigJSON) {
				var metricItem kbapi.MetricChartSchema_0_Metrics_Item
				metricDiags := metric.ConfigJSON.Unmarshal(&metricItem)
				diags.Append(metricDiags...)
				if !metricDiags.HasError() {
					metrics[i] = metricItem
				}
			}
		}
		variant0.Metrics = metrics
	}

	// Set breakdown_by
	if typeutils.IsKnown(m.BreakdownByJSON) {
		var breakdownBy kbapi.MetricChartSchema_0_BreakdownBy
		breakdownDiags := m.BreakdownByJSON.Unmarshal(&breakdownBy)
		diags.Append(breakdownDiags...)
		if !breakdownDiags.HasError() {
			variant0.BreakdownBy = &breakdownBy
		}
	}

	if err := metricChart.FromMetricChartSchema0(variant0); err != nil {
		diags.AddError("Failed to create metric chart schema variant 0", err.Error())
	}

	return metricChart, diags
}

func (m *metricChartConfigModel) toAPIVariant1() (kbapi.MetricChartSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	var metricChart kbapi.MetricChartSchema

	variant1 := kbapi.MetricChartSchema1{
		Type: kbapi.MetricChartSchema1TypeMetric,
	}

	// Set simple fields
	if typeutils.IsKnown(m.Title) {
		variant1.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		variant1.Description = new(m.Description.ValueString())
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		variant1.IgnoreGlobalFilters = new(m.IgnoreGlobalFilters.ValueBool())
	}
	if typeutils.IsKnown(m.Sampling) {
		sampling := float32(m.Sampling.ValueFloat64())
		variant1.Sampling = &sampling
	}

	// Set dataset
	if typeutils.IsKnown(m.DatasetJSON) {
		var dataset kbapi.MetricChartSchema_1_Dataset
		datasetDiags := m.DatasetJSON.Unmarshal(&dataset)
		diags.Append(datasetDiags...)
		if !datasetDiags.HasError() {
			variant1.Dataset = dataset
		}
	}

	// Set filters
	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
		for i, filter := range m.Filters {
			apiFilter, filterDiags := filter.toAPI()
			diags.Append(filterDiags...)
			filters[i] = apiFilter
		}
		variant1.Filters = &filters
	}

	// Set metrics
	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.MetricChartSchema_1_Metrics_Item, len(m.Metrics))
		for i, metric := range m.Metrics {
			if typeutils.IsKnown(metric.ConfigJSON) {
				var metricItem kbapi.MetricChartSchema_1_Metrics_Item
				metricDiags := metric.ConfigJSON.Unmarshal(&metricItem)
				diags.Append(metricDiags...)
				if !metricDiags.HasError() {
					metrics[i] = metricItem
				}
			}
		}
		variant1.Metrics = metrics
	}

	// Set breakdown_by
	if typeutils.IsKnown(m.BreakdownByJSON) {
		var breakdownBy struct {
			CollapseBy kbapi.CollapseBy                             `json:"collapse_by"`
			Column     string                                       `json:"column"`
			Columns    *float32                                     `json:"columns,omitempty"`
			Operation  kbapi.MetricChartSchema1BreakdownByOperation `json:"operation"`
		}
		breakdownDiags := m.BreakdownByJSON.Unmarshal(&breakdownBy)
		diags.Append(breakdownDiags...)
		if !breakdownDiags.HasError() {
			variant1.BreakdownBy = &breakdownBy
		}
	}

	if err := metricChart.FromMetricChartSchema1(variant1); err != nil {
		diags.AddError("Failed to create metric chart schema variant 1", err.Error())
	}

	return metricChart, diags
}
