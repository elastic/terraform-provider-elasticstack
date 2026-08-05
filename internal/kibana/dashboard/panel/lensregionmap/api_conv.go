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

func regionMapConfigFromAPINoESQL(
	ctx context.Context,
	m *models.RegionMapConfigModel,
	_ *models.RegionMapConfigModel,
	api kbapi.KibanaHTTPAPIsRegionMapNoESQL,
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

	regionBytes, err := json.Marshal(api.Region)
	rv, ok := lenscommon.WrapNormalizedJSON(regionBytes, err, "region", &diags)
	if !ok {
		return diags
	}
	m.RegionJSON = rv

	return diags
}

func regionMapConfigFromAPIESQL(
	ctx context.Context,
	m *models.RegionMapConfigModel,
	_ *models.RegionMapConfigModel,
	api kbapi.KibanaHTTPAPIsRegionMapESQL,
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

	return diags
}

func regionMapConfigToAPI(m *models.RegionMapConfigModel) (lenscommon.LensByValueConfig, diag.Diagnostics) {
	var result lenscommon.LensByValueConfig
	var diags diag.Diagnostics
	if m == nil {
		return result, diags
	}

	if lenscommon.ConfigUsesESQL(m.Query) {
		chart, chartDiags := regionMapConfigToAPIESQL(m)
		diags.Append(chartDiags...)
		if chartDiags.HasError() {
			return result, diags
		}
		var err error
		result.Chart, err = regionMapChartToLensAPI(chart)
		if err != nil {
			diags.AddError("Failed to create Lens chart config", err.Error())
			return result, diags
		}
	} else {
		chart, chartDiags := regionMapConfigToAPINoESQL(m)
		diags.Append(chartDiags...)
		if chartDiags.HasError() {
			return result, diags
		}
		var err error
		result.Chart, err = regionMapChartToLensAPI(chart)
		if err != nil {
			diags.AddError("Failed to create Lens chart config", err.Error())
			return result, diags
		}
	}
	if err := preserveRegionMapRawDimensions(&result.Chart, m); err != nil {
		diags.AddError("Failed to preserve region map dimensions", err.Error())
		return result, diags
	}

	writes, presentationDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presentationDiags...)
	if presentationDiags.HasError() {
		return result, diags
	}
	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config_0_Drilldowns_Item](
		writes, &result.Presentation.TimeRange, &result.Presentation.HideTitle, &result.Presentation.HideBorder,
		&result.Presentation.References, &result.Presentation.Drilldowns,
	)...)

	return result, diags
}

func preserveRegionMapRawDimensions(chart *kbapi.KibanaHTTPAPIsLensApiConfig, m *models.RegionMapConfigModel) error {
	raw, err := json.Marshal(chart)
	if err != nil {
		return err
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if typeutils.IsKnown(m.MetricJSON) {
		payload["metric"] = json.RawMessage(m.MetricJSON.ValueString())
	}
	if typeutils.IsKnown(m.RegionJSON) {
		payload["region"] = json.RawMessage(m.RegionJSON.ValueString())
	}
	raw, err = json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, chart)
}

func regionMapConfigToAPINoESQL(m *models.RegionMapConfigModel) (kbapi.KibanaHTTPAPIsRegionMapNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KibanaHTTPAPIsRegionMapNoESQL{
		Type: kbapi.KibanaHTTPAPIsRegionMapNoESQLTypeRegionMap,
	}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)
	if typeutils.IsKnown(m.DataSourceJSON) {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal data_source_json", err.Error())
			return api, diags
		}
	}
	api.Query = lenscommon.FilterSimpleToAPI(m.Query)

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if typeutils.IsKnown(m.MetricJSON) {
		if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
			diags.AddError("Failed to unmarshal metric", err.Error())
			return api, diags
		}
	}
	if typeutils.IsKnown(m.RegionJSON) {
		if err := json.Unmarshal([]byte(m.RegionJSON.ValueString()), &api.Region); err != nil {
			diags.AddError("Failed to unmarshal region", err.Error())
			return api, diags
		}
	}

	return api, diags
}

func regionMapConfigToAPIESQL(m *models.RegionMapConfigModel) (kbapi.KibanaHTTPAPIsRegionMapESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	api := kbapi.KibanaHTTPAPIsRegionMapESQL{
		Type: kbapi.KibanaHTTPAPIsRegionMapESQLTypeRegionMap,
	}

	api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)
	if typeutils.IsKnown(m.DataSourceJSON) {
		if err := json.Unmarshal([]byte(m.DataSourceJSON.ValueString()), &api.DataSource); err != nil {
			diags.AddError("Failed to unmarshal data_source_json", err.Error())
			return api, diags
		}
	}

	api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	if typeutils.IsKnown(m.MetricJSON) {
		if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
			diags.AddError("Failed to unmarshal metric", err.Error())
			return api, diags
		}
	}
	if typeutils.IsKnown(m.RegionJSON) {
		if err := json.Unmarshal([]byte(m.RegionJSON.ValueString()), &api.Region); err != nil {
			diags.AddError("Failed to unmarshal region", err.Error())
			return api, diags
		}
	}

	return api, diags
}
