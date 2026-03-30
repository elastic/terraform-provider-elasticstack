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
	SelectedOptions  types.List                        `tfsdk:"selected_options"`
	VariableName     types.String                      `tfsdk:"variable_name"`
	VariableType     types.String                      `tfsdk:"variable_type"`
	EsqlQuery        types.String                      `tfsdk:"esql_query"`
	ControlType      types.String                      `tfsdk:"control_type"`
	Title            types.String                      `tfsdk:"title"`
	SingleSelect     types.Bool                        `tfsdk:"single_select"`
	AvailableOptions types.List                        `tfsdk:"available_options"`
	DisplaySettings  *esqlControlDisplaySettingsModel  `tfsdk:"display_settings"`
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
func populateEsqlControlFromAPI(pm *panelModel, tfPanel *panelModel, apiConfig kbapi.KbnDashboardPanelEsqlControl_Config) {
	existing := pm.EsqlControlConfig

	// On import (tfPanel == nil) there is no prior intent — populate from API.
	if tfPanel == nil {
		existing = &esqlControlConfigModel{
			SelectedOptions: stringsToList(apiConfig.SelectedOptions),
			VariableName:    types.StringValue(apiConfig.VariableName),
			VariableType:    types.StringValue(string(apiConfig.VariableType)),
			EsqlQuery:       types.StringValue(apiConfig.EsqlQuery),
			ControlType:     types.StringValue(string(apiConfig.ControlType)),
			AvailableOptions: types.ListNull(types.StringType),
		}
		pm.EsqlControlConfig = existing
		if apiConfig.Title != nil {
			existing.Title = types.StringValue(*apiConfig.Title)
		}
		if apiConfig.SingleSelect != nil {
			existing.SingleSelect = types.BoolValue(*apiConfig.SingleSelect)
		}
		if apiConfig.AvailableOptions != nil {
			existing.AvailableOptions = stringsToList(*apiConfig.AvailableOptions)
		}
		if apiConfig.DisplaySettings != nil {
			d := apiConfig.DisplaySettings
			m := &esqlControlDisplaySettingsModel{}
			if d.Placeholder != nil {
				m.Placeholder = types.StringValue(*d.Placeholder)
			}
			if d.HideActionBar != nil {
				m.HideActionBar = types.BoolValue(*d.HideActionBar)
			}
			if d.HideExclude != nil {
				m.HideExclude = types.BoolValue(*d.HideExclude)
			}
			if d.HideExists != nil {
				m.HideExists = types.BoolValue(*d.HideExists)
			}
			if d.HideSort != nil {
				m.HideSort = types.BoolValue(*d.HideSort)
			}
			existing.DisplaySettings = m
		}
		return
	}

	// If the existing state has no config block, preserve nil intent.
	if existing == nil {
		return
	}

	// Required fields always get updated from API.
	existing.SelectedOptions = stringsToList(apiConfig.SelectedOptions)
	existing.VariableName = types.StringValue(apiConfig.VariableName)
	existing.VariableType = types.StringValue(string(apiConfig.VariableType))
	existing.EsqlQuery = types.StringValue(apiConfig.EsqlQuery)
	existing.ControlType = types.StringValue(string(apiConfig.ControlType))

	// Optional fields: update only if already known (non-null) in state.
	if typeutils.IsKnown(existing.Title) && apiConfig.Title != nil {
		existing.Title = types.StringValue(*apiConfig.Title)
	}
	if typeutils.IsKnown(existing.SingleSelect) && apiConfig.SingleSelect != nil {
		existing.SingleSelect = types.BoolValue(*apiConfig.SingleSelect)
	}

	// available_options: if TF state had it set (non-null list), update from API.
	if !existing.AvailableOptions.IsNull() && apiConfig.AvailableOptions != nil {
		existing.AvailableOptions = stringsToList(*apiConfig.AvailableOptions)
	}

	// display_settings: if block is present in state, update from API; otherwise preserve nil.
	if existing.DisplaySettings != nil && apiConfig.DisplaySettings != nil {
		ds := existing.DisplaySettings
		if typeutils.IsKnown(ds.Placeholder) && apiConfig.DisplaySettings.Placeholder != nil {
			ds.Placeholder = types.StringValue(*apiConfig.DisplaySettings.Placeholder)
		}
		if typeutils.IsKnown(ds.HideActionBar) && apiConfig.DisplaySettings.HideActionBar != nil {
			ds.HideActionBar = types.BoolValue(*apiConfig.DisplaySettings.HideActionBar)
		}
		if typeutils.IsKnown(ds.HideExclude) && apiConfig.DisplaySettings.HideExclude != nil {
			ds.HideExclude = types.BoolValue(*apiConfig.DisplaySettings.HideExclude)
		}
		if typeutils.IsKnown(ds.HideExists) && apiConfig.DisplaySettings.HideExists != nil {
			ds.HideExists = types.BoolValue(*apiConfig.DisplaySettings.HideExists)
		}
		if typeutils.IsKnown(ds.HideSort) && apiConfig.DisplaySettings.HideSort != nil {
			ds.HideSort = types.BoolValue(*apiConfig.DisplaySettings.HideSort)
		}
	}
}

// buildEsqlControlConfig writes the TF model fields into the API panel struct.
func buildEsqlControlConfig(pm panelModel, esqlPanel *kbapi.KbnDashboardPanelEsqlControl) {
	cfg := pm.EsqlControlConfig
	if cfg == nil {
		return
	}

	esqlPanel.Config.SelectedOptions = listToStrings(cfg.SelectedOptions)
	esqlPanel.Config.VariableName = cfg.VariableName.ValueString()
	esqlPanel.Config.VariableType = kbapi.KbnDashboardPanelEsqlControlConfigVariableType(cfg.VariableType.ValueString())
	esqlPanel.Config.EsqlQuery = cfg.EsqlQuery.ValueString()
	esqlPanel.Config.ControlType = kbapi.KbnDashboardPanelEsqlControlConfigControlType(cfg.ControlType.ValueString())

	if typeutils.IsKnown(cfg.Title) {
		esqlPanel.Config.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.SingleSelect) {
		esqlPanel.Config.SingleSelect = cfg.SingleSelect.ValueBoolPointer()
	}
	if !cfg.AvailableOptions.IsNull() {
		opts := listToStrings(cfg.AvailableOptions)
		esqlPanel.Config.AvailableOptions = &opts
	}
	if cfg.DisplaySettings != nil {
		ds := cfg.DisplaySettings
		esqlPanel.Config.DisplaySettings = &struct {
			HideActionBar *bool   `json:"hide_action_bar,omitempty"`
			HideExclude   *bool   `json:"hide_exclude,omitempty"`
			HideExists    *bool   `json:"hide_exists,omitempty"`
			HideSort      *bool   `json:"hide_sort,omitempty"`
			Placeholder   *string `json:"placeholder,omitempty"`
		}{}
		if typeutils.IsKnown(ds.Placeholder) {
			esqlPanel.Config.DisplaySettings.Placeholder = ds.Placeholder.ValueStringPointer()
		}
		if typeutils.IsKnown(ds.HideActionBar) {
			esqlPanel.Config.DisplaySettings.HideActionBar = ds.HideActionBar.ValueBoolPointer()
		}
		if typeutils.IsKnown(ds.HideExclude) {
			esqlPanel.Config.DisplaySettings.HideExclude = ds.HideExclude.ValueBoolPointer()
		}
		if typeutils.IsKnown(ds.HideExists) {
			esqlPanel.Config.DisplaySettings.HideExists = ds.HideExists.ValueBoolPointer()
		}
		if typeutils.IsKnown(ds.HideSort) {
			esqlPanel.Config.DisplaySettings.HideSort = ds.HideSort.ValueBoolPointer()
		}
	}
}
