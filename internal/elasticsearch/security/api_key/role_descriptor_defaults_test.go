package api_key

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestPopulateRoleDescriptorsDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]models.ApiKeyRoleDescriptor
		expected map[string]models.ApiKeyRoleDescriptor
	}{
		{
			name:     "empty map returns empty map",
			input:    map[string]models.ApiKeyRoleDescriptor{},
			expected: map[string]models.ApiKeyRoleDescriptor{},
		},
		{
			name: "role with no indices returns unchanged",
			input: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
				},
			},
			expected: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
				},
			},
		},
		{
			name: "role with empty indices slice returns unchanged",
			input: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
					Indices: []models.IndexPerms{},
				},
			},
			expected: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
					Indices: []models.IndexPerms{},
				},
			},
		},
		{
			name: "index without AllowRestrictedIndices gets default false",
			input: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:      []string{"index1"},
							Privileges: []string{"read"},
						},
					},
				},
			},
			expected: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: utils.Pointer(false),
						},
					},
				},
			},
		},
		{
			name: "index with AllowRestrictedIndices true preserves value",
			input: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: utils.Pointer(true),
						},
					},
				},
			},
			expected: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: utils.Pointer(true),
						},
					},
				},
			},
		},
		{
			name: "index with AllowRestrictedIndices false preserves value",
			input: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: utils.Pointer(false),
						},
					},
				},
			},
			expected: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: utils.Pointer(false),
						},
					},
				},
			},
		},
		{
			name: "multiple indices with mixed AllowRestrictedIndices values",
			input: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:      []string{"index1"},
							Privileges: []string{"read"},
							// No AllowRestrictedIndices set
						},
						{
							Names:                  []string{"index2"},
							Privileges:             []string{"write"},
							AllowRestrictedIndices: utils.Pointer(true),
						},
						{
							Names:                  []string{"index3"},
							Privileges:             []string{"read", "write"},
							AllowRestrictedIndices: utils.Pointer(false),
						},
					},
				},
			},
			expected: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: utils.Pointer(false),
						},
						{
							Names:                  []string{"index2"},
							Privileges:             []string{"write"},
							AllowRestrictedIndices: utils.Pointer(true),
						},
						{
							Names:                  []string{"index3"},
							Privileges:             []string{"read", "write"},
							AllowRestrictedIndices: utils.Pointer(false),
						},
					},
				},
			},
		},
		{
			name: "multiple roles with mixed configurations",
			input: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
					Indices: []models.IndexPerms{
						{
							Names:      []string{"admin-*"},
							Privileges: []string{"all"},
						},
					},
				},
				"reader": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"logs-*"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: utils.Pointer(true),
						},
					},
				},
				"writer": {
					Cluster: []string{"monitor"},
					// No indices
				},
			},
			expected: map[string]models.ApiKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"admin-*"},
							Privileges:             []string{"all"},
							AllowRestrictedIndices: utils.Pointer(false),
						},
					},
				},
				"reader": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"logs-*"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: utils.Pointer(true),
						},
					},
				},
				"writer": {
					Cluster: []string{"monitor"},
				},
			},
		},
		{
			name: "role with complex index permissions",
			input: map[string]models.ApiKeyRoleDescriptor{
				"complex": {
					Cluster: []string{"monitor", "manage"},
					Indices: []models.IndexPerms{
						{
							Names:      []string{"sensitive-*"},
							Privileges: []string{"read", "view_index_metadata"},
							Query:      utils.Pointer(`{"term": {"public": true}}`),
							FieldSecurity: &models.FieldSecurity{
								Grant: []string{"public_*"},
							},
						},
					},
					Metadata: map[string]interface{}{
						"version": 1,
						"tags":    []string{"production"},
					},
				},
			},
			expected: map[string]models.ApiKeyRoleDescriptor{
				"complex": {
					Cluster: []string{"monitor", "manage"},
					Indices: []models.IndexPerms{
						{
							Names:      []string{"sensitive-*"},
							Privileges: []string{"read", "view_index_metadata"},
							Query:      utils.Pointer(`{"term": {"public": true}}`),
							FieldSecurity: &models.FieldSecurity{
								Grant: []string{"public_*"},
							},
							AllowRestrictedIndices: utils.Pointer(false),
						},
					},
					Metadata: map[string]interface{}{
						"version": 1,
						"tags":    []string{"production"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := populateRoleDescriptorsDefaults(tt.input)
			assert.Equal(t, tt.expected, result)

			// Verify that the function modifies the input map
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestPopulateRoleDescriptorsDefaults_NilInput(t *testing.T) {
	var input map[string]models.ApiKeyRoleDescriptor
	result := populateRoleDescriptorsDefaults(input)
	assert.Nil(t, result)
}
