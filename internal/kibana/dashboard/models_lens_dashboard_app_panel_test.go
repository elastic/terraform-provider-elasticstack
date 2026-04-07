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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Write-path (toAPI) tests ---

func Test_lensDashboardApp_byValue_writeConverter(t *testing.T) {
	attrsJSON := `{"visualizationType":"lnsXY","title":"My Viz"}`
	refsJSON := `[{"id":"dv-1","name":"indexpattern-datasource-layer-abc","type":"index-pattern"}]`

	pm := panelModel{
		Type: types.StringValue(panelTypeLensDashboardApp),
		Grid: panelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
		},
		LensDashboardAppConfig: &lensDashboardAppConfigModel{
			ByValue: &lensDashboardAppByValueModel{
				AttributesJSON: jsontypes.NewNormalizedValue(attrsJSON),
				ReferencesJSON: jsontypes.NewNormalizedValue(refsJSON),
			},
			Title:       types.StringValue("My Panel"),
			Description: types.StringNull(),
			HideTitle:   types.BoolNull(),
			HideBorder:  types.BoolNull(),
		},
	}

	panelItem, diags := pm.toAPI()
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)

	ldaPanel, err := panelItem.AsKbnDashboardPanelLensDashboardApp()
	require.NoError(t, err)

	config0, err := ldaPanel.Config.AsKbnDashboardPanelLensDashboardAppConfig0()
	require.NoError(t, err)

	// attributes should be present
	attrsBytes, err := json.Marshal(config0.Attributes)
	require.NoError(t, err)
	assert.JSONEq(t, attrsJSON, string(attrsBytes))

	// references should be present
	require.NotNil(t, config0.References)
	assert.Len(t, *config0.References, 1)
	assert.Equal(t, "dv-1", (*config0.References)[0].Id)
	assert.Equal(t, "indexpattern-datasource-layer-abc", (*config0.References)[0].Name)
	assert.Equal(t, "index-pattern", (*config0.References)[0].Type)

	// title should be set
	require.NotNil(t, config0.Title)
	assert.Equal(t, "My Panel", *config0.Title)

	// null optional fields should not be set
	assert.Nil(t, config0.Description)
	assert.Nil(t, config0.HideTitle)
	assert.Nil(t, config0.HideBorder)
}

func Test_lensDashboardApp_byValue_noReferences_writeConverter(t *testing.T) {
	attrsJSON := `{"visualizationType":"lnsMetric"}`

	pm := panelModel{
		Type: types.StringValue(panelTypeLensDashboardApp),
		Grid: panelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
		},
		LensDashboardAppConfig: &lensDashboardAppConfigModel{
			ByValue: &lensDashboardAppByValueModel{
				AttributesJSON: jsontypes.NewNormalizedValue(attrsJSON),
				ReferencesJSON: jsontypes.NewNormalizedNull(),
			},
		},
	}

	panelItem, diags := pm.toAPI()
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)

	ldaPanel, err := panelItem.AsKbnDashboardPanelLensDashboardApp()
	require.NoError(t, err)

	config0, err := ldaPanel.Config.AsKbnDashboardPanelLensDashboardAppConfig0()
	require.NoError(t, err)

	// references should not be set when null
	assert.Nil(t, config0.References)
}

func Test_lensDashboardApp_byReference_writeConverter(t *testing.T) {
	pm := panelModel{
		Type: types.StringValue(panelTypeLensDashboardApp),
		Grid: panelGridModel{
			X: types.Int64Value(2),
			Y: types.Int64Value(3),
		},
		LensDashboardAppConfig: &lensDashboardAppConfigModel{
			ByReference: &lensDashboardAppByReferenceModel{
				SavedObjectID: types.StringValue("abc-123"),
				OverridesJSON: jsontypes.NewNormalizedNull(),
			},
			Title:       types.StringValue("My Shared Viz"),
			HideTitle:   types.BoolValue(true),
			HideBorder:  types.BoolValue(false),
			Description: types.StringNull(),
		},
	}

	panelItem, diags := pm.toAPI()
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)

	ldaPanel, err := panelItem.AsKbnDashboardPanelLensDashboardApp()
	require.NoError(t, err)

	config1, err := ldaPanel.Config.AsKbnDashboardPanelLensDashboardAppConfig1()
	require.NoError(t, err)

	assert.Equal(t, "abc-123", config1.RefId)
	require.NotNil(t, config1.Title)
	assert.Equal(t, "My Shared Viz", *config1.Title)
	require.NotNil(t, config1.HideTitle)
	assert.True(t, *config1.HideTitle)
	require.NotNil(t, config1.HideBorder)
	assert.False(t, *config1.HideBorder)
	assert.Nil(t, config1.Description)
}

func Test_lensDashboardApp_byValue_withTimeRange_writeConverter(t *testing.T) {
	attrsJSON := `{"visualizationType":"lnsXY"}`

	pm := panelModel{
		Type: types.StringValue(panelTypeLensDashboardApp),
		Grid: panelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
		},
		LensDashboardAppConfig: &lensDashboardAppConfigModel{
			ByValue: &lensDashboardAppByValueModel{
				AttributesJSON: jsontypes.NewNormalizedValue(attrsJSON),
				ReferencesJSON: jsontypes.NewNormalizedNull(),
			},
			TimeRange: &lensDashboardAppTimeRangeModel{
				From: types.StringValue("now-7d"),
				To:   types.StringValue("now"),
			},
		},
	}

	panelItem, diags := pm.toAPI()
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)

	ldaPanel, err := panelItem.AsKbnDashboardPanelLensDashboardApp()
	require.NoError(t, err)

	config0, err := ldaPanel.Config.AsKbnDashboardPanelLensDashboardAppConfig0()
	require.NoError(t, err)

	assert.Equal(t, "now-7d", config0.TimeRange.From)
	assert.Equal(t, "now", config0.TimeRange.To)
}

// --- Read-path (fromAPI) tests ---

func buildLensDashboardAppByValueAPIPanel(attrsJSON string, refsJSON *string, title *string) kbapi.KbnDashboardPanelLensDashboardApp {
	var attrs kbapi.LensApiState
	_ = json.Unmarshal([]byte(attrsJSON), &attrs)

	config0 := kbapi.KbnDashboardPanelLensDashboardAppConfig0{
		Attributes: attrs,
		Title:      title,
	}
	if refsJSON != nil {
		var refs []kbapi.KbnContentManagementUtilsReferenceSchema
		_ = json.Unmarshal([]byte(*refsJSON), &refs)
		config0.References = &refs
	}

	var cfg kbapi.KbnDashboardPanelLensDashboardApp_Config
	_ = cfg.FromKbnDashboardPanelLensDashboardAppConfig0(config0)

	return kbapi.KbnDashboardPanelLensDashboardApp{
		Config: cfg,
		Grid: struct {
			H *float32 `json:"h,omitempty"`
			W *float32 `json:"w,omitempty"`
			X float32  `json:"x"`
			Y float32  `json:"y"`
		}{X: 0, Y: 0},
		Type: kbapi.LensDashboardApp,
	}
}

func buildLensDashboardAppByReferenceAPIPanel(refID string, title *string) kbapi.KbnDashboardPanelLensDashboardApp {
	config1 := kbapi.KbnDashboardPanelLensDashboardAppConfig1{
		RefId: refID,
		Title: title,
	}

	var cfg kbapi.KbnDashboardPanelLensDashboardApp_Config
	_ = cfg.FromKbnDashboardPanelLensDashboardAppConfig1(config1)

	return kbapi.KbnDashboardPanelLensDashboardApp{
		Config: cfg,
		Grid: struct {
			H *float32 `json:"h,omitempty"`
			W *float32 `json:"w,omitempty"`
			X float32  `json:"x"`
			Y float32  `json:"y"`
		}{X: 0, Y: 0},
		Type: kbapi.LensDashboardApp,
	}
}

func Test_lensDashboardApp_readConverter_byValue(t *testing.T) {
	attrsJSON := `{"visualizationType":"lnsXY","title":"My Viz"}`
	title := "Panel Title"
	apiPanel := buildLensDashboardAppByValueAPIPanel(attrsJSON, nil, &title)

	pm := panelModel{}
	diags := populateLensDashboardAppFromAPI(&pm, nil, apiPanel)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)

	require.NotNil(t, pm.LensDashboardAppConfig)
	cfg := pm.LensDashboardAppConfig

	// Should be by-value
	require.NotNil(t, cfg.ByValue)
	assert.Nil(t, cfg.ByReference)

	// attributes_json should be populated
	assert.False(t, cfg.ByValue.AttributesJSON.IsNull())
	assert.JSONEq(t, attrsJSON, cfg.ByValue.AttributesJSON.ValueString())

	// title should be populated
	assert.Equal(t, "Panel Title", cfg.Title.ValueString())
}

func Test_lensDashboardApp_readConverter_byValue_withReferences(t *testing.T) {
	attrsJSON := `{"visualizationType":"lnsXY"}`
	refsJSON := `[{"id":"dv-1","name":"ref-name","type":"index-pattern"}]`
	apiPanel := buildLensDashboardAppByValueAPIPanel(attrsJSON, &refsJSON, nil)

	pm := panelModel{}
	diags := populateLensDashboardAppFromAPI(&pm, nil, apiPanel)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)

	require.NotNil(t, pm.LensDashboardAppConfig)
	cfg := pm.LensDashboardAppConfig

	require.NotNil(t, cfg.ByValue)
	assert.False(t, cfg.ByValue.ReferencesJSON.IsNull())
	assert.JSONEq(t, refsJSON, cfg.ByValue.ReferencesJSON.ValueString())
}

func Test_lensDashboardApp_readConverter_byReference(t *testing.T) {
	title := "My Shared Viz"
	apiPanel := buildLensDashboardAppByReferenceAPIPanel("abc-123", &title)

	pm := panelModel{}
	diags := populateLensDashboardAppFromAPI(&pm, nil, apiPanel)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)

	require.NotNil(t, pm.LensDashboardAppConfig)
	cfg := pm.LensDashboardAppConfig

	// Should be by-reference
	assert.Nil(t, cfg.ByValue)
	require.NotNil(t, cfg.ByReference)

	assert.Equal(t, "abc-123", cfg.ByReference.SavedObjectID.ValueString())
	assert.Equal(t, "My Shared Viz", cfg.Title.ValueString())
}

func Test_lensDashboardApp_readConverter_byReference_preservesAbsentOptionals(t *testing.T) {
	// Simulate read-back without title or time_range set in prior TF state
	apiPanel := buildLensDashboardAppByReferenceAPIPanel("xyz-456", nil)

	// Prior state: by-reference, no title, no time_range
	existingConfig := &lensDashboardAppConfigModel{
		ByReference: &lensDashboardAppByReferenceModel{
			SavedObjectID: types.StringValue("xyz-456"),
			OverridesJSON: jsontypes.NewNormalizedNull(),
		},
		Title:       types.StringNull(),
		Description: types.StringNull(),
		HideTitle:   types.BoolNull(),
		HideBorder:  types.BoolNull(),
		TimeRange:   nil,
	}
	tfPanel := &panelModel{
		LensDashboardAppConfig: existingConfig,
	}

	pm := panelModel{}
	diags := populateLensDashboardAppFromAPI(&pm, tfPanel, apiPanel)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)

	cfg := pm.LensDashboardAppConfig
	require.NotNil(t, cfg)

	// Optional fields should remain null per null-preservation semantics
	assert.True(t, cfg.Title.IsNull())
	assert.True(t, cfg.Description.IsNull())
	assert.True(t, cfg.HideTitle.IsNull())
	assert.True(t, cfg.HideBorder.IsNull())
	assert.Nil(t, cfg.TimeRange)
}

// --- config_json error for lens-dashboard-app ---

func Test_lensDashboardApp_configJSON_rejected(t *testing.T) {
	// Task 6.12: config_json on lens-dashboard-app panel type should return error diagnostic
	pm := panelModel{
		Type: types.StringValue(panelTypeLensDashboardApp),
		Grid: panelGridModel{
			X: types.Int64Value(0),
			Y: types.Int64Value(0),
		},
		ConfigJSON: customtypes.NewJSONWithDefaultsValue(`{"key":"value"}`, populatePanelConfigJSONDefaults),
	}

	_, diags := pm.toAPI()
	require.True(t, diags.HasError())
	require.Len(t, diags, 1)
	assert.Equal(t, "Unsupported panel type for config_json", diags[0].Summary())
	assert.Contains(t, diags[0].Detail(), "lens-dashboard-app")
}

// --- lensDashboardAppConfigModeValidator tests ---

func Test_lensDashboardAppConfigModeValidator_bothSet(t *testing.T) {
	diags := panelConfigValidateDiags(
		panelTypeLensDashboardApp,
		panelConfigValueState{},
		panelConfigValueState{},
		panelConfigValueState{},
		panelConfigValueState{},
		lensConfigStates(nil),
		panelConfigValueState{},
		panelConfigValueState{Set: true},
		nil,
	)
	require.False(t, diags.HasError(), "panelConfigValidateDiags accepts lens-dashboard-app with config set")
}

func Test_lensDashboardAppConfigModeValidator_neitherSet(t *testing.T) {
	diags := panelConfigValidateDiags(
		panelTypeLensDashboardApp,
		panelConfigValueState{},
		panelConfigValueState{},
		panelConfigValueState{},
		panelConfigValueState{},
		lensConfigStates(nil),
		panelConfigValueState{},
		panelConfigValueState{},
		nil,
	)
	require.True(t, diags.HasError())
	assert.Equal(t, "Missing lens-dashboard-app panel configuration", diags[0].Summary())
}
