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

package markdown

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_configModeValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
		AttrNames:     []string{"by_value", "by_reference"},
		Summary:       "Invalid markdown_config",
		MissingDetail: "Exactly one of `by_value` or `by_reference` must be set inside `markdown_config`.",
		TooManyDetail: "Exactly one of `by_value` or `by_reference` must be set inside `markdown_config`, not both.",
	})

	settingsAttrs := map[string]attr.Type{
		"open_links_in_new_tab": types.BoolType,
	}
	byValueAttrs := map[string]attr.Type{
		"content":     types.StringType,
		"settings":    types.ObjectType{AttrTypes: settingsAttrs},
		"description": types.StringType,
		"hide_title":  types.BoolType,
		"title":       types.StringType,
		"hide_border": types.BoolType,
	}
	byRefAttrs := map[string]attr.Type{
		"ref_id":      types.StringType,
		"description": types.StringType,
		"hide_title":  types.BoolType,
		"title":       types.StringType,
		"hide_border": types.BoolType,
	}
	mcAttrTypes := map[string]attr.Type{
		"by_value":     types.ObjectType{AttrTypes: byValueAttrs},
		"by_reference": types.ObjectType{AttrTypes: byRefAttrs},
	}

	byValueObj := func() types.Object {
		settings := types.ObjectValueMust(settingsAttrs, map[string]attr.Value{
			"open_links_in_new_tab": types.BoolValue(true),
		})
		return types.ObjectValueMust(byValueAttrs, map[string]attr.Value{
			"content":     types.StringValue("# hi"),
			"settings":    settings,
			"description": types.StringNull(),
			"hide_title":  types.BoolNull(),
			"title":       types.StringNull(),
			"hide_border": types.BoolNull(),
		})
	}

	byRefObj := func() types.Object {
		return types.ObjectValueMust(byRefAttrs, map[string]attr.Value{
			"ref_id":      types.StringValue("lib-md-1"),
			"description": types.StringNull(),
			"hide_title":  types.BoolNull(),
			"title":       types.StringValue("from library"),
			"hide_border": types.BoolNull(),
		})
	}

	t.Run("accepts by_value only", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(mcAttrTypes, map[string]attr.Value{
			"by_value":     byValueObj(),
			"by_reference": types.ObjectNull(byRefAttrs),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("markdown_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("accepts by_reference only", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(mcAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectNull(byValueAttrs),
			"by_reference": byRefObj(),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("markdown_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("rejects both", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(mcAttrTypes, map[string]attr.Value{
			"by_value":     byValueObj(),
			"by_reference": byRefObj(),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("markdown_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "not both")
	})

	t.Run("rejects neither", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(mcAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectNull(byValueAttrs),
			"by_reference": types.ObjectNull(byRefAttrs),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("markdown_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("defers when by_value is unknown", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(mcAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectUnknown(byValueAttrs),
			"by_reference": types.ObjectNull(byRefAttrs),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("markdown_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("defers when by_reference is unknown", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(mcAttrTypes, map[string]attr.Value{
			"by_value":     types.ObjectNull(byValueAttrs),
			"by_reference": types.ObjectUnknown(byRefAttrs),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("markdown_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("no-op on null or unknown markdown_config object", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			name string
			val  attr.Value
		}{
			{"null", types.ObjectNull(mcAttrTypes)},
			{"unknown", types.ObjectUnknown(mcAttrTypes)},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ov, ok := tc.val.(types.Object)
				require.True(t, ok)
				var resp validator.ObjectResponse
				v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("markdown_config")}, &resp)
				require.False(t, resp.Diagnostics.HasError())
			})
		}
	})
}
