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

// TestToCreateBody_conditionHandling asserts that input/stream `condition`
// values are included in the create request body. Version gating for
// `condition` is covered by the resource-level MinVersion 9.5.0 floor in
// models.go (same as policyshape.MinVersionCondition).
func TestToCreateBody_conditionHandling(t *testing.T) {
	t.Run("sends input and stream condition", func(t *testing.T) {
		t.Parallel()
		m := baseTestModel(t)
		m.Inputs = newInputsWithCondition(t, "host.os.family == 'linux'", "data_stream.dataset == 'audit'")

		body, diags := m.toCreateBody(context.Background())
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
}

// TestBuildUpdateBody_conditionHandling is the update-path counterpart of
// TestToCreateBody_conditionHandling.
func TestBuildUpdateBody_conditionHandling(t *testing.T) {
	current := mustPackagePolicyFromJSON(t, typedFormatPackagePolicyJSON)

	t.Run("sends input condition on update", func(t *testing.T) {
		t.Parallel()
		plan := baseTestModel(t)
		plan.Inputs = newInputsWithCondition(t, "host.os.family == 'linux'", "data_stream.dataset == 'audit'")

		body, diags := buildUpdateBody(context.Background(), plan, current)
		require.False(t, diags.HasError(), "%v", diags)

		decoded := decodeRequestJSON(t, body)
		inputs, ok := decoded["inputs"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, inputs)
		in, ok := inputs[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "host.os.family == 'linux'", in["condition"])
	})
}
