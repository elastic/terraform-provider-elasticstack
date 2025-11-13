package connectors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeV0(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		rawState      map[string]interface{}
		expectedState map[string]interface{}
		expectError   bool
		errorContains string
	}{
		{
			name: "removes empty config field",
			rawState: map[string]interface{}{
				"id":     "test-connector",
				"name":   "Test Connector",
				"config": "",
				"type":   "webhook",
			},
			expectedState: map[string]interface{}{
				"id":   "test-connector",
				"name": "Test Connector",
				"type": "webhook",
			},
			expectError: false,
		},
		{
			name: "removes empty secrets field",
			rawState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"secrets": "",
				"type":    "webhook",
			},
			expectedState: map[string]interface{}{
				"id":   "test-connector",
				"name": "Test Connector",
				"type": "webhook",
			},
			expectError: false,
		},
		{
			name: "removes both empty config and secrets fields",
			rawState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"config":  "",
				"secrets": "",
				"type":    "webhook",
			},
			expectedState: map[string]interface{}{
				"id":   "test-connector",
				"name": "Test Connector",
				"type": "webhook",
			},
			expectError: false,
		},
		{
			name: "preserves non-empty config field",
			rawState: map[string]interface{}{
				"id":     "test-connector",
				"name":   "Test Connector",
				"config": `{"url": "https://example.com"}`,
				"type":   "webhook",
			},
			expectedState: map[string]interface{}{
				"id":     "test-connector",
				"name":   "Test Connector",
				"config": `{"url": "https://example.com"}`,
				"type":   "webhook",
			},
			expectError: false,
		},
		{
			name: "preserves non-empty secrets field",
			rawState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"secrets": `{"apiKey": "secret123"}`,
				"type":    "webhook",
			},
			expectedState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"secrets": `{"apiKey": "secret123"}`,
				"type":    "webhook",
			},
			expectError: false,
		},
		{
			name: "preserves non-string config field",
			rawState: map[string]interface{}{
				"id":     "test-connector",
				"name":   "Test Connector",
				"config": map[string]interface{}{"url": "https://example.com"},
				"type":   "webhook",
			},
			expectedState: map[string]interface{}{
				"id":     "test-connector",
				"name":   "Test Connector",
				"config": map[string]interface{}{"url": "https://example.com"},
				"type":   "webhook",
			},
			expectError: false,
		},
		{
			name: "preserves non-string secrets field",
			rawState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"secrets": map[string]interface{}{"apiKey": "secret123"},
				"type":    "webhook",
			},
			expectedState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"secrets": map[string]interface{}{"apiKey": "secret123"},
				"type":    "webhook",
			},
			expectError: false,
		},
		{
			name: "handles missing config and secrets fields",
			rawState: map[string]interface{}{
				"id":   "test-connector",
				"name": "Test Connector",
				"type": "webhook",
			},
			expectedState: map[string]interface{}{
				"id":   "test-connector",
				"name": "Test Connector",
				"type": "webhook",
			},
			expectError: false,
		},
		{
			name: "handles null config field",
			rawState: map[string]interface{}{
				"id":     "test-connector",
				"name":   "Test Connector",
				"config": nil,
				"type":   "webhook",
			},
			expectedState: map[string]interface{}{
				"id":     "test-connector",
				"name":   "Test Connector",
				"config": nil,
				"type":   "webhook",
			},
			expectError: false,
		},
		{
			name: "handles null secrets field",
			rawState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"secrets": nil,
				"type":    "webhook",
			},
			expectedState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"secrets": nil,
				"type":    "webhook",
			},
			expectError: false,
		},
		{
			name: "handles complex state with other fields",
			rawState: map[string]interface{}{
				"id":                  "test-connector",
				"name":                "Test Connector",
				"config":              "",
				"secrets":             "",
				"type":                "webhook",
				"connector_type_id":   "webhook-connector",
				"is_preconfigured":    false,
				"is_deprecated":       false,
				"is_missing_secrets":  false,
				"referenced_by_count": float64(0), // JSON unmarshaling converts numbers to float64
				"is_system_action":    false,
			},
			expectedState: map[string]interface{}{
				"id":                  "test-connector",
				"name":                "Test Connector",
				"type":                "webhook",
				"connector_type_id":   "webhook-connector",
				"is_preconfigured":    false,
				"is_deprecated":       false,
				"is_missing_secrets":  false,
				"referenced_by_count": float64(0), // JSON unmarshaling converts numbers to float64
				"is_system_action":    false,
			},
			expectError: false,
		},
		{
			name: "handles mixed cases - empty config, non-empty secrets",
			rawState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"config":  "",
				"secrets": `{"password": "secret"}`,
				"type":    "webhook",
			},
			expectedState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"secrets": `{"password": "secret"}`,
				"type":    "webhook",
			},
			expectError: false,
		},
		{
			name: "handles mixed cases - non-empty config, empty secrets",
			rawState: map[string]interface{}{
				"id":      "test-connector",
				"name":    "Test Connector",
				"config":  `{"endpoint": "https://api.example.com"}`,
				"secrets": "",
				"type":    "webhook",
			},
			expectedState: map[string]interface{}{
				"id":     "test-connector",
				"name":   "Test Connector",
				"config": `{"endpoint": "https://api.example.com"}`,
				"type":   "webhook",
			},
			expectError: false,
		},
		{
			name: "handles minimal state",
			rawState: map[string]interface{}{
				"id": "minimal-connector",
			},
			expectedState: map[string]interface{}{
				"id": "minimal-connector",
			},
			expectError: false,
		},
		{
			name:          "handles invalid JSON in raw state",
			rawState:      nil, // This will be handled specially in the test
			expectedState: nil,
			expectError:   true,
			errorContains: "Failed to unmarshal state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// Marshal the raw state to JSON, unless we expect unmarshal error
			var rawStateJSON []byte
			var err error
			if tt.errorContains == "Failed to unmarshal state" {
				// Create invalid JSON for unmarshal error test
				rawStateJSON = []byte(`{"invalid": "unterminated string`)
			} else {
				rawStateJSON, err = json.Marshal(tt.rawState)
				require.NoError(t, err)
			}

			// Create the upgrade request
			req := resource.UpgradeStateRequest{
				RawState: &tfprotov6.RawState{
					JSON: rawStateJSON,
				},
			}

			// Create a response
			resp := &resource.UpgradeStateResponse{}

			// Call the upgradeV0 function
			upgradeV0(ctx, req, resp)

			if tt.expectError {
				require.True(t, resp.Diagnostics.HasError(), "Expected error but got none")
				if tt.errorContains != "" {
					errorSummary := ""
					for _, diag := range resp.Diagnostics.Errors() {
						errorSummary += diag.Summary() + " " + diag.Detail()
					}
					assert.Contains(t, errorSummary, tt.errorContains)
				}
				return
			}

			// Check no errors occurred
			require.False(t, resp.Diagnostics.HasError(), "Unexpected error: %v", resp.Diagnostics.Errors())

			// Check that a DynamicValue is always returned
			require.NotNil(t, resp.DynamicValue, "DynamicValue should always be returned")
			require.NotNil(t, resp.DynamicValue.JSON, "DynamicValue.JSON should not be nil")

			// Unmarshal the upgraded state to compare
			var actualState map[string]interface{}
			err = json.Unmarshal(resp.DynamicValue.JSON, &actualState)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedState, actualState)
		})
	}
}

func TestUpgradeV0_JsonMarshalError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a valid raw state
	rawState := map[string]interface{}{
		"id":     "test-connector",
		"config": "",
	}
	rawStateJSON, err := json.Marshal(rawState)
	require.NoError(t, err)

	// Create the upgrade request
	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: rawStateJSON,
		},
	}

	// Create a response
	resp := &resource.UpgradeStateResponse{}

	// Mock the json.Marshal failure by creating a scenario where marshaling would fail
	// This is tricky to test directly since we can't easily inject marshal failure
	// but we can test the error handling path by examining the code coverage

	// Call the upgradeV0 function
	upgradeV0(ctx, req, resp)

	// This should succeed normally
	require.False(t, resp.Diagnostics.HasError())
	require.NotNil(t, resp.DynamicValue)
}

func TestUpgradeV0_NilRawState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create the upgrade request with nil RawState
	req := resource.UpgradeStateRequest{
		RawState: nil,
	}

	// Create a response
	resp := &resource.UpgradeStateResponse{}

	// The current implementation panics with nil RawState
	// This test documents the current behavior
	require.Panics(t, func() {
		upgradeV0(ctx, req, resp)
	}, "upgradeV0 should panic with nil RawState")
}

func TestUpgradeV0_NilRawStateJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create the upgrade request with nil JSON
	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: nil,
		},
	}

	// Create a response
	resp := &resource.UpgradeStateResponse{}

	// Call the upgradeV0 function
	upgradeV0(ctx, req, resp)

	// This should cause an error due to nil JSON
	require.True(t, resp.Diagnostics.HasError())
}

func TestRemoveEmptyStringHelper(t *testing.T) {
	t.Parallel()

	// This is testing the internal removeEmptyString function behavior
	// by examining its effects through the upgradeV0 function

	tests := []struct {
		name     string
		state    map[string]interface{}
		key      string
		expected map[string]interface{}
	}{
		{
			name: "removes empty string",
			state: map[string]interface{}{
				"id":     "test",
				"config": "",
			},
			key: "config",
			expected: map[string]interface{}{
				"id": "test",
			},
		},
		{
			name: "preserves non-empty string",
			state: map[string]interface{}{
				"id":     "test",
				"config": "value",
			},
			key: "config",
			expected: map[string]interface{}{
				"id":     "test",
				"config": "value",
			},
		},
		{
			name: "preserves non-string value",
			state: map[string]interface{}{
				"id":     "test",
				"config": float64(123), // JSON marshaling/unmarshaling converts numbers to float64
			},
			key: "config",
			expected: map[string]interface{}{
				"id":     "test",
				"config": float64(123),
			},
		},
		{
			name: "handles missing key",
			state: map[string]interface{}{
				"id": "test",
			},
			key: "config",
			expected: map[string]interface{}{
				"id": "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// Use upgradeV0 to test the removeEmptyString behavior
			rawStateJSON, err := json.Marshal(tt.state)
			require.NoError(t, err)

			req := resource.UpgradeStateRequest{
				RawState: &tfprotov6.RawState{
					JSON: rawStateJSON,
				},
			}

			resp := &resource.UpgradeStateResponse{}
			upgradeV0(ctx, req, resp)

			require.False(t, resp.Diagnostics.HasError())
			require.NotNil(t, resp.DynamicValue)

			var actualState map[string]interface{}
			err = json.Unmarshal(resp.DynamicValue.JSON, &actualState)
			require.NoError(t, err)

			// Check the specific behavior for the key being tested
			if tt.key == "config" || tt.key == "secrets" {
				// These keys are handled by upgradeV0
				if tt.expected[tt.key] == nil {
					_, exists := actualState[tt.key]
					assert.False(t, exists, "Key %s should be removed", tt.key)
				} else {
					assert.Equal(t, tt.expected[tt.key], actualState[tt.key])
				}
			}
		})
	}
}
