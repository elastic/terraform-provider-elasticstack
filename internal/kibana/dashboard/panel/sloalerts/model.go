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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	drilldownURLEncodeURLDefault    = true
	drilldownURLOpenInNewTabDefault = false
)

// BuildConfig fills panel.Config from Terraform state.
func BuildConfig(pm *models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts) {
	cfg := pm.SloAlertsConfig
	if cfg == nil {
		return
	}

	embeddable := kbapi.KibanaHTTPAPIsSloAlertsEmbeddable{}

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

	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&embeddable.Title, &embeddable.Description, &embeddable.HideTitle, &embeddable.HideBorder)

	if len(cfg.Drilldowns) > 0 {
		drilldowns := make([]struct {
			EncodeUrl    *bool                                                    `json:"encode_url,omitempty"` //nolint:revive
			Label        string                                                   `json:"label"`
			OpenInNewTab *bool                                                    `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.KibanaHTTPAPIsSloAlertsEmbeddableDrilldownsTrigger `json:"trigger"`
			Type         kbapi.KibanaHTTPAPIsSloAlertsEmbeddableDrilldownsType    `json:"type"`
			Url          string                                                   `json:"url"` //nolint:revive
		}, len(cfg.Drilldowns))
		for i, d := range cfg.Drilldowns {
			drilldowns[i].Url = d.URL.ValueString()
			drilldowns[i].Label = d.Label.ValueString()
			drilldowns[i].Trigger = kbapi.KibanaHTTPAPIsSloAlertsEmbeddableDrilldownsTriggerOnOpenPanelMenu
			drilldowns[i].Type = kbapi.KibanaHTTPAPIsSloAlertsEmbeddableDrilldownsTypeUrlDrilldown
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
func PopulateFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts) {
	apiCfg := apiPanel.Config

	if pm.SloAlertsConfig == nil {
		pm.SloAlertsConfig = sloAlertsPanelConfigFromAPIImport(apiCfg)
	}

	if tfPanel == nil {
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

	panelkit.ApplyPresentationFromAPI(&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder,
		apiCfg.Title, apiCfg.Description, apiCfg.HideTitle, apiCfg.HideBorder)

	var priorDrilldowns []models.URLDrilldownModel
	if tfPanel.SloAlertsConfig != nil {
		priorDrilldowns = tfPanel.SloAlertsConfig.Drilldowns
	}
	existing.Drilldowns = sloAlertsAPIEntries(apiCfg.Drilldowns, priorDrilldowns)
}

func sloAlertsPanelConfigFromAPIImport(apiCfg kbapi.KibanaHTTPAPIsSloAlertsEmbeddable) *models.SloAlertsPanelConfigModel {
	cfg := &models.SloAlertsPanelConfigModel{
		Title:       types.StringPointerValue(apiCfg.Title),
		Description: types.StringPointerValue(apiCfg.Description),
		HideTitle:   types.BoolPointerValue(apiCfg.HideTitle),
		HideBorder:  types.BoolPointerValue(apiCfg.HideBorder),
	}
	if apiCfg.Slos != nil {
		cfg.Slos = readSlosFromAPI(*apiCfg.Slos, nil)
	}
	cfg.Drilldowns = sloAlertsAPIEntries(apiCfg.Drilldowns, nil)
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

func sloAlertsAPIEntries(
	apiDrilldowns *[]struct {
		EncodeUrl    *bool                                                    `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                   `json:"label"`
		OpenInNewTab *bool                                                    `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KibanaHTTPAPIsSloAlertsEmbeddableDrilldownsTrigger `json:"trigger"`
		Type         kbapi.KibanaHTTPAPIsSloAlertsEmbeddableDrilldownsType    `json:"type"`
		Url          string                                                   `json:"url"` //nolint:revive
	},
	priorDrilldowns []models.URLDrilldownModel,
) []models.URLDrilldownModel {
	if apiDrilldowns == nil || len(*apiDrilldowns) == 0 {
		return nil
	}
	entries := make([]panelkit.URLDrilldownAPIEntry, len(*apiDrilldowns))
	for i, d := range *apiDrilldowns {
		entries[i] = panelkit.URLDrilldownAPIEntry{
			URL:          d.Url,
			Label:        d.Label,
			EncodeURL:    d.EncodeUrl,
			OpenInNewTab: d.OpenInNewTab,
		}
	}
	return panelkit.ReadURLDrilldownsFromAPI(entries, priorDrilldowns, drilldownURLEncodeURLDefault, drilldownURLOpenInNewTabDefault)
}
