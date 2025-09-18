package output

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputResourceUpgradeState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		rawState      map[string]interface{}
		expectedState map[string]interface{}
		expectError   bool
		errorContains string
	}{
		{
			name: "successful upgrade - ssl list to object",
			rawState: map[string]interface{}{
				"id":   "test-output",
				"name": "Test Output",
				"type": "elasticsearch",
				"ssl": []interface{}{
					map[string]interface{}{
						"certificate":             "cert-content",
						"key":                     "key-content",
						"certificate_authorities": []interface{}{"ca1", "ca2"},
					},
				},
				"hosts": []interface{}{"https://localhost:9200"},
			},
			expectedState: map[string]interface{}{
				"id":   "test-output",
				"name": "Test Output",
				"type": "elasticsearch",
				"ssl": map[string]interface{}{
					"certificate":             "cert-content",
					"key":                     "key-content",
					"certificate_authorities": []interface{}{"ca1", "ca2"},
				},
				"hosts": []interface{}{"https://localhost:9200"},
			},
			expectError: false,
		},
		{
			name: "no ssl field - no changes",
			rawState: map[string]interface{}{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"hosts": []interface{}{"https://localhost:9200"},
			},
			expectedState: map[string]interface{}{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"hosts": []interface{}{"https://localhost:9200"},
			},
			expectError: false,
		},
		{
			name: "empty ssl list - removes ssl field",
			rawState: map[string]interface{}{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"ssl":   []interface{}{},
				"hosts": []interface{}{"https://localhost:9200"},
			},
			expectedState: map[string]interface{}{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"hosts": []interface{}{"https://localhost:9200"},
			},
			expectError: false,
		},
		{
			name: "ssl not an array - returns error",
			rawState: map[string]interface{}{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"ssl":   "invalid-type",
				"hosts": []interface{}{"https://localhost:9200"},
			},
			expectedState: nil,
			expectError:   true,
			errorContains: "Unexpected type for legacy ssl attribute",
		},
		{
			name: "multiple ssl items - takes first item",
			rawState: map[string]interface{}{
				"id":   "test-output",
				"name": "Test Output",
				"type": "elasticsearch",
				"ssl": []interface{}{
					map[string]interface{}{"certificate": "cert1"},
					map[string]interface{}{"certificate": "cert2"},
				},
				"hosts": []interface{}{"https://localhost:9200"},
			},
			expectedState: map[string]interface{}{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"ssl":   map[string]interface{}{"certificate": "cert1"},
				"hosts": []interface{}{"https://localhost:9200"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Marshal the raw state to JSON
			rawStateJSON, err := json.Marshal(tt.rawState)
			require.NoError(t, err)

			// Create the upgrade request
			req := resource.UpgradeStateRequest{
				RawState: &tfprotov6.RawState{
					JSON: rawStateJSON,
				},
			}

			// Create a response
			resp := &resource.UpgradeStateResponse{}

			// Create the resource and call UpgradeState
			r := &outputResource{}
			upgraders := r.UpgradeState(context.Background())
			upgrader := upgraders[0]
			upgrader.StateUpgrader(context.Background(), req, resp)

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

			// Unmarshal the upgraded state to compare
			var actualState map[string]interface{}
			err = json.Unmarshal(resp.DynamicValue.JSON, &actualState)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedState, actualState)
		})
	}
}
