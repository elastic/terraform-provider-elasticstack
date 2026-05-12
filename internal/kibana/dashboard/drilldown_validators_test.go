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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func structuredDrilldownItemAttrTypesForTest(t *testing.T) map[string]attr.Type {
	t.Helper()
	dashboardObject := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"dashboard_id":    types.StringType,
			"label":           types.StringType,
			"use_filters":     types.BoolType,
			"use_time_range":  types.BoolType,
			"open_in_new_tab": types.BoolType,
		},
	}
	discoverObject := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"label":           types.StringType,
			"open_in_new_tab": types.BoolType,
		},
	}
	urlObject := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"url":             types.StringType,
			"label":           types.StringType,
			"trigger":         types.StringType,
			"encode_url":      types.BoolType,
			"open_in_new_tab": types.BoolType,
		},
	}
	return map[string]attr.Type{
		"dashboard": dashboardObject,
		"discover":  discoverObject,
		"url":       urlObject,
	}
}

func Test_drilldownItemModeValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	attrTypes := structuredDrilldownItemAttrTypesForTest(t)
	v := drilldownItemModeValidator{}

	pathRoot := path.Root("drilldowns").AtListIndex(0)

	t.Run("rejects_zero_sub_blocks", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(attrTypes, map[string]attr.Value{
			"dashboard": types.ObjectNull(attrTypes["dashboard"].(types.ObjectType).AttrTypes),
			"discover":  types.ObjectNull(attrTypes["discover"].(types.ObjectType).AttrTypes),
			"url":       types.ObjectNull(attrTypes["url"].(types.ObjectType).AttrTypes),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: pathRoot}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "exactly one")
	})

	t.Run("rejects_multiple_sub_blocks", func(t *testing.T) {
		t.Parallel()
		dashVals := map[string]attr.Value{
			"dashboard_id":    types.StringValue("d1"),
			"label":           types.StringValue("a"),
			"use_filters":     types.BoolNull(),
			"use_time_range":  types.BoolNull(),
			"open_in_new_tab": types.BoolNull(),
		}
		urlVals := map[string]attr.Value{
			"url":             types.StringValue("https://x"),
			"label":           types.StringValue("u"),
			"trigger":         types.StringValue("on_click_value"),
			"encode_url":      types.BoolNull(),
			"open_in_new_tab": types.BoolNull(),
		}
		ov := types.ObjectValueMust(attrTypes, map[string]attr.Value{
			"dashboard": types.ObjectValueMust(attrTypes["dashboard"].(types.ObjectType).AttrTypes, dashVals),
			"discover":  types.ObjectNull(attrTypes["discover"].(types.ObjectType).AttrTypes),
			"url":       types.ObjectValueMust(attrTypes["url"].(types.ObjectType).AttrTypes, urlVals),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: pathRoot}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "mutually exclusive")
	})

	t.Run("accepts_single_dashboard", func(t *testing.T) {
		t.Parallel()
		dashVals := map[string]attr.Value{
			"dashboard_id":    types.StringValue("d1"),
			"label":           types.StringValue("a"),
			"use_filters":     types.BoolNull(),
			"use_time_range":  types.BoolNull(),
			"open_in_new_tab": types.BoolNull(),
		}
		ov := types.ObjectValueMust(attrTypes, map[string]attr.Value{
			"dashboard": types.ObjectValueMust(attrTypes["dashboard"].(types.ObjectType).AttrTypes, dashVals),
			"discover":  types.ObjectNull(attrTypes["discover"].(types.ObjectType).AttrTypes),
			"url":       types.ObjectNull(attrTypes["url"].(types.ObjectType).AttrTypes),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: pathRoot}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("no_op_on_null_unknown", func(t *testing.T) {
		t.Parallel()
		for _, ov := range []attr.Value{
			types.ObjectNull(attrTypes),
			types.ObjectUnknown(attrTypes),
		} {
			obj, ok := ov.(types.Object)
			require.True(t, ok)
			var resp validator.ObjectResponse
			v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: obj, Path: pathRoot}, &resp)
			require.False(t, resp.Diagnostics.HasError())
		}
	})
}

func Test_structuredDrilldown_urlTriggerStringValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	listAttr := getStructuredDrilldownsAttribute().(schema.ListNestedAttribute)
	urlAttr, ok := listAttr.NestedObject.Attributes["url"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	triggerAttr, ok := urlAttr.Attributes["trigger"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, triggerAttr.Required)
	require.NotEmpty(t, triggerAttr.Validators)

	t.Run("rejects_invalid", func(t *testing.T) {
		t.Parallel()
		req := validator.StringRequest{Path: path.Root("trigger"), ConfigValue: types.StringValue("nope")}
		var resp validator.StringResponse
		for _, m := range triggerAttr.Validators {
			m.ValidateString(ctx, req, &resp)
		}
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("allows_known", func(t *testing.T) {
		t.Parallel()
		req := validator.StringRequest{Path: path.Root("trigger"), ConfigValue: types.StringValue("on_click_row")}
		var resp validator.StringResponse
		for _, m := range triggerAttr.Validators {
			m.ValidateString(ctx, req, &resp)
		}
		require.False(t, resp.Diagnostics.HasError())
	})
}
