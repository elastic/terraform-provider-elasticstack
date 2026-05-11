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

func newRegionMapPanelConfigConverter() regionMapPanelConfigConverter {
	return regionMapPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.RegionMapNoESQLTypeRegionMap),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.RegionMapConfig != nil },
		},
	}
}

type regionMapPanelConfigConverter struct {
	lensVisualizationBase
}

func (c regionMapPanelConfigConverter) populateFromAttributes(ctx context.Context, dashboard *dashboardModel, pm *panelModel, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	var prior *regionMapConfigModel
	if pm.RegionMapConfig != nil {
		cpy := *pm.RegionMapConfig
		prior = &cpy
	}
	pm.RegionMapConfig = &regionMapConfigModel{}

	if noESQL, err := attrs.AsRegionMapNoESQL(); err == nil && !isRegionMapNoESQLCandidateActuallyESQL(noESQL) {
		return pm.RegionMapConfig.fromAPINoESQL(ctx, dashboard, prior, noESQL)
	}

	regionMapESQL, err := attrs.AsRegionMapESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.RegionMapConfig.fromAPIESQL(ctx, dashboard, prior, regionMapESQL)
}

func isRegionMapNoESQLCandidateActuallyESQL(api kbapi.RegionMapNoESQL) bool {
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

func (c regionMapPanelConfigConverter) buildAttributes(pm panelModel, dashboard *dashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	configModel := *pm.RegionMapConfig

	attrs, regionDiags := configModel.toAPI(dashboard)
	diags.Append(regionDiags...)
	return attrs, diags
}

type regionMapConfigModel struct {
	lensChartPresentationTFModel
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DataSourceJSON      jsontypes.Normalized                              `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []chartFilterJSONModel                            `tfsdk:"filters"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
	RegionJSON          jsontypes.Normalized                              `tfsdk:"region_json"`
}

func (m *regionMapConfigModel) populateCommonFields(
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

func (m *regionMapConfigModel) fromAPINoESQL(ctx context.Context, dashboard *dashboardModel, prior *regionMapConfigModel, api kbapi.RegionMapNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := api.DataSource.MarshalJSON()
	if !m.populateCommonFields(api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling, datasetBytes, datasetErr, api.Filters, &diags) {
		return diags
	}

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric", populateRegionMapMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = preservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

	regionBytes, err := api.Region.MarshalJSON()
	rv, ok := marshalToNormalized(regionBytes, err, "region", &diags)
	if !ok {
		return diags
	}
	m.RegionJSON = rv

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

func (m *regionMapConfigModel) fromAPIESQL(ctx context.Context, dashboard *dashboardModel, prior *regionMapConfigModel, api kbapi.RegionMapESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	if !m.populateCommonFields(api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling, datasetBytes, datasetErr, api.Filters, &diags) {
		return diags
	}

	m.Query = nil

	metricBytes, err := json.Marshal(api.Metric)
	mv, ok := marshalToJSONWithDefaults(metricBytes, err, "metric", populateRegionMapMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = preservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

	regionBytes, err := json.Marshal(api.Region)
	rv, ok := marshalToNormalized(regionBytes, err, "region", &diags)
	if !ok {
		return diags
	}
	m.RegionJSON = rv

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

func (m *regionMapConfigModel) toAPI(dashboard *dashboardModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	if m.Query != nil && typeutils.IsKnown(m.Query.Expression) {
		api := kbapi.RegionMapNoESQL{
			Type: kbapi.RegionMapNoESQLTypeRegionMap,
		}

		if typeutils.IsKnown(m.Title) {
			api.Title = m.Title.ValueStringPointer()
		}
		if typeutils.IsKnown(m.Description) {
			api.Description = m.Description.ValueStringPointer()
		}
		if typeutils.IsKnown(m.DataSourceJSON) {
			if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
				diags.AddError("Failed to unmarshal data_source_json", err.Error())
				return attrs, diags
			}
		}
		if typeutils.IsKnown(m.IgnoreGlobalFilters) {
			api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
		}
		if typeutils.IsKnown(m.Sampling) {
			sampling := float32(m.Sampling.ValueFloat64())
			api.Sampling = &sampling
		}
		api.Query = m.Query.toAPI()

		api.Filters = buildFiltersForAPI(m.Filters, &diags)

		if typeutils.IsKnown(m.MetricJSON) {
			if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
				diags.AddError("Failed to unmarshal metric", err.Error())
				return attrs, diags
			}
		}
		if typeutils.IsKnown(m.RegionJSON) {
			if err := json.Unmarshal([]byte(m.RegionJSON.ValueString()), &api.Region); err != nil {
				diags.AddError("Failed to unmarshal region", err.Error())
				return attrs, diags
			}
		}

		writes, presDiags := lensChartPresentationWritesFor(dashboard, m.lensChartPresentationTFModel)
		diags.Append(presDiags...)
		if presDiags.HasError() {
			return attrs, diags
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
			items, ddDiags := decodeLensDrilldownSlice[kbapi.RegionMapNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
			diags.Append(ddDiags...)
			if !ddDiags.HasError() {
				api.Drilldowns = &items
			}
		}

		if err := attrs.FromRegionMapNoESQL(api); err != nil {
			diags.AddError("Failed to create region map schema", err.Error())
		}
		return attrs, diags
	}

	api := kbapi.RegionMapESQL{
		Type: kbapi.RegionMapESQLTypeRegionMap,
	}

	if typeutils.IsKnown(m.Title) {
		api.Title = m.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Description) {
		api.Description = m.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(m.DataSourceJSON) {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal data_source_json", err.Error())
			return attrs, diags
		}
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.Sampling) {
		sampling := float32(m.Sampling.ValueFloat64())
		api.Sampling = &sampling
	}

	api.Filters = buildFiltersForAPI(m.Filters, &diags)

	if typeutils.IsKnown(m.MetricJSON) {
		if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
			diags.AddError("Failed to unmarshal metric", err.Error())
			return attrs, diags
		}
	}
	if typeutils.IsKnown(m.RegionJSON) {
		if err := json.Unmarshal([]byte(m.RegionJSON.ValueString()), &api.Region); err != nil {
			diags.AddError("Failed to unmarshal region", err.Error())
			return attrs, diags
		}
	}

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.lensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return attrs, diags
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
		items, ddDiags := decodeLensDrilldownSlice[kbapi.RegionMapESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			api.Drilldowns = &items
		}
	}

	if err := attrs.FromRegionMapESQL(api); err != nil {
		diags.AddError("Failed to create region map schema", err.Error())
	}
	return attrs, diags
}
