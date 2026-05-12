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

package dashboard

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// sloAlertsPanelConfigModel is the Terraform model for `slo_alerts_config`.
type sloAlertsPanelConfigModel struct {
	Slos        []sloAlertsPanelSloModel       `tfsdk:"slos"`
	Title       types.String                   `tfsdk:"title"`
	Description types.String                   `tfsdk:"description"`
	HideTitle   types.Bool                     `tfsdk:"hide_title"`
	HideBorder  types.Bool                     `tfsdk:"hide_border"`
	Drilldowns  []sloAlertsPanelDrilldownModel `tfsdk:"drilldowns"`
}

type sloAlertsPanelSloModel struct {
	SloID         types.String `tfsdk:"slo_id"`
	SloInstanceID types.String `tfsdk:"slo_instance_id"`
}

type sloAlertsPanelDrilldownModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

func sloAlertsPanelToAPI(pm panelModel, grid struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
}, panelID *string) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := pm.SloAlertsConfig
	if cfg == nil {
		diags.AddError("Missing SLO alerts panel configuration", "SLO alerts panels require `slo_alerts_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}

	out := kbapi.KbnDashboardPanelTypeSloAlerts{
		Grid: grid,
		Id:   panelID,
		Type: kbapi.SloAlerts,
	}

	embeddable := kbapi.SloAlertsEmbeddable{}

	slos := make([]struct {
		SloId         string  `json:"slo_id"`                    //nolint:revive // kbapi JSON shape
		SloInstanceId *string `json:"slo_instance_id,omitempty"` //nolint:revive // kbapi JSON shape
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

	out.Config = embeddable

	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeSloAlerts(out); err != nil {
		diags.AddError("Failed to create SLO alerts panel", err.Error())
	}
	return panelItem, diags
}

// populateSloAlertsPanelFromAPI maps an API `slo_alerts` panel into Terraform state.
func populateSloAlertsPanelFromAPI(pm *panelModel, tfPanel *panelModel, apiPanel kbapi.KbnDashboardPanelTypeSloAlerts) {
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

	existing.Slos = readSloAlertsSlosFromAPI(*apiCfg.Slos, existing.Slos)

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

	existing.Drilldowns = readSloAlertsDrilldownsFromAPI(apiCfg.Drilldowns, existing.Drilldowns)
}

func sloAlertsPanelConfigFromAPIImport(apiCfg kbapi.SloAlertsEmbeddable) *sloAlertsPanelConfigModel {
	cfg := &sloAlertsPanelConfigModel{
		Title:       types.StringPointerValue(apiCfg.Title),
		Description: types.StringPointerValue(apiCfg.Description),
		HideTitle:   types.BoolPointerValue(apiCfg.HideTitle),
		HideBorder:  types.BoolPointerValue(apiCfg.HideBorder),
	}
	if apiCfg.Slos != nil {
		cfg.Slos = readSloAlertsSlosFromAPI(*apiCfg.Slos, nil)
	}
	cfg.Drilldowns = readSloAlertsDrilldownsFromAPI(apiCfg.Drilldowns, nil)
	return cfg
}

func readSloAlertsSlosFromAPI(
	apiSlos []struct {
		SloId         string  `json:"slo_id"`                    //nolint:revive // kbapi JSON shape
		SloInstanceId *string `json:"slo_instance_id,omitempty"` //nolint:revive // kbapi JSON shape
	},
	priorSlos []sloAlertsPanelSloModel,
) []sloAlertsPanelSloModel {
	out := make([]sloAlertsPanelSloModel, len(apiSlos))
	for i, apiSlo := range apiSlos {
		out[i].SloID = types.StringValue(apiSlo.SloId)

		var prior *sloAlertsPanelSloModel
		if i < len(priorSlos) {
			prior = &priorSlos[i]
		}

		switch {
		case prior == nil:
			// Import path: treat "*" (all instances) like omitted.
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

func readSloAlertsDrilldownsFromAPI(
	apiDrilldowns *[]struct {
		EncodeUrl    *bool                                      `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                     `json:"label"`
		OpenInNewTab *bool                                      `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.SloAlertsEmbeddableDrilldownsTrigger `json:"trigger"`
		Type         kbapi.SloAlertsEmbeddableDrilldownsType    `json:"type"`
		Url          string                                     `json:"url"` //nolint:revive
	},
	priorDrilldowns []sloAlertsPanelDrilldownModel,
) []sloAlertsPanelDrilldownModel {
	if apiDrilldowns == nil || len(*apiDrilldowns) == 0 {
		return nil
	}

	out := make([]sloAlertsPanelDrilldownModel, len(*apiDrilldowns))
	for i, d := range *apiDrilldowns {
		out[i].URL = types.StringValue(d.Url)
		out[i].Label = types.StringValue(d.Label)

		var prior *sloAlertsPanelDrilldownModel
		if i < len(priorDrilldowns) {
			prior = &priorDrilldowns[i]
		}

		if prior == nil {
			out[i].EncodeURL = panelDrilldownBoolImportPreserving(d.EncodeUrl, drilldownURLEncodeURLDefault)
			out[i].OpenInNewTab = panelDrilldownBoolImportPreserving(d.OpenInNewTab, drilldownURLOpenInNewTabDefault)
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
