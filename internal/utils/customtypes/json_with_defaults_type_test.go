package customtypes

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// testPopulateDefaults is a test function that mimics testPopulateDefaults for testing purposes
func testPopulateDefaults(model map[string]models.ApiKeyRoleDescriptor) map[string]models.ApiKeyRoleDescriptor {
	result := make(map[string]models.ApiKeyRoleDescriptor)

	for role, descriptor := range model {
		resultDescriptor := descriptor

		// Ensure AllowRestrictedIndices is set to false for all indices that don't have it set
		for i, index := range resultDescriptor.Indices {
			if index.AllowRestrictedIndices == nil {
				resultDescriptor.Indices[i].AllowRestrictedIndices = new(bool)
				*resultDescriptor.Indices[i].AllowRestrictedIndices = false
			}
		}

		result[role] = resultDescriptor
	}

	return result
}

func TestRoleDescriptorsType_String(t *testing.T) {
	roleDescriptorsType := NewJSONWithDefaultsType(testPopulateDefaults)
	expected := "customtypes.JSONWithDefaultsType"
	actual := roleDescriptorsType.String()
	assert.Equal(t, expected, actual)
}

func TestRoleDescriptorsType_ValueType(t *testing.T) {
	roleDescriptorsType := NewJSONWithDefaultsType(testPopulateDefaults)
	ctx := context.Background()

	value := roleDescriptorsType.ValueType(ctx)

	expectedType := JSONWithDefaultsValue[map[string]models.ApiKeyRoleDescriptor]{}
	assert.IsType(t, expectedType, value)
}

func TestRoleDescriptorsType_Equal(t *testing.T) {
	tests := []struct {
		name     string
		thisType JSONWithDefaultsType[map[string]models.ApiKeyRoleDescriptor]
		other    attr.Type
		expected bool
	}{
		{
			name:     "equal to same type",
			thisType: NewJSONWithDefaultsType(testPopulateDefaults),
			other:    NewJSONWithDefaultsType(testPopulateDefaults),
			expected: true,
		},
		{
			name:     "not equal to different type",
			thisType: NewJSONWithDefaultsType(testPopulateDefaults),
			other:    basetypes.StringType{},
			expected: false,
		},
		{
			name:     "not equal to nil",
			thisType: NewJSONWithDefaultsType(testPopulateDefaults),
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
		expectedValue JSONWithDefaultsValue[map[string]models.ApiKeyRoleDescriptor]
		expectedDiags bool
	}{
		{
			name:  "valid string value",
			input: basetypes.NewStringValue(`{"role1": {"cluster": ["read"]}}`),
			expectedValue: NewJSONWithDefaultsValue(
				`{"role1": {"cluster": ["read"]}}`,
				testPopulateDefaults,
			),
			expectedDiags: false,
		},
		{
			name:  "null string value",
			input: basetypes.NewStringNull(),
			expectedValue: NewJSONWithDefaultsNull(
				testPopulateDefaults,
			),
			expectedDiags: false,
		},
		{
			name:  "unknown string value",
			input: basetypes.NewStringUnknown(),
			expectedValue: NewJSONWithDefaultsUnknown(
				testPopulateDefaults,
			),
			expectedDiags: false,
		},
		{
			name:  "empty string value",
			input: basetypes.NewStringValue(""),
			expectedValue: JSONWithDefaultsValue[map[string]models.ApiKeyRoleDescriptor]{
				Normalized: jsontypes.NewNormalizedValue(""),
			},
			expectedDiags: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleDescriptorsType := NewJSONWithDefaultsType(testPopulateDefaults)
			ctx := context.Background()

			value, diags := roleDescriptorsType.ValueFromString(ctx, tt.input)

			if tt.expectedDiags {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
			}

			// For value comparison, we check the string representation since the internal structure might differ
			if !tt.expectedDiags {
				actualValue, ok := value.(JSONWithDefaultsValue[map[string]models.ApiKeyRoleDescriptor])
				assert.True(t, ok)
				assert.Equal(t, tt.expectedValue.IsNull(), actualValue.IsNull())
				assert.Equal(t, tt.expectedValue.IsUnknown(), actualValue.IsUnknown())
				if !tt.expectedValue.IsNull() && !tt.expectedValue.IsUnknown() {
					assert.Equal(t, tt.expectedValue.ValueString(), actualValue.ValueString())
				}
			}
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
			expectedType:  JSONWithDefaultsValue[map[string]models.ApiKeyRoleDescriptor]{},
		},
		{
			name:          "null terraform value",
			input:         tftypes.NewValue(tftypes.String, nil),
			expectedError: false,
			expectedType:  JSONWithDefaultsValue[map[string]models.ApiKeyRoleDescriptor]{},
		},
		{
			name:          "unknown terraform value",
			input:         tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectedError: false,
			expectedType:  JSONWithDefaultsValue[map[string]models.ApiKeyRoleDescriptor]{},
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
			roleDescriptorsType := NewJSONWithDefaultsType(testPopulateDefaults)
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
