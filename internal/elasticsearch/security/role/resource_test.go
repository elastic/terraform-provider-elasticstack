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
		input         map[string]any
		expected      map[string]any
		expectError   bool
		errorContains string
	}{
		{
			name: "empty_global_and_metadata_removed",
			input: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"global":      "",
				"metadata":    "",
				"cluster":     []string{"all"},
			},
			expected: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"cluster":     []any{"all"},
			},
		},
		{
			name: "non_empty_global_and_metadata_preserved",
			input: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"global":      `{"profile": {"privileges": ["manage"]}}`,
				"metadata":    `{"version": 1}`,
				"cluster":     []string{"all"},
			},
			expected: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"global":      `{"profile": {"privileges": ["manage"]}}`,
				"metadata":    `{"version": 1}`,
				"cluster":     []any{"all"},
			},
		},
		{
			name: "empty_query_in_indices_removed",
			input: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"indices": []any{
					map[string]any{
						"names":      []string{"index1", "index2"},
						"privileges": []string{"read"},
						"query":      "",
					},
					map[string]any{
						"names":      []string{"index3"},
						"privileges": []string{"write"},
						"query":      `{"match": {"field": "value"}}`,
					},
				},
			},
			expected: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"indices": []any{
					map[string]any{
						"names":      []any{"index1", "index2"},
						"privileges": []any{"read"},
					},
					map[string]any{
						"names":      []any{"index3"},
						"privileges": []any{"write"},
						"query":      `{"match": {"field": "value"}}`,
					},
				},
			},
		},
		{
			name: "empty_query_in_remote_indices_removed",
			input: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"remote_indices": []any{
					map[string]any{
						"clusters":   []string{"cluster1"},
						"names":      []string{"remote-index1"},
						"privileges": []string{"read"},
						"query":      "",
					},
					map[string]any{
						"clusters":   []string{"cluster2"},
						"names":      []string{"remote-index2"},
						"privileges": []string{"write"},
						"query":      `{"term": {"status": "active"}}`,
					},
				},
			},
			expected: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"remote_indices": []any{
					map[string]any{
						"clusters":   []any{"cluster1"},
						"names":      []any{"remote-index1"},
						"privileges": []any{"read"},
					},
					map[string]any{
						"clusters":   []any{"cluster2"},
						"names":      []any{"remote-index2"},
						"privileges": []any{"write"},
						"query":      `{"term": {"status": "active"}}`,
					},
				},
			},
		},
		{
			name: "all_empty_fields_removed_comprehensive",
			input: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"global":      "",
				"metadata":    "",
				"cluster":     []string{"all"},
				"indices": []any{
					map[string]any{
						"names":      []string{"index1"},
						"privileges": []string{"read"},
						"query":      "",
					},
				},
				"remote_indices": []any{
					map[string]any{
						"clusters":   []string{"cluster1"},
						"names":      []string{"remote-index1"},
						"privileges": []string{"read"},
						"query":      "",
					},
				},
			},
			expected: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"cluster":     []any{"all"},
				"indices": []any{
					map[string]any{
						"names":      []any{"index1"},
						"privileges": []any{"read"},
					},
				},
				"remote_indices": []any{
					map[string]any{
						"clusters":   []any{"cluster1"},
						"names":      []any{"remote-index1"},
						"privileges": []any{"read"},
					},
				},
			},
		},
		{
			name: "no_indices_or_remote_indices",
			input: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"global":      "",
				"metadata":    "",
				"cluster":     []string{"all"},
			},
			expected: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"cluster":     []any{"all"},
			},
		},
		{
			name: "index_item_not_map",
			input: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"indices": []any{
					"not-a-map",
					map[string]any{
						"names":      []string{"index1"},
						"privileges": []string{"read"},
						"query":      "",
					},
				},
			},
			expected: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"indices": []any{
					"not-a-map", // Should be preserved as-is if not a map
					map[string]any{
						"names":      []any{"index1"},
						"privileges": []any{"read"},
					},
				},
			},
		},
		{
			name: "remote_index_item_not_map",
			input: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"remote_indices": []any{
					"not-a-map",
					map[string]any{
						"clusters":   []string{"cluster1"},
						"names":      []string{"remote-index1"},
						"privileges": []string{"read"},
						"query":      "",
					},
				},
			},
			expected: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"remote_indices": []any{
					"not-a-map", // Should be preserved as-is if not a map
					map[string]any{
						"clusters":   []any{"cluster1"},
						"names":      []any{"remote-index1"},
						"privileges": []any{"read"},
					},
				},
			},
		},
		{
			name: "nil_global_and_metadata_removed",
			input: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"global":      nil,
				"metadata":    nil,
				"cluster":     []string{"all"},
			},
			expected: map[string]any{
				"name":        "test-role",
				"description": "Test role",
				"cluster":     []any{"all"},
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

			var actualState map[string]any
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
