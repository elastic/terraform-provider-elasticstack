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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const optionsListControlTestDataViewID = "dv1"

// olFieldCfg / olEsqlCfg are the two variants of the options list control config union.
type olFieldCfg = kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField
type olEsqlCfg = kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql

func makeAPIConfig(t *testing.T, dataViewID, fieldName string) *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl {
	p := &kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	require.NoError(t, p.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(olFieldCfg{
		DataViewId: dataViewID,
		FieldName:  fieldName,
	}))
	return p
}

func makeEsqlAPIConfig(t *testing.T, esqlQuery string) *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl {
	p := &kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	require.NoError(t, p.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql(olEsqlCfg{
		EsqlQuery:    esqlQuery,
		ValuesSource: kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlValuesSourceEsql,
	}))
	return p
}

func olConfigField(t *testing.T, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl) olFieldCfg {
	c, err := p.Config.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField()
	require.NoError(t, err)
	return c
}

func olConfigEsql(t *testing.T, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl) olEsqlCfg {
	c, err := p.Config.AsKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql()
	require.NoError(t, err)
	return c
}

// ---------------------------------------------------------------------------
// by_field branch
// ---------------------------------------------------------------------------

// Test: nil config block with non-nil tfPanel preserves nil intent.
func Test_PopulateFromAPI_nilBlock_preservedAsNil(t *testing.T) {
	pm := &models.PanelModel{}
	tfPanel := &models.PanelModel{}
	PopulateFromAPI(pm, tfPanel, makeAPIConfig(t, optionsListControlTestDataViewID, "field1"))
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
		Title:             new("My Control"),
		UseGlobalFilters:  new(true),
		IgnoreValidations: new(false),
		SingleSelect:      new(true),
		Exclude:           new(false),
		ExistsSelected:    new(true),
		RunPastTimeout:    new(false),
		SearchTechnique:   &st,
		DisplaySettings: &displaySettingsAPI{
			Placeholder:   new("Select..."),
			HideActionBar: new(true),
			HideExclude:   new(false),
			HideExists:    new(true),
			HideSort:      new(false),
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
	require.NoError(t, api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(c))
	PopulateFromAPI(pm, nil, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	require.NotNil(t, pm.OptionsListControlConfig.ByField)
	assert.Nil(t, pm.OptionsListControlConfig.ByEsql)
	cfg := pm.OptionsListControlConfig.ByField
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
	PopulateFromAPI(pm, nil, makeAPIConfig(t, "dv2", "status"))
	require.NotNil(t, pm.OptionsListControlConfig)
	require.NotNil(t, pm.OptionsListControlConfig.ByField)
	assert.Equal(t, types.StringValue("dv2"), pm.OptionsListControlConfig.ByField.DataViewID)
	assert.Equal(t, types.StringValue("status"), pm.OptionsListControlConfig.ByField.FieldName)
	assert.Nil(t, pm.OptionsListControlConfig.ByField.DisplaySettings)
	assert.Nil(t, pm.OptionsListControlConfig.ByField.Sort)
}

// Test: existing block with known fields gets updated from API.
func Test_PopulateFromAPI_knownFields_updatedFromAPI(t *testing.T) {
	pm := &models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			ByField: &models.OptionsListControlByFieldModel{
				DataViewID:       types.StringValue("old-dv"),
				FieldName:        types.StringValue("old-field"),
				UseGlobalFilters: types.BoolValue(false),
				SearchTechnique:  types.StringValue("prefix"),
			},
		},
	}
	tfPanel := &models.PanelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	st := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSearchTechniqueWildcard
	c := olFieldCfg{
		DataViewId:       "new-dv",
		FieldName:        "new-field",
		UseGlobalFilters: new(true),
		SearchTechnique:  &st,
	}
	var api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl
	require.NoError(t, api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(c))
	PopulateFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	require.NotNil(t, pm.OptionsListControlConfig.ByField)
	cfg := pm.OptionsListControlConfig.ByField
	assert.Equal(t, types.StringValue("new-dv"), cfg.DataViewID)
	assert.Equal(t, types.StringValue("new-field"), cfg.FieldName)
	assert.Equal(t, types.BoolValue(true), cfg.UseGlobalFilters)
	assert.Equal(t, types.StringValue("wildcard"), cfg.SearchTechnique)
}

// Test: null-preservation — null optional fields in state are not overwritten by API values.
func Test_PopulateFromAPI_nullFields_preservedAsNull(t *testing.T) {
	pm := &models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			ByField: &models.OptionsListControlByFieldModel{
				DataViewID:       types.StringValue(optionsListControlTestDataViewID),
				FieldName:        types.StringValue("f1"),
				UseGlobalFilters: types.BoolNull(),
				SearchTechnique:  types.StringNull(),
			},
		},
	}
	tfPanel := &models.PanelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	st := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSearchTechniqueExact
	c := olFieldCfg{
		DataViewId:       optionsListControlTestDataViewID,
		FieldName:        "f1",
		UseGlobalFilters: new(true),
		SearchTechnique:  &st,
	}
	var api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl
	require.NoError(t, api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(c))
	PopulateFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	require.NotNil(t, pm.OptionsListControlConfig.ByField)
	assert.True(t, pm.OptionsListControlConfig.ByField.UseGlobalFilters.IsNull())
	assert.True(t, pm.OptionsListControlConfig.ByField.SearchTechnique.IsNull())
}

// Test: nil display_settings block in state is preserved as nil even when API returns data.
func Test_PopulateFromAPI_nilDisplaySettings_preservedAsNil(t *testing.T) {
	pm := &models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			ByField: &models.OptionsListControlByFieldModel{
				DataViewID:      types.StringValue(optionsListControlTestDataViewID),
				FieldName:       types.StringValue("f1"),
				DisplaySettings: nil,
			},
		},
	}
	tfPanel := &models.PanelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	c := olFieldCfg{
		DataViewId: optionsListControlTestDataViewID,
		FieldName:  "f1",
		DisplaySettings: &displaySettingsAPI{
			Placeholder: new("test"),
		},
	}
	var api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl
	require.NoError(t, api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField(c))
	PopulateFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	require.NotNil(t, pm.OptionsListControlConfig.ByField)
	assert.Nil(t, pm.OptionsListControlConfig.ByField.DisplaySettings)
}

// Test: if prior state configured by_esql but the API now returns the Field variant (e.g. the
// control was switched out-of-band in Kibana), PopulateFromAPI must switch state to by_field
// rather than silently preserving the stale by_esql block.
func Test_PopulateFromAPI_branchSwitchedRemotely_esqlToField(t *testing.T) {
	priorCfg := &models.OptionsListControlConfigModel{
		ByEsql: &models.OptionsListControlByEsqlModel{
			EsqlQuery:    types.StringValue("FROM logs | STATS BY service.name"),
			ValuesSource: types.StringValue(panelkit.EsqlValuesSourceUserValue),
		},
	}
	pm := &models.PanelModel{OptionsListControlConfig: priorCfg}
	tfPanel := &models.PanelModel{OptionsListControlConfig: priorCfg}

	PopulateFromAPI(pm, tfPanel, makeAPIConfig(t, optionsListControlTestDataViewID, "field1"))

	require.NotNil(t, pm.OptionsListControlConfig)
	require.NotNil(t, pm.OptionsListControlConfig.ByField)
	assert.Nil(t, pm.OptionsListControlConfig.ByEsql)
	assert.Equal(t, optionsListControlTestDataViewID, pm.OptionsListControlConfig.ByField.DataViewID.ValueString())
	assert.Equal(t, "field1", pm.OptionsListControlConfig.ByField.FieldName.ValueString())
}

// Test: the symmetric case — prior state configured by_field but the API now returns the ES|QL
// variant.
func Test_PopulateFromAPI_branchSwitchedRemotely_fieldToEsql(t *testing.T) {
	priorCfg := &models.OptionsListControlConfigModel{
		ByField: &models.OptionsListControlByFieldModel{
			DataViewID: types.StringValue(optionsListControlTestDataViewID),
			FieldName:  types.StringValue("field1"),
		},
	}
	pm := &models.PanelModel{OptionsListControlConfig: priorCfg}
	tfPanel := &models.PanelModel{OptionsListControlConfig: priorCfg}

	PopulateFromAPI(pm, tfPanel, makeEsqlAPIConfig(t, "FROM logs | STATS BY service.name"))

	require.NotNil(t, pm.OptionsListControlConfig)
	require.NotNil(t, pm.OptionsListControlConfig.ByEsql)
	assert.Nil(t, pm.OptionsListControlConfig.ByField)
	assert.Equal(t, "FROM logs | STATS BY service.name", pm.OptionsListControlConfig.ByEsql.EsqlQuery.ValueString())
}

// Test: BuildConfig returns an error when neither branch is set (defensive; schema-level
// ExactlyOneOfNestedAttrsValidator should normally prevent this at plan time).
func Test_BuildConfig_neitherBranchSet_errors(t *testing.T) {
	pm := models.PanelModel{OptionsListControlConfig: &models.OptionsListControlConfigModel{}}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	diags := BuildConfig(pm, &olPanel)
	assert.True(t, diags.HasError())
}

// Test: BuildConfig sets all known fields and leaves values_source unset (see buildFieldConfig).
func Test_BuildConfig_byField_allFields(t *testing.T) {
	pm := models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			ByField: &models.OptionsListControlByFieldModel{
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
		},
	}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	diags := BuildConfig(pm, &olPanel)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := olConfigField(t, olPanel)
	assert.Equal(t, optionsListControlTestDataViewID, cfg.DataViewId)
	assert.Equal(t, "field1", cfg.FieldName)
	// values_source is deliberately left unset on the wire for by_field (see buildFieldConfig):
	// Kibana defaults it to "field" when absent, and older Kibana versions reject the property
	// entirely if present.
	assert.Nil(t, cfg.ValuesSource)
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
func Test_BuildConfig_byField_nullSelectedOptions_omitted(t *testing.T) {
	pm := models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			ByField: &models.OptionsListControlByFieldModel{
				DataViewID:      types.StringValue(optionsListControlTestDataViewID),
				FieldName:       types.StringValue("field1"),
				SelectedOptions: types.ListNull(types.StringType),
			},
		},
	}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	BuildConfig(pm, &olPanel)
	cfg := olConfigField(t, olPanel)
	assert.Nil(t, cfg.SelectedOptions)
}

// Test: BuildConfig with nil optional fields omits them.
func Test_BuildConfig_byField_nullOptionalFields_omitted(t *testing.T) {
	pm := models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			ByField: &models.OptionsListControlByFieldModel{
				DataViewID:       types.StringValue(optionsListControlTestDataViewID),
				FieldName:        types.StringValue("field1"),
				UseGlobalFilters: types.BoolNull(),
				SearchTechnique:  types.StringNull(),
			},
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

// Test: round-trip — build then populate returns identical state (by_field).
func Test_optionsListControl_byField_roundTrip(t *testing.T) {
	original := &models.OptionsListControlConfigModel{
		ByField: &models.OptionsListControlByFieldModel{
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
		},
	}

	pm := models.PanelModel{OptionsListControlConfig: original}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	BuildConfig(pm, &olPanel)

	out := &models.PanelModel{OptionsListControlConfig: &models.OptionsListControlConfigModel{
		ByField: &models.OptionsListControlByFieldModel{
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
		},
	}}
	tfPanel := &models.PanelModel{OptionsListControlConfig: out.OptionsListControlConfig}
	PopulateFromAPI(out, tfPanel, &olPanel)

	require.NotNil(t, out.OptionsListControlConfig)
	require.NotNil(t, out.OptionsListControlConfig.ByField)
	cfg := out.OptionsListControlConfig.ByField
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

// Test: selectedOptionsFieldToList converts string items.
func Test_selectedOptionsFieldToList_stringItems(t *testing.T) {
	var item1 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item
	require.NoError(t, item1.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions0("alpha"))
	var item2 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item
	require.NoError(t, item2.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions0("beta"))

	result := selectedOptionsFieldToList([]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item{item1, item2})
	require.False(t, result.IsNull())
	elems := result.Elements()
	require.Len(t, elems, 2)
	assert.Equal(t, types.StringValue("alpha"), elems[0])
	assert.Equal(t, types.StringValue("beta"), elems[1])
}

// Test: selectedOptionsFieldToList converts numeric items using fixed-point notation.
func Test_selectedOptionsFieldToList_numericItems(t *testing.T) {
	var item1 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item
	require.NoError(t, item1.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions1(1000000))
	var item2 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item
	require.NoError(t, item2.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaFieldSelectedOptions1(3.14))

	result := selectedOptionsFieldToList([]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaField_SelectedOptions_Item{item1, item2})
	require.False(t, result.IsNull())
	elems := result.Elements()
	require.Len(t, elems, 2)
	// Must be fixed-point, not scientific notation.
	assert.Equal(t, types.StringValue("1000000"), elems[0])
	assert.Equal(t, types.StringValue("3.14"), elems[1])
}

// ---------------------------------------------------------------------------
// by_esql branch
// ---------------------------------------------------------------------------

// Test: nil config block with non-nil tfPanel preserves nil intent (ES|QL variant of the API response).
func Test_PopulateFromAPI_esql_nilBlock_preservedAsNil(t *testing.T) {
	pm := &models.PanelModel{}
	tfPanel := &models.PanelModel{}
	PopulateFromAPI(pm, tfPanel, makeEsqlAPIConfig(t, "FROM logs | STATS BY service.name"))
	assert.Nil(t, pm.OptionsListControlConfig)
}

// Test: on import, required fields (esql_query, values_source) and user-configurable optional
// fields are populated; optional booleans and sort are left null.
func Test_PopulateFromAPI_esql_import_populatesUserConfigurableFields(t *testing.T) {
	pm := &models.PanelModel{}
	st := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSearchTechniquePrefix
	c := olEsqlCfg{
		EsqlQuery:         "FROM logs | STATS BY service.name",
		ValuesSource:      kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlValuesSourceEsql,
		Title:             new("My ES|QL Control"),
		UseGlobalFilters:  new(true),
		IgnoreValidations: new(false),
		SingleSelect:      new(true),
		Exclude:           new(false),
		ExistsSelected:    new(true),
		RunPastTimeout:    new(false),
		SearchTechnique:   &st,
		DisplaySettings: &displaySettingsAPI{
			Placeholder: new("Select..."),
		},
		Sort: &struct {
			By        kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSortBy        `json:"by"`
			Direction kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSortDirection `json:"direction"`
		}{
			By:        kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSortBy("_key"),
			Direction: kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSortDirection("asc"),
		},
	}
	var api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl
	require.NoError(t, api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql(c))
	PopulateFromAPI(pm, nil, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	require.NotNil(t, pm.OptionsListControlConfig.ByEsql)
	assert.Nil(t, pm.OptionsListControlConfig.ByField)
	cfg := pm.OptionsListControlConfig.ByEsql
	assert.Equal(t, types.StringValue("FROM logs | STATS BY service.name"), cfg.EsqlQuery)
	assert.Equal(t, types.StringValue("esql_query"), cfg.ValuesSource)
	assert.Equal(t, types.StringValue("My ES|QL Control"), cfg.Title)
	assert.Equal(t, types.BoolValue(true), cfg.SingleSelect)
	assert.Equal(t, types.StringValue("prefix"), cfg.SearchTechnique)
	assert.True(t, cfg.UseGlobalFilters.IsNull())
	assert.True(t, cfg.IgnoreValidations.IsNull())
	assert.True(t, cfg.Exclude.IsNull())
	assert.True(t, cfg.ExistsSelected.IsNull())
	assert.True(t, cfg.RunPastTimeout.IsNull())
	assert.Nil(t, cfg.Sort)
	require.NotNil(t, cfg.DisplaySettings)
	assert.Equal(t, types.StringValue("Select..."), cfg.DisplaySettings.Placeholder)
}

// Test: existing by_esql block with known fields gets updated from API.
func Test_PopulateFromAPI_esql_knownFields_updatedFromAPI(t *testing.T) {
	pm := &models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			ByEsql: &models.OptionsListControlByEsqlModel{
				EsqlQuery:        types.StringValue("FROM old | STATS BY a"),
				ValuesSource:     types.StringValue("esql_query"),
				UseGlobalFilters: types.BoolValue(false),
			},
		},
	}
	tfPanel := &models.PanelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	c := olEsqlCfg{
		EsqlQuery:        "FROM new | STATS BY b",
		ValuesSource:     kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlValuesSourceEsql,
		UseGlobalFilters: new(true),
	}
	var api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl
	require.NoError(t, api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql(c))
	PopulateFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	require.NotNil(t, pm.OptionsListControlConfig.ByEsql)
	cfg := pm.OptionsListControlConfig.ByEsql
	assert.Equal(t, types.StringValue("FROM new | STATS BY b"), cfg.EsqlQuery)
	assert.Equal(t, types.BoolValue(true), cfg.UseGlobalFilters)
}

// Test: null-preservation on the ES|QL branch — null optional fields in state are not overwritten.
func Test_PopulateFromAPI_esql_nullFields_preservedAsNull(t *testing.T) {
	pm := &models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			ByEsql: &models.OptionsListControlByEsqlModel{
				EsqlQuery:        types.StringValue("FROM logs"),
				ValuesSource:     types.StringValue("esql_query"),
				UseGlobalFilters: types.BoolNull(),
				SearchTechnique:  types.StringNull(),
			},
		},
	}
	tfPanel := &models.PanelModel{OptionsListControlConfig: pm.OptionsListControlConfig}
	st := kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSearchTechniqueExact
	c := olEsqlCfg{
		EsqlQuery:        "FROM logs",
		ValuesSource:     kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlValuesSourceEsql,
		UseGlobalFilters: new(true),
		SearchTechnique:  &st,
	}
	var api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl
	require.NoError(t, api.Config.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql(c))
	PopulateFromAPI(pm, tfPanel, &api)
	require.NotNil(t, pm.OptionsListControlConfig)
	require.NotNil(t, pm.OptionsListControlConfig.ByEsql)
	assert.True(t, pm.OptionsListControlConfig.ByEsql.UseGlobalFilters.IsNull())
	assert.True(t, pm.OptionsListControlConfig.ByEsql.SearchTechnique.IsNull())
}

// Test: BuildConfig sets all known fields for by_esql, honoring the user-supplied values_source.
func Test_BuildConfig_byEsql_allFields(t *testing.T) {
	pm := models.PanelModel{
		OptionsListControlConfig: &models.OptionsListControlConfigModel{
			ByEsql: &models.OptionsListControlByEsqlModel{
				EsqlQuery:         types.StringValue("FROM logs | STATS BY service.name"),
				ValuesSource:      types.StringValue("esql_query"),
				Title:             types.StringValue("My Title"),
				UseGlobalFilters:  types.BoolValue(true),
				IgnoreValidations: types.BoolValue(false),
				SingleSelect:      types.BoolValue(true),
				Exclude:           types.BoolValue(false),
				ExistsSelected:    types.BoolValue(false),
				RunPastTimeout:    types.BoolValue(true),
				SearchTechnique:   types.StringValue("exact"),
				SelectedOptions:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("active")}),
				DisplaySettings: &models.OptionsListControlDisplaySettingsModel{
					Placeholder: types.StringValue("Pick one"),
				},
				Sort: &models.OptionsListControlSortModel{
					By:        types.StringValue("_count"),
					Direction: types.StringValue("desc"),
				},
			},
		},
	}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	diags := BuildConfig(pm, &olPanel)
	require.False(t, diags.HasError(), "%v", diags)

	cfg := olConfigEsql(t, olPanel)
	assert.Equal(t, "FROM logs | STATS BY service.name", cfg.EsqlQuery)
	assert.Equal(t, kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlValuesSourceEsql, cfg.ValuesSource)
	require.NotNil(t, cfg.Title)
	assert.Equal(t, "My Title", *cfg.Title)
	require.NotNil(t, cfg.UseGlobalFilters)
	assert.True(t, *cfg.UseGlobalFilters)
	require.NotNil(t, cfg.SearchTechnique)
	assert.Equal(t, kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSearchTechniqueExact, *cfg.SearchTechnique)
	require.NotNil(t, cfg.SelectedOptions)
	require.Len(t, *cfg.SelectedOptions, 1)
	require.NotNil(t, cfg.DisplaySettings)
	assert.Equal(t, "Pick one", *cfg.DisplaySettings.Placeholder)
	require.NotNil(t, cfg.Sort)
	assert.Equal(t, kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSortBy("_count"), cfg.Sort.By)
	assert.Equal(t, kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSortDirection("desc"), cfg.Sort.Direction)
}

// Test: round-trip — build then populate returns identical state (by_esql).
func Test_optionsListControl_byEsql_roundTrip(t *testing.T) {
	original := &models.OptionsListControlConfigModel{
		ByEsql: &models.OptionsListControlByEsqlModel{
			EsqlQuery:        types.StringValue("FROM logs | STATS BY status"),
			ValuesSource:     types.StringValue("esql_query"),
			SearchTechnique:  types.StringValue("prefix"),
			SingleSelect:     types.BoolValue(true),
			UseGlobalFilters: types.BoolValue(false),
		},
	}

	pm := models.PanelModel{OptionsListControlConfig: original}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	BuildConfig(pm, &olPanel)

	out := &models.PanelModel{OptionsListControlConfig: &models.OptionsListControlConfigModel{
		ByEsql: &models.OptionsListControlByEsqlModel{
			EsqlQuery:        types.StringValue("FROM logs | STATS BY status"),
			ValuesSource:     types.StringValue("esql_query"),
			SearchTechnique:  types.StringValue("prefix"),
			SingleSelect:     types.BoolValue(true),
			UseGlobalFilters: types.BoolValue(false),
		},
	}}
	tfPanel := &models.PanelModel{OptionsListControlConfig: out.OptionsListControlConfig}
	PopulateFromAPI(out, tfPanel, &olPanel)

	require.NotNil(t, out.OptionsListControlConfig)
	require.NotNil(t, out.OptionsListControlConfig.ByEsql)
	cfg := out.OptionsListControlConfig.ByEsql
	assert.Equal(t, types.StringValue("FROM logs | STATS BY status"), cfg.EsqlQuery)
	assert.Equal(t, types.StringValue("esql_query"), cfg.ValuesSource)
	assert.Equal(t, types.StringValue("prefix"), cfg.SearchTechnique)
	assert.Equal(t, types.BoolValue(true), cfg.SingleSelect)
	assert.Equal(t, types.BoolValue(false), cfg.UseGlobalFilters)
}

// Test: selectedOptionsEsqlToList converts string items.
func Test_selectedOptionsEsqlToList_stringItems(t *testing.T) {
	var item1 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql_SelectedOptions_Item
	require.NoError(t, item1.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSelectedOptions0("alpha"))
	var item2 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql_SelectedOptions_Item
	require.NoError(t, item2.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSelectedOptions0("beta"))

	result := selectedOptionsEsqlToList([]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql_SelectedOptions_Item{item1, item2})
	require.False(t, result.IsNull())
	elems := result.Elements()
	require.Len(t, elems, 2)
	assert.Equal(t, types.StringValue("alpha"), elems[0])
	assert.Equal(t, types.StringValue("beta"), elems[1])
}

// Test: selectedOptionsEsqlToList converts numeric items using fixed-point notation.
func Test_selectedOptionsEsqlToList_numericItems(t *testing.T) {
	var item1 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql_SelectedOptions_Item
	require.NoError(t, item1.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSelectedOptions1(1000000))
	var item2 kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql_SelectedOptions_Item
	require.NoError(t, item2.FromKibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsqlSelectedOptions1(3.14))

	result := selectedOptionsEsqlToList([]kbapi.KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchemaEsql_SelectedOptions_Item{item1, item2})
	require.False(t, result.IsNull())
	elems := result.Elements()
	require.Len(t, elems, 2)
	assert.Equal(t, types.StringValue("1000000"), elems[0])
	assert.Equal(t, types.StringValue("3.14"), elems[1])
}

// Test: BuildConfig with a nil OptionsListControlConfig is a no-op (no diagnostics, no panic).
func Test_BuildConfig_nilConfig_noop(t *testing.T) {
	pm := models.PanelModel{}
	olPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl{}
	diags := BuildConfig(pm, &olPanel)
	assert.False(t, diags.HasError())
}
