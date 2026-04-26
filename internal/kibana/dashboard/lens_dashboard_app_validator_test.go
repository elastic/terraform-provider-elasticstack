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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_lensDashboardAppConfigModeValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := lensDashboardAppConfigModeValidator{}

	byValueAttrs := map[string]attr.Type{
		"config_json": jsontypes.NormalizedType{},
	}
	byValueObj := func() types.Object {
		return types.ObjectValueMust(byValueAttrs, map[string]attr.Value{
			"config_json": jsontypes.NewNormalizedValue(`{"a":1}`),
		})
	}

	byRefAttrs := map[string]attr.Type{
		"ref_id":          types.StringType,
		"references_json": jsontypes.NormalizedType{},
		"title":           types.StringType,
		"description":     types.StringType,
		"hide_title":      types.BoolType,
		"hide_border":     types.BoolType,
		"drilldowns_json": jsontypes.NormalizedType{},
		"time_range":      types.ObjectType{AttrTypes: map[string]attr.Type{"from": types.StringType, "to": types.StringType, "mode": types.StringType}},
	}
	byRefObj := func() types.Object {
		tr := types.ObjectValueMust(
			map[string]attr.Type{"from": types.StringType, "to": types.StringType, "mode": types.StringType},
			map[string]attr.Value{
				"from": types.StringValue("now-1h"),
				"to":   types.StringValue("now"),
				"mode": types.StringNull(),
			},
		)
		return types.ObjectValueMust(byRefAttrs, map[string]attr.Value{
			"ref_id":          types.StringValue("ref1"),
			"references_json": jsontypes.NewNormalizedNull(),
			"title":           types.StringNull(),
			"description":     types.StringNull(),
			"hide_title":      types.BoolNull(),
			"hide_border":     types.BoolNull(),
			"drilldowns_json": jsontypes.NewNormalizedNull(),
			"time_range":      tr,
		})
	}

	ldaAttrTypes := map[string]attr.Type{
		"by_value":     types.ObjectType{AttrTypes: byValueAttrs},
		"by_reference": types.ObjectType{AttrTypes: byRefAttrs},
	}

	t.Run("accepts by_value only", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(ldaAttrTypes, map[string]attr.Value{
			"by_value":     byValueObj(),
			"by_reference": types.ObjectNull(byRefAttrs),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("lens_dashboard_app_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("accepts by_reference only", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(ldaAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectNull(byValueAttrs),
			"by_reference": byRefObj(),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("lens_dashboard_app_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("rejects both by_value and by_reference", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(ldaAttrTypes, map[string]attr.Value{
			"by_value":     byValueObj(),
			"by_reference": byRefObj(),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("lens_dashboard_app_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "not both")
	})

	t.Run("rejects neither by_value nor by_reference", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(ldaAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectNull(byValueAttrs),
			"by_reference": types.ObjectNull(byRefAttrs),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("lens_dashboard_app_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("defers when by_value is unknown", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(ldaAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectUnknown(byValueAttrs),
			"by_reference": types.ObjectNull(byRefAttrs),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("lens_dashboard_app_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("defers when by_reference is unknown", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(ldaAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectNull(byValueAttrs),
			"by_reference": types.ObjectUnknown(byRefAttrs),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("lens_dashboard_app_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("no-op on null or unknown object", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			name string
			val  attr.Value
		}{
			{"null", types.ObjectNull(ldaAttrTypes)},
			{"unknown", types.ObjectUnknown(ldaAttrTypes)},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ov, ok := tc.val.(types.Object)
				require.True(t, ok)
				resp := validator.ObjectResponse{}
				v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("x")}, &resp)
				require.False(t, resp.Diagnostics.HasError())
			})
		}
	})
}

func Test_lensDashboardAppByReferenceTimeRangeModeStringValidators(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	panel := getPanelSchema()
	lda, ok := panel.Attributes["lens_dashboard_app_config"].(schema.SingleNestedAttribute)
	require.True(t, ok, "lens_dashboard_app_config should be a SingleNestedAttribute")
	byRef, ok := lda.Attributes["by_reference"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	tRange, ok := byRef.Attributes["time_range"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	modeAttr, ok := tRange.Attributes["mode"].(schema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, modeAttr.Validators)
	req := validator.StringRequest{Path: path.Root("mode"), ConfigValue: types.StringValue("invalid")}
	var resp validator.StringResponse
	for _, m := range modeAttr.Validators {
		m.ValidateString(ctx, req, &resp)
	}
	require.True(t, resp.Diagnostics.HasError())
}

func lensByValueAttributeTypes(t *testing.T) map[string]attr.Type {
	t.Helper()
	lda := getLensDashboardAppConfigSchema()
	bv, ok := lda["by_value"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	return bv.GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func lensByValueObjectAllNull(t *testing.T) types.Object {
	t.Helper()
	typesMap := lensByValueAttributeTypes(t)
	vals := make(map[string]attr.Value, len(typesMap))
	for k, at := range typesMap {
		switch k {
		case "config_json":
			vals[k] = jsontypes.NewNormalizedNull()
		default:
			ot, ok := at.(types.ObjectType)
			require.True(t, ok, "expected types.ObjectType for %s, got %T", k, at)
			vals[k] = types.ObjectNull(ot.AttrTypes)
		}
	}
	return types.ObjectValueMust(typesMap, vals)
}

func Test_lensDashboardAppByValueSourceValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := lensDashboardAppByValueSourceValidator{}

	lda := getLensDashboardAppConfigSchema()
	bv, ok := lda["by_value"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	var found bool
	for _, val := range bv.Validators {
		if _, ok := val.(lensDashboardAppByValueSourceValidator); ok {
			found = true
			break
		}
	}
	require.True(t, found, "by_value should include lensDashboardAppByValueSourceValidator")

	t.Run("rejects no source", func(t *testing.T) {
		t.Parallel()
		ov := lensByValueObjectAllNull(t)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("by_value")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "exactly one")
	})

	t.Run("accepts config_json only", func(t *testing.T) {
		t.Parallel()
		typesMap := lensByValueAttributeTypes(t)
		vals := make(map[string]attr.Value, len(typesMap))
		for k, at := range typesMap {
			switch k {
			case "config_json":
				vals[k] = jsontypes.NewNormalizedValue(`{"type":"metric"}`)
			default:
				ot := at.(types.ObjectType)
				vals[k] = types.ObjectNull(ot.AttrTypes)
			}
		}
		ov := types.ObjectValueMust(typesMap, vals)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("by_value")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("defers when a source is unknown", func(t *testing.T) {
		t.Parallel()
		typesMap := lensByValueAttributeTypes(t)
		vals := make(map[string]attr.Value, len(typesMap))
		for k, at := range typesMap {
			switch k {
			case "config_json":
				vals[k] = jsontypes.NewNormalizedUnknown()
			default:
				ot := at.(types.ObjectType)
				vals[k] = types.ObjectNull(ot.AttrTypes)
			}
		}
		ov := types.ObjectValueMust(typesMap, vals)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("by_value")}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})
}
