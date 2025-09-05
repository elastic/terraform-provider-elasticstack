package connectors

import (
	"context"
	"testing"

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
		expectError         bool
	}{
		{
			name:                "valid JSON config with connector type ID",
			input:               basetypes.NewStringValue(`{"key": "value", "__tf_provider_connector_type_id": "my-connector"}`),
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
			input:               basetypes.NewStringValue(`{"key": "value", "__tf_provider_connector_type_id": 123}`),
			expectedConnectorID: "",
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configType := ConfigType{}
			result, diags := configType.ValueFromString(context.Background(), tt.input)

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected an error but got none")
				return
			}

			require.False(t, diags.HasError(), "Unexpected error: %v", diags)
			require.NotNil(t, result, "Result should not be nil")

			configValue, ok := result.(ConfigValue)
			require.True(t, ok, "Result should be of type ConfigValue")

			require.Equal(t, tt.expectedConnectorID, configValue.connectorTypeID, "Connector type ID mismatch")
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
			tfValue: tftypes.NewValue(tftypes.String, `{"key": "value", "__tf_provider_connector_type_id": "test-connector"}`),
			expectedValue: ConfigValue{
				Normalized: func() jsontypes.Normalized {
					return jsontypes.NewNormalizedValue(`{"key": "value", "__tf_provider_connector_type_id": "test-connector"}`)
				}(),
				connectorTypeID: "test-connector",
			},
		},
		{
			name:    "valid string value with empty JSON",
			tfValue: tftypes.NewValue(tftypes.String, `{}`),
			expectedValue: ConfigValue{
				Normalized: func() jsontypes.Normalized {
					n, _ := jsontypes.NewNormalizedValue(`{}`).ToStringValue(context.Background())
					return jsontypes.Normalized{StringValue: n}
				}(),
				connectorTypeID: "",
			},
		},
		{
			name:    "null string value",
			tfValue: tftypes.NewValue(tftypes.String, nil),
			expectedValue: ConfigValue{
				Normalized: func() jsontypes.Normalized {
					return jsontypes.Normalized{StringValue: basetypes.NewStringNull()}
				}(),
				connectorTypeID: "",
			},
		},
		{
			name:    "unknown string value",
			tfValue: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectedValue: ConfigValue{
				Normalized: func() jsontypes.Normalized {
					return jsontypes.Normalized{StringValue: basetypes.NewStringUnknown()}
				}(),
				connectorTypeID: "",
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

			// Compare the connector type ID
			require.Equal(t, expectedConfigValue.connectorTypeID, configValue.connectorTypeID, "Connector type ID mismatch")

			// Compare the underlying string values
			require.Equal(t, expectedConfigValue.StringValue.Equal(configValue.StringValue), true, "String values should be equal")
		})
	}
}
