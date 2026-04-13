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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const optionsListControlTestDataViewID = "dv1"

func makeAPIConfig(dataViewID, fieldName string) *kbapi.KbnDashboardPanelTypeOptionsListControl {
	p := &kbapi.KbnDashboardPanelTypeOptionsListControl{}
	p.Config.DataViewId = dataViewID
	p.Config.FieldName = fieldName
	return p
}

// Test: nil config block with non-nil tfPanel preserves nil intent.
func Test_populateOptionsListControlFromAPI_nilBlock_preservedAsNil(t *testing.T) {
	pm := &panelModel{}
	tfPanel := &panelModel{}
	populateOptionsListControlFromAPI(pm, tfPanel, makeAPIConfig(optionsListControlTestDataViewID, "field1"))
	assert.Nil(t, pm.OptionsListControlConfig)
}

// Test: on import (tfPanel == nil), required fields and user-configurable optional fields are
// populated; server-default boolean flags (use_global_filters, exclude, exists_selected,
// ignore_validations, run_past_timeout) and sort are intentionally left null to avoid
// post-import drift when users have not explicitly configured them.
func Test_populateOptionsListControlFromAPI_import_populatesUserConfigurableFields(t *testing.T) {
	pm := &panelModel{}
	st := kbapi.KbnDashboardPanelTypeOptionsListControlConfigSearchTechniquePrefix
	var api kbapi.KbnDashboardPanelTypeOptionsListControl
	api.Config.DataViewId = optionsListControlTestDataViewID
	api.Config.FieldName = "field1"
	api.Config.Title = new("My Control")
	api.Config.UseGlobalFilters = new(true)
	api.Config.IgnoreValidations = new(false)
	api.Config.SingleSelect = new(true)
	api.Config.Exclude = new(false)
	api.Config.ExistsSelected = new(true)
	api.Config.RunPastTimeout = new(false)
	api.Config.SearchTechnique = &st
	api.Config.DisplaySettings = &struct {
		HideActionBar *bool   `json:"hide_action_bar,omitempty"`
		HideExclude   *bool   `json:"hide_exclude,omitempty"`
		HideExists    *bool   `json:"hide_exists,omitempty"`
		HideSort      *bool   `json:"hide_sort,omitempty"`
		Placeholder   *string `json:"placeholder,omitempty"`
	}{
		Placeholder:   new("Select..."),
		HideActionBar: new(true),
		HideExclude:   new(false),
		HideExists:    new(true),
		HideSort:      new(false),
	}
	api.Config.Sort = &struct {
		By        kbapi.KbnDashboardPanelTypeOptionsListControlConfigSortBy        `json:"by"`
		Direction kbapi.KbnDashboardPanelTypeOptionsListControlConfigSortDirection `json:"direction"`
	}{
		By:        kbapi.KbnDashboardPanelTypeOptionsListControlConfigSortByUnderscoreKey,
		Direction: kbapi.KbnDashboardPanelTypeOptionsListControlConfigSortDirectionAsc,
	}
	populateOptionsListControlFromAPI(pm, nil, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	cfg := pm.OptionsListControlConfig
	assert.Equal(t, types.StringValue(optionsListControlTestDataViewID), cfg.DataViewID)
	assert.Equal(t, types.StringValue("field1"), cfg.FieldName)
	assert.Equal(t, types.StringValue("My Control"), cfg.Title)
	assert.Equal(t, types.BoolValue(true), cfg.SingleSelect)
	assert.Equal(t, types.StringValue("prefix"), cfg.SearchTechnique)
	// Server-default boolean flags are left null on import to match apply-read null-preservation.
	assert.True(t, cfg.UseGlobalFilters.IsNull())
	assert.True(t, cfg.IgnoreValidations.IsNull())
	assert.True(t, cfg.Exclude.IsNull())
	assert.True(t, cfg.ExistsSelected.IsNull())
	assert.True(t, cfg.RunPastTimeout.IsNull())
	assert.Nil(t, cfg.Sort)
	require.NotNil(t, cfg.DisplaySettings)
	assert.Equal(t, types.StringValue("Select..."), cfg.DisplaySettings.Placeholder)
	assert.Equal(t, types.BoolValue(true), cfg.DisplaySettings.HideActionBar)
	assert.Equal(t, types.BoolValue(false), cfg.DisplaySettings.HideExclude)
	assert.Equal(t, types.BoolValue(true), cfg.DisplaySettings.HideExists)
	assert.Equal(t, types.BoolValue(false), cfg.DisplaySettings.HideSort)
}

// Test: on import with no optional fields, only required fields are populated.
func Test_populateOptionsListControlFromAPI_import_requiredFieldsOnly(t *testing.T) {
	pm := &panelModel{}
	populateOptionsListControlFromAPI(pm, nil, makeAPIConfig("dv2", "status"))
	require.NotNil(t, pm.OptionsListControlConfig)
	assert.Equal(t, types.StringValue("dv2"), pm.OptionsListControlConfig.DataViewID)
	assert.Equal(t, types.StringValue("status"), pm.OptionsListControlConfig.FieldName)
	assert.Nil(t, pm.OptionsListControlConfig.DisplaySettings)
	assert.Nil(t, pm.OptionsListControlConfig.Sort)
}

// Test: existing block with known fields gets updated from API.
func Test_populateOptionsListControlFromAPI_knownFields_updatedFromAPI(t *testing.T) {
	pm := &panelModel{
		OptionsListControlConfig: &optionsListControlConfigModel{
			DataViewID:       types.StringValue("old-dv"),
			FieldName:        types.StringValue("old-field"),
			UseGlobalFilters: types.BoolValue(false),
			SearchTechnique:  types.StringValue("prefix"),
		},
	}
	tfPanel := &panelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	st := kbapi.KbnDashboardPanelTypeOptionsListControlConfigSearchTechniqueWildcard
	var api kbapi.KbnDashboardPanelTypeOptionsListControl
	api.Config.DataViewId = "new-dv"
	api.Config.FieldName = "new-field"
	api.Config.UseGlobalFilters = new(true)
	api.Config.SearchTechnique = &st
	populateOptionsListControlFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	assert.Equal(t, types.StringValue("new-dv"), pm.OptionsListControlConfig.DataViewID)
	assert.Equal(t, types.StringValue("new-field"), pm.OptionsListControlConfig.FieldName)
	assert.Equal(t, types.BoolValue(true), pm.OptionsListControlConfig.UseGlobalFilters)
	assert.Equal(t, types.StringValue("wildcard"), pm.OptionsListControlConfig.SearchTechnique)
}

// Test: null-preservation — null optional fields in state are not overwritten by API values.
func Test_populateOptionsListControlFromAPI_nullFields_preservedAsNull(t *testing.T) {
	pm := &panelModel{
		OptionsListControlConfig: &optionsListControlConfigModel{
			DataViewID:       types.StringValue(optionsListControlTestDataViewID),
			FieldName:        types.StringValue("f1"),
			UseGlobalFilters: types.BoolNull(),
			SearchTechnique:  types.StringNull(),
		},
	}
	tfPanel := &panelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	st := kbapi.KbnDashboardPanelTypeOptionsListControlConfigSearchTechniqueExact
	var api kbapi.KbnDashboardPanelTypeOptionsListControl
	api.Config.DataViewId = optionsListControlTestDataViewID
	api.Config.FieldName = "f1"
	api.Config.UseGlobalFilters = new(true)
	api.Config.SearchTechnique = &st
	populateOptionsListControlFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	assert.True(t, pm.OptionsListControlConfig.UseGlobalFilters.IsNull())
	assert.True(t, pm.OptionsListControlConfig.SearchTechnique.IsNull())
}

// Test: nil display_settings block in state is preserved as nil even when API returns data.
func Test_populateOptionsListControlFromAPI_nilDisplaySettings_preservedAsNil(t *testing.T) {
	pm := &panelModel{
		OptionsListControlConfig: &optionsListControlConfigModel{
			DataViewID:      types.StringValue(optionsListControlTestDataViewID),
			FieldName:       types.StringValue("f1"),
			DisplaySettings: nil,
		},
	}
	tfPanel := &panelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	var api kbapi.KbnDashboardPanelTypeOptionsListControl
	api.Config.DataViewId = optionsListControlTestDataViewID
	api.Config.FieldName = "f1"
	api.Config.DisplaySettings = &struct {
		HideActionBar *bool   `json:"hide_action_bar,omitempty"`
		HideExclude   *bool   `json:"hide_exclude,omitempty"`
		HideExists    *bool   `json:"hide_exists,omitempty"`
		HideSort      *bool   `json:"hide_sort,omitempty"`
		Placeholder   *string `json:"placeholder,omitempty"`
	}{
		Placeholder: new("test"),
	}
	populateOptionsListControlFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	assert.Nil(t, pm.OptionsListControlConfig.DisplaySettings)
}

// Test: buildOptionsListControlConfig sets all known fields.
func Test_buildOptionsListControlConfig_allFields(t *testing.T) {
	pm := panelModel{
		OptionsListControlConfig: &optionsListControlConfigModel{
			DataViewID:        types.StringValue(optionsListControlTestDataViewID),
			FieldName:         types.StringValue("field1"),
			Title:             types.StringValue("My Title"),
			UseGlobalFilters:  types.BoolValue(true),
			IgnoreValidations: types.BoolValue(false),
			SingleSelect:      types.BoolValue(true),
			Exclude:           types.BoolValue(false),
			ExistsSelected:    types.BoolValue(false),
			RunPastTimeout:    types.BoolValue(true),
			SearchTechnique:   types.StringValue("exact"),
			SelectedOptions:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("active"), types.StringValue("inactive")}),
			DisplaySettings: &optionsListControlDisplaySettingsModel{
				Placeholder:   types.StringValue("Pick one"),
				HideActionBar: types.BoolValue(true),
				HideExclude:   types.BoolValue(false),
				HideExists:    types.BoolValue(true),
				HideSort:      types.BoolValue(false),
			},
			Sort: &optionsListControlSortModel{
				By:        types.StringValue("_count"),
				Direction: types.StringValue("desc"),
			},
		},
	}
	olPanel := kbapi.KbnDashboardPanelTypeOptionsListControl{}
	buildOptionsListControlConfig(pm, &olPanel)

	assert.Equal(t, optionsListControlTestDataViewID, olPanel.Config.DataViewId)
	assert.Equal(t, "field1", olPanel.Config.FieldName)
	require.NotNil(t, olPanel.Config.Title)
	assert.Equal(t, "My Title", *olPanel.Config.Title)
	require.NotNil(t, olPanel.Config.UseGlobalFilters)
	assert.True(t, *olPanel.Config.UseGlobalFilters)
	require.NotNil(t, olPanel.Config.SingleSelect)
	assert.True(t, *olPanel.Config.SingleSelect)
	require.NotNil(t, olPanel.Config.RunPastTimeout)
	assert.True(t, *olPanel.Config.RunPastTimeout)
	require.NotNil(t, olPanel.Config.SearchTechnique)
	assert.Equal(t, kbapi.KbnDashboardPanelTypeOptionsListControlConfigSearchTechniqueExact, *olPanel.Config.SearchTechnique)
	require.NotNil(t, olPanel.Config.SelectedOptions)
	require.Len(t, *olPanel.Config.SelectedOptions, 2)
	require.NotNil(t, olPanel.Config.DisplaySettings)
	assert.Equal(t, "Pick one", *olPanel.Config.DisplaySettings.Placeholder)
	assert.True(t, *olPanel.Config.DisplaySettings.HideActionBar)
	require.NotNil(t, olPanel.Config.Sort)
	assert.Equal(t, kbapi.KbnDashboardPanelTypeOptionsListControlConfigSortByUnderscoreCount, olPanel.Config.Sort.By)
	assert.Equal(t, kbapi.KbnDashboardPanelTypeOptionsListControlConfigSortDirectionDesc, olPanel.Config.Sort.Direction)
}

// Test: buildOptionsListControlConfig with null SelectedOptions omits the field.
func Test_buildOptionsListControlConfig_nullSelectedOptions_omitted(t *testing.T) {
	pm := panelModel{
		OptionsListControlConfig: &optionsListControlConfigModel{
			DataViewID:      types.StringValue(optionsListControlTestDataViewID),
			FieldName:       types.StringValue("field1"),
			SelectedOptions: types.ListNull(types.StringType),
		},
	}
	olPanel := kbapi.KbnDashboardPanelTypeOptionsListControl{}
	buildOptionsListControlConfig(pm, &olPanel)
	assert.Nil(t, olPanel.Config.SelectedOptions)
}

// Test: buildOptionsListControlConfig with nil optional fields omits them.
func Test_buildOptionsListControlConfig_nullOptionalFields_omitted(t *testing.T) {
	pm := panelModel{
		OptionsListControlConfig: &optionsListControlConfigModel{
			DataViewID:       types.StringValue(optionsListControlTestDataViewID),
			FieldName:        types.StringValue("field1"),
			UseGlobalFilters: types.BoolNull(),
			SearchTechnique:  types.StringNull(),
		},
	}
	olPanel := kbapi.KbnDashboardPanelTypeOptionsListControl{}
	buildOptionsListControlConfig(pm, &olPanel)
	assert.Equal(t, optionsListControlTestDataViewID, olPanel.Config.DataViewId)
	assert.Nil(t, olPanel.Config.UseGlobalFilters)
	assert.Nil(t, olPanel.Config.SearchTechnique)
	assert.Nil(t, olPanel.Config.DisplaySettings)
	assert.Nil(t, olPanel.Config.Sort)
}

// Test: round-trip — build then populate returns identical state.
func Test_optionsListControl_roundTrip(t *testing.T) {
	original := &optionsListControlConfigModel{
		DataViewID:       types.StringValue("my-dv"),
		FieldName:        types.StringValue("status"),
		SearchTechnique:  types.StringValue("prefix"),
		SingleSelect:     types.BoolValue(true),
		UseGlobalFilters: types.BoolValue(false),
		DisplaySettings: &optionsListControlDisplaySettingsModel{
			Placeholder:   types.StringValue("Search..."),
			HideActionBar: types.BoolValue(false),
			HideExclude:   types.BoolNull(),
			HideExists:    types.BoolNull(),
			HideSort:      types.BoolValue(true),
		},
		Sort: &optionsListControlSortModel{
			By:        types.StringValue("_key"),
			Direction: types.StringValue("asc"),
		},
	}

	pm := panelModel{OptionsListControlConfig: original}
	olPanel := kbapi.KbnDashboardPanelTypeOptionsListControl{}
	buildOptionsListControlConfig(pm, &olPanel)

	out := &panelModel{OptionsListControlConfig: &optionsListControlConfigModel{
		DataViewID:       types.StringValue("my-dv"),
		FieldName:        types.StringValue("status"),
		SearchTechnique:  types.StringValue("prefix"),
		SingleSelect:     types.BoolValue(true),
		UseGlobalFilters: types.BoolValue(false),
		DisplaySettings: &optionsListControlDisplaySettingsModel{
			Placeholder:   types.StringValue("Search..."),
			HideActionBar: types.BoolValue(false),
			HideExclude:   types.BoolNull(),
			HideExists:    types.BoolNull(),
			HideSort:      types.BoolValue(true),
		},
		Sort: &optionsListControlSortModel{
			By:        types.StringValue("_key"),
			Direction: types.StringValue("asc"),
		},
	}}
	tfPanel := &panelModel{OptionsListControlConfig: out.OptionsListControlConfig}
	populateOptionsListControlFromAPI(out, tfPanel, &olPanel)

	require.NotNil(t, out.OptionsListControlConfig)
	cfg := out.OptionsListControlConfig
	assert.Equal(t, types.StringValue("my-dv"), cfg.DataViewID)
	assert.Equal(t, types.StringValue("status"), cfg.FieldName)
	assert.Equal(t, types.StringValue("prefix"), cfg.SearchTechnique)
	assert.Equal(t, types.BoolValue(true), cfg.SingleSelect)
	assert.Equal(t, types.BoolValue(false), cfg.UseGlobalFilters)
	require.NotNil(t, cfg.DisplaySettings)
	assert.Equal(t, types.StringValue("Search..."), cfg.DisplaySettings.Placeholder)
	assert.Equal(t, types.BoolValue(false), cfg.DisplaySettings.HideActionBar)
	assert.True(t, cfg.DisplaySettings.HideExclude.IsNull())
	assert.True(t, cfg.DisplaySettings.HideExists.IsNull())
	assert.Equal(t, types.BoolValue(true), cfg.DisplaySettings.HideSort)
	require.NotNil(t, cfg.Sort)
	assert.Equal(t, types.StringValue("_key"), cfg.Sort.By)
	assert.Equal(t, types.StringValue("asc"), cfg.Sort.Direction)
}

// Test: selectedOptionsToList converts string items.
func Test_selectedOptionsToList_stringItems(t *testing.T) {
	var item1 kbapi.KbnDashboardPanelTypeOptionsListControl_Config_SelectedOptions_Item
	require.NoError(t, item1.FromKbnDashboardPanelTypeOptionsListControlConfigSelectedOptions0("alpha"))
	var item2 kbapi.KbnDashboardPanelTypeOptionsListControl_Config_SelectedOptions_Item
	require.NoError(t, item2.FromKbnDashboardPanelTypeOptionsListControlConfigSelectedOptions0("beta"))

	result := selectedOptionsToList([]kbapi.KbnDashboardPanelTypeOptionsListControl_Config_SelectedOptions_Item{item1, item2})
	require.False(t, result.IsNull())
	elems := result.Elements()
	require.Len(t, elems, 2)
	assert.Equal(t, types.StringValue("alpha"), elems[0])
	assert.Equal(t, types.StringValue("beta"), elems[1])
}

// Test: selectedOptionsToList converts numeric items using fixed-point notation.
func Test_selectedOptionsToList_numericItems(t *testing.T) {
	var item1 kbapi.KbnDashboardPanelTypeOptionsListControl_Config_SelectedOptions_Item
	require.NoError(t, item1.FromKbnDashboardPanelTypeOptionsListControlConfigSelectedOptions1(1000000))
	var item2 kbapi.KbnDashboardPanelTypeOptionsListControl_Config_SelectedOptions_Item
	require.NoError(t, item2.FromKbnDashboardPanelTypeOptionsListControlConfigSelectedOptions1(3.14))

	result := selectedOptionsToList([]kbapi.KbnDashboardPanelTypeOptionsListControl_Config_SelectedOptions_Item{item1, item2})
	require.False(t, result.IsNull())
	elems := result.Elements()
	require.Len(t, elems, 2)
	// Must be fixed-point, not scientific notation.
	assert.Equal(t, types.StringValue("1000000"), elems[0])
	assert.Equal(t, types.StringValue("3.14"), elems[1])
}
