package api_key

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestRoleDescriptorsType_String(t *testing.T) {
	roleDescriptorsType := RoleDescriptorsType{}
	expected := "api_key.RoleDescriptorsType"
	actual := roleDescriptorsType.String()
	assert.Equal(t, expected, actual)
}

func TestRoleDescriptorsType_ValueType(t *testing.T) {
	roleDescriptorsType := RoleDescriptorsType{}
	ctx := context.Background()

	value := roleDescriptorsType.ValueType(ctx)

	assert.IsType(t, RoleDescriptorsValue{}, value)
}

func TestRoleDescriptorsType_Equal(t *testing.T) {
	tests := []struct {
		name     string
		thisType RoleDescriptorsType
		other    attr.Type
		expected bool
	}{
		{
			name:     "equal to same type",
			thisType: RoleDescriptorsType{},
			other:    RoleDescriptorsType{},
			expected: true,
		},
		{
			name:     "not equal to different type",
			thisType: RoleDescriptorsType{},
			other:    basetypes.StringType{},
			expected: false,
		},
		{
			name:     "not equal to nil",
			thisType: RoleDescriptorsType{},
			other:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.thisType.Equal(tt.other)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestRoleDescriptorsType_ValueFromString(t *testing.T) {
	tests := []struct {
		name          string
		input         basetypes.StringValue
		expectedValue RoleDescriptorsValue
		expectedDiags bool
	}{
		{
			name:  "valid string value",
			input: basetypes.NewStringValue(`{"role1": {"cluster": ["read"]}}`),
			expectedValue: RoleDescriptorsValue{
				Normalized: jsontypes.Normalized{
					StringValue: basetypes.NewStringValue(`{"role1": {"cluster": ["read"]}}`),
				},
			},
			expectedDiags: false,
		},
		{
			name:  "null string value",
			input: basetypes.NewStringNull(),
			expectedValue: RoleDescriptorsValue{
				Normalized: jsontypes.Normalized{
					StringValue: basetypes.NewStringNull(),
				},
			},
			expectedDiags: false,
		},
		{
			name:  "unknown string value",
			input: basetypes.NewStringUnknown(),
			expectedValue: RoleDescriptorsValue{
				Normalized: jsontypes.Normalized{
					StringValue: basetypes.NewStringUnknown(),
				},
			},
			expectedDiags: false,
		},
		{
			name:  "empty string value",
			input: basetypes.NewStringValue(""),
			expectedValue: RoleDescriptorsValue{
				Normalized: jsontypes.Normalized{
					StringValue: basetypes.NewStringValue(""),
				},
			},
			expectedDiags: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleDescriptorsType := RoleDescriptorsType{}
			ctx := context.Background()

			value, diags := roleDescriptorsType.ValueFromString(ctx, tt.input)

			if tt.expectedDiags {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
			}

			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

func TestRoleDescriptorsType_ValueFromTerraform(t *testing.T) {
	tests := []struct {
		name          string
		input         tftypes.Value
		expectedError bool
		expectedType  interface{}
	}{
		{
			name:          "valid string terraform value",
			input:         tftypes.NewValue(tftypes.String, `{"role1": {"cluster": ["read"]}}`),
			expectedError: false,
			expectedType:  RoleDescriptorsValue{},
		},
		{
			name:          "null terraform value",
			input:         tftypes.NewValue(tftypes.String, nil),
			expectedError: false,
			expectedType:  RoleDescriptorsValue{},
		},
		{
			name:          "unknown terraform value",
			input:         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectedError: false,
			expectedType:  RoleDescriptorsValue{},
		},
		{
			name:          "invalid terraform value type",
			input:         tftypes.NewValue(tftypes.Number, 123),
			expectedError: true,
			expectedType:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleDescriptorsType := RoleDescriptorsType{}
			ctx := context.Background()

			value, err := roleDescriptorsType.ValueFromTerraform(ctx, tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, value)
			} else {
				assert.NoError(t, err)
				if tt.expectedType != nil {
					assert.IsType(t, tt.expectedType, value)
				}
			}
		})
	}
}
