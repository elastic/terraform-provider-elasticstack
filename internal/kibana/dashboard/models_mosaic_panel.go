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

func newMosaicPanelConfigConverter() mosaicPanelConfigConverter {
	return mosaicPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.MosaicNoESQLTypeMosaic),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.MosaicConfig != nil },
		},
	}
}

type mosaicPanelConfigConverter struct {
	lensVisualizationBase
}

func (c mosaicPanelConfigConverter) populateFromAttributes(_ context.Context, pm *panelModel, attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes) diag.Diagnostics {
	mosaicChart, err := attrs.AsMosaicChart()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	if pm.MosaicConfig == nil {
		pm.MosaicConfig = &mosaicConfigModel{}
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

	if datasetType == legacyMetricDatasetTypeESQL {
		mosaicESQL, err := mosaicChart.AsMosaicESQL()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		return pm.MosaicConfig.fromAPIESQL(mosaicESQL)
	}

	mosaicNoESQL, err := mosaicChart.AsMosaicNoESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.MosaicConfig.fromAPINoESQL(mosaicNoESQL)
}

func (c mosaicPanelConfigConverter) buildAttributes(pm panelModel) (kbapi.KbnDashboardPanelLens_Config_0_Attributes, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *pm.MosaicConfig

	mosaicChart, mosaicDiags := configModel.toAPI()
	diags.Append(mosaicDiags...)
	if diags.HasError() {
		return kbapi.KbnDashboardPanelLens_Config_0_Attributes{}, diags
	}

	var attrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	if err := attrs.FromMosaicChart(mosaicChart); err != nil {
		diags.AddError("Failed to create mosaic attributes", err.Error())
		return kbapi.KbnDashboardPanelLens_Config_0_Attributes{}, diags
	}

	return attrs, diags
}

type mosaicConfigModel struct {
	Title               types.String                                        `tfsdk:"title"`
	Description         types.String                                        `tfsdk:"description"`
	Dataset             jsontypes.Normalized                                `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool                                          `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                       `tfsdk:"sampling"`
	Query               *filterSimpleModel                                  `tfsdk:"query"`
	Filters             []searchFilterModel                                 `tfsdk:"filters"`
	GroupBy             customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"group_by_json"`
	GroupBreakdownBy    customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"group_breakdown_by_json"`
	Metrics             customtypes.JSONWithDefaultsValue[[]map[string]any] `tfsdk:"metrics_json"`
	Legend              *partitionLegendModel                               `tfsdk:"legend"`
	ValueDisplay        *partitionValueDisplay                              `tfsdk:"value_display"`
}

func (m *mosaicConfigModel) fromAPINoESQL(api kbapi.MosaicNoESQL) diag.Diagnostics {
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
		groupByBytes, err := json.Marshal(api.GroupBy)
		if err != nil {
			diags.AddError("Failed to marshal group_by", err.Error())
			return diags
		}
		m.GroupBy = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(groupByBytes), populatePartitionGroupByDefaults)
	} else {
		m.GroupBy = customtypes.NewJSONWithDefaultsNull(populatePartitionGroupByDefaults)
	}

	if api.GroupBreakdownBy != nil {
		groupBreakdownByBytes, err := json.Marshal(api.GroupBreakdownBy)
		if err != nil {
			diags.AddError("Failed to marshal group_breakdown_by", err.Error())
			return diags
		}
		m.GroupBreakdownBy = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(groupBreakdownByBytes), populatePartitionGroupByDefaults)
	} else {
		m.GroupBreakdownBy = customtypes.NewJSONWithDefaultsNull(populatePartitionGroupByDefaults)
	}

	metricsBytes, err := json.Marshal(api.Metrics)
	if err != nil {
		diags.AddError("Failed to marshal metrics", err.Error())
		return diags
	}
	m.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(metricsBytes), populatePartitionMetricsDefaults)

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

	m.Legend = &partitionLegendModel{}
	m.Legend.fromMosaicLegend(api.Legend)

	if api.ValueDisplay != nil {
		m.ValueDisplay = &partitionValueDisplay{}
		m.ValueDisplay.fromMosaicNoESQL(api.ValueDisplay)
	} else {
		m.ValueDisplay = nil
	}

	return diags
}

func (m *mosaicConfigModel) fromAPIESQL(api kbapi.MosaicESQL) diag.Diagnostics {
	var diags diag.Diagnostics

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
		groupByBytes, err := json.Marshal(api.GroupBy)
		if err != nil {
			diags.AddError("Failed to marshal group_by", err.Error())
			return diags
		}
		m.GroupBy = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(groupByBytes), populatePartitionGroupByDefaults)
	} else {
		m.GroupBy = customtypes.NewJSONWithDefaultsNull(populatePartitionGroupByDefaults)
	}

	if api.GroupBreakdownBy != nil {
		groupBreakdownByBytes, err := json.Marshal(api.GroupBreakdownBy)
		if err != nil {
			diags.AddError("Failed to marshal group_breakdown_by", err.Error())
			return diags
		}
		m.GroupBreakdownBy = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(groupBreakdownByBytes), populatePartitionGroupByDefaults)
	} else {
		m.GroupBreakdownBy = customtypes.NewJSONWithDefaultsNull(populatePartitionGroupByDefaults)
	}

	metricsBytes, err := json.Marshal(api.Metrics)
	if err != nil {
		diags.AddError("Failed to marshal metrics", err.Error())
		return diags
	}
	m.Metrics = customtypes.NewJSONWithDefaultsValue[[]map[string]any](string(metricsBytes), populatePartitionMetricsDefaults)

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

	m.Legend = &partitionLegendModel{}
	m.Legend.fromMosaicLegend(api.Legend)

	if api.ValueDisplay != nil {
		m.ValueDisplay = &partitionValueDisplay{}
		m.ValueDisplay.fromMosaicESQL(api.ValueDisplay)
	} else {
		m.ValueDisplay = nil
	}

	return diags
}

func (m *mosaicConfigModel) toAPI() (kbapi.MosaicChart, diag.Diagnostics) {
	var diags diag.Diagnostics
	var mosaicChart kbapi.MosaicChart

	if m == nil {
		return mosaicChart, diags
	}

	if m.usesESQL() {
		return m.toAPIESQLChartSchema()
	}

	noESQL, noESQLDiags := m.toAPINoESQL()
	diags.Append(noESQLDiags...)
	if diags.HasError() {
		return mosaicChart, diags
	}
	if err := mosaicChart.FromMosaicNoESQL(noESQL); err != nil {
		diags.AddError("Failed to create mosaic schema", err.Error())
	}

	return mosaicChart, diags
}

func (m *mosaicConfigModel) usesESQL() bool {
	if m == nil {
		return false
	}
	if m.Query == nil {
		return true
	}
	return m.Query.Query.IsNull() && m.Query.Language.IsNull()
}

func (m *mosaicConfigModel) toAPIESQLChartSchema() (kbapi.MosaicChart, diag.Diagnostics) {
	var diags diag.Diagnostics
	var mosaicChart kbapi.MosaicChart

	if m.Dataset.IsNull() {
		diags.AddError("Missing dataset_json", "mosaic_config.dataset_json must be provided")
		return mosaicChart, diags
	}
	if m.GroupBy.IsNull() {
		diags.AddError("Missing group_by_json", "mosaic_config.group_by_json must be provided")
		return mosaicChart, diags
	}
	if m.GroupBreakdownBy.IsNull() {
		diags.AddError("Missing group_breakdown_by_json", "mosaic_config.group_breakdown_by_json must be provided")
		return mosaicChart, diags
	}
	if m.Metrics.IsNull() {
		diags.AddError("Missing metrics_json", "mosaic_config.metrics_json must be provided")
		return mosaicChart, diags
	}
	if m.Legend == nil {
		diags.AddError("Missing legend", "mosaic_config.legend must be provided")
		return mosaicChart, diags
	}
	if diags.HasError() {
		return mosaicChart, diags
	}

	api := kbapi.MosaicESQL{
		Type:   kbapi.MosaicESQLTypeMosaic,
		Legend: m.Legend.toMosaicLegend(),
	}

	if err := json.Unmarshal([]byte(m.Dataset.ValueString()), &api.Dataset); err != nil {
		diags.AddError("Failed to unmarshal dataset", err.Error())
		return mosaicChart, diags
	}
	if err := json.Unmarshal([]byte(m.GroupBy.ValueString()), &api.GroupBy); err != nil {
		diags.AddError("Failed to unmarshal group_by", err.Error())
		return mosaicChart, diags
	}
	if err := json.Unmarshal([]byte(m.GroupBreakdownBy.ValueString()), &api.GroupBreakdownBy); err != nil {
		diags.AddError("Failed to unmarshal group_breakdown_by", err.Error())
		return mosaicChart, diags
	}
	if err := json.Unmarshal([]byte(m.Metrics.ValueString()), &api.Metrics); err != nil {
		diags.AddError("Failed to unmarshal metrics", err.Error())
		return mosaicChart, diags
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

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilter, len(m.Filters))
		for i, filterModel := range m.Filters {
			filter, filterDiags := filterModel.toAPI()
			diags.Append(filterDiags...)
			if diags.HasError() {
				return mosaicChart, diags
			}
			filters[i] = filter
		}
		api.Filters = &filters
	}

	if m.ValueDisplay != nil {
		vd := m.ValueDisplay.toMosaicESQL()
		api.ValueDisplay = &struct {
			Mode            kbapi.MosaicESQLValueDisplayMode `json:"mode"`
			PercentDecimals *float32                         `json:"percent_decimals,omitempty"`
		}{
			Mode:            vd.Mode,
			PercentDecimals: vd.PercentDecimals,
		}
	}

	if err := mosaicChart.FromMosaicESQL(api); err != nil {
		diags.AddError("Failed to create mosaic chart schema", err.Error())
		return mosaicChart, diags
	}

	return mosaicChart, diags
}

func (m *mosaicConfigModel) toAPINoESQL() (kbapi.MosaicNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.MosaicNoESQL{Type: kbapi.MosaicNoESQLTypeMosaic}

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
		diags.AddError("Missing dataset_json", "mosaic_config.dataset_json must be provided")
		return api, diags
	}
	if err := json.Unmarshal([]byte(m.Dataset.ValueString()), &api.Dataset); err != nil {
		diags.AddError("Failed to unmarshal dataset", err.Error())
		return api, diags
	}

	if m.GroupBy.IsNull() {
		diags.AddError("Missing group_by_json", "mosaic_config.group_by_json must be provided")
		return api, diags
	}
	var groupBy []kbapi.MosaicNoESQL_GroupBy_Item
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
	var groupBreakdownBy []kbapi.MosaicNoESQL_GroupBreakdownBy_Item
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
	var metrics []kbapi.MosaicNoESQL_Metrics_Item
	if err := json.Unmarshal([]byte(m.Metrics.ValueString()), &metrics); err != nil {
		diags.AddError("Failed to unmarshal metrics", err.Error())
		return api, diags
	}
	if len(metrics) == 0 {
		diags.AddError("Invalid metrics_json", "mosaic_config.metrics_json must contain at least one item")
		return api, diags
	}
	api.Metrics = metrics

	if m.Query == nil {
		diags.AddError("Missing query", "mosaic_config.query is required for non-ES|QL mosaic charts")
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

	if m.Legend == nil {
		diags.AddError("Missing legend", "mosaic_config.legend must be provided")
		return api, diags
	}
	api.Legend = m.Legend.toMosaicLegend()

	if m.ValueDisplay != nil {
		valueDisplay := m.ValueDisplay.toMosaicNoESQL()
		api.ValueDisplay = &valueDisplay
	}

	return api, diags
}
