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

// stripMetricBreakdownByAPIFields removes server-added fields from breakdown_by JSON
// that are not part of the API spec but are returned by the Kibana API on read.
func stripMetricBreakdownByAPIFields(jsonStr string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return jsonStr
	}
	delete(m, "columns")
	out, err := json.Marshal(m)
	if err != nil {
		return jsonStr
	}
	return string(out)
}

func newMetricChartPanelConfigConverter() metricChartPanelConfigConverter {
	return metricChartPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.MetricNoESQLTypeMetric),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.MetricChartConfig != nil },
		},
	}
}

type metricChartPanelConfigConverter struct {
	lensVisualizationBase
}

func (c metricChartPanelConfigConverter) populateFromAttributes(ctx context.Context, pm *panelModel, attrs kbapi.LensApiState) diag.Diagnostics {
	metricChart, err := attrs.AsMetricChart()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	// Populate the model.
	//
	// Disambiguate variant 0 vs 1 using query presence in decoded variant0; variant1 (ESQL)
	// can decode into variant0 but leaves query empty.
	pm.MetricChartConfig = &metricChartConfigModel{}
	if variant0, err := metricChart.AsMetricNoESQL(); err == nil && (variant0.Query.Query != "" || variant0.Query.Language != nil) {
		return pm.MetricChartConfig.fromAPIVariant0(ctx, variant0)
	}
	variant1, err := metricChart.AsMetricESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.MetricChartConfig.fromAPIVariant1(ctx, variant1)
}

func (c metricChartPanelConfigConverter) buildAttributes(pm panelModel) (kbapi.LensApiState, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *pm.MetricChartConfig

	// Convert the structured model to API schema
	metricChart, metricDiags := configModel.toAPI()
	diags.Append(metricDiags...)
	if diags.HasError() {
		return kbapi.LensApiState{}, diags
	}

	var attrs kbapi.LensApiState
	if err := attrs.FromMetricChart(metricChart); err != nil {
		diags.AddError("Failed to create metric chart attributes", err.Error())
		return kbapi.LensApiState{}, diags
	}

	return attrs, diags
}

type metricChartConfigModel struct {
	Title               types.String           `tfsdk:"title"`
	Description         types.String           `tfsdk:"description"`
	DatasetJSON         jsontypes.Normalized   `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool             `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64          `tfsdk:"sampling"`
	Query               *filterSimpleModel     `tfsdk:"query"`
	Filters             []chartFilterJSONModel `tfsdk:"filters"`
	Metrics             []metricItemModel      `tfsdk:"metrics"`
	BreakdownByJSON     jsontypes.Normalized   `tfsdk:"breakdown_by_json"`
}

type metricItemModel struct {
	ConfigJSON customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config_json"`
}

func (m *metricChartConfigModel) fromAPI(ctx context.Context, apiChart kbapi.MetricChart) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try to get the metric chart variant 0 (non-ESQL) or 1 (ESQL)
	// Both variants share the same "type" field value ("metric"), so we can't use that to distinguish them.
	// The key difference is that variant 0 requires a Query field, while variant 1 doesn't have one.
	// We try variant 0 first, but if the Query field is empty (which happens when decoding variant 1 JSON),
	// we know it's actually variant 1.
	variant0, err := apiChart.AsMetricNoESQL()
	if err == nil {
		// Check if this is actually variant 1 by looking at whether the Query field is empty
		if variant0.Query.Query == "" && variant0.Query.Language == nil {
			// This is likely variant 1 (ESQL), try decoding as that
			variant1, err1 := apiChart.AsMetricESQL()
			if err1 == nil {
				return m.fromAPIVariant1(ctx, variant1)
			}
			// If variant 1 also fails, fall back to variant 0 anyway
		}
		return m.fromAPIVariant0(ctx, variant0)
	}

	variant1, err := apiChart.AsMetricESQL()
	if err == nil {
		return m.fromAPIVariant1(ctx, variant1)
	}

	diags.AddError("Failed to parse metric chart schema", "Could not parse as either variant 0 or 1")
	return diags
}

func (m *metricChartConfigModel) fromAPIVariant0(ctx context.Context, apiChart kbapi.MetricNoESQL) diag.Diagnostics {
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
	if len(apiChart.Filters) > 0 {
		m.Filters = make([]chartFilterJSONModel, 0, len(apiChart.Filters))
		for _, filter := range apiChart.Filters {
			fm := chartFilterJSONModel{}
			filterDiags := fm.populateFromAPIItem(filter)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				m.Filters = append(m.Filters, fm)
			}
		}
	}

	// Set metrics - MetricChart0 has a slice of metrics
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
				populateLensMetricDefaults,
			)
		}
	}

	// Set breakdown_by
	if apiChart.BreakdownBy != nil {
		breakdownJSON, err := json.Marshal(apiChart.BreakdownBy)
		if err != nil {
			diags.AddError("Failed to marshal breakdown_by", err.Error())
		} else {
			m.BreakdownByJSON = jsontypes.NewNormalizedValue(stripMetricBreakdownByAPIFields(string(breakdownJSON)))
		}
	} else {
		m.BreakdownByJSON = jsontypes.NewNormalizedNull()
	}

	return diags
}

func (m *metricChartConfigModel) fromAPIVariant1(ctx context.Context, apiChart kbapi.MetricESQL) diag.Diagnostics {
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
	if len(apiChart.Filters) > 0 {
		m.Filters = make([]chartFilterJSONModel, 0, len(apiChart.Filters))
		for _, filter := range apiChart.Filters {
			fm := chartFilterJSONModel{}
			filterDiags := fm.populateFromAPIItem(filter)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				m.Filters = append(m.Filters, fm)
			}
		}
	}

	// Set metrics - MetricChart1 has a slice of metrics
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
				populateLensMetricDefaults,
			)
		}
	}

	// Set breakdown_by
	if apiChart.BreakdownBy != nil {
		breakdownJSON, err := json.Marshal(apiChart.BreakdownBy)
		if err != nil {
			diags.AddError("Failed to marshal breakdown_by", err.Error())
		} else {
			m.BreakdownByJSON = jsontypes.NewNormalizedValue(stripMetricBreakdownByAPIFields(string(breakdownJSON)))
		}
	} else {
		m.BreakdownByJSON = jsontypes.NewNormalizedNull()
	}

	return diags
}

func (m *metricChartConfigModel) toAPI() (kbapi.MetricChart, diag.Diagnostics) {
	// Determine which variant to use based on whether we have a query
	// Variant 0 (non-ESQL) requires a query
	// Variant 1 (ESQL) doesn't require a query
	if m.Query != nil {
		return m.toAPIVariant0()
	}
	return m.toAPIVariant1()
}

func (m *metricChartConfigModel) toAPIVariant0() (kbapi.MetricChart, diag.Diagnostics) {
	var diags diag.Diagnostics
	var metricChart kbapi.MetricChart

	variant0 := kbapi.MetricNoESQL{
		Type: kbapi.MetricNoESQLTypeMetric,
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
		var dataset kbapi.MetricNoESQL_Dataset
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
	variant0.Filters = []kbapi.LensPanelFilters_Item{}
	if len(m.Filters) > 0 {
		filters := make([]kbapi.LensPanelFilters_Item, 0, len(m.Filters))
		for _, filter := range m.Filters {
			var item kbapi.LensPanelFilters_Item
			filterDiags := decodeChartFilterJSON(filter.FilterJSON, &item)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				filters = append(filters, item)
			}
		}
		if len(filters) > 0 {
			variant0.Filters = filters
		}
	}

	// Set metrics
	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.MetricNoESQL_Metrics_Item, len(m.Metrics))
		for i, metric := range m.Metrics {
			if typeutils.IsKnown(metric.ConfigJSON) {
				var metricItem kbapi.MetricNoESQL_Metrics_Item
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
		var breakdownBy kbapi.MetricNoESQL_BreakdownBy
		breakdownDiags := m.BreakdownByJSON.Unmarshal(&breakdownBy)
		diags.Append(breakdownDiags...)
		if !breakdownDiags.HasError() {
			variant0.BreakdownBy = &breakdownBy
		}
	}

	if err := metricChart.FromMetricNoESQL(variant0); err != nil {
		diags.AddError("Failed to create metric chart schema variant 0", err.Error())
	}

	return metricChart, diags
}

func (m *metricChartConfigModel) toAPIVariant1() (kbapi.MetricChart, diag.Diagnostics) {
	var diags diag.Diagnostics
	var metricChart kbapi.MetricChart

	variant1 := kbapi.MetricESQL{
		Type: kbapi.MetricESQLTypeMetric,
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
		var dataset kbapi.MetricESQL_Dataset
		datasetDiags := m.DatasetJSON.Unmarshal(&dataset)
		diags.Append(datasetDiags...)
		if !datasetDiags.HasError() {
			variant1.Dataset = dataset
		}
	}

	// Set filters
	variant1.Filters = []kbapi.LensPanelFilters_Item{}
	if len(m.Filters) > 0 {
		filters := make([]kbapi.LensPanelFilters_Item, 0, len(m.Filters))
		for _, filter := range m.Filters {
			var item kbapi.LensPanelFilters_Item
			filterDiags := decodeChartFilterJSON(filter.FilterJSON, &item)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				filters = append(filters, item)
			}
		}
		if len(filters) > 0 {
			variant1.Filters = filters
		}
	}

	// Set metrics
	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.MetricESQL_Metrics_Item, len(m.Metrics))
		for i, metric := range m.Metrics {
			if typeutils.IsKnown(metric.ConfigJSON) {
				var metricItem kbapi.MetricESQL_Metrics_Item
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
			CollapseBy kbapi.CollapseBy                     `json:"collapse_by"`
			Column     string                               `json:"column"`
			Columns    *float32                             `json:"columns,omitempty"`
			Format     kbapi.FormatType                     `json:"format"`
			Label      *string                              `json:"label,omitempty"`
			Operation  kbapi.MetricESQLBreakdownByOperation `json:"operation"`
		}
		breakdownDiags := m.BreakdownByJSON.Unmarshal(&breakdownBy)
		diags.Append(breakdownDiags...)
		if !breakdownDiags.HasError() {
			fb, _ := json.Marshal(breakdownBy.Format)
			if string(fb) == jsonNullString || len(fb) == 0 {
				_ = breakdownBy.Format.FromNumericFormat(kbapi.NumericFormat{Type: kbapi.Number})
			}
			variant1.BreakdownBy = &breakdownBy
		}
	}

	if err := metricChart.FromMetricESQL(variant1); err != nil {
		diags.AddError("Failed to create metric chart schema variant 1", err.Error())
	}

	return metricChart, diags
}
