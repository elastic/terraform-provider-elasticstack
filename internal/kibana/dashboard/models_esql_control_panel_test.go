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
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func esqlTestStringList(t *testing.T, vals ...string) types.List {
	t.Helper()
	var diags diag.Diagnostics
	l := typeutils.SliceToListTypeString(context.Background(), vals, path.Empty(), &diags)
	require.False(t, diags.HasError())
	return l
}

func Test_esqlControlConfigModel_roundTrip(t *testing.T) {
	ctx := context.Background()
	truePtr := func(b bool) *bool { return &b }
	strPtr := func(s string) *string { return &s }

	base := kbapi.KbnDashboardPanelEsqlControl_Config{
		Title:            strPtr("Control title"),
		ControlType:      kbapi.KbnDashboardPanelEsqlControlConfigControlTypeSTATICVALUES,
		VariableName:     "my_var",
		VariableType:     kbapi.KbnDashboardPanelEsqlControlConfigVariableTypeValues,
		EsqlQuery:        `ROW x = "opt1"`,
		SelectedOptions:  []string{"opt1"},
		SingleSelect:     truePtr(true),
		AvailableOptions: &[]string{"opt1", "opt2"},
		DisplaySettings: &struct {
			HideActionBar *bool   `json:"hide_action_bar,omitempty"`
			HideExclude   *bool   `json:"hide_exclude,omitempty"`
			HideExists    *bool   `json:"hide_exists,omitempty"`
			HideSort      *bool   `json:"hide_sort,omitempty"`
			Placeholder   *string `json:"placeholder,omitempty"`
		}{
			Placeholder:   strPtr("Choose…"),
			HideActionBar: truePtr(true),
			HideExclude:   truePtr(false),
			HideExists:    truePtr(true),
			HideSort:      truePtr(false),
		},
	}

	var m esqlControlConfigModel
	require.False(t, m.fromAPI(ctx, base).HasError())

	out, toDiags := m.toAPI(ctx)
	require.False(t, toDiags.HasError())

	gotJSON, err := json.Marshal(out)
	require.NoError(t, err)
	wantJSON, err := json.Marshal(base)
	require.NoError(t, err)

	var gotMap, wantMap map[string]any
	require.NoError(t, json.Unmarshal(gotJSON, &gotMap))
	require.NoError(t, json.Unmarshal(wantJSON, &wantMap))
	assert.Equal(t, wantMap, gotMap)
}

func Test_esqlControlConfigModel_variableTypes_roundTrip(t *testing.T) {
	ctx := context.Background()
	for _, vt := range []kbapi.KbnDashboardPanelEsqlControlConfigVariableType{
		kbapi.KbnDashboardPanelEsqlControlConfigVariableTypeFields,
		kbapi.KbnDashboardPanelEsqlControlConfigVariableTypeValues,
		kbapi.KbnDashboardPanelEsqlControlConfigVariableTypeFunctions,
		kbapi.KbnDashboardPanelEsqlControlConfigVariableTypeTimeLiteral,
		kbapi.KbnDashboardPanelEsqlControlConfigVariableTypeMultiValues,
	} {
		t.Run(string(vt), func(t *testing.T) {
			api := kbapi.KbnDashboardPanelEsqlControl_Config{
				ControlType:     kbapi.KbnDashboardPanelEsqlControlConfigControlTypeVALUESFROMQUERY,
				VariableName:    "v",
				VariableType:    vt,
				EsqlQuery:       "FROM logs-* | LIMIT 1",
				SelectedOptions: []string{},
			}
			var m esqlControlConfigModel
			require.False(t, m.fromAPI(ctx, api).HasError())
			out, d := m.toAPI(ctx)
			require.False(t, d.HasError())
			assert.Equal(t, vt, out.VariableType)
			assert.Equal(t, kbapi.KbnDashboardPanelEsqlControlConfigControlTypeVALUESFROMQUERY, out.ControlType)
		})
	}
}

func Test_esqlControlConfigModel_empty_selected_options(t *testing.T) {
	ctx := context.Background()
	api := kbapi.KbnDashboardPanelEsqlControl_Config{
		ControlType:     kbapi.KbnDashboardPanelEsqlControlConfigControlTypeSTATICVALUES,
		VariableName:    "v",
		VariableType:    kbapi.KbnDashboardPanelEsqlControlConfigVariableTypeValues,
		EsqlQuery:       `ROW a = 1`,
		SelectedOptions: []string{},
	}
	var m esqlControlConfigModel
	require.False(t, m.fromAPI(ctx, api).HasError())
	out, d := m.toAPI(ctx)
	require.False(t, d.HasError())
	assert.Equal(t, []string{}, out.SelectedOptions)
}

func Test_panelModel_esql_control_toAPI(t *testing.T) {
	pm := panelModel{
		Type: types.StringValue("esql_control"),
		Grid: panelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
			W: types.Int64Value(12),
			H: types.Int64Value(8),
		},
		ID: types.StringValue("esql-panel-1"),
		EsqlControlConfig: &esqlControlConfigModel{
			Title:            types.StringValue("Static pick"),
			VariableName:     types.StringValue("bucket"),
			VariableType:     types.StringValue("values"),
			EsqlQuery:        types.StringValue(`ROW x = "a"`),
			ControlType:      types.StringValue("STATIC_VALUES"),
			SelectedOptions:  esqlTestStringList(t, "a"),
			AvailableOptions: esqlTestStringList(t, "a", "b"),
			SingleSelect:     types.BoolValue(true),
			DisplaySettings: &esqlControlDisplaySettingsModel{
				Placeholder:   types.StringValue("hint"),
				HideActionBar: types.BoolValue(true),
				HideExclude:   types.BoolNull(),
				HideExists:    types.BoolNull(),
				HideSort:      types.BoolNull(),
			},
		},
	}
	item, diags := pm.toAPI(context.Background())
	require.False(t, diags.HasError())
	esql, err := item.AsKbnDashboardPanelEsqlControl()
	require.NoError(t, err)
	assert.InDelta(t, 0.0, float64(esql.Grid.X), 0.001)
	assert.Equal(t, "bucket", esql.Config.VariableName)
	assert.Equal(t, kbapi.KbnDashboardPanelEsqlControlConfigControlTypeSTATICVALUES, esql.Config.ControlType)
	require.NotNil(t, esql.Config.AvailableOptions)
	assert.Equal(t, []string{"a", "b"}, *esql.Config.AvailableOptions)
	require.NotNil(t, esql.Config.DisplaySettings)
	require.NotNil(t, esql.Config.DisplaySettings.Placeholder)
	assert.Equal(t, "hint", *esql.Config.DisplaySettings.Placeholder)
}

func Test_mapPanelsFromAPI_esql_control(t *testing.T) {
	const apiPanelsJSON = `[
		{
			"grid": {"x": 0, "y": 2, "w": 24, "h": 10},
			"uid": "esql-panel-uid",
			"type": "esql_control",
			"config": {
				"title": "Variable pick",
				"variable_name": "bucket",
				"variable_type": "values",
				"esql_query": "ROW n = 1",
				"control_type": "STATIC_VALUES",
				"selected_options": ["a"],
				"available_options": ["a", "b"],
				"single_select": true,
				"display_settings": {
					"placeholder": "Select value",
					"hide_action_bar": true
				}
			}
		}
	]`
	var apiPanels kbapi.DashboardPanels
	require.NoError(t, json.Unmarshal([]byte(apiPanelsJSON), &apiPanels))

	item, err := apiPanels[0].AsDashboardPanelItem()
	require.NoError(t, err)
	esqlIn, err := item.AsKbnDashboardPanelEsqlControl()
	require.NoError(t, err)
	cfgBytes, err := esqlIn.Config.MarshalJSON()
	require.NoError(t, err)
	wantConfigJSON := customtypes.NewJSONWithDefaultsValue(string(cfgBytes), populatePanelConfigJSONDefaults)

	model := &dashboardModel{}
	panels, _, diags := model.mapPanelsFromAPI(t.Context(), &apiPanels)
	require.False(t, diags.HasError())
	require.Len(t, panels, 1)

	pm := panels[0]
	assert.Equal(t, "esql_control", pm.Type.ValueString())
	assert.Equal(t, int64(0), pm.Grid.X.ValueInt64())
	assert.Equal(t, int64(2), pm.Grid.Y.ValueInt64())
	assert.Equal(t, "esql-panel-uid", pm.ID.ValueString())
	require.NotNil(t, pm.EsqlControlConfig)
	assert.Equal(t, "Variable pick", pm.EsqlControlConfig.Title.ValueString())
	assert.Equal(t, "bucket", pm.EsqlControlConfig.VariableName.ValueString())
	assert.Equal(t, "values", pm.EsqlControlConfig.VariableType.ValueString())
	assert.Equal(t, "ROW n = 1", pm.EsqlControlConfig.EsqlQuery.ValueString())
	assert.Equal(t, "STATIC_VALUES", pm.EsqlControlConfig.ControlType.ValueString())
	assert.True(t, pm.EsqlControlConfig.SingleSelect.ValueBool())
	require.NotNil(t, pm.EsqlControlConfig.DisplaySettings)
	assert.Equal(t, "Select value", pm.EsqlControlConfig.DisplaySettings.Placeholder.ValueString())
	assert.True(t, pm.EsqlControlConfig.DisplaySettings.HideActionBar.ValueBool())

	eq, cjDiags := wantConfigJSON.StringSemanticEquals(context.Background(), pm.ConfigJSON)
	require.False(t, cjDiags.HasError())
	assert.True(t, eq, "ConfigJSON mismatch")
}
