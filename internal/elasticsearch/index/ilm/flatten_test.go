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
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func TestFlattenPhaseRollover(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	emptyRolloverActions := map[string]map[string]any{
		ilmActionRollover:    {},
		ilmActionSetPriority: {"priority": int64(10)},
	}

	t.Run("omits empty rollover when prior rollover is null", func(t *testing.T) {
		t.Parallel()

		priors := []struct {
			name  string
			prior types.Object
		}{
			{name: "null phase", prior: types.ObjectNull(hotPhaseObjectType().AttrTypes)},
			{name: "phase with null rollover attr", prior: hotPhaseObjectWithRollover(t, types.ObjectNull(rolloverObjectType().AttrTypes))},
		}

		for _, p := range priors {
			t.Run(p.name, func(t *testing.T) {
				t.Parallel()

				obj, diags := flattenPhase(ctx, ilmPhaseHot, "1h", emptyRolloverActions, p.prior)
				require.False(t, diags.HasError(), "%s", diags)

				rollover, ok := obj.Attributes()[ilmActionRollover].(types.Object)
				require.True(t, ok)
				assert.True(t, rollover.IsNull())

				priority, ok := obj.Attributes()[ilmActionSetPriority].(types.Object)
				require.True(t, ok)
				assert.False(t, priority.IsNull())
			})
		}
	})

	t.Run("preserves empty rollover when prior declared a non-null object", func(t *testing.T) {
		t.Parallel()

		prior := hotPhaseObjectWithRollover(t, emptyRolloverObject(t))
		obj, diags := flattenPhase(ctx, ilmPhaseHot, "1h", emptyRolloverActions, prior)
		require.False(t, diags.HasError(), "%s", diags)

		rollover, ok := obj.Attributes()[ilmActionRollover].(types.Object)
		require.True(t, ok)
		assert.False(t, rollover.IsNull())

		maxAge, ok := rollover.Attributes()[attrMaxAge].(types.String)
		require.True(t, ok)
		assert.True(t, maxAge.IsNull())
	})

	t.Run("populated rollover conditions flatten unchanged", func(t *testing.T) {
		t.Parallel()

		obj, diags := flattenPhase(ctx, ilmPhaseHot, "1h", map[string]map[string]any{
			ilmActionRollover: {
				attrMaxAge:              "7d",
				"max_docs":              int64(10000),
				attrMaxPrimaryShardSize: "50gb",
			},
		}, types.ObjectNull(hotPhaseObjectType().AttrTypes))
		require.False(t, diags.HasError(), "%s", diags)

		rollover, ok := obj.Attributes()[ilmActionRollover].(types.Object)
		require.True(t, ok)
		assert.False(t, rollover.IsNull())

		maxAge, ok := rollover.Attributes()[attrMaxAge].(types.String)
		require.True(t, ok)
		assert.Equal(t, types.StringValue("7d"), maxAge)

		maxDocs, ok := rollover.Attributes()["max_docs"].(types.Int64)
		require.True(t, ok)
		assert.Equal(t, types.Int64Value(10000), maxDocs)

		maxPrimaryShardSize, ok := rollover.Attributes()[attrMaxPrimaryShardSize].(types.String)
		require.True(t, ok)
		assert.Equal(t, types.StringValue("50gb"), maxPrimaryShardSize)
	})
}

func hotPhaseObjectWithRollover(t *testing.T, rollover types.Object) types.Object {
	t.Helper()
	ot := hotPhaseObjectType()
	attrs := make(map[string]attr.Value, len(ot.AttrTypes))
	for k, typ := range ot.AttrTypes {
		if k == ilmActionRollover {
			attrs[k] = rollover
			continue
		}
		attrs[k] = nullValueForType(typ)
	}
	obj, diags := types.ObjectValue(ot.AttrTypes, attrs)
	require.False(t, diags.HasError(), "%s", diags)
	return obj
}

func emptyRolloverObject(t *testing.T) types.Object {
	t.Helper()
	ot := rolloverObjectType()
	attrs := make(map[string]attr.Value, len(ot.AttrTypes))
	for k, typ := range ot.AttrTypes {
		attrs[k] = nullValueForType(typ)
	}
	obj, diags := types.ObjectValue(ot.AttrTypes, attrs)
	require.False(t, diags.HasError(), "%s", diags)
	return obj
}
