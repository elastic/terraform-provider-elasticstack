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

package parameter

import (
	"context"
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/providerfwtest"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResource_embedsEntityCoreKibanaResource(t *testing.T) {
	t.Parallel()
	rt := reflect.TypeFor[Resource]()
	field, ok := rt.FieldByName("KibanaResource")
	require.True(t, ok)
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.KibanaResource[Model]](), field.Type)
}

func TestResource_importState_compositeID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	const importID = "cluster/uuid-with-slash"
	r.ImportState(ctx, resource.ImportStateRequest{ID: importID}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var id, spaceID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_id"), &spaceID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, importID, id.ValueString())
	require.Equal(t, "cluster", spaceID.ValueString())
}

func TestResource_importState_bareUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	const paramUUID = "550e8400-e29b-41d4-a716-446655440000"
	r.ImportState(ctx, resource.ImportStateRequest{ID: paramUUID}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var id, spaceID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_id"), &spaceID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, clients.DefaultSpaceID+"/"+paramUUID, id.ValueString())
	require.Equal(t, clients.DefaultSpaceID, spaceID.ValueString())
}

func TestResource_importState_rejectsEmptyResourceSegment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "my-space/"}, resp)
	require.True(t, resp.Diagnostics.HasError())
	assert.Equal(t, "Wrong resource ID.", resp.Diagnostics.Errors()[0].Summary())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "<cluster_uuid>/<resource identifier>")
}

func TestResource_importState_rejectsMultipleSlashes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	const importID = "my-space/uuid/extra"
	r.ImportState(ctx, resource.ImportStateRequest{ID: importID}, resp)
	require.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), importID)
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "at most one slash")
}

func TestResource_importState_rejectsEmptyImportID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: ""}, resp)
	require.True(t, resp.Diagnostics.HasError())
	assert.Equal(t, "Wrong resource ID.", resp.Diagnostics.Errors()[0].Summary())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "bare `<parameter_uuid>`")
}
