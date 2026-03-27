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

type sloErrorBudgetConfigModel struct {
	SloID          types.String                        `tfsdk:"slo_id"`
	SloInstanceID  types.String                        `tfsdk:"slo_instance_id"`
	Title          types.String                        `tfsdk:"title"`
	Description    types.String                        `tfsdk:"description"`
	HideTitle      types.Bool                          `tfsdk:"hide_title"`
	HideBorder     types.Bool                          `tfsdk:"hide_border"`
	Drilldowns     []sloErrorBudgetDrilldownModel      `tfsdk:"drilldowns"`
}

type sloErrorBudgetDrilldownModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	Type         types.String `tfsdk:"type"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

// buildSloErrorBudgetConfig writes the TF model fields into the API panel struct.
func buildSloErrorBudgetConfig(pm panelModel, sebPanel *kbapi.KbnDashboardPanelSloErrorBudget) {
	cfg := pm.SloErrorBudgetConfig
	if cfg == nil {
		return
	}

	sebPanel.Config.SloId = cfg.SloID.ValueString()

	if typeutils.IsKnown(cfg.SloInstanceID) {
		v := cfg.SloInstanceID.ValueString()
		sebPanel.Config.SloInstanceId = &v
	}
	if typeutils.IsKnown(cfg.Title) {
		v := cfg.Title.ValueString()
		sebPanel.Config.Title = &v
	}
	if typeutils.IsKnown(cfg.Description) {
		v := cfg.Description.ValueString()
		sebPanel.Config.Description = &v
	}
	if typeutils.IsKnown(cfg.HideTitle) {
		v := cfg.HideTitle.ValueBool()
		sebPanel.Config.HideTitle = &v
	}
	if typeutils.IsKnown(cfg.HideBorder) {
		v := cfg.HideBorder.ValueBool()
		sebPanel.Config.HideBorder = &v
	}

	if len(cfg.Drilldowns) > 0 {
		drilldowns := make([]struct {
			EncodeUrl    *bool                                     `json:"encode_url,omitempty"`
			Label        string                                    `json:"label"`
			OpenInNewTab *bool                                     `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.SloErrorBudgetEmbeddableDrilldownsTrigger `json:"trigger"`
			Type         kbapi.SloErrorBudgetEmbeddableDrilldownsType    `json:"type"`
			Url          string                                    `json:"url"`
		}, len(cfg.Drilldowns))

		for i, d := range cfg.Drilldowns {
			drilldowns[i].Url = d.URL.ValueString()
			drilldowns[i].Label = d.Label.ValueString()
			drilldowns[i].Trigger = kbapi.SloErrorBudgetEmbeddableDrilldownsTrigger(d.Trigger.ValueString())
			drilldowns[i].Type = kbapi.SloErrorBudgetEmbeddableDrilldownsType(d.Type.ValueString())
			if typeutils.IsKnown(d.EncodeURL) {
				v := d.EncodeURL.ValueBool()
				drilldowns[i].EncodeUrl = &v
			}
			if typeutils.IsKnown(d.OpenInNewTab) {
				v := d.OpenInNewTab.ValueBool()
				drilldowns[i].OpenInNewTab = &v
			}
		}
		sebPanel.Config.Drilldowns = &drilldowns
	}
}

// populateSloErrorBudgetFromAPI reads back an SLO error budget config from the API
// response and updates the panel model. Null-preservation semantics apply:
//   - slo_instance_id: if null in prior state, keep null even if the API returns "*"
//   - encode_url / open_in_new_tab: normalize API default true to null (don't overwrite
//     null intent with true), so practitioners who omit these fields don't observe drift
//
// tfPanel is the prior TF state/plan panel, or nil on import.
func populateSloErrorBudgetFromAPI(pm *panelModel, tfPanel *panelModel, apiConfig kbapi.SloErrorBudgetEmbeddable) {
	existing := pm.SloErrorBudgetConfig

	// Determine the prior intent for slo_instance_id null-preservation.
	var priorSloInstanceID types.String
	if tfPanel != nil && tfPanel.SloErrorBudgetConfig != nil {
		priorSloInstanceID = tfPanel.SloErrorBudgetConfig.SloInstanceID
	} else if tfPanel == nil {
		// Import: no prior intent — populate all API-returned values.
		priorSloInstanceID = types.StringValue("*") // treat as "known", so we write the API value
	}

	if existing == nil {
		// If tfPanel is nil (import) or no config block in prior state,
		// only create one if the API returned data.
		if tfPanel != nil {
			// Prior state had no block — preserve nil intent.
			return
		}
		// Import path: create block from API.
		pm.SloErrorBudgetConfig = &sloErrorBudgetConfigModel{}
		existing = pm.SloErrorBudgetConfig
	}

	// Always update slo_id from API.
	existing.SloID = types.StringValue(apiConfig.SloId)

	// slo_instance_id: only write if prior intent was non-null (or import).
	if typeutils.IsKnown(priorSloInstanceID) && apiConfig.SloInstanceId != nil {
		existing.SloInstanceID = types.StringValue(*apiConfig.SloInstanceId)
	}

	// title, description, hide_title, hide_border: only update if already known.
	if typeutils.IsKnown(existing.Title) && apiConfig.Title != nil {
		existing.Title = types.StringValue(*apiConfig.Title)
	}
	if typeutils.IsKnown(existing.Description) && apiConfig.Description != nil {
		existing.Description = types.StringValue(*apiConfig.Description)
	}
	if typeutils.IsKnown(existing.HideTitle) && apiConfig.HideTitle != nil {
		existing.HideBorder = types.BoolValue(*apiConfig.HideTitle)
	}
	if typeutils.IsKnown(existing.HideBorder) && apiConfig.HideBorder != nil {
		existing.HideBorder = types.BoolValue(*apiConfig.HideBorder)
	}

	// Drilldowns round-trip with default normalization.
	if apiConfig.Drilldowns != nil {
		// Determine prior drilldown intent.
		var priorDrilldowns []sloErrorBudgetDrilldownModel
		if tfPanel != nil && tfPanel.SloErrorBudgetConfig != nil {
			priorDrilldowns = tfPanel.SloErrorBudgetConfig.Drilldowns
		}

		newDrilldowns := make([]sloErrorBudgetDrilldownModel, 0, len(*apiConfig.Drilldowns))
		for i, d := range *apiConfig.Drilldowns {
			dm := sloErrorBudgetDrilldownModel{
				URL:     types.StringValue(d.Url),
				Label:   types.StringValue(d.Label),
				Trigger: types.StringValue(string(d.Trigger)),
				Type:    types.StringValue(string(d.Type)),
			}

			// encode_url: normalize API default (true) — only write if prior state had it set,
			// or if API returned false (non-default).
			if d.EncodeUrl != nil {
				priorEncodeURL := types.BoolNull()
				if i < len(priorDrilldowns) {
					priorEncodeURL = priorDrilldowns[i].EncodeURL
				}
				if typeutils.IsKnown(priorEncodeURL) || !*d.EncodeUrl {
					dm.EncodeURL = types.BoolValue(*d.EncodeUrl)
				}
			}

			// open_in_new_tab: same normalization.
			if d.OpenInNewTab != nil {
				priorOpenInNewTab := types.BoolNull()
				if i < len(priorDrilldowns) {
					priorOpenInNewTab = priorDrilldowns[i].OpenInNewTab
				}
				if typeutils.IsKnown(priorOpenInNewTab) || !*d.OpenInNewTab {
					dm.OpenInNewTab = types.BoolValue(*d.OpenInNewTab)
				}
			}

			newDrilldowns = append(newDrilldowns, dm)
		}
		existing.Drilldowns = newDrilldowns
	}
}
