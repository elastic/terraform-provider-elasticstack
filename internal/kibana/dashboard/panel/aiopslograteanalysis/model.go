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

package aiopslograteanalysis

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes Terraform state from pm into the typed API panel config.
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsLogRateAnalysis) diag.Diagnostics {
	cfg := pm.AiopsLogRateAnalysisConfig
	if cfg == nil {
		return nil
	}

	panel.Config.DataViewId = cfg.DataViewID.ValueString()

	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&panel.Config.Title, &panel.Config.Description, &panel.Config.HideTitle, &panel.Config.HideBorder)
	panel.Config.TimeRange = lenscommon.TimeRangeModelToAPI(cfg.TimeRange)

	return nil
}

// PopulateFromAPI maps the Kibana API panel config into Terraform panel state while preserving
// prior null intent (REQ-009). prior is the prior TF state/plan panel, or nil on import.
//
// pm always arrives with AiopsLogRateAnalysisConfig unset (callers build state from a zero-valued
// PanelModel to avoid aliasing plan pointers), so that field cannot be used to detect whether this
// panel was previously this same type. prior.AiopsLogRateAnalysisConfig is the only reliable signal:
// non-nil means the panel was already this type and its null intent must be honored; nil means
// there is no prior null intent for this config block (creation, import, or a type change).
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, api kbapi.KibanaHTTPAPIsAiopsLogRateAnalysis) diag.Diagnostics {
	if prior == nil || prior.AiopsLogRateAnalysisConfig == nil {
		pm.AiopsLogRateAnalysisConfig = aiopsLogRateAnalysisConfigFromAPIImport(api)
		return nil
	}

	// Same-type update: rebuild from the API, then reapply the prior config's null intent for any
	// optional field the plan/state had not set (REQ-009 null-preservation).
	existing := aiopsLogRateAnalysisConfigFromAPIImport(api)
	existing.TimeRange = panelkit.MergeTimeRange(existing.TimeRange, api.TimeRange, prior.AiopsLogRateAnalysisConfig.TimeRange)
	aiopsLogRateAnalysisPreserveNullIntentFromPrior(prior.AiopsLogRateAnalysisConfig, existing)
	pm.AiopsLogRateAnalysisConfig = existing
	return nil
}

func aiopsLogRateAnalysisConfigFromAPIImport(api kbapi.KibanaHTTPAPIsAiopsLogRateAnalysis) *models.AiopsLogRateAnalysisConfigModel {
	cfg := &models.AiopsLogRateAnalysisConfigModel{
		DataViewID:  types.StringValue(api.DataViewId),
		Title:       types.StringPointerValue(api.Title),
		Description: types.StringPointerValue(api.Description),
		HideTitle:   types.BoolPointerValue(api.HideTitle),
		HideBorder:  types.BoolPointerValue(api.HideBorder),
	}
	cfg.TimeRange = panelkit.TimeRangeFromAPI(api.TimeRange, nil)
	return cfg
}

func aiopsLogRateAnalysisPreserveNullIntentFromPrior(prior, existing *models.AiopsLogRateAnalysisConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	panelkit.NullPreservePresentationFromPrior(prior.Title, prior.Description, prior.HideTitle, prior.HideBorder,
		&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder)
	if prior.TimeRange == nil {
		existing.TimeRange = nil
	}
}
