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

package esqlcontrol

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type esqlControlDisplaySettingsAPI = struct {
	HideActionBar *bool   `json:"hide_action_bar,omitempty"`
	HideExclude   *bool   `json:"hide_exclude,omitempty"`
	HideExists    *bool   `json:"hide_exists,omitempty"`
	HideSort      *bool   `json:"hide_sort,omitempty"`
	Placeholder   *string `json:"placeholder,omitempty"`
}

// esqlControlAPIData is a normalized view of either union branch of KbnDashboardPanelTypeEsqlControl_Config.
type esqlControlAPIData struct {
	SelectedOptions  []string
	VariableName     string
	VariableType     string
	EsqlQuery        string
	ControlType      string
	Title            *string
	SingleSelect     *bool
	AvailableOptions []string
	DisplaySettings  *esqlControlDisplaySettingsAPI
	ok               bool
}

func esqlControlAPIDataFromConfig(cfg kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeEsqlControl_Config) esqlControlAPIData {
	// Prefer static-values first: the VALUES_FROM_QUERY union branch is permissive enough that it
	// can incorrectly match some STATIC_VALUES payloads, dropping fields like available_options.
	if sv, err := cfg.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListEsqlControlSchemaStaticValues(); err == nil {
		return esqlControlAPIData{
			SelectedOptions:  sv.SelectedOptions,
			VariableName:     sv.VariableName,
			VariableType:     string(sv.VariableType),
			EsqlQuery:        "",
			ControlType:      string(sv.ControlType),
			Title:            sv.Title,
			SingleSelect:     sv.SingleSelect,
			AvailableOptions: sv.AvailableOptions,
			DisplaySettings:  sv.DisplaySettings,
			ok:               true,
		}
	}
	if vq, err := cfg.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListEsqlControlSchemaValuesFromQuery(); err == nil {
		return esqlControlAPIData{
			SelectedOptions: vq.SelectedOptions,
			VariableName:    vq.VariableName,
			VariableType:    string(vq.VariableType),
			EsqlQuery:       vq.EsqlQuery,
			ControlType:     string(vq.ControlType),
			Title:           vq.Title,
			SingleSelect:    vq.SingleSelect,
			DisplaySettings: vq.DisplaySettings,
			ok:              true,
		}
	}
	return esqlControlAPIData{}
}

// PopulateFromAPI reads back an ES|QL control config from the API response and
// updates the panel model. Null-preservation semantics apply: if a field is null in the
// existing TF state, we do not overwrite it with a Kibana-returned value. If there is no
// existing config block, and Kibana returns an empty/absent config, we leave
// EsqlControlConfig as nil.
//
// tfPanel is the prior TF state/plan panel, or nil on import. When nil, the function
// populates all API-returned fields unconditionally (no prior intent to preserve).
func PopulateFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeEsqlControl_Config) {
	api := esqlControlAPIDataFromConfig(apiConfig)
	if !api.ok {
		return
	}
	existing := pm.EsqlControlConfig

	// On import (tfPanel == nil) there is no prior intent — populate from API.
	if tfPanel == nil {
		existing = &models.EsqlControlConfigModel{
			SelectedOptions:  typeutils.StringsToListMust(api.SelectedOptions),
			VariableName:     types.StringValue(api.VariableName),
			VariableType:     types.StringValue(api.VariableType),
			EsqlQuery:        types.StringValue(api.EsqlQuery),
			ControlType:      types.StringValue(api.ControlType),
			Title:            types.StringPointerValue(api.Title),
			SingleSelect:     types.BoolPointerValue(api.SingleSelect),
			AvailableOptions: types.ListNull(types.StringType),
		}
		pm.EsqlControlConfig = existing
		if len(api.AvailableOptions) > 0 {
			existing.AvailableOptions = typeutils.StringsToListMust(api.AvailableOptions)
		}
		if api.DisplaySettings != nil {
			d := api.DisplaySettings
			existing.DisplaySettings = &models.EsqlControlDisplaySettingsModel{
				Placeholder:   types.StringPointerValue(d.Placeholder),
				HideActionBar: types.BoolPointerValue(d.HideActionBar),
				HideExclude:   types.BoolPointerValue(d.HideExclude),
				HideExists:    types.BoolPointerValue(d.HideExists),
				HideSort:      types.BoolPointerValue(d.HideSort),
			}
		}
		return
	}

	if existing == nil {
		if tfPanel == nil || tfPanel.EsqlControlConfig == nil {
			return
		}
		existing = &models.EsqlControlConfigModel{
			SelectedOptions:  typeutils.StringsToListMust(api.SelectedOptions),
			VariableName:     types.StringValue(api.VariableName),
			VariableType:     types.StringValue(api.VariableType),
			EsqlQuery:        types.StringValue(api.EsqlQuery),
			ControlType:      types.StringValue(api.ControlType),
			Title:            types.StringPointerValue(api.Title),
			SingleSelect:     types.BoolPointerValue(api.SingleSelect),
			AvailableOptions: types.ListNull(types.StringType),
		}
		pm.EsqlControlConfig = existing
		if len(api.AvailableOptions) > 0 {
			existing.AvailableOptions = typeutils.StringsToListMust(api.AvailableOptions)
		}
		if api.DisplaySettings != nil {
			d := api.DisplaySettings
			existing.DisplaySettings = &models.EsqlControlDisplaySettingsModel{
				Placeholder:   types.StringPointerValue(d.Placeholder),
				HideActionBar: types.BoolPointerValue(d.HideActionBar),
				HideExclude:   types.BoolPointerValue(d.HideExclude),
				HideExists:    types.BoolPointerValue(d.HideExists),
				HideSort:      types.BoolPointerValue(d.HideSort),
			}
		}
	}

	prevQuery := existing.EsqlQuery
	prevTitle := existing.Title
	prevAvailableOptions := existing.AvailableOptions

	// Required fields always get updated from API.
	existing.SelectedOptions = typeutils.StringsToListMust(api.SelectedOptions)
	existing.VariableName = types.StringValue(api.VariableName)
	existing.VariableType = types.StringValue(api.VariableType)
	existing.EsqlQuery = types.StringValue(api.EsqlQuery)
	existing.ControlType = types.StringValue(api.ControlType)

	// Optional fields: update only if already known (non-null) in state.
	if typeutils.IsKnown(existing.Title) && api.Title != nil {
		existing.Title = types.StringValue(*api.Title)
	}
	if typeutils.IsKnown(existing.SingleSelect) && api.SingleSelect != nil {
		existing.SingleSelect = types.BoolValue(*api.SingleSelect)
	}

	// available_options: if TF state had it set (known, non-null list), update from API.
	if typeutils.IsKnown(existing.AvailableOptions) && len(api.AvailableOptions) > 0 {
		existing.AvailableOptions = typeutils.StringsToListMust(api.AvailableOptions)
	}
	lenscommon.PreserveKnownStringIfStateBlank(prevQuery, &existing.EsqlQuery)
	lenscommon.PreserveKnownStringIfStateBlank(prevTitle, &existing.Title)
	lenscommon.PreserveKnownTfValueIfStateNull(prevAvailableOptions, &existing.AvailableOptions)

	// display_settings: if block is present in state, update from API; otherwise preserve nil.
	if existing.DisplaySettings != nil && api.DisplaySettings != nil {
		ds := existing.DisplaySettings
		apiDS := api.DisplaySettings
		if typeutils.IsKnown(ds.Placeholder) && apiDS.Placeholder != nil {
			ds.Placeholder = types.StringValue(*apiDS.Placeholder)
		}
		if typeutils.IsKnown(ds.HideActionBar) && apiDS.HideActionBar != nil {
			ds.HideActionBar = types.BoolValue(*apiDS.HideActionBar)
		}
		if typeutils.IsKnown(ds.HideExclude) && apiDS.HideExclude != nil {
			ds.HideExclude = types.BoolValue(*apiDS.HideExclude)
		}
		if typeutils.IsKnown(ds.HideExists) && apiDS.HideExists != nil {
			ds.HideExists = types.BoolValue(*apiDS.HideExists)
		}
		if typeutils.IsKnown(ds.HideSort) && apiDS.HideSort != nil {
			ds.HideSort = types.BoolValue(*apiDS.HideSort)
		}
	}

	if tfPanel != nil && tfPanel.EsqlControlConfig != nil {
		esqlControlPreserveNullIntentFromPrior(tfPanel.EsqlControlConfig, existing)
	}
}

func esqlControlPreserveNullIntentFromPrior(prior, existing *models.EsqlControlConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	panelkit.NullPreserveBoolFromPrior(prior.SingleSelect, &existing.SingleSelect)
	panelkit.NullPreserveStringFromPrior(prior.Title, &existing.Title)
	panelkit.NullPreserveListFromPrior(prior.AvailableOptions, &existing.AvailableOptions)
	if prior.DisplaySettings == nil {
		existing.DisplaySettings = nil
	}
}

// BuildConfig writes TF model fields into the API panel union.
func BuildConfig(pm models.PanelModel, esqlPanel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeEsqlControl) diag.Diagnostics {
	var diags diag.Diagnostics
	cfg := pm.EsqlControlConfig
	if cfg == nil {
		return diags
	}

	displayToAPI := func(ds *models.EsqlControlDisplaySettingsModel) *esqlControlDisplaySettingsAPI {
		if ds == nil {
			return nil
		}
		out := &esqlControlDisplaySettingsAPI{}
		if typeutils.IsKnown(ds.Placeholder) {
			out.Placeholder = ds.Placeholder.ValueStringPointer()
		}
		if typeutils.IsKnown(ds.HideActionBar) {
			out.HideActionBar = ds.HideActionBar.ValueBoolPointer()
		}
		if typeutils.IsKnown(ds.HideExclude) {
			out.HideExclude = ds.HideExclude.ValueBoolPointer()
		}
		if typeutils.IsKnown(ds.HideExists) {
			out.HideExists = ds.HideExists.ValueBoolPointer()
		}
		if typeutils.IsKnown(ds.HideSort) {
			out.HideSort = ds.HideSort.ValueBoolPointer()
		}
		return out
	}

	ct := cfg.ControlType.ValueString()
	if kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListEsqlControlSchemaValuesFromQueryControlType(ct) == kbapi.VALUESFROMQUERY {
		vq := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListEsqlControlSchemaValuesFromQuery{
			SelectedOptions: typeutils.ListToStringsMust(cfg.SelectedOptions),
			VariableName:    cfg.VariableName.ValueString(),
			VariableType: kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListEsqlControlSchemaValuesFromQueryVariableType(
				cfg.VariableType.ValueString(),
			),
			EsqlQuery:   cfg.EsqlQuery.ValueString(),
			ControlType: kbapi.VALUESFROMQUERY,
		}
		if typeutils.IsKnown(cfg.Title) {
			vq.Title = cfg.Title.ValueStringPointer()
		}
		if typeutils.IsKnown(cfg.SingleSelect) {
			vq.SingleSelect = cfg.SingleSelect.ValueBoolPointer()
		}
		vq.DisplaySettings = displayToAPI(cfg.DisplaySettings)
		if err := esqlPanel.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListEsqlControlSchemaValuesFromQuery(vq); err != nil {
			diags.AddError("Failed to build esql control values_from_query config", err.Error())
		}
		return diags
	}

	sv := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListEsqlControlSchemaStaticValues{
		SelectedOptions: typeutils.ListToStringsMust(cfg.SelectedOptions),
		VariableName:    cfg.VariableName.ValueString(),
		VariableType: kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListEsqlControlSchemaStaticValuesVariableType(
			cfg.VariableType.ValueString(),
		),
		ControlType: kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListEsqlControlSchemaStaticValuesControlType(ct),
	}
	if typeutils.IsKnown(cfg.Title) {
		sv.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.SingleSelect) {
		sv.SingleSelect = cfg.SingleSelect.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.AvailableOptions) {
		sv.AvailableOptions = typeutils.ListToStringsMust(cfg.AvailableOptions)
	}
	sv.DisplaySettings = displayToAPI(cfg.DisplaySettings)
	if err := esqlPanel.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListEsqlControlSchemaStaticValues(sv); err != nil {
		diags.AddError("Failed to build esql control static_values config", err.Error())
	}
	return diags
}

// AlignEsqlPanels reapplies practitioner plan fields when Kibana echoes sparse reads (parity with dashboard.alignEsqlControlStateFromPlan).
func AlignEsqlPanels(plan, state *models.PanelModel) {
	if plan == nil || state == nil {
		return
	}
	alignEsql(plan.EsqlControlConfig, state.EsqlControlConfig)
}

func alignEsql(plan, state *models.EsqlControlConfigModel) {
	if plan == nil || state == nil {
		return
	}
	lenscommon.PreserveKnownStringIfStateBlank(plan.EsqlQuery, &state.EsqlQuery)
	lenscommon.PreserveKnownStringIfStateBlank(plan.Title, &state.Title)
	lenscommon.PreserveKnownTfValueIfStateNull(plan.AvailableOptions, &state.AvailableOptions)
}
