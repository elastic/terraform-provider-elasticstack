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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const fieldStatsViewTypeDataview = "dataview"
const fieldStatsViewTypeEsql = "esql"

// fieldStatsTableAPIConfigViewType extracts the `view_type` discriminator from the kbapi union
// type. kbapi's Config union unmarshals successfully into both generated branch structs, so we
// probe the raw JSON for the discriminator field instead. err is non-nil only if the union itself
// fails to (un)marshal, which is distinct from a missing/unexpected view_type value.
func fieldStatsTableAPIConfigViewType(apiCfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats) (string, error) {
	raw, err := apiCfg.MarshalJSON()
	if err != nil {
		return "", err
	}
	var probe struct {
		ViewType string `json:"view_type"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return "", err
	}
	return probe.ViewType, nil
}

// fieldStatsTablePriorTFBranchMismatchesAPI reports out-of-band branch changes (e.g. Kibana flipped
// dataview vs ES|QL). Prior Terraform state used exclusively one branch while the API payload uses the other.
func fieldStatsTablePriorTFBranchMismatchesAPI(viewType string, priorCfg *models.FieldStatsTableConfigModel) bool {
	if priorCfg == nil {
		return false
	}
	hasDataview := priorCfg.ByDataview != nil
	hasEsql := priorCfg.ByEsql != nil
	if viewType == fieldStatsViewTypeEsql && hasDataview && !hasEsql {
		return true
	}
	if viewType == fieldStatsViewTypeDataview && hasEsql && !hasDataview {
		return true
	}
	return false
}

func populateFieldStatsTableFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable) diag.Diagnostics {
	if prior == nil {
		cfg, diags := fieldStatsTableConfigFromAPIImport(apiPanel.Config)
		pm.FieldStatsTableConfig = cfg
		return diags
	}

	if pm.FieldStatsTableConfig == nil {
		cfg, diags := fieldStatsTableConfigFromAPIImport(apiPanel.Config)
		if diags.HasError() {
			return diags
		}
		pm.FieldStatsTableConfig = cfg
	}
	existing := pm.FieldStatsTableConfig

	viewType, err := fieldStatsTableAPIConfigViewType(apiPanel.Config)
	if err != nil {
		return fieldStatsTableProbeDiagnostics(err)
	}

	if fieldStatsTablePriorTFBranchMismatchesAPI(viewType, prior.FieldStatsTableConfig) {
		// Drift import: replace typed config from API so the next plan surfaces the branch change.
		imported, diags := fieldStatsTableConfigFromAPIImport(apiPanel.Config)
		if imported != nil {
			*existing = *imported
		}
		return diags
	}

	switch viewType {
	case fieldStatsViewTypeEsql:
		cfg1, err := apiPanel.Config.AsKibanaHTTPAPIsDataVisualizerFieldStats1()
		if err != nil {
			return fieldStatsTableDecodeDiagnostics(err, attrByEsql)
		}
		return mergeFieldStatsTableEsqlFromAPI(existing, prior, cfg1)
	case fieldStatsViewTypeDataview:
		cfg0, err := apiPanel.Config.AsKibanaHTTPAPIsDataVisualizerFieldStats0()
		if err != nil {
			return fieldStatsTableDecodeDiagnostics(err, attrByDataview)
		}
		return mergeFieldStatsTableDataviewFromAPI(existing, prior, cfg0)
	default:
		return fieldStatsTableInvalidViewTypeDiagnostics(viewType)
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

// fieldStatsTableProbeDiagnostics reports a failure to (un)marshal the kbapi union itself while
// probing for the view_type discriminator, distinct from a missing/unexpected view_type value.
func fieldStatsTableProbeDiagnostics(err error) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddError(
		"Failed to decode field_stats_table API config",
		fmt.Sprintf("Could not determine the field_stats_table view_type: %s.", err.Error()),
	)
	return diags
}

func fieldStatsTableInvalidViewTypeDiagnostics(viewType string) diag.Diagnostics {
	var diags diag.Diagnostics
	detail := "view_type is missing or invalid"
	if viewType != "" {
		detail = fmt.Sprintf("view_type has unexpected value %q", viewType)
	}
	diags.AddError(
		"Failed to decode field_stats_table API config",
		fmt.Sprintf("Could not decode the API field_stats_table config: %s; expected %q or %q.", detail, fieldStatsViewTypeDataview, fieldStatsViewTypeEsql),
	)
	return diags
}

func fieldStatsTableConfigFromAPIImport(apiCfg kbapi.KibanaHTTPAPIsDataVisualizerFieldStats) (*models.FieldStatsTableConfigModel, diag.Diagnostics) {
	viewType, err := fieldStatsTableAPIConfigViewType(apiCfg)
	if err != nil {
		return nil, fieldStatsTableProbeDiagnostics(err)
	}
	switch viewType {
	case fieldStatsViewTypeEsql:
		cfg1, err := apiCfg.AsKibanaHTTPAPIsDataVisualizerFieldStats1()
		if err != nil {
			return nil, fieldStatsTableDecodeDiagnostics(err, attrByEsql)
		}
		return &models.FieldStatsTableConfigModel{ByEsql: fieldStatsTableByEsqlFromAPI(cfg1)}, nil
	case fieldStatsViewTypeDataview:
		cfg0, err := apiCfg.AsKibanaHTTPAPIsDataVisualizerFieldStats0()
		if err != nil {
			return nil, fieldStatsTableDecodeDiagnostics(err, attrByDataview)
		}
		return &models.FieldStatsTableConfigModel{ByDataview: fieldStatsTableByDataviewFromAPI(cfg0)}, nil
	default:
		return nil, fieldStatsTableInvalidViewTypeDiagnostics(viewType)
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
	panelkit.ApplyPresentationFromAPI(&branch.Title, &branch.Description, &branch.HideTitle, &branch.HideBorder,
		api.Title, api.Description, api.HideTitle, api.HideBorder)
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
	panelkit.ApplyPresentationFromAPI(&branch.Title, &branch.Description, &branch.HideTitle, &branch.HideBorder,
		api.Title, api.Description, api.HideTitle, api.HideBorder)
	branch.TimeRange = panelkit.MergeTimeRange(branch.TimeRange, api.TimeRange, priorBranch.TimeRange)

	preserveFieldStatsTableEsqlNullIntent(priorBranch, branch)
	existing.ByDataview = nil
	return nil
}

// fieldStatsTableBranchNullIntentFields holds pointers to the shared fields of both branch model
// types (ByDataview and ByEsql) that require REQ-009 null-intent preservation, allowing
// preserveFieldStatsTableBranchNullIntent to handle both branches without duplicating logic.
type fieldStatsTableBranchNullIntentFields struct {
	ShowDistributions *types.Bool
	Title             *types.String
	Description       *types.String
	HideTitle         *types.Bool
	HideBorder        *types.Bool
	TimeRange         **models.TimeRangeModel
}

// preserveFieldStatsTableBranchNullIntent is the shared implementation for null-intent preservation
// across the dataview and ES|QL branch models.
func preserveFieldStatsTableBranchNullIntent(prior, existing fieldStatsTableBranchNullIntentFields) {
	if !typeutils.IsKnown(*prior.ShowDistributions) {
		*existing.ShowDistributions = types.BoolNull()
	}
	panelkit.NullPreservePresentationFromPrior(*prior.Title, *prior.Description, *prior.HideTitle, *prior.HideBorder,
		existing.Title, existing.Description, existing.HideTitle, existing.HideBorder)
	if *prior.TimeRange == nil {
		*existing.TimeRange = nil
	} else {
		*existing.TimeRange = panelkit.PreserveTimeRangeNullIntentFromPrior(*prior.TimeRange, *existing.TimeRange)
	}
}

func preserveFieldStatsTableDataviewNullIntent(prior, existing *models.FieldStatsTableByDataviewModel) {
	if prior == nil || existing == nil {
		return
	}
	preserveFieldStatsTableBranchNullIntent(
		fieldStatsTableBranchNullIntentFields{&prior.ShowDistributions, &prior.Title, &prior.Description, &prior.HideTitle, &prior.HideBorder, &prior.TimeRange},
		fieldStatsTableBranchNullIntentFields{&existing.ShowDistributions, &existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder, &existing.TimeRange},
	)
}

func preserveFieldStatsTableEsqlNullIntent(prior, existing *models.FieldStatsTableByEsqlModel) {
	if prior == nil || existing == nil {
		return
	}
	preserveFieldStatsTableBranchNullIntent(
		fieldStatsTableBranchNullIntentFields{&prior.ShowDistributions, &prior.Title, &prior.Description, &prior.HideTitle, &prior.HideBorder, &prior.TimeRange},
		fieldStatsTableBranchNullIntentFields{&existing.ShowDistributions, &existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder, &existing.TimeRange},
	)
}
