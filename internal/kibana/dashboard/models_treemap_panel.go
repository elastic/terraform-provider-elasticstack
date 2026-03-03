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
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.TreemapNoESQLTypeTreemap),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.TreemapConfig != nil },
		},
	}
}

type treemapPanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c treemapPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.TreemapConfig != nil
}

func (c treemapPanelConfigConverter) populateFromAPIPanel(_ context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	cfgMap, err := config.AsDashboardPanelItemConfig8()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	attrs, ok := cfgMap["attributes"]
	if !ok {
		return nil
	}

	attrsMap, ok := attrs.(map[string]any)
	if !ok {
		return nil
	}

	attrsJSON, err := json.Marshal(attrsMap)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var treemapChart kbapi.TreemapChart
	if err := json.Unmarshal(attrsJSON, &treemapChart); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	_, hasQuery := attrsMap["query"]

	if pm.TreemapConfig == nil {
		pm.TreemapConfig = &treemapConfigModel{}
	}
	if hasQuery {
		treemapNoESQL, err := treemapChart.AsTreemapNoESQL()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		return pm.TreemapConfig.fromAPINoESQL(treemapNoESQL)
	}

	treemapESQL, err := treemapChart.AsTreemapESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.TreemapConfig.fromAPIESQL(treemapESQL)
}

func (c treemapPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	configModel := *pm.TreemapConfig

	treemapChart, treemapDiags := configModel.toAPI()
	diags.Append(treemapDiags...)
	if diags.HasError() {
		return diags
	}

	var attrs0 kbapi.DashboardPanelItemConfig70Attributes0
	if err := attrs0.FromTreemapChart(treemapChart); err != nil {
		diags.AddError("Failed to create treemap attributes", err.Error())
		return diags
	}

	var configAttrs kbapi.DashboardPanelItem_Config_7_0_Attributes
	if err := configAttrs.FromDashboardPanelItemConfig70Attributes0(attrs0); err != nil {
		diags.AddError("Failed to create config attributes", err.Error())
		return diags
	}

	config10 := kbapi.DashboardPanelItemConfig70{
		Attributes: configAttrs,
	}

	var config1 kbapi.DashboardPanelItemConfig7
	if err := config1.FromDashboardPanelItemConfig70(config10); err != nil {
		diags.AddError("Failed to create config1", err.Error())
		return diags
	}

	if err := apiConfig.FromDashboardPanelItemConfig7(config1); err != nil {
		diags.AddError("Failed to marshal treemap config", err.Error())
	}

	return diags
}

type treemapConfigModel struct {
	Title               types.String                                        `tfsdk:"title"`
	Description         types.String                                        `tfsdk:"description"`
	Dataset             jsontypes.Normalized                                `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool                                          `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                       `tfsdk:"sampling"`
	Query               *filterSimpleModel                                  `tfsdk:"query"`
	Filters             []searchFilterModel                                 `tfsdk:"filters"`
	GroupBy             customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"group_by_json"`
	Metrics             customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"metrics_json"`
	LabelPosition       types.String                                        `tfsdk:"label_position"`
	Legend              *treemapLegendModel                                 `tfsdk:"legend"`
	ValueDisplay        *treemapValueDisplay                                `tfsdk:"value_display"`
}

type treemapLegendModel struct {
	Nested            types.Bool    `tfsdk:"nested"`
	Size              types.String  `tfsdk:"size"`
	TruncateAfterLine types.Float64 `tfsdk:"truncate_after_lines"`
	Visible           types.String  `tfsdk:"visible"`
}

type treemapValueDisplay struct {
	Mode            types.String  `tfsdk:"mode"`
	PercentDecimals types.Float64 `tfsdk:"percent_decimals"`
}

func (m *treemapConfigModel) fromAPINoESQL(api kbapi.TreemapNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	// Kibana may omit these optional attributes in GET responses even when they were
	// provided on write. Preserve any already-known value (typically from the plan)
	// to avoid "inconsistent result after apply" drift.
	if api.IgnoreGlobalFilters != nil {
		m.IgnoreGlobalFilters = types.BoolValue(*api.IgnoreGlobalFilters)
	} else if !typeutils.IsKnown(m.IgnoreGlobalFilters) {
		m.IgnoreGlobalFilters = types.BoolNull()
	}
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else if !typeutils.IsKnown(m.Sampling) {
		m.Sampling = types.Float64Null()
	}

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.Dataset = jsontypes.NewNormalizedValue(string(datasetBytes))

	if api.GroupBy != nil {
		groupByBytes, err := json.Marshal(api.GroupBy)
		if err != nil {
			diags.AddError("Failed to marshal group_by", err.Error())
			return diags
		}
		m.GroupBy = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(groupByBytes), populateTreemapGroupByDefaults)
	} else {
		m.GroupBy = customtypes.NewJSONWithDefaultsNull(populateTreemapGroupByDefaults)
	}

	metricsBytes, err := json.Marshal(api.Metrics)
	if err != nil {
		diags.AddError("Failed to marshal metrics", err.Error())
		return diags
	}
	m.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(metricsBytes), populateTreemapMetricsDefaults)

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, 0, len(*api.Filters))
		for _, filter := range *api.Filters {
			filterModel := searchFilterModel{}
			filterDiags := filterModel.fromAPI(filter)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				m.Filters = append(m.Filters, filterModel)
			}
		}
	} else {
		m.Filters = nil
	}

	if api.LabelPosition != nil {
		m.LabelPosition = types.StringValue(string(*api.LabelPosition))
	} else if !typeutils.IsKnown(m.LabelPosition) {
		m.LabelPosition = types.StringNull()
	}

	m.Legend = &treemapLegendModel{}
	m.Legend.fromAPI(api.Legend)

	if api.ValueDisplay != nil {
		m.ValueDisplay = &treemapValueDisplay{}
		m.ValueDisplay.fromAPINoESQL(api.ValueDisplay)
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
	// Kibana may omit these optional attributes in GET responses even when they were
	// provided on write. Preserve any already-known value (typically from the plan)
	// to avoid "inconsistent result after apply" drift.
	if api.IgnoreGlobalFilters != nil {
		m.IgnoreGlobalFilters = types.BoolValue(*api.IgnoreGlobalFilters)
	} else if !typeutils.IsKnown(m.IgnoreGlobalFilters) {
		m.IgnoreGlobalFilters = types.BoolNull()
	}
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else if !typeutils.IsKnown(m.Sampling) {
		m.Sampling = types.Float64Null()
	}

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.Dataset = jsontypes.NewNormalizedValue(string(datasetBytes))

	if api.GroupBy != nil {
		groupByBytes, err := json.Marshal(api.GroupBy)
		if err != nil {
			diags.AddError("Failed to marshal group_by", err.Error())
			return diags
		}
		m.GroupBy = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(groupByBytes), populateTreemapGroupByDefaults)
	} else {
		m.GroupBy = customtypes.NewJSONWithDefaultsNull(populateTreemapGroupByDefaults)
	}

	metricsBytes, err := json.Marshal(api.Metrics)
	if err != nil {
		diags.AddError("Failed to marshal metrics", err.Error())
		return diags
	}
	m.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(metricsBytes), populateTreemapMetricsDefaults)

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, 0, len(*api.Filters))
		for _, filter := range *api.Filters {
			filterModel := searchFilterModel{}
			filterDiags := filterModel.fromAPI(filter)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				m.Filters = append(m.Filters, filterModel)
			}
		}
	} else {
		m.Filters = nil
	}

	if api.LabelPosition != nil {
		m.LabelPosition = types.StringValue(string(*api.LabelPosition))
	} else if !typeutils.IsKnown(m.LabelPosition) {
		m.LabelPosition = types.StringNull()
	}

	m.Legend = &treemapLegendModel{}
	m.Legend.fromAPI(api.Legend)

	if api.ValueDisplay != nil {
		m.ValueDisplay = &treemapValueDisplay{}
		m.ValueDisplay.fromAPIESQL(api.ValueDisplay)
	} else {
		m.ValueDisplay = nil
	}

	return diags
}

func (m *treemapLegendModel) fromAPI(api kbapi.TreemapLegend) {
	m.Nested = types.BoolPointerValue(api.Nested)
	m.Size = types.StringValue(string(api.Size))
	if api.TruncateAfterLines != nil {
		m.TruncateAfterLine = types.Float64Value(float64(*api.TruncateAfterLines))
	} else {
		m.TruncateAfterLine = types.Float64Null()
	}
	if api.Visible != nil {
		m.Visible = types.StringValue(string(*api.Visible))
	} else {
		m.Visible = types.StringNull()
	}
}

func (m *treemapValueDisplay) fromAPINoESQL(api *struct {
	Mode            kbapi.TreemapNoESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                            `json:"percent_decimals,omitempty"`
}) {
	m.Mode = types.StringValue(string(api.Mode))
	if api.PercentDecimals != nil {
		m.PercentDecimals = types.Float64Value(float64(*api.PercentDecimals))
	} else {
		m.PercentDecimals = types.Float64Null()
	}
}

func (m *treemapValueDisplay) fromAPIESQL(api *struct {
	Mode            kbapi.TreemapESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                          `json:"percent_decimals,omitempty"`
}) {
	m.Mode = types.StringValue(string(api.Mode))
	if api.PercentDecimals != nil {
		m.PercentDecimals = types.Float64Value(float64(*api.PercentDecimals))
	} else {
		m.PercentDecimals = types.Float64Null()
	}
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

	if len(m.Filters) > 0 {
		filters := make([]any, 0, len(m.Filters))
		for _, filterModel := range m.Filters {
			filter, filterDiags := filterModel.toAPI()
			diags.Append(filterDiags...)
			if diags.HasError() {
				return treemapChart, diags
			}

			filterBytes, err := json.Marshal(filter)
			if err != nil {
				diags.AddError("Failed to marshal filter", err.Error())
				return treemapChart, diags
			}
			var filterAny any
			if err := json.Unmarshal(filterBytes, &filterAny); err != nil {
				diags.AddError("Failed to unmarshal filter", err.Error())
				return treemapChart, diags
			}
			filters = append(filters, filterAny)
		}
		attrs["filters"] = filters
	}

	if typeutils.IsKnown(m.LabelPosition) {
		attrs["label_position"] = m.LabelPosition.ValueString()
	}

	if m.Legend == nil {
		diags.AddError("Missing legend", "treemap_config.legend must be provided")
		return treemapChart, diags
	}
	legendBytes, err := json.Marshal(m.Legend.toAPI())
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
		valueDisplayBytes, err := json.Marshal(m.ValueDisplay.toAPIESQL())
		if err != nil {
			diags.AddError("Failed to marshal value_display", err.Error())
			return treemapChart, diags
		}
		var valueDisplay any
		if err := json.Unmarshal(valueDisplayBytes, &valueDisplay); err != nil {
			diags.AddError("Failed to unmarshal value_display", err.Error())
			return treemapChart, diags
		}
		attrs["value_display"] = valueDisplay
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

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilter, len(m.Filters))
		for i, filterModel := range m.Filters {
			filter, filterDiags := filterModel.toAPI()
			diags.Append(filterDiags...)
			filters[i] = filter
		}
		api.Filters = &filters
	}

	if typeutils.IsKnown(m.LabelPosition) {
		lp := kbapi.TreemapNoESQLLabelPosition(m.LabelPosition.ValueString())
		api.LabelPosition = &lp
	}

	if m.Legend == nil {
		diags.AddError("Missing legend", "treemap_config.legend must be provided")
		return api, diags
	}
	api.Legend = m.Legend.toAPI()

	if m.ValueDisplay != nil {
		valueDisplay := m.ValueDisplay.toAPINoESQL()
		api.ValueDisplay = &valueDisplay
	}

	return api, diags
}

func (m *treemapLegendModel) toAPI() kbapi.TreemapLegend {
	legend := kbapi.TreemapLegend{Size: kbapi.LegendSize(m.Size.ValueString())}
	if typeutils.IsKnown(m.Nested) {
		legend.Nested = new(m.Nested.ValueBool())
	}
	if typeutils.IsKnown(m.TruncateAfterLine) {
		legend.TruncateAfterLines = new(float32(m.TruncateAfterLine.ValueFloat64()))
	}
	if typeutils.IsKnown(m.Visible) {
		v := kbapi.TreemapLegendVisible(m.Visible.ValueString())
		legend.Visible = &v
	}
	return legend
}

func (m *treemapValueDisplay) toAPINoESQL() struct {
	Mode            kbapi.TreemapNoESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                            `json:"percent_decimals,omitempty"`
} {
	vd := struct {
		Mode            kbapi.TreemapNoESQLValueDisplayMode `json:"mode"`
		PercentDecimals *float32                            `json:"percent_decimals,omitempty"`
	}{
		Mode: kbapi.TreemapNoESQLValueDisplayMode(m.Mode.ValueString()),
	}
	if typeutils.IsKnown(m.PercentDecimals) {
		vd.PercentDecimals = new(float32(m.PercentDecimals.ValueFloat64()))
	}
	return vd
}

func (m *treemapValueDisplay) toAPIESQL() struct {
	Mode            kbapi.TreemapESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                          `json:"percent_decimals,omitempty"`
} {
	vd := struct {
		Mode            kbapi.TreemapESQLValueDisplayMode `json:"mode"`
		PercentDecimals *float32                          `json:"percent_decimals,omitempty"`
	}{
		Mode: kbapi.TreemapESQLValueDisplayMode(m.Mode.ValueString()),
	}
	if typeutils.IsKnown(m.PercentDecimals) {
		vd.PercentDecimals = new(float32(m.PercentDecimals.ValueFloat64()))
	}
	return vd
}
