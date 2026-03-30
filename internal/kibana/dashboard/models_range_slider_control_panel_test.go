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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func apiRangeSliderConfig(opts ...func(*kbapi.KbnDashboardPanelRangeSliderControl_Config)) kbapi.KbnDashboardPanelRangeSliderControl_Config {
	cfg := kbapi.KbnDashboardPanelRangeSliderControl_Config{
		DataViewId: "dv-1",
		FieldName:  "bytes",
	}
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}

func withTitle(t string) func(*kbapi.KbnDashboardPanelRangeSliderControl_Config) {
	return func(c *kbapi.KbnDashboardPanelRangeSliderControl_Config) { c.Title = &t }
}
func withUseGlobalFilters(b bool) func(*kbapi.KbnDashboardPanelRangeSliderControl_Config) {
	return func(c *kbapi.KbnDashboardPanelRangeSliderControl_Config) { c.UseGlobalFilters = &b }
}
func withIgnoreValidations(b bool) func(*kbapi.KbnDashboardPanelRangeSliderControl_Config) {
	return func(c *kbapi.KbnDashboardPanelRangeSliderControl_Config) { c.IgnoreValidations = &b }
}
func withValue(lo, hi string) func(*kbapi.KbnDashboardPanelRangeSliderControl_Config) {
	return func(c *kbapi.KbnDashboardPanelRangeSliderControl_Config) {
		v := []string{lo, hi}
		c.Value = &v
	}
}
func withStep(s float32) func(*kbapi.KbnDashboardPanelRangeSliderControl_Config) {
	return func(c *kbapi.KbnDashboardPanelRangeSliderControl_Config) { c.Step = &s }
}

func mustStringList(elems ...string) types.List {
	v, _ := types.ListValueFrom(context.Background(), types.StringType, elems)
	return v
}

// Test: on import (tfPanel == nil) with API data, populate all fields.
func Test_populateRangeSliderControlFromAPI_import_allFields(t *testing.T) {
	pm := &panelModel{}
	apiCfg := apiRangeSliderConfig(
		withTitle("My Control"),
		withUseGlobalFilters(true),
		withIgnoreValidations(false),
		withValue("100", "500"),
		withStep(10),
	)
	populateRangeSliderControlFromAPI(context.Background(), pm, nil, apiCfg)
	require.NotNil(t, pm.RangeSliderControlConfig)
	assert.Equal(t, types.StringValue("dv-1"), pm.RangeSliderControlConfig.DataViewID)
	assert.Equal(t, types.StringValue("bytes"), pm.RangeSliderControlConfig.FieldName)
	assert.Equal(t, types.StringValue("My Control"), pm.RangeSliderControlConfig.Title)
	assert.Equal(t, types.BoolValue(true), pm.RangeSliderControlConfig.UseGlobalFilters)
	assert.Equal(t, types.BoolValue(false), pm.RangeSliderControlConfig.IgnoreValidations)
	assert.Equal(t, mustStringList("100", "500"), pm.RangeSliderControlConfig.Value)
	assert.Equal(t, types.Float32Value(10), pm.RangeSliderControlConfig.Step)
}

// Test: on import (tfPanel == nil) with minimal API data (only required fields).
func Test_populateRangeSliderControlFromAPI_import_requiredOnly(t *testing.T) {
	pm := &panelModel{}
	populateRangeSliderControlFromAPI(context.Background(), pm, nil, apiRangeSliderConfig())
	require.NotNil(t, pm.RangeSliderControlConfig)
	assert.Equal(t, types.StringValue("dv-1"), pm.RangeSliderControlConfig.DataViewID)
	assert.Equal(t, types.StringValue("bytes"), pm.RangeSliderControlConfig.FieldName)
	assert.True(t, pm.RangeSliderControlConfig.Title.IsNull())
	assert.True(t, pm.RangeSliderControlConfig.UseGlobalFilters.IsNull())
	assert.True(t, pm.RangeSliderControlConfig.IgnoreValidations.IsNull())
	assert.True(t, pm.RangeSliderControlConfig.Value.IsNull())
	assert.True(t, pm.RangeSliderControlConfig.Step.IsNull())
}

// Test: when existing config block is nil and API returns data, preserve nil (null-preservation).
func Test_populateRangeSliderControlFromAPI_nilBlock_preservesNil(t *testing.T) {
	pm := &panelModel{}
	tfPanel := &panelModel{}
	populateRangeSliderControlFromAPI(context.Background(), pm, tfPanel, apiRangeSliderConfig())
	assert.Nil(t, pm.RangeSliderControlConfig)
}

// Test: when config block exists with known fields, they are updated from API.
func Test_populateRangeSliderControlFromAPI_knownFields_updatedFromAPI(t *testing.T) {
	pm := &panelModel{
		RangeSliderControlConfig: &rangeSliderControlConfigModel{
			DataViewID:        types.StringValue("dv-old"),
			FieldName:         types.StringValue("old-field"),
			Title:             types.StringValue("old title"),
			UseGlobalFilters:  types.BoolValue(false),
			IgnoreValidations: types.BoolValue(false),
			Value:             mustStringList("1", "9"),
			Step:              types.Float32Value(1),
		},
	}
	tfPanel := &panelModel{RangeSliderControlConfig: pm.RangeSliderControlConfig}
	apiCfg := apiRangeSliderConfig(
		withTitle("new title"),
		withUseGlobalFilters(true),
		withIgnoreValidations(true),
		withValue("10", "90"),
		withStep(5),
	)
	populateRangeSliderControlFromAPI(context.Background(), pm, tfPanel, apiCfg)
	require.NotNil(t, pm.RangeSliderControlConfig)
	assert.Equal(t, types.StringValue("dv-1"), pm.RangeSliderControlConfig.DataViewID)
	assert.Equal(t, types.StringValue("bytes"), pm.RangeSliderControlConfig.FieldName)
	assert.Equal(t, types.StringValue("new title"), pm.RangeSliderControlConfig.Title)
	assert.Equal(t, types.BoolValue(true), pm.RangeSliderControlConfig.UseGlobalFilters)
	assert.Equal(t, types.BoolValue(true), pm.RangeSliderControlConfig.IgnoreValidations)
	assert.Equal(t, mustStringList("10", "90"), pm.RangeSliderControlConfig.Value)
	assert.Equal(t, types.Float32Value(5), pm.RangeSliderControlConfig.Step)
}

// Test: null-preservation — null optional fields in state are not overwritten by API values.
func Test_populateRangeSliderControlFromAPI_nullOptionalFields_preserved(t *testing.T) {
	pm := &panelModel{
		RangeSliderControlConfig: &rangeSliderControlConfigModel{
			DataViewID:        types.StringValue("dv-1"),
			FieldName:         types.StringValue("bytes"),
			Title:             types.StringNull(),
			UseGlobalFilters:  types.BoolNull(),
			IgnoreValidations: types.BoolNull(),
			Value:             types.ListNull(types.StringType),
			Step:              types.Float32Null(),
		},
	}
	tfPanel := &panelModel{RangeSliderControlConfig: pm.RangeSliderControlConfig}
	apiCfg := apiRangeSliderConfig(
		withTitle("ignored"),
		withUseGlobalFilters(true),
		withIgnoreValidations(true),
		withValue("10", "90"),
		withStep(5),
	)
	populateRangeSliderControlFromAPI(context.Background(), pm, tfPanel, apiCfg)
	require.NotNil(t, pm.RangeSliderControlConfig)
	// Required fields are always updated.
	assert.Equal(t, types.StringValue("dv-1"), pm.RangeSliderControlConfig.DataViewID)
	assert.Equal(t, types.StringValue("bytes"), pm.RangeSliderControlConfig.FieldName)
	// Null optional fields are preserved.
	assert.True(t, pm.RangeSliderControlConfig.Title.IsNull())
	assert.True(t, pm.RangeSliderControlConfig.UseGlobalFilters.IsNull())
	assert.True(t, pm.RangeSliderControlConfig.IgnoreValidations.IsNull())
	assert.True(t, pm.RangeSliderControlConfig.Value.IsNull())
	assert.True(t, pm.RangeSliderControlConfig.Step.IsNull())
}

// Test: buildRangeSliderControlConfig sets known fields and omits null fields.
func Test_buildRangeSliderControlConfig_knownFields(t *testing.T) {
	pm := panelModel{
		RangeSliderControlConfig: &rangeSliderControlConfigModel{
			DataViewID:        types.StringValue("dv-1"),
			FieldName:         types.StringValue("bytes"),
			Title:             types.StringValue("My Slider"),
			UseGlobalFilters:  types.BoolValue(true),
			IgnoreValidations: types.BoolValue(false),
			Value:             mustStringList("10", "100"),
			Step:              types.Float32Value(5),
		},
	}
	rsPanel := kbapi.KbnDashboardPanelRangeSliderControl{
		Config: kbapi.KbnDashboardPanelRangeSliderControl_Config{},
	}
	buildRangeSliderControlConfig(pm, &rsPanel)
	assert.Equal(t, "dv-1", rsPanel.Config.DataViewId)
	assert.Equal(t, "bytes", rsPanel.Config.FieldName)
	require.NotNil(t, rsPanel.Config.Title)
	assert.Equal(t, "My Slider", *rsPanel.Config.Title)
	require.NotNil(t, rsPanel.Config.UseGlobalFilters)
	assert.True(t, *rsPanel.Config.UseGlobalFilters)
	require.NotNil(t, rsPanel.Config.IgnoreValidations)
	assert.False(t, *rsPanel.Config.IgnoreValidations)
	require.NotNil(t, rsPanel.Config.Value)
	assert.Equal(t, []string{"10", "100"}, *rsPanel.Config.Value)
	require.NotNil(t, rsPanel.Config.Step)
	assert.InEpsilon(t, float32(5), *rsPanel.Config.Step, 1e-6)
}

// Test: buildRangeSliderControlConfig omits null optional fields.
func Test_buildRangeSliderControlConfig_nullOptionalFields(t *testing.T) {
	pm := panelModel{
		RangeSliderControlConfig: &rangeSliderControlConfigModel{
			DataViewID:        types.StringValue("dv-1"),
			FieldName:         types.StringValue("bytes"),
			Title:             types.StringNull(),
			UseGlobalFilters:  types.BoolNull(),
			IgnoreValidations: types.BoolNull(),
			Value:             types.ListNull(types.StringType),
			Step:              types.Float32Null(),
		},
	}
	rsPanel := kbapi.KbnDashboardPanelRangeSliderControl{
		Config: kbapi.KbnDashboardPanelRangeSliderControl_Config{},
	}
	buildRangeSliderControlConfig(pm, &rsPanel)
	assert.Equal(t, "dv-1", rsPanel.Config.DataViewId)
	assert.Equal(t, "bytes", rsPanel.Config.FieldName)
	assert.Nil(t, rsPanel.Config.Title)
	assert.Nil(t, rsPanel.Config.UseGlobalFilters)
	assert.Nil(t, rsPanel.Config.IgnoreValidations)
	assert.Nil(t, rsPanel.Config.Value)
	assert.Nil(t, rsPanel.Config.Step)
}

// Test: round-trip — write then read back yields the same values.
func Test_rangeSliderControl_roundTrip(t *testing.T) {
	original := rangeSliderControlConfigModel{
		DataViewID:        types.StringValue("dv-1"),
		FieldName:         types.StringValue("price"),
		Title:             types.StringValue("Price Range"),
		UseGlobalFilters:  types.BoolValue(true),
		IgnoreValidations: types.BoolValue(false),
		Value:             mustStringList("50", "200"),
		Step:              types.Float32Value(10),
	}
	pm := panelModel{RangeSliderControlConfig: &original}
	rsPanel := kbapi.KbnDashboardPanelRangeSliderControl{
		Config: kbapi.KbnDashboardPanelRangeSliderControl_Config{},
	}
	buildRangeSliderControlConfig(pm, &rsPanel)

	out := &panelModel{
		RangeSliderControlConfig: &rangeSliderControlConfigModel{
			DataViewID:        types.StringValue("dv-1"),
			FieldName:         types.StringValue("price"),
			Title:             types.StringValue("Price Range"),
			UseGlobalFilters:  types.BoolValue(true),
			IgnoreValidations: types.BoolValue(false),
			Value:             mustStringList("50", "200"),
			Step:              types.Float32Value(10),
		},
	}
	tfPanel := &panelModel{RangeSliderControlConfig: out.RangeSliderControlConfig}
	populateRangeSliderControlFromAPI(context.Background(), out, tfPanel, rsPanel.Config)

	require.NotNil(t, out.RangeSliderControlConfig)
	assert.Equal(t, original.DataViewID, out.RangeSliderControlConfig.DataViewID)
	assert.Equal(t, original.FieldName, out.RangeSliderControlConfig.FieldName)
	assert.Equal(t, original.Title, out.RangeSliderControlConfig.Title)
	assert.Equal(t, original.UseGlobalFilters, out.RangeSliderControlConfig.UseGlobalFilters)
	assert.Equal(t, original.IgnoreValidations, out.RangeSliderControlConfig.IgnoreValidations)
	assert.Equal(t, original.Value, out.RangeSliderControlConfig.Value)
	assert.Equal(t, original.Step, out.RangeSliderControlConfig.Step)
}

// Test: value list with exactly 2 elements is preserved correctly.
func Test_rangeSliderControl_value_exactlyTwoElements(t *testing.T) {
	pm := panelModel{
		RangeSliderControlConfig: &rangeSliderControlConfigModel{
			DataViewID: types.StringValue("dv-1"),
			FieldName:  types.StringValue("bytes"),
			Value:      mustStringList("0", "1000"),
		},
	}
	rsPanel := kbapi.KbnDashboardPanelRangeSliderControl{
		Config: kbapi.KbnDashboardPanelRangeSliderControl_Config{},
	}
	buildRangeSliderControlConfig(pm, &rsPanel)
	require.NotNil(t, rsPanel.Config.Value)
	assert.Len(t, *rsPanel.Config.Value, 2)
	assert.Equal(t, "0", (*rsPanel.Config.Value)[0])
	assert.Equal(t, "1000", (*rsPanel.Config.Value)[1])
}
