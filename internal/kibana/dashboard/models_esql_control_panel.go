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

package dashboard

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type esqlControlDisplaySettingsModel struct {
	Placeholder   types.String `tfsdk:"placeholder"`
	HideActionBar types.Bool   `tfsdk:"hide_action_bar"`
	HideExclude   types.Bool   `tfsdk:"hide_exclude"`
	HideExists    types.Bool   `tfsdk:"hide_exists"`
	HideSort      types.Bool   `tfsdk:"hide_sort"`
}

type esqlControlConfigModel struct {
	SelectedOptions  types.List                       `tfsdk:"selected_options"`
	VariableName     types.String                     `tfsdk:"variable_name"`
	VariableType     types.String                     `tfsdk:"variable_type"`
	EsqlQuery        types.String                     `tfsdk:"esql_query"`
	ControlType      types.String                     `tfsdk:"control_type"`
	Title            types.String                     `tfsdk:"title"`
	SingleSelect     types.Bool                       `tfsdk:"single_select"`
	AvailableOptions types.List                       `tfsdk:"available_options"`
	DisplaySettings  *esqlControlDisplaySettingsModel `tfsdk:"display_settings"`
}

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

func esqlControlAPIDataFromConfig(cfg kbapi.KbnDashboardPanelTypeEsqlControl_Config) esqlControlAPIData {
	// Prefer static-values first: the VALUES_FROM_QUERY union branch is permissive enough that it
	// can incorrectly match some STATIC_VALUES payloads, dropping fields like available_options.
	if sv, err := cfg.AsKbnControlsSchemasOptionsListEsqlControlSchemaStaticValues(); err == nil {
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
	if vq, err := cfg.AsKbnControlsSchemasOptionsListEsqlControlSchemaValuesFromQuery(); err == nil {
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

// stringsToList converts a []string to a types.List of string elements.
func stringsToList(strs []string) types.List {
	vals := make([]attr.Value, len(strs))
	for i, s := range strs {
		vals[i] = types.StringValue(s)
	}
	return types.ListValueMust(types.StringType, vals)
}

// listToStrings extracts a []string from a types.List of string elements.
func listToStrings(list types.List) []string {
	elems := list.Elements()
	strs := make([]string, len(elems))
	for i, v := range elems {
		strs[i] = v.(types.String).ValueString()
	}
	return strs
}

// populateEsqlControlFromAPI reads back an ES|QL control config from the API response and
// updates the panel model. Null-preservation semantics apply: if a field is null in the
// existing TF state, we do not overwrite it with a Kibana-returned value. If there is no
// existing config block, and Kibana returns an empty/absent config, we leave
// EsqlControlConfig as nil.
//
// tfPanel is the prior TF state/plan panel, or nil on import. When nil, the function
// populates all API-returned fields unconditionally (no prior intent to preserve).
func populateEsqlControlFromAPI(pm *panelModel, tfPanel *panelModel, apiConfig kbapi.KbnDashboardPanelTypeEsqlControl_Config) {
	api := esqlControlAPIDataFromConfig(apiConfig)
	if !api.ok {
		return
	}
	existing := pm.EsqlControlConfig

	// On import (tfPanel == nil) there is no prior intent — populate from API.
	if tfPanel == nil {
		singleSelect := types.BoolNull()
		if api.SingleSelect != nil {
			singleSelect = types.BoolValue(*api.SingleSelect)
		}
		existing = &esqlControlConfigModel{
			SelectedOptions:  stringsToList(api.SelectedOptions),
			VariableName:     types.StringValue(api.VariableName),
			VariableType:     types.StringValue(api.VariableType),
			EsqlQuery:        types.StringValue(api.EsqlQuery),
			ControlType:      types.StringValue(api.ControlType),
			Title:            types.StringPointerValue(api.Title),
			SingleSelect:     singleSelect,
			AvailableOptions: types.ListNull(types.StringType),
		}
		pm.EsqlControlConfig = existing
		if len(api.AvailableOptions) > 0 {
			existing.AvailableOptions = stringsToList(api.AvailableOptions)
		}
		if api.DisplaySettings != nil {
			d := api.DisplaySettings
			existing.DisplaySettings = &esqlControlDisplaySettingsModel{
				Placeholder:   types.StringPointerValue(d.Placeholder),
				HideActionBar: types.BoolPointerValue(d.HideActionBar),
				HideExclude:   types.BoolPointerValue(d.HideExclude),
				HideExists:    types.BoolPointerValue(d.HideExists),
				HideSort:      types.BoolPointerValue(d.HideSort),
			}
		}
		return
	}

	// If the existing state has no config block, preserve nil intent.
	if existing == nil {
		return
	}

	prevQuery := existing.EsqlQuery
	prevTitle := existing.Title
	prevAvailableOptions := existing.AvailableOptions

	// Required fields always get updated from API.
	existing.SelectedOptions = stringsToList(api.SelectedOptions)
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
		existing.AvailableOptions = stringsToList(api.AvailableOptions)
	}
	preserveKnownStringIfStateBlank(prevQuery, &existing.EsqlQuery)
	preserveKnownStringIfStateBlank(prevTitle, &existing.Title)
	preserveKnownListIfStateNull(prevAvailableOptions, &existing.AvailableOptions)

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
}

// buildEsqlControlConfig writes the TF model fields into the API panel struct.
func buildEsqlControlConfig(pm panelModel, esqlPanel *kbapi.KbnDashboardPanelTypeEsqlControl) {
	cfg := pm.EsqlControlConfig
	if cfg == nil {
		return
	}

	displayToAPI := func(ds *esqlControlDisplaySettingsModel) *esqlControlDisplaySettingsAPI {
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
	if kbapi.KbnControlsSchemasOptionsListEsqlControlSchemaValuesFromQueryControlType(ct) == kbapi.VALUESFROMQUERY {
		vq := kbapi.KbnControlsSchemasOptionsListEsqlControlSchemaValuesFromQuery{
			SelectedOptions: listToStrings(cfg.SelectedOptions),
			VariableName:    cfg.VariableName.ValueString(),
			VariableType: kbapi.KbnControlsSchemasOptionsListEsqlControlSchemaValuesFromQueryVariableType(
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
		_ = esqlPanel.Config.FromKbnControlsSchemasOptionsListEsqlControlSchemaValuesFromQuery(vq)
		return
	}

	sv := kbapi.KbnControlsSchemasOptionsListEsqlControlSchemaStaticValues{
		SelectedOptions: listToStrings(cfg.SelectedOptions),
		VariableName:    cfg.VariableName.ValueString(),
		VariableType: kbapi.KbnControlsSchemasOptionsListEsqlControlSchemaStaticValuesVariableType(
			cfg.VariableType.ValueString(),
		),
		ControlType: kbapi.KbnControlsSchemasOptionsListEsqlControlSchemaStaticValuesControlType(ct),
	}
	if typeutils.IsKnown(cfg.Title) {
		sv.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.SingleSelect) {
		sv.SingleSelect = cfg.SingleSelect.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.AvailableOptions) {
		sv.AvailableOptions = listToStrings(cfg.AvailableOptions)
	}
	sv.DisplaySettings = displayToAPI(cfg.DisplaySettings)
	_ = esqlPanel.Config.FromKbnControlsSchemasOptionsListEsqlControlSchemaStaticValues(sv)
}
