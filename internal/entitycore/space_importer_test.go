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

// fakeSpaceListResource is a minimal resource.Resource whose schema contains
// the attributes exercised by SpaceImporter tests: resource_id, extra_id, space_ids.
type fakeSpaceListResource struct {
	*SpaceImporter
}

func (f *fakeSpaceListResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "test_fake"
}

func (f *fakeSpaceListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{Optional: true, Computed: true},
			"extra_id":    schema.StringAttribute{Optional: true, Computed: true},
			"space_ids":   schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		},
	}
}

func (f *fakeSpaceListResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
}
func (f *fakeSpaceListResource) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {}
func (f *fakeSpaceListResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (f *fakeSpaceListResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

// TestSpaceImporter_compositeID verifies that a "<space>/<id>" import string
// splits correctly: idField is set to the resource-ID portion and space_ids
// receives a one-element list with the space.
func TestSpaceImporter_compositeID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeSpaceListResource{SpaceImporter: NewSpaceImporter(path.Root("resource_id"))}
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

	r := &fakeSpaceListResource{SpaceImporter: NewSpaceImporter(path.Root("resource_id"))}
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

	r := &fakeSpaceListResource{
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

	r := &fakeSpaceListResource{
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

// fakeKibanaSpaceResource is a minimal resource.Resource whose schema contains
// the attributes exercised by KibanaSpaceImporter tests: id, space_id, rule_id.
type fakeKibanaSpaceResource struct {
	*KibanaSpaceImporter
}

func (f *fakeKibanaSpaceResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "test_fake_kibana"
}

func (f *fakeKibanaSpaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":       schema.StringAttribute{Optional: true, Computed: true},
			"space_id": schema.StringAttribute{Optional: true, Computed: true},
			"rule_id":  schema.StringAttribute{Optional: true, Computed: true},
			"extra_id": schema.StringAttribute{Optional: true, Computed: true},
		},
	}
}

func (f *fakeKibanaSpaceResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
}
func (f *fakeKibanaSpaceResource) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {
}
func (f *fakeKibanaSpaceResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (f *fakeKibanaSpaceResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

// TestKibanaSpaceImporter_compositeID verifies that a "<space>/<id>" import
// string is split into the id, space_id, and resource-ID fields.
func TestKibanaSpaceImporter_compositeID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeKibanaSpaceResource{
		KibanaSpaceImporter: NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id")),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "my-space/my-rule-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var id, spaceID, ruleID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_id"), &spaceID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("rule_id"), &ruleID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "my-space/my-rule-id", id.ValueString())
	require.Equal(t, "my-space", spaceID.ValueString())
	require.Equal(t, "my-rule-id", ruleID.ValueString())
}

// TestKibanaSpaceImporter_plainID verifies that a plain (non-composite) import
// ID results in an error diagnostic, since Kibana composite IDs are required.
func TestKibanaSpaceImporter_plainID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeKibanaSpaceResource{
		KibanaSpaceImporter: NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id")),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "plain-rule-id"}, resp)
	require.True(t, resp.Diagnostics.HasError())
}

// TestKibanaSpaceImporter_multipleResourceIDFields verifies that when multiple
// resourceIDFields are configured, all of them receive the resource-ID portion.
func TestKibanaSpaceImporter_multipleResourceIDFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeKibanaSpaceResource{
		KibanaSpaceImporter: NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id"), path.Root("extra_id")),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "my-space/shared-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var ruleID, extraID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("rule_id"), &ruleID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("extra_id"), &extraID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "shared-id", ruleID.ValueString())
	require.Equal(t, "shared-id", extraID.ValueString())
}

// TestKibanaSpaceImporter_emptySpaceSegment_defaultsToEmptyString verifies
// that, absent RequireSpaceID/DefaultSpaceID, a composite ID with an empty
// space segment (e.g. "/my-id") sets spaceIDField to an empty string rather
// than erroring.
func TestKibanaSpaceImporter_emptySpaceSegment_defaultsToEmptyString(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeKibanaSpaceResource{
		KibanaSpaceImporter: NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id")),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "/my-rule-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var spaceID, ruleID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_id"), &spaceID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("rule_id"), &ruleID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Empty(t, spaceID.ValueString())
	require.Equal(t, "my-rule-id", ruleID.ValueString())
}

// TestKibanaSpaceImporter_requireSpaceID_errorsOnEmptySpaceSegment verifies
// that RequireSpaceID turns an empty space segment into an error diagnostic
// using the configured summary/detail.
func TestKibanaSpaceImporter_requireSpaceID_errorsOnEmptySpaceSegment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeKibanaSpaceResource{
		KibanaSpaceImporter: NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id")).
			RequireSpaceID("Wrong resource ID.", "Import ID must include a Kibana space in the form `<space_id>/<rule_id>`."),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "/my-rule-id"}, resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Equal(t, "Wrong resource ID.", resp.Diagnostics.Errors()[0].Summary())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "<space_id>/<rule_id>")
}

// TestKibanaSpaceImporter_defaultSpaceID_fallsBackOnEmptySpaceSegment verifies
// that DefaultSpaceID substitutes the configured space ID when the composite
// ID's space segment is empty, and rewrites id to the canonical form.
func TestKibanaSpaceImporter_defaultSpaceID_fallsBackOnEmptySpaceSegment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := &fakeKibanaSpaceResource{
		KibanaSpaceImporter: NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id")).
			DefaultSpaceID("default"),
	}
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "/my-rule-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var id, spaceID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_id"), &spaceID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "default", spaceID.ValueString())
	require.Equal(t, "default/my-rule-id", id.ValueString())
}

func TestKibanaSpaceImporter_requireAndDefaultSpaceID_mutuallyExclusive(t *testing.T) {
	t.Parallel()

	t.Run("default then require", func(t *testing.T) {
		t.Parallel()
		defer func() {
			require.NotNil(t, recover())
		}()
		NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id")).
			DefaultSpaceID("default").
			RequireSpaceID("summary", "detail")
	})

	t.Run("require then default", func(t *testing.T) {
		t.Parallel()
		defer func() {
			require.NotNil(t, recover())
		}()
		NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id")).
			RequireSpaceID("summary", "detail").
			DefaultSpaceID("default")
	})

	t.Run("empty default", func(t *testing.T) {
		t.Parallel()
		defer func() {
			require.NotNil(t, recover())
		}()
		NewKibanaSpaceImporter(path.Root("id"), path.Root("space_id"), path.Root("rule_id")).
			DefaultSpaceID("")
	})
}
