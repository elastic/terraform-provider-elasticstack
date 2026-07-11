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

package mlanomalyswimlane

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
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeMlAnomalySwimlane) diag.Diagnostics {
	cfg := pm.MlAnomalySwimlaneConfig
	if cfg == nil {
		var diags diag.Diagnostics
		diags.AddError(
			"Missing ML anomaly swim lane panel configuration",
			"ML anomaly swim lane panels require `ml_anomaly_swimlane_config`.",
		)
		return diags
	}

	jobIDs := typeutils.ValueStringSlice(cfg.JobIDs)

	var union kbapi.KibanaHTTPAPIsMlAnomalySwimlane
	switch cfg.SwimlaneType.ValueString() {
	case swimlaneTypeViewBy:
		branch := kbapi.KibanaHTTPAPIsMlAnomalySwimlane1{
			SwimlaneType: kbapi.ViewBy,
			JobIds:       jobIDs,
			ViewBy:       cfg.ViewBy.ValueString(),
		}
		mlAnomalySwimlaneApplyOptionalFields(&branch.Description, &branch.HideBorder, &branch.HideTitle, &branch.PerPage, &branch.TimeRange, &branch.Title, cfg)
		if err := union.FromKibanaHTTPAPIsMlAnomalySwimlane1(branch); err != nil {
			var diags diag.Diagnostics
			diags.AddError("Invalid ML anomaly swim lane configuration", err.Error())
			return diags
		}
	default:
		branch := kbapi.KibanaHTTPAPIsMlAnomalySwimlane0{
			SwimlaneType: kbapi.Overall,
			JobIds:       jobIDs,
		}
		mlAnomalySwimlaneApplyOptionalFields(&branch.Description, &branch.HideBorder, &branch.HideTitle, &branch.PerPage, &branch.TimeRange, &branch.Title, cfg)
		if err := union.FromKibanaHTTPAPIsMlAnomalySwimlane0(branch); err != nil {
			var diags diag.Diagnostics
			diags.AddError("Invalid ML anomaly swim lane configuration", err.Error())
			return diags
		}
	}

	panel.Config = union
	return nil
}

func mlAnomalySwimlaneApplyOptionalFields(
	description **string,
	hideBorder, hideTitle **bool,
	perPage **float32,
	timeRange **kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema,
	title **string,
	cfg *models.MlAnomalySwimlaneConfigModel,
) {
	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		title, description, hideTitle, hideBorder)
	if typeutils.IsKnown(cfg.PerPage) {
		v := cfg.PerPage.ValueFloat32()
		*perPage = &v
	}
	if cfg.TimeRange != nil {
		*timeRange = lenscommon.TimeRangeModelToAPI(cfg.TimeRange)
	}
}

// PopulateFromAPI maps Kibana ML anomaly swim lane config into Terraform panel state while preserving prior null intent.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsMlAnomalySwimlane) diag.Diagnostics {
	if prior == nil {
		cfg, diags := mlAnomalySwimlaneConfigFromAPIImport(apiConfig)
		if diags.HasError() {
			return diags
		}
		pm.MlAnomalySwimlaneConfig = cfg
		return nil
	}

	if pm.MlAnomalySwimlaneConfig == nil && prior.MlAnomalySwimlaneConfig != nil {
		cfg, diags := mlAnomalySwimlaneConfigFromAPIImport(apiConfig)
		if diags.HasError() {
			return diags
		}
		pm.MlAnomalySwimlaneConfig = cfg
	}

	existing := pm.MlAnomalySwimlaneConfig
	if existing == nil {
		return nil
	}

	if cfg1, err := apiConfig.AsKibanaHTTPAPIsMlAnomalySwimlane1(); err == nil && cfg1.SwimlaneType == kbapi.ViewBy {
		mlAnomalySwimlaneMergeViewByFromAPI(existing, prior.MlAnomalySwimlaneConfig, cfg1)
		return nil
	}

	cfg0, err := apiConfig.AsKibanaHTTPAPIsMlAnomalySwimlane0()
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid ML anomaly swim lane panel configuration on read", err.Error())
		return diags
	}
	mlAnomalySwimlaneMergeOverallFromAPI(existing, prior.MlAnomalySwimlaneConfig, cfg0)
	return nil
}

func mlAnomalySwimlaneConfigFromAPIImport(apiConfig kbapi.KibanaHTTPAPIsMlAnomalySwimlane) (*models.MlAnomalySwimlaneConfigModel, diag.Diagnostics) {
	if cfg1, err := apiConfig.AsKibanaHTTPAPIsMlAnomalySwimlane1(); err == nil && cfg1.SwimlaneType == kbapi.ViewBy {
		return mlAnomalySwimlaneConfigFromViewByAPI(cfg1), nil
	}

	cfg0, err := apiConfig.AsKibanaHTTPAPIsMlAnomalySwimlane0()
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid ML anomaly swim lane panel configuration on read", err.Error())
		return nil, diags
	}
	return mlAnomalySwimlaneConfigFromOverallAPI(cfg0), nil
}

func mlAnomalySwimlaneConfigFromOverallAPI(cfg kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) *models.MlAnomalySwimlaneConfigModel {
	out := &models.MlAnomalySwimlaneConfigModel{
		SwimlaneType: types.StringValue(string(cfg.SwimlaneType)),
		JobIDs:       typeutils.StringSliceValue(cfg.JobIds),
		ViewBy:       types.StringNull(),
		Title:        types.StringPointerValue(cfg.Title),
		Description:  types.StringPointerValue(cfg.Description),
		HideTitle:    types.BoolPointerValue(cfg.HideTitle),
		HideBorder:   types.BoolPointerValue(cfg.HideBorder),
		PerPage:      types.Float32PointerValue(cfg.PerPage),
		TimeRange:    panelkit.TimeRangeFromAPI(cfg.TimeRange, nil),
	}
	return out
}

func mlAnomalySwimlaneConfigFromViewByAPI(cfg kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) *models.MlAnomalySwimlaneConfigModel {
	out := &models.MlAnomalySwimlaneConfigModel{
		SwimlaneType: types.StringValue(string(cfg.SwimlaneType)),
		JobIDs:       typeutils.StringSliceValue(cfg.JobIds),
		ViewBy:       types.StringValue(cfg.ViewBy),
		Title:        types.StringPointerValue(cfg.Title),
		Description:  types.StringPointerValue(cfg.Description),
		HideTitle:    types.BoolPointerValue(cfg.HideTitle),
		HideBorder:   types.BoolPointerValue(cfg.HideBorder),
		PerPage:      types.Float32PointerValue(cfg.PerPage),
		TimeRange:    panelkit.TimeRangeFromAPI(cfg.TimeRange, nil),
	}
	return out
}

func mlAnomalySwimlaneMergeOverallFromAPI(existing, prior *models.MlAnomalySwimlaneConfigModel, cfg kbapi.KibanaHTTPAPIsMlAnomalySwimlane0) {
	existing.SwimlaneType = types.StringValue(string(cfg.SwimlaneType))
	existing.JobIDs = typeutils.StringSliceValue(cfg.JobIds)
	existing.ViewBy = types.StringNull()
	mlAnomalySwimlaneMergeOptionalFromAPI(existing, prior, cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder, cfg.PerPage, cfg.TimeRange)
}

func mlAnomalySwimlaneMergeViewByFromAPI(existing, prior *models.MlAnomalySwimlaneConfigModel, cfg kbapi.KibanaHTTPAPIsMlAnomalySwimlane1) {
	existing.SwimlaneType = types.StringValue(string(cfg.SwimlaneType))
	existing.JobIDs = typeutils.StringSliceValue(cfg.JobIds)
	existing.ViewBy = types.StringValue(cfg.ViewBy)
	mlAnomalySwimlaneMergeOptionalFromAPI(existing, prior, cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder, cfg.PerPage, cfg.TimeRange)
}

func mlAnomalySwimlaneMergeOptionalFromAPI(
	existing, prior *models.MlAnomalySwimlaneConfigModel,
	title, description *string,
	hideTitle, hideBorder *bool,
	perPage *float32,
	timeRange *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema,
) {
	panelkit.ApplyPresentationFromAPI(&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder,
		title, description, hideTitle, hideBorder)
	existing.PerPage = panelkit.PreserveFloat32(existing.PerPage, perPage)

	var priorTR *models.TimeRangeModel
	if prior != nil {
		priorTR = prior.TimeRange
	}
	existing.TimeRange = panelkit.MergeTimeRange(existing.TimeRange, timeRange, priorTR)

	if prior != nil {
		mlAnomalySwimlanePreserveNullIntentFromPrior(prior, existing)
	}
}

func mlAnomalySwimlanePreserveNullIntentFromPrior(prior, existing *models.MlAnomalySwimlaneConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	switch existing.SwimlaneType.ValueString() {
	case swimlaneTypeOverall:
		existing.ViewBy = types.StringNull()
	case swimlaneTypeViewBy:
		if prior.SwimlaneType.ValueString() == swimlaneTypeViewBy && !typeutils.IsKnown(prior.ViewBy) {
			existing.ViewBy = types.StringNull()
		}
	}
	panelkit.NullPreserveFloat32FromPrior(prior.PerPage, &existing.PerPage)
	panelkit.NullPreserveBaseFromPrior(prior.Title, prior.Description, prior.HideTitle, prior.HideBorder,
		&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder)
	existing.TimeRange = panelkit.PreserveTimeRangeNullIntentFromPrior(prior.TimeRange, existing.TimeRange)
}
