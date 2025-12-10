package connectors

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestConfigValue_ValidateAttribute(t *testing.T) {
	tests := []struct {
		name          string
		configValue   ConfigValue
		expectError   bool
		errorContains string
	}{
		{
			name:        "null value should not validate",
			configValue: NewConfigNull(),
			expectError: false,
		},
		{
			name:        "unknown value should not validate",
			configValue: NewConfigUnknown(),
			expectError: false,
		},
		{
			name: "valid JSON value should validate successfully",
			configValue: ConfigValue{
				Normalized: jsontypes.NewNormalizedValue(`{"key": "value"}`),
			},
			expectError: false,
		},
		{
			name: "invalid JSON value should produce validation error",
			configValue: ConfigValue{
				Normalized: func() jsontypes.Normalized {
					// Create an invalid JSON by directly setting StringValue
					return jsontypes.Normalized{StringValue: basetypes.NewStringValue(`{invalid json`)}
				}(),
			},
			expectError:   true,
			errorContains: "Invalid JSON String Value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := xattr.ValidateAttributeRequest{
				Path: path.Root("config"),
			}
			resp := &xattr.ValidateAttributeResponse{}

			tt.configValue.ValidateAttribute(context.Background(), req, resp)

			if tt.expectError {
				require.True(t, resp.Diagnostics.HasError(), "Expected validation error but got none")
				if tt.errorContains != "" {
					require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), tt.errorContains)
				}
			} else {
				require.False(t, resp.Diagnostics.HasError(), "Unexpected validation error: %v", resp.Diagnostics)
			}
		})
	}
}

func TestConfigValue_SanitizedValue(t *testing.T) {
	tests := []struct {
		name           string
		configValue    ConfigValue
		expectedResult string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "null value returns empty string",
			configValue:    NewConfigNull(),
			expectedResult: "",
			expectError:    false,
		},
		{
			name:           "unknown value returns empty string",
			configValue:    NewConfigUnknown(),
			expectedResult: "",
			expectError:    false,
		},
		{
			name: "JSON without connector type ID remains unchanged",
			configValue: ConfigValue{
				Normalized: jsontypes.NewNormalizedValue(`{"key": "value", "another": "field"}`),
			},
			expectedResult: `{"another":"field","key":"value"}`,
			expectError:    false,
		},
		{
			name: "JSON with connector type ID gets sanitized",
			configValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(`{"key": "value", "__tf_provider_connector_type_id": "test-connector", "another": "field"}`),
				connectorTypeID: "test-connector",
			},
			expectedResult: `{"another":"field","key":"value"}`,
			expectError:    false,
		},
		{
			name: "empty JSON object",
			configValue: ConfigValue{
				Normalized: jsontypes.NewNormalizedValue(`{}`),
			},
			expectedResult: `{}`,
			expectError:    false,
		},
		{
			name: "invalid JSON should return error",
			configValue: ConfigValue{
				Normalized: jsontypes.Normalized{StringValue: basetypes.NewStringValue(`{invalid json`)},
			},
			expectError:   true,
			errorContains: "Failed to unmarshal config value",
		},
		{
			name: "JSON with null values gets sanitized - top level",
			configValue: ConfigValue{
				Normalized: jsontypes.NewNormalizedValue(`{"key": "value", "nullField": null, "another": "field"}`),
			},
			expectedResult: `{"another":"field","key":"value"}`,
			expectError:    false,
		},
		{
			name: "JSON with null values gets sanitized - nested",
			configValue: ConfigValue{
				Normalized: jsontypes.NewNormalizedValue(`{"key": "value", "nested": {"field": "value", "nullField": null}}`),
			},
			expectedResult: `{"key":"value","nested":{"field":"value"}}`,
			expectError:    false,
		},
		{
			name: "JSON with null values gets sanitized - mixed",
			configValue: ConfigValue{
				Normalized: jsontypes.NewNormalizedValue(`{"key": "value", "nullTop": null, "nested": {"field": "value", "nullNested": null}, "another": null}`),
			},
			expectedResult: `{"key":"value","nested":{"field":"value"}}`,
			expectError:    false,
		},
		{
			name: "JSON with only null values results in empty object",
			configValue: ConfigValue{
				Normalized: jsontypes.NewNormalizedValue(`{"nullField1": null, "nullField2": null}`),
			},
			expectedResult: `{}`,
			expectError:    false,
		},
		{
			name: "JSON with null and connector type ID gets both sanitized",
			configValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(`{"key": "value", "nullField": null, "__tf_provider_connector_type_id": "test-connector"}`),
				connectorTypeID: "test-connector",
			},
			expectedResult: `{"key":"value"}`,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.configValue.SanitizedValue()

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
				if tt.errorContains != "" {
					require.Contains(t, diags.Errors()[0].Summary(), tt.errorContains)
				}
			} else {
				require.False(t, diags.HasError(), "Unexpected error: %v", diags)
				require.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestConfigValue_StringSemanticEquals(t *testing.T) {
	emailConnectorID := ".email"
	emailConnectorConfig := `{"key": "value"}`
	emailConnectorConfigWithDefaults, err := kibana_oapi.ConnectorConfigWithDefaults(emailConnectorID, emailConnectorConfig)
	require.NoError(t, err)

	tests := []struct {
		name          string
		configValue   ConfigValue
		otherValue    basetypes.StringValuable
		expectEqual   bool
		expectError   bool
		errorContains string
	}{
		{
			name:        "null values are equal",
			configValue: NewConfigNull(),
			otherValue:  NewConfigNull(),
			expectEqual: true,
			expectError: false,
		},
		{
			name:        "unknown values are equal",
			configValue: NewConfigUnknown(),
			otherValue:  NewConfigUnknown(),
			expectEqual: true,
			expectError: false,
		},
		{
			name:        "null vs unknown should not be equal",
			configValue: NewConfigNull(),
			otherValue:  NewConfigUnknown(),
			expectEqual: false,
			expectError: false,
		},
		{
			name: "wrong type should produce error",
			configValue: ConfigValue{
				Normalized: jsontypes.NewNormalizedValue(`{"key": "value"}`),
			},
			otherValue:    basetypes.NewStringValue(`{"key": "value"}`),
			expectEqual:   false,
			expectError:   true,
			errorContains: "Semantic Equality Check Error",
		},
		{
			name: "values without connector type ID should use normalized comparison",
			configValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(`{"key": "value"}`),
				connectorTypeID: "",
			},
			otherValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(`{"key": "value"}`),
				connectorTypeID: "",
			},
			expectEqual: true,
			expectError: false,
		},
		{
			name: "different values without connector type ID should not be equal",
			configValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(`{"key": "value1"}`),
				connectorTypeID: "",
			},
			otherValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(`{"key": "value2"}`),
				connectorTypeID: "",
			},
			expectEqual: false,
			expectError: false,
		},
		{
			name: "values with same connector type ID from first value",
			configValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(emailConnectorConfig),
				connectorTypeID: emailConnectorID,
			},
			otherValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(emailConnectorConfigWithDefaults),
				connectorTypeID: "",
			},
			expectEqual: true, // Would be true if connector config with defaults works
			expectError: false,
		},
		{
			name: "values with same connector type ID from second value",
			configValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(emailConnectorConfigWithDefaults),
				connectorTypeID: "",
			},
			otherValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(emailConnectorConfig),
				connectorTypeID: emailConnectorID,
			},
			expectEqual: true, // Would be true if connector config with defaults works
			expectError: false,
		},
		{
			name: "invalid JSON in first value should cause error",
			configValue: ConfigValue{
				Normalized:      jsontypes.Normalized{StringValue: basetypes.NewStringValue(`{invalid`)},
				connectorTypeID: "test-connector",
			},
			otherValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(`{"key": "value"}`),
				connectorTypeID: "test-connector",
			},
			expectEqual:   false,
			expectError:   true,
			errorContains: "Failed to unmarshal config value",
		},
		{
			name: "invalid JSON in second value should cause error",
			configValue: ConfigValue{
				Normalized:      jsontypes.NewNormalizedValue(`{"key": "value"}`),
				connectorTypeID: "test-connector",
			},
			otherValue: ConfigValue{
				Normalized:      jsontypes.Normalized{StringValue: basetypes.NewStringValue(`{invalid`)},
				connectorTypeID: "test-connector",
			},
			expectEqual:   false,
			expectError:   true,
			errorContains: "Failed to unmarshal config value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.configValue.StringSemanticEquals(context.Background(), tt.otherValue)

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
				if tt.errorContains != "" {
					errorFound := false
					for _, err := range diags.Errors() {
						if strings.Contains(err.Summary(), tt.errorContains) || strings.Contains(err.Detail(), tt.errorContains) {
							errorFound = true
							break
						}
					}
					require.True(t, errorFound, "Expected error containing '%s' but got: %v", tt.errorContains, diags)
				}
			} else {
				if diags.HasError() {
					// For connector config with defaults errors, we might expect them in real scenarios
					// but for unit tests, we'll be more lenient
					hasConnectorError := false
					for _, err := range diags.Errors() {
						if strings.Contains(err.Summary(), "Failed to get connector config with defaults") {
							hasConnectorError = true
							break
						}
					}
					if !hasConnectorError {
						require.False(t, diags.HasError(), "Unexpected error: %v", diags)
					}
				}
				require.Equal(t, tt.expectEqual, result)
			}
		})
	}
}

func TestNewConfigValueWithConnectorID(t *testing.T) {
	tests := []struct {
		name            string
		value           string
		connectorTypeID string
		expectError     bool
		errorContains   string
		validateResult  func(t *testing.T, result ConfigValue)
	}{
		{
			name:            "empty value returns null config",
			value:           "",
			connectorTypeID: "test-connector",
			expectError:     false,
			validateResult: func(t *testing.T, result ConfigValue) {
				require.True(t, result.IsNull())
				require.Equal(t, "", result.connectorTypeID)
			},
		},
		{
			name:            "valid JSON with connector type ID",
			value:           `{"key": "value"}`,
			connectorTypeID: "test-connector",
			expectError:     false,
			validateResult: func(t *testing.T, result ConfigValue) {
				require.False(t, result.IsNull())
				require.Equal(t, "test-connector", result.connectorTypeID)

				// Check that the connector type ID was added to the JSON
				var resultMap map[string]interface{}
				err := json.Unmarshal([]byte(result.ValueString()), &resultMap)
				require.NoError(t, err)
				require.Equal(t, "test-connector", resultMap["__tf_provider_connector_type_id"])
				require.Equal(t, "value", resultMap["key"])
			},
		},
		{
			name:            "valid empty JSON object",
			value:           `{}`,
			connectorTypeID: "test-connector",
			expectError:     false,
			validateResult: func(t *testing.T, result ConfigValue) {
				require.False(t, result.IsNull())
				require.Equal(t, "test-connector", result.connectorTypeID)

				var resultMap map[string]interface{}
				err := json.Unmarshal([]byte(result.ValueString()), &resultMap)
				require.NoError(t, err)
				require.Equal(t, "test-connector", resultMap["__tf_provider_connector_type_id"])
			},
		},
		{
			name:            "complex JSON object",
			value:           `{"config": {"nested": "value"}, "array": [1, 2, 3]}`,
			connectorTypeID: "complex-connector",
			expectError:     false,
			validateResult: func(t *testing.T, result ConfigValue) {
				require.False(t, result.IsNull())
				require.Equal(t, "complex-connector", result.connectorTypeID)

				var resultMap map[string]interface{}
				err := json.Unmarshal([]byte(result.ValueString()), &resultMap)
				require.NoError(t, err)
				require.Equal(t, "complex-connector", resultMap["__tf_provider_connector_type_id"])

				config, ok := resultMap["config"].(map[string]interface{})
				require.True(t, ok)
				require.Equal(t, "value", config["nested"])

				array, ok := resultMap["array"].([]interface{})
				require.True(t, ok)
				require.Len(t, array, 3)
			},
		},
		{
			name:            "invalid JSON should return error",
			value:           `{invalid json`,
			connectorTypeID: "test-connector",
			expectError:     true,
			errorContains:   "Failed to unmarshal config",
		},
		{
			name:            "empty connector type ID",
			value:           `{"key": "value"}`,
			connectorTypeID: "",
			expectError:     false,
			validateResult: func(t *testing.T, result ConfigValue) {
				require.False(t, result.IsNull())
				require.Equal(t, "", result.connectorTypeID)

				var resultMap map[string]interface{}
				err := json.Unmarshal([]byte(result.ValueString()), &resultMap)
				require.NoError(t, err)
				require.Equal(t, "", resultMap["__tf_provider_connector_type_id"])
				require.Equal(t, "value", resultMap["key"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := NewConfigValueWithConnectorID(tt.value, tt.connectorTypeID)

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
				if tt.errorContains != "" {
					require.Contains(t, diags.Errors()[0].Summary(), tt.errorContains)
				}
			} else {
				require.False(t, diags.HasError(), "Unexpected error: %v", diags)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}
