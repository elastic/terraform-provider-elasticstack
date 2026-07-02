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

package rangeslider

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rsFieldCfg = kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaField
type rsEsqlCfg = kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsql

func apiRangeSliderFieldConfig(t *testing.T, opts ...func(*rsFieldCfg)) *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl {
	c := rsFieldCfg{DataViewId: "dv-1", FieldName: "bytes"}
	for _, o := range opts {
		o(&c)
	}
	p := &kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
	require.NoError(t, p.Config.FromKibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaField(c))
	return p
}

func apiRangeSliderEsqlConfig(t *testing.T, opts ...func(*rsEsqlCfg)) *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl {
	c := rsEsqlCfg{
		EsqlQuery:    "FROM orders | STATS min = MIN(price), max = MAX(price)",
		ValuesSource: kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsqlValuesSourceEsql,
	}
	for _, o := range opts {
		o(&c)
	}
	p := &kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
	require.NoError(t, p.Config.FromKibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsql(c))
	return p
}

func rsConfigField(t *testing.T, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl) rsFieldCfg {
	c, err := p.Config.AsKibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaField()
	require.NoError(t, err)
	return c
}

func rsConfigEsql(t *testing.T, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl) rsEsqlCfg {
	c, err := p.Config.AsKibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsql()
	require.NoError(t, err)
	return c
}

func rsWithTitle(t string) func(*rsFieldCfg) {
	return func(c *rsFieldCfg) { c.Title = &t }
}
func withUseGlobalFilters(b bool) func(*rsFieldCfg) {
	return func(c *rsFieldCfg) { c.UseGlobalFilters = &b }
}
func withIgnoreValidations(b bool) func(*rsFieldCfg) {
	return func(c *rsFieldCfg) { c.IgnoreValidations = &b }
}
func withValue(lo, hi string) func(*rsFieldCfg) {
	return func(c *rsFieldCfg) {
		v := []string{lo, hi}
		c.Value = &v
	}
}
func withStep(s float32) func(*rsFieldCfg) {
	return func(c *rsFieldCfg) { c.Step = &s }
}

func esqlWithTitle(t string) func(*rsEsqlCfg) {
	return func(c *rsEsqlCfg) { c.Title = &t }
}
func esqlWithUseGlobalFilters(b bool) func(*rsEsqlCfg) {
	return func(c *rsEsqlCfg) { c.UseGlobalFilters = &b }
}
func esqlWithIgnoreValidations(b bool) func(*rsEsqlCfg) {
	return func(c *rsEsqlCfg) { c.IgnoreValidations = &b }
}
func esqlWithValue(lo, hi string) func(*rsEsqlCfg) {
	return func(c *rsEsqlCfg) {
		v := []string{lo, hi}
		c.Value = &v
	}
}
func esqlWithStep(s float32) func(*rsEsqlCfg) {
	return func(c *rsEsqlCfg) { c.Step = &s }
}

func mustStringList(elems ...string) types.List {
	v, _ := types.ListValueFrom(context.Background(), types.StringType, elems)
	return v
}

// ---- by_field: PopulateFromAPI ----

// Test: on import (tfPanel == nil) with API data, populate all fields.
func Test_PopulateFromAPI_byField_import_allFields(t *testing.T) {
	pm := &models.PanelModel{}
	apiCfg := apiRangeSliderFieldConfig(t,
		rsWithTitle("My Control"),
		withUseGlobalFilters(true),
		withIgnoreValidations(false),
		withValue("100", "500"),
		withStep(10),
	)
	PopulateFromAPI(context.Background(), pm, nil, apiCfg)
	require.NotNil(t, pm.RangeSliderControlConfig)
	require.NotNil(t, pm.RangeSliderControlConfig.ByField)
	require.Nil(t, pm.RangeSliderControlConfig.ByEsql)
	bf := pm.RangeSliderControlConfig.ByField
	assert.Equal(t, types.StringValue("dv-1"), bf.DataViewID)
	assert.Equal(t, types.StringValue("bytes"), bf.FieldName)
	assert.Equal(t, types.StringValue("My Control"), bf.Title)
	assert.Equal(t, types.BoolValue(true), bf.UseGlobalFilters)
	assert.Equal(t, types.BoolValue(false), bf.IgnoreValidations)
	assert.Equal(t, mustStringList("100", "500"), bf.Value)
	assert.Equal(t, types.Float32Value(10), bf.Step)
}

// Test: on import (tfPanel == nil) with minimal API data (only required fields).
func Test_PopulateFromAPI_byField_import_requiredOnly(t *testing.T) {
	pm := &models.PanelModel{}
	PopulateFromAPI(context.Background(), pm, nil, apiRangeSliderFieldConfig(t))
	require.NotNil(t, pm.RangeSliderControlConfig)
	require.NotNil(t, pm.RangeSliderControlConfig.ByField)
	bf := pm.RangeSliderControlConfig.ByField
	assert.Equal(t, types.StringValue("dv-1"), bf.DataViewID)
	assert.Equal(t, types.StringValue("bytes"), bf.FieldName)
	assert.True(t, bf.Title.IsNull())
	assert.True(t, bf.UseGlobalFilters.IsNull())
	assert.True(t, bf.IgnoreValidations.IsNull())
	assert.True(t, bf.Value.IsNull())
	assert.True(t, bf.Step.IsNull())
}

// Test: when existing config block is nil and API returns data, preserve nil (null-preservation).
func Test_PopulateFromAPI_nilBlock_preservesNil(t *testing.T) {
	pm := &models.PanelModel{}
	tfPanel := &models.PanelModel{}
	PopulateFromAPI(context.Background(), pm, tfPanel, apiRangeSliderFieldConfig(t))
	assert.Nil(t, pm.RangeSliderControlConfig)
}

// Test: when by_field block exists with known fields, they are updated from API.
func Test_PopulateFromAPI_byField_knownFields_updatedFromAPI(t *testing.T) {
	prior := &models.RangeSliderControlByFieldModel{
		DataViewID:        types.StringValue("dv-old"),
		FieldName:         types.StringValue("old-field"),
		Title:             types.StringValue("old title"),
		UseGlobalFilters:  types.BoolValue(false),
		IgnoreValidations: types.BoolValue(false),
		Value:             mustStringList("1", "9"),
		Step:              types.Float32Value(1),
	}
	pm := &models.PanelModel{RangeSliderControlConfig: &models.RangeSliderControlConfigModel{ByField: prior}}
	tfPanel := &models.PanelModel{RangeSliderControlConfig: pm.RangeSliderControlConfig}
	apiCfg := apiRangeSliderFieldConfig(t,
		rsWithTitle("new title"),
		withUseGlobalFilters(true),
		withIgnoreValidations(true),
		withValue("10", "90"),
		withStep(5),
	)
	PopulateFromAPI(context.Background(), pm, tfPanel, apiCfg)
	require.NotNil(t, pm.RangeSliderControlConfig)
	bf := pm.RangeSliderControlConfig.ByField
	require.NotNil(t, bf)
	assert.Equal(t, types.StringValue("dv-1"), bf.DataViewID)
	assert.Equal(t, types.StringValue("bytes"), bf.FieldName)
	assert.Equal(t, types.StringValue("new title"), bf.Title)
	assert.Equal(t, types.BoolValue(true), bf.UseGlobalFilters)
	assert.Equal(t, types.BoolValue(true), bf.IgnoreValidations)
	assert.Equal(t, mustStringList("10", "90"), bf.Value)
	assert.Equal(t, types.Float32Value(5), bf.Step)
}

// Test: null-preservation — null optional fields in state are not overwritten by API values.
func Test_PopulateFromAPI_byField_nullOptionalFields_preserved(t *testing.T) {
	prior := &models.RangeSliderControlByFieldModel{
		DataViewID:        types.StringValue("dv-1"),
		FieldName:         types.StringValue("bytes"),
		Title:             types.StringNull(),
		UseGlobalFilters:  types.BoolNull(),
		IgnoreValidations: types.BoolNull(),
		Value:             types.ListNull(types.StringType),
		Step:              types.Float32Null(),
	}
	pm := &models.PanelModel{RangeSliderControlConfig: &models.RangeSliderControlConfigModel{ByField: prior}}
	tfPanel := &models.PanelModel{RangeSliderControlConfig: pm.RangeSliderControlConfig}
	apiCfg := apiRangeSliderFieldConfig(t,
		rsWithTitle("ignored"),
		withUseGlobalFilters(true),
		withIgnoreValidations(true),
		withValue("10", "90"),
		withStep(5),
	)
	PopulateFromAPI(context.Background(), pm, tfPanel, apiCfg)
	require.NotNil(t, pm.RangeSliderControlConfig)
	bf := pm.RangeSliderControlConfig.ByField
	require.NotNil(t, bf)
	// Required fields are always updated.
	assert.Equal(t, types.StringValue("dv-1"), bf.DataViewID)
	assert.Equal(t, types.StringValue("bytes"), bf.FieldName)
	// Null optional fields are preserved.
	assert.True(t, bf.Title.IsNull())
	assert.True(t, bf.UseGlobalFilters.IsNull())
	assert.True(t, bf.IgnoreValidations.IsNull())
	assert.True(t, bf.Value.IsNull())
	assert.True(t, bf.Step.IsNull())
}

// ---- by_esql: PopulateFromAPI ----

func Test_PopulateFromAPI_byEsql_import_allFields(t *testing.T) {
	pm := &models.PanelModel{}
	apiCfg := apiRangeSliderEsqlConfig(t,
		esqlWithTitle("Price Range"),
		esqlWithUseGlobalFilters(true),
		esqlWithIgnoreValidations(false),
		esqlWithValue("10", "1000"),
		esqlWithStep(5),
	)
	PopulateFromAPI(context.Background(), pm, nil, apiCfg)
	require.NotNil(t, pm.RangeSliderControlConfig)
	require.Nil(t, pm.RangeSliderControlConfig.ByField)
	require.NotNil(t, pm.RangeSliderControlConfig.ByEsql)
	be := pm.RangeSliderControlConfig.ByEsql
	assert.Equal(t, types.StringValue("FROM orders | STATS min = MIN(price), max = MAX(price)"), be.EsqlQuery)
	assert.Equal(t, types.StringValue("esql_query"), be.ValuesSource)
	assert.Equal(t, types.StringValue("Price Range"), be.Title)
	assert.Equal(t, types.BoolValue(true), be.UseGlobalFilters)
	assert.Equal(t, types.BoolValue(false), be.IgnoreValidations)
	assert.Equal(t, mustStringList("10", "1000"), be.Value)
	assert.Equal(t, types.Float32Value(5), be.Step)
}

func Test_PopulateFromAPI_byEsql_import_requiredOnly(t *testing.T) {
	pm := &models.PanelModel{}
	PopulateFromAPI(context.Background(), pm, nil, apiRangeSliderEsqlConfig(t))
	require.NotNil(t, pm.RangeSliderControlConfig)
	be := pm.RangeSliderControlConfig.ByEsql
	require.NotNil(t, be)
	assert.False(t, be.EsqlQuery.IsNull())
	assert.Equal(t, types.StringValue("esql_query"), be.ValuesSource)
	assert.True(t, be.Title.IsNull())
	assert.True(t, be.UseGlobalFilters.IsNull())
	assert.True(t, be.IgnoreValidations.IsNull())
	assert.True(t, be.Value.IsNull())
	assert.True(t, be.Step.IsNull())
}

func Test_PopulateFromAPI_byEsql_knownFields_updatedFromAPI(t *testing.T) {
	prior := &models.RangeSliderControlByEsqlModel{
		EsqlQuery:         types.StringValue("FROM old | STATS x = MIN(y)"),
		ValuesSource:      types.StringValue("esql_query"),
		Title:             types.StringValue("old title"),
		UseGlobalFilters:  types.BoolValue(false),
		IgnoreValidations: types.BoolValue(false),
		Value:             mustStringList("1", "9"),
		Step:              types.Float32Value(1),
	}
	pm := &models.PanelModel{RangeSliderControlConfig: &models.RangeSliderControlConfigModel{ByEsql: prior}}
	tfPanel := &models.PanelModel{RangeSliderControlConfig: pm.RangeSliderControlConfig}
	apiCfg := apiRangeSliderEsqlConfig(t,
		esqlWithTitle("new title"),
		esqlWithUseGlobalFilters(true),
		esqlWithIgnoreValidations(true),
		esqlWithValue("10", "90"),
		esqlWithStep(5),
	)
	PopulateFromAPI(context.Background(), pm, tfPanel, apiCfg)
	be := pm.RangeSliderControlConfig.ByEsql
	require.NotNil(t, be)
	assert.Equal(t, types.StringValue("new title"), be.Title)
	assert.Equal(t, types.BoolValue(true), be.UseGlobalFilters)
	assert.Equal(t, types.BoolValue(true), be.IgnoreValidations)
	assert.Equal(t, mustStringList("10", "90"), be.Value)
	assert.Equal(t, types.Float32Value(5), be.Step)
}

func Test_PopulateFromAPI_byEsql_nullOptionalFields_preserved(t *testing.T) {
	prior := &models.RangeSliderControlByEsqlModel{
		EsqlQuery:         types.StringValue("FROM orders | STATS x = MIN(price)"),
		ValuesSource:      types.StringValue("esql_query"),
		Title:             types.StringNull(),
		UseGlobalFilters:  types.BoolNull(),
		IgnoreValidations: types.BoolNull(),
		Value:             types.ListNull(types.StringType),
		Step:              types.Float32Null(),
	}
	pm := &models.PanelModel{RangeSliderControlConfig: &models.RangeSliderControlConfigModel{ByEsql: prior}}
	tfPanel := &models.PanelModel{RangeSliderControlConfig: pm.RangeSliderControlConfig}
	apiCfg := apiRangeSliderEsqlConfig(t,
		esqlWithTitle("ignored"),
		esqlWithUseGlobalFilters(true),
		esqlWithIgnoreValidations(true),
		esqlWithValue("10", "90"),
		esqlWithStep(5),
	)
	PopulateFromAPI(context.Background(), pm, tfPanel, apiCfg)
	be := pm.RangeSliderControlConfig.ByEsql
	require.NotNil(t, be)
	assert.True(t, be.Title.IsNull())
	assert.True(t, be.UseGlobalFilters.IsNull())
	assert.True(t, be.IgnoreValidations.IsNull())
	assert.True(t, be.Value.IsNull())
	assert.True(t, be.Step.IsNull())
}

// Test: branch switch — prior state described by_field, API now returns by_esql (e.g. resource
// replaced out of band). No prior intent exists for by_esql, so all its optional fields populate
// unconditionally rather than being forced null.
func Test_PopulateFromAPI_branchSwitch_fieldToEsql_populatesUnconditionally(t *testing.T) {
	priorField := &models.RangeSliderControlByFieldModel{
		DataViewID: types.StringValue("dv-1"),
		FieldName:  types.StringValue("bytes"),
	}
	pm := &models.PanelModel{RangeSliderControlConfig: &models.RangeSliderControlConfigModel{ByField: priorField}}
	tfPanel := &models.PanelModel{RangeSliderControlConfig: pm.RangeSliderControlConfig}
	apiCfg := apiRangeSliderEsqlConfig(t, esqlWithTitle("Now ES|QL"))
	PopulateFromAPI(context.Background(), pm, tfPanel, apiCfg)
	require.Nil(t, pm.RangeSliderControlConfig.ByField)
	require.NotNil(t, pm.RangeSliderControlConfig.ByEsql)
	assert.Equal(t, types.StringValue("Now ES|QL"), pm.RangeSliderControlConfig.ByEsql.Title)
}

// ---- by_field: BuildConfig ----

func Test_BuildConfig_byField_knownFields(t *testing.T) {
	pm := models.PanelModel{
		RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
			ByField: &models.RangeSliderControlByFieldModel{
				DataViewID:        types.StringValue("dv-1"),
				FieldName:         types.StringValue("bytes"),
				Title:             types.StringValue("My Slider"),
				UseGlobalFilters:  types.BoolValue(true),
				IgnoreValidations: types.BoolValue(false),
				Value:             mustStringList("10", "100"),
				Step:              types.Float32Value(5),
			},
		},
	}
	rsPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
	require.False(t, BuildConfig(pm, &rsPanel).HasError(), "BuildConfig failed")
	cfg := rsConfigField(t, rsPanel)
	assert.Equal(t, "dv-1", cfg.DataViewId)
	assert.Equal(t, "bytes", cfg.FieldName)
	// values_source is deliberately left unset on the wire for by_field (see buildFieldConfig):
	// Kibana defaults it to "field" when absent, and older Kibana versions reject the property
	// entirely if present.
	assert.Nil(t, cfg.ValuesSource)
	require.NotNil(t, cfg.Title)
	assert.Equal(t, "My Slider", *cfg.Title)
	require.NotNil(t, cfg.UseGlobalFilters)
	assert.True(t, *cfg.UseGlobalFilters)
	require.NotNil(t, cfg.IgnoreValidations)
	assert.False(t, *cfg.IgnoreValidations)
	require.NotNil(t, cfg.Value)
	assert.Equal(t, []string{"10", "100"}, *cfg.Value)
	require.NotNil(t, cfg.Step)
	assert.InEpsilon(t, float32(5), *cfg.Step, 1e-6)
}

func Test_BuildConfig_byField_nullOptionalFields(t *testing.T) {
	pm := models.PanelModel{
		RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
			ByField: &models.RangeSliderControlByFieldModel{
				DataViewID:        types.StringValue("dv-1"),
				FieldName:         types.StringValue("bytes"),
				Title:             types.StringNull(),
				UseGlobalFilters:  types.BoolNull(),
				IgnoreValidations: types.BoolNull(),
				Value:             types.ListNull(types.StringType),
				Step:              types.Float32Null(),
			},
		},
	}
	rsPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
	require.False(t, BuildConfig(pm, &rsPanel).HasError(), "BuildConfig failed")
	cfg := rsConfigField(t, rsPanel)
	assert.Equal(t, "dv-1", cfg.DataViewId)
	assert.Equal(t, "bytes", cfg.FieldName)
	assert.Nil(t, cfg.Title)
	assert.Nil(t, cfg.UseGlobalFilters)
	assert.Nil(t, cfg.IgnoreValidations)
	assert.Nil(t, cfg.Value)
	assert.Nil(t, cfg.Step)
}

// ---- by_esql: BuildConfig ----

func Test_BuildConfig_byEsql_knownFields(t *testing.T) {
	pm := models.PanelModel{
		RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
			ByEsql: &models.RangeSliderControlByEsqlModel{
				EsqlQuery:         types.StringValue("FROM orders | STATS min = MIN(price), max = MAX(price)"),
				ValuesSource:      types.StringValue("esql_query"),
				Title:             types.StringValue("Price Range"),
				UseGlobalFilters:  types.BoolValue(true),
				IgnoreValidations: types.BoolValue(false),
				Value:             mustStringList("10", "1000"),
				Step:              types.Float32Value(5),
			},
		},
	}
	rsPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
	require.False(t, BuildConfig(pm, &rsPanel).HasError(), "BuildConfig failed")
	cfg := rsConfigEsql(t, rsPanel)
	assert.Equal(t, "FROM orders | STATS min = MIN(price), max = MAX(price)", cfg.EsqlQuery)
	// The wire enum's only legal value is "esql" even though the TF-facing attribute is "esql_query".
	assert.Equal(t, kbapi.KibanaHTTPAPIsKbnControlsSchemasRangeSliderControlSchemaEsqlValuesSourceEsql, cfg.ValuesSource)
	require.NotNil(t, cfg.Title)
	assert.Equal(t, "Price Range", *cfg.Title)
	require.NotNil(t, cfg.UseGlobalFilters)
	assert.True(t, *cfg.UseGlobalFilters)
	require.NotNil(t, cfg.IgnoreValidations)
	assert.False(t, *cfg.IgnoreValidations)
	require.NotNil(t, cfg.Value)
	assert.Equal(t, []string{"10", "1000"}, *cfg.Value)
	require.NotNil(t, cfg.Step)
	assert.InEpsilon(t, float32(5), *cfg.Step, 1e-6)
}

func Test_BuildConfig_byEsql_nullOptionalFields(t *testing.T) {
	pm := models.PanelModel{
		RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
			ByEsql: &models.RangeSliderControlByEsqlModel{
				EsqlQuery:         types.StringValue("FROM orders | STATS min = MIN(price)"),
				ValuesSource:      types.StringValue("esql_query"),
				Title:             types.StringNull(),
				UseGlobalFilters:  types.BoolNull(),
				IgnoreValidations: types.BoolNull(),
				Value:             types.ListNull(types.StringType),
				Step:              types.Float32Null(),
			},
		},
	}
	rsPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
	require.False(t, BuildConfig(pm, &rsPanel).HasError(), "BuildConfig failed")
	cfg := rsConfigEsql(t, rsPanel)
	assert.Equal(t, "FROM orders | STATS min = MIN(price)", cfg.EsqlQuery)
	assert.Nil(t, cfg.Title)
	assert.Nil(t, cfg.UseGlobalFilters)
	assert.Nil(t, cfg.IgnoreValidations)
	assert.Nil(t, cfg.Value)
	assert.Nil(t, cfg.Step)
}

// Test: BuildConfig returns an error when neither branch is set (defensive; schema-level
// ExactlyOneOfNestedAttrsValidator should normally prevent this at plan time).
func Test_BuildConfig_neitherBranchSet_errors(t *testing.T) {
	pm := models.PanelModel{RangeSliderControlConfig: &models.RangeSliderControlConfigModel{}}
	rsPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
	diags := BuildConfig(pm, &rsPanel)
	assert.True(t, diags.HasError())
}

// ---- Round trips ----

// Test: round-trip — write then read back yields the same values for by_field.
func Test_rangeSliderControl_byField_roundTrip(t *testing.T) {
	original := models.RangeSliderControlByFieldModel{
		DataViewID:        types.StringValue("dv-1"),
		FieldName:         types.StringValue("price"),
		Title:             types.StringValue("Price Range"),
		UseGlobalFilters:  types.BoolValue(true),
		IgnoreValidations: types.BoolValue(false),
		Value:             mustStringList("50", "200"),
		Step:              types.Float32Value(10),
	}
	pm := models.PanelModel{RangeSliderControlConfig: &models.RangeSliderControlConfigModel{ByField: &original}}
	rsPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
	require.False(t, BuildConfig(pm, &rsPanel).HasError(), "BuildConfig failed")

	out := &models.PanelModel{
		RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
			ByField: &models.RangeSliderControlByFieldModel{
				DataViewID:        types.StringValue("dv-1"),
				FieldName:         types.StringValue("price"),
				Title:             types.StringValue("Price Range"),
				UseGlobalFilters:  types.BoolValue(true),
				IgnoreValidations: types.BoolValue(false),
				Value:             mustStringList("50", "200"),
				Step:              types.Float32Value(10),
			},
		},
	}
	tfPanel := &models.PanelModel{RangeSliderControlConfig: out.RangeSliderControlConfig}
	PopulateFromAPI(context.Background(), out, tfPanel, &rsPanel)

	require.NotNil(t, out.RangeSliderControlConfig)
	bf := out.RangeSliderControlConfig.ByField
	require.NotNil(t, bf)
	assert.Equal(t, original.DataViewID, bf.DataViewID)
	assert.Equal(t, original.FieldName, bf.FieldName)
	assert.Equal(t, original.Title, bf.Title)
	assert.Equal(t, original.UseGlobalFilters, bf.UseGlobalFilters)
	assert.Equal(t, original.IgnoreValidations, bf.IgnoreValidations)
	assert.Equal(t, original.Value, bf.Value)
	assert.Equal(t, original.Step, bf.Step)
}

// Test: round-trip — write then read back yields the same values for by_esql.
func Test_rangeSliderControl_byEsql_roundTrip(t *testing.T) {
	original := models.RangeSliderControlByEsqlModel{
		EsqlQuery:         types.StringValue("FROM orders | STATS min = MIN(price), max = MAX(price)"),
		ValuesSource:      types.StringValue("esql_query"),
		Title:             types.StringValue("Price Range"),
		UseGlobalFilters:  types.BoolValue(true),
		IgnoreValidations: types.BoolValue(false),
		Value:             mustStringList("50", "200"),
		Step:              types.Float32Value(10),
	}
	pm := models.PanelModel{RangeSliderControlConfig: &models.RangeSliderControlConfigModel{ByEsql: &original}}
	rsPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
	require.False(t, BuildConfig(pm, &rsPanel).HasError(), "BuildConfig failed")

	out := &models.PanelModel{
		RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
			ByEsql: &models.RangeSliderControlByEsqlModel{
				EsqlQuery:         types.StringValue("FROM orders | STATS min = MIN(price), max = MAX(price)"),
				ValuesSource:      types.StringValue("esql_query"),
				Title:             types.StringValue("Price Range"),
				UseGlobalFilters:  types.BoolValue(true),
				IgnoreValidations: types.BoolValue(false),
				Value:             mustStringList("50", "200"),
				Step:              types.Float32Value(10),
			},
		},
	}
	tfPanel := &models.PanelModel{RangeSliderControlConfig: out.RangeSliderControlConfig}
	PopulateFromAPI(context.Background(), out, tfPanel, &rsPanel)

	require.NotNil(t, out.RangeSliderControlConfig)
	be := out.RangeSliderControlConfig.ByEsql
	require.NotNil(t, be)
	assert.Equal(t, original.EsqlQuery, be.EsqlQuery)
	assert.Equal(t, original.ValuesSource, be.ValuesSource)
	assert.Equal(t, original.Title, be.Title)
	assert.Equal(t, original.UseGlobalFilters, be.UseGlobalFilters)
	assert.Equal(t, original.IgnoreValidations, be.IgnoreValidations)
	assert.Equal(t, original.Value, be.Value)
	assert.Equal(t, original.Step, be.Step)
}

// Test: value list with exactly 2 elements is preserved correctly, for both branches.
func Test_rangeSliderControl_value_exactlyTwoElements(t *testing.T) {
	t.Run("by_field", func(t *testing.T) {
		pm := models.PanelModel{
			RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
				ByField: &models.RangeSliderControlByFieldModel{
					DataViewID: types.StringValue("dv-1"),
					FieldName:  types.StringValue("bytes"),
					Value:      mustStringList("0", "1000"),
				},
			},
		}
		rsPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
		require.False(t, BuildConfig(pm, &rsPanel).HasError(), "BuildConfig failed")
		cfg := rsConfigField(t, rsPanel)
		require.NotNil(t, cfg.Value)
		assert.Len(t, *cfg.Value, 2)
		assert.Equal(t, "0", (*cfg.Value)[0])
		assert.Equal(t, "1000", (*cfg.Value)[1])
	})

	t.Run("by_esql", func(t *testing.T) {
		pm := models.PanelModel{
			RangeSliderControlConfig: &models.RangeSliderControlConfigModel{
				ByEsql: &models.RangeSliderControlByEsqlModel{
					EsqlQuery:    types.StringValue("FROM orders | STATS x = MIN(price)"),
					ValuesSource: types.StringValue("esql_query"),
					Value:        mustStringList("0", "1000"),
				},
			},
		}
		rsPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeRangeSliderControl{}
		require.False(t, BuildConfig(pm, &rsPanel).HasError(), "BuildConfig failed")
		cfg := rsConfigEsql(t, rsPanel)
		require.NotNil(t, cfg.Value)
		assert.Len(t, *cfg.Value, 2)
		assert.Equal(t, "0", (*cfg.Value)[0])
		assert.Equal(t, "1000", (*cfg.Value)[1])
	})
}
