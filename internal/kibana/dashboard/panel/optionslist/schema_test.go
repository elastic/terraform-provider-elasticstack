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

package optionslist

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// byFieldObjType/byEsqlObjType are minimal attr.Type maps sufficient to exercise the
// ExactlyOneOfNestedAttrsValidator, which only inspects top-level attribute presence/nullness —
// it does not require the full branch attribute set.
var (
	byFieldTestAttrs = map[string]attr.Type{
		"data_view_id": types.StringType,
		"field_name":   types.StringType,
	}
	byEsqlTestAttrs = map[string]attr.Type{
		"esql_query":    types.StringType,
		"values_source": types.StringType,
	}
	optionsListConfigTestAttrTypes = map[string]attr.Type{
		BranchByField: types.ObjectType{AttrTypes: byFieldTestAttrs},
		BranchByEsql:  types.ObjectType{AttrTypes: byEsqlTestAttrs},
	}
)

func optionsListExactlyOneOfValidator() validator.Object {
	return validators.ExactlyOneOfNestedAttrsValidator(validators.ExactlyOneOfNestedAttrsOpts{
		AttrNames:     []string{BranchByField, BranchByEsql},
		Summary:       "Invalid options_list_control_config",
		MissingDetail: "Exactly one of `by_field` or `by_esql` must be configured inside `options_list_control_config`.",
		TooManyDetail: "Exactly one of `by_field` or `by_esql` must be configured inside `options_list_control_config`, not both.",
	})
}

func Test_optionsListConfigModeValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := optionsListExactlyOneOfValidator()

	byFieldObj := types.ObjectValueMust(byFieldTestAttrs, map[string]attr.Value{
		"data_view_id": types.StringValue("dv1"),
		"field_name":   types.StringValue("host.name"),
	})
	byEsqlObj := types.ObjectValueMust(byEsqlTestAttrs, map[string]attr.Value{
		"esql_query":    types.StringValue("FROM logs"),
		"values_source": types.StringValue("esql_query"),
	})

	t.Run("accepts by_field only", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(optionsListConfigTestAttrTypes, map[string]attr.Value{
			BranchByField: byFieldObj,
			BranchByEsql:  types.ObjectNull(byEsqlTestAttrs),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("options_list_control_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("accepts by_esql only", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(optionsListConfigTestAttrTypes, map[string]attr.Value{
			BranchByField: types.ObjectNull(byFieldTestAttrs),
			BranchByEsql:  byEsqlObj,
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("options_list_control_config")}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("rejects both", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(optionsListConfigTestAttrTypes, map[string]attr.Value{
			BranchByField: byFieldObj,
			BranchByEsql:  byEsqlObj,
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("options_list_control_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "not both")
	})

	t.Run("rejects neither", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(optionsListConfigTestAttrTypes, map[string]attr.Value{
			BranchByField: types.ObjectNull(byFieldTestAttrs),
			BranchByEsql:  types.ObjectNull(byEsqlTestAttrs),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{ConfigValue: ov, Path: path.Root("options_list_control_config")}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "Exactly one of")
	})
}

// Test_optionsListValuesSourceValidator exercises the same stringvalidator.OneOf("esql_query")
// instance wired onto by_esql.values_source in byEsqlAttributes.
func Test_optionsListValuesSourceValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := stringvalidator.OneOf("esql_query")

	t.Run("accepts esql_query", func(t *testing.T) {
		t.Parallel()
		var resp validator.StringResponse
		v.ValidateString(ctx, validator.StringRequest{
			ConfigValue: types.StringValue("esql_query"),
			Path:        path.Root("options_list_control_config").AtName("by_esql").AtName("values_source"),
		}, &resp)
		require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	})

	t.Run("rejects field", func(t *testing.T) {
		t.Parallel()
		var resp validator.StringResponse
		v.ValidateString(ctx, validator.StringRequest{
			ConfigValue: types.StringValue("field"),
			Path:        path.Root("options_list_control_config").AtName("by_esql").AtName("values_source"),
		}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})
}
