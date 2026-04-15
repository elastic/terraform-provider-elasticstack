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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_dashboardModel_optionsToAPI(t *testing.T) {
	tests := []struct {
		name      string
		options   *optionsModel
		want      *kbapi.KbnDashboardOptions
		wantDiags bool
	}{
		{
			name:    "returns nil when options is nil",
			options: nil,
			want:    nil,
		},
		{
			name: "converts all fields when set to true",
			options: &optionsModel{
				HidePanelTitles:  types.BoolValue(true),
				UseMargins:       types.BoolValue(true),
				SyncColors:       types.BoolValue(true),
				SyncTooltips:     types.BoolValue(true),
				SyncCursor:       types.BoolValue(true),
				AutoApplyFilters: types.BoolValue(true),
				HidePanelBorders: types.BoolValue(true),
			},
			want: &kbapi.KbnDashboardOptions{
				HidePanelTitles:  new(true),
				UseMargins:       new(true),
				SyncColors:       new(true),
				SyncTooltips:     new(true),
				SyncCursor:       new(true),
				AutoApplyFilters: new(true),
				HidePanelBorders: new(true),
			},
		},
		{
			name: "converts all fields when set to false",
			options: &optionsModel{
				HidePanelTitles: types.BoolValue(false),
				UseMargins:      types.BoolValue(false),
				SyncColors:      types.BoolValue(false),
				SyncTooltips:    types.BoolValue(false),
				SyncCursor:      types.BoolValue(false),
			},
			want: &kbapi.KbnDashboardOptions{
				HidePanelTitles: new(false),
				UseMargins:      new(false),
				SyncColors:      new(false),
				SyncTooltips:    new(false),
				SyncCursor:      new(false),
			},
		},
		{
			name: "handles mixed null and set values",
			options: &optionsModel{
				HidePanelTitles: types.BoolValue(true),
				UseMargins:      types.BoolNull(),
				SyncColors:      types.BoolValue(false),
				SyncTooltips:    types.BoolNull(),
				SyncCursor:      types.BoolValue(true),
			},
			want: &kbapi.KbnDashboardOptions{
				HidePanelTitles: new(true),
				UseMargins:      nil,
				SyncColors:      new(false),
				SyncTooltips:    nil,
				SyncCursor:      new(true),
			},
		},
		{
			name: "handles mixed unknown and set values",
			options: &optionsModel{
				HidePanelTitles: types.BoolUnknown(),
				UseMargins:      types.BoolValue(true),
				SyncColors:      types.BoolUnknown(),
				SyncTooltips:    types.BoolValue(false),
				SyncCursor:      types.BoolUnknown(),
			},
			want: &kbapi.KbnDashboardOptions{
				HidePanelTitles: nil,
				UseMargins:      new(true),
				SyncColors:      nil,
				SyncTooltips:    new(false),
				SyncCursor:      nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &dashboardModel{
				Options: tt.options,
			}
			got, diags := m.optionsToAPI()

			if tt.wantDiags {
				assert.True(t, diags.HasError(), "expected diagnostics but got none")
			} else {
				assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			}

			if tt.want == nil {
				assert.Equal(t, kbapi.KbnDashboardOptions{}, got)
			} else {
				assert.Equal(t, *tt.want, got)
			}
		})
	}
}

func Test_dashboardModel_mapOptionsFromAPI(t *testing.T) {
	tests := []struct {
		name         string
		stateOptions *optionsModel
		options      *kbapi.KbnDashboardOptions
		want         *optionsModel
	}{
		{
			name: "keeps options nil for API defaults when state omitted options",
			options: &kbapi.KbnDashboardOptions{
				HidePanelTitles:  new(false),
				UseMargins:       new(true),
				SyncColors:       new(false),
				SyncTooltips:     new(false),
				SyncCursor:       new(true),
				AutoApplyFilters: new(true),
				HidePanelBorders: new(false),
			},
			want: nil,
		},
		{
			name: "maps API defaults when state already has options",
			stateOptions: &optionsModel{
				HidePanelTitles:  types.BoolValue(false),
				UseMargins:       types.BoolValue(true),
				SyncColors:       types.BoolValue(false),
				SyncTooltips:     types.BoolValue(false),
				SyncCursor:       types.BoolValue(true),
				AutoApplyFilters: types.BoolValue(true),
				HidePanelBorders: types.BoolValue(false),
			},
			options: &kbapi.KbnDashboardOptions{
				HidePanelTitles:  new(false),
				UseMargins:       new(true),
				SyncColors:       new(false),
				SyncTooltips:     new(false),
				SyncCursor:       new(true),
				AutoApplyFilters: new(true),
				HidePanelBorders: new(false),
			},
			want: &optionsModel{
				HidePanelTitles:  types.BoolValue(false),
				UseMargins:       types.BoolValue(true),
				SyncColors:       types.BoolValue(false),
				SyncTooltips:     types.BoolValue(false),
				SyncCursor:       types.BoolValue(true),
				AutoApplyFilters: types.BoolValue(true),
				HidePanelBorders: types.BoolValue(false),
			},
		},
		{
			name: "converts all fields when set to true",
			options: &kbapi.KbnDashboardOptions{
				HidePanelTitles: new(true),
				UseMargins:      new(true),
				SyncColors:      new(true),
				SyncTooltips:    new(true),
				SyncCursor:      new(true),
			},
			want: &optionsModel{
				HidePanelTitles:  types.BoolValue(true),
				UseMargins:       types.BoolValue(true),
				SyncColors:       types.BoolValue(true),
				SyncTooltips:     types.BoolValue(true),
				SyncCursor:       types.BoolValue(true),
				AutoApplyFilters: types.BoolNull(),
				HidePanelBorders: types.BoolNull(),
			},
		},
		{
			name: "converts all fields when set to false",
			options: &kbapi.KbnDashboardOptions{
				HidePanelTitles: new(false),
				UseMargins:      new(false),
				SyncColors:      new(false),
				SyncTooltips:    new(false),
				SyncCursor:      new(false),
			},
			want: &optionsModel{
				HidePanelTitles:  types.BoolValue(false),
				UseMargins:       types.BoolValue(false),
				SyncColors:       types.BoolValue(false),
				SyncTooltips:     types.BoolValue(false),
				SyncCursor:       types.BoolValue(false),
				AutoApplyFilters: types.BoolNull(),
				HidePanelBorders: types.BoolNull(),
			},
		},
		{
			name: "handles mixed nil and set values",
			options: &kbapi.KbnDashboardOptions{
				HidePanelTitles: new(true),
				UseMargins:      nil,
				SyncColors:      new(false),
				SyncTooltips:    nil,
				SyncCursor:      new(true),
			},
			want: &optionsModel{
				HidePanelTitles:  types.BoolValue(true),
				UseMargins:       types.BoolNull(),
				SyncColors:       types.BoolValue(false),
				SyncTooltips:     types.BoolNull(),
				SyncCursor:       types.BoolValue(true),
				AutoApplyFilters: types.BoolNull(),
				HidePanelBorders: types.BoolNull(),
			},
		},
		{
			name: "handles all nil values",
			options: &kbapi.KbnDashboardOptions{
				HidePanelTitles: nil,
				UseMargins:      nil,
				SyncColors:      nil,
				SyncTooltips:    nil,
				SyncCursor:      nil,
			},
			want: &optionsModel{
				HidePanelTitles:  types.BoolNull(),
				UseMargins:       types.BoolNull(),
				SyncColors:       types.BoolNull(),
				SyncTooltips:     types.BoolNull(),
				SyncCursor:       types.BoolNull(),
				AutoApplyFilters: types.BoolNull(),
				HidePanelBorders: types.BoolNull(),
			},
		},
		{
			name:    "handles empty struct",
			options: &kbapi.KbnDashboardOptions{},
			want: &optionsModel{
				HidePanelTitles:  types.BoolNull(),
				UseMargins:       types.BoolNull(),
				SyncColors:       types.BoolNull(),
				SyncTooltips:     types.BoolNull(),
				SyncCursor:       types.BoolNull(),
				AutoApplyFilters: types.BoolNull(),
				HidePanelBorders: types.BoolNull(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &dashboardModel{
				Options: tt.stateOptions,
			}
			apiOpts := kbapi.KbnDashboardOptions{}
			if tt.options != nil {
				apiOpts = *tt.options
			}
			got := m.mapOptionsFromAPI(apiOpts)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_optionsModel_toAPI(t *testing.T) {
	tests := []struct {
		name  string
		model optionsModel
		want  *kbapi.KbnDashboardOptions
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
			want: &kbapi.KbnDashboardOptions{
				HidePanelTitles: new(true),
				UseMargins:      new(true),
				SyncColors:      new(true),
				SyncTooltips:    new(true),
				SyncCursor:      new(true),
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
			want: &kbapi.KbnDashboardOptions{
				HidePanelTitles: new(false),
				UseMargins:      new(false),
				SyncColors:      new(false),
				SyncTooltips:    new(false),
				SyncCursor:      new(false),
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
			want: &kbapi.KbnDashboardOptions{
				HidePanelTitles: new(true),
				UseMargins:      nil,
				SyncColors:      new(false),
				SyncTooltips:    nil,
				SyncCursor:      new(true),
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
			want: &kbapi.KbnDashboardOptions{
				HidePanelTitles: nil,
				UseMargins:      new(true),
				SyncColors:      nil,
				SyncTooltips:    new(false),
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
			want: &kbapi.KbnDashboardOptions{
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
			want: &kbapi.KbnDashboardOptions{
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
			want: &kbapi.KbnDashboardOptions{
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
			want: &kbapi.KbnDashboardOptions{
				HidePanelTitles: new(true),
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
			roundTripModel := dm.mapOptionsFromAPI(*apiModel)
			require.NotNil(t, roundTripModel)

			// Compare the original and round-trip models
			assert.Equal(t, tt.model.HidePanelTitles, roundTripModel.HidePanelTitles)
			assert.Equal(t, tt.model.UseMargins, roundTripModel.UseMargins)
			assert.Equal(t, tt.model.SyncColors, roundTripModel.SyncColors)
			assert.Equal(t, tt.model.SyncTooltips, roundTripModel.SyncTooltips)
			assert.Equal(t, tt.model.SyncCursor, roundTripModel.SyncCursor)
		})
	}
}
