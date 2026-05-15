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

package lensdashboardapp

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func lensDashboardAppByValueToAPI(
	byValue models.LensDashboardAppByValueModel,
	grid kbapi.KbnDashboardPanelGrid,
	panelID *string,
	parentDashboard *models.DashboardModel,
) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	if blocks, ok := lensByValueChartBlocksForTypedLensApp(byValue); ok {
		conv, okConv := lenscommon.FirstForBlocks(blocks)
		if !okConv {
			diags.AddError("Invalid `by_value` for lens-dashboard-app", "The typed by-value chart block could not be resolved to a Lens visualization converter.")
			return kbapi.DashboardPanelItem{}, diags
		}
		vis0, d := conv.BuildAttributes(blocks, lensChartResolver(parentDashboard))
		diags.Append(d...)
		if d.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		config, err := lensByValueConfigFromVisConfig0(vis0)
		if err != nil {
			diags.AddError("Invalid typed by-value config for lens-dashboard-app", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		ldPanel := kbapi.KbnDashboardPanelTypeLensDashboardApp{
			Config: config,
			Grid:   grid,
			Id:     panelID,
			Type:   kbapi.LensDashboardApp,
		}
		var panelItem kbapi.DashboardPanelItem
		if err := panelItem.FromKbnDashboardPanelTypeLensDashboardApp(ldPanel); err != nil {
			diags.AddError("Failed to create lens-dashboard-app panel", err.Error())
		}
		return panelItem, diags
	}
	if !typeutils.IsKnown(byValue.ConfigJSON) {
		diags.AddError(
			"Invalid `by_value.config_json` for lens-dashboard-app",
			"by_value.config_json is unknown. Ensure it is set to a non-null JSON value when using `config_json` as the by-value source.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}
	var config kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	if err := config.UnmarshalJSON([]byte(byValue.ConfigJSON.ValueString())); err != nil {
		diags.AddError("Invalid `by_value.config_json` for lens-dashboard-app", err.Error())
		return kbapi.DashboardPanelItem{}, diags
	}
	ldPanel := kbapi.KbnDashboardPanelTypeLensDashboardApp{
		Config: config,
		Grid:   grid,
		Id:     panelID,
		Type:   kbapi.LensDashboardApp,
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeLensDashboardApp(ldPanel); err != nil {
		diags.AddError("Failed to create lens-dashboard-app panel", err.Error())
	}
	return panelItem, diags
}

func lensDashboardAppByReferenceToAPI(
	byRef models.LensDashboardAppByReferenceModel,
	grid kbapi.KbnDashboardPanelGrid,
	panelID *string,
) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	api1, d := lenscommon.LensDashboardAppByReferenceModelToAPIConfig1(byRef, "by_reference.references_json")
	diags.Append(d...)
	if d.HasError() {
		return kbapi.DashboardPanelItem{}, diags
	}
	var config kbapi.KbnDashboardPanelTypeLensDashboardApp_Config
	if err := config.FromKbnDashboardPanelTypeLensDashboardAppConfig1(api1); err != nil {
		diags.AddError("Failed to set lens-dashboard-app by_reference config", err.Error())
		return kbapi.DashboardPanelItem{}, diags
	}
	ldPanel := kbapi.KbnDashboardPanelTypeLensDashboardApp{
		Config: config,
		Grid:   grid,
		Id:     panelID,
		Type:   kbapi.LensDashboardApp,
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeLensDashboardApp(ldPanel); err != nil {
		diags.AddError("Failed to create lens-dashboard-app panel", err.Error())
	}
	return panelItem, diags
}
