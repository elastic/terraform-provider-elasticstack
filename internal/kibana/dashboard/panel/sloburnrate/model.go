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

	var diags diag.Diagnostics
	if len(cfg.Drilldowns) > 0 {
		diags.Append(panelkit.InjectDrilldownsJSON(&embeddable, cfg.Drilldowns)...)
	}

	panel.Config = embeddable
	return diags
}

// PopulateFromAPI maps Kibana SLO burn rate embeddable config into Terraform panel state while
// preserving prior null intent (REQ-009). prior is the prior TF state/plan panel, or nil on import.
//
// pm always arrives with SloBurnRateConfig unset (callers build state from a zero-valued
// PanelModel to avoid aliasing plan pointers), so that field cannot be used to detect whether this
// panel was previously this same type. prior.SloBurnRateConfig is the only reliable signal:
// non-nil means the panel was already this type and its null intent must be honored; nil means
// there is no prior null intent for this config block (creation, import, or a type change).
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsSloBurnRateEmbeddable) diag.Diagnostics {
	if prior == nil || prior.SloBurnRateConfig == nil {
		pm.SloBurnRateConfig = sloBurnRateConfigFromAPIImport(apiConfig)
		return nil
	}

	// Same-type update: rebuild from the API, then reapply the prior config's null intent for any
	// optional field the plan/state had not set (REQ-009 null-preservation).
	existing := sloBurnRateConfigFromAPIImport(apiConfig)
	existing.Drilldowns = readSloBurnRateDrilldownsFromAPI(apiConfig.Drilldowns, prior.SloBurnRateConfig.Drilldowns)
	// sloBurnRateConfigFromAPIImport always normalizes the API's "*" (all-instances) sentinel to
	// null, which is the right default when there is no prior value to consult (creation/import).
	// On an update, a practitioner who explicitly set slo_instance_id = "*" must see it round-trip
	// as "*" rather than be silently nulled, so recompute it from the raw API value using prior
	// knowledge instead of the unconditionally-normalized import value.
	existing.SloInstanceID = sloBurnRateInstanceIDFromAPI(apiConfig.SloInstanceId, prior.SloBurnRateConfig.SloInstanceID)
	sloBurnRatePreserveNullIntentFromPrior(prior.SloBurnRateConfig, existing)
	pm.SloBurnRateConfig = existing
	return nil
}

// sloBurnRateInstanceIDFromAPI computes slo_instance_id for a same-type update: if the prior state
// never had a known value (practitioner omitted it), the field stays null regardless of what the
// API echoes back (the API's "*" all-instances sentinel has no meaningful null-free TF value).
// Otherwise the practitioner is explicitly managing this field, so it round-trips the raw API value
// (including "*") without the import path's blanket normalization.
func sloBurnRateInstanceIDFromAPI(api *string, prior types.String) types.String {
	if !typeutils.IsKnown(prior) {
		return types.StringNull()
	}
	return types.StringPointerValue(api)
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
	panelkit.NullPreserveStringFromPrior(prior.SloInstanceID, &existing.SloInstanceID)
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
	items := make([]panelkit.URLDrilldownAPIItemData, len(*apiDrilldowns))
	for i, d := range *apiDrilldowns {
		items[i] = panelkit.URLDrilldownAPIItemData{
			URL:          d.Url,
			Label:        d.Label,
			EncodeUrl:    d.EncodeUrl,
			OpenInNewTab: d.OpenInNewTab,
		}
	}
	return panelkit.ReadURLDrilldownsFromAPI(items, priorDrilldowns)
}
