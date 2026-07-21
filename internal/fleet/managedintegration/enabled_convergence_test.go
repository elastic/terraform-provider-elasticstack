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

package managedintegration

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nonPersistingEnabledUpdateResponseJSON simulates a managed_integrations
// read-after-write response where enabled did not persist.
const nonPersistingEnabledUpdateResponseJSON = `{
	"id": "policy-1",
	"name": "test-policy",
	"namespace": "default",
	"created_at": "2024-01-01T00:00:00.000Z",
	"created_by": "elastic",
	"updated_at": "2024-01-02T00:00:00.000Z",
	"updated_by": "elastic",
	"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "Security Posture Management"},
	"vars": {"posture": "cspm", "deployment": "aws"},
	"description": "old description",
	"inputs": {
		"cspm-cloudbeat/cis_aws": {
			"enabled": true,
			"streams": {
				"cloud_security_posture.findings": {
					"enabled": true,
					"vars": {"aws.account_type": "single-account"}
				}
			}
		}
	}
}`

// TestEnabledChangeNonPersistence_stateReflectsAPIReality is a regression
// test for the Task 3 spike finding (update.go's header comment, and
// overlayInputFromPlan's comment on why no package-specific workaround was
// added): Kibana's PUT /api/fleet/package_policies/{id} can return 200 for a
// request that changes an input's `enabled` flag without actually persisting
// that change. Nothing previously proved this resource's Update path
// converges to the API's real state rather than looping (state believing the
// requested change succeeded, producing a permanent, incorrect non-diff, or
// worse, resurfacing a stale value on the next plan).
//
// This test round-trips through buildUpdateBody and populateFromManagedIntegration:
//  1. buildUpdateBody builds the full-replace PUT body from a plan that asks
//     to flip `cspm-cloudbeat/cis_aws`'s enabled to false.
//  2. populateFromManagedIntegration is then fed a simulated response
//     (nonPersistingEnabledUpdateResponseJSON) that echoes back enabled=true.
func TestEnabledChangeNonPersistence_stateReflectsAPIReality(t *testing.T) {
	ctx := context.Background()

	prior := baseTestModel(t)
	plan := prior

	streamsMap, diags := types.MapValueFrom(ctx, policyshape.StreamType(), map[string]policyshape.InputStreamModel{
		"cloud_security_posture.findings": {
			Enabled: types.BoolValue(true),
		},
	})
	require.False(t, diags.HasError())
	inputsValue, diags := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{
		"cspm-cloudbeat/cis_aws": {
			Enabled: types.BoolValue(false),
			Streams: streamsMap,
		},
	})
	require.False(t, diags.HasError())
	plan.Inputs = inputsValue

	body, bodyDiags := buildUpdateBody(ctx, plan, prior)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)
	decoded := decodeRequestJSON(t, body)
	inputs, ok := decoded["inputs"].(map[string]any)
	require.True(t, ok)
	in, ok := inputs["cspm-cloudbeat/cis_aws"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, false, in["enabled"], "the update request must ask Kibana to disable the input")

	data := mustManagedIntegrationFromJSON(t, nonPersistingEnabledUpdateResponseJSON)
	popDiags := plan.populateFromManagedIntegration(ctx, "default", data, nil)
	require.False(t, popDiags.HasError(), "%v", popDiags)

	var resultInputs map[string]agentlessInputModel
	require.False(t, plan.Inputs.ElementsAs(ctx, &resultInputs, false).HasError())
	require.Contains(t, resultInputs, "cspm-cloudbeat/cis_aws")

	// The state must reflect the API's actual (unchanged) value, not the
	// plan's requested value: this is the "we don't silently believe our own
	// request succeeded" safety property this test locks in.
	assert.True(t, resultInputs["cspm-cloudbeat/cis_aws"].Enabled.ValueBool(),
		"state must reflect the API's real (unchanged) enabled value, not the plan's requested value -- "+
			"a false here would mean the provider is trusting its own request instead of the API's response")
}
