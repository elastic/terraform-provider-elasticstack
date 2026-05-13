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

func Test_visConfigModeValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := visConfigModeValidator{}

	visAttrs := getVisConfigSchema()
	visAttrTypes := make(map[string]attr.Type)
	for name, attrDef := range visAttrs {
		visAttrTypes[name] = attrDef.GetType()
	}
	byValueAttrs := visAttrTypes["by_value"].(types.ObjectType).AttrTypes
	byRefAttrs := visAttrTypes["by_reference"].(types.ObjectType).AttrTypes
	drillElemType := byRefAttrs["drilldowns"].(types.ListType).ElemType

	byValueObj := func() types.Object {
		return visByValueObjectAllChartsNull(t, byValueAttrs)
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
			"drilldowns":      types.ListNull(drillElemType),
			"time_range":      tr,
		})
	}

	t.Run("accepts by_value only", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(visAttrTypes, map[string]attr.Value{
			"by_value":     byValueObj(),
			"by_reference": types.ObjectNull(byRefAttrs),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("accepts by_reference only", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(visAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectNull(byValueAttrs),
			"by_reference": byRefObj(),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("rejects both by_value and by_reference", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(visAttrTypes, map[string]attr.Value{
			"by_value":     byValueObj(),
			"by_reference": byRefObj(),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "not both")
	})

	t.Run("rejects neither by_value nor by_reference", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(visAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectNull(byValueAttrs),
			"by_reference": types.ObjectNull(byRefAttrs),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "Exactly one")
	})

	t.Run("defers when by_value is unknown", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(visAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectUnknown(byValueAttrs),
			"by_reference": types.ObjectNull(byRefAttrs),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("defers when by_reference is unknown", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(visAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectNull(byValueAttrs),
			"by_reference": types.ObjectUnknown(byRefAttrs),
		})
		resp := validator.ObjectResponse{}
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("no-op on null or unknown object", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			name string
			val  attr.Value
		}{
			{"null", types.ObjectNull(visAttrTypes)},
			{"unknown", types.ObjectUnknown(visAttrTypes)},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ov, ok := tc.val.(types.Object)
				require.True(t, ok)
				resp := validator.ObjectResponse{}
				v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config")}, &resp)
				require.False(t, resp.Diagnostics.HasError())
			})
		}
	})
}

func visByValueObjectAllChartsNull(t *testing.T, typesMap map[string]attr.Type) types.Object {
	t.Helper()
	vals := visByValueAllNullAttributeValues(t, typesMap)
	return types.ObjectValueMust(typesMap, vals)
}

func visByValueAllNullAttributeValues(t *testing.T, typesMap map[string]attr.Type) map[string]attr.Value {
	t.Helper()
	vals := make(map[string]attr.Value, len(typesMap))
	for k, at := range typesMap {
		vals[k] = types.ObjectNull(at.(types.ObjectType).AttrTypes)
	}
	return vals
}

func visByValueAttributeTypes(t *testing.T) map[string]attr.Type {
	t.Helper()
	visSchema := getVisConfigSchema()
	bv, ok := visSchema["by_value"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	return bv.GetType().(types.ObjectType).AttrTypes
}

func Test_visByValueSourceValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := visByValueSourceValidator{}

	visSchema := getVisConfigSchema()
	bv, ok := visSchema["by_value"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	var found bool
	for _, val := range bv.Validators {
		if _, ok := val.(visByValueSourceValidator); ok {
			found = true
			break
		}
	}
	require.True(t, found, "vis_config.by_value must include visByValueSourceValidator")

	t.Run("rejects no chart", func(t *testing.T) {
		t.Parallel()
		typesMap := visByValueAttributeTypes(t)
		ov := types.ObjectValueMust(typesMap, visByValueAllNullAttributeValues(t, typesMap))
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config.by_value")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "exactly one")
	})

	t.Run("defers when a chart attribute is unknown", func(t *testing.T) {
		t.Parallel()
		typesMap := visByValueAttributeTypes(t)
		vals := visByValueAllNullAttributeValues(t, typesMap)
		vals["metric_chart_config"] = types.ObjectUnknown(typesMap["metric_chart_config"].(types.ObjectType).AttrTypes)
		ov := types.ObjectValueMust(typesMap, vals)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config.by_value")}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("accepts one typed chart block only", func(t *testing.T) {
		t.Parallel()
		typesMap := visByValueAttributeTypes(t)
		vals := visByValueAllNullAttributeValues(t, typesMap)
		metricOT := typesMap["metric_chart_config"].(types.ObjectType)
		mv, err := knownNullableObjectAllFieldsNull(ctx, t, metricOT)
		require.NoError(t, err)
		vals["metric_chart_config"] = mv
		ov := types.ObjectValueMust(typesMap, vals)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config.by_value")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("rejects two typed chart blocks", func(t *testing.T) {
		t.Parallel()
		typesMap := visByValueAttributeTypes(t)
		vals := visByValueAllNullAttributeValues(t, typesMap)
		mOT := typesMap["metric_chart_config"].(types.ObjectType)
		pOT := typesMap["pie_chart_config"].(types.ObjectType)
		mv, err := knownNullableObjectAllFieldsNull(ctx, t, mOT)
		require.NoError(t, err)
		pv, err := knownNullableObjectAllFieldsNull(ctx, t, pOT)
		require.NoError(t, err)
		vals["metric_chart_config"] = mv
		vals["pie_chart_config"] = pv
		ov := types.ObjectValueMust(typesMap, vals)
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("vis_config.by_value")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		d := resp.Diagnostics.Errors()[0]
		require.Equal(t, "Invalid vis_config.by_value", d.Summary())
		require.Contains(t, d.Detail(), "exactly one")
	})
}
