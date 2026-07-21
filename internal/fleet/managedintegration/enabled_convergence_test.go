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

// nonPersistingEnabledUpdateResponseJSON is the mapped-format
// GET/PUT-response shape (see mappedFormatPackagePolicyJSON) that
// TestEnabledChangeNonPersistence_stateReflectsAPIReality uses to simulate
// the Decision 3 spike finding documented in update.go's header comment and
// overlayInputFromPlan's own comment: Kibana accepted (200) a PUT that
// changed inputs["cspm-cloudbeat/cis_aws"].enabled from true to false, but
// silently did NOT persist it -- a subsequent read still reports `true`.
// This fixture stands in for that subsequent read: enabled is `true` (the
// OLD/unchanged value), not `false` (what the plan actually requested).
const nonPersistingEnabledUpdateResponseJSON = `{
	"id": "policy-1",
	"name": "test-policy",
	"namespace": "default",
	"enabled": true,
	"created_at": "2024-01-01T00:00:00.000Z",
	"created_by": "elastic",
	"updated_at": "2024-01-02T00:00:00.000Z",
	"updated_by": "elastic",
	"revision": 2,
	"policy_id": "policy-1",
	"policy_ids": ["policy-1"],
	"spaceIds": ["default"],
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
// This test round-trips through the real production call chain:
//  1. buildUpdateInputs/overlayInputFromPlan build the PUT request body from a
//     plan that asks to flip `cspm-cloudbeat/cis_aws`'s enabled from true
//     (current, per typedFormatPackagePolicyJSON) to false -- confirming the
//     request really did ask for the change (this resource is not silently
//     failing to send it).
//  2. populateFromPackagePolicy is then fed a simulated response
//     (nonPersistingEnabledUpdateResponseJSON) that echoes back enabled=true,
//     standing in for the entitycore envelope's real read-after-write refresh
//     (kibana_resource_envelope.go's runKibanaWrite calls Read with the
//     Update callback's returned model before persisting state).
//
// The correct, honest outcome asserted here is that the resulting model
// reflects the API's actual (unchanged) value -- NOT the plan's requested
// value. Terraform will show a diff on the next plan as a result; that is
// expected/correct behavior given the API's limitation, not a bug this test
// is trying to catch.
func TestEnabledChangeNonPersistence_stateReflectsAPIReality(t *testing.T) {
	ctx := context.Background()

	current := mustPackagePolicyFromJSON(t, typedFormatPackagePolicyJSON)

	// Plan requests enabled=false for cspm-cloudbeat/cis_aws (current/typed
	// fixture above has it at true). The stream/vars shape matches
	// typedFormatPackagePolicyJSON's cis_aws input so overlayInputFromPlan's
	// stream-matching path (DataStream.Dataset ==
	// "cloud_security_posture.findings") is also exercised normally, rather
	// than being incidentally skipped.
	plan := baseTestModel(t)
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

	// Step 1: confirm the request body really does ask Kibana for the
	// change -- this resource is not silently failing to even send it.
	decodedInputs := plan.decodeInputs(ctx, &diags)
	require.False(t, diags.HasError())
	reqInputs, inputDiags := buildUpdateInputs(current, decodedInputs)
	require.False(t, inputDiags.HasError(), "%v", inputDiags)
	require.Len(t, reqInputs, 2)
	var sawRequestedDisable bool
	for _, in := range reqInputs {
		if in.PolicyTemplate != nil && *in.PolicyTemplate == "cspm" && in.Type == "cloudbeat/cis_aws" {
			sawRequestedDisable = true
			require.False(t, in.Enabled, "the update request must actually ask Kibana to disable the input")
		}
	}
	require.True(t, sawRequestedDisable, "expected to find the cis_aws input in the built request")

	// Step 2: simulate Kibana accepting the PUT (200) but NOT persisting the
	// enabled change -- the subsequent read-after-write Read call sees the
	// input still enabled. plan.Inputs is left as the plan set it (mirroring
	// updateAgentlessPolicy's real sequence: the entitycore envelope calls
	// Read with the Update callback's returned model, which still carries
	// plan.Inputs at that point), so inputsKnownKeySet correctly limits
	// populateInputsModel's output to just the configured input.
	data := mustPackagePolicyFromJSON(t, nonPersistingEnabledUpdateResponseJSON)
	popDiags := plan.populateFromPackagePolicy(ctx, "default", data)
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
