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

func (c metricChartPanelConfigConverter) populateFromAttributes(ctx context.Context, pm *panelModel, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	// Populate the model.
	//
	// Disambiguate variant 0 vs 1 using dataset type. The regenerated API can
	// return an empty standard-query object, so query presence is not reliable.
	pm.MetricChartConfig = &metricChartConfigModel{}
	if variant0, err := attrs.AsMetricNoESQL(); err == nil && !isMetricNoESQLCandidateActuallyESQL(variant0) {
		return pm.MetricChartConfig.fromAPIVariant0(ctx, variant0)
	}
	variant1, err := attrs.AsMetricESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.MetricChartConfig.fromAPIVariant1(ctx, variant1)
}

func (c metricChartPanelConfigConverter) buildAttributes(pm panelModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *pm.MetricChartConfig

	// Convert the structured model to API schema
	attrs, metricDiags := configModel.toAPI()
	diags.Append(metricDiags...)
	if diags.HasError() {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	return attrs, diags
}

type metricChartConfigModel struct {
	Title               types.String           `tfsdk:"title"`
	Description         types.String           `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized   `tfsdk:"data_source_json"`
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

func isMetricNoESQLCandidateActuallyESQL(apiChart kbapi.MetricNoESQL) bool {
	body, err := json.Marshal(apiChart.DataSource)
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

func (m *metricChartConfigModel) fromAPI(ctx context.Context, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try to get the metric chart variant 0 (non-ESQL) or 1 (ESQL)
	// Both variants share the same "type" field value ("metric"), so we can't use that to distinguish them.
	// The key difference is the dataset type: data views and indices are no-ESQL,
	// while ESQL/table datasets belong to the ESQL variant.
	variant0, err := attrs.AsMetricNoESQL()
	if err == nil {
		if isMetricNoESQLCandidateActuallyESQL(variant0) {
			variant1, err1 := attrs.AsMetricESQL()
			if err1 == nil {
				return m.fromAPIVariant1(ctx, variant1)
			}
		}
		return m.fromAPIVariant0(ctx, variant0)
	}

	variant1, err := attrs.AsMetricESQL()
	if err == nil {
		return m.fromAPIVariant1(ctx, variant1)
	}

	diags.AddError("Failed to parse metric chart schema", "Could not parse as either variant 0 or 1")
	return diags
}

func (m *metricChartConfigModel) populateCommonFields(
	title, description *string,
	ignoreGlobalFilters *bool,
	sampling *float32,
	datasetBytes []byte,
	datasetErr error,
	filters []kbapi.LensPanelFilters_Item,
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
	dv, ok := marshalToNormalized(datasetBytes, datasetErr, "dataset", diags)
	if !ok {
		return false
	}
	m.DataSourceJSON = dv
	m.Filters = populateFiltersFromAPI(filters, diags)
	return !diags.HasError()
}

func (m *metricChartConfigModel) fromAPIVariant0(ctx context.Context, apiChart kbapi.MetricNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(apiChart.DataSource)
	if !m.populateCommonFields(apiChart.Title, apiChart.Description, apiChart.IgnoreGlobalFilters, apiChart.Sampling, datasetBytes, datasetErr, apiChart.Filters, &diags) {
		return diags
	}

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(apiChart.Query)

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

	datasetBytes, datasetErr := json.Marshal(apiChart.DataSource)
	if !m.populateCommonFields(apiChart.Title, apiChart.Description, apiChart.IgnoreGlobalFilters, apiChart.Sampling, datasetBytes, datasetErr, apiChart.Filters, &diags) {
		return diags
	}

	m.Query = nil

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

func (m *metricChartConfigModel) toAPI() (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	// Determine which variant to use based on whether we have a query
	// Variant 0 (non-ESQL) requires a query
	// Variant 1 (ESQL) doesn't require a query
	if m.Query != nil {
		return m.toAPIVariant0()
	}
	return m.toAPIVariant1()
}

func (m *metricChartConfigModel) toAPIVariant0() (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0

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
	if typeutils.IsKnown(m.DataSourceJSON) {
		var dataset kbapi.MetricNoESQL_DataSource
		datasetDiags := m.DataSourceJSON.Unmarshal(&dataset)
		diags.Append(datasetDiags...)
		if !datasetDiags.HasError() {
			variant0.DataSource = dataset
		}
	}

	// Set query
	if m.Query != nil {
		variant0.Query = m.Query.toAPI()
	}

	// Set filters
	variant0.Filters = buildFiltersForAPI(m.Filters, &diags)

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

	if err := attrs.FromMetricNoESQL(variant0); err != nil {
		diags.AddError("Failed to create metric chart schema variant 0", err.Error())
	}

	return attrs, diags
}

func (m *metricChartConfigModel) toAPIVariant1() (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0

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
	if typeutils.IsKnown(m.DataSourceJSON) {
		var dataset kbapi.EsqlDataSource
		datasetDiags := m.DataSourceJSON.Unmarshal(&dataset)
		diags.Append(datasetDiags...)
		if !datasetDiags.HasError() {
			variant1.DataSource = dataset
		}
	}

	// Set filters
	variant1.Filters = buildFiltersForAPI(m.Filters, &diags)

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
			CollapseBy kbapi.CollapseBy `json:"collapse_by"`
			Column     string           `json:"column"`
			Columns    *float32         `json:"columns,omitempty"`
			Format     kbapi.FormatType `json:"format"`
			Label      *string          `json:"label,omitempty"`
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

	if err := attrs.FromMetricESQL(variant1); err != nil {
		diags.AddError("Failed to create metric chart schema variant 1", err.Error())
	}

	return attrs, diags
}
