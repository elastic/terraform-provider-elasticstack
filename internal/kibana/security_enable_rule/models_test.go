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

package securityenablerule

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()

	const wantMessage = "Security detection rules bulk actions are not supported until Elastic Stack v8.11.0. Upgrade the target server to use this resource"

	t.Run("empty model", func(t *testing.T) {
		t.Parallel()
		var m enableRuleModel
		reqs, diags := m.GetVersionRequirements()
		require.False(t, diags.HasError())
		require.Len(t, reqs, 1)
		require.Equal(t, *minSupportedVersion, reqs[0].MinVersion)
		require.Equal(t, wantMessage, reqs[0].ErrorMessage)
	})

	t.Run("populated model", func(t *testing.T) {
		t.Parallel()
		m := enableRuleModel{
			ID:               types.StringValue("default/tag:production"),
			SpaceID:          types.StringValue("default"),
			Key:              types.StringValue("tag"),
			Value:            types.StringValue("production"),
			DisableOnDestroy: types.BoolValue(true),
			AllRulesEnabled:  types.BoolValue(true),
		}
		reqs, diags := m.GetVersionRequirements()
		require.False(t, diags.HasError())
		require.Len(t, reqs, 1)
		require.Equal(t, *minSupportedVersion, reqs[0].MinVersion)
		require.Equal(t, wantMessage, reqs[0].ErrorMessage)
	})
}
