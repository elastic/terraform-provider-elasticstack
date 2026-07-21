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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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
// values are included in the create request body when set and omitted when
// unset. Version gating is covered by the resource-level MinVersion 9.5.0
// floor in models.go (same as policyshape.MinVersionCondition).
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

	t.Run("omits condition keys when unset", func(t *testing.T) {
		t.Parallel()
		m := baseTestModel(t)
		m.Inputs = newInputsWithCondition(t, "", "")

		body, diags := m.toCreateBody(context.Background())
		require.False(t, diags.HasError(), "%v", diags)

		decoded := decodeRequestJSON(t, body)
		inputs, ok := decoded["inputs"].(map[string]any)
		require.True(t, ok)
		in, ok := inputs["cspm-cloudbeat/cis_aws"].(map[string]any)
		require.True(t, ok)
		_, hasInputCondition := in["condition"]
		assert.False(t, hasInputCondition)

		streams, ok := in["streams"].(map[string]any)
		require.True(t, ok)
		stream, ok := streams["cloud_security_posture.findings"].(map[string]any)
		require.True(t, ok)
		_, hasStreamCondition := stream["condition"]
		assert.False(t, hasStreamCondition)
	})
}

// TestBuildUpdateBody_conditionHandling is the update-path counterpart of
// TestToCreateBody_conditionHandling.
func TestBuildUpdateBody_conditionHandling(t *testing.T) {
	t.Run("sends input and stream condition on update", func(t *testing.T) {
		t.Parallel()
		prior := baseTestModel(t)
		plan := prior
		plan.Inputs = newInputsWithCondition(t, "host.os.family == 'linux'", "data_stream.dataset == 'audit'")

		body, diags := buildUpdateBody(context.Background(), plan, prior)
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

	t.Run("omits condition keys when unset", func(t *testing.T) {
		t.Parallel()
		prior := baseTestModel(t)
		plan := prior
		plan.Inputs = newInputsWithCondition(t, "", "")

		body, diags := buildUpdateBody(context.Background(), plan, prior)
		require.False(t, diags.HasError(), "%v", diags)

		decoded := decodeRequestJSON(t, body)
		inputs, ok := decoded["inputs"].(map[string]any)
		require.True(t, ok)
		in, ok := inputs["cspm-cloudbeat/cis_aws"].(map[string]any)
		require.True(t, ok)
		_, hasInputCondition := in["condition"]
		assert.False(t, hasInputCondition)

		streams, ok := in["streams"].(map[string]any)
		require.True(t, ok)
		stream, ok := streams["cloud_security_posture.findings"].(map[string]any)
		require.True(t, ok)
		_, hasStreamCondition := stream["condition"]
		assert.False(t, hasStreamCondition)
	})
}

// TestPopulateFromCreateResponse_roundTripsCondition decodes `condition` from
// a KibanaHTTPAPIsManagedIntegration create/read response (mapped inputs map).
func TestPopulateFromCreateResponse_roundTripsCondition(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := kbapi.KibanaHTTPAPIsManagedIntegration{
		Id:        "policy-1",
		Name:      "test-policy",
		CreatedAt: "2024-01-01T00:00:00.000Z",
		CreatedBy: "elastic",
		UpdatedAt: "2024-01-02T00:00:00.000Z",
		UpdatedBy: "elastic",
		Package: kbapi.KibanaHTTPAPIsManagedIntegrationPackage{
			Name:    "cloud_security_posture",
			Version: "3.4.0",
			Title:   "Security Posture Management",
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{
		"cspm-cloudbeat/cis_aws": {
			"enabled": true,
			"condition": "host.os.family == 'linux'",
			"streams": {
				"cloud_security_posture.findings": {
					"enabled": true,
					"condition": "data_stream.dataset == 'audit'"
				}
			}
		}
	}`), &item.Inputs))

	m := baseTestModel(t)
	m.PolicyTemplate = types.StringValue("cspm")
	diags := m.populateFromManagedIntegration(ctx, "default", &item, nil)
	require.False(t, diags.HasError(), "%v", diags)

	var inputs map[string]agentlessInputModel
	require.False(t, m.Inputs.ElementsAs(ctx, &inputs, false).HasError())
	require.Contains(t, inputs, "cspm-cloudbeat/cis_aws")
	assert.Equal(t, "host.os.family == 'linux'", inputs["cspm-cloudbeat/cis_aws"].Condition.ValueString())

	var streams map[string]policyshape.InputStreamModel
	require.False(t, inputs["cspm-cloudbeat/cis_aws"].Streams.ElementsAs(ctx, &streams, false).HasError())
	require.Contains(t, streams, "cloud_security_posture.findings")
	assert.Equal(t, "data_stream.dataset == 'audit'", streams["cloud_security_posture.findings"].Condition.ValueString())
}

// TestPopulateFromCreateResponse_leavesConditionNullWhenAbsent decodes inputs
// from a managed-integration response that omits `condition` on inputs/streams.
func TestPopulateFromCreateResponse_leavesConditionNullWhenAbsent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := kbapi.KibanaHTTPAPIsManagedIntegration{
		Id:        "policy-1",
		Name:      "test-policy",
		CreatedAt: "2024-01-01T00:00:00.000Z",
		CreatedBy: "elastic",
		UpdatedAt: "2024-01-02T00:00:00.000Z",
		UpdatedBy: "elastic",
		Package: kbapi.KibanaHTTPAPIsManagedIntegrationPackage{
			Name:    "cloud_security_posture",
			Version: "3.4.0",
			Title:   "Security Posture Management",
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{
		"cspm-cloudbeat/cis_aws": {
			"enabled": true,
			"streams": {
				"cloud_security_posture.findings": {
					"enabled": true
				}
			}
		}
	}`), &item.Inputs))

	m := baseTestModel(t)
	m.PolicyTemplate = types.StringValue("cspm")
	diags := m.populateFromManagedIntegration(ctx, "default", &item, nil)
	require.False(t, diags.HasError(), "%v", diags)

	var inputs map[string]agentlessInputModel
	require.False(t, m.Inputs.ElementsAs(ctx, &inputs, false).HasError())
	require.Contains(t, inputs, "cspm-cloudbeat/cis_aws")
	assert.True(t, inputs["cspm-cloudbeat/cis_aws"].Condition.IsNull())

	var streams map[string]policyshape.InputStreamModel
	require.False(t, inputs["cspm-cloudbeat/cis_aws"].Streams.ElementsAs(ctx, &streams, false).HasError())
	require.Contains(t, streams, "cloud_security_posture.findings")
	assert.True(t, streams["cloud_security_posture.findings"].Condition.IsNull())
}

// TestPopulateFromManagedIntegration_leavesConditionNullWhenAbsent is the read-path
// counterpart using the managed_integrations response fixture.
func TestPopulateFromManagedIntegration_leavesConditionNullWhenAbsent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	data := mustManagedIntegrationFromJSON(t, mappedFormatManagedIntegrationJSON)
	m := agentlessPolicyModel{
		Force:                  types.BoolValue(true),
		CreateDatasetTemplates: types.BoolValue(true),
		PolicyTemplate:         types.StringValue("cspm"),
		CloudConnector:         types.ObjectNull(cloudConnectorAttrTypes()),
	}

	diags := m.populateFromManagedIntegration(ctx, "default", data, nil)
	require.False(t, diags.HasError(), "%v", diags)

	var inputs map[string]agentlessInputModel
	require.False(t, m.Inputs.ElementsAs(ctx, &inputs, false).HasError())
	require.Contains(t, inputs, "cspm-cloudbeat/cis_aws")
	assert.True(t, inputs["cspm-cloudbeat/cis_aws"].Condition.IsNull())

	var streams map[string]policyshape.InputStreamModel
	require.False(t, inputs["cspm-cloudbeat/cis_aws"].Streams.ElementsAs(ctx, &streams, false).HasError())
	require.Contains(t, streams, "cloud_security_posture.findings")
	assert.True(t, streams["cloud_security_posture.findings"].Condition.IsNull())
}

// TestPopulateFromManagedIntegration_roundTripsCondition is the read-path
// counterpart of TestPopulateFromCreateResponse_roundTripsCondition.
func TestPopulateFromManagedIntegration_roundTripsCondition(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "policy-1",
		"name": "test-policy",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-02T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "Security Posture Management"},
		"inputs": {
			"cspm-cloudbeat/cis_aws": {
				"enabled": true,
				"condition": "host.os.family == 'linux'",
				"streams": {
					"cloud_security_posture.findings": {
						"enabled": true,
						"condition": "data_stream.dataset == 'audit'"
					}
				}
			}
		}
	}`)

	m := baseTestModel(t)
	m.PolicyTemplate = types.StringValue("cspm")
	diags := m.populateFromManagedIntegration(ctx, "default", item, nil)
	require.False(t, diags.HasError(), "%v", diags)

	var inputs map[string]agentlessInputModel
	require.False(t, m.Inputs.ElementsAs(ctx, &inputs, false).HasError())
	require.Contains(t, inputs, "cspm-cloudbeat/cis_aws")
	assert.Equal(t, "host.os.family == 'linux'", inputs["cspm-cloudbeat/cis_aws"].Condition.ValueString())

	var streams map[string]policyshape.InputStreamModel
	require.False(t, inputs["cspm-cloudbeat/cis_aws"].Streams.ElementsAs(ctx, &streams, false).HasError())
	require.Contains(t, streams, "cloud_security_posture.findings")
	assert.Equal(t, "data_stream.dataset == 'audit'", streams["cloud_security_posture.findings"].Condition.ValueString())
}
