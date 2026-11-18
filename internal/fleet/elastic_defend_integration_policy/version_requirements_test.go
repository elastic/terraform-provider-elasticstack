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

package elasticdefendintegrationpolicy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestElasticDefendVersionRequirements(t *testing.T) {
	t.Parallel()

	model := elasticDefendIntegrationPolicyModel{
		AgentPolicyIDs: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("policy-id")}),
	}
	reqs, diags := model.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.Equal(t, MinVersionPolicyIDs.String(), reqs[0].MinVersion.String())
	require.Equal(t, "agent_policy_ids", reqs[0].AttributePath.String())
}

func TestElasticDefendVersionRequirementsAdditionalDatastreamsPermissions(t *testing.T) {
	t.Parallel()

	ids := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("policy-id")})
	perms := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("logs-nginx-*")})

	tests := map[string]struct {
		model    elasticDefendIntegrationPolicyModel
		expected []string
	}{
		"permissions only": {
			model:    elasticDefendIntegrationPolicyModel{AdditionalDatastreamsPermissions: perms, AgentPolicyIDs: types.ListNull(types.StringType)},
			expected: []string{"additional_datastreams_permissions"},
		},
		"permissions with unknown agent_policy_ids": {
			model:    elasticDefendIntegrationPolicyModel{AdditionalDatastreamsPermissions: perms, AgentPolicyIDs: types.ListUnknown(types.StringType)},
			expected: []string{"additional_datastreams_permissions"},
		},
		"empty permissions": {
			model:    elasticDefendIntegrationPolicyModel{AdditionalDatastreamsPermissions: types.ListValueMust(types.StringType, []attr.Value{}), AgentPolicyIDs: types.ListNull(types.StringType)},
			expected: []string{"additional_datastreams_permissions"},
		},
		"agent_policy_ids only": {
			model:    elasticDefendIntegrationPolicyModel{AdditionalDatastreamsPermissions: types.ListNull(types.StringType), AgentPolicyIDs: ids},
			expected: []string{"agent_policy_ids"},
		},
		"both": {
			model:    elasticDefendIntegrationPolicyModel{AdditionalDatastreamsPermissions: perms, AgentPolicyIDs: ids},
			expected: []string{"agent_policy_ids", "additional_datastreams_permissions"},
		},
		"neither": {
			model:    elasticDefendIntegrationPolicyModel{AdditionalDatastreamsPermissions: types.ListUnknown(types.StringType), AgentPolicyIDs: types.ListNull(types.StringType)},
			expected: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			reqs, diags := tc.model.GetVersionRequirements(context.Background())
			require.False(t, diags.HasError())
			var paths []string
			for _, r := range reqs {
				paths = append(paths, r.AttributePath.String())
				if r.AttributePath.String() == "additional_datastreams_permissions" {
					require.Equal(t, MinVersionAdditionalDatastreamsPermissions.String(), r.MinVersion.String())
				}
			}
			require.Equal(t, tc.expected, paths)
		})
	}
}
