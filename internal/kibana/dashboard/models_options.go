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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type optionsModel struct {
	HidePanelTitles types.Bool `tfsdk:"hide_panel_titles"`
	UseMargins      types.Bool `tfsdk:"use_margins"`
	SyncColors      types.Bool `tfsdk:"sync_colors"`
	SyncTooltips    types.Bool `tfsdk:"sync_tooltips"`
	SyncCursor      types.Bool `tfsdk:"sync_cursor"`
}

func (m *dashboardModel) optionsToAPI() (*optionsAPIModel, diag.Diagnostics) {
	if m.Options == nil {
		return nil, diag.Diagnostics{}
	}

	return m.Options.toAPI(), diag.Diagnostics{}
}

// optionsAPIModel introduces a type alias for the generated API model.
// The current API spec defines these types inline, resulting in anonymous structs.
// A new type definition won't exactly match the API struct, however an alias will.
type optionsAPIModel = struct {
	// AutoApplyFilters Auto apply control filters.
	AutoApplyFilters *bool `json:"auto_apply_filters,omitempty"`

	// HidePanelBorders Hide the panel borders in the dashboard.
	HidePanelBorders *bool `json:"hide_panel_borders,omitempty"`

	// HidePanelTitles Hide the panel titles in the dashboard.
	HidePanelTitles *bool `json:"hide_panel_titles,omitempty"`

	// SyncColors Synchronize colors between related panels in the dashboard.
	SyncColors *bool `json:"sync_colors,omitempty"`

	// SyncCursor Synchronize cursor position between related panels in the dashboard.
	SyncCursor *bool `json:"sync_cursor,omitempty"`

	// SyncTooltips Synchronize tooltips between related panels in the dashboard.
	SyncTooltips *bool `json:"sync_tooltips,omitempty"`

	// UseMargins Show margins between panels in the dashboard layout.
	UseMargins *bool `json:"use_margins,omitempty"`
}

func (m *dashboardModel) mapOptionsFromAPI(options *optionsAPIModel) *optionsModel {
	if options == nil {
		return nil
	}

	model := optionsModel{
		HidePanelTitles: types.BoolPointerValue(options.HidePanelTitles),
		UseMargins:      types.BoolPointerValue(options.UseMargins),
		SyncColors:      types.BoolPointerValue(options.SyncColors),
		SyncTooltips:    types.BoolPointerValue(options.SyncTooltips),
		SyncCursor:      types.BoolPointerValue(options.SyncCursor),
	}

	return &model
}

func (m optionsModel) toAPI() *optionsAPIModel {
	options := optionsAPIModel{}
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

	return &options
}
