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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func testPinnedDisplaySettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"placeholder":     types.StringType,
		"hide_action_bar": types.BoolType,
		"hide_exclude":    types.BoolType,
		"hide_exists":     types.BoolType,
		"hide_sort":       types.BoolType,
	}
}

func testPinnedSortAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"by":        types.StringType,
		"direction": types.StringType,
	}
}

func testPinnedOptionsListControlAttrTypes() map[string]attr.Type {
	ds := testPinnedDisplaySettingsAttrTypes()
	st := testPinnedSortAttrTypes()
	return map[string]attr.Type{
		"data_view_id":       types.StringType,
		"field_name":         types.StringType,
		"title":              types.StringType,
		"use_global_filters": types.BoolType,
		"ignore_validations": types.BoolType,
		"single_select":      types.BoolType,
		"exclude":            types.BoolType,
		"exists_selected":    types.BoolType,
		"run_past_timeout":   types.BoolType,
		"search_technique":   types.StringType,
		"selected_options":   types.ListType{ElemType: types.StringType},
		"display_settings":   types.ObjectType{AttrTypes: ds},
		"sort":               types.ObjectType{AttrTypes: st},
	}
}

func testPinnedOptionsListControlObject() types.Object {
	t := testPinnedOptionsListControlAttrTypes()
	dsTypes := testPinnedDisplaySettingsAttrTypes()
	sortTypes := testPinnedSortAttrTypes()
	return types.ObjectValueMust(t, map[string]attr.Value{
		"data_view_id":       types.StringValue("dv1"),
		"field_name":         types.StringValue("status"),
		"title":              types.StringNull(),
		"use_global_filters": types.BoolNull(),
		"ignore_validations": types.BoolNull(),
		"single_select":      types.BoolNull(),
		"exclude":            types.BoolNull(),
		"exists_selected":    types.BoolNull(),
		"run_past_timeout":   types.BoolNull(),
		"search_technique":   types.StringNull(),
		"selected_options":   types.ListNull(types.StringType),
		"display_settings":   types.ObjectNull(dsTypes),
		"sort":               types.ObjectNull(sortTypes),
	})
}

func testPinnedRangeSliderControlAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"title":              types.StringType,
		"data_view_id":       types.StringType,
		"field_name":         types.StringType,
		"use_global_filters": types.BoolType,
		"ignore_validations": types.BoolType,
		"value":              types.ListType{ElemType: types.StringType},
		"step":               types.Float32Type,
	}
}

func testPinnedRangeSliderControlObject() types.Object {
	t := testPinnedRangeSliderControlAttrTypes()
	return types.ObjectValueMust(t, map[string]attr.Value{
		"title":              types.StringNull(),
		"data_view_id":       types.StringValue("dv1"),
		"field_name":         types.StringValue("bytes"),
		"use_global_filters": types.BoolNull(),
		"ignore_validations": types.BoolNull(),
		"value":              types.ListNull(types.StringType),
		"step":               types.Float32Null(),
	})
}

func testPinnedTimeSliderControlAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"start_percentage_of_time_range": types.Float32Type,
		"end_percentage_of_time_range":   types.Float32Type,
		"is_anchored":                    types.BoolType,
	}
}

func testPinnedTimeSliderControlObject() types.Object {
	t := testPinnedTimeSliderControlAttrTypes()
	return types.ObjectValueMust(t, map[string]attr.Value{
		"start_percentage_of_time_range": types.Float32Null(),
		"end_percentage_of_time_range":   types.Float32Null(),
		"is_anchored":                    types.BoolNull(),
	})
}

func testPinnedEsqlDisplaySettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"placeholder":     types.StringType,
		"hide_action_bar": types.BoolType,
		"hide_exclude":    types.BoolType,
		"hide_exists":     types.BoolType,
	}
}

func testPinnedEsqlControlAttrTypes() map[string]attr.Type {
	ds := testPinnedEsqlDisplaySettingsAttrTypes()
	return map[string]attr.Type{
		"selected_options":  types.ListType{ElemType: types.StringType},
		"variable_name":     types.StringType,
		"variable_type":     types.StringType,
		"esql_query":        types.StringType,
		"control_type":      types.StringType,
		"title":             types.StringType,
		"single_select":     types.BoolType,
		"available_options": types.ListType{ElemType: types.StringType},
		"display_settings":  types.ObjectType{AttrTypes: ds},
	}
}

func testPinnedEsqlControlObject() types.Object {
	t := testPinnedEsqlControlAttrTypes()
	dsTypes := testPinnedEsqlDisplaySettingsAttrTypes()
	return types.ObjectValueMust(t, map[string]attr.Value{
		"selected_options":  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("opt")}),
		"variable_name":     types.StringValue("host"),
		"variable_type":     types.StringValue("fields"),
		"esql_query":        types.StringValue("FROM logs-* | KEEP host.name"),
		"control_type":      types.StringValue("VALUES_FROM_QUERY"),
		"title":             types.StringNull(),
		"single_select":     types.BoolNull(),
		"available_options": types.ListNull(types.StringType),
		"display_settings":  types.ObjectNull(dsTypes),
	})
}

func testPinnedPanelRootAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":                        types.StringType,
		"time_slider_control_config":  types.ObjectType{AttrTypes: testPinnedTimeSliderControlAttrTypes()},
		"esql_control_config":         types.ObjectType{AttrTypes: testPinnedEsqlControlAttrTypes()},
		"options_list_control_config": types.ObjectType{AttrTypes: testPinnedOptionsListControlAttrTypes()},
		"range_slider_control_config": types.ObjectType{AttrTypes: testPinnedRangeSliderControlAttrTypes()},
	}
}

func testPinnedPanelObject(typeVal attr.Value, ts, esql, ol, rs attr.Value) types.Object {
	rootTypes := testPinnedPanelRootAttrTypes()
	return types.ObjectValueMust(rootTypes, map[string]attr.Value{
		"type":                        typeVal,
		"time_slider_control_config":  ts,
		"esql_control_config":         esql,
		"options_list_control_config": ol,
		"range_slider_control_config": rs,
	})
}

func Test_pinnedPanelControlValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := pinnedPanelControlValidator{}
	nullTS := types.ObjectNull(testPinnedTimeSliderControlAttrTypes())
	nullEsql := types.ObjectNull(testPinnedEsqlControlAttrTypes())
	nullOL := types.ObjectNull(testPinnedOptionsListControlAttrTypes())
	nullRS := types.ObjectNull(testPinnedRangeSliderControlAttrTypes())

	typeStr := func(s string) attr.Value { return types.StringValue(s) }

	type testCase struct {
		name        string
		obj         types.Object
		wantErr     bool
		errSummary  string
		errContains []string
	}

	tests := []testCase{
		{
			name: "valid options_list_control with matching block only",
			obj: testPinnedPanelObject(
				typeStr(panelTypeOptionsListControl),
				nullTS, nullEsql, testPinnedOptionsListControlObject(), nullRS,
			),
			wantErr: false,
		},
		{
			name: "valid range_slider_control with matching block only",
			obj: testPinnedPanelObject(
				typeStr(panelTypeRangeSlider),
				nullTS, nullEsql, nullOL, testPinnedRangeSliderControlObject(),
			),
			wantErr: false,
		},
		{
			name: "valid time_slider_control with matching block only",
			obj: testPinnedPanelObject(
				typeStr(panelTypeTimeSlider),
				testPinnedTimeSliderControlObject(), nullEsql, nullOL, nullRS,
			),
			wantErr: false,
		},
		{
			name: "valid esql_control with matching block only",
			obj: testPinnedPanelObject(
				typeStr(panelTypeEsqlControl),
				nullTS, testPinnedEsqlControlObject(), nullOL, nullRS,
			),
			wantErr: false,
		},
		{
			name: "mismatched range_slider type with options_list block only",
			obj: testPinnedPanelObject(
				typeStr(panelTypeRangeSlider),
				nullTS, nullEsql, testPinnedOptionsListControlObject(), nullRS,
			),
			wantErr:    true,
			errSummary: "Pinned panel control does not match type",
			errContains: []string{
				"range_slider_control",
				"options_list_control_config",
				"range_slider_control_config",
			},
		},
		{
			name: "multiple blocks set",
			obj: testPinnedPanelObject(
				typeStr(panelTypeOptionsListControl),
				nullTS, nullEsql, testPinnedOptionsListControlObject(), testPinnedRangeSliderControlObject(),
			),
			wantErr:    true,
			errSummary: "Invalid pinned panel entry configuration",
			errContains: []string{
				"exactly one",
				"`options_list_control_config`",
				"`range_slider_control_config`",
			},
		},
		{
			name: "missing block for known options_list_control type",
			obj: testPinnedPanelObject(
				typeStr(panelTypeOptionsListControl),
				nullTS, nullEsql, nullOL, nullRS,
			),
			wantErr:    true,
			errSummary: "Missing pinned panel control configuration",
			errContains: []string{
				`type = "options_list_control"`,
				"`options_list_control_config`",
			},
		},
		{
			name: "invalid pinned panel type lens",
			obj: testPinnedPanelObject(
				typeStr("lens"),
				nullTS, nullEsql, nullOL, nullRS,
			),
			wantErr:    true,
			errSummary: "Invalid pinned panel entry type",
			errContains: []string{
				`got "lens"`,
			},
		},
		{
			name: "deferred unknown type with one block set",
			obj: testPinnedPanelObject(
				types.StringUnknown(),
				nullTS, nullEsql, testPinnedOptionsListControlObject(), nullRS,
			),
			wantErr: false,
		},
		{
			name: "deferred known type with zero blocks and unknown slot",
			obj: testPinnedPanelObject(
				typeStr(panelTypeOptionsListControl),
				nullTS, nullEsql, nullOL, types.ObjectUnknown(testPinnedRangeSliderControlAttrTypes()),
			),
			wantErr: false,
		},
		{
			name: "deferred known type with correct block and another unknown slot",
			obj: testPinnedPanelObject(
				typeStr(panelTypeOptionsListControl),
				nullTS, nullEsql, testPinnedOptionsListControlObject(), types.ObjectUnknown(testPinnedRangeSliderControlAttrTypes()),
			),
			wantErr: false,
		},
		{
			name: "invalid type still errors when another slot is unknown",
			obj: testPinnedPanelObject(
				typeStr("lens"),
				nullTS, nullEsql, testPinnedOptionsListControlObject(), types.ObjectUnknown(testPinnedRangeSliderControlAttrTypes()),
			),
			wantErr:    true,
			errSummary: "Invalid pinned panel entry type",
		},
		{
			name: "two blocks set with unknown type still errors definitively",
			obj: testPinnedPanelObject(
				types.StringUnknown(),
				nullTS, nullEsql, testPinnedOptionsListControlObject(), testPinnedRangeSliderControlObject(),
			),
			wantErr:    true,
			errSummary: "Invalid pinned panel entry configuration",
			errContains: []string{
				"exactly one",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := validator.ObjectResponse{}
			v.ValidateObject(ctx, validator.ObjectRequest{
				ConfigValue: tc.obj,
				Path:        path.Root("pinned_panels").AtListIndex(0),
			}, &resp)

			if !tc.wantErr {
				require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
				return
			}

			require.True(t, resp.Diagnostics.HasError(), "expected error diagnostics")
			require.Equal(t, tc.errSummary, resp.Diagnostics.Errors()[0].Summary())
			detail := resp.Diagnostics.Errors()[0].Detail()
			for _, frag := range tc.errContains {
				require.Contains(t, detail, frag)
			}
		})
	}
}
