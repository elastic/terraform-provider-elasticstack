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

func newTreemapPanelConfigConverter() treemapPanelConfigConverter {
	return treemapPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.TreemapNoESQLTypeTreemap),
			hasTFChartBlock: func(blocks *models.LensByValueChartBlocks) bool {
				return blocks != nil && blocks.TreemapConfig != nil
			},
		},
	}
}

type treemapPanelConfigConverter struct {
	lensVisualizationBase
}

func (c treemapPanelConfigConverter) populateFromAttributes(
	ctx context.Context,
	dashboard *models.DashboardModel,
	tfPanel *models.PanelModel,
	blocks *models.LensByValueChartBlocks,
	attrs kbapi.KbnDashboardPanelTypeVisConfig0,
) diag.Diagnostics {
	var prior *models.TreemapConfigModel
	if b := lensByValueChartBlocksFromPanel(tfPanel); b != nil && b.TreemapConfig != nil {
		cpy := *b.TreemapConfig
		prior = &cpy
	}
	blocks.TreemapConfig = &models.TreemapConfigModel{}

	if noESQL, err := attrs.AsTreemapNoESQL(); err == nil && !isTreemapNoESQLCandidateActuallyESQL(noESQL) {
		return treemapConfigFromAPINoESQL(ctx, blocks.TreemapConfig, dashboard, prior, noESQL)
	}

	treemapESQL, err := attrs.AsTreemapESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return treemapConfigFromAPIESQL(ctx, blocks.TreemapConfig, dashboard, prior, treemapESQL)
}

func (c treemapPanelConfigConverter) buildAttributes(blocks *models.LensByValueChartBlocks, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *blocks.TreemapConfig

	attrs, treemapDiags := treemapConfigToAPI(&configModel, dashboard)
	diags.Append(treemapDiags...)
	return attrs, diags
}

func isTreemapNoESQLCandidateActuallyESQL(api kbapi.TreemapNoESQL) bool {
	body, err := api.DataSource.MarshalJSON()
	if err != nil {
		return false
	}
	var ds struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &ds); err != nil {
		return false
	}
	return ds.Type == legacyMetricDatasetTypeESQL || ds.Type == legacyMetricDatasetTypeTable
}

func treemapConfigFromAPINoESQL(ctx context.Context, m *models.TreemapConfigModel, dashboard *models.DashboardModel, prior *models.TreemapConfigModel, api kbapi.TreemapNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = mapOptionalBoolWithSnapshotDefault(m.IgnoreGlobalFilters, api.IgnoreGlobalFilters, false)
	m.Sampling = mapOptionalFloatWithSnapshotDefault(m.Sampling, api.Sampling, 1)

	datasetBytes, err := api.DataSource.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal data_source_json", err.Error())
		return diags
	}
	m.DataSourceJSON = jsontypes.NewNormalizedValue(string(datasetBytes))

	if api.GroupBy != nil {
		gb, gbDiags := newPartitionGroupByJSONFromAPI(api.GroupBy)
		diags.Append(gbDiags...)
		if !gbDiags.HasError() {
			m.GroupBy = gb
		}
	} else {
		m.GroupBy = customtypes.NewJSONWithDefaultsNull(populatePartitionGroupByDefaults)
	}

	metricsBytes, err := json.Marshal(api.Metrics)
	if err != nil {
		diags.AddError("Failed to marshal metrics", err.Error())
		return diags
	}
	m.Metrics = customtypes.NewJSONWithDefaultsValue(string(metricsBytes), populatePartitionMetricsDefaults)

	m.Query = &models.FilterSimpleModel{}
	filterSimpleFromAPI(m.Query, api.Query)

	if len(api.Filters) > 0 {
		m.Filters = populateFiltersFromAPI(api.Filters, &diags)
	} else {
		m.Filters = nil
	}

	m.Legend = &models.PartitionLegendModel{}
	partitionLegendFromTreemapLegend(m.Legend, api.Legend)

	if api.Styling.Values.Mode != nil || api.Styling.Values.PercentDecimals != nil {
		m.ValueDisplay = &models.PartitionValueDisplay{}
		partitionValueDisplayFromValueDisplay(m.ValueDisplay, api.Styling.Values)
	} else {
		m.ValueDisplay = nil
	}

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lensDrilldownsAPIToWire(api.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lensChartPresentationReadsFor(ctx, dashboard, priorLens, api.TimeRange, api.HideTitle, api.HideBorder, api.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres
	m.EsqlMetrics = nil
	m.EsqlGroupBy = nil

	return diags
}

func treemapConfigFromAPIESQL(ctx context.Context, m *models.TreemapConfigModel, dashboard *models.DashboardModel, prior *models.TreemapConfigModel, api kbapi.TreemapESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	// ES|QL charts don't have a query block. Clear it to avoid carrying over
	// query state from a previous non-ES|QL config.
	m.Query = nil

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = mapOptionalBoolWithSnapshotDefault(m.IgnoreGlobalFilters, api.IgnoreGlobalFilters, false)
	m.Sampling = mapOptionalFloatWithSnapshotDefault(m.Sampling, api.Sampling, 1)

	datasetBytes, err := json.Marshal(api.DataSource)
	if err != nil {
		diags.AddError("Failed to marshal data_source_json", err.Error())
		return diags
	}
	m.DataSourceJSON = jsontypes.NewNormalizedValue(string(datasetBytes))

	m.GroupBy = customtypes.NewJSONWithDefaultsNull(populatePartitionGroupByDefaults)
	m.Metrics = customtypes.NewJSONWithDefaultsNull(populatePartitionMetricsDefaults)

	if len(api.Metrics) > 0 {
		m.EsqlMetrics = make([]models.TreemapEsqlMetric, len(api.Metrics))
		for i, met := range api.Metrics {
			m.EsqlMetrics[i].Column = types.StringValue(met.Column)
			if met.Label != nil {
				m.EsqlMetrics[i].Label = types.StringValue(*met.Label)
			} else {
				m.EsqlMetrics[i].Label = types.StringNull()
			}
			formatVal, ok := lensESQLNumberFormatJSONFromAPI(met.Format, "esql_metrics.format_json", &diags)
			if !ok {
				continue
			}
			m.EsqlMetrics[i].FormatJSON = formatVal
			if met.Color != nil {
				staticColor, colorErr := met.Color.AsStaticColor()
				if colorErr == nil {
					m.EsqlMetrics[i].Color = &models.TreemapEsqlMetricColor{
						Type:  types.StringValue(string(staticColor.Type)),
						Color: types.StringValue(staticColor.Color),
					}
				}
			}
		}
	}

	if api.GroupBy != nil && len(*api.GroupBy) > 0 {
		m.EsqlGroupBy = make([]models.TreemapEsqlGroupBy, len(*api.GroupBy))
		for i, gb := range *api.GroupBy {
			m.EsqlGroupBy[i].Column = types.StringValue(gb.Column)
			m.EsqlGroupBy[i].CollapseBy = types.StringValue(string(gb.CollapseBy))
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
			if gb.Label != nil {
				m.EsqlGroupBy[i].Label = types.StringValue(*gb.Label)
			} else {
				m.EsqlGroupBy[i].Label = types.StringNull()
			}
		}
	}

	if len(api.Filters) > 0 {
		m.Filters = populateFiltersFromAPI(api.Filters, &diags)
	} else {
		m.Filters = nil
	}

	m.Legend = &models.PartitionLegendModel{}
	partitionLegendFromTreemapLegend(m.Legend, api.Legend)

	if api.Styling.Values.Mode != nil || api.Styling.Values.PercentDecimals != nil {
		m.ValueDisplay = &models.PartitionValueDisplay{}
		partitionValueDisplayFromValueDisplay(m.ValueDisplay, api.Styling.Values)
	} else {
		m.ValueDisplay = nil
	}

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lensDrilldownsAPIToWire(api.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lensChartPresentationReadsFor(ctx, dashboard, priorLens, api.TimeRange, api.HideTitle, api.HideBorder, api.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres

	return diags
}

func treemapConfigToAPI(m *models.TreemapConfigModel, dashboard *models.DashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	if treemapConfigUsesESQL(m) {
		esql, esqlDiags := treemapConfigToAPITreemapESQL(m, dashboard)
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromTreemapESQL(esql); err != nil {
			diags.AddError("Failed to create treemap ES|QL schema", err.Error())
		}
		return attrs, diags
	}

	noESQL, noESQLDiags := treemapConfigToAPINoESQL(m, dashboard)
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromTreemapNoESQL(noESQL); err != nil {
		diags.AddError("Failed to create treemap schema", err.Error())
	}

	return attrs, diags
}

func treemapConfigToAPITreemapESQL(m *models.TreemapConfigModel, dashboard *models.DashboardModel) (kbapi.TreemapESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.TreemapESQL
	api.Type = kbapi.TreemapESQLTypeTreemap

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
		Color  *kbapi.TreemapESQL_Metrics_Color `json:"color,omitempty"`
		Column string                           `json:"column"`
		Format kbapi.FormatType                 `json:"format"`
		Label  *string                          `json:"label,omitempty"`
	}, len(m.EsqlMetrics))
	for i, em := range m.EsqlMetrics {
		api.Metrics[i].Column = em.Column.ValueString()
		if typeutils.IsKnown(em.Label) {
			l := em.Label.ValueString()
			api.Metrics[i].Label = &l
		}
		if err := json.Unmarshal([]byte(em.FormatJSON.ValueString()), &api.Metrics[i].Format); err != nil {
			diags.AddError("Failed to unmarshal esql metric format_json", err.Error())
			return api, diags
		}
		if em.Color == nil {
			diags.AddError("Missing color", "treemap_config.esql_metrics color is required")
			return api, diags
		}
		staticColor := kbapi.StaticColor{
			Type:  kbapi.StaticColorType(em.Color.Type.ValueString()),
			Color: em.Color.Color.ValueString(),
		}
		var color kbapi.TreemapESQL_Metrics_Color
		if err := color.FromStaticColor(staticColor); err != nil {
			diags.AddError("Failed to marshal metric color", err.Error())
			return api, diags
		}
		api.Metrics[i].Color = &color
	}

	groupBy := make([]struct {
		CollapseBy kbapi.CollapseBy   `json:"collapse_by"`
		Color      kbapi.ColorMapping `json:"color"`
		Column     string             `json:"column"`
		Format     kbapi.FormatType   `json:"format"`
		Label      *string            `json:"label,omitempty"`
	}, len(m.EsqlGroupBy))
	for i, eg := range m.EsqlGroupBy {
		groupBy[i].Column = eg.Column.ValueString()
		groupBy[i].CollapseBy = kbapi.CollapseBy(eg.CollapseBy.ValueString())
		if err := json.Unmarshal([]byte(eg.ColorJSON.ValueString()), &groupBy[i].Color); err != nil {
			diags.AddError("Failed to unmarshal esql group_by color_json", err.Error())
			return api, diags
		}
		formatSrc := defaultNumberFormatJSON
		if typeutils.IsKnown(eg.FormatJSON) {
			formatSrc = eg.FormatJSON.ValueString()
		}
		if err := json.Unmarshal([]byte(formatSrc), &groupBy[i].Format); err != nil {
			diags.AddError("Failed to unmarshal esql group_by format_json", err.Error())
			return api, diags
		}
		if typeutils.IsKnown(eg.Label) {
			l := eg.Label.ValueString()
			groupBy[i].Label = &l
		}
	}
	api.GroupBy = &groupBy

	api.Legend = partitionLegendToTreemapLegend(m.Legend)

	if typeutils.IsKnown(m.Title) {
		api.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = new(m.Description.ValueString())
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = new(m.IgnoreGlobalFilters.ValueBool())
	}
	if typeutils.IsKnown(m.Sampling) {
		api.Sampling = new(float32(m.Sampling.ValueFloat64()))
	}

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if m.ValueDisplay != nil {
		api.Styling.Values = partitionValueDisplayToValueDisplay(m.ValueDisplay)
	} else {
		defaultMode := kbapi.ValueDisplayModePercentage
		api.Styling.Values = kbapi.ValueDisplay{Mode: &defaultMode}
	}

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	api.TimeRange = writes.TimeRange
	if writes.HideTitle != nil {
		api.HideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		api.HideBorder = writes.HideBorder
	}
	if writes.References != nil {
		api.References = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := decodeLensDrilldownSlice[kbapi.TreemapESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}

func treemapConfigUsesESQL(m *models.TreemapConfigModel) bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Expression.IsNull() && m.Query.Language.IsNull()
}

func treemapConfigToAPINoESQL(m *models.TreemapConfigModel, dashboard *models.DashboardModel) (kbapi.TreemapNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.TreemapNoESQL{
		Type: kbapi.TreemapNoESQLTypeTreemap,
	}

	if typeutils.IsKnown(m.Title) {
		api.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = new(m.Description.ValueString())
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = new(m.IgnoreGlobalFilters.ValueBool())
	}
	if typeutils.IsKnown(m.Sampling) {
		api.Sampling = new(float32(m.Sampling.ValueFloat64()))
	}

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
	var groupBy []kbapi.TreemapNoESQL_GroupBy_Item
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
	var metrics []kbapi.TreemapNoESQL_Metrics_Item
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
	api.Query = filterSimpleToAPI(m.Query)

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if m.Legend == nil {
		diags.AddError("Missing legend", "treemap_config.legend must be provided")
		return api, diags
	}
	api.Legend = partitionLegendToTreemapLegend(m.Legend)

	if m.ValueDisplay != nil {
		api.Styling.Values = partitionValueDisplayToValueDisplay(m.ValueDisplay)
	}

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return api, diags
	}

	api.TimeRange = writes.TimeRange
	if writes.HideTitle != nil {
		api.HideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		api.HideBorder = writes.HideBorder
	}
	if writes.References != nil {
		api.References = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := decodeLensDrilldownSlice[kbapi.TreemapNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	return api, diags
}
