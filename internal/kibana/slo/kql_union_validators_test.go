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

package slo

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// kqlSiblingsTestSchema models filter + filter_kql as root siblings so path expressions match
// kql_custom_indicator (same parent/child path layout for MatchRelative().AtParent().AtName).
func kqlSiblingsTestSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filter": schema.StringAttribute{Optional: true},
			"filter_kql": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"kql_query": schema.StringAttribute{Optional: true, Computed: true},
					"filters": schema.ListNestedAttribute{
						Optional: true,
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"query": schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}},
							},
						},
					},
				},
			},
		},
	}
}

func testKqlObject(t *testing.T) types.Object {
	t.Helper()
	emptyFilters := types.ListValueMust(tfKqlFilterRowObjectType, nil)
	return types.ObjectValueMust(tfKqlKqlObjectAttrTypes, map[string]attr.Value{
		"kql_query": types.StringValue("host.name:*"),
		"filters":   emptyFilters,
	})
}

func testConfigFilterAndKQL(t *testing.T, filterVal tftypes.Value, kqlObj types.Object) tfsdk.Config {
	t.Helper()
	ktf, err := kqlObj.ToTerraformValue(context.Background())
	require.NoError(t, err)
	sch := kqlSiblingsTestSchema()
	raw := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"filter":     tftypes.String,
			"filter_kql": ktf.Type(),
		}},
		map[string]tftypes.Value{
			"filter":     filterVal,
			"filter_kql": ktf,
		},
	)
	return tfsdk.Config{Raw: raw, Schema: sch}
}

func TestKqlObjectFormMeaningful(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	emptyList := types.ListValueMust(tfKqlFilterRowObjectType, nil)

	t.Run("rejects known-empty object form", func(t *testing.T) {
		t.Parallel()
		obj := types.ObjectValueMust(tfKqlKqlObjectAttrTypes, map[string]attr.Value{
			"kql_query": types.StringValue(""),
			"filters":   emptyList,
		})
		var resp validator.ObjectResponse
		kqlObjectFormMeaningful{}.ValidateObject(ctx, validator.ObjectRequest{
			Path:        path.Root("filter_kql"),
			Config:      tfsdk.Config{},
			ConfigValue: obj,
		}, &resp)
		require.True(t, resp.Diagnostics.HasError(), "empty kql_query and no filters should fail")
	})

	t.Run("allows non-blank kql_query", func(t *testing.T) {
		t.Parallel()
		obj := types.ObjectValueMust(tfKqlKqlObjectAttrTypes, map[string]attr.Value{
			"kql_query": types.StringValue("host.name:*"),
			"filters":   emptyList,
		})
		var resp validator.ObjectResponse
		kqlObjectFormMeaningful{}.ValidateObject(ctx, validator.ObjectRequest{
			Path:        path.Root("filter_kql"),
			Config:      tfsdk.Config{},
			ConfigValue: obj,
		}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})
}

func TestKqlLegacyStringExclusiveWithObject_siblings(t *testing.T) {
	t.Parallel()
	v := kqlLegacyStringExclusiveWithObject{parallelObjectAttr: "filter_kql", treatEmptyStringAsUnset: true}
	kq := testKqlObject(t)

	t.Run("conflict when both string and object set", func(t *testing.T) {
		t.Parallel()
		cfg := testConfigFilterAndKQL(t, tftypes.NewValue(tftypes.String, "host:*"), kq)
		var resp validator.StringResponse
		v.ValidateString(context.Background(), validator.StringRequest{
			Path:        path.Root("filter"),
			ConfigValue: types.StringValue("host:*"),
			Config:      cfg,
		}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("no conflict when filter is empty and object is set", func(t *testing.T) {
		t.Parallel()
		cfg := testConfigFilterAndKQL(t, tftypes.NewValue(tftypes.String, ""), kq)
		var resp validator.StringResponse
		v.ValidateString(context.Background(), validator.StringRequest{
			Path:        path.Root("filter"),
			ConfigValue: types.StringValue(""),
			Config:      cfg,
		}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})
}

func TestKqlObjectFormExclusiveWithString_siblings(t *testing.T) {
	t.Parallel()
	v := kqlObjectFormExclusiveWithString{parallelStringAttr: "filter", treatEmptyStringAsUnset: true}
	kq := testKqlObject(t)

	t.Run("conflict when both meaningfully set", func(t *testing.T) {
		t.Parallel()
		cfg := testConfigFilterAndKQL(t, tftypes.NewValue(tftypes.String, "a"), kq)
		var resp validator.ObjectResponse
		v.ValidateObject(context.Background(), validator.ObjectRequest{
			Path:        path.Root("filter_kql"),
			ConfigValue: kq,
			Config:      cfg,
		}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})
}

// goodSiblingsTestSchema matches good + good_kql layout under a single parent object.
func goodSiblingsTestSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"good": schema.StringAttribute{Optional: true},
			"good_kql": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"kql_query": schema.StringAttribute{Optional: true, Computed: true},
					"filters": schema.ListNestedAttribute{
						Optional: true,
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"query": schema.StringAttribute{Optional: true, Computed: true, CustomType: jsontypes.NormalizedType{}},
							},
						},
					},
				},
			},
		},
	}
}

func testConfigGoodAndGoodKql(t *testing.T, goodVal tftypes.Value, kqlObj types.Object) tfsdk.Config {
	t.Helper()
	ktf, err := kqlObj.ToTerraformValue(context.Background())
	require.NoError(t, err)
	sch := goodSiblingsTestSchema()
	raw := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"good":     tftypes.String,
			"good_kql": ktf.Type(),
		}},
		map[string]tftypes.Value{
			"good":     goodVal,
			"good_kql": ktf,
		},
	)
	return tfsdk.Config{Raw: raw, Schema: sch}
}

func TestKqlLegacyStringExclusiveWithObject_goodPair(t *testing.T) {
	t.Parallel()
	v := kqlLegacyStringExclusiveWithObject{parallelObjectAttr: "good_kql", treatEmptyStringAsUnset: true}
	kq := testKqlObject(t)

	t.Run("conflict when string good and good_kql both set", func(t *testing.T) {
		t.Parallel()
		cfg := testConfigGoodAndGoodKql(t, tftypes.NewValue(tftypes.String, "status:ok"), kq)
		var resp validator.StringResponse
		v.ValidateString(context.Background(), validator.StringRequest{
			Path:        path.Root("good"),
			ConfigValue: types.StringValue("status:ok"),
			Config:      cfg,
		}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})
}
