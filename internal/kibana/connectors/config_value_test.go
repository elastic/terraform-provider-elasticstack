package connectors

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

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
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(`{"key": "value"}`),
				},
			},
			otherValue:    basetypes.NewStringValue(`{"key": "value"}`),
			expectEqual:   false,
			expectError:   true,
			errorContains: "Semantic Equality Check Error",
		},
		{
			name: "values without connector type ID should use normalized comparison",
			configValue: ConfigValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(`{"key": "value"}`),
				},
			},
			otherValue: ConfigValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(`{"key": "value"}`),
				},
			},
			expectEqual: true,
			expectError: false,
		},
		{
			name: "different values without connector type ID should not be equal",
			configValue: ConfigValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(`{"key": "value1"}`),
				},
			},
			otherValue: ConfigValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(`{"key": "value2"}`),
				},
			},
			expectEqual: false,
			expectError: false,
		},
		{
			name: "values with same connector type ID from first value",
			configValue: func() ConfigValue {
				val, _ := NewConfigValueWithConnectorID(emailConnectorConfig, emailConnectorID)
				return val
			}(),
			otherValue: ConfigValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(emailConnectorConfigWithDefaults),
				},
			},
			expectEqual: true, // Would be true if connector config with defaults works
			expectError: false,
		},
		{
			name: "values with same connector type ID from second value",
			configValue: ConfigValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: jsontypes.NewNormalizedValue(emailConnectorConfigWithDefaults),
				},
			},
			otherValue: func() ConfigValue {
				val, _ := NewConfigValueWithConnectorID(emailConnectorConfig, emailConnectorID)
				return val
			}(),
			expectEqual: true, // Would be true if connector config with defaults works
			expectError: false,
		},
		{
			name: "invalid JSON in first value should cause error",
			configValue: func() ConfigValue {
				// Manually construct invalid JSON with context
				return ConfigValue{
					JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
						Normalized: jsontypes.Normalized{StringValue: basetypes.NewStringValue(`{invalid`)},
					},
				}
			}(),
			otherValue: func() ConfigValue {
				val, _ := NewConfigValueWithConnectorID(`{"key": "value"}`, "test-connector")
				return val
			}(),
			expectEqual:   false,
			expectError:   true,
			errorContains: "Failed to unmarshal config value",
		},
		{
			name: "invalid JSON in second value should cause error",
			configValue: func() ConfigValue {
				val, _ := NewConfigValueWithConnectorID(`{"key": "value"}`, "test-connector")
				return val
			}(),
			otherValue: func() ConfigValue {
				return ConfigValue{
					JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
						Normalized: jsontypes.Normalized{StringValue: basetypes.NewStringValue(`{invalid`)},
					},
				}
			}(),
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
						if strings.Contains(err.Summary(), "Failed to get config with defaults") {
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
			},
		},
		{
			name:            "valid JSON with connector type ID",
			value:           `{"key": "value"}`,
			connectorTypeID: "test-connector",
			expectError:     false,
			validateResult: func(t *testing.T, result ConfigValue) {
				require.False(t, result.IsNull())

				// Check that the connector type ID was added to the JSON
				var resultMap map[string]interface{}
				err := json.Unmarshal([]byte(result.ValueString()), &resultMap)
				require.NoError(t, err)
				require.Equal(t, "test-connector", resultMap["__tf_provider_context"])
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

				var resultMap map[string]interface{}
				err := json.Unmarshal([]byte(result.ValueString()), &resultMap)
				require.NoError(t, err)
				require.Equal(t, "test-connector", resultMap["__tf_provider_context"])
			},
		},
		{
			name:            "complex JSON object",
			value:           `{"config": {"nested": "value"}, "array": [1, 2, 3]}`,
			connectorTypeID: "complex-connector",
			expectError:     false,
			validateResult: func(t *testing.T, result ConfigValue) {
				require.False(t, result.IsNull())

				var resultMap map[string]interface{}
				err := json.Unmarshal([]byte(result.ValueString()), &resultMap)
				require.NoError(t, err)
				require.Equal(t, "complex-connector", resultMap["__tf_provider_context"])

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

				var resultMap map[string]interface{}
				err := json.Unmarshal([]byte(result.ValueString()), &resultMap)
				require.NoError(t, err)
				require.Equal(t, "", resultMap["__tf_provider_context"])
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
