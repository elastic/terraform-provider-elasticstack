package connectors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestConfigType_ValueFromString(t *testing.T) {
	tests := []struct {
		name                string
		input               basetypes.StringValue
		expectedConnectorID string
		contextKey          string
		expectError         bool
	}{
		{
			name:                "valid JSON config with connector type ID",
			input:               basetypes.NewStringValue(`{"key": "value", "__tf_provider_connector_type_id": "my-connector"}`),
			expectedConnectorID: "my-connector",
			contextKey:          "__tf_provider_connector_type_id",
			expectError:         false,
		},
		{
			name:                "valid JSON config with context key ID",
			input:               basetypes.NewStringValue(`{"key": "value", "__tf_provider_context": "my-connector"}`),
			expectedConnectorID: "my-connector",
			expectError:         false,
		},
		{
			name:                "valid JSON config without connector type ID",
			input:               basetypes.NewStringValue(`{"key": "value"}`),
			expectedConnectorID: "",
			expectError:         false,
		},
		{
			name:                "empty JSON config",
			input:               basetypes.NewStringValue(`{}`),
			expectedConnectorID: "",
			expectError:         false,
		},
		{
			name:        "invalid JSON config",
			input:       basetypes.NewStringValue(`{invalid json`),
			expectError: true,
		},
		{
			name:                "null string value",
			input:               basetypes.NewStringNull(),
			expectedConnectorID: "",
			expectError:         false,
		},
		{
			name:                "unknown string value",
			input:               basetypes.NewStringUnknown(),
			expectedConnectorID: "",
			expectError:         false,
		},
		{
			name:                "JSON with non-string connector type ID",
			input:               basetypes.NewStringValue(`{"key": "value", "__tf_provider_context": 123}`),
			expectedConnectorID: "",
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configType := ConfigType{}
			result, diags := configType.ValueFromString(t.Context(), tt.input)

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected an error but got none")
				return
			}

			require.False(t, diags.HasError(), "Unexpected error: %v", diags)
			require.NotNil(t, result, "Result should not be nil")

			configValue, ok := result.(ConfigValue)
			require.True(t, ok, "Result should be of type ConfigValue")

			if !configValue.IsNull() && !configValue.IsUnknown() {
				var resultMap map[string]interface{}
				err := json.Unmarshal([]byte(configValue.ValueString()), &resultMap)
				require.NoError(t, err)

				if tt.expectedConnectorID != "" {
					contextKey := "__tf_provider_context"
					if tt.contextKey != "" {
						contextKey = tt.contextKey
					}
					require.Equal(t, tt.expectedConnectorID, resultMap[contextKey], "Connector type ID mismatch")
				}
			}

			require.Equal(t, tt.input, configValue.StringValue, "String value should be preserved")
		})
	}
}

func TestConfigType_ValueFromTerraform(t *testing.T) {
	tests := []struct {
		name          string
		tfValue       tftypes.Value
		expectedValue attr.Value
		expectedError string
	}{
		{
			name:    "valid string value with JSON config",
			tfValue: tftypes.NewValue(tftypes.String, `{"key": "value", "__tf_provider_context": "test-connector"}`),
			expectedValue: ConfigValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: func() jsontypes.Normalized {
						return jsontypes.NewNormalizedValue(`{"key": "value", "__tf_provider_context": "test-connector"}`)
					}(),
				},
			},
		},
		{
			name:    "valid string value with empty JSON",
			tfValue: tftypes.NewValue(tftypes.String, `{}`),
			expectedValue: ConfigValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: func() jsontypes.Normalized {
						n, _ := jsontypes.NewNormalizedValue(`{}`).ToStringValue(context.Background())
						return jsontypes.Normalized{StringValue: n}
					}(),
				},
			},
		},
		{
			name:    "null string value",
			tfValue: tftypes.NewValue(tftypes.String, nil),
			expectedValue: ConfigValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: func() jsontypes.Normalized {
						return jsontypes.Normalized{StringValue: basetypes.NewStringNull()}
					}(),
				},
			},
		},
		{
			name:    "unknown string value",
			tfValue: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectedValue: ConfigValue{
				JSONWithContextualDefaultsValue: customtypes.JSONWithContextualDefaultsValue{
					Normalized: func() jsontypes.Normalized {
						return jsontypes.Normalized{StringValue: basetypes.NewStringUnknown()}
					}(),
				},
			},
		},
		{
			name:          "non-string terraform value",
			tfValue:       tftypes.NewValue(tftypes.Bool, true),
			expectedValue: nil,
			expectedError: "expected string",
		},
		{
			name:          "invalid JSON in string value",
			tfValue:       tftypes.NewValue(tftypes.String, `{invalid json`),
			expectedValue: nil,
			expectedError: "unexpected error converting StringValue to StringValuable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configType := ConfigType{}
			result, err := configType.ValueFromTerraform(context.Background(), tt.tfValue)

			if tt.expectedError != "" {
				require.Error(t, err, "Expected an error but got none")
				require.Contains(t, err.Error(), tt.expectedError, "Error message should contain expected text")
				require.Nil(t, result, "Result should be nil when there's an error")
				return
			}

			require.NoError(t, err, "Unexpected error: %v", err)
			require.NotNil(t, result, "Result should not be nil")

			configValue, ok := result.(ConfigValue)
			require.True(t, ok, "Result should be of type ConfigValue")

			expectedConfigValue, ok := tt.expectedValue.(ConfigValue)
			require.True(t, ok, "Expected value should be of type ConfigValue")

			// Compare the underlying string values
			require.Equal(t, expectedConfigValue.StringValue.Equal(configValue.StringValue), true, "String values should be equal")
		})
	}
}
