package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (m *dashboardModel) optionsToAPI(ctx context.Context) (*optionsAPIModel, diag.Diagnostics) {
	if !utils.IsKnown(m.Options) {
		return nil, nil
	}

	var optModel optionsModel
	diags := m.Options.As(ctx, &optModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	return optModel.toAPI(), diags
}

// optionsAPIModel introduces a type alias for the generated API model.
// The current API spec defines these types inline, resulting in anonymous structs.
// A new type definition won't exactly match the API struct, howeven an alias will.
type optionsAPIModel = struct {
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

func (m *dashboardModel) mapOptionsFromAPI(ctx context.Context, options *optionsAPIModel) (types.Object, diag.Diagnostics) {
	if options == nil {
		return types.ObjectNull(getOptionsAttrTypes()), nil
	}

	model := optionsModel{
		HidePanelTitles: types.BoolPointerValue(options.HidePanelTitles),
		UseMargins:      types.BoolPointerValue(options.UseMargins),
		SyncColors:      types.BoolPointerValue(options.SyncColors),
		SyncTooltips:    types.BoolPointerValue(options.SyncTooltips),
		SyncCursor:      types.BoolPointerValue(options.SyncCursor),
	}

	return types.ObjectValueFrom(ctx, getOptionsAttrTypes(), model)
}

func (m optionsModel) toAPI() *optionsAPIModel {
	options := optionsAPIModel{}
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
