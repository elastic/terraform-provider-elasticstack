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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Registry panel type key (`lens_dashboard_app` + `_config` => `lens_dashboard_app_config` on PanelModel).
// Kibana's wire discriminator remains `lens-dashboard-app` (see kbapi.LensDashboardApp).
const panelType = "lens_dashboard_app"

// Handler implements iface.Handler for `lens-dashboard-app` dashboard panels (`lens_dashboard_app_config`).
type Handler struct{}

func (Handler) PanelType() string                 { return panelType }
func (Handler) SchemaAttribute() schema.Attribute { return SchemaAttribute() }

func (Handler) ClassifyJSON(map[string]any) bool { return false }

func (Handler) PopulateJSONDefaults(config map[string]any) map[string]any { return config }

func (Handler) PinnedHandler() iface.PinnedHandler { return nil }

func (Handler) AlignStateFromPlan(context.Context, *models.PanelModel, *models.PanelModel) {}

func (Handler) ValidatePanelConfig(context.Context, map[string]attr.Value, path.Path) diag.Diagnostics {
	return nil
}

// FromAPI maps a kbapi lens-dashboard-app panel into Terraform panel models (parity with legacy populateLensDashboardAppFromAPI).
func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	ldPanel, err := item.AsKbnDashboardPanelTypeLensDashboardApp()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	dashboardModel := iface.EnclosingDashboard(ctx)

	pm.Grid = panelkit.GridFromAPI(ldPanel.Grid.X, ldPanel.Grid.Y, ldPanel.Grid.W, ldPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(ldPanel.Id)
	pm.ConfigJSON = panelkit.PanelConfigJSONNull()

	return populateLensDashboardAppFromAPI(ctx, dashboardModel, pm, prior, ldPanel)
}

// ToAPI serializes Terraform lens-dashboard-app panel state into kbapi (parity with legacy lensDashboardAppToAPI).
func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	gridTF := panelkit.GridToAPI(pm.Grid)
	grid := kbapi.KbnDashboardPanelGrid{
		H: gridTF.H,
		W: gridTF.W,
		X: gridTF.X,
		Y: gridTF.Y,
	}
	panelID := panelkit.IDToAPI(pm.ID)

	cfg := pm.LensDashboardAppConfig
	if cfg == nil {
		var diags diag.Diagnostics
		diags.AddError("Missing `lens_dashboard_app_config`", "The `lens_dashboard_app_config` block is required for `lens-dashboard-app` panels.")
		return kbapi.DashboardPanelItem{}, diags
	}
	switch {
	case cfg.ByValue != nil:
		return lensDashboardAppByValueToAPI(*cfg.ByValue, grid, panelID, dashboard)
	case cfg.ByReference != nil:
		return lensDashboardAppByReferenceToAPI(*cfg.ByReference, grid, panelID)
	default:
		var diags diag.Diagnostics
		diags.AddError("Invalid `lens_dashboard_app_config`", "Exactly one of `by_value` or `by_reference` must be set inside `lens_dashboard_app_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}
}
