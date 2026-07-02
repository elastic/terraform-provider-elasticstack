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

package aiopspatternanalysis

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes Terraform state from pm into the typed API panel config.
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsPatternAnalysis) diag.Diagnostics {
	cfg := pm.AiopsPatternAnalysisConfig
	if cfg == nil {
		return nil
	}

	panel.Config.DataViewId = cfg.DataViewID.ValueString()
	panel.Config.FieldName = cfg.FieldName.ValueString()

	if typeutils.IsKnown(cfg.MinimumTimeRange) {
		v := kbapi.KibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange(cfg.MinimumTimeRange.ValueString())
		panel.Config.MinimumTimeRange = &v
	}
	if typeutils.IsKnown(cfg.RandomSamplerMode) {
		v := kbapi.KibanaHTTPAPIsAiopsPatternAnalysisRandomSamplerMode(cfg.RandomSamplerMode.ValueString())
		panel.Config.RandomSamplerMode = &v
	}
	if typeutils.IsKnown(cfg.RandomSamplerProbability) {
		v := cfg.RandomSamplerProbability.ValueFloat32()
		panel.Config.RandomSamplerProbability = &v
	}

	if typeutils.IsKnown(cfg.Title) {
		panel.Config.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Description) {
		panel.Config.Description = cfg.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.HideTitle) {
		panel.Config.HideTitle = cfg.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.HideBorder) {
		panel.Config.HideBorder = cfg.HideBorder.ValueBoolPointer()
	}
	panel.Config.TimeRange = lenscommon.TimeRangeModelToAPI(cfg.TimeRange)

	return nil
}

// PopulateFromAPI maps the Kibana API panel config into Terraform panel state while preserving
// prior null intent (REQ-009). prior is the prior TF state/plan panel, or nil on import.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, api kbapi.KibanaHTTPAPIsAiopsPatternAnalysis) diag.Diagnostics {
	// On import (prior == nil): populate required fields unconditionally; optional fields only when API non-nil.
	if prior == nil {
		pm.AiopsPatternAnalysisConfig = aiopsPatternAnalysisConfigFromAPIImport(api)
		return nil
	}

	if pm.AiopsPatternAnalysisConfig == nil && prior.AiopsPatternAnalysisConfig != nil {
		pm.AiopsPatternAnalysisConfig = aiopsPatternAnalysisConfigFromAPIImport(api)
	}

	existing := pm.AiopsPatternAnalysisConfig
	if existing == nil {
		return nil
	}

	// Required fields always update from the API.
	existing.DataViewID = types.StringValue(api.DataViewId)
	existing.FieldName = types.StringValue(api.FieldName)

	// Optional enum/float fields: only update from API when already known in state (REQ-009 null-preservation).
	if typeutils.IsKnown(existing.MinimumTimeRange) {
		existing.MinimumTimeRange = patternAnalysisMinimumTimeRangeValue(api.MinimumTimeRange)
	}
	if typeutils.IsKnown(existing.RandomSamplerMode) {
		existing.RandomSamplerMode = patternAnalysisRandomSamplerModeValue(api.RandomSamplerMode)
	}
	existing.RandomSamplerProbability = panelkit.PreserveFloat32(existing.RandomSamplerProbability, api.RandomSamplerProbability)

	existing.Title = panelkit.PreserveString(existing.Title, api.Title)
	existing.Description = panelkit.PreserveString(existing.Description, api.Description)
	existing.HideTitle = panelkit.PreserveBool(existing.HideTitle, api.HideTitle)
	existing.HideBorder = panelkit.PreserveBool(existing.HideBorder, api.HideBorder)

	var priorTR *models.TimeRangeModel
	if prior.AiopsPatternAnalysisConfig != nil {
		priorTR = prior.AiopsPatternAnalysisConfig.TimeRange
	}
	existing.TimeRange = panelkit.MergeTimeRange(existing.TimeRange, api.TimeRange, priorTR)

	if prior.AiopsPatternAnalysisConfig != nil {
		preserveNullIntentFromPrior(prior.AiopsPatternAnalysisConfig, existing)
	}
	return nil
}

func aiopsPatternAnalysisConfigFromAPIImport(api kbapi.KibanaHTTPAPIsAiopsPatternAnalysis) *models.AiopsPatternAnalysisConfigModel {
	cfg := &models.AiopsPatternAnalysisConfigModel{
		DataViewID:        types.StringValue(api.DataViewId),
		FieldName:         types.StringValue(api.FieldName),
		MinimumTimeRange:  patternAnalysisMinimumTimeRangeValue(api.MinimumTimeRange),
		RandomSamplerMode: patternAnalysisRandomSamplerModeValue(api.RandomSamplerMode),
		Title:             types.StringPointerValue(api.Title),
		Description:       types.StringPointerValue(api.Description),
		HideTitle:         types.BoolPointerValue(api.HideTitle),
		HideBorder:        types.BoolPointerValue(api.HideBorder),
	}
	cfg.RandomSamplerProbability = types.Float32PointerValue(api.RandomSamplerProbability)
	cfg.TimeRange = panelkit.TimeRangeFromAPI(api.TimeRange, nil)
	return cfg
}

func preserveNullIntentFromPrior(prior, existing *models.AiopsPatternAnalysisConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	if !typeutils.IsKnown(prior.MinimumTimeRange) {
		existing.MinimumTimeRange = types.StringNull()
	}
	if !typeutils.IsKnown(prior.RandomSamplerMode) {
		existing.RandomSamplerMode = types.StringNull()
	}
	if !typeutils.IsKnown(prior.RandomSamplerProbability) {
		existing.RandomSamplerProbability = types.Float32Null()
	}
	if !typeutils.IsKnown(prior.Title) {
		existing.Title = types.StringNull()
	}
	if !typeutils.IsKnown(prior.Description) {
		existing.Description = types.StringNull()
	}
	if !typeutils.IsKnown(prior.HideTitle) {
		existing.HideTitle = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.HideBorder) {
		existing.HideBorder = types.BoolNull()
	}
	if prior.TimeRange == nil {
		existing.TimeRange = nil
	}
}

func patternAnalysisMinimumTimeRangeValue(v *kbapi.KibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*v))
}

func patternAnalysisRandomSamplerModeValue(v *kbapi.KibanaHTTPAPIsAiopsPatternAnalysisRandomSamplerMode) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*v))
}
