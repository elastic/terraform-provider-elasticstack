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

const (
	legacyMetricDatasetTypeDataView = "dataView"
	legacyMetricDatasetTypeIndex    = "index"
	legacyMetricDatasetTypeESQL     = "esql"
	legacyMetricDatasetTypeTable    = "table"
)

func newLegacyMetricPanelConfigConverter() legacyMetricPanelConfigConverter {
	return legacyMetricPanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.LegacyMetricNoESQLTypeLegacyMetric),
		},
	}
}

type legacyMetricPanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c legacyMetricPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.LegacyMetricConfig != nil
}

func (c legacyMetricPanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	cfgMap, err := config.AsDashboardPanelItemConfig2()
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

	var legacyMetricChart kbapi.LegacyMetricChartSchema
	if err := json.Unmarshal(attrsJSON, &legacyMetricChart); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.LegacyMetricConfig = &legacyMetricConfigModel{}
	return pm.LegacyMetricConfig.fromAPI(ctx, legacyMetricChart)
}

func (c legacyMetricPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	configModel := *pm.LegacyMetricConfig

	legacyMetricChart, legacyDiags := configModel.toAPI()
	diags.Append(legacyDiags...)
	if diags.HasError() {
		return diags
	}

	var attrs0 kbapi.DashboardPanelItemConfig10Attributes0
	if err := attrs0.FromLegacyMetricChartSchema(legacyMetricChart); err != nil {
		diags.AddError("Failed to create legacy metric attributes", err.Error())
		return diags
	}

	var configAttrs kbapi.DashboardPanelItem_Config_1_0_Attributes
	if err := configAttrs.FromDashboardPanelItemConfig10Attributes0(attrs0); err != nil {
		diags.AddError("Failed to create config attributes", err.Error())
		return diags
	}

	config10 := kbapi.DashboardPanelItemConfig10{
		Attributes: configAttrs,
	}

	var config1 kbapi.DashboardPanelItemConfig1
	if err := config1.FromDashboardPanelItemConfig10(config10); err != nil {
		diags.AddError("Failed to create config1", err.Error())
		return diags
	}

	if err := apiConfig.FromDashboardPanelItemConfig1(config1); err != nil {
		diags.AddError("Failed to marshal legacy metric config", err.Error())
	}

	return diags
}

type legacyMetricConfigModel struct {
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DatasetJSON         jsontypes.Normalized                              `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []searchFilterModel                               `tfsdk:"filters"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
}

func (m *legacyMetricConfigModel) fromAPI(ctx context.Context, apiChart kbapi.LegacyMetricChartSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	legacyNoESQL, err := apiChart.AsLegacyMetricNoESQL()
	if err == nil {
		return m.fromAPINoESQL(ctx, legacyNoESQL)
	}

	legacyESQL, err := apiChart.AsLegacyMetricESQL()
	if err == nil {
		return m.fromAPIESQL(ctx, legacyESQL)
	}

	diags.AddError("Failed to parse legacy metric chart", "Unable to parse as legacy metric (ESQL or NoESQL)")
	return diags
}

func (m *legacyMetricConfigModel) fromAPINoESQL(ctx context.Context, api kbapi.LegacyMetricNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.DatasetJSON = jsontypes.NewNormalizedValue(string(datasetBytes))

	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	metricBytes, err := api.Metric.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal metric", err.Error())
		return diags
	}
	m.MetricJSON = customtypes.NewJSONWithDefaultsValue(
		string(metricBytes),
		populateLegacyMetricMetricDefaults,
	)

	return diags
}

func (m *legacyMetricConfigModel) fromAPIESQL(ctx context.Context, api kbapi.LegacyMetricESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.DatasetJSON = jsontypes.NewNormalizedValue(string(datasetBytes))

	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	metricBytes, err := json.Marshal(api.Metric)
	if err != nil {
		diags.AddError("Failed to marshal metric", err.Error())
		return diags
	}
	m.MetricJSON = customtypes.NewJSONWithDefaultsValue[map[string]any](
		string(metricBytes),
		populateLegacyMetricMetricDefaults,
	)

	m.Query = nil

	return diags
}

func (m *legacyMetricConfigModel) toAPI() (kbapi.LegacyMetricChartSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.LegacyMetricChartSchema

	if m == nil {
		diags.AddError("Legacy metric config is nil", "Legacy metric configuration is required")
		return result, diags
	}

	datasetType, typeDiags := m.datasetType()
	diags.Append(typeDiags...)
	if diags.HasError() {
		return result, diags
	}

	switch datasetType {
	case legacyMetricDatasetTypeESQL, legacyMetricDatasetTypeTable:
		api := kbapi.LegacyMetricESQL{
			Type: kbapi.LegacyMetricESQLTypeLegacyMetric,
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
			sampling := float32(m.Sampling.ValueFloat64())
			api.Sampling = &sampling
		}

		if len(m.Filters) > 0 {
			filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
			for i, filterModel := range m.Filters {
				filter, filterDiags := filterModel.toAPI()
				diags.Append(filterDiags...)
				filters[i] = filter
			}
			api.Filters = &filters
		}

		if m.Query != nil {
			diags.AddError("Invalid legacy metric query", "Query is not supported for ESQL legacy metric charts")
			return result, diags
		}

		if typeutils.IsKnown(m.DatasetJSON) {
			if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &api.Dataset); err != nil {
				diags.AddError("Failed to unmarshal dataset", err.Error())
				return result, diags
			}
		}

		if !typeutils.IsKnown(m.MetricJSON) {
			diags.AddError("Missing metric", "Metric is required for ESQL legacy metric charts")
			return result, diags
		}

		var metric legacyMetricESQLMetricAPIModel
		if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &metric); err != nil {
			diags.AddError("Failed to unmarshal metric", err.Error())
			return result, diags
		}
		api.Metric.Alignments = metric.Alignments
		api.Metric.ApplyColorTo = metric.ApplyColorTo
		api.Metric.Color = metric.Color
		api.Metric.Column = metric.Column
		api.Metric.Format = metric.Format
		api.Metric.Label = metric.Label
		api.Metric.Operation = metric.Operation
		api.Metric.Size = metric.Size

		if err := result.FromLegacyMetricESQL(api); err != nil {
			diags.AddError("Failed to marshal legacy metric ESQL", err.Error())
		}
		return result, diags
	case legacyMetricDatasetTypeDataView, legacyMetricDatasetTypeIndex:
		api := kbapi.LegacyMetricNoESQL{
			Type: kbapi.LegacyMetricNoESQLTypeLegacyMetric,
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
			sampling := float32(m.Sampling.ValueFloat64())
			api.Sampling = &sampling
		}

		if len(m.Filters) > 0 {
			filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
			for i, filterModel := range m.Filters {
				filter, filterDiags := filterModel.toAPI()
				diags.Append(filterDiags...)
				filters[i] = filter
			}
			api.Filters = &filters
		}

		if m.Query != nil {
			api.Query = m.Query.toAPI()
		} else {
			diags.AddError("Missing legacy metric query", "Query is required for non-ESQL legacy metric charts")
			return result, diags
		}

		if typeutils.IsKnown(m.DatasetJSON) {
			if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &api.Dataset); err != nil {
				diags.AddError("Failed to unmarshal dataset", err.Error())
				return result, diags
			}
		}

		if !typeutils.IsKnown(m.MetricJSON) {
			diags.AddError("Missing metric", "Metric is required for legacy metric charts")
			return result, diags
		}
		if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
			diags.AddError("Failed to unmarshal metric", err.Error())
			return result, diags
		}

		if err := result.FromLegacyMetricNoESQL(api); err != nil {
			diags.AddError("Failed to marshal legacy metric", err.Error())
		}
		return result, diags
	default:
		diags.AddError("Unsupported legacy metric dataset", "Legacy metric dataset type must be one of dataView, index, esql, or table")
		return result, diags
	}
}

func (m *legacyMetricConfigModel) datasetType() (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(m.DatasetJSON) {
		diags.AddError("Missing dataset", "Dataset is required for legacy metric charts")
		return "", diags
	}

	var datasetType struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &datasetType); err != nil {
		diags.AddError("Failed to decode dataset type", err.Error())
		return "", diags
	}

	if datasetType.Type == "" {
		diags.AddError("Missing dataset type", "Dataset type is required for legacy metric charts")
		return "", diags
	}

	return datasetType.Type, diags
}

type legacyMetricESQLMetricAPIModel struct {
	Alignments *struct {
		Labels *kbapi.LegacyMetricESQLMetricAlignmentsLabels `json:"labels,omitempty"`
		Value  *kbapi.LegacyMetricESQLMetricAlignmentsValue  `json:"value,omitempty"`
	} `json:"alignments,omitempty"`
	ApplyColorTo *kbapi.LegacyMetricESQLMetricApplyColorTo `json:"apply_color_to,omitempty"`
	Color        kbapi.ColorByValueAbsolute                `json:"color"`
	Column       string                                    `json:"column"`
	Format       kbapi.FormatTypeSchema                    `json:"format"`
	Label        *string                                   `json:"label,omitempty"`
	Operation    kbapi.LegacyMetricESQLMetricOperation     `json:"operation"`
	Size         *kbapi.LegacyMetricESQLMetricSize         `json:"size,omitempty"`
}
