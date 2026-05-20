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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
