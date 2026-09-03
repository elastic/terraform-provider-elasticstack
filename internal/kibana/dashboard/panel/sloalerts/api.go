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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Handler implements iface.Handler for slo_alerts panels.
type Handler struct {
	panelkit.NoopHandlerBase
}

func (Handler) PanelType() string                 { return panelType }
func (Handler) SchemaAttribute() schema.Attribute { return SchemaAttribute() }

func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return panelkit.SimpleFromAPI(ctx, pm, prior,
		item.AsKibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts,
		func(p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts) (kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, *string) {
			return p.Grid, p.Id
		},
		func(pm, prior *models.PanelModel, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts) diag.Diagnostics {
			PopulateFromAPI(pm, prior, p)
			return nil
		},
	)
}

func (Handler) ToAPI(pm models.PanelModel, _ *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	return panelkit.SimpleToAPI(pm,
		func(grid kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, id *string) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts, diag.Diagnostics) {
			if diags := panelkit.RejectConfigJSON(pm, panelType); diags.HasError() {
				return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts{}, diags
			}
			if pm.SloAlertsConfig == nil {
				var diags diag.Diagnostics
				diags.AddError("Missing SLO alerts panel configuration", "SLO alerts panels require `slo_alerts_config`.")
				return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts{}, diags
			}
			panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts{Grid: grid, Id: id, Type: kbapi.SloAlerts}
			return panel, BuildConfig(&pm, &panel)
		},
		func(item *kbapi.DashboardPanelItem, panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts) error {
			return item.FromKibanaHTTPAPIsKbnDashboardPanelTypeSloAlerts(panel)
		},
		"Failed to create SLO alerts panel",
	)
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
