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

package customtypes

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
)

func TestRoleDescriptorsValue_Type(t *testing.T) {
	value := NewJSONWithDefaultsNull(testPopulateDefaults)
	ctx := context.Background()

	attrType := value.Type(ctx)

	expectedType := NewJSONWithDefaultsType(testPopulateDefaults)
	assert.IsType(t, expectedType, attrType)
}

func TestRoleDescriptorsValue_WithDefaults(t *testing.T) {
	tests := []struct {
		name           string
		input          JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor]
		expectedResult func(t *testing.T, result JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor], diags diag.Diagnostics)
		expectError    bool
	}{
		{
			name:  "null value returns same value without error",
			input: NewJSONWithDefaultsNull(testPopulateDefaults),
			expectedResult: func(t *testing.T, result JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor], diags diag.Diagnostics) {
				assert.True(t, result.IsNull())
				assert.False(t, diags.HasError())
			},
			expectError: false,
		},
		{
			name:  "unknown value returns same value without error",
			input: NewJSONWithDefaultsUnknown(testPopulateDefaults),
			expectedResult: func(t *testing.T, result JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor], diags diag.Diagnostics) {
				assert.True(t, result.IsUnknown())
				assert.False(t, diags.HasError())
			},
			expectError: false,
		},
		{
			name:  "valid JSON with missing allow_restricted_indices sets default",
			input: NewJSONWithDefaultsValue(`{"admin":{"indices":[{"names":["index1"],"privileges":["read"]}]}}`, testPopulateDefaults),
			expectedResult: func(t *testing.T, result JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor], diags diag.Diagnostics) {
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
			input: NewJSONWithDefaultsValue(`{"admin":{"indices":[{"names":["index1"],"privileges":["read"],"allow_restricted_indices":true}]}}`, testPopulateDefaults),
			expectedResult: func(t *testing.T, result JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor], diags diag.Diagnostics) {
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
			input: NewJSONWithDefaultsValue(`{"admin":{}}`, testPopulateDefaults),
			expectedResult: func(t *testing.T, result JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor], diags diag.Diagnostics) {
				assert.False(t, result.IsNull())
				assert.False(t, result.IsUnknown())
				assert.False(t, diags.HasError())
			},
			expectError: false,
		},
		{
			name:        "invalid JSON returns error",
			input:       NewJSONWithDefaultsValue(`{"invalid json"`, testPopulateDefaults),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.input.WithDefaults()

			if tt.expectError {
				assert.True(t, diags.HasError())
			} else if tt.expectedResult != nil {
				tt.expectedResult(t, result, diags)
			}
		})
	}
}

func TestRoleDescriptorsValue_StringSemanticEquals(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		value1      JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor]
		value2      basetypes.StringValuable
		expected    bool
		expectError bool
	}{
		{
			name:        "both null values are equal",
			value1:      NewJSONWithDefaultsNull(testPopulateDefaults),
			value2:      NewJSONWithDefaultsNull(testPopulateDefaults),
			expected:    true,
			expectError: false,
		},
		{
			name:        "both unknown values are equal",
			value1:      NewJSONWithDefaultsUnknown(testPopulateDefaults),
			value2:      NewJSONWithDefaultsUnknown(testPopulateDefaults),
			expected:    true,
			expectError: false,
		},
		{
			name:        "null vs unknown are not equal",
			value1:      NewJSONWithDefaultsNull(testPopulateDefaults),
			value2:      NewJSONWithDefaultsUnknown(testPopulateDefaults),
			expected:    false,
			expectError: false,
		},
		{
			name:        "same JSON content are equal",
			value1:      NewJSONWithDefaultsValue(`{"admin":{"cluster":["read"]}}`, testPopulateDefaults),
			value2:      NewJSONWithDefaultsValue(`{"admin":{"cluster":["read"]}}`, testPopulateDefaults),
			expected:    true,
			expectError: false,
		},
		{
			name:        "different JSON content are not equal",
			value1:      NewJSONWithDefaultsValue(`{"admin":{"cluster":["read"]}}`, testPopulateDefaults),
			value2:      NewJSONWithDefaultsValue(`{"user":{"cluster":["write"]}}`, testPopulateDefaults),
			expected:    false,
			expectError: false,
		},
		{
			name:        "semantic equality with defaults - missing vs explicit false",
			value1:      NewJSONWithDefaultsValue(`{"admin":{"indices":[{"names":["index1"],"privileges":["read"]}]}}`, testPopulateDefaults),
			value2:      NewJSONWithDefaultsValue(`{"admin":{"indices":[{"names":["index1"],"privileges":["read"],"allow_restricted_indices":false}]}}`, testPopulateDefaults),
			expected:    true,
			expectError: false,
		},
		{
			name:        "wrong type returns error",
			value1:      NewJSONWithDefaultsValue(`{"admin":{}}`, testPopulateDefaults),
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

	value := NewJSONWithDefaultsValue(complexJSON, testPopulateDefaults)
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
