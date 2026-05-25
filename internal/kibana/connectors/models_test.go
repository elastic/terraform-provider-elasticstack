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

package connectors

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestTfModel_GetResourceID(t *testing.T) {
	t.Parallel()

	t.Run("connector_id unset", func(t *testing.T) {
		t.Parallel()
		require.Empty(t, tfModel{}.GetResourceID().ValueString())
	})

	t.Run("connector_id empty", func(t *testing.T) {
		t.Parallel()
		m := tfModel{ConnectorID: types.StringValue("")}
		require.Empty(t, m.GetResourceID().ValueString())
	})

	t.Run("connector_id set", func(t *testing.T) {
		t.Parallel()
		m := tfModel{ConnectorID: types.StringValue("abc-123")}
		require.Equal(t, "abc-123", m.GetResourceID().ValueString())
	})
}

func TestTfModel_GetCompositeID(t *testing.T) {
	t.Parallel()

	t.Run("valid composite id", func(t *testing.T) {
		t.Parallel()
		m := tfModel{ID: types.StringValue("default/my-connector")}
		comp, diags := m.GetCompositeID()
		require.False(t, diags.HasError())
		require.Equal(t, "default", comp.ClusterID)
		require.Equal(t, "my-connector", comp.ResourceID)
	})

	t.Run("invalid composite id", func(t *testing.T) {
		t.Parallel()
		m := tfModel{ID: types.StringValue("not-a-composite")}
		_, diags := m.GetCompositeID()
		require.True(t, diags.HasError())
	})
}
