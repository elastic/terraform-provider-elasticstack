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

package agentlesspolicy

import (
	"context"
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/providerfwtest"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestResource_embedsKibanaResource(t *testing.T) {
	t.Parallel()
	rt := reflect.TypeFor[Resource]()
	field, ok := rt.FieldByName("KibanaResource")
	require.True(t, ok)
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.KibanaResource[agentlessPolicyModel]](), field.Type)
}

func TestResource_importState_customCompositeID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "myspace/my-policy-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var policyID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("policy_id"), &policyID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "my-policy-id", policyID.ValueString())

	var spaceIDs types.Set
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_ids"), &spaceIDs)...)
	require.False(t, resp.Diagnostics.HasError())

	var elems []types.String
	resp.Diagnostics.Append(spaceIDs.ElementsAs(ctx, &elems, false)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Len(t, elems, 1)
	require.Equal(t, "myspace", elems[0].ValueString())
}

func TestAgentlessPolicyModel_getVersionRequirements(t *testing.T) {
	t.Parallel()

	m := agentlessPolicyModel{}
	reqs, diags := m.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.True(t, reqs[0].MinVersion.Equal(version.Must(version.NewVersion("9.3.0"))))
}

func TestAgentlessPolicyModel_getSpaceID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("null space_ids defaults to \"default\"", func(t *testing.T) {
		t.Parallel()
		// space_ids has no schema-level Default plan modifier (matching
		// internal/fleet/output and internal/fleet/serverhost), so Create
		// must fall back to "default" here or every Create without an
		// explicit space_ids would fail entitycore's validateSpaceID check
		// (this resource is space-scoped, unlike output/serverhost, which
		// opt out via KibanaUnscopedSpace). See the GetSpaceID doc comment.
		m := agentlessPolicyModel{SpaceIDs: types.SetNull(types.StringType)}
		require.Equal(t, "default", m.GetSpaceID().ValueString())
	})

	t.Run("unknown space_ids defaults to \"default\"", func(t *testing.T) {
		t.Parallel()
		m := agentlessPolicyModel{SpaceIDs: types.SetUnknown(types.StringType)}
		require.Equal(t, "default", m.GetSpaceID().ValueString())
	})

	t.Run("returns first non-empty element", func(t *testing.T) {
		t.Parallel()
		set, diags := types.SetValueFrom(ctx, types.StringType, []string{"myspace"})
		require.False(t, diags.HasError())
		m := agentlessPolicyModel{SpaceIDs: set}
		require.Equal(t, "myspace", m.GetSpaceID().ValueString())
	})
}
