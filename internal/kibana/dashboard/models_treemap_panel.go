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

func newTreemapPanelConfigConverter() treemapPanelConfigConverter {
	return treemapPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.TreemapNoESQLTypeTreemap),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.TreemapConfig != nil },
		},
	}
}

type treemapPanelConfigConverter struct {
	lensVisualizationBase
}

func (c treemapPanelConfigConverter) populateFromAttributes(_ context.Context, pm *panelModel, attrs kbapi.LensApiState) diag.Diagnostics {
	treemapChart, err := attrs.AsTreemapChart()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	if pm.TreemapConfig == nil {
		pm.TreemapConfig = &treemapConfigModel{}
	}

	datasetType := ""
	if attrsJSON, err := attrs.MarshalJSON(); err == nil {
		var attrsMap map[string]any
		if err := json.Unmarshal(attrsJSON, &attrsMap); err == nil {
			if dataset, ok := attrsMap["dataset"].(map[string]any); ok {
				if t, ok := dataset["type"].(string); ok {
					datasetType = t
				}
			}
		}
	}

	if datasetType == "esql" {
		treemapESQL, err := treemapChart.AsTreemapESQL()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		return pm.TreemapConfig.fromAPIESQL(treemapESQL)
	}

	treemapNoESQL, err := treemapChart.AsTreemapNoESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.TreemapConfig.fromAPINoESQL(treemapNoESQL)
}

func (c treemapPanelConfigConverter) buildAttributes(pm panelModel) (kbapi.LensApiState, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *pm.TreemapConfig

	treemapChart, treemapDiags := configModel.toAPI()
	diags.Append(treemapDiags...)
	if diags.HasError() {
		return kbapi.LensApiState{}, diags
	}

	var attrs kbapi.LensApiState
	if err := attrs.FromTreemapChart(treemapChart); err != nil {
		diags.AddError("Failed to create treemap attributes", err.Error())
		return kbapi.LensApiState{}, diags
	}

	return attrs, diags
}

type treemapConfigModel struct {
	Title               types.String                                        `tfsdk:"title"`
	Description         types.String                                        `tfsdk:"description"`
	Dataset             jsontypes.Normalized                                `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool                                          `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                       `tfsdk:"sampling"`
	Query               *filterSimpleModel                                  `tfsdk:"query"`
	Filters             []chartFilterJSONModel                              `tfsdk:"filters"`
	GroupBy             customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"group_by_json"`
	Metrics             customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"metrics_json"`
	Legend              *partitionLegendModel                               `tfsdk:"legend"`
	ValueDisplay        *partitionValueDisplay                              `tfsdk:"value_display"`
}

func (m *treemapConfigModel) fromAPINoESQL(api kbapi.TreemapNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = mapOptionalBoolWithSnapshotDefault(m.IgnoreGlobalFilters, api.IgnoreGlobalFilters, false)
	m.Sampling = mapOptionalFloatWithSnapshotDefault(m.Sampling, api.Sampling, 1)

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.Dataset = jsontypes.NewNormalizedValue(string(datasetBytes))

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
	m.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(metricsBytes), populatePartitionMetricsDefaults)

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	if len(api.Filters) > 0 {
		m.Filters = make([]chartFilterJSONModel, 0, len(api.Filters))
		for _, filter := range api.Filters {
			fm := chartFilterJSONModel{}
			filterDiags := fm.populateFromAPIItem(filter)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				m.Filters = append(m.Filters, fm)
			}
		}
	} else {
		m.Filters = nil
	}

	m.Legend = &partitionLegendModel{}
	m.Legend.fromTreemapLegend(api.Legend)

	if api.Values.Mode != nil || api.Values.PercentDecimals != nil {
		m.ValueDisplay = &partitionValueDisplay{}
		m.ValueDisplay.fromValueDisplay(api.Values)
	} else {
		m.ValueDisplay = nil
	}

	return diags
}

func (m *treemapConfigModel) fromAPIESQL(api kbapi.TreemapESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	// ES|QL charts don't have a query block. Clear it to avoid carrying over
	// query state from a previous non-ES|QL config.
	m.Query = nil

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	m.IgnoreGlobalFilters = mapOptionalBoolWithSnapshotDefault(m.IgnoreGlobalFilters, api.IgnoreGlobalFilters, false)
	m.Sampling = mapOptionalFloatWithSnapshotDefault(m.Sampling, api.Sampling, 1)

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.Dataset = jsontypes.NewNormalizedValue(string(datasetBytes))

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
	m.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(metricsBytes), populatePartitionMetricsDefaults)

	if len(api.Filters) > 0 {
		m.Filters = make([]chartFilterJSONModel, 0, len(api.Filters))
		for _, filter := range api.Filters {
			fm := chartFilterJSONModel{}
			filterDiags := fm.populateFromAPIItem(filter)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				m.Filters = append(m.Filters, fm)
			}
		}
	} else {
		m.Filters = nil
	}

	m.Legend = &partitionLegendModel{}
	m.Legend.fromTreemapLegend(api.Legend)

	if api.Values.Mode != nil || api.Values.PercentDecimals != nil {
		m.ValueDisplay = &partitionValueDisplay{}
		m.ValueDisplay.fromValueDisplay(api.Values)
	} else {
		m.ValueDisplay = nil
	}

	return diags
}

func (m *treemapConfigModel) toAPI() (kbapi.TreemapChart, diag.Diagnostics) {
	var diags diag.Diagnostics
	var treemapChart kbapi.TreemapChart

	if m == nil {
		return treemapChart, diags
	}

	if m.usesESQL() {
		return m.toAPIESQLChartSchema()
	}

	noESQL, noESQLDiags := m.toAPINoESQL()
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return treemapChart, diags
	}
	if err := treemapChart.FromTreemapNoESQL(noESQL); err != nil {
		diags.AddError("Failed to create treemap schema", err.Error())
	}

	return treemapChart, diags
}

func (m *treemapConfigModel) toAPIESQLChartSchema() (kbapi.TreemapChart, diag.Diagnostics) {
	var diags diag.Diagnostics
	var treemapChart kbapi.TreemapChart

	attrs := map[string]any{
		"type": string(kbapi.TreemapESQLTypeTreemap),
	}

	if typeutils.IsKnown(m.Title) {
		attrs["title"] = m.Title.ValueString()
	}
	if typeutils.IsKnown(m.Description) {
		attrs["description"] = m.Description.ValueString()
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		attrs["ignore_global_filters"] = m.IgnoreGlobalFilters.ValueBool()
	}
	if typeutils.IsKnown(m.Sampling) {
		attrs["sampling"] = m.Sampling.ValueFloat64()
	}

	if m.Dataset.IsNull() {
		diags.AddError("Missing dataset_json", "treemap_config.dataset_json must be provided")
		return treemapChart, diags
	}
	var dataset any
	if err := json.Unmarshal([]byte(m.Dataset.ValueString()), &dataset); err != nil {
		diags.AddError("Failed to unmarshal dataset", err.Error())
		return treemapChart, diags
	}
	attrs["dataset"] = dataset

	if m.GroupBy.IsNull() {
		diags.AddError("Missing group_by_json", "treemap_config.group_by_json must be provided")
		return treemapChart, diags
	}
	var groupBy any
	if err := json.Unmarshal([]byte(m.GroupBy.ValueString()), &groupBy); err != nil {
		diags.AddError("Failed to unmarshal group_by", err.Error())
		return treemapChart, diags
	}
	attrs["group_by"] = groupBy

	if m.Metrics.IsNull() {
		diags.AddError("Missing metrics_json", "treemap_config.metrics_json must be provided")
		return treemapChart, diags
	}
	var metrics any
	if err := json.Unmarshal([]byte(m.Metrics.ValueString()), &metrics); err != nil {
		diags.AddError("Failed to unmarshal metrics", err.Error())
		return treemapChart, diags
	}
	attrs["metrics"] = metrics

	attrs["filters"] = []any{}
	if len(m.Filters) > 0 {
		filters := make([]any, 0, len(m.Filters))
		for _, filterModel := range m.Filters {
			var filterAny map[string]any
			filterDiags := decodeChartFilterJSON(filterModel.FilterJSON, &filterAny)
			diags.Append(filterDiags...)
			if diags.HasError() {
				return treemapChart, diags
			}
			filters = append(filters, filterAny)
		}
		attrs["filters"] = filters
	}

	if m.Legend == nil {
		diags.AddError("Missing legend", "treemap_config.legend must be provided")
		return treemapChart, diags
	}
	legendBytes, err := json.Marshal(m.Legend.toTreemapLegend())
	if err != nil {
		diags.AddError("Failed to marshal legend", err.Error())
		return treemapChart, diags
	}
	var legend any
	if err := json.Unmarshal(legendBytes, &legend); err != nil {
		diags.AddError("Failed to unmarshal legend", err.Error())
		return treemapChart, diags
	}
	attrs["legend"] = legend

	if m.ValueDisplay != nil {
		valueDisplayBytes, err := json.Marshal(m.ValueDisplay.toValueDisplay())
		if err != nil {
			diags.AddError("Failed to marshal value_display", err.Error())
			return treemapChart, diags
		}
		var valueDisplay any
		if err := json.Unmarshal(valueDisplayBytes, &valueDisplay); err != nil {
			diags.AddError("Failed to unmarshal value_display", err.Error())
			return treemapChart, diags
		}
		attrs["values"] = valueDisplay
	}

	attrsJSON, err := json.Marshal(attrs)
	if err != nil {
		diags.AddError("Failed to marshal treemap attributes", err.Error())
		return treemapChart, diags
	}
	if err := json.Unmarshal(attrsJSON, &treemapChart); err != nil {
		diags.AddError("Failed to create treemap chart schema", err.Error())
		return treemapChart, diags
	}

	return treemapChart, diags
}

func (m *treemapConfigModel) usesESQL() bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Query.IsNull() && m.Query.Language.IsNull()
}

func (m *treemapConfigModel) toAPINoESQL() (kbapi.TreemapNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.TreemapNoESQL{Type: kbapi.TreemapNoESQLTypeTreemap}

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

	if m.Dataset.IsNull() {
		diags.AddError("Missing dataset_json", "treemap_config.dataset_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.Dataset.ValueString()), &api.Dataset); err != nil {
		diags.AddError("Failed to unmarshal dataset", err.Error())
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

	api.Filters = []kbapi.LensPanelFilters_Item{}
	if len(m.Filters) > 0 {
		filters := make([]kbapi.LensPanelFilters_Item, 0, len(m.Filters))
		for _, filterModel := range m.Filters {
			var item kbapi.LensPanelFilters_Item
			filterDiags := decodeChartFilterJSON(filterModel.FilterJSON, &item)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				filters = append(filters, item)
			}
		}
		if len(filters) > 0 {
			api.Filters = filters
		}
	}

	if m.Legend == nil {
		diags.AddError("Missing legend", "treemap_config.legend must be provided")
		return api, diags
	}
	api.Legend = m.Legend.toTreemapLegend()

	if m.ValueDisplay != nil {
		api.Values = m.ValueDisplay.toValueDisplay()
	}

	return api, diags
}
