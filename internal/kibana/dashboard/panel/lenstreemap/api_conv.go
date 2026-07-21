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

package lenstreemap

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func treemapConfigFromAPINoESQL(
	ctx context.Context,
	m *models.TreemapConfigModel,
	prior *models.TreemapConfigModel,
	api kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	snapshotIgnoreGlobalFilters := m.IgnoreGlobalFilters
	snapshotSampling := m.Sampling

	datasetBytes, datasetErr := api.DataSource.MarshalJSON()
	base, ok := lenscommon.PopulateLensChartBaseFromAPI(
		api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling,
		datasetBytes, datasetErr, "data_source_json", api.Filters, &diags,
	)
	if !ok {
		return diags
	}
	m.LensChartBaseTFModel = base
	m.IgnoreGlobalFilters = lenscommon.MapOptionalBoolWithSnapshotDefault(snapshotIgnoreGlobalFilters, api.IgnoreGlobalFilters, false)
	m.Sampling = lenscommon.MapOptionalFloatWithSnapshotDefault(snapshotSampling, api.Sampling, 1)

	if api.GroupBy != nil {
		gb, gbDiags := lenscommon.NewPartitionGroupByJSONFromAPI(api.GroupBy)
		diags.Append(gbDiags...)
		if !gbDiags.HasError() {
			m.GroupBy = gb
		}
	} else {
		m.GroupBy = customtypes.NewJSONWithDefaultsNull(lenscommon.PopulatePartitionGroupByDefaults)
	}

	metricsBytes, err := json.Marshal(api.Metrics)
	if err != nil {
		diags.AddError("Failed to marshal metrics", err.Error())
		return diags
	}
	m.Metrics = customtypes.NewJSONWithDefaultsValue(string(metricsBytes), lenscommon.PopulatePartitionMetricsDefaults)

	m.Query = &models.FilterSimpleModel{}
	lenscommon.FilterSimpleFromAPI(m.Query, api.Query)

	m.Legend = &models.PartitionLegendModel{}
	lenscommon.PartitionLegendFromTreemapLegend(m.Legend, api.Legend)

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

func treemapConfigFromAPIESQL(ctx context.Context, m *models.TreemapConfigModel, prior *models.TreemapConfigModel, api kbapi.KibanaHTTPAPIsTreemapESQLByValuePanel) diag.Diagnostics {
	var diags diag.Diagnostics

	// ES|QL charts don't have a query block. Clear it to avoid carrying over
	// query state from a previous non-ES|QL config.
	m.Query = nil

	snapshotIgnoreGlobalFilters := m.IgnoreGlobalFilters
	snapshotSampling := m.Sampling

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	base, ok := lenscommon.PopulateLensChartBaseFromAPI(
		api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling,
		datasetBytes, datasetErr, "data_source_json", api.Filters, &diags,
	)
	if !ok {
		return diags
	}
	m.LensChartBaseTFModel = base
	m.IgnoreGlobalFilters = lenscommon.MapOptionalBoolWithSnapshotDefault(snapshotIgnoreGlobalFilters, api.IgnoreGlobalFilters, false)
	m.Sampling = lenscommon.MapOptionalFloatWithSnapshotDefault(snapshotSampling, api.Sampling, 1)

	m.GroupBy = customtypes.NewJSONWithDefaultsNull(lenscommon.PopulatePartitionGroupByDefaults)
	m.Metrics = customtypes.NewJSONWithDefaultsNull(lenscommon.PopulatePartitionMetricsDefaults)

	if len(api.Metrics) > 0 {
		m.EsqlMetrics = make([]models.TreemapEsqlMetric, len(api.Metrics))
		for i, met := range api.Metrics {
			m.EsqlMetrics[i].Column = types.StringValue(met.Column)
			m.EsqlMetrics[i].Label = typeutils.StringishPointerValue(met.Label)
			formatVal, ok := lenscommon.LensESQLNumberFormatJSONFromAPI(met.Format, "esql_metrics.format_json", &diags)
			if !ok {
				continue
			}
			m.EsqlMetrics[i].FormatJSON = formatVal
			if met.Color != nil {
				staticColor, colorErr := met.Color.AsKibanaHTTPAPIsStaticColor()
				if colorErr == nil {
					m.EsqlMetrics[i].Color = &models.LensStaticColorModel{
						Type:  types.StringValue(string(staticColor.Type)),
						Color: types.StringValue(staticColor.Color),
					}
				}
			}
		}
	}

	if api.GroupBy != nil && len(*api.GroupBy) > 0 {
		src := make([]lenscommon.EsqlGroupByAPIFields, len(*api.GroupBy))
		for i, gb := range *api.GroupBy {
			src[i] = lenscommon.EsqlGroupByAPIFields{CollapseBy: gb.CollapseBy, Color: gb.Color, Column: gb.Column, Format: gb.Format, Label: gb.Label}
		}
		m.EsqlGroupBy = lenscommon.PopulatePartitionEsqlGroupByFromAPI(src, &diags)
	}

	m.Legend = &models.PartitionLegendModel{}
	lenscommon.PartitionLegendFromTreemapLegend(m.Legend, api.Legend)

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

func treemapConfigToAPI(m *models.TreemapConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var attrs lenscommon.VisByValueConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	if lenscommon.ConfigUsesESQL(m.Query) {
		esql, esqlDiags := treemapConfigToAPITreemapESQL(m)
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromKibanaHTTPAPIsTreemapESQLByValuePanel(esql); err != nil {
			diags.AddError("Failed to create treemap ES|QL schema", err.Error())
		}
		return attrs, diags
	}

	noESQL, noESQLDiags := treemapConfigToAPINoESQL(m)
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromKibanaHTTPAPIsTreemapNoESQLByValuePanel(noESQL); err != nil {
		diags.AddError("Failed to create treemap schema", err.Error())
	}

	return attrs, diags
}

func treemapConfigToAPITreemapESQL(m *models.TreemapConfigModel) (kbapi.KibanaHTTPAPIsTreemapESQLByValuePanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.KibanaHTTPAPIsTreemapESQLByValuePanel
	api.Type = kbapi.KibanaHTTPAPIsTreemapESQLByValuePanelTypeTreemap

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing data_source_json", "treemap_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal data_source_json", err.Error())
		return api, diags
	}
	if m.Legend == nil {
		diags.AddError("Missing legend", "treemap_config.legend must be provided")
		return api, diags
	}
	if len(m.EsqlMetrics) == 0 {
		diags.AddError("Missing esql_metrics", "treemap_config.esql_metrics must contain at least one item")
		return api, diags
	}
	if len(m.EsqlGroupBy) == 0 {
		diags.AddError("Missing esql_group_by", "treemap_config.esql_group_by must contain at least one item")
		return api, diags
	}

	api.Metrics = make([]struct {
		Color  *kbapi.KibanaHTTPAPIsTreemapESQLByValuePanel_Metrics_Color `json:"color,omitempty"`
		Column string                                                     `json:"column"`
		Format *kbapi.KibanaHTTPAPIsFormatType                            `json:"format,omitempty"`
		Label  *string                                                    `json:"label,omitempty"`
	}, len(m.EsqlMetrics))
	for i, em := range m.EsqlMetrics {
		api.Metrics[i].Column = em.Column.ValueString()
		if typeutils.IsKnown(em.Label) {
			l := em.Label.ValueString()
			api.Metrics[i].Label = &l
		}
		var format kbapi.KibanaHTTPAPIsFormatType
		if err := json.Unmarshal([]byte(em.FormatJSON.ValueString()), &format); err != nil {
			diags.AddError("Failed to unmarshal esql metric format_json", err.Error())
			return api, diags
		}
		api.Metrics[i].Format = &format
		if em.Color == nil {
			diags.AddError("Missing color", "treemap_config.esql_metrics color is required")
			return api, diags
		}
		staticColor := kbapi.KibanaHTTPAPIsStaticColor{
			Type:  kbapi.KibanaHTTPAPIsStaticColorType(em.Color.Type.ValueString()),
			Color: em.Color.Color.ValueString(),
		}
		var color kbapi.KibanaHTTPAPIsTreemapESQLByValuePanel_Metrics_Color
		if err := color.FromKibanaHTTPAPIsStaticColor(staticColor); err != nil {
			diags.AddError("Failed to marshal metric color", err.Error())
			return api, diags
		}
		api.Metrics[i].Color = &color
	}

	entries := lenscommon.BuildPartitionEsqlGroupByForAPI(m.EsqlGroupBy, &diags)
	if diags.HasError() {
		return api, diags
	}
	groupBy := lenscommon.BuildEsqlGroupBySliceForAPI[struct {
		CollapseBy *kbapi.KibanaHTTPAPIsCollapseBy   `json:"collapse_by,omitempty"`
		Color      *kbapi.KibanaHTTPAPIsColorMapping `json:"color,omitempty"`
		Column     string                            `json:"column"`
		Format     *kbapi.KibanaHTTPAPIsFormatType   `json:"format,omitempty"`
		Label      *string                           `json:"label,omitempty"`
	}](entries, &diags)
	if diags.HasError() {
		return api, diags
	}
	api.GroupBy = &groupBy

	api.Legend = lenscommon.PartitionLegendToTreemapLegend(m.Legend)

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if m.ValueDisplay != nil {
		api.Styling = &kbapi.KibanaHTTPAPIsTreemapStyling{Values: lenscommon.PartitionValueDisplayToAPI(m.ValueDisplay)}
	} else {
		defaultMode := kbapi.KibanaHTTPAPIsValueDisplayModePercentage
		api.Styling = &kbapi.KibanaHTTPAPIsTreemapStyling{
			Values: &kbapi.KibanaHTTPAPIsValueDisplay{Mode: &defaultMode},
		}
	}

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsTreemapESQLByValuePanel_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	return api, diags
}

func treemapConfigToAPINoESQL(m *models.TreemapConfigModel) (kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanel{
		Type: kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanelTypeTreemap,
	}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing data_source_json", "treemap_config.data_source_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
		diags.AddError("Failed to unmarshal data_source_json", err.Error())
		return api, diags
	}

	if m.GroupBy.IsNull() {
		diags.AddError("Missing group_by_json", "treemap_config.group_by_json must be provided")
		return api, diags
	}
	var groupBy []kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanel_GroupBy_Item
	if err := json.Unmarshal([]byte(m.GroupBy.ValueString()), &groupBy); err != nil {
		diags.AddError("Failed to unmarshal group_by", err.Error())
		return api, diags
	}
	if len(groupBy) == 0 {
		diags.AddError("Invalid group_by_json", "treemap_config.group_by_json must contain at least one item")
		return api, diags
	}
	api.GroupBy = &groupBy

	if m.Metrics.IsNull() {
		diags.AddError("Missing metrics_json", "treemap_config.metrics_json must be provided")
		return api, diags
	}
	var metrics []kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanel_Metrics_Item
	if err := json.Unmarshal([]byte(m.Metrics.ValueString()), &metrics); err != nil {
		diags.AddError("Failed to unmarshal metrics", err.Error())
		return api, diags
	}
	if len(metrics) == 0 {
		diags.AddError("Invalid metrics_json", "treemap_config.metrics_json must contain at least one item")
		return api, diags
	}
	api.Metrics = metrics

	if m.Query == nil {
		diags.AddError("Missing query", "treemap_config.query is required for non-ES|QL treemap charts")
		return api, diags
	}
	api.Query = lenscommon.FilterSimpleToAPI(m.Query)

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if m.Legend == nil {
		diags.AddError("Missing legend", "treemap_config.legend must be provided")
		return api, diags
	}
	api.Legend = lenscommon.PartitionLegendToTreemapLegend(m.Legend)

	if m.ValueDisplay != nil {
		api.Styling = &kbapi.KibanaHTTPAPIsTreemapStyling{Values: lenscommon.PartitionValueDisplayToAPI(m.ValueDisplay)}
	}

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsTreemapNoESQLByValuePanel_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	return api, diags
}
