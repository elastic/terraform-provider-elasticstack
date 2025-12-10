package dashboard

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_dashboardModel_optionsToAPI(t *testing.T) {
	tests := []struct {
		name      string
		options   types.Object
		want      *optionsAPIModel
		wantErr   bool
		wantDiags bool
	}{
		{
			name:    "returns nil when options is null",
			options: types.ObjectNull(getOptionsAttrTypes()),
			want:    nil,
		},
		{
			name:    "returns nil when options is unknown",
			options: types.ObjectUnknown(getOptionsAttrTypes()),
			want:    nil,
		},
		{
			name: "converts all fields when set to true",
			options: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolValue(true),
					"use_margins":       types.BoolValue(true),
					"sync_colors":       types.BoolValue(true),
					"sync_tooltips":     types.BoolValue(true),
					"sync_cursor":       types.BoolValue(true),
				},
			),
			want: &optionsAPIModel{
				HidePanelTitles: utils.Pointer(true),
				UseMargins:      utils.Pointer(true),
				SyncColors:      utils.Pointer(true),
				SyncTooltips:    utils.Pointer(true),
				SyncCursor:      utils.Pointer(true),
			},
		},
		{
			name: "converts all fields when set to false",
			options: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolValue(false),
					"use_margins":       types.BoolValue(false),
					"sync_colors":       types.BoolValue(false),
					"sync_tooltips":     types.BoolValue(false),
					"sync_cursor":       types.BoolValue(false),
				},
			),
			want: &optionsAPIModel{
				HidePanelTitles: utils.Pointer(false),
				UseMargins:      utils.Pointer(false),
				SyncColors:      utils.Pointer(false),
				SyncTooltips:    utils.Pointer(false),
				SyncCursor:      utils.Pointer(false),
			},
		},
		{
			name: "handles mixed null and set values",
			options: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolValue(true),
					"use_margins":       types.BoolNull(),
					"sync_colors":       types.BoolValue(false),
					"sync_tooltips":     types.BoolNull(),
					"sync_cursor":       types.BoolValue(true),
				},
			),
			want: &optionsAPIModel{
				HidePanelTitles: utils.Pointer(true),
				UseMargins:      nil,
				SyncColors:      utils.Pointer(false),
				SyncTooltips:    nil,
				SyncCursor:      utils.Pointer(true),
			},
		},
		{
			name: "handles mixed unknown and set values",
			options: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolUnknown(),
					"use_margins":       types.BoolValue(true),
					"sync_colors":       types.BoolUnknown(),
					"sync_tooltips":     types.BoolValue(false),
					"sync_cursor":       types.BoolUnknown(),
				},
			),
			want: &optionsAPIModel{
				HidePanelTitles: nil,
				UseMargins:      utils.Pointer(true),
				SyncColors:      nil,
				SyncTooltips:    utils.Pointer(false),
				SyncCursor:      nil,
			},
		},
		{
			name: "handles all null values",
			options: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolNull(),
					"use_margins":       types.BoolNull(),
					"sync_colors":       types.BoolNull(),
					"sync_tooltips":     types.BoolNull(),
					"sync_cursor":       types.BoolNull(),
				},
			),
			want: &optionsAPIModel{
				HidePanelTitles: nil,
				UseMargins:      nil,
				SyncColors:      nil,
				SyncTooltips:    nil,
				SyncCursor:      nil,
			},
		},
		{
			name: "handles all unknown values",
			options: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolUnknown(),
					"use_margins":       types.BoolUnknown(),
					"sync_colors":       types.BoolUnknown(),
					"sync_tooltips":     types.BoolUnknown(),
					"sync_cursor":       types.BoolUnknown(),
				},
			),
			want: &optionsAPIModel{
				HidePanelTitles: nil,
				UseMargins:      nil,
				SyncColors:      nil,
				SyncTooltips:    nil,
				SyncCursor:      nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &dashboardModel{
				Options: tt.options,
			}
			got, diags := m.optionsToAPI(context.Background())

			if tt.wantDiags {
				assert.True(t, diags.HasError(), "expected diagnostics but got none")
			} else {
				assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_dashboardModel_mapOptionsFromAPI(t *testing.T) {
	tests := []struct {
		name      string
		options   *optionsAPIModel
		want      types.Object
		wantErr   bool
		wantDiags bool
	}{
		{
			name:    "returns null object when options is nil",
			options: nil,
			want:    types.ObjectNull(getOptionsAttrTypes()),
		},
		{
			name: "converts all fields when set to true",
			options: &optionsAPIModel{
				HidePanelTitles: utils.Pointer(true),
				UseMargins:      utils.Pointer(true),
				SyncColors:      utils.Pointer(true),
				SyncTooltips:    utils.Pointer(true),
				SyncCursor:      utils.Pointer(true),
			},
			want: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolValue(true),
					"use_margins":       types.BoolValue(true),
					"sync_colors":       types.BoolValue(true),
					"sync_tooltips":     types.BoolValue(true),
					"sync_cursor":       types.BoolValue(true),
				},
			),
		},
		{
			name: "converts all fields when set to false",
			options: &optionsAPIModel{
				HidePanelTitles: utils.Pointer(false),
				UseMargins:      utils.Pointer(false),
				SyncColors:      utils.Pointer(false),
				SyncTooltips:    utils.Pointer(false),
				SyncCursor:      utils.Pointer(false),
			},
			want: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolValue(false),
					"use_margins":       types.BoolValue(false),
					"sync_colors":       types.BoolValue(false),
					"sync_tooltips":     types.BoolValue(false),
					"sync_cursor":       types.BoolValue(false),
				},
			),
		},
		{
			name: "handles mixed nil and set values",
			options: &optionsAPIModel{
				HidePanelTitles: utils.Pointer(true),
				UseMargins:      nil,
				SyncColors:      utils.Pointer(false),
				SyncTooltips:    nil,
				SyncCursor:      utils.Pointer(true),
			},
			want: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolValue(true),
					"use_margins":       types.BoolNull(),
					"sync_colors":       types.BoolValue(false),
					"sync_tooltips":     types.BoolNull(),
					"sync_cursor":       types.BoolValue(true),
				},
			),
		},
		{
			name: "handles all nil values",
			options: &optionsAPIModel{
				HidePanelTitles: nil,
				UseMargins:      nil,
				SyncColors:      nil,
				SyncTooltips:    nil,
				SyncCursor:      nil,
			},
			want: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolNull(),
					"use_margins":       types.BoolNull(),
					"sync_colors":       types.BoolNull(),
					"sync_tooltips":     types.BoolNull(),
					"sync_cursor":       types.BoolNull(),
				},
			),
		},
		{
			name:    "handles empty struct",
			options: &optionsAPIModel{},
			want: types.ObjectValueMust(
				getOptionsAttrTypes(),
				map[string]attr.Value{
					"hide_panel_titles": types.BoolNull(),
					"use_margins":       types.BoolNull(),
					"sync_colors":       types.BoolNull(),
					"sync_tooltips":     types.BoolNull(),
					"sync_cursor":       types.BoolNull(),
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &dashboardModel{}
			got, diags := m.mapOptionsFromAPI(context.Background(), tt.options)

			if tt.wantDiags {
				assert.True(t, diags.HasError(), "expected diagnostics but got none")
			} else {
				assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_optionsModel_toAPI(t *testing.T) {
	tests := []struct {
		name  string
		model optionsModel
		want  *optionsAPIModel
	}{
		{
			name: "converts all fields when set to true",
			model: optionsModel{
				HidePanelTitles: types.BoolValue(true),
				UseMargins:      types.BoolValue(true),
				SyncColors:      types.BoolValue(true),
				SyncTooltips:    types.BoolValue(true),
				SyncCursor:      types.BoolValue(true),
			},
			want: &optionsAPIModel{
				HidePanelTitles: utils.Pointer(true),
				UseMargins:      utils.Pointer(true),
				SyncColors:      utils.Pointer(true),
				SyncTooltips:    utils.Pointer(true),
				SyncCursor:      utils.Pointer(true),
			},
		},
		{
			name: "converts all fields when set to false",
			model: optionsModel{
				HidePanelTitles: types.BoolValue(false),
				UseMargins:      types.BoolValue(false),
				SyncColors:      types.BoolValue(false),
				SyncTooltips:    types.BoolValue(false),
				SyncCursor:      types.BoolValue(false),
			},
			want: &optionsAPIModel{
				HidePanelTitles: utils.Pointer(false),
				UseMargins:      utils.Pointer(false),
				SyncColors:      utils.Pointer(false),
				SyncTooltips:    utils.Pointer(false),
				SyncCursor:      utils.Pointer(false),
			},
		},
		{
			name: "handles mixed null and set values",
			model: optionsModel{
				HidePanelTitles: types.BoolValue(true),
				UseMargins:      types.BoolNull(),
				SyncColors:      types.BoolValue(false),
				SyncTooltips:    types.BoolNull(),
				SyncCursor:      types.BoolValue(true),
			},
			want: &optionsAPIModel{
				HidePanelTitles: utils.Pointer(true),
				UseMargins:      nil,
				SyncColors:      utils.Pointer(false),
				SyncTooltips:    nil,
				SyncCursor:      utils.Pointer(true),
			},
		},
		{
			name: "handles mixed unknown and set values",
			model: optionsModel{
				HidePanelTitles: types.BoolUnknown(),
				UseMargins:      types.BoolValue(true),
				SyncColors:      types.BoolUnknown(),
				SyncTooltips:    types.BoolValue(false),
				SyncCursor:      types.BoolUnknown(),
			},
			want: &optionsAPIModel{
				HidePanelTitles: nil,
				UseMargins:      utils.Pointer(true),
				SyncColors:      nil,
				SyncTooltips:    utils.Pointer(false),
				SyncCursor:      nil,
			},
		},
		{
			name: "handles all null values",
			model: optionsModel{
				HidePanelTitles: types.BoolNull(),
				UseMargins:      types.BoolNull(),
				SyncColors:      types.BoolNull(),
				SyncTooltips:    types.BoolNull(),
				SyncCursor:      types.BoolNull(),
			},
			want: &optionsAPIModel{
				HidePanelTitles: nil,
				UseMargins:      nil,
				SyncColors:      nil,
				SyncTooltips:    nil,
				SyncCursor:      nil,
			},
		},
		{
			name: "handles all unknown values",
			model: optionsModel{
				HidePanelTitles: types.BoolUnknown(),
				UseMargins:      types.BoolUnknown(),
				SyncColors:      types.BoolUnknown(),
				SyncTooltips:    types.BoolUnknown(),
				SyncCursor:      types.BoolUnknown(),
			},
			want: &optionsAPIModel{
				HidePanelTitles: nil,
				UseMargins:      nil,
				SyncColors:      nil,
				SyncTooltips:    nil,
				SyncCursor:      nil,
			},
		},
		{
			name:  "handles zero-value model",
			model: optionsModel{},
			want: &optionsAPIModel{
				HidePanelTitles: nil,
				UseMargins:      nil,
				SyncColors:      nil,
				SyncTooltips:    nil,
				SyncCursor:      nil,
			},
		},
		{
			name: "handles single field set",
			model: optionsModel{
				HidePanelTitles: types.BoolValue(true),
				UseMargins:      types.BoolNull(),
				SyncColors:      types.BoolNull(),
				SyncTooltips:    types.BoolNull(),
				SyncCursor:      types.BoolNull(),
			},
			want: &optionsAPIModel{
				HidePanelTitles: utils.Pointer(true),
				UseMargins:      nil,
				SyncColors:      nil,
				SyncTooltips:    nil,
				SyncCursor:      nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.model.toAPI()
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_optionsModel_roundTrip(t *testing.T) {
	// Test round-trip conversion: model -> API -> model
	tests := []struct {
		name  string
		model optionsModel
	}{
		{
			name: "all true values",
			model: optionsModel{
				HidePanelTitles: types.BoolValue(true),
				UseMargins:      types.BoolValue(true),
				SyncColors:      types.BoolValue(true),
				SyncTooltips:    types.BoolValue(true),
				SyncCursor:      types.BoolValue(true),
			},
		},
		{
			name: "all false values",
			model: optionsModel{
				HidePanelTitles: types.BoolValue(false),
				UseMargins:      types.BoolValue(false),
				SyncColors:      types.BoolValue(false),
				SyncTooltips:    types.BoolValue(false),
				SyncCursor:      types.BoolValue(false),
			},
		},
		{
			name: "mixed values",
			model: optionsModel{
				HidePanelTitles: types.BoolValue(true),
				UseMargins:      types.BoolValue(false),
				SyncColors:      types.BoolValue(true),
				SyncTooltips:    types.BoolValue(false),
				SyncCursor:      types.BoolValue(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to API model
			apiModel := tt.model.toAPI()
			require.NotNil(t, apiModel)

			// Convert back to Terraform model
			dm := &dashboardModel{}
			obj, diags := dm.mapOptionsFromAPI(context.Background(), apiModel)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			// Extract the model from the object
			var roundTripModel optionsModel
			diags = obj.As(context.Background(), &roundTripModel, basetypes.ObjectAsOptions{})
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			// Compare the original and round-trip models
			assert.Equal(t, tt.model.HidePanelTitles, roundTripModel.HidePanelTitles)
			assert.Equal(t, tt.model.UseMargins, roundTripModel.UseMargins)
			assert.Equal(t, tt.model.SyncColors, roundTripModel.SyncColors)
			assert.Equal(t, tt.model.SyncTooltips, roundTripModel.SyncTooltips)
			assert.Equal(t, tt.model.SyncCursor, roundTripModel.SyncCursor)
		})
	}
}
