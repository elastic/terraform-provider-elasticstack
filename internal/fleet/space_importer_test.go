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

package fleet

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/providerfwtest"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// fakeResource is a minimal resource.Resource whose schema contains the
// attributes exercised by SpaceImporter tests: resource_id, extra_id, space_ids.
type fakeResource struct {
	*SpaceImporter
}

func (f *fakeResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "test_fake"
}

func (f *fakeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{Optional: true, Computed: true},
			"extra_id":    schema.StringAttribute{Optional: true, Computed: true},
			"space_ids":   schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		},
	}
}

func (f *fakeResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {}
func (f *fakeResource) Read(context.Context, resource.ReadRequest, *resource.ReadResponse)       {}
func (f *fakeResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {}
func (f *fakeResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {}

// TestSpaceImporter_compositeID verifies that a "<space>/<id>" import string
// splits correctly: idField is set to the resource-ID portion and space_ids
// receives a one-element list with the space.
func TestSpaceImporter_compositeID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeResource{SpaceImporter: NewSpaceImporter(path.Root("resource_id"))}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "my-space/my-resource-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var resourceID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("resource_id"), &resourceID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "my-resource-id", resourceID.ValueString())

	var spaceIDs types.Set
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_ids"), &spaceIDs)...)
	require.False(t, resp.Diagnostics.HasError())

	var elems []types.String
	resp.Diagnostics.Append(spaceIDs.ElementsAs(ctx, &elems, false)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Len(t, elems, 1)
	require.Equal(t, "my-space", elems[0].ValueString())
}

// TestSpaceImporter_plainID verifies that a plain (non-composite) import ID
// is placed into the idField as-is, and space_ids is left nil (not set).
func TestSpaceImporter_plainID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeResource{SpaceImporter: NewSpaceImporter(path.Root("resource_id"))}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "plain-resource-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var resourceID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("resource_id"), &resourceID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "plain-resource-id", resourceID.ValueString())

	var spaceIDs types.Set
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_ids"), &spaceIDs)...)
	require.False(t, resp.Diagnostics.HasError())
	require.True(t, spaceIDs.IsNull(), "space_ids should be nil for a plain import ID")
}

// TestSpaceImporter_multipleIDFields verifies that when multiple idFields are
// configured, all of them receive the resource-ID portion of the import string.
func TestSpaceImporter_multipleIDFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeResource{
		SpaceImporter: NewSpaceImporter(path.Root("resource_id"), path.Root("extra_id")),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "my-space/shared-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var resourceID, extraID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("resource_id"), &resourceID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("extra_id"), &extraID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "shared-id", resourceID.ValueString())
	require.Equal(t, "shared-id", extraID.ValueString())
}

// TestSpaceImporter_multipleIDFields_plainID verifies that when multiple
// idFields are configured and a plain (non-composite) ID is given, all fields
// receive the full import ID and space_ids is left null.
func TestSpaceImporter_multipleIDFields_plainID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeResource{
		SpaceImporter: NewSpaceImporter(path.Root("resource_id"), path.Root("extra_id")),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "plain-resource-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var resourceID, extraID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("resource_id"), &resourceID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("extra_id"), &extraID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "plain-resource-id", resourceID.ValueString())
	require.Equal(t, "plain-resource-id", extraID.ValueString())

	var spaceIDs types.Set
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_ids"), &spaceIDs)...)
	require.False(t, resp.Diagnostics.HasError())
	require.True(t, spaceIDs.IsNull(), "space_ids should be null for a plain import ID")
}
