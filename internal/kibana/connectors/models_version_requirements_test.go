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

func TestTfModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()

	const wantMessage = "Preconfigured connector IDs are only supported for Elastic Stack v8.8.0 and above. Either remove the `connector_id` attribute or upgrade your target cluster to supported version"

	t.Run("connector_id unset", func(t *testing.T) {
		t.Parallel()
		reqs, diags := (tfModel{}).GetVersionRequirements()
		require.False(t, diags.HasError())
		require.Nil(t, reqs)
	})

	t.Run("connector_id empty", func(t *testing.T) {
		t.Parallel()
		m := tfModel{ConnectorID: types.StringValue("")}
		reqs, diags := m.GetVersionRequirements()
		require.False(t, diags.HasError())
		require.Nil(t, reqs)
	})

	t.Run("connector_id set", func(t *testing.T) {
		t.Parallel()
		m := tfModel{ConnectorID: types.StringValue(".email")}
		reqs, diags := m.GetVersionRequirements()
		require.False(t, diags.HasError())
		require.Len(t, reqs, 1)
		require.Equal(t, *MinVersionSupportingPreconfiguredIDs, reqs[0].MinVersion)
		require.Equal(t, wantMessage, reqs[0].ErrorMessage)
	})
}
