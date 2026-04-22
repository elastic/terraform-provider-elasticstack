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
	SloID         types.String                   `tfsdk:"slo_id"`
	SloInstanceID types.String                   `tfsdk:"slo_instance_id"`
	Title         types.String                   `tfsdk:"title"`
	Description   types.String                   `tfsdk:"description"`
	HideTitle     types.Bool                     `tfsdk:"hide_title"`
	HideBorder    types.Bool                     `tfsdk:"hide_border"`
	Drilldowns    []sloErrorBudgetDrilldownModel `tfsdk:"drilldowns"`
}

type sloErrorBudgetDrilldownModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

// buildSloErrorBudgetConfig writes the TF model fields into the API panel struct.
func buildSloErrorBudgetConfig(pm panelModel, sebPanel *kbapi.KbnDashboardPanelTypeSloErrorBudget) {
	cfg := pm.SloErrorBudgetConfig
	if cfg == nil {
		return
	}

	sebPanel.Config.SloId = cfg.SloID.ValueString()

	if typeutils.IsKnown(cfg.SloInstanceID) {
		sebPanel.Config.SloInstanceId = cfg.SloInstanceID.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Title) {
		sebPanel.Config.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Description) {
		sebPanel.Config.Description = cfg.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.HideTitle) {
		sebPanel.Config.HideTitle = cfg.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.HideBorder) {
		sebPanel.Config.HideBorder = cfg.HideBorder.ValueBoolPointer()
	}

	if len(cfg.Drilldowns) > 0 {
		drilldowns := make([]struct {
			EncodeUrl    *bool                                           `json:"encode_url,omitempty"` //nolint:revive
			Label        string                                          `json:"label"`
			OpenInNewTab *bool                                           `json:"open_in_new_tab,omitempty"`
			Trigger      kbapi.SloErrorBudgetEmbeddableDrilldownsTrigger `json:"trigger"`
			Type         kbapi.SloErrorBudgetEmbeddableDrilldownsType    `json:"type"`
			Url          string                                          `json:"url"` //nolint:revive
		}, len(cfg.Drilldowns))

		for i, d := range cfg.Drilldowns {
			drilldowns[i].Url = d.URL.ValueString()
			drilldowns[i].Label = d.Label.ValueString()
			drilldowns[i].Trigger = kbapi.SloErrorBudgetEmbeddableDrilldownsTriggerOnOpenPanelMenu
			drilldowns[i].Type = kbapi.SloErrorBudgetEmbeddableDrilldownsTypeUrlDrilldown
			if typeutils.IsKnown(d.EncodeURL) {
				drilldowns[i].EncodeUrl = d.EncodeURL.ValueBoolPointer()
			}
			if typeutils.IsKnown(d.OpenInNewTab) {
				drilldowns[i].OpenInNewTab = d.OpenInNewTab.ValueBoolPointer()
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
	// Normalize "*" (all-instances wildcard) to null, matching create+refresh behaviour.
	if typeutils.IsKnown(priorSloInstanceID) && apiConfig.SloInstanceId != nil && *apiConfig.SloInstanceId != "*" {
		existing.SloInstanceID = types.StringValue(*apiConfig.SloInstanceId)
	}

	// Import has no prior intent, so populate all optional display fields. During normal
	// refreshes, preserve omitted values unless the practitioner configured them.
	if (tfPanel == nil || typeutils.IsKnown(existing.Title)) && apiConfig.Title != nil {
		existing.Title = types.StringValue(*apiConfig.Title)
	}
	if (tfPanel == nil || typeutils.IsKnown(existing.Description)) && apiConfig.Description != nil {
		existing.Description = types.StringValue(*apiConfig.Description)
	}
	if (tfPanel == nil || typeutils.IsKnown(existing.HideTitle)) && apiConfig.HideTitle != nil {
		existing.HideTitle = types.BoolValue(*apiConfig.HideTitle)
	}
	if (tfPanel == nil || typeutils.IsKnown(existing.HideBorder)) && apiConfig.HideBorder != nil {
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
				URL:   types.StringValue(d.Url),
				Label: types.StringValue(d.Label),
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
