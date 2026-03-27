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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// sloBurnRateConfigModel is the Terraform model for the slo_burn_rate_config block.
type sloBurnRateConfigModel struct {
	SloID          types.String             `tfsdk:"slo_id"`
	Duration       types.String             `tfsdk:"duration"`
	SloInstanceID  types.String             `tfsdk:"slo_instance_id"`
	Title          types.String             `tfsdk:"title"`
	Description    types.String             `tfsdk:"description"`
	HideTitle      types.Bool               `tfsdk:"hide_title"`
	HideBorder     types.Bool               `tfsdk:"hide_border"`
	Drilldowns     []sloBurnRateDrilldownModel `tfsdk:"drilldowns"`
}

// sloBurnRateDrilldownModel represents a single drilldown entry within slo_burn_rate_config.
type sloBurnRateDrilldownModel struct {
	URL           types.String `tfsdk:"url"`
	Label         types.String `tfsdk:"label"`
	Trigger       types.String `tfsdk:"trigger"`
	Type          types.String `tfsdk:"type"`
	EncodeURL     types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab  types.Bool   `tfsdk:"open_in_new_tab"`
}

// buildSloBurnRateConfig writes the TF model fields into the API panel struct.
func buildSloBurnRateConfig(pm panelModel, panel *kbapi.KbnDashboardPanelSloBurnRate) {
	cfg := pm.SloBurnRateConfig
	if cfg == nil {
		return
	}

	embeddable := kbapi.SloBurnRateEmbeddable{
		SloId:    cfg.SloID.ValueString(),
		Duration: cfg.Duration.ValueString(),
	}

	if typeutils.IsKnown(cfg.SloInstanceID) {
		embeddable.SloInstanceId = cfg.SloInstanceID.ValueStringPointer()
	}
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
			EncodeUrl    *bool                                  `json:"encode_url,omitempty"`
			Label        string                                 `json:"label"`
			OpenInNewTab *bool                                  `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.SloBurnRateEmbeddableDrilldownsTrigger `json:"trigger"`
			Type         kbapi.SloBurnRateEmbeddableDrilldownsType    `json:"type"`
			Url          string                                 `json:"url"`
		}, len(cfg.Drilldowns))

		for i, d := range cfg.Drilldowns {
			drilldowns[i].Url = d.URL.ValueString()
			drilldowns[i].Label = d.Label.ValueString()
			drilldowns[i].Trigger = kbapi.SloBurnRateEmbeddableDrilldownsTrigger(d.Trigger.ValueString())
			drilldowns[i].Type = kbapi.SloBurnRateEmbeddableDrilldownsType(d.Type.ValueString())
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

// populateSloBurnRateFromAPI reads back a SLO burn rate config from the API response and
// updates the panel model. Null-preservation semantics apply: if slo_instance_id was null
// in the prior state and the API returns "*", keep it null. If there is no existing config
// block in TF state and tfPanel is non-nil, preserve that nil intent.
//
// tfPanel is the prior TF state/plan panel, or nil on import. When nil, all API-returned
// fields are populated unconditionally (no prior intent to preserve).
func populateSloBurnRateFromAPI(pm *panelModel, tfPanel *panelModel, apiConfig kbapi.SloBurnRateEmbeddable) {
	// On import (tfPanel == nil) populate from API unconditionally.
	if tfPanel == nil {
		cfg := &sloBurnRateConfigModel{
			SloID:    types.StringValue(apiConfig.SloId),
			Duration: types.StringValue(apiConfig.Duration),
		}
		if apiConfig.SloInstanceId != nil {
			cfg.SloInstanceID = types.StringValue(*apiConfig.SloInstanceId)
		} else {
			cfg.SloInstanceID = types.StringNull()
		}
		if apiConfig.Title != nil {
			cfg.Title = types.StringValue(*apiConfig.Title)
		} else {
			cfg.Title = types.StringNull()
		}
		if apiConfig.Description != nil {
			cfg.Description = types.StringValue(*apiConfig.Description)
		} else {
			cfg.Description = types.StringNull()
		}
		if apiConfig.HideTitle != nil {
			cfg.HideTitle = types.BoolValue(*apiConfig.HideTitle)
		} else {
			cfg.HideTitle = types.BoolNull()
		}
		if apiConfig.HideBorder != nil {
			cfg.HideBorder = types.BoolValue(*apiConfig.HideBorder)
		} else {
			cfg.HideBorder = types.BoolNull()
		}
		cfg.Drilldowns = readSloBurnRateDrilldownsFromAPI(apiConfig.Drilldowns, nil)
		pm.SloBurnRateConfig = cfg
		return
	}

	existing := pm.SloBurnRateConfig

	// If there was no config block in prior state, preserve nil intent.
	if existing == nil {
		return
	}

	// Block exists in state — update required fields always, optional fields using null-preservation.
	existing.SloID = types.StringValue(apiConfig.SloId)
	existing.Duration = types.StringValue(apiConfig.Duration)

	// slo_instance_id null-preservation: if state is null and API returns "*", keep null.
	if typeutils.IsKnown(existing.SloInstanceID) {
		// Practitioner explicitly set a value — round-trip normally.
		if apiConfig.SloInstanceId != nil {
			existing.SloInstanceID = types.StringValue(*apiConfig.SloInstanceId)
		} else {
			existing.SloInstanceID = types.StringNull()
		}
	} else if existing.SloInstanceID.IsNull() {
		// Practitioner did not configure slo_instance_id — preserve null regardless of API response.
		// (If API returns "*", we do NOT populate it.)
	}

	// Optional string fields: only update from API when they were already known in state.
	if typeutils.IsKnown(existing.Title) {
		if apiConfig.Title != nil {
			existing.Title = types.StringValue(*apiConfig.Title)
		} else {
			existing.Title = types.StringNull()
		}
	}
	if typeutils.IsKnown(existing.Description) {
		if apiConfig.Description != nil {
			existing.Description = types.StringValue(*apiConfig.Description)
		} else {
			existing.Description = types.StringNull()
		}
	}

	// Optional bool fields: only update from API when they were already known in state.
	if typeutils.IsKnown(existing.HideTitle) {
		if apiConfig.HideTitle != nil {
			existing.HideTitle = types.BoolValue(*apiConfig.HideTitle)
		} else {
			existing.HideTitle = types.BoolNull()
		}
	}
	if typeutils.IsKnown(existing.HideBorder) {
		if apiConfig.HideBorder != nil {
			existing.HideBorder = types.BoolValue(*apiConfig.HideBorder)
		} else {
			existing.HideBorder = types.BoolNull()
		}
	}

	// Drilldowns: update from API preserving optional bool null-preservation per drilldown.
	existing.Drilldowns = readSloBurnRateDrilldownsFromAPI(apiConfig.Drilldowns, existing.Drilldowns)
}

// readSloBurnRateDrilldownsFromAPI converts the API drilldowns slice into TF models.
// priorDrilldowns is the existing TF state slice (may be nil). When present, optional
// bool fields (encode_url, open_in_new_tab) use null-preservation: if the prior state
// value was null, keep null even if the API returns a value.
func readSloBurnRateDrilldownsFromAPI(
	apiDrilldowns *[]struct {
		EncodeUrl    *bool                                  `json:"encode_url,omitempty"`
		Label        string                                 `json:"label"`
		OpenInNewTab *bool                                  `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.SloBurnRateEmbeddableDrilldownsTrigger `json:"trigger"`
		Type         kbapi.SloBurnRateEmbeddableDrilldownsType    `json:"type"`
		Url          string                                 `json:"url"`
	},
	priorDrilldowns []sloBurnRateDrilldownModel,
) []sloBurnRateDrilldownModel {
	if apiDrilldowns == nil || len(*apiDrilldowns) == 0 {
		return nil
	}

	result := make([]sloBurnRateDrilldownModel, len(*apiDrilldowns))
	for i, d := range *apiDrilldowns {
		result[i] = sloBurnRateDrilldownModel{
			URL:     types.StringValue(d.Url),
			Label:   types.StringValue(d.Label),
			Trigger: types.StringValue(string(d.Trigger)),
			Type:    types.StringValue(string(d.Type)),
		}

		// Determine prior state for this drilldown (if it exists at this index).
		var prior *sloBurnRateDrilldownModel
		if i < len(priorDrilldowns) {
			prior = &priorDrilldowns[i]
		}

		// encode_url: null-preserve if prior was null, otherwise populate from API.
		if prior != nil && prior.EncodeURL.IsNull() {
			result[i].EncodeURL = types.BoolNull()
		} else if d.EncodeUrl != nil {
			result[i].EncodeURL = types.BoolValue(*d.EncodeUrl)
		} else {
			result[i].EncodeURL = types.BoolNull()
		}

		// open_in_new_tab: null-preserve if prior was null, otherwise populate from API.
		if prior != nil && prior.OpenInNewTab.IsNull() {
			result[i].OpenInNewTab = types.BoolNull()
		} else if d.OpenInNewTab != nil {
			result[i].OpenInNewTab = types.BoolValue(*d.OpenInNewTab)
		} else {
			result[i].OpenInNewTab = types.BoolNull()
		}
	}

	return result
}
