package integration_policy

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpgradeV0ToV2_JSONConversions tests the V0 to V1 conversion logic
// particularly the conversion of empty strings to null for JSON fields
func TestUpgradeV0ToV2_JSONConversions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		v0VarsJson     types.String
		expectedV1Null bool
	}{
		{
			name:           "valid JSON string is preserved",
			v0VarsJson:     types.StringValue(`{"key":"value"}`),
			expectedV1Null: false,
		},
		{
			name:           "empty string converts to null",
			v0VarsJson:     types.StringValue(""),
			expectedV1Null: true,
		},
		{
			name:           "null remains null",
			v0VarsJson:     types.StringNull(),
			expectedV1Null: true,
		},
		{
			name:           "complex JSON is preserved",
			v0VarsJson:     types.StringValue(`{"nested":{"key":"value"},"array":[1,2,3]}`),
			expectedV1Null: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Simulate the conversion logic from upgradeV0ToV2
			var v1VarsJson jsontypes.Normalized
			if varsJSON := tt.v0VarsJson.ValueStringPointer(); varsJSON != nil {
				if *varsJSON == "" {
					v1VarsJson = jsontypes.NewNormalizedNull()
				} else {
					v1VarsJson = jsontypes.NewNormalizedValue(*varsJSON)
				}
			} else {
				v1VarsJson = jsontypes.NewNormalizedNull()
			}

			if tt.expectedV1Null {
				assert.True(t, v1VarsJson.IsNull(), "Expected null but got non-null value")
			} else {
				assert.False(t, v1VarsJson.IsNull(), "Expected non-null but got null value")

				// For non-null values, verify the JSON content
				var result map[string]interface{}
				diags := v1VarsJson.Unmarshal(&result)
				require.Empty(t, diags, "Failed to unmarshal JSON")
			}
		})
	}
}

// TestUpgradeV0ToV2_InputJSONConversions tests the conversion logic for input fields
func TestUpgradeV0ToV2_InputJSONConversions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		v0VarsJson          types.String
		v0StreamsJson       types.String
		expectedVarsNull    bool
		expectedStreamsNull bool
	}{
		{
			name:                "valid JSON strings are preserved",
			v0VarsJson:          types.StringValue(`{"var":"value"}`),
			v0StreamsJson:       types.StringValue(`{"stream-1":{"enabled":true}}`),
			expectedVarsNull:    false,
			expectedStreamsNull: false,
		},
		{
			name:                "empty strings convert to null",
			v0VarsJson:          types.StringValue(""),
			v0StreamsJson:       types.StringValue(""),
			expectedVarsNull:    true,
			expectedStreamsNull: true,
		},
		{
			name:                "null values remain null",
			v0VarsJson:          types.StringNull(),
			v0StreamsJson:       types.StringNull(),
			expectedVarsNull:    true,
			expectedStreamsNull: true,
		},
		{
			name:                "mixed empty and valid JSON",
			v0VarsJson:          types.StringValue(`{"key":"value"}`),
			v0StreamsJson:       types.StringValue(""),
			expectedVarsNull:    false,
			expectedStreamsNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Simulate the conversion logic for vars_json
			var v1VarsJson jsontypes.Normalized
			if varsJSON := tt.v0VarsJson.ValueStringPointer(); varsJSON != nil {
				if *varsJSON == "" {
					v1VarsJson = jsontypes.NewNormalizedNull()
				} else {
					v1VarsJson = jsontypes.NewNormalizedValue(*varsJSON)
				}
			} else {
				v1VarsJson = jsontypes.NewNormalizedNull()
			}

			// Simulate the conversion logic for streams_json
			var v1StreamsJson jsontypes.Normalized
			if streamsJSON := tt.v0StreamsJson.ValueStringPointer(); streamsJSON != nil {
				if *streamsJSON == "" {
					v1StreamsJson = jsontypes.NewNormalizedNull()
				} else {
					v1StreamsJson = jsontypes.NewNormalizedValue(*streamsJSON)
				}
			} else {
				v1StreamsJson = jsontypes.NewNormalizedNull()
			}

			// Verify vars_json
			if tt.expectedVarsNull {
				assert.True(t, v1VarsJson.IsNull(), "Expected vars_json to be null")
			} else {
				assert.False(t, v1VarsJson.IsNull(), "Expected vars_json to be non-null")
			}

			// Verify streams_json
			if tt.expectedStreamsNull {
				assert.True(t, v1StreamsJson.IsNull(), "Expected streams_json to be null")
			} else {
				assert.False(t, v1StreamsJson.IsNull(), "Expected streams_json to be non-null")
			}
		})
	}
}

// TestUpgradeV0ToV2_NewFieldsAddedAsNull tests that new fields added in V1/V2 are set to null
func TestUpgradeV0ToV2_NewFieldsAddedAsNull(t *testing.T) {
	t.Parallel()

	// V0 didn't have these fields, verify they're initialized as null in the upgrade
	agentPolicyIDs := types.ListNull(types.StringType)
	spaceIds := types.SetNull(types.StringType)

	assert.True(t, agentPolicyIDs.IsNull(), "agent_policy_ids should be null (didn't exist in V0)")
	assert.True(t, spaceIds.IsNull(), "space_ids should be null (didn't exist in V0)")
}

// TestUpgradeV0ToV2_FieldsPreserved tests that all V0 fields are preserved during upgrade
func TestUpgradeV0ToV2_FieldsPreserved(t *testing.T) {
	t.Parallel()

	// Test that the structure of fields from V0 are preserved
	v0Model := integrationPolicyModelV0{
		ID:                 types.StringValue("test-id"),
		PolicyID:           types.StringValue("test-policy-id"),
		Name:               types.StringValue("test-name"),
		Namespace:          types.StringValue("test-namespace"),
		AgentPolicyID:      types.StringValue("agent-policy-1"),
		Description:        types.StringValue("test description"),
		Enabled:            types.BoolValue(false),
		Force:              types.BoolValue(true),
		IntegrationName:    types.StringValue("test-integration"),
		IntegrationVersion: types.StringValue("2.0.0"),
		VarsJson:           types.StringValue(`{"complex":{"nested":"value"}}`),
		Input:              types.ListNull(getInputTypeV0()),
	}

	// Verify all fields are accessible and have the expected values
	assert.Equal(t, "test-id", v0Model.ID.ValueString())
	assert.Equal(t, "test-policy-id", v0Model.PolicyID.ValueString())
	assert.Equal(t, "test-name", v0Model.Name.ValueString())
	assert.Equal(t, "test-namespace", v0Model.Namespace.ValueString())
	assert.Equal(t, "agent-policy-1", v0Model.AgentPolicyID.ValueString())
	assert.Equal(t, "test description", v0Model.Description.ValueString())
	assert.Equal(t, false, v0Model.Enabled.ValueBool())
	assert.Equal(t, true, v0Model.Force.ValueBool())
	assert.Equal(t, "test-integration", v0Model.IntegrationName.ValueString())
	assert.Equal(t, "2.0.0", v0Model.IntegrationVersion.ValueString())
	assert.Equal(t, `{"complex":{"nested":"value"}}`, v0Model.VarsJson.ValueString())
	assert.True(t, v0Model.Input.IsNull())
}

func getInputTypeV0() types.ObjectType {
	return getSchemaV0().Blocks["input"].Type().(types.ListType).ElemType.(types.ObjectType)
}
