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

package agentlesspolicy

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newInputsWithCondition builds an `inputs` map with one input
// ("cspm-cloudbeat/cis_aws", matching typedFormatPackagePolicyJSON's
// PolicyTemplate/Type via mappedInputKey, so it also exercises
// TestBuildUpdateBody_conditionHandling's update path) carrying the given
// input-level and stream-level condition expressions (empty string = unset).
// Mirrors internal/fleet/integration_policy/models_test.go's
// TestConditionHandling helper of the same shape.
func newInputsWithCondition(t *testing.T, inputCondition, streamCondition string) policyshape.InputsValue {
	t.Helper()
	ctx := context.Background()

	streamModel := policyshape.InputStreamModel{
		Enabled: types.BoolValue(true),
	}
	if streamCondition != "" {
		streamModel.Condition = types.StringValue(streamCondition)
	}

	streams, diags := types.MapValueFrom(ctx, policyshape.StreamType(), map[string]policyshape.InputStreamModel{
		"cloud_security_posture.findings": streamModel,
	})
	require.False(t, diags.HasError())

	inputModel := agentlessInputModel{
		Enabled: types.BoolValue(true),
		Streams: streams,
	}
	if inputCondition != "" {
		inputModel.Condition = types.StringValue(inputCondition)
	}

	inputs, diags := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{
		"cspm-cloudbeat/cis_aws": inputModel,
	})
	require.False(t, diags.HasError())
	return inputs
}

// TestToCreateBody_conditionHandling covers the fleet-agentless-policy Fix 2
// gap: unlike internal/fleet/integration_policy, this resource previously
// sent `condition` on inputs/streams with no version gate at all, even though
// Kibana only accepts it from 9.5.0 onward (policyshape.MinVersionCondition)
// -- a user on this resource's own 9.3.0 floor would get a raw, unhelpful
// Kibana 400 ("Additional properties are not allowed") instead of a clean
// attribute-scoped Terraform diagnostic. Subtests mirror
// integration_policy/models_test.go's TestConditionHandling gating cases.
func TestToCreateBody_conditionHandling(t *testing.T) {
	// Deliberately not hoisted to a shared `ctx := context.Background()` at
	// this function's top level: baseTestModel/newInputsWithCondition below
	// each mint their own context.Background() internally, and golangci-lint's
	// contextcheck flags that pattern when an ancestor scope's ctx variable is
	// lexically reachable from the call site (as it would be from these
	// subtest closures) but isn't threaded through instead.
	t.Run("sends condition when supported", func(t *testing.T) {
		t.Parallel()
		m := baseTestModel(t)
		m.Inputs = newInputsWithCondition(t, "host.os.family == 'linux'", "data_stream.dataset == 'audit'")

		body, diags := m.toCreateBody(context.Background(), agentlessPolicyFeatures{SupportsCondition: true})
		require.False(t, diags.HasError(), "%v", diags)

		decoded := decodeRequestJSON(t, body)
		inputs, ok := decoded["inputs"].(map[string]any)
		require.True(t, ok)
		in, ok := inputs["cspm-cloudbeat/cis_aws"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "host.os.family == 'linux'", in["condition"])

		streams, ok := in["streams"].(map[string]any)
		require.True(t, ok)
		stream, ok := streams["cloud_security_posture.findings"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "data_stream.dataset == 'audit'", stream["condition"])
	})

	t.Run("rejects input condition when version unsupported", func(t *testing.T) {
		t.Parallel()
		m := baseTestModel(t)
		m.Inputs = newInputsWithCondition(t, "host.os.family == 'linux'", "")

		_, diags := m.toCreateBody(context.Background(), agentlessPolicyFeatures{SupportsCondition: false})
		require.True(t, diags.HasError())
		require.Equal(t, "Unsupported Elasticsearch version", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), "Input condition is only supported in Elastic Stack")
		require.Contains(t, diags[0].Detail(), policyshape.MinVersionCondition.String())
	})

	t.Run("rejects stream condition when version unsupported", func(t *testing.T) {
		t.Parallel()
		m := baseTestModel(t)
		m.Inputs = newInputsWithCondition(t, "", "data_stream.dataset == 'audit'")

		_, diags := m.toCreateBody(context.Background(), agentlessPolicyFeatures{SupportsCondition: false})
		require.True(t, diags.HasError())
		require.Equal(t, "Unsupported Elasticsearch version", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), "Stream condition is only supported in Elastic Stack")
		require.Contains(t, diags[0].Detail(), policyshape.MinVersionCondition.String())
	})

	t.Run("allows unset condition when version unsupported", func(t *testing.T) {
		t.Parallel()
		m := baseTestModel(t)
		m.Inputs = newInputsWithCondition(t, "", "")

		_, diags := m.toCreateBody(context.Background(), agentlessPolicyFeatures{SupportsCondition: false})
		require.False(t, diags.HasError(), "condition gating must not affect requests that never set condition: %v", diags)
	})
}

// TestBuildUpdateBody_conditionHandling is the update-path counterpart of
// TestToCreateBody_conditionHandling: buildUpdateBody must apply the exact
// same gating (via the shared validateInputConditionSupport), since
// overlayInputFromPlan sends the plan's `condition` value on every Update
// just as toCreateBody's applyCreateInputs does on Create.
func TestBuildUpdateBody_conditionHandling(t *testing.T) {
	// See TestToCreateBody_conditionHandling's comment on why no shared `ctx`
	// variable is hoisted to this function's top level.
	current := mustPackagePolicyFromJSON(t, typedFormatPackagePolicyJSON)

	t.Run("sends condition when supported", func(t *testing.T) {
		t.Parallel()
		plan := baseTestModel(t)
		plan.Inputs = newInputsWithCondition(t, "host.os.family == 'linux'", "data_stream.dataset == 'audit'")

		body, diags := buildUpdateBody(context.Background(), plan, current, agentlessPolicyFeatures{SupportsCondition: true})
		require.False(t, diags.HasError(), "%v", diags)

		decoded := decodeRequestJSON(t, body)
		inputs, ok := decoded["inputs"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, inputs)
		in, ok := inputs[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "host.os.family == 'linux'", in["condition"])
	})

	t.Run("rejects input condition when version unsupported", func(t *testing.T) {
		t.Parallel()
		plan := baseTestModel(t)
		plan.Inputs = newInputsWithCondition(t, "host.os.family == 'linux'", "")

		_, diags := buildUpdateBody(context.Background(), plan, current, agentlessPolicyFeatures{SupportsCondition: false})
		require.True(t, diags.HasError())
		require.Equal(t, "Unsupported Elasticsearch version", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), "Input condition is only supported in Elastic Stack")
		require.Contains(t, diags[0].Detail(), policyshape.MinVersionCondition.String())
	})

	t.Run("allows unset condition when version unsupported", func(t *testing.T) {
		t.Parallel()
		plan := baseTestModel(t)
		plan.Inputs = newInputsWithCondition(t, "", "")

		_, diags := buildUpdateBody(context.Background(), plan, current, agentlessPolicyFeatures{SupportsCondition: false})
		require.False(t, diags.HasError(), "condition gating must not affect requests that never set condition: %v", diags)
	})
}
