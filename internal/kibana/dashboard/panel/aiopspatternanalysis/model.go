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
		panel.Config.MinimumTimeRange = patternAnalysisMinimumTimeRangeAPIValue(cfg.MinimumTimeRange.ValueString())
	}
	if typeutils.IsKnown(cfg.RandomSamplerMode) {
		panel.Config.RandomSamplerMode = patternAnalysisRandomSamplerModeAPIValue(cfg.RandomSamplerMode.ValueString())
	}
	if typeutils.IsKnown(cfg.RandomSamplerProbability) {
		v := cfg.RandomSamplerProbability.ValueFloat32()
		probability := &kbapi.KibanaHTTPAPIsAiopsPatternAnalysis_RandomSamplerProbability{}
		_ = probability.FromKibanaHTTPAPIsAiopsPatternAnalysisRandomSamplerProbability0(v)
		panel.Config.RandomSamplerProbability = probability
	}

	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&panel.Config.Title, &panel.Config.Description, &panel.Config.HideTitle, &panel.Config.HideBorder)
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

	// Type-change recovery: the plan dropped this config block but prior still has it.
	// Rebuild entirely from the API and skip null-preservation, since there is no
	// current-plan null intent to honor.
	if pm.AiopsPatternAnalysisConfig == nil && prior.AiopsPatternAnalysisConfig != nil {
		pm.AiopsPatternAnalysisConfig = aiopsPatternAnalysisConfigFromAPIImport(api)
		return nil
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
	existing.RandomSamplerProbability = panelkit.PreserveFloat32(existing.RandomSamplerProbability, patternAnalysisRandomSamplerProbabilityValue(api.RandomSamplerProbability))

	panelkit.ApplyPresentationFromAPI(&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder,
		api.Title, api.Description, api.HideTitle, api.HideBorder)

	var priorTR *models.TimeRangeModel
	if prior.AiopsPatternAnalysisConfig != nil {
		priorTR = prior.AiopsPatternAnalysisConfig.TimeRange
	}
	existing.TimeRange = panelkit.MergeTimeRange(existing.TimeRange, api.TimeRange, priorTR)

	if prior.AiopsPatternAnalysisConfig != nil {
		aiopsPatternAnalysisPreserveNullIntentFromPrior(prior.AiopsPatternAnalysisConfig, existing)
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
	cfg.RandomSamplerProbability = types.Float32PointerValue(patternAnalysisRandomSamplerProbabilityValue(api.RandomSamplerProbability))
	cfg.TimeRange = panelkit.TimeRangeFromAPI(api.TimeRange, nil)
	return cfg
}

func aiopsPatternAnalysisPreserveNullIntentFromPrior(prior, existing *models.AiopsPatternAnalysisConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	panelkit.NullPreserveStringFromPrior(prior.MinimumTimeRange, &existing.MinimumTimeRange)
	panelkit.NullPreserveStringFromPrior(prior.RandomSamplerMode, &existing.RandomSamplerMode)
	panelkit.NullPreserveFloat32FromPrior(prior.RandomSamplerProbability, &existing.RandomSamplerProbability)
	panelkit.NullPreservePresentationFromPrior(prior.Title, prior.Description, prior.HideTitle, prior.HideBorder,
		&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder)
	if prior.TimeRange == nil {
		existing.TimeRange = nil
	}
}

func patternAnalysisMinimumTimeRangeAPIValue(value string) *kbapi.KibanaHTTPAPIsAiopsPatternAnalysis_MinimumTimeRange {
	v := &kbapi.KibanaHTTPAPIsAiopsPatternAnalysis_MinimumTimeRange{}
	switch value {
	case "no_minimum":
		_ = v.FromKibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange0(kbapi.NoMinimum)
	case "1_week":
		_ = v.FromKibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange1(kbapi.N1Week)
	case "1_month":
		_ = v.FromKibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange2(kbapi.N1Month)
	case "3_months":
		_ = v.FromKibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange3(kbapi.N3Months)
	case "6_months":
		_ = v.FromKibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange4(kbapi.N6Months)
	default:
		return nil
	}
	return v
}

func patternAnalysisMinimumTimeRangeValue(v *kbapi.KibanaHTTPAPIsAiopsPatternAnalysis_MinimumTimeRange) types.String {
	if v == nil {
		return types.StringNull()
	}
	value, err := v.AsKibanaHTTPAPIsAiopsPatternAnalysisMinimumTimeRange0()
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(value))
}

func patternAnalysisRandomSamplerModeAPIValue(value string) *kbapi.KibanaHTTPAPIsAiopsPatternAnalysis_RandomSamplerMode {
	v := &kbapi.KibanaHTTPAPIsAiopsPatternAnalysis_RandomSamplerMode{}
	switch value {
	case "on_automatic":
		_ = v.FromKibanaHTTPAPIsAiopsPatternAnalysisRandomSamplerMode0(kbapi.OnAutomatic)
	case "on_manual":
		_ = v.FromKibanaHTTPAPIsAiopsPatternAnalysisRandomSamplerMode1(kbapi.OnManual)
	case "off":
		_ = v.FromKibanaHTTPAPIsAiopsPatternAnalysisRandomSamplerMode2(kbapi.KibanaHTTPAPIsAiopsPatternAnalysisRandomSamplerMode2Off)
	default:
		return nil
	}
	return v
}

func patternAnalysisRandomSamplerModeValue(v *kbapi.KibanaHTTPAPIsAiopsPatternAnalysis_RandomSamplerMode) types.String {
	if v == nil {
		return types.StringNull()
	}
	value, err := v.AsKibanaHTTPAPIsAiopsPatternAnalysisRandomSamplerMode0()
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(value))
}

func patternAnalysisRandomSamplerProbabilityValue(v *kbapi.KibanaHTTPAPIsAiopsPatternAnalysis_RandomSamplerProbability) *float32 {
	if v == nil {
		return nil
	}
	value, err := v.AsKibanaHTTPAPIsAiopsPatternAnalysisRandomSamplerProbability0()
	if err != nil {
		return nil
	}
	return &value
}
