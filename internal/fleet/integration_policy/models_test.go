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

package integrationpolicy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestOutputIdHandling(t *testing.T) {
	t.Run("populateFromAPI", func(t *testing.T) {
		model := &integrationPolicyModel{}
		outputID := "test-output-id"
		testID := "test-id"
		data := &kbapi.PackagePolicy{
			Id:      &testID,
			Name:    "test-policy",
			Enabled: true,
			Package: &kbapi.KibanaHTTPAPIsPackagePolicyPackage{
				Name:    "test-integration",
				Version: "1.0.0",
			},
			OutputId: &outputID,
		}

		diags := model.populateFromAPI(context.Background(), nil, data)
		require.Empty(t, diags)
		require.Equal(t, "test-output-id", model.OutputID.ValueString())
	})

	t.Run("toAPIModel", func(t *testing.T) {
		model := integrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			OutputID:           types.StringValue("test-output-id"),
			AgentPolicyID:      types.StringValue("test-policy-id"),
			AgentPolicyIDs:     types.ListNull(types.StringType),
		}

		feat := features{
			SupportsPolicyIDs: true,
			SupportsOutputID:  true,
		}

		body, diags := model.toAPIModel(context.Background(), feat)
		require.Empty(t, diags)

		raw, err := body.MarshalJSON()
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(raw, &decoded))
		require.Equal(t, "test-output-id", decoded["output_id"])
		require.Equal(t, "test-policy-id", decoded["policy_id"])
		require.Equal(t, []any{}, decoded["policy_ids"])
	})

	t.Run("toAPIModel_unsupported_version", func(t *testing.T) {
		model := integrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			OutputID:           types.StringValue("test-output-id"),
		}

		feat := features{
			SupportsPolicyIDs: true,
			SupportsOutputID:  false, // Simulate unsupported version
		}

		_, diags := model.toAPIModel(context.Background(), feat)
		require.Len(t, diags, 1)
		require.Equal(t, "Unsupported Elasticsearch version", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), "Output ID is only supported in Elastic Stack")
	})
}
