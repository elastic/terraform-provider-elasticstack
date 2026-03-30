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

func minimalEsqlAPIConfig() kbapi.KbnDashboardPanelEsqlControl_Config {
	return kbapi.KbnDashboardPanelEsqlControl_Config{
		SelectedOptions: []string{"opt_a"},
		VariableName:    "my_var",
		VariableType:    "values",
		EsqlQuery:       "FROM logs-*",
		ControlType:     "STATIC_VALUES",
	}
}

// Test: on import (tfPanel == nil) populate all fields from API.
func Test_populateEsqlControlFromAPI_import_populatesAllFields(t *testing.T) {
	cfg := minimalEsqlAPIConfig()
	cfg.Title = new("My Control")
	cfg.SingleSelect = new(true)
	opts := []string{"a", "b"}
	cfg.AvailableOptions = &opts
	cfg.DisplaySettings = &struct {
		HideActionBar *bool   `json:"hide_action_bar,omitempty"`
		HideExclude   *bool   `json:"hide_exclude,omitempty"`
		HideExists    *bool   `json:"hide_exists,omitempty"`
		HideSort      *bool   `json:"hide_sort,omitempty"`
		Placeholder   *string `json:"placeholder,omitempty"`
	}{
		Placeholder:   new("Select..."),
		HideActionBar: new(true),
	}

	pm := &panelModel{}
	populateEsqlControlFromAPI(pm, nil, cfg)

	require.NotNil(t, pm.EsqlControlConfig)
	assert.Equal(t, stringsToList([]string{"opt_a"}), pm.EsqlControlConfig.SelectedOptions)
	assert.Equal(t, types.StringValue("my_var"), pm.EsqlControlConfig.VariableName)
	assert.Equal(t, types.StringValue("values"), pm.EsqlControlConfig.VariableType)
	assert.Equal(t, types.StringValue("FROM logs-*"), pm.EsqlControlConfig.EsqlQuery)
	assert.Equal(t, types.StringValue("STATIC_VALUES"), pm.EsqlControlConfig.ControlType)
	assert.Equal(t, types.StringValue("My Control"), pm.EsqlControlConfig.Title)
	assert.Equal(t, types.BoolValue(true), pm.EsqlControlConfig.SingleSelect)
	assert.Equal(t, stringsToList([]string{"a", "b"}), pm.EsqlControlConfig.AvailableOptions)
	require.NotNil(t, pm.EsqlControlConfig.DisplaySettings)
	assert.Equal(t, types.StringValue("Select..."), pm.EsqlControlConfig.DisplaySettings.Placeholder)
	assert.Equal(t, types.BoolValue(true), pm.EsqlControlConfig.DisplaySettings.HideActionBar)
}

// Test: when existing config block is nil, preserve nil intent.
func Test_populateEsqlControlFromAPI_nilBlock_preservesNil(t *testing.T) {
	pm := &panelModel{}
	tfPanel := &panelModel{}
	populateEsqlControlFromAPI(pm, tfPanel, minimalEsqlAPIConfig())
	assert.Nil(t, pm.EsqlControlConfig)
}

// Test: when config block exists, required fields are updated from API.
func Test_populateEsqlControlFromAPI_existingBlock_requiredFieldsUpdated(t *testing.T) {
	pm := &panelModel{
		EsqlControlConfig: &esqlControlConfigModel{
			SelectedOptions:  stringsToList([]string{"old"}),
			VariableName:     types.StringValue("old_var"),
			VariableType:     types.StringValue("fields"),
			EsqlQuery:        types.StringValue("FROM old-*"),
			ControlType:      types.StringValue("STATIC_VALUES"),
			AvailableOptions: types.ListNull(types.StringType),
		},
	}
	tfPanel := &panelModel{EsqlControlConfig: pm.EsqlControlConfig}
	populateEsqlControlFromAPI(pm, tfPanel, minimalEsqlAPIConfig())

	require.NotNil(t, pm.EsqlControlConfig)
	assert.Equal(t, stringsToList([]string{"opt_a"}), pm.EsqlControlConfig.SelectedOptions)
	assert.Equal(t, types.StringValue("my_var"), pm.EsqlControlConfig.VariableName)
	assert.Equal(t, types.StringValue("values"), pm.EsqlControlConfig.VariableType)
	assert.Equal(t, types.StringValue("FROM logs-*"), pm.EsqlControlConfig.EsqlQuery)
	assert.Equal(t, types.StringValue("STATIC_VALUES"), pm.EsqlControlConfig.ControlType)
}

// Test: null-preservation — null optional fields in state are not overwritten by API values.
func Test_populateEsqlControlFromAPI_nullOptionalFields_preserved(t *testing.T) {
	pm := &panelModel{
		EsqlControlConfig: &esqlControlConfigModel{
			SelectedOptions:  stringsToList([]string{}),
			VariableName:     types.StringValue("my_var"),
			VariableType:     types.StringValue("values"),
			EsqlQuery:        types.StringValue("FROM logs-*"),
			ControlType:      types.StringValue("STATIC_VALUES"),
			Title:            types.StringNull(),
			SingleSelect:     types.BoolNull(),
			AvailableOptions: types.ListNull(types.StringType),
		},
	}
	tfPanel := &panelModel{EsqlControlConfig: pm.EsqlControlConfig}
	cfg := minimalEsqlAPIConfig()
	cfg.Title = new("API Title")
	cfg.SingleSelect = new(true)
	populateEsqlControlFromAPI(pm, tfPanel, cfg)

	require.NotNil(t, pm.EsqlControlConfig)
	assert.True(t, pm.EsqlControlConfig.Title.IsNull())
	assert.True(t, pm.EsqlControlConfig.SingleSelect.IsNull())
}

// Test: null display_settings block preserved when API returns display_settings.
func Test_populateEsqlControlFromAPI_nilDisplaySettings_preserved(t *testing.T) {
	pm := &panelModel{
		EsqlControlConfig: &esqlControlConfigModel{
			SelectedOptions:  stringsToList([]string{}),
			VariableName:     types.StringValue("v"),
			VariableType:     types.StringValue("values"),
			EsqlQuery:        types.StringValue("FROM logs-*"),
			ControlType:      types.StringValue("STATIC_VALUES"),
			AvailableOptions: types.ListNull(types.StringType),
			DisplaySettings:  nil,
		},
	}
	tfPanel := &panelModel{EsqlControlConfig: pm.EsqlControlConfig}
	cfg := minimalEsqlAPIConfig()
	cfg.DisplaySettings = &struct {
		HideActionBar *bool   `json:"hide_action_bar,omitempty"`
		HideExclude   *bool   `json:"hide_exclude,omitempty"`
		HideExists    *bool   `json:"hide_exists,omitempty"`
		HideSort      *bool   `json:"hide_sort,omitempty"`
		Placeholder   *string `json:"placeholder,omitempty"`
	}{Placeholder: new("hint")}
	populateEsqlControlFromAPI(pm, tfPanel, cfg)
	assert.Nil(t, pm.EsqlControlConfig.DisplaySettings)
}

// Test: display_settings null fields within existing block are preserved.
func Test_populateEsqlControlFromAPI_displaySettings_nullFieldsPreserved(t *testing.T) {
	pm := &panelModel{
		EsqlControlConfig: &esqlControlConfigModel{
			SelectedOptions:  stringsToList([]string{}),
			VariableName:     types.StringValue("v"),
			VariableType:     types.StringValue("values"),
			EsqlQuery:        types.StringValue("FROM logs-*"),
			ControlType:      types.StringValue("STATIC_VALUES"),
			AvailableOptions: types.ListNull(types.StringType),
			DisplaySettings: &esqlControlDisplaySettingsModel{
				Placeholder:   types.StringNull(),
				HideActionBar: types.BoolValue(true),
			},
		},
	}
	tfPanel := &panelModel{EsqlControlConfig: pm.EsqlControlConfig}
	cfg := minimalEsqlAPIConfig()
	cfg.DisplaySettings = &struct {
		HideActionBar *bool   `json:"hide_action_bar,omitempty"`
		HideExclude   *bool   `json:"hide_exclude,omitempty"`
		HideExists    *bool   `json:"hide_exists,omitempty"`
		HideSort      *bool   `json:"hide_sort,omitempty"`
		Placeholder   *string `json:"placeholder,omitempty"`
	}{
		Placeholder:   new("hint"),
		HideActionBar: new(false),
	}
	populateEsqlControlFromAPI(pm, tfPanel, cfg)

	require.NotNil(t, pm.EsqlControlConfig.DisplaySettings)
	// null field stays null
	assert.True(t, pm.EsqlControlConfig.DisplaySettings.Placeholder.IsNull())
	// known field is updated
	assert.Equal(t, types.BoolValue(false), pm.EsqlControlConfig.DisplaySettings.HideActionBar)
}

// Test: buildEsqlControlConfig writes required fields to API struct.
func Test_buildEsqlControlConfig_requiredFields(t *testing.T) {
	pm := panelModel{
		EsqlControlConfig: &esqlControlConfigModel{
			SelectedOptions:  stringsToList([]string{"opt1", "opt2"}),
			VariableName:     types.StringValue("my_var"),
			VariableType:     types.StringValue("values"),
			EsqlQuery:        types.StringValue("FROM logs-*"),
			ControlType:      types.StringValue("STATIC_VALUES"),
			AvailableOptions: types.ListNull(types.StringType),
		},
	}
	esqlPanel := kbapi.KbnDashboardPanelEsqlControl{}
	buildEsqlControlConfig(pm, &esqlPanel)

	assert.Equal(t, []string{"opt1", "opt2"}, esqlPanel.Config.SelectedOptions)
	assert.Equal(t, "my_var", esqlPanel.Config.VariableName)
	assert.Equal(t, kbapi.KbnDashboardPanelEsqlControlConfigVariableType("values"), esqlPanel.Config.VariableType)
	assert.Equal(t, "FROM logs-*", esqlPanel.Config.EsqlQuery)
	assert.Equal(t, kbapi.KbnDashboardPanelEsqlControlConfigControlType("STATIC_VALUES"), esqlPanel.Config.ControlType)
	assert.Nil(t, esqlPanel.Config.Title)
	assert.Nil(t, esqlPanel.Config.SingleSelect)
	assert.Nil(t, esqlPanel.Config.AvailableOptions)
	assert.Nil(t, esqlPanel.Config.DisplaySettings)
}

// Test: buildEsqlControlConfig writes optional fields when set.
func Test_buildEsqlControlConfig_optionalFields(t *testing.T) {
	pm := panelModel{
		EsqlControlConfig: &esqlControlConfigModel{
			SelectedOptions:  stringsToList([]string{}),
			VariableName:     types.StringValue("v"),
			VariableType:     types.StringValue("fields"),
			EsqlQuery:        types.StringValue("FROM *"),
			ControlType:      types.StringValue("VALUES_FROM_QUERY"),
			Title:            types.StringValue("My Control"),
			SingleSelect:     types.BoolValue(true),
			AvailableOptions: stringsToList([]string{"a"}),
			DisplaySettings: &esqlControlDisplaySettingsModel{
				Placeholder:   types.StringValue("hint"),
				HideActionBar: types.BoolValue(true),
				HideExclude:   types.BoolNull(),
				HideExists:    types.BoolNull(),
				HideSort:      types.BoolNull(),
			},
		},
	}
	esqlPanel := kbapi.KbnDashboardPanelEsqlControl{}
	buildEsqlControlConfig(pm, &esqlPanel)

	require.NotNil(t, esqlPanel.Config.Title)
	assert.Equal(t, "My Control", *esqlPanel.Config.Title)
	require.NotNil(t, esqlPanel.Config.SingleSelect)
	assert.True(t, *esqlPanel.Config.SingleSelect)
	require.NotNil(t, esqlPanel.Config.AvailableOptions)
	assert.Equal(t, []string{"a"}, *esqlPanel.Config.AvailableOptions)
	require.NotNil(t, esqlPanel.Config.DisplaySettings)
	require.NotNil(t, esqlPanel.Config.DisplaySettings.Placeholder)
	assert.Equal(t, "hint", *esqlPanel.Config.DisplaySettings.Placeholder)
	require.NotNil(t, esqlPanel.Config.DisplaySettings.HideActionBar)
	assert.True(t, *esqlPanel.Config.DisplaySettings.HideActionBar)
	assert.Nil(t, esqlPanel.Config.DisplaySettings.HideExclude)
	assert.Nil(t, esqlPanel.Config.DisplaySettings.HideExists)
	assert.Nil(t, esqlPanel.Config.DisplaySettings.HideSort)
}

// Test: buildEsqlControlConfig omits nil optional fields.
func Test_buildEsqlControlConfig_nullOptionalFields_omitted(t *testing.T) {
	pm := panelModel{
		EsqlControlConfig: &esqlControlConfigModel{
			SelectedOptions:  stringsToList([]string{}),
			VariableName:     types.StringValue("v"),
			VariableType:     types.StringValue("values"),
			EsqlQuery:        types.StringValue("FROM *"),
			ControlType:      types.StringValue("STATIC_VALUES"),
			Title:            types.StringNull(),
			SingleSelect:     types.BoolNull(),
			AvailableOptions: types.ListNull(types.StringType),
		},
	}
	esqlPanel := kbapi.KbnDashboardPanelEsqlControl{}
	buildEsqlControlConfig(pm, &esqlPanel)

	assert.Nil(t, esqlPanel.Config.Title)
	assert.Nil(t, esqlPanel.Config.SingleSelect)
	assert.Nil(t, esqlPanel.Config.AvailableOptions)
	assert.Nil(t, esqlPanel.Config.DisplaySettings)
}

// Test: round-trip — set values, build to API, populate back, same values.
func Test_esqlControl_roundTrip(t *testing.T) {
	original := &esqlControlConfigModel{
		SelectedOptions:  stringsToList([]string{"opt_a", "opt_b"}),
		VariableName:     types.StringValue("my_var"),
		VariableType:     types.StringValue("values"),
		EsqlQuery:        types.StringValue("FROM logs-* | STATS count = COUNT(*) BY host.name"),
		ControlType:      types.StringValue("STATIC_VALUES"),
		Title:            types.StringValue("My Control"),
		SingleSelect:     types.BoolValue(false),
		AvailableOptions: types.ListNull(types.StringType),
	}
	pm := panelModel{EsqlControlConfig: original}

	esqlPanel := kbapi.KbnDashboardPanelEsqlControl{}
	buildEsqlControlConfig(pm, &esqlPanel)

	out := &panelModel{EsqlControlConfig: &esqlControlConfigModel{
		SelectedOptions:  original.SelectedOptions,
		VariableName:     original.VariableName,
		VariableType:     original.VariableType,
		EsqlQuery:        original.EsqlQuery,
		ControlType:      original.ControlType,
		Title:            original.Title,
		SingleSelect:     original.SingleSelect,
		AvailableOptions: original.AvailableOptions,
	}}
	tfPanel := &panelModel{EsqlControlConfig: out.EsqlControlConfig}
	populateEsqlControlFromAPI(out, tfPanel, esqlPanel.Config)

	require.NotNil(t, out.EsqlControlConfig)
	assert.Equal(t, original.SelectedOptions, out.EsqlControlConfig.SelectedOptions)
	assert.Equal(t, original.VariableName, out.EsqlControlConfig.VariableName)
	assert.Equal(t, original.VariableType, out.EsqlControlConfig.VariableType)
	assert.Equal(t, original.EsqlQuery, out.EsqlControlConfig.EsqlQuery)
	assert.Equal(t, original.ControlType, out.EsqlControlConfig.ControlType)
	assert.Equal(t, original.Title, out.EsqlControlConfig.Title)
	assert.Equal(t, original.SingleSelect, out.EsqlControlConfig.SingleSelect)
}

