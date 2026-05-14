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

package sloalerts

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	drilldownURLEncodeURLDefault    = true
	drilldownURLOpenInNewTabDefault = false
)

func drilldownBoolImportPreserving(api *bool, serverDefault bool) types.Bool {
	if api == nil {
		return types.BoolNull()
	}
	if *api == serverDefault {
		return types.BoolNull()
	}
	return types.BoolValue(*api)
}

// BuildConfig fills panel.Config from Terraform state.
func BuildConfig(pm *models.PanelModel, panel *kbapi.KbnDashboardPanelTypeSloAlerts) {
	cfg := pm.SloAlertsConfig
	if cfg == nil {
		return
	}

	embeddable := kbapi.SloAlertsEmbeddable{}

	slos := make([]struct {
		SloId         string  `json:"slo_id"`                    //nolint:revive
		SloInstanceId *string `json:"slo_instance_id,omitempty"` //nolint:revive
	}, len(cfg.Slos))
	for i, s := range cfg.Slos {
		slos[i].SloId = s.SloID.ValueString()
		if typeutils.IsKnown(s.SloInstanceID) {
			slos[i].SloInstanceId = s.SloInstanceID.ValueStringPointer()
		}
	}
	embeddable.Slos = &slos

	if typeutils.IsKnown(cfg.Title) {
		embeddable.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Description) {
		embeddable.Description = cfg.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.HideTitle) {
		embeddable.HideTitle = cfg.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.HideBorder) {
		embeddable.HideBorder = cfg.HideBorder.ValueBoolPointer()
	}

	if len(cfg.Drilldowns) > 0 {
		drilldowns := make([]struct {
			EncodeUrl    *bool                                      `json:"encode_url,omitempty"` //nolint:revive
			Label        string                                     `json:"label"`
			OpenInNewTab *bool                                      `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.SloAlertsEmbeddableDrilldownsTrigger `json:"trigger"`
			Type         kbapi.SloAlertsEmbeddableDrilldownsType    `json:"type"`
			Url          string                                     `json:"url"` //nolint:revive
		}, len(cfg.Drilldowns))
		for i, d := range cfg.Drilldowns {
			drilldowns[i].Url = d.URL.ValueString()
			drilldowns[i].Label = d.Label.ValueString()
			drilldowns[i].Trigger = kbapi.SloAlertsEmbeddableDrilldownsTriggerOnOpenPanelMenu
			drilldowns[i].Type = kbapi.SloAlertsEmbeddableDrilldownsTypeUrlDrilldown
			if typeutils.IsKnown(d.EncodeURL) {
				drilldowns[i].EncodeUrl = d.EncodeURL.ValueBoolPointer()
			}
			if typeutils.IsKnown(d.OpenInNewTab) {
				drilldowns[i].OpenInNewTab = d.OpenInNewTab.ValueBoolPointer()
			}
		}
		embeddable.Drilldowns = &drilldowns
	}

	panel.Config = embeddable
}

// PopulateFromAPI merges API config into practitioner state seeded from tfPanel.
func PopulateFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, apiPanel kbapi.KbnDashboardPanelTypeSloAlerts) {
	apiCfg := apiPanel.Config

	if tfPanel == nil {
		pm.SloAlertsConfig = sloAlertsPanelConfigFromAPIImport(apiCfg)
		return
	}

	existing := pm.SloAlertsConfig
	if existing == nil {
		return
	}

	if apiCfg.Slos == nil || len(*apiCfg.Slos) == 0 {
		return
	}

	existing.Slos = readSlosFromAPI(*apiCfg.Slos, existing.Slos)

	if typeutils.IsKnown(existing.Title) {
		existing.Title = types.StringPointerValue(apiCfg.Title)
	}
	if typeutils.IsKnown(existing.Description) {
		existing.Description = types.StringPointerValue(apiCfg.Description)
	}
	if typeutils.IsKnown(existing.HideTitle) {
		existing.HideTitle = types.BoolPointerValue(apiCfg.HideTitle)
	}
	if typeutils.IsKnown(existing.HideBorder) {
		existing.HideBorder = types.BoolPointerValue(apiCfg.HideBorder)
	}

	existing.Drilldowns = readDrilldownsFromAPI(apiCfg.Drilldowns, existing.Drilldowns)
}

func sloAlertsPanelConfigFromAPIImport(apiCfg kbapi.SloAlertsEmbeddable) *models.SloAlertsPanelConfigModel {
	cfg := &models.SloAlertsPanelConfigModel{
		Title:       types.StringPointerValue(apiCfg.Title),
		Description: types.StringPointerValue(apiCfg.Description),
		HideTitle:   types.BoolPointerValue(apiCfg.HideTitle),
		HideBorder:  types.BoolPointerValue(apiCfg.HideBorder),
	}
	if apiCfg.Slos != nil {
		cfg.Slos = readSlosFromAPI(*apiCfg.Slos, nil)
	}
	cfg.Drilldowns = readDrilldownsFromAPI(apiCfg.Drilldowns, nil)
	return cfg
}

func readSlosFromAPI(
	apiSlos []struct {
		SloId         string  `json:"slo_id"`                    //nolint:revive
		SloInstanceId *string `json:"slo_instance_id,omitempty"` //nolint:revive
	},
	priorSlos []models.SloAlertsPanelSloModel,
) []models.SloAlertsPanelSloModel {
	out := make([]models.SloAlertsPanelSloModel, len(apiSlos))
	for i, apiSlo := range apiSlos {
		out[i].SloID = types.StringValue(apiSlo.SloId)

		var prior *models.SloAlertsPanelSloModel
		if i < len(priorSlos) {
			prior = &priorSlos[i]
		}

		switch {
		case prior == nil:
			if apiSlo.SloInstanceId != nil && *apiSlo.SloInstanceId != "*" {
				out[i].SloInstanceID = types.StringValue(*apiSlo.SloInstanceId)
			} else {
				out[i].SloInstanceID = types.StringNull()
			}
		case typeutils.IsKnown(prior.SloInstanceID):
			out[i].SloInstanceID = types.StringPointerValue(apiSlo.SloInstanceId)
		default:
			out[i].SloInstanceID = types.StringNull()
		}
	}
	return out
}

func readDrilldownsFromAPI(
	apiDrilldowns *[]struct {
		EncodeUrl    *bool                                      `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                     `json:"label"`
		OpenInNewTab *bool                                      `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.SloAlertsEmbeddableDrilldownsTrigger `json:"trigger"`
		Type         kbapi.SloAlertsEmbeddableDrilldownsType    `json:"type"`
		Url          string                                     `json:"url"` //nolint:revive
	},
	priorDrilldowns []models.URLDrilldownModel,
) []models.URLDrilldownModel {
	if apiDrilldowns == nil || len(*apiDrilldowns) == 0 {
		return nil
	}

	out := make([]models.URLDrilldownModel, len(*apiDrilldowns))
	for i, d := range *apiDrilldowns {
		out[i].URL = types.StringValue(d.Url)
		out[i].Label = types.StringValue(d.Label)

		var prior *models.URLDrilldownModel
		if i < len(priorDrilldowns) {
			prior = &priorDrilldowns[i]
		}

		if prior == nil {
			out[i].EncodeURL = drilldownBoolImportPreserving(d.EncodeUrl, drilldownURLEncodeURLDefault)
			out[i].OpenInNewTab = drilldownBoolImportPreserving(d.OpenInNewTab, drilldownURLOpenInNewTabDefault)
			continue
		}

		switch {
		case prior.EncodeURL.IsNull():
			out[i].EncodeURL = types.BoolNull()
		case d.EncodeUrl != nil:
			out[i].EncodeURL = types.BoolValue(*d.EncodeUrl)
		default:
			out[i].EncodeURL = types.BoolNull()
		}

		switch {
		case prior.OpenInNewTab.IsNull():
			out[i].OpenInNewTab = types.BoolNull()
		case d.OpenInNewTab != nil:
			out[i].OpenInNewTab = types.BoolValue(*d.OpenInNewTab)
		default:
			out[i].OpenInNewTab = types.BoolNull()
		}
	}
	return out
}
