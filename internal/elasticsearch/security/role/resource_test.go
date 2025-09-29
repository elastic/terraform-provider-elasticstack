package role

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV0ToV1(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]interface{}
		expected      map[string]interface{}
		expectError   bool
		errorContains string
	}{
		{
			name: "empty_global_and_metadata_removed",
			input: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"global":      "",
				"metadata":    "",
				"cluster":     []string{"all"},
			},
			expected: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"cluster":     []interface{}{"all"},
			},
		},
		{
			name: "non_empty_global_and_metadata_preserved",
			input: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"global":      `{"profile": {"privileges": ["manage"]}}`,
				"metadata":    `{"version": 1}`,
				"cluster":     []string{"all"},
			},
			expected: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"global":      `{"profile": {"privileges": ["manage"]}}`,
				"metadata":    `{"version": 1}`,
				"cluster":     []interface{}{"all"},
			},
		},
		{
			name: "empty_query_in_indices_removed",
			input: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"indices": []interface{}{
					map[string]interface{}{
						"names":      []string{"index1", "index2"},
						"privileges": []string{"read"},
						"query":      "",
					},
					map[string]interface{}{
						"names":      []string{"index3"},
						"privileges": []string{"write"},
						"query":      `{"match": {"field": "value"}}`,
					},
				},
			},
			expected: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"indices": []interface{}{
					map[string]interface{}{
						"names":      []interface{}{"index1", "index2"},
						"privileges": []interface{}{"read"},
					},
					map[string]interface{}{
						"names":      []interface{}{"index3"},
						"privileges": []interface{}{"write"},
						"query":      `{"match": {"field": "value"}}`,
					},
				},
			},
		},
		{
			name: "empty_query_in_remote_indices_removed",
			input: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"remote_indices": []interface{}{
					map[string]interface{}{
						"clusters":   []string{"cluster1"},
						"names":      []string{"remote-index1"},
						"privileges": []string{"read"},
						"query":      "",
					},
					map[string]interface{}{
						"clusters":   []string{"cluster2"},
						"names":      []string{"remote-index2"},
						"privileges": []string{"write"},
						"query":      `{"term": {"status": "active"}}`,
					},
				},
			},
			expected: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"remote_indices": []interface{}{
					map[string]interface{}{
						"clusters":   []interface{}{"cluster1"},
						"names":      []interface{}{"remote-index1"},
						"privileges": []interface{}{"read"},
					},
					map[string]interface{}{
						"clusters":   []interface{}{"cluster2"},
						"names":      []interface{}{"remote-index2"},
						"privileges": []interface{}{"write"},
						"query":      `{"term": {"status": "active"}}`,
					},
				},
			},
		},
		{
			name: "all_empty_fields_removed_comprehensive",
			input: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"global":      "",
				"metadata":    "",
				"cluster":     []string{"all"},
				"indices": []interface{}{
					map[string]interface{}{
						"names":      []string{"index1"},
						"privileges": []string{"read"},
						"query":      "",
					},
				},
				"remote_indices": []interface{}{
					map[string]interface{}{
						"clusters":   []string{"cluster1"},
						"names":      []string{"remote-index1"},
						"privileges": []string{"read"},
						"query":      "",
					},
				},
			},
			expected: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"cluster":     []interface{}{"all"},
				"indices": []interface{}{
					map[string]interface{}{
						"names":      []interface{}{"index1"},
						"privileges": []interface{}{"read"},
					},
				},
				"remote_indices": []interface{}{
					map[string]interface{}{
						"clusters":   []interface{}{"cluster1"},
						"names":      []interface{}{"remote-index1"},
						"privileges": []interface{}{"read"},
					},
				},
			},
		},
		{
			name: "no_indices_or_remote_indices",
			input: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"global":      "",
				"metadata":    "",
				"cluster":     []string{"all"},
			},
			expected: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"cluster":     []interface{}{"all"},
			},
		},
		{
			name: "indices_not_array",
			input: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"indices":     "not-an-array",
			},
			expected: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"indices":     "not-an-array", // Should be preserved as-is if not an array
			},
		},
		{
			name: "remote_indices_not_array",
			input: map[string]interface{}{
				"name":           "test-role",
				"description":    "Test role",
				"remote_indices": "not-an-array",
			},
			expected: map[string]interface{}{
				"name":           "test-role",
				"description":    "Test role",
				"remote_indices": "not-an-array", // Should be preserved as-is if not an array
			},
		},
		{
			name: "index_item_not_map",
			input: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"indices": []interface{}{
					"not-a-map",
					map[string]interface{}{
						"names":      []string{"index1"},
						"privileges": []string{"read"},
						"query":      "",
					},
				},
			},
			expected: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"indices": []interface{}{
					"not-a-map", // Should be preserved as-is if not a map
					map[string]interface{}{
						"names":      []interface{}{"index1"},
						"privileges": []interface{}{"read"},
					},
				},
			},
		},
		{
			name: "remote_index_item_not_map",
			input: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"remote_indices": []interface{}{
					"not-a-map",
					map[string]interface{}{
						"clusters":   []string{"cluster1"},
						"names":      []string{"remote-index1"},
						"privileges": []string{"read"},
						"query":      "",
					},
				},
			},
			expected: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"remote_indices": []interface{}{
					"not-a-map", // Should be preserved as-is if not a map
					map[string]interface{}{
						"clusters":   []interface{}{"cluster1"},
						"names":      []interface{}{"remote-index1"},
						"privileges": []interface{}{"read"},
					},
				},
			},
		},
		{
			name: "nil_global_and_metadata_preserved",
			input: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"global":      nil,
				"metadata":    nil,
				"cluster":     []string{"all"},
			},
			expected: map[string]interface{}{
				"name":        "test-role",
				"description": "Test role",
				"global":      nil,
				"metadata":    nil,
				"cluster":     []interface{}{"all"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare the raw state JSON
			inputJSON, err := json.Marshal(tt.input)
			require.NoError(t, err)

			// Create the request
			req := resource.UpgradeStateRequest{
				RawState: &tfprotov6.RawState{
					JSON: inputJSON,
				},
			}

			// Create the response
			resp := &resource.UpgradeStateResponse{}

			// Call the function
			v0ToV1(context.Background(), req, resp)

			if tt.expectError {
				assert.True(t, resp.Diagnostics.HasError())
				if tt.errorContains != "" {
					found := false
					for _, diag := range resp.Diagnostics.Errors() {
						if assert.Contains(t, diag.Detail(), tt.errorContains) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error message not found")
				}
				return
			}

			// Should not have errors
			assert.False(t, resp.Diagnostics.HasError(), "Unexpected errors: %v", resp.Diagnostics)

			// Parse the output
			require.NotNil(t, resp.DynamicValue)
			require.NotNil(t, resp.DynamicValue.JSON)

			var actualState map[string]interface{}
			err = json.Unmarshal(resp.DynamicValue.JSON, &actualState)
			require.NoError(t, err)

			// Compare the results
			assert.Equal(t, tt.expected, actualState)
		})
	}
}

func TestV0ToV1_InvalidJSON(t *testing.T) {
	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: []byte("invalid json"),
		},
	}

	resp := &resource.UpgradeStateResponse{}

	v0ToV1(context.Background(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "State Upgrade Error")
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "Could not unmarshal prior state")
}
