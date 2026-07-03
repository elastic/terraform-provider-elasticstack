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

package fieldstatstable

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestFieldStatsTableConfigModeValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	v := fieldStatsTableConfigModeValidator{}

	byDataviewAttrs := map[string]attr.Type{
		"data_view_id": types.StringType,
	}
	byEsqlAttrs := map[string]attr.Type{
		"query": types.StringType,
	}
	configAttrTypes := map[string]attr.Type{
		"by_dataview": types.ObjectType{AttrTypes: byDataviewAttrs},
		"by_esql":     types.ObjectType{AttrTypes: byEsqlAttrs},
	}

	t.Run("both branches set", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(configAttrTypes, map[string]attr.Value{
			"by_dataview": types.ObjectValueMust(byDataviewAttrs, map[string]attr.Value{
				"data_view_id": types.StringValue("dv"),
			}),
			"by_esql": types.ObjectValueMust(byEsqlAttrs, map[string]attr.Value{
				"query": types.StringValue("FROM logs"),
			}),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{
			Path:        path.Root("field_stats_table_config"),
			ConfigValue: ov,
		}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("neither branch set", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(configAttrTypes, map[string]attr.Value{
			"by_dataview": types.ObjectNull(byDataviewAttrs),
			"by_esql":     types.ObjectNull(byEsqlAttrs),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{
			Path:        path.Root("field_stats_table_config"),
			ConfigValue: ov,
		}, &resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("exactly one branch set", func(t *testing.T) {
		t.Parallel()
		ov := types.ObjectValueMust(configAttrTypes, map[string]attr.Value{
			"by_dataview": types.ObjectValueMust(byDataviewAttrs, map[string]attr.Value{
				"data_view_id": types.StringValue("dv"),
			}),
			"by_esql": types.ObjectNull(byEsqlAttrs),
		})
		var resp validator.ObjectResponse
		v.ValidateObject(ctx, validator.ObjectRequest{
			Path:        path.Root("field_stats_table_config"),
			ConfigValue: ov,
		}, &resp)
		require.False(t, resp.Diagnostics.HasError())
	})
}
