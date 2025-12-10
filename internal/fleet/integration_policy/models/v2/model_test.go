package v2

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy/models"
	v1 "github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy/models/v1"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputIdHandling(t *testing.T) {
	t.Run("populateFromAPI", func(t *testing.T) {
		model := &IntegrationPolicyModel{}
		outputId := "test-output-id"
		data := &kbapi.PackagePolicy{
			Id:      "test-id",
			Name:    "test-policy",
			Enabled: true,
			Package: &struct {
				ExperimentalDataStreamFeatures *[]struct {
					DataStream string `json:"data_stream"`
					Features   struct {
						DocValueOnlyNumeric *bool `json:"doc_value_only_numeric,omitempty"`
						DocValueOnlyOther   *bool `json:"doc_value_only_other,omitempty"`
						SyntheticSource     *bool `json:"synthetic_source,omitempty"`
						Tsdb                *bool `json:"tsdb,omitempty"`
					} `json:"features"`
				} `json:"experimental_data_stream_features,omitempty"`
				FipsCompatible *bool   `json:"fips_compatible,omitempty"`
				Name           string  `json:"name"`
				RequiresRoot   *bool   `json:"requires_root,omitempty"`
				Title          *string `json:"title,omitempty"`
				Version        string  `json:"version"`
			}{
				Name:    "test-integration",
				Version: "1.0.0",
			},
			OutputId: &outputId,
		}

		diags := model.PopulateFromAPI(context.Background(), data)
		require.Empty(t, diags)
		require.Equal(t, "test-output-id", model.OutputID.ValueString())
	})

	t.Run("toAPIModel", func(t *testing.T) {
		model := IntegrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			OutputID:           types.StringValue("test-output-id"),
		}

		feat := models.Features{
			SupportsPolicyIds: true,
			SupportsOutputId:  true,
		}

		result, diags := model.ToAPIModel(context.Background(), false, feat)
		require.Empty(t, diags)
		require.NotNil(t, result.OutputId)
		require.Equal(t, "test-output-id", *result.OutputId)
	})

	t.Run("toAPIModel_unsupported_version", func(t *testing.T) {
		model := IntegrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			OutputID:           types.StringValue("test-output-id"),
		}

		feat := models.Features{
			SupportsPolicyIds: true,
			SupportsOutputId:  false, // Simulate unsupported version
		}

		_, diags := model.ToAPIModel(context.Background(), false, feat)
		require.Len(t, diags, 1)
		require.Equal(t, "Unsupported Elasticsearch version", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), "Output ID is only supported in Elastic Stack")
	})
}

func TestUpdateStreamsV1ToV2(t *testing.T) {
	ctx := context.Background()

	t.Run("null streams", func(t *testing.T) {
		result, diags := updateStreamsV1ToV2(ctx, jsontypes.NewNormalizedNull(), "test-input")
		require.Empty(t, diags)
		assert.True(t, result.IsNull())
	})

	t.Run("unknown streams", func(t *testing.T) {
		result, diags := updateStreamsV1ToV2(ctx, jsontypes.NewNormalizedUnknown(), "test-input")
		require.Empty(t, diags)
		assert.True(t, result.IsNull())
	})

	t.Run("empty streams", func(t *testing.T) {
		emptyJSON := jsontypes.NewNormalizedValue("{}")
		result, diags := updateStreamsV1ToV2(ctx, emptyJSON, "test-input")
		require.Empty(t, diags)
		assert.True(t, result.IsNull())
	})

	t.Run("single stream with enabled and vars", func(t *testing.T) {
		enabled := true
		vars := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		apiStreams := map[string]kbapi.PackagePolicyInputStream{
			"stream-1": {
				Enabled: &enabled,
				Vars:    &vars,
			},
		}
		streamsJSON, err := json.Marshal(apiStreams)
		require.NoError(t, err)
		normalized := jsontypes.NewNormalizedValue(string(streamsJSON))

		result, diags := updateStreamsV1ToV2(ctx, normalized, "test-input")
		require.Empty(t, diags)
		assert.False(t, result.IsNull())

		// Convert back to verify
		var resultStreams map[string]integrationPolicyInputStreamModel
		d := result.ElementsAs(ctx, &resultStreams, false)
		require.Empty(t, d)
		require.Len(t, resultStreams, 1)
		require.Contains(t, resultStreams, "stream-1")

		stream := resultStreams["stream-1"]
		assert.Equal(t, types.BoolValue(true), stream.Enabled)
		assert.False(t, stream.Vars.IsNull())
	})

	t.Run("multiple streams with different configurations", func(t *testing.T) {
		enabled1 := true
		enabled2 := false
		vars1 := map[string]interface{}{"key1": "value1"}
		vars2 := map[string]interface{}{"key2": "value2"}

		apiStreams := map[string]kbapi.PackagePolicyInputStream{
			"stream-1": {
				Enabled: &enabled1,
				Vars:    &vars1,
			},
			"stream-2": {
				Enabled: &enabled2,
				Vars:    &vars2,
			},
		}
		streamsJSON, err := json.Marshal(apiStreams)
		require.NoError(t, err)
		normalized := jsontypes.NewNormalizedValue(string(streamsJSON))

		result, diags := updateStreamsV1ToV2(ctx, normalized, "test-input")
		require.Empty(t, diags)
		assert.False(t, result.IsNull())

		var resultStreams map[string]integrationPolicyInputStreamModel
		d := result.ElementsAs(ctx, &resultStreams, false)
		require.Empty(t, d)
		require.Len(t, resultStreams, 2)

		assert.Equal(t, types.BoolValue(true), resultStreams["stream-1"].Enabled)
		assert.Equal(t, types.BoolValue(false), resultStreams["stream-2"].Enabled)
	})

	t.Run("stream with nil enabled", func(t *testing.T) {
		vars := map[string]interface{}{"key": "value"}
		apiStreams := map[string]kbapi.PackagePolicyInputStream{
			"stream-1": {
				Enabled: nil,
				Vars:    &vars,
			},
		}
		streamsJSON, err := json.Marshal(apiStreams)
		require.NoError(t, err)
		normalized := jsontypes.NewNormalizedValue(string(streamsJSON))

		result, diags := updateStreamsV1ToV2(ctx, normalized, "test-input")
		require.Empty(t, diags)
		assert.False(t, result.IsNull())

		var resultStreams map[string]integrationPolicyInputStreamModel
		d := result.ElementsAs(ctx, &resultStreams, false)
		require.Empty(t, d)
		require.Len(t, resultStreams, 1)

		stream := resultStreams["stream-1"]
		assert.True(t, stream.Enabled.IsNull())
	})

	t.Run("stream with nil vars", func(t *testing.T) {
		enabled := true
		apiStreams := map[string]kbapi.PackagePolicyInputStream{
			"stream-1": {
				Enabled: &enabled,
				Vars:    nil,
			},
		}
		streamsJSON, err := json.Marshal(apiStreams)
		require.NoError(t, err)
		normalized := jsontypes.NewNormalizedValue(string(streamsJSON))

		result, diags := updateStreamsV1ToV2(ctx, normalized, "test-input")
		require.Empty(t, diags)
		assert.False(t, result.IsNull())

		var resultStreams map[string]integrationPolicyInputStreamModel
		d := result.ElementsAs(ctx, &resultStreams, false)
		require.Empty(t, d)
		require.Len(t, resultStreams, 1)

		stream := resultStreams["stream-1"]
		assert.True(t, stream.Vars.IsNull())
	})

	t.Run("invalid JSON", func(t *testing.T) {
		normalized := jsontypes.NewNormalizedValue("not valid json")
		result, diags := updateStreamsV1ToV2(ctx, normalized, "test-input")
		require.NotEmpty(t, diags)
		assert.True(t, result.IsNull())
	})
}

func TestIntegrationPolicyModelV1ToV2(t *testing.T) {
	ctx := context.Background()

	t.Run("basic model conversion", func(t *testing.T) {
		v1Model := v1.IntegrationPolicyModel{
			ID:                 types.StringValue("test-id"),
			PolicyID:           types.StringValue("test-policy-id"),
			Name:               types.StringValue("test-name"),
			Namespace:          types.StringValue("test-namespace"),
			AgentPolicyID:      types.StringValue("agent-policy-1"),
			Description:        types.StringValue("test description"),
			Enabled:            types.BoolValue(true),
			Force:              types.BoolValue(false),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			VarsJson:           jsontypes.NewNormalizedValue(`{"var1":"value1"}`),
			Input:              types.ListNull(v1.GetInputType()),
		}

		v2Model, diags := NewFromV1(t.Context(), v1Model)
		require.Empty(t, diags)

		assert.Equal(t, v1Model.ID, v2Model.ID)
		assert.Equal(t, v1Model.PolicyID, v2Model.PolicyID)
		assert.Equal(t, v1Model.Name, v2Model.Name)
		assert.Equal(t, v1Model.Namespace, v2Model.Namespace)
		assert.Equal(t, v1Model.AgentPolicyID, v2Model.AgentPolicyID)
		assert.Equal(t, v1Model.Description, v2Model.Description)
		assert.Equal(t, v1Model.Enabled, v2Model.Enabled)
		assert.Equal(t, v1Model.Force, v2Model.Force)
		assert.Equal(t, v1Model.IntegrationName, v2Model.IntegrationName)
		assert.Equal(t, v1Model.IntegrationVersion, v2Model.IntegrationVersion)
		assert.Equal(t, v1Model.VarsJson, v2Model.VarsJson)
	})

	t.Run("conversion with agent_policy_ids", func(t *testing.T) {
		policyIDs, diags := types.ListValueFrom(ctx, types.StringType, []string{"policy-1", "policy-2"})
		require.Empty(t, diags)

		v1Model := v1.IntegrationPolicyModel{
			ID:                 types.StringValue("test-id"),
			Name:               types.StringValue("test-name"),
			Namespace:          types.StringValue("test-namespace"),
			AgentPolicyIDs:     policyIDs,
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			Input:              types.ListNull(v1.GetInputType()),
		}

		v2Model, diags := NewFromV1(t.Context(), v1Model)
		require.Empty(t, diags)

		assert.Equal(t, v1Model.AgentPolicyIDs, v2Model.AgentPolicyIDs)
	})

	t.Run("conversion with space_ids", func(t *testing.T) {
		spaceIDs, diags := types.SetValueFrom(ctx, types.StringType, []string{"space-1", "space-2"})
		require.Empty(t, diags)

		v1Model := v1.IntegrationPolicyModel{
			ID:                 types.StringValue("test-id"),
			Name:               types.StringValue("test-name"),
			Namespace:          types.StringValue("test-namespace"),
			AgentPolicyID:      types.StringValue("agent-policy-1"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			SpaceIds:           spaceIDs,
			Input:              types.ListNull(v1.GetInputType()),
		}

		v2Model, diags := NewFromV1(t.Context(), v1Model)
		require.Empty(t, diags)

		assert.Equal(t, v1Model.SpaceIds, v2Model.SpaceIds)
	})

	t.Run("conversion with empty inputs", func(t *testing.T) {
		v1Model := v1.IntegrationPolicyModel{
			ID:                 types.StringValue("test-id"),
			Name:               types.StringValue("test-name"),
			Namespace:          types.StringValue("test-namespace"),
			AgentPolicyID:      types.StringValue("agent-policy-1"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			Input:              types.ListNull(v1.GetInputType()),
		}

		v2Model, diags := NewFromV1(t.Context(), v1Model)
		require.Empty(t, diags)

		assert.True(t, v2Model.Inputs.IsNull() || len(v2Model.Inputs.Elements()) == 0)
	})

	t.Run("conversion with single input without streams", func(t *testing.T) {
		inputsV1 := []v1.IntegrationPolicyInputModel{
			{
				InputID:     types.StringValue("input-1"),
				Enabled:     types.BoolValue(true),
				VarsJson:    jsontypes.NewNormalizedValue(`{"input_var":"value"}`),
				StreamsJson: jsontypes.NewNormalizedNull(),
			},
		}
		inputList, diags := types.ListValueFrom(ctx, v1.GetInputType(), inputsV1)
		require.Empty(t, diags)

		v1Model := v1.IntegrationPolicyModel{
			ID:                 types.StringValue("test-id"),
			Name:               types.StringValue("test-name"),
			Namespace:          types.StringValue("test-namespace"),
			AgentPolicyID:      types.StringValue("agent-policy-1"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			Input:              inputList,
		}

		v2Model, diags := NewFromV1(t.Context(), v1Model)
		require.Empty(t, diags)
		assert.False(t, v2Model.Inputs.IsNull())

		var inputsMap map[string]integrationPolicyInputsModel
		d := v2Model.Inputs.ElementsAs(ctx, &inputsMap, false)
		require.Empty(t, d)
		require.Len(t, inputsMap, 1)
		require.Contains(t, inputsMap, "input-1")

		input := inputsMap["input-1"]
		assert.Equal(t, types.BoolValue(true), input.Enabled)
		assert.False(t, input.Vars.IsNull())
		assert.True(t, input.Streams.IsNull())
	})

	t.Run("conversion with input and streams", func(t *testing.T) {
		enabled := true
		vars := map[string]interface{}{"stream_var": "value"}
		apiStreams := map[string]kbapi.PackagePolicyInputStream{
			"stream-1": {
				Enabled: &enabled,
				Vars:    &vars,
			},
		}
		streamsJSON, err := json.Marshal(apiStreams)
		require.NoError(t, err)

		inputsV1 := []v1.IntegrationPolicyInputModel{
			{
				InputID:     types.StringValue("input-1"),
				Enabled:     types.BoolValue(true),
				VarsJson:    jsontypes.NewNormalizedValue(`{"input_var":"value"}`),
				StreamsJson: jsontypes.NewNormalizedValue(string(streamsJSON)),
			},
		}
		inputList, diags := types.ListValueFrom(ctx, v1.GetInputType(), inputsV1)
		require.Empty(t, diags)

		v1Model := v1.IntegrationPolicyModel{
			ID:                 types.StringValue("test-id"),
			Name:               types.StringValue("test-name"),
			Namespace:          types.StringValue("test-namespace"),
			AgentPolicyID:      types.StringValue("agent-policy-1"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			Input:              inputList,
		}

		v2Model, diags := NewFromV1(t.Context(), v1Model)
		require.Empty(t, diags)
		assert.False(t, v2Model.Inputs.IsNull())

		var inputsMap map[string]integrationPolicyInputsModel
		d := v2Model.Inputs.ElementsAs(ctx, &inputsMap, false)
		require.Empty(t, d)
		require.Len(t, inputsMap, 1)

		input := inputsMap["input-1"]
		assert.Equal(t, types.BoolValue(true), input.Enabled)
		assert.False(t, input.Vars.IsNull())
		assert.False(t, input.Streams.IsNull())

		var streamsMap map[string]integrationPolicyInputStreamModel
		d = input.Streams.ElementsAs(ctx, &streamsMap, false)
		require.Empty(t, d)
		require.Len(t, streamsMap, 1)
		require.Contains(t, streamsMap, "stream-1")

		stream := streamsMap["stream-1"]
		assert.Equal(t, types.BoolValue(true), stream.Enabled)
		assert.False(t, stream.Vars.IsNull())
	})

	t.Run("conversion with multiple inputs and streams", func(t *testing.T) {
		enabled1 := true
		enabled2 := false
		vars1 := map[string]interface{}{"stream1_var": "value1"}
		vars2 := map[string]interface{}{"stream2_var": "value2"}

		apiStreams1 := map[string]kbapi.PackagePolicyInputStream{
			"stream-1": {Enabled: &enabled1, Vars: &vars1},
		}
		apiStreams2 := map[string]kbapi.PackagePolicyInputStream{
			"stream-2": {Enabled: &enabled2, Vars: &vars2},
		}

		streamsJSON1, err := json.Marshal(apiStreams1)
		require.NoError(t, err)
		streamsJSON2, err := json.Marshal(apiStreams2)
		require.NoError(t, err)

		inputsV1 := []v1.IntegrationPolicyInputModel{
			{
				InputID:     types.StringValue("input-1"),
				Enabled:     types.BoolValue(true),
				VarsJson:    jsontypes.NewNormalizedValue(`{"input1_var":"value1"}`),
				StreamsJson: jsontypes.NewNormalizedValue(string(streamsJSON1)),
			},
			{
				InputID:     types.StringValue("input-2"),
				Enabled:     types.BoolValue(false),
				VarsJson:    jsontypes.NewNormalizedValue(`{"input2_var":"value2"}`),
				StreamsJson: jsontypes.NewNormalizedValue(string(streamsJSON2)),
			},
		}
		inputList, diags := types.ListValueFrom(ctx, v1.GetInputType(), inputsV1)
		require.Empty(t, diags)

		v1Model := v1.IntegrationPolicyModel{
			ID:                 types.StringValue("test-id"),
			Name:               types.StringValue("test-name"),
			Namespace:          types.StringValue("test-namespace"),
			AgentPolicyID:      types.StringValue("agent-policy-1"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			Input:              inputList,
		}

		v2Model, diags := NewFromV1(t.Context(), v1Model)
		require.Empty(t, diags)
		assert.False(t, v2Model.Inputs.IsNull())

		var inputsMap map[string]integrationPolicyInputsModel
		d := v2Model.Inputs.ElementsAs(ctx, &inputsMap, false)
		require.Empty(t, d)
		require.Len(t, inputsMap, 2)

		// Verify input-1
		input1 := inputsMap["input-1"]
		assert.Equal(t, types.BoolValue(true), input1.Enabled)
		var streams1 map[string]integrationPolicyInputStreamModel
		d = input1.Streams.ElementsAs(ctx, &streams1, false)
		require.Empty(t, d)
		require.Contains(t, streams1, "stream-1")

		// Verify input-2
		input2 := inputsMap["input-2"]
		assert.Equal(t, types.BoolValue(false), input2.Enabled)
		var streams2 map[string]integrationPolicyInputStreamModel
		d = input2.Streams.ElementsAs(ctx, &streams2, false)
		require.Empty(t, d)
		require.Contains(t, streams2, "stream-2")
	})

	t.Run("conversion with invalid streams JSON", func(t *testing.T) {
		// Use valid JSON that doesn't match the expected structure
		inputsV1 := []v1.IntegrationPolicyInputModel{
			{
				InputID:     types.StringValue("input-1"),
				Enabled:     types.BoolValue(true),
				VarsJson:    jsontypes.NewNormalizedValue(`{"input_var":"value"}`),
				StreamsJson: jsontypes.NewNormalizedValue(`["array", "instead", "of", "map"]`),
			},
		}
		inputList, diags := types.ListValueFrom(ctx, v1.GetInputType(), inputsV1)
		require.Empty(t, diags)

		v1Model := v1.IntegrationPolicyModel{
			ID:                 types.StringValue("test-id"),
			Name:               types.StringValue("test-name"),
			Namespace:          types.StringValue("test-namespace"),
			AgentPolicyID:      types.StringValue("agent-policy-1"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			Input:              inputList,
		}

		_, diags = NewFromV1(t.Context(), v1Model)
		require.NotEmpty(t, diags)
		assert.True(t, diags.HasError())
	})

	t.Run("conversion preserves null and unknown values", func(t *testing.T) {
		v1Model := v1.IntegrationPolicyModel{
			ID:                 types.StringValue("test-id"),
			Name:               types.StringValue("test-name"),
			Namespace:          types.StringValue("test-namespace"),
			AgentPolicyID:      types.StringValue("agent-policy-1"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			Description:        types.StringNull(),
			VarsJson:           jsontypes.NewNormalizedNull(),
			SpaceIds:           types.SetNull(types.StringType),
			Input:              types.ListNull(v1.GetInputType()),
		}

		v2Model, diags := NewFromV1(t.Context(), v1Model)
		require.Empty(t, diags)

		assert.True(t, v2Model.Description.IsNull())
		assert.True(t, v2Model.VarsJson.IsNull())
		assert.True(t, v2Model.SpaceIds.IsNull())
	})
}
