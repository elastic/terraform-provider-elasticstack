package api_key

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
)

func TestRoleDescriptorsValue_Type(t *testing.T) {
	value := RoleDescriptorsValue{}
	ctx := context.Background()

	attrType := value.Type(ctx)

	assert.IsType(t, RoleDescriptorsType{}, attrType)
}

func TestRoleDescriptorsValue_WithDefaults(t *testing.T) {
	tests := []struct {
		name           string
		input          RoleDescriptorsValue
		expectedResult func(t *testing.T, result RoleDescriptorsValue, diags diag.Diagnostics)
		expectError    bool
	}{
		{
			name:  "null value returns same value without error",
			input: NewRoleDescriptorsNull(),
			expectedResult: func(t *testing.T, result RoleDescriptorsValue, diags diag.Diagnostics) {
				assert.True(t, result.IsNull())
				assert.False(t, diags.HasError())
			},
			expectError: false,
		},
		{
			name:  "unknown value returns same value without error",
			input: NewRoleDescriptorsUnknown(),
			expectedResult: func(t *testing.T, result RoleDescriptorsValue, diags diag.Diagnostics) {
				assert.True(t, result.IsUnknown())
				assert.False(t, diags.HasError())
			},
			expectError: false,
		},
		{
			name:  "valid JSON with missing allow_restricted_indices sets default",
			input: NewRoleDescriptorsValue(`{"admin":{"indices":[{"names":["index1"],"privileges":["read"]}]}}`),
			expectedResult: func(t *testing.T, result RoleDescriptorsValue, diags diag.Diagnostics) {
				assert.False(t, result.IsNull())
				assert.False(t, result.IsUnknown())
				assert.False(t, diags.HasError())
				assert.Contains(t, result.ValueString(), "allow_restricted_indices")
				assert.Contains(t, result.ValueString(), "false")
			},
			expectError: false,
		},
		{
			name:  "valid JSON with existing allow_restricted_indices preserves value",
			input: NewRoleDescriptorsValue(`{"admin":{"indices":[{"names":["index1"],"privileges":["read"],"allow_restricted_indices":true}]}}`),
			expectedResult: func(t *testing.T, result RoleDescriptorsValue, diags diag.Diagnostics) {
				assert.False(t, result.IsNull())
				assert.False(t, result.IsUnknown())
				assert.False(t, diags.HasError())
				assert.Contains(t, result.ValueString(), "allow_restricted_indices")
				assert.Contains(t, result.ValueString(), "true")
			},
			expectError: false,
		},
		{
			name:  "empty role descriptor object",
			input: NewRoleDescriptorsValue(`{"admin":{}}`),
			expectedResult: func(t *testing.T, result RoleDescriptorsValue, diags diag.Diagnostics) {
				assert.False(t, result.IsNull())
				assert.False(t, result.IsUnknown())
				assert.False(t, diags.HasError())
			},
			expectError: false,
		},
		{
			name:        "invalid JSON returns error",
			input:       NewRoleDescriptorsValue(`{"invalid json"`),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.input.WithDefaults()

			if tt.expectError {
				assert.True(t, diags.HasError())
			} else {
				if tt.expectedResult != nil {
					tt.expectedResult(t, result, diags)
				}
			}
		})
	}
}

func TestRoleDescriptorsValue_StringSemanticEquals(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		value1      RoleDescriptorsValue
		value2      basetypes.StringValuable
		expected    bool
		expectError bool
	}{
		{
			name:        "both null values are equal",
			value1:      NewRoleDescriptorsNull(),
			value2:      NewRoleDescriptorsNull(),
			expected:    true,
			expectError: false,
		},
		{
			name:        "both unknown values are equal",
			value1:      NewRoleDescriptorsUnknown(),
			value2:      NewRoleDescriptorsUnknown(),
			expected:    true,
			expectError: false,
		},
		{
			name:        "null vs unknown are not equal",
			value1:      NewRoleDescriptorsNull(),
			value2:      NewRoleDescriptorsUnknown(),
			expected:    false,
			expectError: false,
		},
		{
			name:        "same JSON content are equal",
			value1:      NewRoleDescriptorsValue(`{"admin":{"cluster":["read"]}}`),
			value2:      NewRoleDescriptorsValue(`{"admin":{"cluster":["read"]}}`),
			expected:    true,
			expectError: false,
		},
		{
			name:        "different JSON content are not equal",
			value1:      NewRoleDescriptorsValue(`{"admin":{"cluster":["read"]}}`),
			value2:      NewRoleDescriptorsValue(`{"user":{"cluster":["write"]}}`),
			expected:    false,
			expectError: false,
		},
		{
			name:        "semantic equality with defaults - missing vs explicit false",
			value1:      NewRoleDescriptorsValue(`{"admin":{"indices":[{"names":["index1"],"privileges":["read"]}]}}`),
			value2:      NewRoleDescriptorsValue(`{"admin":{"indices":[{"names":["index1"],"privileges":["read"],"allow_restricted_indices":false}]}}`),
			expected:    true,
			expectError: false,
		},
		{
			name:        "wrong type returns error",
			value1:      NewRoleDescriptorsValue(`{"admin":{}}`),
			value2:      basetypes.NewStringValue("not a role descriptors value"),
			expected:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.value1.StringSemanticEquals(ctx, tt.value2)

			if tt.expectError {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRoleDescriptorsValue_WithDefaults_ComplexJSON(t *testing.T) {
	// Test with complex role descriptor JSON that has multiple roles and indices
	complexJSON := `{
		"admin": {
			"cluster": ["all"],
			"indices": [
				{"names": ["index1"], "privileges": ["read"]},
				{"names": ["index2"], "privileges": ["write"], "allow_restricted_indices": true}
			]
		},
		"user": {
			"indices": [
				{"names": ["public*"], "privileges": ["read"]}
			]
		}
	}`

	value := NewRoleDescriptorsValue(complexJSON)
	result, diags := value.WithDefaults()

	assert.False(t, diags.HasError())
	assert.False(t, result.IsNull())
	assert.False(t, result.IsUnknown())

	resultJSON := result.ValueString()

	// Should contain the original true value
	assert.Contains(t, resultJSON, `"allow_restricted_indices":true`)
	// Should contain default false values for indices without the field
	assert.Contains(t, resultJSON, `"allow_restricted_indices":false`)
}
