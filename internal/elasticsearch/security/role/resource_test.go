package role

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
			name: "nil_global_and_metadata_removed",
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

// TestFromAPIModel_DefaultValuesNotNull tests that when the API returns nil for optional
// fields with defaults (due to omitempty), we return the actual default values, not null.
// This prevents the "planned set element does not correlate with any element in actual" error.
// See: https://github.com/elastic/terraform-provider-elasticstack/issues/XXX
func TestFromAPIModel_DefaultValuesNotNull(t *testing.T) {
	ctx := context.Background()

	t.Run("allow_restricted_indices_nil_returns_false", func(t *testing.T) {
		// Simulate API response where allow_restricted_indices is nil (omitted due to omitempty)
		role := &models.Role{
			Name: "test-role",
			Indices: []models.IndexPerms{
				{
					Names:                  []string{"index1"},
					Privileges:             []string{"read"},
					AllowRestrictedIndices: nil, // API omits false values
					FieldSecurity: &models.FieldSecurity{
						Grant:  []string{"*"},
						Except: nil, // API omits empty arrays
					},
				},
			},
		}

		data := &RoleData{}
		diags := data.fromAPIModel(ctx, role)

		require.False(t, diags.HasError(), "fromAPIModel should not return errors")

		// Extract the indices set
		var indicesList []IndexPermsData
		diags = data.Indices.ElementsAs(ctx, &indicesList, false)
		require.False(t, diags.HasError())
		require.Len(t, indicesList, 1)

		// Verify allow_restricted_indices is false, not null
		assert.False(t, indicesList[0].AllowRestrictedIndices.IsNull(),
			"allow_restricted_indices should not be null when API returns nil")
		assert.Equal(t, false, indicesList[0].AllowRestrictedIndices.ValueBool(),
			"allow_restricted_indices should be false when API returns nil")
	})

	t.Run("field_security_except_nil_returns_empty_set", func(t *testing.T) {
		// Simulate API response where except is nil (omitted due to omitempty)
		role := &models.Role{
			Name: "test-role",
			Indices: []models.IndexPerms{
				{
					Names:      []string{"index1"},
					Privileges: []string{"read"},
					FieldSecurity: &models.FieldSecurity{
						Grant:  []string{"*"},
						Except: nil, // API omits empty arrays
					},
				},
			},
		}

		data := &RoleData{}
		diags := data.fromAPIModel(ctx, role)

		require.False(t, diags.HasError(), "fromAPIModel should not return errors")

		// Extract the indices set
		var indicesList []IndexPermsData
		diags = data.Indices.ElementsAs(ctx, &indicesList, false)
		require.False(t, diags.HasError())
		require.Len(t, indicesList, 1)

		// Extract field_security
		var fieldSec FieldSecurityData
		diags = indicesList[0].FieldSecurity.As(ctx, &fieldSec, basetypes.ObjectAsOptions{})
		require.False(t, diags.HasError())

		// Verify except is an empty set, not null
		assert.False(t, fieldSec.Except.IsNull(),
			"except should not be null when API returns nil")
		assert.Equal(t, 0, len(fieldSec.Except.Elements()),
			"except should be an empty set when API returns nil")
	})

	t.Run("remote_indices_field_security_except_nil_returns_empty_set", func(t *testing.T) {
		// Simulate API response for remote_indices where except is nil
		role := &models.Role{
			Name: "test-role",
			RemoteIndices: []models.RemoteIndexPerms{
				{
					IndexPerms: models.IndexPerms{
						Names:      []string{"remote-index1"},
						Privileges: []string{"read"},
						FieldSecurity: &models.FieldSecurity{
							Grant:  []string{"*"},
							Except: nil, // API omits empty arrays
						},
					},
					Clusters: []string{"cluster1"},
				},
			},
		}

		data := &RoleData{}
		diags := data.fromAPIModel(ctx, role)

		require.False(t, diags.HasError(), "fromAPIModel should not return errors")

		// Extract the remote_indices set
		var remoteIndicesList []RemoteIndexPermsData
		diags = data.RemoteIndices.ElementsAs(ctx, &remoteIndicesList, false)
		require.False(t, diags.HasError())
		require.Len(t, remoteIndicesList, 1)

		// Extract field_security
		var fieldSec FieldSecurityData
		diags = remoteIndicesList[0].FieldSecurity.As(ctx, &fieldSec, basetypes.ObjectAsOptions{})
		require.False(t, diags.HasError())

		// Verify except is an empty set, not null
		assert.False(t, fieldSec.Except.IsNull(),
			"except should not be null when API returns nil for remote_indices")
		assert.Equal(t, 0, len(fieldSec.Except.Elements()),
			"except should be an empty set when API returns nil for remote_indices")
	})

	t.Run("explicit_values_preserved", func(t *testing.T) {
		// Simulate API response with explicit values (not defaults)
		allowRestricted := true
		role := &models.Role{
			Name: "test-role",
			Indices: []models.IndexPerms{
				{
					Names:                  []string{"index1"},
					Privileges:             []string{"read"},
					AllowRestrictedIndices: &allowRestricted, // Explicit true
					FieldSecurity: &models.FieldSecurity{
						Grant:  []string{"*"},
						Except: []string{"secret_field"}, // Explicit non-empty
					},
				},
			},
		}

		data := &RoleData{}
		diags := data.fromAPIModel(ctx, role)

		require.False(t, diags.HasError(), "fromAPIModel should not return errors")

		// Extract the indices set
		var indicesList []IndexPermsData
		diags = data.Indices.ElementsAs(ctx, &indicesList, false)
		require.False(t, diags.HasError())
		require.Len(t, indicesList, 1)

		// Verify allow_restricted_indices is true
		assert.Equal(t, true, indicesList[0].AllowRestrictedIndices.ValueBool(),
			"allow_restricted_indices should preserve explicit true value")

		// Extract field_security
		var fieldSec FieldSecurityData
		diags = indicesList[0].FieldSecurity.As(ctx, &fieldSec, basetypes.ObjectAsOptions{})
		require.False(t, diags.HasError())

		// Verify except has the explicit value
		var exceptList []string
		diags = fieldSec.Except.ElementsAs(ctx, &exceptList, false)
		require.False(t, diags.HasError())
		assert.Equal(t, []string{"secret_field"}, exceptList,
			"except should preserve explicit values")
	})
}
