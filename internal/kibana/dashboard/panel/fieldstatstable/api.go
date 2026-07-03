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

package fieldstatstable

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

// Handler implements iface.Handler for `field_stats_table` dashboard panels.
type Handler struct {
	panelkit.NoopHandlerBase
}

func (Handler) PanelType() string                 { return panelType }
func (Handler) SchemaAttribute() schema.Attribute { return SchemaAttribute() }

func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	block := attrs["field_stats_table_config"]
	if panelkit.AttrConcreteSet(block) {
		return diags
	}
	if panelkit.AttrUnknown(block) {
		return diags
	}
	diags.AddAttributeError(attrPath, "Missing field_stats_table panel configuration", "Field statistics table panels require `field_stats_table_config`.")
	return diags
}

func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return panelkit.SimpleFromAPI(ctx, pm, prior,
		item.AsKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable,
		func(p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable) (kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, *string) {
			return p.Grid, p.Id
		},
		populateFieldStatsTableFromAPI,
	)
}

func (Handler) ToAPI(pm models.PanelModel, _ *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	return panelkit.SimpleToAPI(pm,
		func(grid kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, id *string) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable, diag.Diagnostics) {
			if diags := panelkit.RejectConfigJSON(pm, panelType); diags.HasError() {
				return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{}, diags
			}
			panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable{
				Grid: grid,
				Id:   id,
				Type: kbapi.FieldStatsTable,
			}
			return panel, buildFieldStatsTableConfig(pm, &panel)
		},
		func(item *kbapi.DashboardPanelItem, panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable) error {
			return item.FromKibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable(panel)
		},
		"Failed to create field_stats_table panel",
	)
}
