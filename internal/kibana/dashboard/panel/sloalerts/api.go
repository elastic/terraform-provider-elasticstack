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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Handler implements iface.Handler for slo_alerts panels.
type Handler struct{}

func (Handler) PanelType() string                  { return panelType }
func (Handler) SchemaAttribute() schema.Attribute  { return SchemaAttribute() }
func (Handler) ClassifyJSON(_ map[string]any) bool { return false }
func (Handler) PopulateJSONDefaults(config map[string]any) map[string]any {
	return config
}
func (Handler) PinnedHandler() iface.PinnedHandler { return nil }
func (Handler) AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel) {
	_, _, _ = ctx, plan, state
}

func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	apiPanel, err := item.AsKbnDashboardPanelTypeSloAlerts()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Dashboard panel decode", err.Error())
		return d
	}

	pm.Grid = panelkit.GridFromAPI(apiPanel.Grid.X, apiPanel.Grid.Y, apiPanel.Grid.W, apiPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(apiPanel.Id)
	pm.ConfigJSON = panelkit.PanelConfigJSONNull()
	PopulateFromAPI(pm, prior, apiPanel)
	_ = ctx
	return nil
}

func (Handler) ToAPI(pm models.PanelModel, _ *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := pm.SloAlertsConfig
	if cfg == nil {
		diags.AddError("Missing SLO alerts panel configuration", "SLO alerts panels require `slo_alerts_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}

	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	panel := kbapi.KbnDashboardPanelTypeSloAlerts{
		Grid: grid,
		Id:   id,
		Type: kbapi.SloAlerts,
	}
	BuildConfig(&pm, &panel)

	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeSloAlerts(panel); err != nil {
		diags.AddError("Failed to create SLO alerts panel", err.Error())
	}
	return panelItem, diags
}

func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	cv := attrs[panelConfigBlock]
	if panelkit.AttrConcreteSet(cv) || panelkit.AttrUnknown(cv) {
		return diags
	}
	diags.AddAttributeError(attrPath, "Missing SLO alerts panel configuration", "SLO alerts panels require `slo_alerts_config`.")
	return diags
}
