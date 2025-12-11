package dashboard

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// optionsAPIModel introduces a type alias for the generated API model.
// The current API spec defines these types inline, resulting in anonymous structs.
// A new type definition won't exactly match the API struct, howeven an alias will.
type optionsAPIModel = struct {
	HidePanelTitles *bool `json:"hidePanelTitles,omitempty"`
	SyncColors      *bool `json:"syncColors,omitempty"`
	SyncCursor      *bool `json:"syncCursor,omitempty"`
	SyncTooltips    *bool `json:"syncTooltips,omitempty"`
	UseMargins      *bool `json:"useMargins,omitempty"`
}

func newOptionsFromAPI(options *optionsAPIModel) *optionsModel {
	if options == nil {
		return nil
	}

	return &optionsModel{
		HidePanelTitles: types.BoolPointerValue(options.HidePanelTitles),
		UseMargins:      types.BoolPointerValue(options.UseMargins),
		SyncColors:      types.BoolPointerValue(options.SyncColors),
		SyncTooltips:    types.BoolPointerValue(options.SyncTooltips),
		SyncCursor:      types.BoolPointerValue(options.SyncCursor),
	}
}

func (m *optionsModel) toAPI() *optionsAPIModel {
	if m == nil {
		return nil
	}

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
