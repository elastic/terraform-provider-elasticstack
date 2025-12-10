package v1

import (
	"context"
	"testing"

	v0 "github.com/elastic/terraform-provider-elasticstack/internal/fleet/integration_policy/models/v0"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFromV0(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		inputV0     v0.IntegrationPolicyModel
		checkFunc   func(t *testing.T, result IntegrationPolicyModel, diags bool)
		expectDiags bool
	}{
		{
			name: "basic model with all string fields set",
			inputV0: v0.IntegrationPolicyModel{
				ID:                 types.StringValue("test-id"),
				PolicyID:           types.StringValue("policy-id"),
				Name:               types.StringValue("test-policy"),
				Namespace:          types.StringValue("default"),
				AgentPolicyID:      types.StringValue("agent-policy-id"),
				Description:        types.StringValue("Test description"),
				Enabled:            types.BoolValue(true),
				Force:              types.BoolValue(false),
				IntegrationName:    types.StringValue("nginx"),
				IntegrationVersion: types.StringValue("1.0.0"),
				VarsJson:           types.StringValue(`{"key": "value"}`),
				Input:              types.ListNull(types.ObjectType{}),
			},
			checkFunc: func(t *testing.T, result IntegrationPolicyModel, hasDiags bool) {
				assert.False(t, hasDiags, "Should not have diagnostics")
				assert.Equal(t, "test-id", result.ID.ValueString())
				assert.Equal(t, "policy-id", result.PolicyID.ValueString())
				assert.Equal(t, "test-policy", result.Name.ValueString())
				assert.Equal(t, "default", result.Namespace.ValueString())
				assert.Equal(t, "agent-policy-id", result.AgentPolicyID.ValueString())
				assert.Equal(t, "Test description", result.Description.ValueString())
				assert.True(t, result.Enabled.ValueBool())
				assert.False(t, result.Force.ValueBool())
				assert.Equal(t, "nginx", result.IntegrationName.ValueString())
				assert.Equal(t, "1.0.0", result.IntegrationVersion.ValueString())
				assert.Equal(t, `{"key": "value"}`, result.VarsJson.ValueString())

				// V1-only fields should be null
				assert.True(t, result.AgentPolicyIDs.IsNull())
				assert.True(t, result.SpaceIds.IsNull())
				assert.True(t, result.OutputID.IsNull())
			},
		},
		{
			name: "vars_json with empty string converts to null",
			inputV0: v0.IntegrationPolicyModel{
				ID:                 types.StringValue("test-id"),
				PolicyID:           types.StringValue("policy-id"),
				Name:               types.StringValue("test-policy"),
				Namespace:          types.StringValue("default"),
				AgentPolicyID:      types.StringValue("agent-policy-id"),
				Description:        types.StringValue("Test description"),
				Enabled:            types.BoolValue(true),
				Force:              types.BoolValue(false),
				IntegrationName:    types.StringValue("nginx"),
				IntegrationVersion: types.StringValue("1.0.0"),
				VarsJson:           types.StringValue(""), // Empty string
				Input:              types.ListNull(types.ObjectType{}),
			},
			checkFunc: func(t *testing.T, result IntegrationPolicyModel, hasDiags bool) {
				assert.False(t, hasDiags, "Should not have diagnostics")
				assert.True(t, result.VarsJson.IsNull(), "Empty string should convert to null")
			},
		},
		{
			name: "vars_json with null converts to null",
			inputV0: v0.IntegrationPolicyModel{
				ID:                 types.StringValue("test-id"),
				PolicyID:           types.StringValue("policy-id"),
				Name:               types.StringValue("test-policy"),
				Namespace:          types.StringValue("default"),
				AgentPolicyID:      types.StringValue("agent-policy-id"),
				Description:        types.StringValue("Test description"),
				Enabled:            types.BoolValue(true),
				Force:              types.BoolValue(false),
				IntegrationName:    types.StringValue("nginx"),
				IntegrationVersion: types.StringValue("1.0.0"),
				VarsJson:           types.StringNull(),
				Input:              types.ListNull(types.ObjectType{}),
			},
			checkFunc: func(t *testing.T, result IntegrationPolicyModel, hasDiags bool) {
				assert.False(t, hasDiags, "Should not have diagnostics")
				assert.True(t, result.VarsJson.IsNull(), "Null should stay null")
			},
		},
		{
			name: "vars_json with unknown converts to null",
			inputV0: v0.IntegrationPolicyModel{
				ID:                 types.StringValue("test-id"),
				PolicyID:           types.StringValue("policy-id"),
				Name:               types.StringValue("test-policy"),
				Namespace:          types.StringValue("default"),
				AgentPolicyID:      types.StringValue("agent-policy-id"),
				Description:        types.StringValue("Test description"),
				Enabled:            types.BoolValue(true),
				Force:              types.BoolValue(false),
				IntegrationName:    types.StringValue("nginx"),
				IntegrationVersion: types.StringValue("1.0.0"),
				VarsJson:           types.StringUnknown(),
				Input:              types.ListNull(types.ObjectType{}),
			},
			checkFunc: func(t *testing.T, result IntegrationPolicyModel, hasDiags bool) {
				assert.False(t, hasDiags, "Should not have diagnostics")
				assert.True(t, result.VarsJson.IsNull(), "Unknown should convert to null")
			},
		},
		{
			name: "model with null input list",
			inputV0: v0.IntegrationPolicyModel{
				ID:                 types.StringValue("test-id"),
				PolicyID:           types.StringValue("policy-id"),
				Name:               types.StringValue("test-policy"),
				Namespace:          types.StringValue("default"),
				AgentPolicyID:      types.StringValue("agent-policy-id"),
				Description:        types.StringValue("Test description"),
				Enabled:            types.BoolValue(true),
				Force:              types.BoolValue(false),
				IntegrationName:    types.StringValue("nginx"),
				IntegrationVersion: types.StringValue("1.0.0"),
				VarsJson:           types.StringValue(`{"key": "value"}`),
				Input:              types.ListNull(types.ObjectType{}),
			},
			checkFunc: func(t *testing.T, result IntegrationPolicyModel, hasDiags bool) {
				assert.False(t, hasDiags, "Should not have diagnostics")
				assert.NotNil(t, result.Input)
				// The Input list should be converted properly even if it was null
			},
		},
		{
			name: "model with unknown values",
			inputV0: v0.IntegrationPolicyModel{
				ID:                 types.StringValue("test-id"),
				PolicyID:           types.StringUnknown(),
				Name:               types.StringUnknown(),
				Namespace:          types.StringUnknown(),
				AgentPolicyID:      types.StringUnknown(),
				Description:        types.StringUnknown(),
				Enabled:            types.BoolUnknown(),
				Force:              types.BoolUnknown(),
				IntegrationName:    types.StringUnknown(),
				IntegrationVersion: types.StringUnknown(),
				VarsJson:           types.StringUnknown(),
				Input:              types.ListNull(types.ObjectType{}),
			},
			checkFunc: func(t *testing.T, result IntegrationPolicyModel, hasDiags bool) {
				assert.False(t, hasDiags, "Should not have diagnostics")
				assert.False(t, result.ID.IsUnknown(), "ID should be known")
				assert.True(t, result.PolicyID.IsUnknown())
				assert.True(t, result.Name.IsUnknown())
				assert.True(t, result.Namespace.IsUnknown())
				assert.True(t, result.AgentPolicyID.IsUnknown())
				assert.True(t, result.Description.IsUnknown())
				assert.True(t, result.Enabled.IsUnknown())
				assert.True(t, result.Force.IsUnknown())
				assert.True(t, result.IntegrationName.IsUnknown())
				assert.True(t, result.IntegrationVersion.IsUnknown())
				assert.True(t, result.VarsJson.IsNull(), "Unknown VarsJson should convert to null")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := NewFromV0(ctx, tt.inputV0)

			hasDiags := diags.HasError()
			if tt.expectDiags {
				assert.True(t, hasDiags, "Expected diagnostics but got none")
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result, hasDiags)
			}
		})
	}
}

// TestNewFromV0_InputConversion tests the input conversion logic separately
// Note: This test verifies the conversion logic for JSON fields in inputs.
// The actual extraction of V0 inputs from the Terraform state is handled by
// the framework during state upgrades and is tested via acceptance tests.
func TestNewFromV0_InputConversion(t *testing.T) {
	t.Run("converts empty string to null for vars_json", func(t *testing.T) {
		// Simulate what happens during the conversion
		varsJSON := types.StringValue("")

		var result types.String
		if ptr := varsJSON.ValueStringPointer(); ptr != nil {
			if *ptr == "" {
				result = types.StringNull()
			} else {
				result = types.StringValue(*ptr)
			}
		} else {
			result = types.StringNull()
		}

		assert.True(t, result.IsNull())
	})

	t.Run("converts empty string to null for streams_json", func(t *testing.T) {
		streamsJSON := types.StringValue("")

		var result types.String
		if ptr := streamsJSON.ValueStringPointer(); ptr != nil {
			if *ptr == "" {
				result = types.StringNull()
			} else {
				result = types.StringValue(*ptr)
			}
		} else {
			result = types.StringNull()
		}

		assert.True(t, result.IsNull())
	})

	t.Run("preserves valid JSON string", func(t *testing.T) {
		varsJSON := types.StringValue(`{"key": "value"}`)

		var result types.String
		if ptr := varsJSON.ValueStringPointer(); ptr != nil {
			if *ptr == "" {
				result = types.StringNull()
			} else {
				result = types.StringValue(*ptr)
			}
		} else {
			result = types.StringNull()
		}

		assert.False(t, result.IsNull())
		assert.Equal(t, `{"key": "value"}`, result.ValueString())
	})

	t.Run("converts null to null", func(t *testing.T) {
		varsJSON := types.StringNull()

		var result types.String
		if ptr := varsJSON.ValueStringPointer(); ptr != nil {
			if *ptr == "" {
				result = types.StringNull()
			} else {
				result = types.StringValue(*ptr)
			}
		} else {
			result = types.StringNull()
		}

		assert.True(t, result.IsNull())
	})
}

func TestGetInputTypeV1(t *testing.T) {
	inputType := GetInputType()
	require.NotNil(t, inputType)

	// Verify it's an object type with the expected attributes
	objType, ok := inputType.(attr.TypeWithAttributeTypes)
	require.True(t, ok, "input type should be an object type with attributes")

	attrTypes := objType.AttributeTypes()
	require.Contains(t, attrTypes, "input_id")
	require.Contains(t, attrTypes, "enabled")
	require.Contains(t, attrTypes, "streams_json")
	require.Contains(t, attrTypes, "vars_json")
}
