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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func discoverSessionTestGrid() struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
} {
	w := float32(24)
	h := float32(15)
	return struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	}{
		X: 0, Y: 0, W: &w, H: &h,
	}
}

func discoverSessionSortElementType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType, "direction": types.StringType}}
}

func discoverSessionDSLTabAttrTypes() map[string]attr.Type {
	colSettingsElem := types.ObjectType{AttrTypes: map[string]attr.Type{"width": types.Float64Type}}
	return map[string]attr.Type{
		"column_order":      types.ListType{ElemType: types.StringType},
		"column_settings":   types.MapType{ElemType: colSettingsElem},
		"sort":              types.ListType{ElemType: discoverSessionSortElementType()},
		"density":           types.StringType,
		"header_row_height": types.StringType,
		"row_height":        types.StringType,
		"rows_per_page":     types.Int64Type,
		"sample_size":       types.Int64Type,
		"view_mode":         types.StringType,
		"query": types.ObjectType{AttrTypes: map[string]attr.Type{
			"language": types.StringType, "expression": types.StringType,
		}},
		"data_source_json": jsontypes.NormalizedType{},
		"filters":          types.ListType{ElemType: dashboardRootSavedFiltersElementType()},
	}
}

func discoverSessionESQLTabAttrTypes() map[string]attr.Type {
	colSettingsElem := types.ObjectType{AttrTypes: map[string]attr.Type{"width": types.Float64Type}}
	return map[string]attr.Type{
		"column_order":      types.ListType{ElemType: types.StringType},
		"column_settings":   types.MapType{ElemType: colSettingsElem},
		"sort":              types.ListType{ElemType: discoverSessionSortElementType()},
		"density":           types.StringType,
		"header_row_height": types.StringType,
		"row_height":        types.StringType,
		"data_source_json":  jsontypes.NormalizedType{},
	}
}

func discoverSessionOverridesAttrTypes() map[string]attr.Type {
	colSettingsElem := types.ObjectType{AttrTypes: map[string]attr.Type{"width": types.Float64Type}}
	return map[string]attr.Type{
		"column_order":      types.ListType{ElemType: types.StringType},
		"column_settings":   types.MapType{ElemType: colSettingsElem},
		"sort":              types.ListType{ElemType: discoverSessionSortElementType()},
		"density":           types.StringType,
		"header_row_height": types.StringType,
		"row_height":        types.StringType,
		"rows_per_page":     types.Int64Type,
		"sample_size":       types.Int64Type,
	}
}

func Test_discoverSessionConfigModeValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := discoverSessionConfigModeValidator{}

	dslT := discoverSessionDSLTabAttrTypes()
	esqlT := discoverSessionESQLTabAttrTypes()
	emptyTabAttrs := map[string]attr.Type{
		"dsl":  types.ObjectType{AttrTypes: dslT},
		"esql": types.ObjectType{AttrTypes: esqlT},
	}

	byValueAttrs := map[string]attr.Type{
		"time_range": types.ObjectType{AttrTypes: map[string]attr.Type{
			"from": types.StringType, "to": types.StringType, "mode": types.StringType}},
		"tab": types.ObjectType{AttrTypes: emptyTabAttrs},
	}
	byRefAttrs := map[string]attr.Type{
		"time_range": types.ObjectType{AttrTypes: map[string]attr.Type{
			"from": types.StringType, "to": types.StringType, "mode": types.StringType}},
		"ref_id":          types.StringType,
		"selected_tab_id": types.StringType,
		"overrides":       types.ObjectType{AttrTypes: discoverSessionOverridesAttrTypes()},
	}

	tr := types.ObjectValueMust(
		map[string]attr.Type{"from": types.StringType, "to": types.StringType, "mode": types.StringType},
		map[string]attr.Value{
			"from": types.StringValue("now-15m"),
			"to":   types.StringValue("now"),
			"mode": types.StringNull(),
		},
	)

	filtersElem := dashboardRootSavedFiltersElementType()
	byValueObj := types.ObjectValueMust(byValueAttrs, map[string]attr.Value{
		"time_range": tr,
		"tab": types.ObjectValueMust(emptyTabAttrs, map[string]attr.Value{
			"dsl": types.ObjectValueMust(dslT, map[string]attr.Value{
				"column_order":      types.ListNull(types.StringType),
				"column_settings":   types.MapNull(types.ObjectType{AttrTypes: map[string]attr.Type{"width": types.Float64Type}}),
				"sort":              types.ListNull(discoverSessionSortElementType()),
				"density":           types.StringNull(),
				"header_row_height": types.StringNull(),
				"row_height":        types.StringNull(),
				"rows_per_page":     types.Int64Null(),
				"sample_size":       types.Int64Null(),
				"view_mode":         types.StringNull(),
				"query": types.ObjectValueMust(map[string]attr.Type{
					"language": types.StringType, "expression": types.StringType,
				}, map[string]attr.Value{
					"language": types.StringValue("kql"), "expression": types.StringValue("*"),
				}),
				"data_source_json": jsontypes.NewNormalizedValue(`{"type":"data_view_reference","id":"logs-*"}`),
				"filters":          types.ListNull(filtersElem),
			}),
			"esql": types.ObjectNull(esqlT),
		}),
	})

	byRefObj := types.ObjectValueMust(byRefAttrs, map[string]attr.Value{
		"time_range":      tr,
		"ref_id":          types.StringValue("discover-1"),
		"selected_tab_id": types.StringNull(),
		"overrides":       types.ObjectNull(discoverSessionOverridesAttrTypes()),
	})

	ddElem := types.ObjectType{AttrTypes: map[string]attr.Type{
		"url": types.StringType, "label": types.StringType, "encode_url": types.BoolType, "open_in_new_tab": types.BoolType,
	}}
	rootAttrTypes := map[string]attr.Type{
		"title": types.StringType, "description": types.StringType, "hide_title": types.BoolType, "hide_border": types.BoolType,
		"drilldowns":   types.ListType{ElemType: ddElem},
		"by_value":     types.ObjectType{AttrTypes: byValueAttrs},
		"by_reference": types.ObjectType{AttrTypes: byRefAttrs},
	}

	rootObj := func(byValue, byRef attr.Value) types.Object {
		return types.ObjectValueMust(rootAttrTypes, map[string]attr.Value{
			"title": types.StringNull(), "description": types.StringNull(), "hide_title": types.BoolNull(), "hide_border": types.BoolNull(),
			"drilldowns":   types.ListNull(ddElem),
			"by_value":     byValue,
			"by_reference": byRef,
		})
	}

	t.Run("rejects both by_value and by_reference", func(t *testing.T) {
		t.Parallel()
		ov := rootObj(byValueObj, byRefObj)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("discover_session_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "not both")
	})

	t.Run("rejects neither branch", func(t *testing.T) {
		t.Parallel()
		ov := rootObj(types.ObjectNull(byValueAttrs), types.ObjectNull(byRefAttrs))
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("discover_session_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("accepts by_value only", func(t *testing.T) {
		t.Parallel()
		ov := rootObj(byValueObj, types.ObjectNull(byRefAttrs))
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("discover_session_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("accepts by_reference only", func(t *testing.T) {
		t.Parallel()
		ov := rootObj(types.ObjectNull(byValueAttrs), byRefObj)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("discover_session_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})
}

func Test_discoverSessionTabModeValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := discoverSessionTabModeValidator{}

	dslT := discoverSessionDSLTabAttrTypes()
	esqlT := discoverSessionESQLTabAttrTypes()
	tabTypes := map[string]attr.Type{
		"dsl":  types.ObjectType{AttrTypes: dslT},
		"esql": types.ObjectType{AttrTypes: esqlT},
	}

	colSettingsElem := types.ObjectType{AttrTypes: map[string]attr.Type{"width": types.Float64Type}}
	dslSet := types.ObjectValueMust(dslT, map[string]attr.Value{
		"column_order":      types.ListNull(types.StringType),
		"column_settings":   types.MapNull(colSettingsElem),
		"sort":              types.ListNull(discoverSessionSortElementType()),
		"density":           types.StringNull(),
		"header_row_height": types.StringNull(),
		"row_height":        types.StringNull(),
		"rows_per_page":     types.Int64Null(),
		"sample_size":       types.Int64Null(),
		"view_mode":         types.StringNull(),
		"query": types.ObjectValueMust(map[string]attr.Type{
			"language": types.StringType, "expression": types.StringType,
		}, map[string]attr.Value{
			"language": types.StringValue("kql"), "expression": types.StringValue("*"),
		}),
		"data_source_json": jsontypes.NewNormalizedValue(`{"type":"data_view_reference","id":"logs-*"}`),
		"filters":          types.ListNull(dashboardRootSavedFiltersElementType()),
	})
	esqlSet := types.ObjectValueMust(esqlT, map[string]attr.Value{
		"column_order":      types.ListNull(types.StringType),
		"column_settings":   types.MapNull(colSettingsElem),
		"sort":              types.ListNull(discoverSessionSortElementType()),
		"density":           types.StringNull(),
		"header_row_height": types.StringNull(),
		"row_height":        types.StringNull(),
		"data_source_json":  jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM logs-*"}`),
	})

	t.Run("rejects both dsl and esql", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(tabTypes, map[string]attr.Value{"dsl": dslSet, "esql": esqlSet})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("tab")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "not both")
	})

	t.Run("rejects neither dsl nor esql", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(tabTypes, map[string]attr.Value{
			"dsl":  types.ObjectNull(dslT),
			"esql": types.ObjectNull(esqlT),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("tab")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("accepts dsl only", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(tabTypes, map[string]attr.Value{
			"dsl": dslSet, "esql": types.ObjectNull(esqlT),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("tab")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("accepts esql only", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(tabTypes, map[string]attr.Value{
			"dsl": types.ObjectNull(dslT), "esql": esqlSet,
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("tab")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})
}

func Test_discoverSession_rowHeightStringValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := makeDiscoverSessionRowHeightStringValidator(20)

	t.Run("rejects out of range", func(t *testing.T) {
		var resp validator.StringResponse
		v.ValidateString(ctx, validator.StringRequest{
			Path:           path.Root("row_height"),
			PathExpression: path.MatchRoot("row_height"),
			ConfigValue:    types.StringValue("25"),
		}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("accepts auto", func(t *testing.T) {
		var resp validator.StringResponse
		v.ValidateString(ctx, validator.StringRequest{
			Path:           path.Root("row_height"),
			PathExpression: path.MatchRoot("row_height"),
			ConfigValue:    types.StringValue("auto"),
		}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})
}

func Test_discoverSessionDSLTabToAPI_invalidDataSourceJSON(t *testing.T) {
	m := models.DiscoverSessionDSLTabModel{
		Query: &models.FilterSimpleModel{
			Language:   types.StringValue("kql"),
			Expression: types.StringValue("*"),
		},
		DataSourceJSON: jsontypes.NewNormalizedValue(`not-json`),
	}
	_, diags := discoverSessionDSLTabToAPI(context.Background(), m)
	require.True(t, diags.HasError())
}

func Test_discoverSession_viewMode_invalidRejectedBySchemaValidators(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	panel := getPanelSchema()
	dsc, ok := panel.Attributes["discover_session_config"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	byVal, ok := dsc.Attributes["by_value"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	tab, ok := byVal.Attributes["tab"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	dsl, ok := tab.Attributes["dsl"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	vmAttr, ok := dsl.Attributes["view_mode"].(schema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, vmAttr.Validators)

	req := validator.StringRequest{
		Path:           path.Root("view_mode"),
		PathExpression: path.MatchRoot("view_mode"),
		ConfigValue:    types.StringValue("invalid"),
	}
	var resp validator.StringResponse
	for _, val := range vmAttr.Validators {
		val.ValidateString(ctx, req, &resp)
	}
	require.True(t, resp.Diagnostics.HasError())
}

func Test_discoverSession_byValue_dsl_roundTrip(t *testing.T) {
	ctx := context.Background()
	rawFilter := `{"condition":{"field":"host.name","operator":"is","value":"web"},"type":"condition"}`
	pm := models.PanelModel{
		DiscoverSessionConfig: &models.DiscoverSessionPanelConfigModel{
			ByValue: &models.DiscoverSessionPanelByValueModel{
				TimeRange: &models.TimeRangeModel{
					From: types.StringValue("now-30m"),
					To:   types.StringValue("now"),
					Mode: types.StringNull(),
				},
				Tab: models.DiscoverSessionTabModel{
					DSL: &models.DiscoverSessionDSLTabModel{
						ColumnOrder: types.ListValueMust(types.StringType, []attr.Value{
							types.StringValue("@timestamp"),
							types.StringValue("message"),
						}),
						Query: &models.FilterSimpleModel{
							Language:   types.StringValue("kql"),
							Expression: types.StringValue(`host.name : "web-01"`),
						},
						DataSourceJSON: jsontypes.NewNormalizedValue(`{"id":"logs-*","type":"data_view_reference"}`),
						Filters: []models.ChartFilterJSONModel{
							{FilterJSON: jsontypes.NewNormalizedValue(rawFilter)},
						},
					},
				},
			},
		},
	}

	grid := discoverSessionTestGrid()
	item, diags := discoverSessionPanelToAPI(context.Background(), pm, grid, nil, nil)
	require.False(t, diags.HasError())

	dsPanel, err := item.AsKbnDashboardPanelTypeDiscoverSession()
	require.NoError(t, err)

	next := pm
	populateDiscoverSessionPanelFromAPI(context.Background(), &next, &pm, dsPanel)

	require.Nil(t, next.DiscoverSessionConfig.ByValue.Tab.ESQL)
	require.NotNil(t, next.DiscoverSessionConfig.ByValue.Tab.DSL)

	dsl := next.DiscoverSessionConfig.ByValue.Tab.DSL
	ctxSE := context.Background()
	if assert.Len(t, dsl.Filters, 1) {
		eq, d := dsl.Filters[0].FilterJSON.StringSemanticEquals(ctxSE, jsontypes.NewNormalizedValue(rawFilter))
		require.False(t, d.HasError())
		assert.True(t, eq)
	}

	co := dsl.ColumnOrder.Elements()
	require.Len(t, co, 2)
	assert.Equal(t, "@timestamp", co[0].(types.String).ValueString())
	assert.Equal(t, "message", co[1].(types.String).ValueString())
	assert.Equal(t, `host.name : "web-01"`, dsl.Query.Expression.ValueString())

	dsJSONEq, d := dsl.DataSourceJSON.StringSemanticEquals(ctx, jsontypes.NewNormalizedValue(`{"type":"data_view_reference","id":"logs-*"}`))
	require.False(t, d.HasError())
	assert.True(t, dsJSONEq)
}

func Test_discoverSession_byValue_esql_roundTrip(t *testing.T) {
	ctx := context.Background()
	pm := models.PanelModel{
		DiscoverSessionConfig: &models.DiscoverSessionPanelConfigModel{
			ByValue: &models.DiscoverSessionPanelByValueModel{
				TimeRange: &models.TimeRangeModel{
					From: types.StringValue("now-30m"),
					To:   types.StringValue("now"),
					Mode: types.StringNull(),
				},
				Tab: models.DiscoverSessionTabModel{
					ESQL: &models.DiscoverSessionESQLTabModel{
						DataSourceJSON: jsontypes.NewNormalizedValue(`{"query":"FROM logs-*","type":"esql"}`),
					},
				},
			},
		},
	}

	item, diags := discoverSessionPanelToAPI(context.Background(), pm, discoverSessionTestGrid(), nil, nil)
	require.False(t, diags.HasError())

	dsPanel, err := item.AsKbnDashboardPanelTypeDiscoverSession()
	require.NoError(t, err)

	next := pm
	populateDiscoverSessionPanelFromAPI(context.Background(), &next, &pm, dsPanel)

	require.Nil(t, next.DiscoverSessionConfig.ByValue.Tab.DSL)
	require.NotNil(t, next.DiscoverSessionConfig.ByValue.Tab.ESQL)

	esql := next.DiscoverSessionConfig.ByValue.Tab.ESQL
	dsJSONEq, d := esql.DataSourceJSON.StringSemanticEquals(ctx, jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM logs-*"}`))
	require.False(t, d.HasError())
	assert.True(t, dsJSONEq)
}

func Test_discoverSession_byReference_preservesPractitionerSelectedTabID(t *testing.T) {
	pm := models.PanelModel{
		DiscoverSessionConfig: &models.DiscoverSessionPanelConfigModel{
			ByReference: &models.DiscoverSessionPanelByRefModel{
				RefID:         types.StringValue("discover-1"),
				SelectedTabID: types.StringValue("user-tab"),
				TimeRange: &models.TimeRangeModel{
					From: types.StringValue("now-30m"),
					To:   types.StringValue("now"),
					Mode: types.StringNull(),
				},
			},
		},
	}

	apiTab := "api-default-tab"
	cfg1 := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1{
		RefId:         "discover-1",
		TimeRange:     kbapi.KbnEsQueryServerTimeRangeSchema{From: "now-30m", To: "now"},
		SelectedTabId: &apiTab,
	}
	var cfgUnion kbapi.KbnDashboardPanelTypeDiscoverSession_Config
	require.NoError(t, cfgUnion.FromKbnDashboardPanelTypeDiscoverSessionConfig1(cfg1))

	apiPanel := kbapi.KbnDashboardPanelTypeDiscoverSession{
		Config: cfgUnion,
		Type:   kbapi.DiscoverSession,
		Grid:   discoverSessionTestGrid(),
	}

	next := pm
	populateDiscoverSessionPanelFromAPI(context.Background(), &next, &pm, apiPanel)
	assert.Equal(t, "user-tab", next.DiscoverSessionConfig.ByReference.SelectedTabID.ValueString())
}

func Test_discoverSession_byReference_selectedTabID_fromAPI_thenStable(t *testing.T) {
	pm := models.PanelModel{
		DiscoverSessionConfig: &models.DiscoverSessionPanelConfigModel{
			ByReference: &models.DiscoverSessionPanelByRefModel{
				RefID:         types.StringValue("discover-1"),
				SelectedTabID: types.StringNull(),
				TimeRange: &models.TimeRangeModel{
					From: types.StringValue("now-30m"),
					To:   types.StringValue("now"),
					Mode: types.StringNull(),
				},
			},
		},
	}

	apiTab := "resolved-tab-id"
	cfg1 := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1{
		RefId:         "discover-1",
		TimeRange:     kbapi.KbnEsQueryServerTimeRangeSchema{From: "now-30m", To: "now"},
		SelectedTabId: &apiTab,
	}
	var cfgUnion kbapi.KbnDashboardPanelTypeDiscoverSession_Config
	require.NoError(t, cfgUnion.FromKbnDashboardPanelTypeDiscoverSessionConfig1(cfg1))

	apiPanel := kbapi.KbnDashboardPanelTypeDiscoverSession{
		Config: cfgUnion,
		Type:   kbapi.DiscoverSession,
		Grid:   discoverSessionTestGrid(),
	}

	first := pm
	populateDiscoverSessionPanelFromAPI(context.Background(), &first, &pm, apiPanel)
	require.Equal(t, "resolved-tab-id", first.DiscoverSessionConfig.ByReference.SelectedTabID.ValueString())

	second := first
	populateDiscoverSessionPanelFromAPI(context.Background(), &second, &first, apiPanel)
	require.Equal(t, "resolved-tab-id", second.DiscoverSessionConfig.ByReference.SelectedTabID.ValueString())
}

func Test_populateDiscoverSessionPanelFromAPI_branchMismatch_repopulates(t *testing.T) {
	ctx := context.Background()

	t.Run("API by_reference replaces prior by_value", func(t *testing.T) {
		shared := &models.DiscoverSessionPanelConfigModel{
			ByValue: &models.DiscoverSessionPanelByValueModel{
				TimeRange: &models.TimeRangeModel{
					From: types.StringValue("now-30m"),
					To:   types.StringValue("now"),
					Mode: types.StringNull(),
				},
				Tab: models.DiscoverSessionTabModel{
					DSL: &models.DiscoverSessionDSLTabModel{
						Query: &models.FilterSimpleModel{
							Language:   types.StringValue("kql"),
							Expression: types.StringValue("*"),
						},
						DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_reference","id":"logs-*"}`),
					},
				},
			},
		}
		pm := models.PanelModel{DiscoverSessionConfig: shared}
		tfPanel := models.PanelModel{DiscoverSessionConfig: shared}

		cfg1 := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1{
			RefId:     "saved-session-99",
			TimeRange: kbapi.KbnEsQueryServerTimeRangeSchema{From: "now-30m", To: "now"},
		}
		var u kbapi.KbnDashboardPanelTypeDiscoverSession_Config
		require.NoError(t, u.FromKbnDashboardPanelTypeDiscoverSessionConfig1(cfg1))
		apiPanel := kbapi.KbnDashboardPanelTypeDiscoverSession{Config: u, Type: kbapi.DiscoverSession, Grid: discoverSessionTestGrid()}

		populateDiscoverSessionPanelFromAPI(ctx, &pm, &tfPanel, apiPanel)

		require.Nil(t, pm.DiscoverSessionConfig.ByValue)
		require.NotNil(t, pm.DiscoverSessionConfig.ByReference)
		assert.Equal(t, "saved-session-99", pm.DiscoverSessionConfig.ByReference.RefID.ValueString())
	})

	t.Run("API by_value replaces prior by_reference", func(t *testing.T) {
		shared := &models.DiscoverSessionPanelConfigModel{
			ByReference: &models.DiscoverSessionPanelByRefModel{
				TimeRange: &models.TimeRangeModel{
					From: types.StringValue("now-30m"),
					To:   types.StringValue("now"),
					Mode: types.StringNull(),
				},
				RefID: types.StringValue("old-ref-id"),
			},
		}
		pm := models.PanelModel{DiscoverSessionConfig: shared}
		tfPanel := models.PanelModel{DiscoverSessionConfig: shared}

		tabItem := kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item{}
		dsl := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0{
			Query: kbapi.KbnAsCodeQuery{Expression: `host.name : "x"`, Language: kbapi.Kql},
		}
		require.NoError(t, json.Unmarshal([]byte(`{"type":"data_view_reference","id":"metrics-*"}`), &dsl.DataSource))
		require.NoError(t, tabItem.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0(dsl))

		cfg0 := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0{
			TimeRange: kbapi.KbnEsQueryServerTimeRangeSchema{From: "now-30m", To: "now"},
			Tabs:      []kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item{tabItem},
		}
		var u kbapi.KbnDashboardPanelTypeDiscoverSession_Config
		require.NoError(t, u.FromKbnDashboardPanelTypeDiscoverSessionConfig0(cfg0))
		apiPanel := kbapi.KbnDashboardPanelTypeDiscoverSession{Config: u, Type: kbapi.DiscoverSession, Grid: discoverSessionTestGrid()}

		populateDiscoverSessionPanelFromAPI(ctx, &pm, &tfPanel, apiPanel)

		require.Nil(t, pm.DiscoverSessionConfig.ByReference)
		require.NotNil(t, pm.DiscoverSessionConfig.ByValue)
		require.NotNil(t, pm.DiscoverSessionConfig.ByValue.Tab.DSL)
		assert.Equal(t, `host.name : "x"`, pm.DiscoverSessionConfig.ByValue.Tab.DSL.Query.Expression.ValueString())
	})
}

func Test_discoverSession_byReference_roundTrip(t *testing.T) {
	ctx := context.Background()
	grid := discoverSessionTestGrid()

	pm := models.PanelModel{
		DiscoverSessionConfig: &models.DiscoverSessionPanelConfigModel{
			Title:       types.StringValue("Discover link"),
			Description: types.StringValue("linked panel"),
			ByReference: &models.DiscoverSessionPanelByRefModel{
				TimeRange: &models.TimeRangeModel{
					From: types.StringValue("now-1h"),
					To:   types.StringValue("now"),
					Mode: types.StringNull(),
				},
				RefID:         types.StringValue("saved-discover-abc"),
				SelectedTabID: types.StringValue("tab-explicit"),
				Overrides: &models.DiscoverSessionOverridesModel{
					Density:     types.StringValue("compact"),
					RowsPerPage: types.Int64Value(50),
					SampleSize:  types.Int64Value(500),
					Sort: []models.DiscoverSessionSortModel{
						{Name: types.StringValue("@timestamp"), Direction: types.StringValue("desc")},
					},
				},
			},
		},
	}

	item1, diags := discoverSessionPanelToAPI(ctx, pm, grid, nil, nil)
	require.False(t, diags.HasError(), "%s", diags)

	dsPanel, err := item1.AsKbnDashboardPanelTypeDiscoverSession()
	require.NoError(t, err)

	next := pm
	populateDiscoverSessionPanelFromAPI(ctx, &next, &pm, dsPanel)

	require.Nil(t, next.DiscoverSessionConfig.ByValue)
	br := next.DiscoverSessionConfig.ByReference
	require.NotNil(t, br)
	assert.Equal(t, "saved-discover-abc", br.RefID.ValueString())
	assert.Equal(t, "tab-explicit", br.SelectedTabID.ValueString())
	assert.Equal(t, "now-1h", br.TimeRange.From.ValueString())
	assert.Equal(t, "now", br.TimeRange.To.ValueString())
	require.NotNil(t, br.Overrides)
	assert.Equal(t, "compact", br.Overrides.Density.ValueString())
	assert.Equal(t, int64(50), br.Overrides.RowsPerPage.ValueInt64())
	assert.Equal(t, int64(500), br.Overrides.SampleSize.ValueInt64())
	require.Len(t, br.Overrides.Sort, 1)
	assert.Equal(t, "@timestamp", br.Overrides.Sort[0].Name.ValueString())
	assert.Equal(t, "desc", br.Overrides.Sort[0].Direction.ValueString())

	assert.Equal(t, "Discover link", next.DiscoverSessionConfig.Title.ValueString())
	assert.Equal(t, "linked panel", next.DiscoverSessionConfig.Description.ValueString())

	item2, diags2 := discoverSessionPanelToAPI(ctx, next, grid, nil, nil)
	require.False(t, diags2.HasError(), "%s", diags2)

	raw1, err := item1.MarshalJSON()
	require.NoError(t, err)
	raw2, err := item2.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, string(raw1), string(raw2))
}

func Test_discoverSession_timeRange_inheritsDashboardAtWrite_keepsNullOnRead(t *testing.T) {
	dashTR := &models.TimeRangeModel{
		From: types.StringValue("now-15m"),
		To:   types.StringValue("now"),
		Mode: types.StringNull(),
	}

	pm := models.PanelModel{
		DiscoverSessionConfig: &models.DiscoverSessionPanelConfigModel{
			ByValue: &models.DiscoverSessionPanelByValueModel{
				TimeRange: nil,
				Tab: models.DiscoverSessionTabModel{
					DSL: &models.DiscoverSessionDSLTabModel{
						Query: &models.FilterSimpleModel{
							Language:   types.StringValue("kql"),
							Expression: types.StringValue("*"),
						},
						DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_reference","id":"logs-*"}`),
					},
				},
			},
		},
	}

	item, diags := discoverSessionPanelToAPI(context.Background(), pm, discoverSessionTestGrid(), nil, dashTR)
	require.False(t, diags.HasError())

	dsPanel, err := item.AsKbnDashboardPanelTypeDiscoverSession()
	require.NoError(t, err)
	cfg0, err := dsPanel.Config.AsKbnDashboardPanelTypeDiscoverSessionConfig0()
	require.NoError(t, err)
	assert.Equal(t, "now-15m", cfg0.TimeRange.From)
	assert.Equal(t, "now", cfg0.TimeRange.To)

	next := pm
	populateDiscoverSessionPanelFromAPI(context.Background(), &next, &pm, dsPanel)
	require.Nil(t, next.DiscoverSessionConfig.ByValue.TimeRange)
}

func Test_populateDiscoverSessionPanelFromAPI_import_drilldownDefaults(t *testing.T) {
	encodeTrue := true
	openFalse := false
	dd := []struct {
		EncodeUrl    *bool                                                              `json:"encode_url,omitempty"` //nolint:revive
		Label        string                                                             `json:"label"`
		OpenInNewTab *bool                                                              `json:"open_in_new_tab,omitempty"`
		Trigger      kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsTrigger `json:"trigger"`
		Type         kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsType    `json:"type"`
		Url          string                                                             `json:"url"` //nolint:revive
	}{
		{
			Url:          "https://example.test/x",
			Label:        "open",
			Trigger:      kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsTriggerOnOpenPanelMenu,
			Type:         kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0DrilldownsTypeUrlDrilldown,
			EncodeUrl:    &encodeTrue,
			OpenInNewTab: &openFalse,
		},
	}

	tabItem := kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item{}
	dsl := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0Tabs0{
		Query: kbapi.KbnAsCodeQuery{Expression: "*", Language: kbapi.Kql},
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"data_view_reference","id":"logs-*"}`), &dsl.DataSource))
	require.NoError(t, tabItem.FromKbnDashboardPanelTypeDiscoverSessionConfig0Tabs0(dsl))

	cfg0 := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig0{
		TimeRange:  kbapi.KbnEsQueryServerTimeRangeSchema{From: "now-30m", To: "now"},
		Drilldowns: &dd,
		Tabs:       []kbapi.KbnDashboardPanelTypeDiscoverSession_Config_0_Tabs_Item{tabItem},
	}

	var cfgUnion kbapi.KbnDashboardPanelTypeDiscoverSession_Config
	require.NoError(t, cfgUnion.FromKbnDashboardPanelTypeDiscoverSessionConfig0(cfg0))

	apiPanel := kbapi.KbnDashboardPanelTypeDiscoverSession{
		Config: cfgUnion,
		Type:   kbapi.DiscoverSession,
		Grid:   discoverSessionTestGrid(),
	}

	var pm models.PanelModel
	populateDiscoverSessionPanelFromAPI(context.Background(), &pm, nil, apiPanel)
	require.Len(t, pm.DiscoverSessionConfig.Drilldowns, 1)
	d := pm.DiscoverSessionConfig.Drilldowns[0]
	assert.True(t, d.EncodeURL.IsNull())
	assert.True(t, d.OpenInNewTab.IsNull())
}

func Test_populateDiscoverSessionPanelFromAPI_import_byReference_selectedTabID(t *testing.T) {
	selected := "sel-tab"
	cfg1 := kbapi.KbnDashboardPanelTypeDiscoverSessionConfig1{
		RefId:         "discover-so",
		TimeRange:     kbapi.KbnEsQueryServerTimeRangeSchema{From: "now-30m", To: "now"},
		SelectedTabId: &selected,
	}
	var u kbapi.KbnDashboardPanelTypeDiscoverSession_Config
	require.NoError(t, u.FromKbnDashboardPanelTypeDiscoverSessionConfig1(cfg1))
	apiPanel := kbapi.KbnDashboardPanelTypeDiscoverSession{Config: u, Type: kbapi.DiscoverSession, Grid: discoverSessionTestGrid()}

	var pm models.PanelModel
	populateDiscoverSessionPanelFromAPI(context.Background(), &pm, nil, apiPanel)
	require.NotNil(t, pm.DiscoverSessionConfig.ByReference)
	assert.Equal(t, "sel-tab", pm.DiscoverSessionConfig.ByReference.SelectedTabID.ValueString())
}
