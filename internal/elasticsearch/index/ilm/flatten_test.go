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

package ilm

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenPhaseAllocateOmitsAbsentReplicaShardFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	obj, diags := flattenPhase(ctx, ilmPhaseWarm, "", map[string]map[string]any{
		"allocate": {
			"require": map[string]any{"zone": "zone-1"},
		},
	}, types.ObjectNull(warmPhaseObjectType().AttrTypes))
	require.False(t, diags.HasError(), "%s", diags)

	allocateObj, ok := obj.Attributes()["allocate"].(types.Object)
	require.True(t, ok)

	allocateAttrs := allocateObj.Attributes()

	replicas, ok := allocateAttrs["number_of_replicas"].(types.Int64)
	require.True(t, ok)
	assert.True(t, replicas.IsNull())

	totalShards, ok := allocateAttrs["total_shards_per_node"].(types.Int64)
	require.True(t, ok)
	assert.True(t, totalShards.IsNull())
	assert.False(t, totalShards.Equal(types.Int64Value(-1)))

	requireVal, ok := allocateAttrs["require"].(jsontypes.Normalized)
	require.True(t, ok)
	assert.JSONEq(t, `{"zone":"zone-1"}`, requireVal.ValueString())
}

func TestFlattenPhaseSearchableSnapshotForceMergeOnClone(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := types.ObjectNull(hotPhaseObjectType().AttrTypes)

	t.Run("backfills true when omitted and force_merge_index is true", func(t *testing.T) {
		t.Parallel()

		obj, diags := flattenPhase(ctx, ilmPhaseHot, "", map[string]map[string]any{
			"searchable_snapshot": {
				"snapshot_repository": "repo-a",
				"force_merge_index":   true,
			},
		}, prior)
		require.False(t, diags.HasError(), "%s", diags)

		ss, ok := obj.Attributes()["searchable_snapshot"].(types.Object)
		require.True(t, ok)
		got, ok := ss.Attributes()["force_merge_on_clone"].(types.Bool)
		require.True(t, ok)
		assert.Equal(t, types.BoolValue(true), got)
	})

	t.Run("does not backfill when force_merge_index is false", func(t *testing.T) {
		t.Parallel()

		obj, diags := flattenPhase(ctx, ilmPhaseHot, "", map[string]map[string]any{
			"searchable_snapshot": {
				"snapshot_repository": "repo-a",
				"force_merge_index":   false,
			},
		}, prior)
		require.False(t, diags.HasError(), "%s", diags)

		ss, ok := obj.Attributes()["searchable_snapshot"].(types.Object)
		require.True(t, ok)
		got, ok := ss.Attributes()["force_merge_on_clone"].(types.Bool)
		require.True(t, ok)
		assert.True(t, got.IsNull())
	})

	t.Run("preserves explicit false", func(t *testing.T) {
		t.Parallel()

		obj, diags := flattenPhase(ctx, ilmPhaseHot, "", map[string]map[string]any{
			"searchable_snapshot": {
				"snapshot_repository":  "repo-a",
				"force_merge_index":    true,
				"force_merge_on_clone": false,
			},
		}, prior)
		require.False(t, diags.HasError(), "%s", diags)

		ss, ok := obj.Attributes()["searchable_snapshot"].(types.Object)
		require.True(t, ok)
		got, ok := ss.Attributes()["force_merge_on_clone"].(types.Bool)
		require.True(t, ok)
		assert.Equal(t, types.BoolValue(false), got)
	})
}
