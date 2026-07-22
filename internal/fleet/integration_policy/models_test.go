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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestOutputIdHandling(t *testing.T) {
	t.Run("populateFromAPI", func(t *testing.T) {
		model := &integrationPolicyModel{}
		outputID := "test-output-id"
		testID := "test-id"
		data := &kbapi.PackagePolicy{
			Id:      testID,
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

		feat := integrationPolicyFeatures{
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

		feat := integrationPolicyFeatures{
			SupportsPolicyIDs: true,
			SupportsOutputID:  false, // Simulate unsupported version
		}

		_, diags := model.toAPIModel(context.Background(), feat)
		require.Len(t, diags, 1)
		require.Equal(t, "Unsupported Elasticsearch version", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), "Output ID is only supported in Elastic Stack")
	})
}

// TestConditionHandling verifies the additive `condition` attribute (Phase 1,
// openspec/changes/archive/2026-07-02-fleet-agentless-policy) round-trips end to end for both
// inputs and streams: sent on the request when set, omitted when unset, and
// read back from the API response into state.
func TestConditionHandling(t *testing.T) {
	ctx := context.Background()

	newInputsWithCondition := func(t *testing.T, inputCondition, streamCondition string) InputsValue {
		t.Helper()

		streamModel := integrationPolicyInputStreamModel{
			Enabled: types.BoolValue(true),
		}
		if streamCondition != "" {
			streamModel.Condition = types.StringValue(streamCondition)
		}

		streams, diags := types.MapValueFrom(ctx, getInputStreamType(), map[string]integrationPolicyInputStreamModel{
			"test.stream": streamModel,
		})
		require.False(t, diags.HasError())

		inputModel := integrationPolicyInputsModel{
			Enabled:  types.BoolValue(true),
			Streams:  streams,
			Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
		}
		if inputCondition != "" {
			inputModel.Condition = types.StringValue(inputCondition)
		}

		inputs, diags := NewInputsValueFrom(ctx, getInputsElementType(), map[string]integrationPolicyInputsModel{
			"test-input": inputModel,
		})
		require.False(t, diags.HasError())
		return inputs
	}

	t.Run("toAPIModel sends condition when set", func(t *testing.T) {
		model := integrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			AgentPolicyIDs:     types.ListNull(types.StringType),
			Inputs:             newInputsWithCondition(t, "host.os.family == 'linux'", "data_stream.dataset == 'audit'"),
		}

		body, diags := model.toAPIModel(ctx, integrationPolicyFeatures{SupportsPolicyIDs: true, SupportsOutputID: true, SupportsCondition: true})
		require.False(t, diags.HasError())

		raw, err := body.MarshalJSON()
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(raw, &decoded))

		inputs := decoded["inputs"].(map[string]any)
		input := inputs["test-input"].(map[string]any)
		require.Equal(t, "host.os.family == 'linux'", input["condition"])

		streams := input["streams"].(map[string]any)
		stream := streams["test.stream"].(map[string]any)
		require.Equal(t, "data_stream.dataset == 'audit'", stream["condition"])
	})

	t.Run("toAPIModel omits condition when unset", func(t *testing.T) {
		model := integrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			AgentPolicyIDs:     types.ListNull(types.StringType),
			Inputs:             newInputsWithCondition(t, "", ""),
		}

		body, diags := model.toAPIModel(ctx, integrationPolicyFeatures{SupportsPolicyIDs: true, SupportsOutputID: true, SupportsCondition: true})
		require.False(t, diags.HasError())

		raw, err := body.MarshalJSON()
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(raw, &decoded))

		inputs := decoded["inputs"].(map[string]any)
		input := inputs["test-input"].(map[string]any)
		_, hasInputCondition := input["condition"]
		require.False(t, hasInputCondition, "condition should be omitted from the request body when unset")

		streams := input["streams"].(map[string]any)
		stream := streams["test.stream"].(map[string]any)
		_, hasStreamCondition := stream["condition"]
		require.False(t, hasStreamCondition, "stream condition should be omitted from the request body when unset")
	})

	// The condition field is gated behind MinVersionCondition (Kibana 9.5.0):
	// it is rejected by Kibana 9.4.x and earlier ("Additional properties are
	// not allowed" 400), confirmed empirically against a 9.5.0-SNAPSHOT
	// Kibana. See design.md Open Question 4 resolution and MinVersionCondition.
	t.Run("toAPIModel rejects input condition when version unsupported", func(t *testing.T) {
		model := integrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			AgentPolicyIDs:     types.ListNull(types.StringType),
			Inputs:             newInputsWithCondition(t, "host.os.family == 'linux'", ""),
		}

		_, diags := model.toAPIModel(ctx, integrationPolicyFeatures{SupportsPolicyIDs: true, SupportsOutputID: true, SupportsCondition: false})
		require.True(t, diags.HasError())
		require.Equal(t, "Unsupported Elasticsearch version", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), "Input condition is only supported in Elastic Stack")
		require.Contains(t, diags[0].Detail(), MinVersionCondition.String())
	})

	t.Run("toAPIModel rejects stream condition when version unsupported", func(t *testing.T) {
		model := integrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			AgentPolicyIDs:     types.ListNull(types.StringType),
			Inputs:             newInputsWithCondition(t, "", "data_stream.dataset == 'audit'"),
		}

		_, diags := model.toAPIModel(ctx, integrationPolicyFeatures{SupportsPolicyIDs: true, SupportsOutputID: true, SupportsCondition: false})
		require.True(t, diags.HasError())
		require.Equal(t, "Unsupported Elasticsearch version", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), "Stream condition is only supported in Elastic Stack")
		require.Contains(t, diags[0].Detail(), MinVersionCondition.String())
	})

	t.Run("toAPIModel allows unset condition when version unsupported", func(t *testing.T) {
		model := integrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			AgentPolicyIDs:     types.ListNull(types.StringType),
			Inputs:             newInputsWithCondition(t, "", ""),
		}

		_, diags := model.toAPIModel(ctx, integrationPolicyFeatures{SupportsPolicyIDs: true, SupportsOutputID: true, SupportsCondition: false})
		require.False(t, diags.HasError(), "condition gating must not affect requests that never set condition")
	})

	t.Run("populateFromAPI reads condition back into state", func(t *testing.T) {
		testID := "test-id"
		inputCondition := "host.os.family == 'linux'"
		streamCondition := "data_stream.dataset == 'audit'"

		respInputs := kbapi.PackagePolicyMappedInputs{
			"test-input": {
				Condition: &inputCondition,
				Enabled:   new(true),
				Streams: &map[string]kbapi.PackagePolicyMappedInputStream{
					"test.stream": {
						Condition: &streamCondition,
						Enabled:   new(true),
					},
				},
			},
		}

		data := &kbapi.PackagePolicy{
			Id:   testID,
			Name: "test-policy",
			Package: &kbapi.KibanaHTTPAPIsPackagePolicyPackage{
				Name:    "test-integration",
				Version: "1.0.0",
			},
		}
		require.NoError(t, data.Inputs.FromPackagePolicyMappedInputs(respInputs))

		model := &integrationPolicyModel{}
		diags := model.populateFromAPI(ctx, nil, data)
		require.False(t, diags.HasError())

		inputsMap := model.Inputs.Elements()
		require.Contains(t, inputsMap, "test-input")
		var inputModel integrationPolicyInputsModel
		d := inputsMap["test-input"].(InputValue).As(ctx, &inputModel, basetypes.ObjectAsOptions{})
		require.False(t, d.HasError())
		require.Equal(t, inputCondition, inputModel.Condition.ValueString())

		streamsMap := inputModel.Streams.Elements()
		require.Contains(t, streamsMap, "test.stream")
		var streamModel integrationPolicyInputStreamModel
		d = streamsMap["test.stream"].(types.Object).As(ctx, &streamModel, basetypes.ObjectAsOptions{})
		require.False(t, d.HasError())
		require.Equal(t, streamCondition, streamModel.Condition.ValueString())
	})

	t.Run("populateFromAPI leaves condition null when API omits it", func(t *testing.T) {
		testID := "test-id"

		respInputs := kbapi.PackagePolicyMappedInputs{
			"test-input": {
				Enabled: new(true),
			},
		}

		data := &kbapi.PackagePolicy{
			Id:   testID,
			Name: "test-policy",
			Package: &kbapi.KibanaHTTPAPIsPackagePolicyPackage{
				Name:    "test-integration",
				Version: "1.0.0",
			},
		}
		require.NoError(t, data.Inputs.FromPackagePolicyMappedInputs(respInputs))

		model := &integrationPolicyModel{}
		diags := model.populateFromAPI(ctx, nil, data)
		require.False(t, diags.HasError())

		inputsMap := model.Inputs.Elements()
		require.Contains(t, inputsMap, "test-input")
		var inputModel integrationPolicyInputsModel
		d := inputsMap["test-input"].(InputValue).As(ctx, &inputModel, basetypes.ObjectAsOptions{})
		require.False(t, d.HasError())
		require.True(t, inputModel.Condition.IsNull(), "condition should be null in state when the API doesn't return it")
	})
}
