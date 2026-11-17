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

package validators_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestAtLeastOneOfNestedAttrsValidator(t *testing.T) {
	v := validators.AtLeastOneOfNestedAttrsValidator(validators.AtLeastOneOfNestedAttrsOpts{
		AttrNames: []string{"a", "b"},
		Summary:   "Empty block",
		Detail:    "At least one of `a` or `b` must be set.",
	})

	objectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"a": types.StringType,
			"b": types.StringType,
		},
	}

	build := func(a, b attr.Value) types.Object {
		o, diags := types.ObjectValue(objectType.AttrTypes, map[string]attr.Value{"a": a, "b": b})
		require.False(t, diags.HasError(), diags)
		return o
	}

	run := func(v validator.Object, configValue types.Object) (resp validator.ObjectResponse) {
		v.ValidateObject(context.Background(), validator.ObjectRequest{
			Path:        path.Root("block"),
			ConfigValue: configValue,
		}, &resp)
		return resp
	}

	t.Run("one set passes", func(t *testing.T) {
		resp := run(v, build(types.StringValue("x"), types.StringNull()))
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("both set passes", func(t *testing.T) {
		resp := run(v, build(types.StringValue("x"), types.StringValue("y")))
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("neither set fails", func(t *testing.T) {
		resp := run(v, build(types.StringNull(), types.StringNull()))
		require.True(t, resp.Diagnostics.HasError())
		require.Equal(t, "Empty block", resp.Diagnostics[0].Summary())
		require.Equal(t, "At least one of `a` or `b` must be set.", resp.Diagnostics[0].Detail())
	})

	t.Run("unknown child defers", func(t *testing.T) {
		resp := run(v, build(types.StringUnknown(), types.StringNull()))
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("unknown object defers", func(t *testing.T) {
		resp := run(v, types.ObjectUnknown(objectType.AttrTypes))
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("null object defers", func(t *testing.T) {
		resp := run(v, types.ObjectNull(objectType.AttrTypes))
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("description defaults when empty", func(t *testing.T) {
		require.Equal(t, "Ensures at least one of the listed nested attributes is set.", v.Description(context.Background()))
	})

	t.Run("custom description returned", func(t *testing.T) {
		v2 := validators.AtLeastOneOfNestedAttrsValidator(validators.AtLeastOneOfNestedAttrsOpts{
			AttrNames:   []string{"a", "b"},
			Description: "custom desc",
		})
		require.Equal(t, "custom desc", v2.Description(context.Background()))
		require.Equal(t, "custom desc", v2.MarkdownDescription(context.Background()))
	})
}
