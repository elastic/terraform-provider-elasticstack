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

package optionslist

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

func (Handler) PinnedHandler() iface.PinnedHandler { return newPinnedHandler() }

func (Handler) AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel) {
	_, _, _ = ctx, plan, state
}

func (Handler) FromAPI(_ context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	olPanel, err := item.AsKibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Dashboard panel decode", err.Error())
		return d
	}

	pm.Grid = panelkit.GridFromAPI(olPanel.Grid.X, olPanel.Grid.Y, olPanel.Grid.W, olPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(olPanel.Id)
	if configBytes, err := json.Marshal(olPanel.Config); err == nil {
		pm.ConfigJSON = customtypes.NewJSONWithDefaultsValue(string(configBytes), panelkit.PanelJSONDefaultsFunc())
	}
	return PopulateFromAPI(pm, prior, &olPanel)
}

func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	_ = dashboard
	if pm.OptionsListControlConfig == nil {
		diags.AddError(
			"Missing options list control panel configuration",
			"Options list control panels require `options_list_control_config`.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{
		Grid: grid,
		Id:   id,
	}
	diags.Append(BuildConfig(pm, &panel)...)
	if diags.HasError() {
		return kbapi.DashboardPanelItem{}, diags
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl(panel); err != nil {
		diags.AddError("Failed to create options list control panel", err.Error())
		return panelItem, diags
	}
	return panelItem, diags
}

// ValidatePanelConfig is a no-op for options_list panels. Historically this enforced presence of
// data_view_id/field_name via panelkit.ValidateDataViewFieldName, but now that those attributes
// live under the required `by_field` nested block (schema.Attribute.Required), Terraform enforces
// their presence natively whenever `by_field` is configured. Branch selection itself (exactly one
// of `by_field`/`by_esql`) is enforced by the ExactlyOneOfNestedAttrsValidator on the
// options_list_control_config schema attribute (see SchemaAttribute in schema.go).
func (Handler) ValidatePanelConfig(_ context.Context, _ map[string]attr.Value, _ path.Path) diag.Diagnostics {
	return nil
}
