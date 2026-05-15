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

package discoversession

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

// Registry panel type key (`discover_session` + `_config` => `discover_session_config` on PanelModel).
// Matches kbapi.KbnDashboardPanelTypeDiscoverSessionType (`discover_session`).
const panelType = "discover_session"

// Handler implements iface.Handler for `discover_session` dashboard panels (`discover_session_config`).
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

// FromAPI maps a kbapi discover_session panel into Terraform panel models (parity with legacy populateDiscoverSessionPanelFromAPI).
func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	dsPanel, err := item.AsKbnDashboardPanelTypeDiscoverSession()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.Grid = panelkit.GridFromAPI(dsPanel.Grid.X, dsPanel.Grid.Y, dsPanel.Grid.W, dsPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(dsPanel.Id)
	pm.ConfigJSON = panelkit.PanelConfigJSONNull()

	populateDiscoverSessionPanelFromAPI(ctx, pm, prior, dsPanel)
	return nil
}

// ToAPI serializes Terraform discover_session panel state into kbapi (parity with legacy discoverSessionPanelToAPI).
func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	gridTF := panelkit.GridToAPI(pm.Grid)
	grid := struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	}{
		H: gridTF.H,
		W: gridTF.W,
		X: gridTF.X,
		Y: gridTF.Y,
	}
	panelID := panelkit.IDToAPI(pm.ID)

	var dashTR *models.TimeRangeModel
	if dashboard != nil {
		dashTR = dashboard.TimeRange
	}
	return discoverSessionPanelToAPI(context.Background(), pm, grid, panelID, dashTR)
}
