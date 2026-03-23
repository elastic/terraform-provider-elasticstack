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
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	Title            types.String                     `tfsdk:"title"`
	DisplaySettings  *esqlControlDisplaySettingsModel `tfsdk:"display_settings"`
	SingleSelect     types.Bool                       `tfsdk:"single_select"`
	SelectedOptions  types.List                       `tfsdk:"selected_options"`
	VariableName     types.String                     `tfsdk:"variable_name"`
	VariableType     types.String                     `tfsdk:"variable_type"`
	EsqlQuery        types.String                     `tfsdk:"esql_query"`
	ControlType      types.String                     `tfsdk:"control_type"`
	AvailableOptions types.List                       `tfsdk:"available_options"`
}

func (m *esqlControlConfigModel) fromAPI(ctx context.Context, api kbapi.KbnDashboardPanelEsqlControl_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.VariableName = types.StringValue(api.VariableName)
	m.VariableType = types.StringValue(string(api.VariableType))
	m.EsqlQuery = types.StringValue(api.EsqlQuery)
	m.ControlType = types.StringValue(string(api.ControlType))
	m.SingleSelect = types.BoolPointerValue(api.SingleSelect)

	m.SelectedOptions = typeutils.SliceToListTypeString(ctx, api.SelectedOptions, path.Empty(), &diags)

	if api.AvailableOptions != nil {
		m.AvailableOptions = typeutils.SliceToListTypeString(ctx, *api.AvailableOptions, path.Empty(), &diags)
	} else {
		m.AvailableOptions = types.ListNull(types.StringType)
	}

	if api.DisplaySettings != nil {
		ds := &esqlControlDisplaySettingsModel{
			Placeholder:   types.StringPointerValue(api.DisplaySettings.Placeholder),
			HideActionBar: types.BoolPointerValue(api.DisplaySettings.HideActionBar),
			HideExclude:   types.BoolPointerValue(api.DisplaySettings.HideExclude),
			HideExists:    types.BoolPointerValue(api.DisplaySettings.HideExists),
			HideSort:      types.BoolPointerValue(api.DisplaySettings.HideSort),
		}
		if typeutils.IsKnown(ds.Placeholder) || typeutils.IsKnown(ds.HideActionBar) ||
			typeutils.IsKnown(ds.HideExclude) || typeutils.IsKnown(ds.HideExists) || typeutils.IsKnown(ds.HideSort) {
			m.DisplaySettings = ds
		}
	}

	return diags
}

func (m *esqlControlConfigModel) toAPI(ctx context.Context) (kbapi.KbnDashboardPanelEsqlControl_Config, diag.Diagnostics) {
	var diags diag.Diagnostics

	selected := typeutils.ListTypeToSliceString(ctx, m.SelectedOptions, path.Empty(), &diags)
	if selected == nil {
		selected = []string{}
	}

	cfgMap := map[string]any{
		"variable_name":    m.VariableName.ValueString(),
		"variable_type":    m.VariableType.ValueString(),
		"esql_query":       m.EsqlQuery.ValueString(),
		"control_type":     m.ControlType.ValueString(),
		"selected_options": selected,
	}

	if typeutils.IsKnown(m.Title) {
		cfgMap["title"] = m.Title.ValueString()
	}
	if typeutils.IsKnown(m.SingleSelect) {
		cfgMap["single_select"] = m.SingleSelect.ValueBool()
	}
	if typeutils.IsKnown(m.AvailableOptions) {
		avail := typeutils.ListTypeToSliceString(ctx, m.AvailableOptions, path.Empty(), &diags)
		if len(avail) > 0 {
			cfgMap["available_options"] = avail
		}
	}

	if m.DisplaySettings != nil {
		dsMap := map[string]any{}
		if typeutils.IsKnown(m.DisplaySettings.Placeholder) {
			dsMap["placeholder"] = m.DisplaySettings.Placeholder.ValueString()
		}
		if typeutils.IsKnown(m.DisplaySettings.HideActionBar) {
			dsMap["hide_action_bar"] = m.DisplaySettings.HideActionBar.ValueBool()
		}
		if typeutils.IsKnown(m.DisplaySettings.HideExclude) {
			dsMap["hide_exclude"] = m.DisplaySettings.HideExclude.ValueBool()
		}
		if typeutils.IsKnown(m.DisplaySettings.HideExists) {
			dsMap["hide_exists"] = m.DisplaySettings.HideExists.ValueBool()
		}
		if typeutils.IsKnown(m.DisplaySettings.HideSort) {
			dsMap["hide_sort"] = m.DisplaySettings.HideSort.ValueBool()
		}
		if len(dsMap) > 0 {
			cfgMap["display_settings"] = dsMap
		}
	}

	raw, err := json.Marshal(cfgMap)
	if err != nil {
		diags.AddError("Failed to marshal ES|QL control config", err.Error())
		return kbapi.KbnDashboardPanelEsqlControl_Config{}, diags
	}
	var out kbapi.KbnDashboardPanelEsqlControl_Config
	if err := out.UnmarshalJSON(raw); err != nil {
		diags.AddError("Failed to parse ES|QL control config", err.Error())
		return kbapi.KbnDashboardPanelEsqlControl_Config{}, diags
	}
	return out, diags
}

// mergeOptionalEsqlControlFromPrior copies optional fields from a prior Terraform model when
// the API omitted them on read (mirrors the markdown panel seeding pattern in mapPanelFromAPI).
func (m *esqlControlConfigModel) mergeOptionalEsqlControlFromPrior(from *esqlControlConfigModel) {
	if m == nil || from == nil {
		return
	}
	if !typeutils.IsKnown(m.Title) && typeutils.IsKnown(from.Title) {
		m.Title = from.Title
	}
	if !typeutils.IsKnown(m.SingleSelect) && typeutils.IsKnown(from.SingleSelect) {
		m.SingleSelect = from.SingleSelect
	}
	if (!typeutils.IsKnown(m.AvailableOptions) || m.AvailableOptions.IsNull()) && typeutils.IsKnown(from.AvailableOptions) && !from.AvailableOptions.IsNull() {
		m.AvailableOptions = from.AvailableOptions
	}
	switch {
	case m.DisplaySettings == nil && from.DisplaySettings != nil:
		m.DisplaySettings = cloneEsqlControlDisplaySettings(from.DisplaySettings)
	case m.DisplaySettings != nil && from.DisplaySettings != nil:
		mergeEsqlDisplaySettingsOptional(m.DisplaySettings, from.DisplaySettings)
	}
}

func cloneEsqlControlDisplaySettings(from *esqlControlDisplaySettingsModel) *esqlControlDisplaySettingsModel {
	if from == nil {
		return nil
	}
	return &esqlControlDisplaySettingsModel{
		Placeholder:   from.Placeholder,
		HideActionBar: from.HideActionBar,
		HideExclude:   from.HideExclude,
		HideExists:    from.HideExists,
		HideSort:      from.HideSort,
	}
}

func mergeEsqlDisplaySettingsOptional(to, from *esqlControlDisplaySettingsModel) {
	if !typeutils.IsKnown(to.Placeholder) && typeutils.IsKnown(from.Placeholder) {
		to.Placeholder = from.Placeholder
	}
	if !typeutils.IsKnown(to.HideActionBar) && typeutils.IsKnown(from.HideActionBar) {
		to.HideActionBar = from.HideActionBar
	}
	if !typeutils.IsKnown(to.HideExclude) && typeutils.IsKnown(from.HideExclude) {
		to.HideExclude = from.HideExclude
	}
	if !typeutils.IsKnown(to.HideExists) && typeutils.IsKnown(from.HideExists) {
		to.HideExists = from.HideExists
	}
	if !typeutils.IsKnown(to.HideSort) && typeutils.IsKnown(from.HideSort) {
		to.HideSort = from.HideSort
	}
}
