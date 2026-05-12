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
	legacyMetricDatasetTypeDataViewReference = "data_view_reference"
	legacyMetricDatasetTypeDataViewSpec      = "data_view_spec"
	legacyMetricDatasetTypeESQL              = "esql"
	legacyMetricDatasetTypeTable             = "table"
)

func newLegacyMetricPanelConfigConverter() legacyMetricPanelConfigConverter {
	return legacyMetricPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.LegacyMetric),
			hasTFChartBlock: func(blocks *lensByValueChartBlocks) bool {
				return blocks != nil && blocks.LegacyMetricConfig != nil
			},
		},
	}
}

type legacyMetricPanelConfigConverter struct {
	lensVisualizationBase
}

func (c legacyMetricPanelConfigConverter) populateFromAttributes(
	ctx context.Context,
	dashboard *dashboardModel,
	tfPanel *panelModel,
	blocks *lensByValueChartBlocks,
	attrs kbapi.KbnDashboardPanelTypeVisConfig0,
) diag.Diagnostics {
	legacyMetric, err := attrs.AsLegacyMetricNoESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var prior *legacyMetricConfigModel
	if b := lensByValueChartBlocksFromPanel(tfPanel); b != nil && b.LegacyMetricConfig != nil {
		cpy := *b.LegacyMetricConfig
		prior = &cpy
	}
	blocks.LegacyMetricConfig = &legacyMetricConfigModel{}
	return blocks.LegacyMetricConfig.fromAPINoESQL(ctx, dashboard, prior, legacyMetric)
}

func (c legacyMetricPanelConfigConverter) buildAttributes(blocks *lensByValueChartBlocks, dashboard *dashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *blocks.LegacyMetricConfig

	attrs, legacyDiags := configModel.toAPI(dashboard)
	diags.Append(legacyDiags...)
	if diags.HasError() {
		return kbapi.KbnDashboardPanelTypeVisConfig0{}, diags
	}

	return attrs, diags
}

type legacyMetricConfigModel struct {
	lensChartPresentationTFModel
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized                              `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []chartFilterJSONModel                            `tfsdk:"filters"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
}

func (m *legacyMetricConfigModel) populateCommonFields(
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
	dv, ok := marshalToNormalized(datasetBytes, datasetErr, "data_source_json", diags)
	if !ok {
		return false
	}
	m.DataSourceJSON = dv
	m.Filters = populateFiltersFromAPI(filters, diags)
	return !diags.HasError()
}

func (m *legacyMetricConfigModel) fromAPINoESQL(ctx context.Context, dashboard *dashboardModel, prior *legacyMetricConfigModel, api kbapi.LegacyMetricNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := api.DataSource.MarshalJSON()
	if !m.populateCommonFields(api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling, datasetBytes, datasetErr, api.Filters, &diags) {
		return diags
	}

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric", populateLegacyMetricMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = preservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

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
func (m *legacyMetricConfigModel) toAPI(dashboard *dashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.KbnDashboardPanelTypeVisConfig0

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
	case legacyMetricDatasetTypeDataViewReference, legacyMetricDatasetTypeDataViewSpec:
		api := kbapi.LegacyMetricNoESQL{
			Type: kbapi.LegacyMetric,
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

		api.Filters = buildFiltersForAPI(m.Filters, &diags)

		if m.Query != nil {
			api.Query = m.Query.toAPI()
		} else {
			diags.AddError("Missing legacy metric query", "Query is required for non-ESQL legacy metric charts")
			return result, diags
		}

		if typeutils.IsKnown(m.DataSourceJSON) {
			if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
				diags.AddError("Failed to unmarshal legacy_metric_config.data_source_json", err.Error())
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

		writes, presDiags := lensChartPresentationWritesFor(dashboard, m.lensChartPresentationTFModel)
		diags.Append(presDiags...)
		if presDiags.HasError() {
			return result, diags
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
			items, ddDiags := decodeLensDrilldownSlice[kbapi.LegacyMetricNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
			diags.Append(ddDiags...)
			if !ddDiags.HasError() {
				api.Drilldowns = &items
			}
		}

		if err := result.FromLegacyMetricNoESQL(api); err != nil {
			diags.AddError("Failed to marshal legacy metric", err.Error())
		}
		return result, diags
	default:
		diags.AddError("Unsupported legacy metric dataset", "Legacy metric dataset type must be one of data_view_reference or data_view_spec")
		return result, diags
	}
}

func (m *legacyMetricConfigModel) datasetType() (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(m.DataSourceJSON) {
		diags.AddError("Missing dataset", "Dataset is required for legacy metric charts")
		return "", diags
	}

	var datasetType struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &datasetType); err != nil {
		diags.AddError("Failed to decode dataset type", err.Error())
		return "", diags
	}

	if datasetType.Type == "" {
		diags.AddError("Missing dataset type", "Dataset type is required for legacy metric charts")
		return "", diags
	}

	return datasetType.Type, diags
}
