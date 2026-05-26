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

package clients

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompositeIDFromStr(t *testing.T) {
	t.Run("two segments", func(t *testing.T) {
		compID, diags := CompositeIDFromStr("cluster-uuid/my-resource")
		require.False(t, diags.HasError())
		assert.Equal(t, "cluster-uuid", compID.ClusterID)
		assert.Equal(t, "my-resource", compID.ResourceID)
	})

	t.Run("resource segment may contain slashes", func(t *testing.T) {
		compID, diags := CompositeIDFromStr("cluster-uuid/cal-1/evt-2")
		require.False(t, diags.HasError())
		assert.Equal(t, "cluster-uuid", compID.ClusterID)
		assert.Equal(t, "cal-1/evt-2", compID.ResourceID)
	})

	t.Run("legacy empty cluster segment", func(t *testing.T) {
		compID, diags := CompositeIDFromStr("/legacy-resource")
		require.False(t, diags.HasError())
		assert.Empty(t, compID.ClusterID)
		assert.Equal(t, "legacy-resource", compID.ResourceID)
	})

	t.Run("rejects empty resource segment", func(t *testing.T) {
		_, diags := CompositeIDFromStr("cluster-uuid/")
		require.True(t, diags.HasError())
	})

	t.Run("rejects missing slash", func(t *testing.T) {
		_, diags := CompositeIDFromStr("not-a-composite-import-id")
		require.True(t, diags.HasError())
	})
}
