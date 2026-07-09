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
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes the TF model fields into the API panel struct.
func BuildConfig(pm models.PanelModel, sebPanel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloErrorBudget) diag.Diagnostics {
	cfg := pm.SloErrorBudgetConfig
	if cfg == nil {
		return nil
	}

	sebPanel.Config.SloId = cfg.SloID.ValueString()

	if typeutils.IsKnown(cfg.SloInstanceID) {
		sebPanel.Config.SloInstanceId = cfg.SloInstanceID.ValueStringPointer()
	}
	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&sebPanel.Config.Title, &sebPanel.Config.Description, &sebPanel.Config.HideTitle, &sebPanel.Config.HideBorder)

	var diags diag.Diagnostics
	if len(cfg.Drilldowns) > 0 {
		diags.Append(panelkit.InjectDrilldownsJSON(&sebPanel.Config, cfg.Drilldowns)...)
	}
	return diags
}

// PopulateFromAPI reads back an SLO error budget config from the API response.
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsSloErrorBudgetEmbeddable) diag.Diagnostics {
	existing := pm.SloErrorBudgetConfig

	var priorSloInstanceID types.String
	if prior != nil && prior.SloErrorBudgetConfig != nil {
		priorSloInstanceID = prior.SloErrorBudgetConfig.SloInstanceID
	} else if prior == nil {
		priorSloInstanceID = types.StringValue("*")
	}

	if existing == nil {
		if prior != nil && prior.SloErrorBudgetConfig == nil {
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
		existing.Drilldowns = sloErrorBudgetDrilldownsFromAPI(apiConfig.Drilldowns, priorDrilldowns)
	}
	return nil
}

func sloErrorBudgetDrilldownsFromAPI(
	apiDrilldowns *[]struct {
		EncodeUrl    *bool                                                         `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                        `json:"label"`
		OpenInNewTab *bool                                                         `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KibanaHTTPAPIsSloErrorBudgetEmbeddableDrilldownsTrigger `json:"trigger"`
		Type         kbapi.KibanaHTTPAPIsSloErrorBudgetEmbeddableDrilldownsType    `json:"type"`
		Url          string                                                        `json:"url"` //nolint:revive
	},
	priorDrilldowns []models.URLDrilldownModel,
) []models.URLDrilldownModel {
	if apiDrilldowns == nil || len(*apiDrilldowns) == 0 {
		return nil
	}
	b, err := json.Marshal(*apiDrilldowns)
	if err != nil {
		return nil
	}
	return panelkit.ReadDrilldownsFromWireJSON(b, priorDrilldowns)
}
