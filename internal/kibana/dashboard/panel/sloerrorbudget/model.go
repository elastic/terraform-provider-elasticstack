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

package sloerrorbudget

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes the TF model fields into the API panel struct.
func BuildConfig(pm models.PanelModel, sebPanel *kbapi.KbnDashboardPanelTypeSloErrorBudget) diag.Diagnostics {
	cfg := pm.SloErrorBudgetConfig
	if cfg == nil {
		return nil
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
	return nil
}

// PopulateFromAPI reads back an SLO error budget config from the API response.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiConfig kbapi.SloErrorBudgetEmbeddable) diag.Diagnostics {
	existing := pm.SloErrorBudgetConfig

	var priorSloInstanceID types.String
	if prior != nil && prior.SloErrorBudgetConfig != nil {
		priorSloInstanceID = prior.SloErrorBudgetConfig.SloInstanceID
	} else if prior == nil {
		priorSloInstanceID = types.StringValue("*")
	}

	if existing == nil {
		if prior != nil {
			return nil
		}
		pm.SloErrorBudgetConfig = &models.SloErrorBudgetConfigModel{}
		existing = pm.SloErrorBudgetConfig
	}

	existing.SloID = types.StringValue(apiConfig.SloId)

	if typeutils.IsKnown(priorSloInstanceID) && apiConfig.SloInstanceId != nil && *apiConfig.SloInstanceId != "*" {
		existing.SloInstanceID = types.StringValue(*apiConfig.SloInstanceId)
	}

	if (prior == nil || typeutils.IsKnown(existing.Title)) && apiConfig.Title != nil {
		existing.Title = types.StringValue(*apiConfig.Title)
	}
	if (prior == nil || typeutils.IsKnown(existing.Description)) && apiConfig.Description != nil {
		existing.Description = types.StringValue(*apiConfig.Description)
	}
	if (prior == nil || typeutils.IsKnown(existing.HideTitle)) && apiConfig.HideTitle != nil {
		existing.HideTitle = types.BoolValue(*apiConfig.HideTitle)
	}
	if (prior == nil || typeutils.IsKnown(existing.HideBorder)) && apiConfig.HideBorder != nil {
		existing.HideBorder = types.BoolValue(*apiConfig.HideBorder)
	}

	if apiConfig.Drilldowns != nil {
		var priorDrilldowns []models.URLDrilldownModel
		if prior != nil && prior.SloErrorBudgetConfig != nil {
			priorDrilldowns = prior.SloErrorBudgetConfig.Drilldowns
		}

		newDrilldowns := make([]models.URLDrilldownModel, 0, len(*apiConfig.Drilldowns))
		for i, d := range *apiConfig.Drilldowns {
			dm := models.URLDrilldownModel{
				URL:   types.StringValue(d.Url),
				Label: types.StringValue(d.Label),
			}

			if d.EncodeUrl != nil {
				priorEncodeURL := types.BoolNull()
				if i < len(priorDrilldowns) {
					priorEncodeURL = priorDrilldowns[i].EncodeURL
				}
				if typeutils.IsKnown(priorEncodeURL) || !*d.EncodeUrl {
					dm.EncodeURL = types.BoolValue(*d.EncodeUrl)
				}
			}

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
	return nil
}
