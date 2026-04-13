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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type optionsModel struct {
	HidePanelTitles  types.Bool `tfsdk:"hide_panel_titles"`
	UseMargins       types.Bool `tfsdk:"use_margins"`
	SyncColors       types.Bool `tfsdk:"sync_colors"`
	SyncTooltips     types.Bool `tfsdk:"sync_tooltips"`
	SyncCursor       types.Bool `tfsdk:"sync_cursor"`
	AutoApplyFilters types.Bool `tfsdk:"auto_apply_filters"`
	HidePanelBorders types.Bool `tfsdk:"hide_panel_borders"`
}

func (m *dashboardModel) optionsToAPI() (kbapi.KbnDashboardOptions, diag.Diagnostics) {
	if m.Options == nil {
		return kbapi.KbnDashboardOptions{}, diag.Diagnostics{}
	}
	o := m.Options.toAPI()
	if o == nil {
		return kbapi.KbnDashboardOptions{}, diag.Diagnostics{}
	}
	return *o, diag.Diagnostics{}
}

func (m *dashboardModel) mapOptionsFromAPI(options kbapi.KbnDashboardOptions) *optionsModel {
	// Kibana snapshots can materialize dashboard option defaults even when the
	// options block was omitted in Terraform config. Preserve a nil options block
	// in that case to avoid "inconsistent result after apply".
	if m.Options == nil && isDashboardOptionsDefaultSet(&options) {
		return nil
	}

	model := optionsModel{
		HidePanelTitles:  types.BoolPointerValue(options.HidePanelTitles),
		UseMargins:       types.BoolPointerValue(options.UseMargins),
		SyncColors:       types.BoolPointerValue(options.SyncColors),
		SyncTooltips:     types.BoolPointerValue(options.SyncTooltips),
		SyncCursor:       types.BoolPointerValue(options.SyncCursor),
		AutoApplyFilters: types.BoolPointerValue(options.AutoApplyFilters),
		HidePanelBorders: types.BoolPointerValue(options.HidePanelBorders),
	}

	return &model
}

func (m optionsModel) toAPI() *kbapi.KbnDashboardOptions {
	options := kbapi.KbnDashboardOptions{}
	if typeutils.IsKnown(m.HidePanelTitles) {
		options.HidePanelTitles = m.HidePanelTitles.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.UseMargins) {
		options.UseMargins = m.UseMargins.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.SyncColors) {
		options.SyncColors = m.SyncColors.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.SyncTooltips) {
		options.SyncTooltips = m.SyncTooltips.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.SyncCursor) {
		options.SyncCursor = m.SyncCursor.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.AutoApplyFilters) {
		options.AutoApplyFilters = m.AutoApplyFilters.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.HidePanelBorders) {
		options.HidePanelBorders = m.HidePanelBorders.ValueBoolPointer()
	}

	return &options
}

func isDashboardOptionsDefaultSet(options *kbapi.KbnDashboardOptions) bool {
	if options == nil {
		return false
	}

	// OpenAPI examples use auto_apply_filters=true and hide_panel_borders=false as defaults.
	// When those pointers are omitted on GET, treat them as matching defaults so an omitted
	// Terraform `options` block stays null in state (REQ-009).
	return boolPtrEquals(options.HidePanelTitles, false) &&
		boolPtrEquals(options.UseMargins, true) &&
		boolPtrEquals(options.SyncColors, false) &&
		boolPtrEquals(options.SyncTooltips, false) &&
		boolPtrEquals(options.SyncCursor, true) &&
		boolPtrEqualsOrOmitted(options.AutoApplyFilters, true) &&
		boolPtrEqualsOrOmitted(options.HidePanelBorders, false)
}

func boolPtrEqualsOrOmitted(value *bool, expected bool) bool {
	if value == nil {
		return true
	}
	return *value == expected
}

func boolPtrEquals(value *bool, expected bool) bool {
	return value != nil && *value == expected
}
