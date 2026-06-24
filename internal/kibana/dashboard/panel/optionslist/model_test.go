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

package optionslist

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const optionsListControlTestDataViewID = "dv1"

// olFieldCfg is the field-based variant of the options list control config.
// The control config is a discriminated union (field-based vs ES|QL); the TF
// model only describes the field-based variant.
type olFieldCfg = kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField

func ptr[T any](v T) *T { return &v }

func makeAPIConfig(dataViewID, fieldName string) *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl {
	p := &kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	_ = p.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(olFieldCfg{
		DataViewId: dataViewID,
		FieldName:  fieldName,
	})
	return p
}

func olConfigField(t *testing.T, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl) olFieldCfg {
	c, err := p.Config.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField()
	require.NoError(t, err)
	return c
}

// Test: nil config block with non-nil tfPanel preserves nil intent.
func Test_PopulateFromAPI_nilBlock_preservedAsNil(t *testing.T) {
	pm := &models.PanelModel{}
	tfPanel := &models.PanelModel{}
	PopulateFromAPI(pm, tfPanel, makeAPIConfig(optionsListControlTestDataViewID, "field1"))
	assert.Nil(t, pm.OptionsListControlConfig)
}

// Test: on import (tfPanel == nil), required fields and user-configurable optional fields are
// populated; server-default boolean flags (use_global_filters, exclude, exists_selected,
// ignore_validations, run_past_timeout) and sort are intentionally left null to avoid
// post-import drift when users have not explicitly configured them.
func Test_PopulateFromAPI_import_populatesUserConfigurableFields(t *testing.T) {
	pm := &models.PanelModel{}
	st := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSearchTechniquePrefix
	c := olFieldCfg{
		DataViewId:        optionsListControlTestDataViewID,
		FieldName:         "field1",
		Title:             ptr("My Control"),
		UseGlobalFilters:  ptr(true),
		IgnoreValidations: ptr(false),
		SingleSelect:      ptr(true),
		Exclude:           ptr(false),
		ExistsSelected:    ptr(true),
		RunPastTimeout:    ptr(false),
		SearchTechnique:   &st,
		DisplaySettings: &struct {
			HideActionBar *bool   `json:"hide_action_bar,omitempty"`
			HideExclude   *bool   `json:"hide_exclude,omitempty"`
			HideExists    *bool   `json:"hide_exists,omitempty"`
			HideSort      *bool   `json:"hide_sort,omitempty"`
			Placeholder   *string `json:"placeholder,omitempty"`
		}{
			Placeholder:   ptr("Select..."),
			HideActionBar: ptr(true),
			HideExclude:   ptr(false),
			HideExists:    ptr(true),
			HideSort:      ptr(false),
		},
		Sort: &struct {
			By        kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSortBy        `json:"by"`
			Direction kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSortDirection `json:"direction"`
		}{
			By:        kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSortByUnderscoreKey,
			Direction: kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSortDirectionAsc,
		},
	}
	var api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl
	api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(c)
	PopulateFromAPI(pm, nil, &api)
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
func Test_PopulateFromAPI_import_requiredFieldsOnly(t *testing.T) {
	pm := &models.PanelModel{}
	PopulateFromAPI(pm, nil, makeAPIConfig("dv2", "status"))
	require.NotNil(t, pm.OptionsListControlConfig)
	assert.Equal(t, types.StringValue("dv2"), pm.OptionsListControlConfig.DataViewID)
	assert.Equal(t, types.StringValue("status"), pm.OptionsListControlConfig.FieldName)
	assert.Nil(t, pm.OptionsListControlConfig.DisplaySettings)
	assert.Nil(t, pm.OptionsListControlConfig.Sort)
}

// Test: existing block with known fields gets updated from API.
func Test_PopulateFromAPI_knownFields_updatedFromAPI(t *testing.T) {
	pm := &models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			DataViewID:       types.StringValue("old-dv"),
			FieldName:        types.StringValue("old-field"),
			UseGlobalFilters: types.BoolValue(false),
			SearchTechnique:  types.StringValue("prefix"),
		},
	}
	tfPanel := &models.PanelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	st := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSearchTechniqueWildcard
	c := olFieldCfg{
		DataViewId:       "new-dv",
		FieldName:        "new-field",
		UseGlobalFilters: ptr(true),
		SearchTechnique:  &st,
	}
	var api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl
	api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(c)
	PopulateFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	assert.Equal(t, types.StringValue("new-dv"), pm.OptionsListControlConfig.DataViewID)
	assert.Equal(t, types.StringValue("new-field"), pm.OptionsListControlConfig.FieldName)
	assert.Equal(t, types.BoolValue(true), pm.OptionsListControlConfig.UseGlobalFilters)
	assert.Equal(t, types.StringValue("wildcard"), pm.OptionsListControlConfig.SearchTechnique)
}

// Test: null-preservation — null optional fields in state are not overwritten by API values.
func Test_PopulateFromAPI_nullFields_preservedAsNull(t *testing.T) {
	pm := &models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			DataViewID:       types.StringValue(optionsListControlTestDataViewID),
			FieldName:        types.StringValue("f1"),
			UseGlobalFilters: types.BoolNull(),
			SearchTechnique:  types.StringNull(),
		},
	}
	tfPanel := &models.PanelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	st := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSearchTechniqueExact
	c := olFieldCfg{
		DataViewId:       optionsListControlTestDataViewID,
		FieldName:        "f1",
		UseGlobalFilters: ptr(true),
		SearchTechnique:  &st,
	}
	var api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl
	api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(c)
	PopulateFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	assert.True(t, pm.OptionsListControlConfig.UseGlobalFilters.IsNull())
	assert.True(t, pm.OptionsListControlConfig.SearchTechnique.IsNull())
}

// Test: nil display_settings block in state is preserved as nil even when API returns data.
func Test_PopulateFromAPI_nilDisplaySettings_preservedAsNil(t *testing.T) {
	pm := &models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			DataViewID:      types.StringValue(optionsListControlTestDataViewID),
			FieldName:       types.StringValue("f1"),
			DisplaySettings: nil,
		},
	}
	tfPanel := &models.PanelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	c := olFieldCfg{
		DataViewId: optionsListControlTestDataViewID,
		FieldName:  "f1",
		DisplaySettings: &struct {
			HideActionBar *bool   `json:"hide_action_bar,omitempty"`
			HideExclude   *bool   `json:"hide_exclude,omitempty"`
			HideExists    *bool   `json:"hide_exists,omitempty"`
			HideSort      *bool   `json:"hide_sort,omitempty"`
			Placeholder   *string `json:"placeholder,omitempty"`
		}{
			Placeholder: ptr("test"),
		},
	}
	var api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl
	api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(c)
	PopulateFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	assert.Nil(t, pm.OptionsListControlConfig.DisplaySettings)
}

// Test: BuildConfig sets all known fields.
func Test_BuildConfig_allFields(t *testing.T) {
	pm := models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
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
			DisplaySettings: &models.OptionsListControlDisplaySettingsModel{
				Placeholder:   types.StringValue("Pick one"),
				HideActionBar: types.BoolValue(true),
				HideExclude:   types.BoolValue(false),
				HideExists:    types.BoolValue(true),
				HideSort:      types.BoolValue(false),
			},
			Sort: &models.OptionsListControlSortModel{
				By:        types.StringValue("_count"),
				Direction: types.StringValue("desc"),
			},
		},
	}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	BuildConfig(pm, &olPanel)

	cfg := olConfigField(t, olPanel)
	assert.Equal(t, optionsListControlTestDataViewID, cfg.DataViewId)
	assert.Equal(t, "field1", cfg.FieldName)
	require.NotNil(t, cfg.Title)
	assert.Equal(t, "My Title", *cfg.Title)
	require.NotNil(t, cfg.UseGlobalFilters)
	assert.True(t, *cfg.UseGlobalFilters)
	require.NotNil(t, cfg.SingleSelect)
	assert.True(t, *cfg.SingleSelect)
	require.NotNil(t, cfg.RunPastTimeout)
	assert.True(t, *cfg.RunPastTimeout)
	require.NotNil(t, cfg.SearchTechnique)
	assert.Equal(t, kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSearchTechniqueExact, *cfg.SearchTechnique)
	require.NotNil(t, cfg.SelectedOptions)
	require.Len(t, *cfg.SelectedOptions, 2)
	require.NotNil(t, cfg.DisplaySettings)
	assert.Equal(t, "Pick one", *cfg.DisplaySettings.Placeholder)
	assert.True(t, *cfg.DisplaySettings.HideActionBar)
	require.NotNil(t, cfg.Sort)
	assert.Equal(t, kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSortByUnderscoreCount, cfg.Sort.By)
	assert.Equal(t, kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSortDirectionDesc, cfg.Sort.Direction)
}

// Test: BuildConfig with null SelectedOptions omits the field.
func Test_BuildConfig_nullSelectedOptions_omitted(t *testing.T) {
	pm := models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			DataViewID:      types.StringValue(optionsListControlTestDataViewID),
			FieldName:       types.StringValue("field1"),
			SelectedOptions: types.ListNull(types.StringType),
		},
	}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	BuildConfig(pm, &olPanel)
	cfg := olConfigField(t, olPanel)
	assert.Nil(t, cfg.SelectedOptions)
}

// Test: BuildConfig with nil optional fields omits them.
func Test_BuildConfig_nullOptionalFields_omitted(t *testing.T) {
	pm := models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			DataViewID:       types.StringValue(optionsListControlTestDataViewID),
			FieldName:        types.StringValue("field1"),
			UseGlobalFilters: types.BoolNull(),
			SearchTechnique:  types.StringNull(),
		},
	}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	BuildConfig(pm, &olPanel)
	cfg := olConfigField(t, olPanel)
	assert.Equal(t, optionsListControlTestDataViewID, cfg.DataViewId)
	assert.Nil(t, cfg.UseGlobalFilters)
	assert.Nil(t, cfg.SearchTechnique)
	assert.Nil(t, cfg.DisplaySettings)
	assert.Nil(t, cfg.Sort)
}

// Test: round-trip — build then populate returns identical state.
func Test_optionsListControl_roundTrip(t *testing.T) {
	original := &models.OptionsListControlConfigModel{
		DataViewID:       types.StringValue("my-dv"),
		FieldName:        types.StringValue("status"),
		SearchTechnique:  types.StringValue("prefix"),
		SingleSelect:     types.BoolValue(true),
		UseGlobalFilters: types.BoolValue(false),
		DisplaySettings: &models.OptionsListControlDisplaySettingsModel{
			Placeholder:   types.StringValue("Search..."),
			HideActionBar: types.BoolValue(false),
			HideExclude:   types.BoolNull(),
			HideExists:    types.BoolNull(),
			HideSort:      types.BoolValue(true),
		},
		Sort: &models.OptionsListControlSortModel{
			By:        types.StringValue("_key"),
			Direction: types.StringValue("asc"),
		},
	}

	pm := models.PanelModel{OptionsListControlConfig: original}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	BuildConfig(pm, &olPanel)

	out := &models.PanelModel{OptionsListControlConfig: &models.OptionsListControlConfigModel{
		DataViewID:       types.StringValue("my-dv"),
		FieldName:        types.StringValue("status"),
		SearchTechnique:  types.StringValue("prefix"),
		SingleSelect:     types.BoolValue(true),
		UseGlobalFilters: types.BoolValue(false),
		DisplaySettings: &models.OptionsListControlDisplaySettingsModel{
			Placeholder:   types.StringValue("Search..."),
			HideActionBar: types.BoolValue(false),
			HideExclude:   types.BoolNull(),
			HideExists:    types.BoolNull(),
			HideSort:      types.BoolValue(true),
		},
		Sort: &models.OptionsListControlSortModel{
			By:        types.StringValue("_key"),
			Direction: types.StringValue("asc"),
		},
	}}
	tfPanel := &models.PanelModel{OptionsListControlConfig: out.OptionsListControlConfig}
	PopulateFromAPI(out, tfPanel, &olPanel)

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
	var item1 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item
	require.NoError(t, item1.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions0("alpha"))
	var item2 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item
	require.NoError(t, item2.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions0("beta"))

	result := selectedOptionsToList([]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item{item1, item2})
	require.False(t, result.IsNull())
	elems := result.Elements()
	require.Len(t, elems, 2)
	assert.Equal(t, types.StringValue("alpha"), elems[0])
	assert.Equal(t, types.StringValue("beta"), elems[1])
}

// Test: selectedOptionsToList converts numeric items using fixed-point notation.
func Test_selectedOptionsToList_numericItems(t *testing.T) {
	var item1 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item
	require.NoError(t, item1.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions1(1000000))
	var item2 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item
	require.NoError(t, item2.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions1(3.14))

	result := selectedOptionsToList([]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item{item1, item2})
	require.False(t, result.IsNull())
	elems := result.Elements()
	require.Len(t, elems, 2)
	// Must be fixed-point, not scientific notation.
	assert.Equal(t, types.StringValue("1000000"), elems[0])
	assert.Equal(t, types.StringValue("3.14"), elems[1])
}
