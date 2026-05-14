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
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func configPriorForVisRead(tfPanel, pm *models.PanelModel) *models.VisConfigModel {
	if tfPanel != nil && tfPanel.VisConfig != nil {
		return tfPanel.VisConfig
	}
	if pm != nil && pm.VisConfig != nil {
		return pm.VisConfig
	}
	return nil
}

// populateVisByReferenceFromAPI maps API vis config branch 1 (by-reference saved object panel).
func populateVisByReferenceFromAPI(
	ctx context.Context,
	prior *models.VisConfigModel,
	pm *models.PanelModel,
	cfg1 kbapi.KbnDashboardPanelTypeVisConfig1,
) diag.Diagnostics {
	var priorBR *models.LensDashboardAppByReferenceModel
	if prior != nil {
		priorBR = prior.ByReference
	}

	var lensCfg kbapi.KbnDashboardPanelTypeLensDashboardAppConfig1
	payload, err := json.Marshal(cfg1)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	if err := json.Unmarshal(payload, &lensCfg); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	by, diags := lenscommon.PopulateLensByReferenceTFModelFromLensAppConfig1(ctx, lensCfg, priorBR)
	if diags.HasError() {
		return diags
	}

	brCopy := by
	pm.VisConfig = &models.VisConfigModel{
		ByReference: &brCopy,
	}
	return diags
}

func visByReferenceToAPI(
	byRef models.LensDashboardAppByReferenceModel,
	grid struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	},
	panelID *string,
) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	api1Lens, d := lenscommon.LensDashboardAppByReferenceModelToAPIConfig1(byRef, "vis_config.by_reference.references_json")
	diags.Append(d...)
	if d.HasError() {
		return kbapi.DashboardPanelItem{}, diags
	}
	api1, convDiags := lenscommon.VisByReferenceConfig1FromLens(api1Lens)
	diags.Append(convDiags...)
	if convDiags.HasError() {
		return kbapi.DashboardPanelItem{}, diags
	}
	var config kbapi.KbnDashboardPanelTypeVis_Config
	if err := config.FromKbnDashboardPanelTypeVisConfig1(api1); err != nil {
		diags.AddError("Failed to set vis by_reference config", err.Error())
		return kbapi.DashboardPanelItem{}, diags
	}
	visPanel := kbapi.KbnDashboardPanelTypeVis{
		Config: config,
		Grid:   grid,
		Id:     panelID,
		Type:   kbapi.Vis,
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeVis(visPanel); err != nil {
		diags.AddError("Failed to create visualization panel", err.Error())
	}
	return panelItem, diags
}

func visConfigToAPI(pm models.PanelModel, dashboard *models.DashboardModel, grid struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
}, panelID *string) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := pm.VisConfig
	if cfg == nil {
		diags.AddError("Missing `vis_config`", "The `vis_config` block is required for typed `vis` panels.")
		return kbapi.DashboardPanelItem{}, diags
	}
	switch {
	case cfg.ByReference != nil:
		return visByReferenceToAPI(*cfg.ByReference, grid, panelID)
	case cfg.ByValue != nil:
		blocks := &cfg.ByValue.LensByValueChartBlocks
		conv, okConv := lenscommon.FirstForBlocks(blocks)
		if !okConv {
			diags.AddError("Invalid `vis_config.by_value`", "The typed chart block could not be resolved to a Lens visualization converter.")
			return kbapi.DashboardPanelItem{}, diags
		}
		config0, d := conv.BuildAttributes(blocks, lensChartResolver(dashboard))
		diags.Append(d...)
		if d.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		var config kbapi.KbnDashboardPanelTypeVis_Config
		if err := config.FromKbnDashboardPanelTypeVisConfig0(config0); err != nil {
			diags.AddError("Failed to create visualization panel config", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		visPanel := kbapi.KbnDashboardPanelTypeVis{
			Config: config,
			Grid:   grid,
			Id:     panelID,
			Type:   kbapi.Vis,
		}
		var panelItem kbapi.DashboardPanelItem
		if err := panelItem.FromKbnDashboardPanelTypeVis(visPanel); err != nil {
			diags.AddError("Failed to create visualization panel", err.Error())
		}
		return panelItem, diags
	default:
		diags.AddError("Invalid `vis_config`", "Exactly one of `by_value` or `by_reference` must be set inside `vis_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}
}
