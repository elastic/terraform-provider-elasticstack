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

package lensmetric

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const jsonNullString = "null"

func metricChartAttrsFromPayload(payload any) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs lenscommon.VisByValueConfig0

	rawBytes, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Failed to marshal metric chart payload", err.Error())
		return attrs, diags
	}

	var raw map[string]any
	if err := json.Unmarshal(rawBytes, &raw); err != nil {
		diags.AddError("Failed to decode metric chart payload", err.Error())
		return attrs, diags
	}

	if styling, ok := raw["styling"].(map[string]any); ok {
		if icon, ok := styling["icon"].(map[string]any); ok {
			if name, _ := icon["name"].(string); name == "" {
				delete(styling, "icon")
			}
		}
		if len(styling) == 0 {
			delete(raw, "styling")
		} else {
			raw["styling"] = styling
		}
	}

	cleanedBytes, err := json.Marshal(raw)
	if err != nil {
		diags.AddError("Failed to marshal cleaned metric chart payload", err.Error())
		return attrs, diags
	}

	if err := json.Unmarshal(cleanedBytes, &attrs); err != nil {
		diags.AddError("Failed to create metric chart schema", err.Error())
	}

	return attrs, diags
}

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

func isMetricNoESQLCandidateActuallyESQL(apiChart kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel) bool {
	body, err := json.Marshal(apiChart.DataSource)
	return lenscommon.LensDataSourceIsESQLOrTable(body, err)
}

func metricChartConfigFromAPIVariant0(
	ctx context.Context,
	m *models.MetricChartConfigModel,
	prior *models.MetricChartConfigModel,
	apiChart kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(apiChart.DataSource)
	base, ok := lenscommon.PopulateLensChartBaseFromAPI(
		apiChart.Title, apiChart.Description, apiChart.IgnoreGlobalFilters, apiChart.Sampling,
		datasetBytes, datasetErr, "dataset", apiChart.Filters, &diags,
	)
	if !ok {
		return diags
	}
	m.LensChartBaseTFModel = base

	m.Query = &models.FilterSimpleModel{}
	lenscommon.FilterSimpleFromAPI(m.Query, apiChart.Query)

	if len(apiChart.Metrics) > 0 {
		priorMetrics := m.Metrics
		m.Metrics = make([]models.MetricItemModel, len(apiChart.Metrics))
		for i, metric := range apiChart.Metrics {
			metricJSON, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				continue
			}
			cfg := customtypes.NewJSONWithDefaultsValue(
				string(metricJSON),
				lenscommon.PopulateMetricChartMetricDefaults,
			)
			if i < len(priorMetrics) && lenscommon.MetricChartMetricConfigsEquivalent(priorMetrics[i].ConfigJSON, cfg) {
				cfg = priorMetrics[i].ConfigJSON
			}
			m.Metrics[i].ConfigJSON = cfg
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

	if !lenscommon.PopulateLensChartPresentation(
		ctx, &m.LensChartPresentationTFModel, prior, apiChart.TimeRange,
		apiChart.HideTitle, apiChart.HideBorder, apiChart.References, apiChart.Drilldowns, &diags,
	) {
		return diags
	}

	return diags
}

func metricChartConfigFromAPIVariant1(
	ctx context.Context,
	m *models.MetricChartConfigModel,
	prior *models.MetricChartConfigModel,
	apiChart kbapi.KibanaHTTPAPIsMetricESQLByValuePanel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(apiChart.DataSource)
	base, ok := lenscommon.PopulateLensChartBaseFromAPI(
		apiChart.Title, apiChart.Description, apiChart.IgnoreGlobalFilters, apiChart.Sampling,
		datasetBytes, datasetErr, "dataset", apiChart.Filters, &diags,
	)
	if !ok {
		return diags
	}
	m.LensChartBaseTFModel = base

	m.Query = nil

	if len(apiChart.Metrics) > 0 {
		priorMetrics := m.Metrics
		m.Metrics = make([]models.MetricItemModel, len(apiChart.Metrics))
		for i, metric := range apiChart.Metrics {
			metricJSON, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				continue
			}
			cfg := customtypes.NewJSONWithDefaultsValue(
				string(metricJSON),
				lenscommon.PopulateMetricChartMetricDefaults,
			)
			if i < len(priorMetrics) && lenscommon.MetricChartMetricConfigsEquivalent(priorMetrics[i].ConfigJSON, cfg) {
				cfg = priorMetrics[i].ConfigJSON
			}
			m.Metrics[i].ConfigJSON = cfg
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

	if !lenscommon.PopulateLensChartPresentation(
		ctx, &m.LensChartPresentationTFModel, prior, apiChart.TimeRange,
		apiChart.HideTitle, apiChart.HideBorder, apiChart.References, apiChart.Drilldowns, &diags,
	) {
		return diags
	}

	return diags
}

func metricChartConfigUsesESQL(m *models.MetricChartConfigModel) bool {
	if m == nil || !typeutils.IsKnown(m.DataSourceJSON) {
		return false
	}
	return lenscommon.LensDataSourceIsESQLOrTable([]byte(m.DataSourceJSON.ValueString()), nil)
}

func metricChartConfigToAPI(m *models.MetricChartConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var attrs lenscommon.VisByValueConfig0
	var diags diag.Diagnostics
	if m == nil {
		return attrs, diags
	}
	if metricChartConfigUsesESQL(m) {
		return metricChartConfigToAPIVariant1(m)
	}
	return metricChartConfigToAPIVariant0(m)
}

func metricChartConfigToAPIVariant0(m *models.MetricChartConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs lenscommon.VisByValueConfig0

	variant0 := kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel{
		Type: kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanelTypeMetric,
	}
	styling0 := kbapi.KibanaHTTPAPIsMetricStyling{}
	variant0.Styling = &styling0

	// Set simple fields
	variant0.Title, variant0.Description, variant0.IgnoreGlobalFilters, variant0.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	// Set dataset
	if typeutils.IsKnown(m.DataSourceJSON) {
		var dataset kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel_DataSource
		datasetDiags := m.DataSourceJSON.Unmarshal(&dataset)
		diags.Append(datasetDiags...)
		if !datasetDiags.HasError() {
			variant0.DataSource = dataset
		}
	}

	// Set query
	if m.Query != nil {
		variant0.Query = lenscommon.FilterSimpleToAPI(m.Query)
	}

	// Set filters
	variant0.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	// Set metrics
	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel_Metrics_Item, len(m.Metrics))
		for i, metric := range m.Metrics {
			if typeutils.IsKnown(metric.ConfigJSON) {
				var metricItem kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel_Metrics_Item
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
		var breakdownBy kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel_BreakdownBy
		breakdownDiags := m.BreakdownByJSON.Unmarshal(&breakdownBy)
		diags.Append(breakdownDiags...)
		if !breakdownDiags.HasError() {
			variant0.BreakdownBy = &breakdownBy
		}
	}

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return lenscommon.VisByValueConfig0{}, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsMetricNoESQLByValuePanel_Drilldowns_Item](
		writes, &variant0.TimeRange, &variant0.HideTitle, &variant0.HideBorder, &variant0.References, &variant0.Drilldowns,
	)...)

	attrs, attrsDiags := metricChartAttrsFromPayload(variant0)
	diags.Append(attrsDiags...)
	return attrs, diags
}

func metricChartConfigToAPIVariant1(m *models.MetricChartConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs lenscommon.VisByValueConfig0

	variant1 := kbapi.KibanaHTTPAPIsMetricESQLByValuePanel{
		Type: kbapi.KibanaHTTPAPIsMetricESQLByValuePanelTypeMetric,
	}
	styling1 := kbapi.KibanaHTTPAPIsMetricStyling{}
	variant1.Styling = &styling1

	// Set simple fields
	variant1.Title, variant1.Description, variant1.IgnoreGlobalFilters, variant1.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	// Set dataset
	if typeutils.IsKnown(m.DataSourceJSON) {
		var dataset kbapi.KibanaHTTPAPIsEsqlDataSource
		datasetDiags := m.DataSourceJSON.Unmarshal(&dataset)
		diags.Append(datasetDiags...)
		if !datasetDiags.HasError() {
			variant1.DataSource = dataset
		}
	}

	// Set filters
	variant1.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	// Set metrics
	if len(m.Metrics) > 0 {
		metrics := make([]kbapi.KibanaHTTPAPIsMetricESQLByValuePanel_Metrics_Item, len(m.Metrics))
		for i, metric := range m.Metrics {
			if typeutils.IsKnown(metric.ConfigJSON) {
				var metricItem kbapi.KibanaHTTPAPIsMetricESQLByValuePanel_Metrics_Item
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
			CollapseBy *kbapi.KibanaHTTPAPIsCollapseBy `json:"collapse_by,omitempty"`
			Column     string                          `json:"column"`
			Columns    *float32                        `json:"columns,omitempty"`
			Format     *kbapi.KibanaHTTPAPIsFormatType `json:"format,omitempty"`
			Label      *string                         `json:"label,omitempty"`
		}
		breakdownDiags := m.BreakdownByJSON.Unmarshal(&breakdownBy)
		diags.Append(breakdownDiags...)
		if !breakdownDiags.HasError() {
			if breakdownBy.Format != nil {
				fb, _ := json.Marshal(breakdownBy.Format)
				if string(fb) == jsonNullString || len(fb) == 0 {
					var format kbapi.KibanaHTTPAPIsFormatType
					_ = format.FromKibanaHTTPAPIsNumericFormat(kbapi.KibanaHTTPAPIsNumericFormat{Type: kbapi.Number})
					breakdownBy.Format = &format
				}
			}
			variant1.BreakdownBy = &breakdownBy
		}
	}

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return lenscommon.VisByValueConfig0{}, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsMetricESQLByValuePanel_Drilldowns_Item](
		writes, &variant1.TimeRange, &variant1.HideTitle, &variant1.HideBorder, &variant1.References, &variant1.Drilldowns,
	)...)

	attrs, attrsDiags := metricChartAttrsFromPayload(variant1)
	diags.Append(attrsDiags...)
	return attrs, diags
}
