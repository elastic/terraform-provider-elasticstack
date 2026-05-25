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

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

func TestConnectorReadExists(t *testing.T) {
	t.Parallel()

	t.Run("nil connector with no error means gone", func(t *testing.T) {
		t.Parallel()
		require.False(t, connectorReadExists(nil, nil))
	})

	t.Run("nil connector with read error is not gone", func(t *testing.T) {
		t.Parallel()
		var readDiags diag.Diagnostics
		readDiags.AddError("read failed", "boom")
		require.True(t, connectorReadExists(nil, readDiags))
	})

	t.Run("connector present", func(t *testing.T) {
		t.Parallel()
		require.True(t, connectorReadExists(&models.KibanaActionConnector{}, nil))
	})
}

func TestFinishConnectorRead(t *testing.T) {
	t.Parallel()

	t.Run("populate success", func(t *testing.T) {
		t.Parallel()
		model := tfModel{}
		apiModel := &models.KibanaActionConnector{
			ConnectorID:     "abc-123",
			SpaceID:         "default",
			Name:            "test",
			ConnectorTypeID: ".index",
			ConfigJSON:      `{"index":".kibana","refresh":true}`,
		}

		result, found, diags := finishConnectorRead(model, apiModel, "default", "abc-123")
		require.False(t, diags.HasError())
		require.True(t, found)
		require.Equal(t, "default/abc-123", result.ID.ValueString())
	})

	t.Run("populate error preserves found", func(t *testing.T) {
		t.Parallel()
		model := tfModel{}
		apiModel := &models.KibanaActionConnector{
			ConnectorID:     "abc-123",
			SpaceID:         "default",
			Name:            "test",
			ConnectorTypeID: ".index",
			ConfigJSON:      `{invalid json`,
		}

		_, found, diags := finishConnectorRead(model, apiModel, "default", "abc-123")
		require.True(t, diags.HasError())
		require.True(t, found, "populate errors must not remove resource from state")
	})
}
