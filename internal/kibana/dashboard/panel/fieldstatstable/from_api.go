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
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const fieldStatsViewTypeDataview = "dataview"
const fieldStatsViewTypeEsql = "esql"

func fieldStatsTableAPIConfigViewType(apiCfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats) string {
	raw, err := apiCfg.MarshalJSON()
	if err != nil {
		return ""
	}
	var probe struct {
		ViewType string `json:"view_type"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return ""
	}
	return probe.ViewType
}

func populateFieldStatsTableFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable) diag.Diagnostics {
	if prior == nil {
		cfg, diags := fieldStatsTableConfigFromAPIImport(apiPanel.Config)
		pm.FieldStatsTableConfig = cfg
		return diags
	}

	if pm.FieldStatsTableConfig == nil {
		cfg, diags := fieldStatsTableConfigFromAPIImport(apiPanel.Config)
		pm.FieldStatsTableConfig = cfg
		if prior == nil {
			return diags
		}
	}

	existing := pm.FieldStatsTableConfig
	if existing == nil {
		return nil
	}

	viewType := fieldStatsTableAPIConfigViewType(apiPanel.Config)
	switch viewType {
	case fieldStatsViewTypeEsql:
		cfg1, err := apiPanel.Config.AsKibanaHTTPAPIsDataVisualizerFieldStats1()
		if err != nil {
			return fieldStatsTableDecodeDiagnostics(err, "by_esql")
		}
		return mergeFieldStatsTableEsqlFromAPI(existing, prior, cfg1)
	case fieldStatsViewTypeDataview:
		fallthrough
	default:
		cfg0, err := apiPanel.Config.AsKibanaHTTPAPIsDataVisualizerFieldStats0()
		if err != nil {
			return fieldStatsTableDecodeDiagnostics(err, "by_dataview")
		}
		return mergeFieldStatsTableDataviewFromAPI(existing, prior, cfg0)
	}
}

func fieldStatsTableDecodeDiagnostics(err error, branch string) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Failed to decode field_stats_table API config",
		"Could not decode the API field_stats_table "+branch+" config: "+err.Error(),
	)
	return diags
}

func fieldStatsTableConfigFromAPIImport(apiCfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats) (*models.FieldStatsTableConfigModel, diag.Diagnostics) {
	switch fieldStatsTableAPIConfigViewType(apiCfg) {
	case fieldStatsViewTypeEsql:
		cfg1, err := apiCfg.AsKibanaHTTPAPIsDataVisualizerFieldStats1()
		if err != nil {
			return nil, fieldStatsTableDecodeDiagnostics(err, "by_esql")
		}
		return fieldStatsTableEsqlFromAPIImport(cfg1), nil
	default:
		cfg0, err := apiCfg.AsKibanaHTTPAPIsDataVisualizerFieldStats0()
		if err != nil {
			return nil, fieldStatsTableDecodeDiagnostics(err, "by_dataview")
		}
		return fieldStatsTableDataviewFromAPIImport(cfg0), nil
	}
}

func fieldStatsTableDataviewFromAPIImport(api kbapi.KibanaHTTPAPIsDataVisualizerFieldStats0) *models.FieldStatsTableConfigModel {
	return &models.FieldStatsTableConfigModel{
		ByDataview: fieldStatsTableByDataviewFromAPI(api),
	}
}

func fieldStatsTableEsqlFromAPIImport(api kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1) *models.FieldStatsTableConfigModel {
	return &models.FieldStatsTableConfigModel{
		ByEsql: fieldStatsTableByEsqlFromAPI(api),
	}
}

func fieldStatsTableByDataviewFromAPI(api kbapi.KibanaHTTPAPIsDataVisualizerFieldStats0) *models.FieldStatsTableByDataviewModel {
	return &models.FieldStatsTableByDataviewModel{
		DataViewID:        types.StringValue(api.DataViewId),
		ShowDistributions: types.BoolPointerValue(api.ShowDistributions),
		Title:             types.StringPointerValue(api.Title),
		Description:       types.StringPointerValue(api.Description),
		HideTitle:         types.BoolPointerValue(api.HideTitle),
		HideBorder:        types.BoolPointerValue(api.HideBorder),
		TimeRange:         panelkit.TimeRangeFromAPI(api.TimeRange, nil),
	}
}

func fieldStatsTableByEsqlFromAPI(api kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1) *models.FieldStatsTableByEsqlModel {
	return &models.FieldStatsTableByEsqlModel{
		Query:             types.StringValue(api.Query.Esql),
		ShowDistributions: types.BoolPointerValue(api.ShowDistributions),
		Title:             types.StringPointerValue(api.Title),
		Description:       types.StringPointerValue(api.Description),
		HideTitle:         types.BoolPointerValue(api.HideTitle),
		HideBorder:        types.BoolPointerValue(api.HideBorder),
		TimeRange:         panelkit.TimeRangeFromAPI(api.TimeRange, nil),
	}
}

func mergeFieldStatsTableDataviewFromAPI(
	existing *models.FieldStatsTableConfigModel,
	prior *models.PanelModel,
	api kbapi.KibanaHTTPAPIsDataVisualizerFieldStats0,
) diag.Diagnostics {
	priorCfg := prior.FieldStatsTableConfig
	if priorCfg == nil || priorCfg.ByDataview == nil {
		return nil
	}

	if existing.ByDataview == nil {
		existing.ByDataview = &models.FieldStatsTableByDataviewModel{}
	}
	branch := existing.ByDataview
	priorBranch := priorCfg.ByDataview

	branch.DataViewID = types.StringValue(api.DataViewId)
	branch.ShowDistributions = panelkit.PreserveBool(branch.ShowDistributions, api.ShowDistributions)
	branch.Title = panelkit.PreserveString(branch.Title, api.Title)
	branch.Description = panelkit.PreserveString(branch.Description, api.Description)
	branch.HideTitle = panelkit.PreserveBool(branch.HideTitle, api.HideTitle)
	branch.HideBorder = panelkit.PreserveBool(branch.HideBorder, api.HideBorder)
	branch.TimeRange = panelkit.MergeTimeRange(branch.TimeRange, api.TimeRange, priorBranch.TimeRange)

	preserveFieldStatsTableDataviewNullIntent(priorBranch, branch)
	existing.ByEsql = nil
	return nil
}

func mergeFieldStatsTableEsqlFromAPI(
	existing *models.FieldStatsTableConfigModel,
	prior *models.PanelModel,
	api kbapi.KibanaHTTPAPIsDataVisualizerFieldStats1,
) diag.Diagnostics {
	priorCfg := prior.FieldStatsTableConfig
	if priorCfg == nil || priorCfg.ByEsql == nil {
		return nil
	}

	if existing.ByEsql == nil {
		existing.ByEsql = &models.FieldStatsTableByEsqlModel{}
	}
	branch := existing.ByEsql
	priorBranch := priorCfg.ByEsql

	branch.Query = types.StringValue(api.Query.Esql)
	branch.ShowDistributions = panelkit.PreserveBool(branch.ShowDistributions, api.ShowDistributions)
	branch.Title = panelkit.PreserveString(branch.Title, api.Title)
	branch.Description = panelkit.PreserveString(branch.Description, api.Description)
	branch.HideTitle = panelkit.PreserveBool(branch.HideTitle, api.HideTitle)
	branch.HideBorder = panelkit.PreserveBool(branch.HideBorder, api.HideBorder)
	branch.TimeRange = panelkit.MergeTimeRange(branch.TimeRange, api.TimeRange, priorBranch.TimeRange)

	preserveFieldStatsTableEsqlNullIntent(priorBranch, branch)
	existing.ByDataview = nil
	return nil
}

func preserveFieldStatsTableDataviewNullIntent(prior, existing *models.FieldStatsTableByDataviewModel) {
	if prior == nil || existing == nil {
		return
	}
	if !typeutils.IsKnown(prior.ShowDistributions) {
		existing.ShowDistributions = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.Title) {
		existing.Title = types.StringNull()
	}
	if !typeutils.IsKnown(prior.Description) {
		existing.Description = types.StringNull()
	}
	if !typeutils.IsKnown(prior.HideTitle) {
		existing.HideTitle = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.HideBorder) {
		existing.HideBorder = types.BoolNull()
	}
	if prior.TimeRange == nil {
		existing.TimeRange = nil
	} else {
		existing.TimeRange = panelkit.PreserveTimeRangeNullIntentFromPrior(prior.TimeRange, existing.TimeRange)
	}
}

func preserveFieldStatsTableEsqlNullIntent(prior, existing *models.FieldStatsTableByEsqlModel) {
	if prior == nil || existing == nil {
		return
	}
	if !typeutils.IsKnown(prior.ShowDistributions) {
		existing.ShowDistributions = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.Title) {
		existing.Title = types.StringNull()
	}
	if !typeutils.IsKnown(prior.Description) {
		existing.Description = types.StringNull()
	}
	if !typeutils.IsKnown(prior.HideTitle) {
		existing.HideTitle = types.BoolNull()
	}
	if !typeutils.IsKnown(prior.HideBorder) {
		existing.HideBorder = types.BoolNull()
	}
	if prior.TimeRange == nil {
		existing.TimeRange = nil
	} else {
		existing.TimeRange = panelkit.PreserveTimeRangeNullIntentFromPrior(prior.TimeRange, existing.TimeRange)
	}
}
