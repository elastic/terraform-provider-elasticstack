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

package mlanomalycharts

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes Terraform state from pm into panel's typed API config.
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalyCharts) diag.Diagnostics {
	cfg := pm.MlAnomalyChartsConfig
	if cfg == nil {
		var diags diag.Diagnostics
		diags.AddError(
			"Missing ML anomaly charts panel configuration",
			"ML anomaly charts panels require `ml_anomaly_charts_config`.",
		)
		return diags
	}

	apiConfig := kbapi.KibanaHTTPAPIsMlAnomalyCharts{
		JobIds: typeutils.ValueStringSlice(cfg.JobIDs),
	}

	apiConfig.MaxSeriesToPlot = typeutils.Int64ToFloat32Ptr(cfg.MaxSeriesToPlot)
	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&apiConfig.Title, &apiConfig.Description, &apiConfig.HideTitle, &apiConfig.HideBorder)
	if cfg.TimeRange != nil {
		apiConfig.TimeRange = lenscommon.TimeRangeModelToAPI(cfg.TimeRange)
	}

	if len(cfg.SeverityThreshold) > 0 {
		items, diags := buildSeverityThresholdItems(cfg.SeverityThreshold)
		if diags.HasError() {
			return diags
		}
		apiConfig.SeverityThreshold = items
	}

	panel.Config = apiConfig
	return nil
}

// PopulateFromAPI maps Kibana ML anomaly charts config into Terraform panel state while preserving prior null intent.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsMlAnomalyCharts) diag.Diagnostics {
	if prior == nil {
		cfg, diags := mlAnomalyChartsConfigFromAPIImport(apiConfig)
		if diags.HasError() {
			return diags
		}
		pm.MlAnomalyChartsConfig = cfg
		return nil
	}

	if pm.MlAnomalyChartsConfig == nil && prior.MlAnomalyChartsConfig != nil {
		cfg, diags := mlAnomalyChartsConfigFromAPIImport(apiConfig)
		if diags.HasError() {
			return diags
		}
		pm.MlAnomalyChartsConfig = cfg
	}

	existing := pm.MlAnomalyChartsConfig
	if existing == nil {
		return nil
	}

	existing.JobIDs = typeutils.StringSliceValue(apiConfig.JobIds)
	return mlAnomalyChartsMergeOptionalFromAPI(existing, prior.MlAnomalyChartsConfig, apiConfig)
}

func mlAnomalyChartsConfigFromAPIImport(apiConfig kbapi.KibanaHTTPAPIsMlAnomalyCharts) (*models.MlAnomalyChartsConfigModel, diag.Diagnostics) {
	severityThreshold, diags := readSeverityThresholdFromAPI(apiConfig.SeverityThreshold, nil)
	if diags.HasError() {
		return nil, diags
	}

	return &models.MlAnomalyChartsConfigModel{
		JobIDs:            typeutils.StringSliceValue(apiConfig.JobIds),
		MaxSeriesToPlot:   types.Int64PointerValue(typeutils.Float32PointerToInt64Pointer(apiConfig.MaxSeriesToPlot)),
		SeverityThreshold: severityThreshold,
		TimeRange:         panelkit.TimeRangeFromAPI(apiConfig.TimeRange, nil),
		Title:             types.StringPointerValue(apiConfig.Title),
		Description:       types.StringPointerValue(apiConfig.Description),
		HideTitle:         types.BoolPointerValue(apiConfig.HideTitle),
		HideBorder:        types.BoolPointerValue(apiConfig.HideBorder),
	}, nil
}

func mlAnomalyChartsMergeOptionalFromAPI(
	existing, prior *models.MlAnomalyChartsConfigModel,
	apiConfig kbapi.KibanaHTTPAPIsMlAnomalyCharts,
) diag.Diagnostics {
	panelkit.ApplyPresentationFromAPI(&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder,
		apiConfig.Title, apiConfig.Description, apiConfig.HideTitle, apiConfig.HideBorder)
	existing.MaxSeriesToPlot = panelkit.PreserveInt64(existing.MaxSeriesToPlot, typeutils.Float32PointerToInt64Pointer(apiConfig.MaxSeriesToPlot))

	var priorTR *models.TimeRangeModel
	var priorSeverity []models.MlAnomalyChartsSeverityThresholdModel
	if prior != nil {
		priorTR = prior.TimeRange
		priorSeverity = prior.SeverityThreshold
	}
	existing.TimeRange = panelkit.MergeTimeRange(existing.TimeRange, apiConfig.TimeRange, priorTR)

	if len(priorSeverity) > 0 || len(existing.SeverityThreshold) > 0 {
		severityThreshold, diags := readSeverityThresholdFromAPI(apiConfig.SeverityThreshold, priorSeverity)
		if diags.HasError() {
			return diags
		}
		existing.SeverityThreshold = severityThreshold
	}

	if prior != nil {
		mlAnomalyChartsPreserveNullIntentFromPrior(prior, existing)
	}
	return nil
}

func mlAnomalyChartsPreserveNullIntentFromPrior(prior, existing *models.MlAnomalyChartsConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	panelkit.NullPreserveInt64FromPrior(prior.MaxSeriesToPlot, &existing.MaxSeriesToPlot)
	panelkit.NullPreservePresentationFromPrior(prior.Title, prior.Description, prior.HideTitle, prior.HideBorder,
		&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder)
	if len(prior.SeverityThreshold) == 0 {
		existing.SeverityThreshold = nil
	}
	existing.TimeRange = panelkit.PreserveTimeRangeNullIntentFromPrior(prior.TimeRange, existing.TimeRange)
}
