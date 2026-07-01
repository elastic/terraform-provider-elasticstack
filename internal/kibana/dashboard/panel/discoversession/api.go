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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Registry panel type key (`discover_session` + `_config` => `discover_session_config` on PanelModel).
// Matches kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionType (`discover_session`).
const panelType = "discover_session"

// Handler implements iface.Handler for `discover_session` dashboard panels (`discover_session_config`).
type Handler struct{}

func (Handler) PanelType() string                 { return panelType }
func (Handler) SchemaAttribute() schema.Attribute { return SchemaAttribute() }

func (Handler) ClassifyJSON(map[string]any) bool { return false }

func (Handler) PopulateJSONDefaults(config map[string]any) map[string]any { return config }

func (Handler) PinnedHandler() iface.PinnedHandler { return nil }

func (Handler) AlignStateFromPlan(context.Context, *models.PanelModel, *models.PanelModel) {}

func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	block := attrs["discover_session_config"]
	if panelkit.AttrConcreteSet(block) {
		return diags
	}
	if panelkit.AttrUnknown(block) {
		return diags
	}
	diags.AddAttributeError(attrPath, "Missing discover_session panel configuration", "Discover session panels require `discover_session_config`.")
	return diags
}

// FromAPI maps a kbapi discover_session panel into Terraform panel models (parity with legacy populateDiscoverSessionPanelFromAPI).
func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return panelkit.SimpleFromAPI(ctx, pm, prior,
		item.AsKibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession,
		func(p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession) (kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, *string) {
			return p.Grid, p.Id
		},
		func(pm *models.PanelModel, prior *models.PanelModel, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession) diag.Diagnostics {
			return populateDiscoverSessionPanelFromAPI(ctx, pm, prior, p)
		},
	)
}

// ToAPI serializes Terraform discover_session panel state into kbapi (parity with legacy discoverSessionPanelToAPI).
func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	if diags := panelkit.RejectConfigJSON(pm, panelType); diags.HasError() {
		return kbapi.DashboardPanelItem{}, diags
	}

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
