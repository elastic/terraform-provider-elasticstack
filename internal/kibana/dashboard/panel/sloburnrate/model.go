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

package sloburnrate

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes Terraform state from pm into panel's typed API config.
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloBurnRate) diag.Diagnostics {
	cfg := pm.SloBurnRateConfig
	if cfg == nil {
		var diags diag.Diagnostics
		diags.AddError(
			"Missing SLO burn rate panel configuration",
			"SLO burn rate panels require `slo_burn_rate_config`.",
		)
		return diags
	}

	embeddable := kbapi.KibanaHTTPAPIsSloBurnRateEmbeddable{
		SloId:    cfg.SloID.ValueString(),
		Duration: cfg.Duration.ValueString(),
	}

	if typeutils.IsKnown(cfg.SloInstanceID) {
		embeddable.SloInstanceId = cfg.SloInstanceID.ValueStringPointer()
	}
	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&embeddable.Title, &embeddable.Description, &embeddable.HideTitle, &embeddable.HideBorder)

	if len(cfg.Drilldowns) > 0 {
		panelkit.InjectDrilldownsJSON(&embeddable, cfg.Drilldowns)
	}

	panel.Config = embeddable
	return nil
}

// PopulateFromAPI maps Kibana SLO burn rate embeddable config into Terraform panel state while preserving prior null intent.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsSloBurnRateEmbeddable) diag.Diagnostics {
	// On import (prior == nil) populate from API unconditionally.
	if prior == nil {
		pm.SloBurnRateConfig = sloBurnRateConfigFromAPIImport(apiConfig)
		return nil
	}

	if pm.SloBurnRateConfig == nil && prior.SloBurnRateConfig != nil {
		pm.SloBurnRateConfig = sloBurnRateConfigFromAPIImport(apiConfig)
	}

	existing := pm.SloBurnRateConfig
	if existing == nil {
		return nil
	}

	// Block exists in state — update required fields always, optional fields using null-preservation.
	existing.SloID = types.StringValue(apiConfig.SloId)
	existing.Duration = types.StringValue(apiConfig.Duration)

	// slo_instance_id null-preservation: if state is null (practitioner omitted it), keep null
	// regardless of what the API returns — the API echoes "*" for all-instances which has no
	// meaningful TF representation.
	existing.SloInstanceID = panelkit.PreserveString(existing.SloInstanceID, apiConfig.SloInstanceId)

	// Optional fields: only update from API when they were already known in state.
	panelkit.ApplyPresentationFromAPI(&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder,
		apiConfig.Title, apiConfig.Description, apiConfig.HideTitle, apiConfig.HideBorder)

	var priorDrilldowns []models.URLDrilldownModel
	if prior != nil && prior.SloBurnRateConfig != nil {
		priorDrilldowns = prior.SloBurnRateConfig.Drilldowns
	}
	existing.Drilldowns = readSloBurnRateDrilldownsFromAPI(apiConfig.Drilldowns, priorDrilldowns)

	if prior != nil && prior.SloBurnRateConfig != nil {
		sloBurnRatePreserveNullIntentFromPrior(prior.SloBurnRateConfig, existing)
	}

	return nil
}

func sloBurnRateConfigFromAPIImport(apiConfig kbapi.KibanaHTTPAPIsSloBurnRateEmbeddable) *models.SloBurnRateConfigModel {
	cfg := &models.SloBurnRateConfigModel{
		SloID:    types.StringValue(apiConfig.SloId),
		Duration: types.StringValue(apiConfig.Duration),
	}
	// Normalize "*" (all-instances wildcard) to null, matching create+refresh behaviour.
	if apiConfig.SloInstanceId != nil && *apiConfig.SloInstanceId != "*" {
		cfg.SloInstanceID = types.StringValue(*apiConfig.SloInstanceId)
	} else {
		cfg.SloInstanceID = types.StringNull()
	}
	cfg.Title = types.StringPointerValue(apiConfig.Title)
	cfg.Description = types.StringPointerValue(apiConfig.Description)
	cfg.HideTitle = types.BoolPointerValue(apiConfig.HideTitle)
	cfg.HideBorder = types.BoolPointerValue(apiConfig.HideBorder)
	cfg.Drilldowns = readSloBurnRateDrilldownsFromAPI(apiConfig.Drilldowns, nil)
	return cfg
}

func sloBurnRatePreserveNullIntentFromPrior(prior, existing *models.SloBurnRateConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	if !typeutils.IsKnown(prior.SloInstanceID) {
		existing.SloInstanceID = types.StringNull()
	}
	panelkit.NullPreservePresentationFromPrior(prior.Title, prior.Description, prior.HideTitle, prior.HideBorder,
		&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder)
	if len(prior.Drilldowns) == 0 {
		existing.Drilldowns = nil
	}
}

func readSloBurnRateDrilldownsFromAPI(
	apiDrilldowns *[]struct {
		EncodeUrl    *bool                                                      `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                     `json:"label"`
		OpenInNewTab *bool                                                      `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KibanaHTTPAPIsSloBurnRateEmbeddableDrilldownsTrigger `json:"trigger"`
		Type         kbapi.KibanaHTTPAPIsSloBurnRateEmbeddableDrilldownsType    `json:"type"`
		Url          string                                                     `json:"url"` //nolint:revive
	},
	priorDrilldowns []models.URLDrilldownModel,
) []models.URLDrilldownModel {
	if apiDrilldowns == nil || len(*apiDrilldowns) == 0 {
		return nil
	}

	result := make([]models.URLDrilldownModel, len(*apiDrilldowns))
	for i, d := range *apiDrilldowns {
		result[i] = models.URLDrilldownModel{
			URL:   types.StringValue(d.Url),
			Label: types.StringValue(d.Label),
		}

		// Determine prior state for this drilldown (if it exists at this index).
		var prior *models.URLDrilldownModel
		if i < len(priorDrilldowns) {
			prior = &priorDrilldowns[i]
		}

		// encode_url: null-preserve if prior was null, otherwise populate from API.
		switch {
		case prior != nil && prior.EncodeURL.IsNull():
			result[i].EncodeURL = types.BoolNull()
		case d.EncodeUrl != nil:
			result[i].EncodeURL = types.BoolValue(*d.EncodeUrl)
		default:
			result[i].EncodeURL = types.BoolNull()
		}

		// open_in_new_tab: null-preserve if prior was null, otherwise populate from API.
		switch {
		case prior != nil && prior.OpenInNewTab.IsNull():
			result[i].OpenInNewTab = types.BoolNull()
		case d.OpenInNewTab != nil:
			result[i].OpenInNewTab = types.BoolValue(*d.OpenInNewTab)
		default:
			result[i].OpenInNewTab = types.BoolNull()
		}
	}

	return result
}
