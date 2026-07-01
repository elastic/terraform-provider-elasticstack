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

package aiopschangepointchart

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

// Handler implements the iface.Handler contract for aiops_change_point_chart panels.
type Handler struct {
	panelkit.NoopHandlerBase
}

func (Handler) PanelType() string                 { return panelType }
func (Handler) SchemaAttribute() schema.Attribute { return SchemaAttribute() }

func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return panelkit.SimpleFromAPI(ctx, pm, prior,
		item.AsKibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart,
		func(p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart) (kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, *string) {
			return p.Grid, p.Id
		},
		func(pm, prior *models.PanelModel, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart) diag.Diagnostics {
			return PopulateFromAPI(pm, prior, p.Config)
		},
	)
}

func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	return panelkit.SimpleToAPI(pm,
		func(grid kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, id *string) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart, diag.Diagnostics) {
			if diags := panelkit.RejectConfigJSON(pm, panelType); diags.HasError() {
				return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart{}, diags
			}
			panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart{Grid: grid, Id: id, Type: kbapi.AiopsChangePointChart}
			return panel, BuildConfig(pm, &panel)
		},
		func(item *kbapi.DashboardPanelItem, panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart) error {
			return item.FromKibanaHTTPAPIsKbnDashboardPanelTypeAiopsChangePointChart(panel)
		},
		"Failed to create AIOps change point chart panel",
	)
}

// ValidatePanelConfig enforces presence of data_view_id and metric_field for aiops_change_point_chart panels.
func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var out diag.Diagnostics
	flat, obj, cfgPath, skip, diags := panelkit.ResolveConfigBlock(attrs, attrPath, panelType+"_config",
		"Missing AIOps change point chart panel configuration",
		"AIOps change point chart panels require `aiops_change_point_chart_config`.",
		"data_view_id", "metric_field")
	out.Append(diags...)
	if skip {
		return out
	}

	if deferred, d := panelkit.ValidateRequiredStringField(attrs, obj, flat, cfgPath, "data_view_id",
		"Invalid AIOps change point chart configuration", "`data_view_id` is required."); !deferred {
		out.Append(d...)
	}
	if deferred, d := panelkit.ValidateRequiredStringField(attrs, obj, flat, cfgPath, "metric_field",
		"Invalid AIOps change point chart configuration", "`metric_field` is required."); !deferred {
		out.Append(d...)
	}
	return out
}
