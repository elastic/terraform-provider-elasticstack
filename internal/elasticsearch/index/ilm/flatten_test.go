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
	assert.Equal(t, `{"zone":"zone-1"}`, requireVal.ValueString())
}
