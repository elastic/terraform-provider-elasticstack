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

package lensregionmap

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func isRegionMapNoESQLCandidateActuallyESQL(api kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanel) bool {
	return lenscommon.LensDataSourceIsESQLOrTable(api.DataSource.MarshalJSON())
}

func regionMapConfigFromAPINoESQL(
	ctx context.Context,
	m *models.RegionMapConfigModel,
	prior *models.RegionMapConfigModel,
	api kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	datasetBytes, datasetErr := api.DataSource.MarshalJSON()
	base, ok := lenscommon.PopulateLensChartBaseFromAPI(
		api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling,
		datasetBytes, datasetErr, "data_source_json", api.Filters, &diags,
	)
	if !ok {
		return diags
	}
	m.LensChartBaseTFModel = base

	m.Query = &models.FilterSimpleModel{}
	lenscommon.FilterSimpleFromAPI(m.Query, api.Query)

	metricBytes, err := api.Metric.MarshalJSON()
	mv, ok := lenscommon.MarshalToJSONWithDefaults(metricBytes, err, "metric", lenscommon.PopulateRegionMapMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

	regionBytes, err := api.Region.MarshalJSON()
	rv, ok := lenscommon.WrapNormalizedJSON(regionBytes, err, "region", &diags)
	if !ok {
		return diags
	}
	m.RegionJSON = rv

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}

	return diags
}

func regionMapConfigFromAPIESQL(
	ctx context.Context,
	m *models.RegionMapConfigModel,
	prior *models.RegionMapConfigModel,
	api kbapi.KibanaHTTPAPIsRegionMapESQLByValuePanel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	datasetBytes, datasetErr := json.Marshal(api.DataSource)
	base, ok := lenscommon.PopulateLensChartBaseFromAPI(
		api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling,
		datasetBytes, datasetErr, "data_source_json", api.Filters, &diags,
	)
	if !ok {
		return diags
	}
	m.LensChartBaseTFModel = base

	m.Query = nil

	metricBytes, err := json.Marshal(api.Metric)
	mv, ok := lenscommon.MarshalToJSONWithDefaults(metricBytes, err, "metric", lenscommon.PopulateRegionMapMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

	regionBytes, err := json.Marshal(api.Region)
	rv, ok := lenscommon.WrapNormalizedJSON(regionBytes, err, "region", &diags)
	if !ok {
		return diags
	}
	m.RegionJSON = rv

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}

	return diags
}

func regionMapConfigToAPI(m *models.RegionMapConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var attrs lenscommon.VisByValueConfig0
	var diags diag.Diagnostics

	if m == nil {
		return attrs, diags
	}

	if m.Query != nil && typeutils.IsKnown(m.Query.Expression) {
		api := kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanel{
			Type: kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanelTypeRegionMap,
		}

		api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)
		if typeutils.IsKnown(m.DataSourceJSON) {
			if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
				diags.AddError("Failed to unmarshal data_source_json", err.Error())
				return attrs, diags
			}
		}
		api.Query = lenscommon.FilterSimpleToAPI(m.Query)

		api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

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

		writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
		diags.Append(presDiags...)
		if presDiags.HasError() {
			return attrs, diags
		}

		diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsRegionMapNoESQLByValuePanel_Drilldowns_Item](
			writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
		)...)

		if err := attrs.FromKibanaHTTPAPIsRegionMapNoESQLByValuePanel(api); err != nil {
			diags.AddError("Failed to create region map schema", err.Error())
		}
		return attrs, diags
	}

	api := kbapi.KibanaHTTPAPIsRegionMapESQLByValuePanel{
		Type: kbapi.KibanaHTTPAPIsRegionMapESQLByValuePanelTypeRegionMap,
	}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)
	if typeutils.IsKnown(m.DataSourceJSON) {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal data_source_json", err.Error())
			return attrs, diags
		}
	}

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

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

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return attrs, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsRegionMapESQLByValuePanel_Drilldowns_Item](
		writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
	)...)

	if err := attrs.FromKibanaHTTPAPIsRegionMapESQLByValuePanel(api); err != nil {
		diags.AddError("Failed to create region map schema", err.Error())
	}
	return attrs, diags
}
