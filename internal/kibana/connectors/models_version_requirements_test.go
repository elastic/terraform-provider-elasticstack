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
	"context"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestUserSuppliedConnectorID(t *testing.T) {
	t.Parallel()

	t.Run("connector_id unset", func(t *testing.T) {
		t.Parallel()
		require.False(t, userSuppliedConnectorID(tfModel{}))
	})

	t.Run("connector_id empty", func(t *testing.T) {
		t.Parallel()
		require.False(t, userSuppliedConnectorID(tfModel{ConnectorID: types.StringValue("")}))
	})

	t.Run("connector_id set", func(t *testing.T) {
		t.Parallel()
		require.True(t, userSuppliedConnectorID(tfModel{ConnectorID: types.StringValue(".email")}))
	})

	t.Run("api assigned connector_id in state does not enable gate alone", func(t *testing.T) {
		t.Parallel()
		// After create, state carries API UUID; gate predicate is the same shape but
		// createConnector is the only caller of enforceUserSuppliedConnectorIDVersion.
		require.True(t, userSuppliedConnectorID(tfModel{ConnectorID: types.StringValue("550e8400-e29b-41d4-a716-446655440000")}))
	})
}

type mockMinVersionClient struct {
	called     bool
	minVersion *version.Version
	supported  bool
	diags      diag.Diagnostics
}

func (m *mockMinVersionClient) EnforceMinVersion(_ context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	m.called = true
	m.minVersion = minVersion
	return m.supported, m.diags
}

func TestEnforceUserSuppliedConnectorIDVersion(t *testing.T) {
	t.Parallel()

	const wantMessage = "Preconfigured connector IDs are only supported for Elastic Stack v8.8.0 and above. Either remove the `connector_id` attribute or upgrade your target cluster to supported version"

	t.Run("skips when connector_id unset", func(t *testing.T) {
		t.Parallel()
		mock := &mockMinVersionClient{supported: true}
		diags := enforceUserSuppliedConnectorIDVersion(context.Background(), mock, tfModel{})
		require.False(t, diags.HasError())
		require.False(t, mock.called)
	})

	t.Run("checks version when connector_id set", func(t *testing.T) {
		t.Parallel()
		mock := &mockMinVersionClient{supported: true}
		plan := tfModel{ConnectorID: types.StringValue(".email")}
		diags := enforceUserSuppliedConnectorIDVersion(context.Background(), mock, plan)
		require.False(t, diags.HasError())
		require.True(t, mock.called)
		require.Equal(t, MinVersionSupportingPreconfiguredIDs, mock.minVersion)
	})

	t.Run("returns error when version unsupported", func(t *testing.T) {
		t.Parallel()
		mock := &mockMinVersionClient{supported: false}
		plan := tfModel{ConnectorID: types.StringValue(".email")}
		diags := enforceUserSuppliedConnectorIDVersion(context.Background(), mock, plan)
		require.True(t, diags.HasError())
		require.Equal(t, wantMessage, diags.Errors()[0].Detail())
	})
}
