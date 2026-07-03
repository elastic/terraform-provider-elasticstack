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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func buildFieldStatsTableConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable) diag.Diagnostics {
	var diags diag.Diagnostics
	cfg := pm.FieldStatsTableConfig
	if cfg == nil {
		diags.AddError("Missing field_stats_table panel configuration", "Field statistics table panels require `field_stats_table_config`.")
		return diags
	}

	switch {
	case cfg.ByDataview != nil:
		api := kbapi.KibanaHTTPAPIsDataVisualizerFieldStats0{
			DataViewId: cfg.ByDataview.DataViewID.ValueString(),
			ViewType:   kbapi.Dataview,
		}
		applyFieldStatsTableBranchToAPI0(cfg.ByDataview, &api)
		if err := panel.Config.FromKibanaHTTPAPIsDataVisualizerFieldStats0(api); err != nil {
			diags.AddError("Failed to build field_stats_table config", err.Error())
		}
	case cfg.ByEsql != nil:
		api := kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1{
			ViewType: kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1ViewTypeEsql,
		}
		api.Query.Esql = cfg.ByEsql.Query.ValueString()
		applyFieldStatsTableBranchToAPI1(cfg.ByEsql, &api)
		if err := panel.Config.FromKibanaHTTPAPIsDataVisualizerFieldStats1(api); err != nil {
			diags.AddError("Failed to build field_stats_table config", err.Error())
		}
	default:
		diags.AddError("Invalid field_stats_table_config", "Exactly one of `by_dataview` or `by_esql` must be set.")
	}
	return diags
}

func applyFieldStatsTableBranchToAPI0(m *models.FieldStatsTableByDataviewModel, api *kbapi.KibanaHTTPAPIsDataVisualizerFieldStats0) {
	if typeutils.IsKnown(m.ShowDistributions) {
		api.ShowDistributions = m.ShowDistributions.ValueBoolPointer()
	}
	panelkit.BuildPresentationConfig(m.Title, m.Description, m.HideTitle, m.HideBorder,
		&api.Title, &api.Description, &api.HideTitle, &api.HideBorder)
	api.TimeRange = lenscommon.TimeRangeModelToAPI(m.TimeRange)
}

func applyFieldStatsTableBranchToAPI1(m *models.FieldStatsTableByEsqlModel, api *kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1) {
	if typeutils.IsKnown(m.ShowDistributions) {
		api.ShowDistributions = m.ShowDistributions.ValueBoolPointer()
	}
	panelkit.BuildPresentationConfig(m.Title, m.Description, m.HideTitle, m.HideBorder,
		&api.Title, &api.Description, &api.HideTitle, &api.HideBorder)
	api.TimeRange = lenscommon.TimeRangeModelToAPI(m.TimeRange)
}
