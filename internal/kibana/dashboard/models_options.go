package dashboard

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type optionsModel struct {
	AutoApplyFilters types.Bool `tfsdk:"auto_apply_filters"`
	HidePanelTitles types.Bool `tfsdk:"hide_panel_titles"`
	UseMargins      types.Bool `tfsdk:"use_margins"`
	SyncColors      types.Bool `tfsdk:"sync_colors"`
	SyncTooltips    types.Bool `tfsdk:"sync_tooltips"`
	SyncCursor      types.Bool `tfsdk:"sync_cursor"`
}

func (m *dashboardModel) optionsToAPI() (*optionsAPIModel, diag.Diagnostics) {
	if m.Options == nil {
		return nil, nil
	}

	return m.Options.toAPI(), nil
}

// optionsAPIModel introduces a type alias for the generated API model.
// The current API spec defines these types inline, resulting in anonymous structs.
// A new type definition won't exactly match the API struct, howeven an alias will.
type optionsAPIModel = struct {
	// AutoApplyFilters Auto apply control filters.
	AutoApplyFilters *bool `json:"auto_apply_filters,omitempty"`

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
		AutoApplyFilters: types.BoolPointerValue(options.AutoApplyFilters),
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
	if utils.IsKnown(m.AutoApplyFilters) {
		options.AutoApplyFilters = m.AutoApplyFilters.ValueBoolPointer()
	}
	if utils.IsKnown(m.HidePanelTitles) {
		options.HidePanelTitles = m.HidePanelTitles.ValueBoolPointer()
	}
	if utils.IsKnown(m.UseMargins) {
		options.UseMargins = m.UseMargins.ValueBoolPointer()
	}
	if utils.IsKnown(m.SyncColors) {
		options.SyncColors = m.SyncColors.ValueBoolPointer()
	}
	if utils.IsKnown(m.SyncTooltips) {
		options.SyncTooltips = m.SyncTooltips.ValueBoolPointer()
	}
	if utils.IsKnown(m.SyncCursor) {
		options.SyncCursor = m.SyncCursor.ValueBoolPointer()
	}

	return &options
}
