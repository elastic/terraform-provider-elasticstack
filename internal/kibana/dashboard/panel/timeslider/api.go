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

package timeslider

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type Handler struct{}

func (Handler) PanelType() string                  { return panelType }
func (Handler) SchemaAttribute() schema.Attribute  { return SchemaAttribute() }
func (Handler) ClassifyJSON(_ map[string]any) bool { return false }
func (Handler) PopulateJSONDefaults(config map[string]any) map[string]any {
	return config
}

func (Handler) PinnedHandler() iface.PinnedHandler { return pinnedHandler{} }

func (Handler) AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel) {
	_, _, _ = ctx, plan, state
}

// FromAPI maps a dashboard time_slider_control panel union item onto pm (null-preserving).
func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	tsPanel, err := item.AsKbnDashboardPanelTypeTimeSliderControl()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Dashboard panel decode", err.Error())
		return d
	}

	pm.Grid = panelkit.GridFromAPI(tsPanel.Grid.X, tsPanel.Grid.Y, tsPanel.Grid.W, tsPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(tsPanel.Id)
	if configBytes, err := json.Marshal(tsPanel.Config); err == nil {
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsValue(string(configBytes), panelkit.PanelJSONDefaultsFunc())
	}
	PopulateFromAPI(pm, prior, tsPanel.Config)
	_ = ctx
	return nil
}

// ToAPI serializes Terraform panel state into a kbapi union item.
func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	panel := kbapi.KbnDashboardPanelTypeTimeSliderControl{
		Grid: grid,
		Id:   id,
		Config: struct {
			EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
			IsAnchored                 *bool    `json:"is_anchored,omitempty"`
			StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
		}{},
	}
	BuildConfig(pm, &panel)
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeTimeSliderControl(panel); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to create time slider control panel", err.Error())
		return panelItem, diags
	}
	return panelItem, nil
}

// ValidatePanelConfig is a no-op: all time_slider_control_config attributes are optional.
func (Handler) ValidatePanelConfig(_ context.Context, _ map[string]attr.Value, _ path.Path) diag.Diagnostics {
	return nil
}
