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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

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
