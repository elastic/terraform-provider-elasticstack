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

package panelkit_test

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

func TestExactlyOneOfNestedAttrsValidator(t *testing.T) {
	v := panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
		AttrNames:     []string{"a", "b"},
		Summary:       "Invalid block",
		MissingDetail: "Exactly one of `a` or `b` must be set.",
		TooManyDetail: "Exactly one of `a` or `b` must be set, not both.",
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

	t.Run("exactly one set passes", func(t *testing.T) {
		resp := run(v, build(types.StringValue("x"), types.StringNull()))
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("both set fails with too-many detail", func(t *testing.T) {
		resp := run(v, build(types.StringValue("x"), types.StringValue("y")))
		require.True(t, resp.Diagnostics.HasError())
		require.Equal(t, "Invalid block", resp.Diagnostics[0].Summary())
		require.Contains(t, resp.Diagnostics[0].Detail(), "not both")
	})

	t.Run("neither set fails with missing detail", func(t *testing.T) {
		resp := run(v, build(types.StringNull(), types.StringNull()))
		require.True(t, resp.Diagnostics.HasError())
		require.Equal(t, "Invalid block", resp.Diagnostics[0].Summary())
		require.NotContains(t, resp.Diagnostics[0].Detail(), "not both")
	})

	t.Run("unknown defers", func(t *testing.T) {
		resp := run(v, build(types.StringUnknown(), types.StringNull()))
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("null config defers", func(t *testing.T) {
		resp := validator.ObjectResponse{}
		v.ValidateObject(context.Background(), validator.ObjectRequest{
			Path:        path.Root("block"),
			ConfigValue: types.ObjectNull(objectType.AttrTypes),
		}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})
}
