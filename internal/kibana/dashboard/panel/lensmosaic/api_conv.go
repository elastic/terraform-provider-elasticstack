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

package lensmosaic

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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func isMosaicNoESQLCandidateActuallyESQL(api kbapi.KibanaHTTPAPIsMosaicNoESQL) bool {
	return lenscommon.LensDataSourceIsESQLOrTable(api.DataSource.MarshalJSON())
}

func mosaicConfigFromAPINoESQL(ctx context.Context, m *models.MosaicConfigModel, prior *models.MosaicConfigModel, api kbapi.KibanaHTTPAPIsMosaicNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = lenscommon.MapOptionalBoolWithSnapshotDefault(m.IgnoreGlobalFilters, api.IgnoreGlobalFilters, false)
	m.Sampling = lenscommon.MapOptionalFloatWithSnapshotDefault(m.Sampling, api.Sampling, 1)

	datasetBytes, err := api.DataSource.MarshalJSON()
	dv, ok := lenscommon.WrapNormalizedJSON(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = dv

	if api.GroupBy != nil {
		gb, gbDiags := lenscommon.NewPartitionGroupByJSONFromAPI(api.GroupBy)
		diags.Append(gbDiags...)
		if !gbDiags.HasError() {
			m.GroupBy = gb
		}
	} else {
		m.GroupBy = customtypes.NewJSONWithDefaultsNull(lenscommon.PopulatePartitionGroupByDefaults)
	}

	if api.GroupBreakdownBy != nil {
		gbb, gbbDiags := lenscommon.NewPartitionGroupByJSONFromAPI(api.GroupBreakdownBy)
		diags.Append(gbbDiags...)
		if !gbbDiags.HasError() {
			m.GroupBreakdownBy = gbb
		}
	} else {
		m.GroupBreakdownBy = customtypes.NewJSONWithDefaultsNull(lenscommon.PopulatePartitionGroupByDefaults)
	}

	metricBytes, err := api.Metric.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal metric", err.Error())
		return diags
	}
	metricsWrapped, err := json.Marshal([]json.RawMessage{json.RawMessage(metricBytes)})
	if err != nil {
		diags.AddError("Failed to marshal metrics_json", err.Error())
		return diags
	}
	m.Metrics = customtypes.NewJSONWithDefaultsValue(string(metricsWrapped), lenscommon.PopulatePartitionMetricsDefaults)

	m.Query = &models.FilterSimpleModel{}
	lenscommon.FilterSimpleFromAPI(m.Query, api.Query)

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = lenscommon.PopulateFiltersFromAPI(api.Filters, &diags)
	} else {
		m.Filters = nil
	}

	m.Legend = &models.PartitionLegendModel{}
	lenscommon.PartitionLegendFromMosaicLegend(m.Legend, api.Legend)

	if api.Styling != nil && api.Styling.Values != nil && (api.Styling.Values.Mode != nil || api.Styling.Values.PercentDecimals != nil) {
		m.ValueDisplay = &models.PartitionValueDisplay{}
		lenscommon.PartitionValueDisplayFromAPI(m.ValueDisplay, api.Styling.Values)
	} else {
		m.ValueDisplay = nil
	}

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}
	m.EsqlMetrics = nil
	m.EsqlGroupBy = nil

	return diags
}

func mosaicConfigFromAPIESQL(ctx context.Context, m *models.MosaicConfigModel, prior *models.MosaicConfigModel, api kbapi.KibanaHTTPAPIsMosaicESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Query = nil

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = lenscommon.MapOptionalBoolWithSnapshotDefault(m.IgnoreGlobalFilters, api.IgnoreGlobalFilters, false)
	m.Sampling = lenscommon.MapOptionalFloatWithSnapshotDefault(m.Sampling, api.Sampling, 1)

	datasetBytes, err := json.Marshal(api.DataSource)
	dv, ok := lenscommon.WrapNormalizedJSON(datasetBytes, err, "data_source_json", &diags)
	if !ok {
		return diags
	}
	m.DataSourceJSON = dv

	m.GroupBy = customtypes.NewJSONWithDefaultsNull(lenscommon.PopulatePartitionGroupByDefaults)
	m.Metrics = customtypes.NewJSONWithDefaultsNull(lenscommon.PopulatePartitionMetricsDefaults)

	if api.GroupBreakdownBy != nil {
		gbb, gbbDiags := lenscommon.NewPartitionGroupByJSONFromAPI(api.GroupBreakdownBy)
		diags.Append(gbbDiags...)
		if !gbbDiags.HasError() {
			m.GroupBreakdownBy = gbb
		}
	} else {
		m.GroupBreakdownBy = customtypes.NewJSONWithDefaultsNull(lenscommon.PopulatePartitionGroupByDefaults)
	}

	metricFormat, ok := lenscommon.LensESQLNumberFormatJSONFromAPI(api.Metric.Format, "esql_metrics.format_json", &diags)
	if !ok {
		return diags
	}
	m.EsqlMetrics = []models.MosaicEsqlMetric{
		{
			Column:     types.StringValue(api.Metric.Column),
			FormatJSON: metricFormat,
			Label:      types.StringNull(),
		},
	}
	if api.Metric.Label != nil {
		m.EsqlMetrics[0].Label = types.StringValue(*api.Metric.Label)
	}

	if api.GroupBy != nil && len(*api.GroupBy) > 0 {
		m.EsqlGroupBy = make([]models.PartitionEsqlGroupByModel, len(*api.GroupBy))
		for i, gb := range *api.GroupBy {
			collapseBy := ""
			if gb.CollapseBy != nil {
				collapseBy = string(*gb.CollapseBy)
			}
			m.EsqlGroupBy[i].Column = types.StringValue(gb.Column)
			m.EsqlGroupBy[i].CollapseBy = types.StringValue(collapseBy)
			colorBytes, err := json.Marshal(gb.Color)
			if err != nil {
				diags.AddError("Failed to marshal esql group_by color", err.Error())
				continue
			}
			m.EsqlGroupBy[i].ColorJSON = jsontypes.NewNormalizedValue(string(colorBytes))
			formatBytes, err := json.Marshal(gb.Format)
			if err != nil {
				diags.AddError("Failed to marshal esql group_by format", err.Error())
				continue
			}
			m.EsqlGroupBy[i].FormatJSON = jsontypes.NewNormalizedValue(string(formatBytes))
			m.EsqlGroupBy[i].Label = typeutils.StringishPointerValue(gb.Label)
		}
	}

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = lenscommon.PopulateFiltersFromAPI(api.Filters, &diags)
	} else {
		m.Filters = nil
	}

	m.Legend = &models.PartitionLegendModel{}
	lenscommon.PartitionLegendFromMosaicLegend(m.Legend, api.Legend)

	if api.Styling != nil && api.Styling.Values != nil && (api.Styling.Values.Mode != nil || api.Styling.Values.PercentDecimals != nil) {
		m.ValueDisplay = &models.PartitionValueDisplay{}
		lenscommon.PartitionValueDisplayFromAPI(m.ValueDisplay, api.Styling.Values)
	} else {
		m.ValueDisplay = nil
	}

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}

	return diags
}

func mosaicConfigToAPI(m *models.MosaicConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var attrs lenscommon.VisByValueConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	if lenscommon.ConfigUsesESQL(m.Query) {
		esql, esqlDiags := mosaicConfigToAPIMosaicESQL(m)
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromKibanaHTTPAPIsMosaicESQL(esql); err != nil {
			diags.AddError("Failed to create mosaic ES|QL schema", err.Error())
		}
		return attrs, diags
	}

	noESQL, noESQLDiags := mosaicConfigToAPINoESQL(m)
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromKibanaHTTPAPIsMosaicNoESQL(noESQL); err != nil {
		diags.AddError("Failed to create mosaic schema", err.Error())
	}

	return attrs, diags
}

func mosaicConfigToAPIMosaicESQL(m *models.MosaicConfigModel) (kbapi.KibanaHTTPAPIsMosaicESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.KibanaHTTPAPIsMosaicESQL
	api.Type = kbapi.KibanaHTTPAPIsMosaicESQLTypeMosaic

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing data_source_json", "mosaic_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal data_source_json", err.Error())
		return api, diags
	}
	if m.Legend == nil {
		diags.AddError("Missing legend", "mosaic_config.legend must be provided")
		return api, diags
	}
	api.Legend = lenscommon.PartitionLegendToMosaicLegend(m.Legend)

	if m.GroupBreakdownBy.IsNull() {
		diags.AddError("Missing group_breakdown_by_json", "mosaic_config.group_breakdown_by_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.GroupBreakdownBy.ValueString()), &api.GroupBreakdownBy); err != nil {
		diags.AddError("Failed to unmarshal group_breakdown_by", err.Error())
		return api, diags
	}

	if len(m.EsqlMetrics) != 1 {
		diags.AddError("Invalid esql_metrics", "mosaic_config.esql_metrics must contain exactly one item")
		return api, diags
	}
	em := m.EsqlMetrics[0]
	api.Metric.Column = em.Column.ValueString()
	if typeutils.IsKnown(em.Label) {
		l := em.Label.ValueString()
		api.Metric.Label = &l
	}
	if err := json.Unmarshal([]byte(em.FormatJSON.ValueString()), &api.Metric.Format); err != nil {
		diags.AddError("Failed to unmarshal esql metric format_json", err.Error())
		return api, diags
	}

	if len(m.EsqlGroupBy) == 0 {
		diags.AddError("Missing esql_group_by", "mosaic_config.esql_group_by must contain at least one item")
		return api, diags
	}
	groupBy := make([]struct {
		CollapseBy *kbapi.KibanaHTTPAPIsCollapseBy   `json:"collapse_by,omitempty"`
		Color      *kbapi.KibanaHTTPAPIsColorMapping `json:"color,omitempty"`
		Column     string                            `json:"column"`
		Format     *kbapi.KibanaHTTPAPIsFormatType   `json:"format,omitempty"`
		Label      *string                           `json:"label,omitempty"`
	}, len(m.EsqlGroupBy))
	for i, eg := range m.EsqlGroupBy {
		groupBy[i].Column = eg.Column.ValueString()
		collapseBy := kbapi.KibanaHTTPAPIsCollapseBy(eg.CollapseBy.ValueString())
		groupBy[i].CollapseBy = &collapseBy
		var color kbapi.KibanaHTTPAPIsColorMapping
		if err := json.Unmarshal([]byte(eg.ColorJSON.ValueString()), &color); err != nil {
			diags.AddError("Failed to unmarshal esql group_by color_json", err.Error())
			return api, diags
		}
		groupBy[i].Color = &color
		formatSrc := lenscommon.DefaultLensNumberFormatJSON
		if typeutils.IsKnown(eg.FormatJSON) {
			formatSrc = eg.FormatJSON.ValueString()
		}
		var format kbapi.KibanaHTTPAPIsFormatType
		if err := json.Unmarshal([]byte(formatSrc), &format); err != nil {
			diags.AddError("Failed to unmarshal esql group_by format_json", err.Error())
			return api, diags
		}
		groupBy[i].Format = &format
		if typeutils.IsKnown(eg.Label) {
			l := eg.Label.ValueString()
			groupBy[i].Label = &l
		}
	}
	api.GroupBy = &groupBy

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if m.ValueDisplay != nil {
		api.Styling = &kbapi.KibanaHTTPAPIsMosaicStyling{Values: lenscommon.PartitionValueDisplayToAPI(m.ValueDisplay)}
	} else {
		defaultMode := kbapi.KibanaHTTPAPIsValueDisplayModePercentage
		api.Styling = &kbapi.KibanaHTTPAPIsMosaicStyling{
			Values: &kbapi.KibanaHTTPAPIsValueDisplay{Mode: &defaultMode},
		}
	}

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsMosaicESQL_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	return api, diags
}

func mosaicConfigToAPINoESQL(m *models.MosaicConfigModel) (kbapi.KibanaHTTPAPIsMosaicNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KibanaHTTPAPIsMosaicNoESQL{
		Type: kbapi.KibanaHTTPAPIsMosaicNoESQLTypeMosaic,
	}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing data_source_json", "mosaic_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal data_source_json", err.Error())
		return api, diags
	}

	if m.GroupBy.IsNull() {
		diags.AddError("Missing group_by_json", "mosaic_config.group_by_json must be provided")
		return api, diags
	}
	var groupBy []kbapi.KibanaHTTPAPIsMosaicNoESQL_GroupBy_Item
	if err := json.Unmarshal([]byte(m.GroupBy.ValueString()), &groupBy); err != nil {
		diags.AddError("Failed to unmarshal group_by", err.Error())
		return api, diags
	}
	if len(groupBy) == 0 {
		diags.AddError("Invalid group_by_json", "mosaic_config.group_by_json must contain at least one item")
		return api, diags
	}
	api.GroupBy = &groupBy

	if m.GroupBreakdownBy.IsNull() {
		diags.AddError("Missing group_breakdown_by_json", "mosaic_config.group_breakdown_by_json must be provided")
		return api, diags
	}
	var groupBreakdownBy []kbapi.KibanaHTTPAPIsMosaicNoESQL_GroupBreakdownBy_Item
	if err := json.Unmarshal([]byte(m.GroupBreakdownBy.ValueString()), &groupBreakdownBy); err != nil {
		diags.AddError("Failed to unmarshal group_breakdown_by", err.Error())
		return api, diags
	}
	if len(groupBreakdownBy) == 0 {
		diags.AddError("Invalid group_breakdown_by_json", "mosaic_config.group_breakdown_by_json must contain at least one item")
		return api, diags
	}
	api.GroupBreakdownBy = &groupBreakdownBy

	if m.Metrics.IsNull() {
		diags.AddError("Missing metrics_json", "mosaic_config.metrics_json must be provided")
		return api, diags
	}
	var rawMetrics []json.RawMessage
	if err := json.Unmarshal([]byte(m.Metrics.ValueString()), &rawMetrics); err != nil {
		diags.AddError("Failed to unmarshal metrics_json", err.Error())
		return api, diags
	}
	if len(rawMetrics) != 1 {
		diags.AddError("Invalid metrics_json", "mosaic_config.metrics_json must contain exactly one item")
		return api, diags
	}
	if err := json.Unmarshal(rawMetrics[0], &api.Metric); err != nil {
		diags.AddError("Failed to unmarshal metric", err.Error())
		return api, diags
	}

	if m.Query == nil {
		diags.AddError("Missing query", "mosaic_config.query is required for non-ES|QL mosaic charts")
		return api, diags
	}
	api.Query = lenscommon.FilterSimpleToAPI(m.Query)

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if m.Legend == nil {
		diags.AddError("Missing legend", "mosaic_config.legend must be provided")
		return api, diags
	}
	api.Legend = lenscommon.PartitionLegendToMosaicLegend(m.Legend)

	if m.ValueDisplay != nil {
		api.Styling = &kbapi.KibanaHTTPAPIsMosaicStyling{Values: lenscommon.PartitionValueDisplayToAPI(m.ValueDisplay)}
	}

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsMosaicNoESQL_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	return api, diags
}
