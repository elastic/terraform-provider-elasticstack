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

package slooverview

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

// Handler implements iface.Handler for the slo_overview dashboard panel discriminator.
type Handler struct {
	panelkit.NoopHandlerBase
}

func (Handler) PanelType() string                 { return panelType }
func (Handler) SchemaAttribute() schema.Attribute { return SchemaAttribute() }

func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return panelkit.SimpleFromAPI(ctx, pm, prior, item,
		item.AsKibanaHTTPAPIsKbnDashboardPanelTypeSloOverview,
		func(p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloOverview) (kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, *string) {
			return p.Grid, p.Id
		},
		PopulateFromAPI,
	)
}

func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	return panelkit.SimpleToAPI(pm,
		func(grid kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, id *string) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloOverview, diag.Diagnostics) {
			panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloOverview{Grid: grid, Id: id, Type: kbapi.SloOverview}
			return panel, BuildConfig(pm, &panel)
		},
		func(item *kbapi.DashboardPanelItem, panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloOverview) error {
			return item.FromKibanaHTTPAPIsKbnDashboardPanelTypeSloOverview(panel)
		},
		"Failed to create SLO overview panel",
	)
}

func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var out diag.Diagnostics
	cfg := attrs[panelType+"_config"]
	if panelkit.AttrConcreteSet(cfg) || panelkit.AttrUnknown(cfg) {
		return out
	}
	out.AddAttributeError(attrPath, "Missing SLO overview panel configuration", "SLO overview panels require `slo_overview_config`.")
	return out
}
