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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
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
			hasTFChartBlock: func(blocks *lensByValueChartBlocks) bool {
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
	dashboard *dashboardModel,
	tfPanel *panelModel,
	blocks *lensByValueChartBlocks,
	attrs kbapi.KbnDashboardPanelTypeVisConfig0,
) diag.Diagnostics {
	var prior *treemapConfigModel
	if b := lensByValueChartBlocksFromPanel(tfPanel); b != nil && b.TreemapConfig != nil {
		cpy := *b.TreemapConfig
		prior = &cpy
	}
	blocks.TreemapConfig = &treemapConfigModel{}

	if noESQL, err := attrs.AsTreemapNoESQL(); err == nil && !isTreemapNoESQLCandidateActuallyESQL(noESQL) {
		return blocks.TreemapConfig.fromAPINoESQL(ctx, dashboard, prior, noESQL)
	}

	treemapESQL, err := attrs.AsTreemapESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return blocks.TreemapConfig.fromAPIESQL(ctx, dashboard, prior, treemapESQL)
}

func (c treemapPanelConfigConverter) buildAttributes(blocks *lensByValueChartBlocks, dashboard *dashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *blocks.TreemapConfig

	attrs, treemapDiags := configModel.toAPI(dashboard)
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

type treemapConfigModel struct {
	lensChartPresentationTFModel
	Title               types.String                                        `tfsdk:"title"`
	Description         types.String                                        `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized                                `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool                                          `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                       `tfsdk:"sampling"`
	Query               *filterSimpleModel                                  `tfsdk:"query"`
	Filters             []chartFilterJSONModel                              `tfsdk:"filters"`
	GroupBy             customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"group_by_json"`
	Metrics             customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"metrics_json"`
	Legend              *partitionLegendModel                               `tfsdk:"legend"`
	ValueDisplay        *partitionValueDisplay                              `tfsdk:"value_display"`
}

func (m *treemapConfigModel) fromAPINoESQL(ctx context.Context, dashboard *dashboardModel, prior *treemapConfigModel, api kbapi.TreemapNoESQL) diag.Diagnostics {
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

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	if len(api.Filters) > 0 {
		m.Filters = populateFiltersFromAPI(api.Filters, &diags)
	} else {
		m.Filters = nil
	}

	m.Legend = &partitionLegendModel{}
	m.Legend.fromTreemapLegend(api.Legend)

	if api.Styling.Values.Mode != nil || api.Styling.Values.PercentDecimals != nil {
		m.ValueDisplay = &partitionValueDisplay{}
		m.ValueDisplay.fromValueDisplay(api.Styling.Values)
	} else {
		m.ValueDisplay = nil
	}

	var priorLens *lensChartPresentationTFModel
	if prior != nil {
		p := prior.lensChartPresentationTFModel
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
	m.lensChartPresentationTFModel = pres

	return diags
}

func (m *treemapConfigModel) fromAPIESQL(ctx context.Context, dashboard *dashboardModel, prior *treemapConfigModel, api kbapi.TreemapESQL) diag.Diagnostics {
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

	if len(api.Filters) > 0 {
		m.Filters = populateFiltersFromAPI(api.Filters, &diags)
	} else {
		m.Filters = nil
	}

	m.Legend = &partitionLegendModel{}
	m.Legend.fromTreemapLegend(api.Legend)

	if api.Styling.Values.Mode != nil || api.Styling.Values.PercentDecimals != nil {
		m.ValueDisplay = &partitionValueDisplay{}
		m.ValueDisplay.fromValueDisplay(api.Styling.Values)
	} else {
		m.ValueDisplay = nil
	}

	var priorLens *lensChartPresentationTFModel
	if prior != nil {
		p := prior.lensChartPresentationTFModel
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
	m.lensChartPresentationTFModel = pres

	return diags
}

func (m *treemapConfigModel) toAPI(dashboard *dashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	if m.usesESQL() {
		esql, esqlDiags := m.toAPITreemapESQL(dashboard)
		diags.Append(esqlDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromTreemapESQL(esql); err != nil {
			diags.AddError("Failed to create treemap ES|QL schema", err.Error())
		}
		return attrs, diags
	}

	noESQL, noESQLDiags := m.toAPINoESQL(dashboard)
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromTreemapNoESQL(noESQL); err != nil {
		diags.AddError("Failed to create treemap schema", err.Error())
	}

	return attrs, diags
}

func (m *treemapConfigModel) toAPITreemapESQL(dashboard *dashboardModel) (kbapi.TreemapESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.TreemapESQL

	if m.DataSourceJSON.IsNull() {
		diags.AddError("Missing data_source_json", "treemap_config.data_source_json must be provided")
		return api, diags
	}
	if m.GroupBy.IsNull() {
		diags.AddError("Missing group_by_json", "treemap_config.group_by_json must be provided")
		return api, diags
	}
	if m.Metrics.IsNull() {
		diags.AddError("Missing metrics_json", "treemap_config.metrics_json must be provided")
		return api, diags
	}
	if m.Legend == nil {
		diags.AddError("Missing legend", "treemap_config.legend must be provided")
		return api, diags
	}
	if diags.HasError() {
		return api, diags
	}

	mergeJSON := fmt.Sprintf(
		`{"type":"treemap","data_source":%s,"group_by":%s,"metrics":%s}`,
		m.DataSourceJSON.ValueString(),
		m.GroupBy.ValueString(),
		m.Metrics.ValueString(),
	)
	if err := json.Unmarshal([]byte(mergeJSON), &api); err != nil {
		diags.AddError("Failed to unmarshal treemap ES|QL payload", err.Error())
		return api, diags
	}
	if api.GroupBy == nil || len(*api.GroupBy) == 0 {
		diags.AddError("Invalid group_by_json", "treemap_config.group_by_json must contain at least one item")
		return api, diags
	}
	if len(api.Metrics) == 0 {
		diags.AddError("Invalid metrics_json", "treemap_config.metrics_json must contain at least one item")
		return api, diags
	}

	api.Legend = m.Legend.toTreemapLegend()

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
		api.Styling.Values = m.ValueDisplay.toValueDisplay()
	} else {
		defaultMode := kbapi.ValueDisplayModePercentage
		api.Styling.Values = kbapi.ValueDisplay{Mode: &defaultMode}
	}

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.lensChartPresentationTFModel)
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

func (m *treemapConfigModel) usesESQL() bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Expression.IsNull() && m.Query.Language.IsNull()
}

func (m *treemapConfigModel) toAPINoESQL(dashboard *dashboardModel) (kbapi.TreemapNoESQL, diag.Diagnostics) {
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
	api.Query = m.Query.toAPI()

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if m.Legend == nil {
		diags.AddError("Missing legend", "treemap_config.legend must be provided")
		return api, diags
	}
	api.Legend = m.Legend.toTreemapLegend()

	if m.ValueDisplay != nil {
		api.Styling.Values = m.ValueDisplay.toValueDisplay()
	}

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.lensChartPresentationTFModel)
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
