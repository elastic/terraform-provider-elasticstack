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

func TestExactlyOneOfBranchValidator(t *testing.T) {
	v := panelkit.ExactlyOneOfBranchValidator("options_list_control_config", "by_field", "by_esql")

	objectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"by_field": types.StringType,
			"by_esql":  types.StringType,
		},
	}

	build := func(byField, byEsql attr.Value) types.Object {
		o, diags := types.ObjectValue(objectType.AttrTypes, map[string]attr.Value{"by_field": byField, "by_esql": byEsql})
		require.False(t, diags.HasError(), diags)
		return o
	}

	run := func(configValue types.Object) (resp validator.ObjectResponse) {
		v.ValidateObject(context.Background(), validator.ObjectRequest{
			Path:        path.Root("options_list_control_config"),
			ConfigValue: configValue,
		}, &resp)
		return resp
	}

	t.Run("exactly one set passes", func(t *testing.T) {
		resp := run(build(types.StringValue("x"), types.StringNull()))
		require.False(t, resp.Diagnostics.HasError())
	})

	t.Run("both set fails with too-many detail naming the config block", func(t *testing.T) {
		resp := run(build(types.StringValue("x"), types.StringValue("y")))
		require.True(t, resp.Diagnostics.HasError())
		require.Equal(t, "Invalid options_list_control_config", resp.Diagnostics[0].Summary())
		require.Contains(t, resp.Diagnostics[0].Detail(), "options_list_control_config")
		require.Contains(t, resp.Diagnostics[0].Detail(), "not both")
	})

	t.Run("neither set fails with missing detail naming the config block", func(t *testing.T) {
		resp := run(build(types.StringNull(), types.StringNull()))
		require.True(t, resp.Diagnostics.HasError())
		require.Equal(t, "Invalid options_list_control_config", resp.Diagnostics[0].Summary())
		require.Contains(t, resp.Diagnostics[0].Detail(), "options_list_control_config")
		require.NotContains(t, resp.Diagnostics[0].Detail(), "not both")
	})

	t.Run("distinct config names produce distinct messages", func(t *testing.T) {
		other := panelkit.ExactlyOneOfBranchValidator("range_slider_control_config", "by_field", "by_esql")
		var resp validator.ObjectResponse
		other.ValidateObject(context.Background(), validator.ObjectRequest{
			Path:        path.Root("range_slider_control_config"),
			ConfigValue: build(types.StringNull(), types.StringNull()),
		}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Equal(t, "Invalid range_slider_control_config", resp.Diagnostics[0].Summary())
	})
}
