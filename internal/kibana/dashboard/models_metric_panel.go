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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func metricChartAttrsFromPayload(payload any) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0

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

func newMetricChartPanelConfigConverter() metricChartPanelConfigConverter {
	return metricChartPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.MetricNoESQLTypeMetric),
			hasTFChartBlock: func(blocks *models.LensByValueChartBlocks) bool {
				return blocks != nil && blocks.MetricChartConfig != nil
			},
		},
	}
}

type metricChartPanelConfigConverter struct {
	lensVisualizationBase
}

func (c metricChartPanelConfigConverter) populateFromAttributes(
	ctx context.Context,
	dashboard *models.DashboardModel,
	tfPanel *models.PanelModel,
	blocks *models.LensByValueChartBlocks,
	attrs kbapi.KbnDashboardPanelTypeVisConfig0,
) diag.Diagnostics {
	// Populate the model.
	//
	// Disambiguate variant 0 vs 1 using dataset type. The regenerated API can
	// return an empty standard-query object, so query presence is not reliable.
	//
	// Always allocate a fresh models.MetricChartConfigModel so that fromAPIVariant0/1
	// does not mutate the plan's struct (blocks is seeded from the plan via panel
	// read path). Seed the fresh struct with the prior metrics slice so the inline priorMetrics preservation
	// inside fromAPIVariant0 can still compare against plan values.
	priorConfig := blocks.MetricChartConfig
	if priorConfig == nil {
		if b := lensByValueChartBlocksFromPanel(tfPanel); b != nil && b.MetricChartConfig != nil {
			priorConfig = b.MetricChartConfig
		}
	}
	blocks.MetricChartConfig = &models.MetricChartConfigModel{}
	if priorConfig != nil {
		blocks.MetricChartConfig.Metrics = priorConfig.Metrics
	}
	if variant0, err := attrs.AsMetricNoESQL(); err == nil && !isMetricNoESQLCandidateActuallyESQL(variant0) {
		return metricChartConfigFromAPIVariant0(ctx, blocks.MetricChartConfig, dashboard, priorConfig, variant0)
	}
	variant1, err := attrs.AsMetricESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return metricChartConfigFromAPIVariant1(ctx, blocks.MetricChartConfig, dashboard, priorConfig, variant1)
}

func (c metricChartPanelConfigConverter) buildAttributes(blocks *models.LensByValueChartBlocks, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *blocks.MetricChartConfig

	// Convert the structured model to API schema
	attrs, metricDiags := metricChartConfigToAPI(&configModel, dashboard)
	diags.Append(metricDiags...)
	if diags.HasError() {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	return attrs, diags
}

// models.MetricChartCoreTFModel carries metric chart Terraform attributes that exist for both vis panels and
// lens-dashboard-app typed by-value blocks (REQ-037 presentation siblings are modeled separately).

// Schema uses getMetricChart(false), so REQ-037 presentation siblings are omitted from the Terraform object even though
// the vis conversion path temporarily expands this into models.MetricChartConfigModel (with null presentation defaults).

func metricChartLensByValueTFExpandToVisMetricChart(s *models.MetricChartLensByValueTFModel) *models.MetricChartConfigModel {
	if s == nil {
		return nil
	}
	out := &models.MetricChartConfigModel{MetricChartCoreTFModel: s.MetricChartCoreTFModel}
	out.LensChartPresentationTFModel = newNullLensChartPresentationTFModel()
	return out
}

func metricLensByValueFromVisFull(m *models.MetricChartConfigModel) *models.MetricChartLensByValueTFModel {
	if m == nil {
		return nil
	}
	return &models.MetricChartLensByValueTFModel{MetricChartCoreTFModel: m.MetricChartCoreTFModel}
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

func metricChartConfigFromAPI(ctx context.Context, m *models.MetricChartConfigModel, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
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
				return metricChartConfigFromAPIVariant1(ctx, m, nil, nil, variant1)
			}
		}
		return metricChartConfigFromAPIVariant0(ctx, m, nil, nil, variant0)
	}

	variant1, err := attrs.AsMetricESQL()
	if err == nil {
		return metricChartConfigFromAPIVariant1(ctx, m, nil, nil, variant1)
	}

	diags.AddError("Failed to parse metric chart schema", "Could not parse as either variant 0 or 1")
	return diags
}

func metricChartConfigPopulateCommonFields(m *models.MetricChartConfigModel,
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

func metricChartConfigFromAPIVariant0(
	ctx context.Context,
	m *models.MetricChartConfigModel,
	dashboard *models.DashboardModel,
	prior *models.MetricChartConfigModel,
	apiChart kbapi.MetricNoESQL,
) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(apiChart.DataSource)
	if !metricChartConfigPopulateCommonFields(m, apiChart.Title, apiChart.Description, apiChart.IgnoreGlobalFilters, apiChart.Sampling, datasetBytes, datasetErr, apiChart.Filters, &diags) {
		return diags
	}

	m.Query = &models.FilterSimpleModel{}
	filterSimpleFromAPI(m.Query, apiChart.Query)

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
				populateMetricChartMetricDefaults,
			)
			if i < len(priorMetrics) && metricChartMetricConfigsEquivalent(priorMetrics[i].ConfigJSON, cfg) {
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

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lensDrilldownsAPIToWire(apiChart.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lensChartPresentationReadsFor(ctx, dashboard, priorLens, apiChart.TimeRange, apiChart.HideTitle, apiChart.HideBorder, apiChart.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres

	return diags
}

func metricChartConfigFromAPIVariant1(
	ctx context.Context,
	m *models.MetricChartConfigModel,
	dashboard *models.DashboardModel,
	prior *models.MetricChartConfigModel,
	apiChart kbapi.MetricESQL,
) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(apiChart.DataSource)
	if !metricChartConfigPopulateCommonFields(m, apiChart.Title, apiChart.Description, apiChart.IgnoreGlobalFilters, apiChart.Sampling, datasetBytes, datasetErr, apiChart.Filters, &diags) {
		return diags
	}

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
				populateMetricChartMetricDefaults,
			)
			if i < len(priorMetrics) && metricChartMetricConfigsEquivalent(priorMetrics[i].ConfigJSON, cfg) {
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

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lensDrilldownsAPIToWire(apiChart.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lensChartPresentationReadsFor(ctx, dashboard, priorLens, apiChart.TimeRange, apiChart.HideTitle, apiChart.HideBorder, apiChart.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres

	return diags
}

func metricChartConfigToAPI(m *models.MetricChartConfigModel, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	// Determine which variant to use based on whether we have a query
	// Variant 0 (non-ESQL) requires a query
	// Variant 1 (ESQL) doesn't require a query
	if m.Query != nil {
		return metricChartConfigToAPIVariant0(m, dashboard)
	}
	return metricChartConfigToAPIVariant1(m, dashboard)
}

func metricChartConfigToAPIVariant0(m *models.MetricChartConfigModel, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0

	variant0 := kbapi.MetricNoESQL{
		Type: kbapi.MetricNoESQLTypeMetric,
	}
	variant0.Styling = kbapi.MetricStyling{}

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
		variant0.Query = filterSimpleToAPI(m.Query)
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

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	variant0.TimeRange = writes.TimeRange
	if writes.HideTitle != nil {
		variant0.HideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		variant0.HideBorder = writes.HideBorder
	}
	if writes.References != nil {
		variant0.References = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := decodeLensDrilldownSlice[kbapi.MetricNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			variant0.Drilldowns = &items
		}
	}

	attrs, attrsDiags := metricChartAttrsFromPayload(variant0)
	diags.Append(attrsDiags...)
	return attrs, diags
}

func metricChartConfigToAPIVariant1(m *models.MetricChartConfigModel, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0

	variant1 := kbapi.MetricESQL{
		Type: kbapi.MetricESQLTypeMetric,
	}
	variant1.Styling = kbapi.MetricStyling{}

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

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	variant1.TimeRange = writes.TimeRange
	if writes.HideTitle != nil {
		variant1.HideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		variant1.HideBorder = writes.HideBorder
	}
	if writes.References != nil {
		variant1.References = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := decodeLensDrilldownSlice[kbapi.MetricESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			variant1.Drilldowns = &items
		}
	}

	attrs, attrsDiags := metricChartAttrsFromPayload(variant1)
	diags.Append(attrsDiags...)
	return attrs, diags
}
