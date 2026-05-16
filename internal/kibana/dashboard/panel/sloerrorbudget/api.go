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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelConfigAttrsKeyPrefix = panelType + "_config"

// Handler implements iface.Handler for the slo_error_budget dashboard panel discriminator.
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

// FromAPI fills pm from kbapi DashboardPanelItem for this panel discriminator.
func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	apiPanel, err := item.AsKbnDashboardPanelTypeSloErrorBudget()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Dashboard panel decode", err.Error())
		return d
	}

	pm.Grid = panelkit.GridFromAPI(apiPanel.Grid.X, apiPanel.Grid.Y, apiPanel.Grid.W, apiPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(apiPanel.Id)
	pm.ConfigJSON = panelkit.PanelConfigJSONNull()
	diags := PopulateFromAPI(pm, prior, apiPanel.Config)
	_ = ctx
	return diags
}

// ToAPI serializes Terraform panel state into a kbapi union item.
func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	panel := kbapi.KbnDashboardPanelTypeSloErrorBudget{
		Grid: grid,
		Id:   id,
		Type: kbapi.SloErrorBudget,
	}
	diags := BuildConfig(pm, &panel)
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeSloErrorBudget(panel); err != nil {
		diags.AddError("Failed to create SLO error budget panel", err.Error())
	}
	return panelItem, diags
}

func sloErrorBudgetAttrsShape(attrs map[string]attr.Value) (flat bool, obj types.Object, ok bool) {
	if attrs == nil {
		return false, types.Object{}, false
	}
	if _, id := attrs["slo_id"]; id {
		return true, types.Object{}, true
	}
	if raw, nested := attrs[panelConfigAttrsKeyPrefix]; nested {
		obj, ok := raw.(types.Object)
		return false, obj, ok
	}
	return false, types.Object{}, false
}

func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var out diag.Diagnostics
	flat, obj, shaped := sloErrorBudgetAttrsShape(attrs)
	if !shaped {
		out.AddAttributeError(attrPath.AtName(panelConfigAttrsKeyPrefix), "Missing slo_error_budget panel configuration", "SLO error budget panels require `slo_error_budget_config`.")
		return out
	}

	cfgPath := attrPath
	if !flat {
		cfgPath = attrPath.AtName(panelConfigAttrsKeyPrefix)
		nestedRaw := attrs[panelConfigAttrsKeyPrefix]
		if nestedRaw != nil {
			switch {
			case nestedRaw.IsUnknown():
				return out
			case nestedRaw.IsNull():
				out.AddAttributeError(cfgPath, "Missing slo_error_budget panel configuration", "SLO error budget panels require `slo_error_budget_config`.")
				return out
			}
		}
	}

	var sloVal attr.Value
	if flat {
		sloVal = attrs["slo_id"]
	} else {
		sloVal = obj.Attributes()["slo_id"]
	}

	if deferSLO, missSLO := panelkit.StringAttrDeferOrMissing(sloVal); !deferSLO && missSLO {
		out.AddAttributeError(cfgPath.AtName("slo_id"), `Invalid SLO error budget configuration`, "`slo_id` is required.")
	}
	return out
}
