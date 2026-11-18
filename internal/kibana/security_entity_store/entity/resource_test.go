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

package entity

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/providerfwtest"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompositeIDForEntity(t *testing.T) {
	tests := []struct {
		name     string
		spaceID  string
		entityID string
		want     string
	}{
		{
			name:     "default space",
			spaceID:  clients.DefaultSpaceID,
			entityID: "host:web-01",
			want:     "default/host:web-01",
		},
		{
			name:     "custom space",
			spaceID:  "production",
			entityID: "host:web-01",
			want:     "production/host:web-01",
		},
		{
			name:     "entity ID with colons",
			spaceID:  "default",
			entityID: "user:john:doe",
			want:     "default/user:john:doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := (&clients.CompositeID{ClusterID: tt.spaceID, ResourceID: tt.entityID}).String()
			if got != tt.want {
				t.Errorf("(&CompositeID{%q, %q}).String() = %q, want %q", tt.spaceID, tt.entityID, got, tt.want)
			}
		})
	}
}

func TestResource_importState_seedsCompositeIdentity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)

	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	const importID = "production/host:web-01"
	r.ImportState(ctx, resource.ImportStateRequest{ID: importID}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var id, spaceID, entityID, entityType types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_id"), &spaceID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("entity_id"), &entityID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("entity_type"), &entityType)...)
	require.False(t, resp.Diagnostics.HasError())

	assert.Equal(t, importID, id.ValueString())
	assert.Equal(t, "production", spaceID.ValueString())
	assert.Equal(t, "host:web-01", entityID.ValueString())
	assert.Equal(t, "host", entityType.ValueString())
}

func TestResource_importState_defaultsEmptySpaceSegment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)

	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "/host:web-02"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var id, spaceID, entityID, entityType types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_id"), &spaceID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("entity_id"), &entityID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("entity_type"), &entityType)...)
	require.False(t, resp.Diagnostics.HasError())

	assert.Equal(t, clients.DefaultSpaceID+"/host:web-02", id.ValueString())
	assert.Equal(t, clients.DefaultSpaceID, spaceID.ValueString())
	assert.Equal(t, "host:web-02", entityID.ValueString())
	assert.Equal(t, "host", entityType.ValueString())
}

func TestResource_importState_rejectsMissingEntityTypePrefix(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)

	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "production/web-01"}, resp)
	require.True(t, resp.Diagnostics.HasError())
	assert.Equal(t, "Invalid import ID", resp.Diagnostics.Errors()[0].Summary())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "type prefix")

	var entityID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("entity_id"), &entityID)...)
	assert.True(t, entityID.IsNull() || entityID.ValueString() == "")
}

func TestResource_importState_rejectsMissingSlash(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)

	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "host-web-01"}, resp)
	require.True(t, resp.Diagnostics.HasError())
	assert.Equal(t, "Invalid import ID", resp.Diagnostics.Errors()[0].Summary())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "<space_id>/<entity_id>")
}
