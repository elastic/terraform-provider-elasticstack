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

package lenslegacymetric

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

const (
	datasetTypeDataViewReference = "data_view_reference"
	datasetTypeDataViewSpec      = "data_view_spec"
)

func legacyMetricConfigFromAPINoESQL(
	ctx context.Context,
	m *models.LegacyMetricConfigModel,
	prior *models.LegacyMetricConfigModel,
	api kbapi.KibanaHTTPAPIsLegacyMetricNoESQL,
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
	mv, ok := lenscommon.MarshalToJSONWithDefaults(metricBytes, err, "metric", lenscommon.PopulateLegacyMetricMetricDefaults, &diags)
	if !ok {
		return diags
	}
	m.MetricJSON = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, m.MetricJSON, mv, &diags)

	if !lenscommon.PopulateLensChartPresentation(ctx, &m.LensChartPresentationTFModel, prior, api.TimeRange, api.HideTitle, api.HideBorder, api.References, api.Drilldowns, &diags) {
		return diags
	}

	return diags
}

func legacyMetricConfigToAPI(m *models.LegacyMetricConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result lenscommon.VisByValueConfig0

	if m == nil {
		diags.AddError("Legacy metric config is nil", "Legacy metric configuration is required")
		return result, diags
	}

	datasetType, typeDiags := legacyMetricConfigDatasetType(m)
	diags.Append(typeDiags...)
	if diags.HasError() {
		return result, diags
	}

	switch datasetType {
	case datasetTypeDataViewReference, datasetTypeDataViewSpec:
		api := kbapi.KibanaHTTPAPIsLegacyMetricNoESQL{
			Type: kbapi.LegacyMetric,
		}

		api.Title, api.Description, api.IgnoreGlobalFilters, api.Sampling = lenscommon.LensChartBaseFieldsForAPI(m.LensChartBaseTFModel)

		api.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

		if m.Query != nil {
			api.Query = lenscommon.FilterSimpleToAPI(m.Query)
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

		writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
		diags.Append(presDiags...)
		if presDiags.HasError() {
			return result, diags
		}

		diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsLegacyMetricNoESQL_Drilldowns_Item](
			writes, &api.TimeRange, &api.HideTitle, &api.HideBorder, &api.References, &api.Drilldowns,
		)...)

		if err := result.FromKibanaHTTPAPIsLegacyMetricNoESQL(api); err != nil {
			diags.AddError("Failed to marshal legacy metric", err.Error())
		}
		return result, diags
	default:
		diags.AddError("Unsupported legacy metric dataset", "Legacy metric dataset type must be one of data_view_reference or data_view_spec")
		return result, diags
	}
}

func legacyMetricConfigDatasetType(m *models.LegacyMetricConfigModel) (string, diag.Diagnostics) {
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
