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

package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestModifyPlanRequiresReplaceOnConnectionChange_differs(t *testing.T) {
	t.Parallel()

	plan := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("plan")})
	state := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("state")})

	var resp resource.ModifyPlanResponse
	ModifyPlanRequiresReplaceOnConnectionChange(plan, state, path.Root("elasticsearch_connection"), &resp)

	require.Len(t, resp.RequiresReplace, 1)
	require.Equal(t, path.Root("elasticsearch_connection"), resp.RequiresReplace[0])
}

func TestModifyPlanRequiresReplaceOnConnectionChange_equal(t *testing.T) {
	t.Parallel()

	conn := types.ListNull(types.StringType)

	var resp resource.ModifyPlanResponse
	ModifyPlanRequiresReplaceOnConnectionChange(conn, conn, path.Root("elasticsearch_connection"), &resp)

	require.Empty(t, resp.RequiresReplace)
}

func TestModifyPlanRequiresReplaceOnConnectionChange_unknownIgnored(t *testing.T) {
	t.Parallel()

	known := types.ListNull(types.StringType)
	unknown := types.ListUnknown(types.StringType)

	var resp resource.ModifyPlanResponse
	ModifyPlanRequiresReplaceOnConnectionChange(unknown, known, path.Root("elasticsearch_connection"), &resp)

	require.Empty(t, resp.RequiresReplace)
}

func TestModifyPlanRequiresReplaceOnConnectionChange_planNullStateConfigured(t *testing.T) {
	t.Parallel()

	plan := types.ListNull(types.StringType)
	state := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("configured")})

	var resp resource.ModifyPlanResponse
	ModifyPlanRequiresReplaceOnConnectionChange(plan, state, path.Root("elasticsearch_connection"), &resp)

	require.Len(t, resp.RequiresReplace, 1)
	require.Equal(t, path.Root("elasticsearch_connection"), resp.RequiresReplace[0])
}

func TestListRequiresReplaceIf_planNullStateConfigured(t *testing.T) {
	t.Parallel()

	elemType := types.ObjectType{AttrTypes: map[string]attr.Type{"endpoints": types.ListType{ElemType: types.StringType}}}
	plan := types.ListNull(elemType)
	state := types.ListValueMust(elemType, []attr.Value{
		types.ObjectValueMust(elemType.AttrTypes, map[string]attr.Value{
			"endpoints": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("https://localhost:9200")}),
		}),
	})

	var resp planmodifier.ListResponse
	ListRequiresReplaceIf("connection changed", func(_ context.Context, p, s types.List) bool {
		return !p.Equal(s)
	}).PlanModifyList(context.Background(), planmodifier.ListRequest{
		PlanValue:  plan,
		StateValue: state,
	}, &resp)

	require.True(t, resp.RequiresReplace)
}
