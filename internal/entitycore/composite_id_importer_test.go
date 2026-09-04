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

package entitycore

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

// fakeCompositeIDResource is a minimal resource.Resource whose schema
// contains the attributes exercised by CompositeIDImporter tests: id,
// resource_id, extra_id.
type fakeCompositeIDResource struct {
	*CompositeIDImporter
}

func (f *fakeCompositeIDResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "test_fake_composite_id"
}

func (f *fakeCompositeIDResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Optional: true, Computed: true},
			"resource_id": schema.StringAttribute{Optional: true, Computed: true},
			"extra_id":    schema.StringAttribute{Optional: true, Computed: true},
		},
	}
}

func (f *fakeCompositeIDResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
}
func (f *fakeCompositeIDResource) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {
}
func (f *fakeCompositeIDResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (f *fakeCompositeIDResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

// TestCompositeIDImporter_compositeID verifies that a "<cluster>/<id>" import
// string sets idField to the raw import ID and resourceIDFields to the
// resource-ID portion.
func TestCompositeIDImporter_compositeID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeCompositeIDResource{
		CompositeIDImporter: NewCompositeIDImporter(path.Root("id"), path.Root("resource_id")),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "my-cluster/my-resource-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var id, resourceID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("resource_id"), &resourceID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "my-cluster/my-resource-id", id.ValueString())
	require.Equal(t, "my-resource-id", resourceID.ValueString())
}

// TestCompositeIDImporter_multipleResourceIDFields verifies that when
// multiple resourceIDFields are configured, all of them receive the
// resource-ID portion of the import string.
func TestCompositeIDImporter_multipleResourceIDFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeCompositeIDResource{
		CompositeIDImporter: NewCompositeIDImporter(path.Root("id"), path.Root("resource_id"), path.Root("extra_id")),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "my-cluster/shared-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var resourceID, extraID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("resource_id"), &resourceID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("extra_id"), &extraID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "shared-id", resourceID.ValueString())
	require.Equal(t, "shared-id", extraID.ValueString())
}

// TestCompositeIDImporter_invalidID verifies that an import ID without a
// "<cluster>/<resource_id>" shape results in an error diagnostic and no
// attributes are set.
func TestCompositeIDImporter_invalidID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeCompositeIDResource{
		CompositeIDImporter: NewCompositeIDImporter(path.Root("id"), path.Root("resource_id")),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "plain-id-without-cluster"}, resp)
	require.True(t, resp.Diagnostics.HasError())

	var resourceID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("resource_id"), &resourceID)...)
	require.True(t, resourceID.IsNull())
}

func TestNewCompositeIDImporter_panicsWithoutResourceIDFields(t *testing.T) {
	t.Parallel()
	defer func() {
		require.NotNil(t, recover())
	}()
	NewCompositeIDImporter(path.Root("id"))
}

// TestParseCompositeID_valid verifies that ParseCompositeID splits a valid
// composite ID without adding any diagnostics.
func TestParseCompositeID_valid(t *testing.T) {
	t.Parallel()

	resp := &resource.ImportStateResponse{}
	compID, ok := ParseCompositeID(resource.ImportStateRequest{ID: "my-cluster/my-resource-id"}, resp)
	require.True(t, ok)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "my-cluster", compID.ClusterID)
	require.Equal(t, "my-resource-id", compID.ResourceID)
}

// TestParseCompositeID_invalid verifies that ParseCompositeID appends an
// error diagnostic and returns ok=false for a malformed import ID.
func TestParseCompositeID_invalid(t *testing.T) {
	t.Parallel()

	resp := &resource.ImportStateResponse{}
	compID, ok := ParseCompositeID(resource.ImportStateRequest{ID: "plain-id-without-cluster"}, resp)
	require.False(t, ok)
	require.True(t, resp.Diagnostics.HasError())
	require.Nil(t, compID)
}
