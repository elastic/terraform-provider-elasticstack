// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package apikey

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestPopulateRoleDescriptorsDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]models.APIKeyRoleDescriptor
		expected map[string]models.APIKeyRoleDescriptor
	}{
		{
			name:     "empty map returns empty map",
			input:    map[string]models.APIKeyRoleDescriptor{},
			expected: map[string]models.APIKeyRoleDescriptor{},
		},
		{
			name: "role with no indices returns unchanged",
			input: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
				},
			},
			expected: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
				},
			},
		},
		{
			name: "role with empty indices slice returns unchanged",
			input: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
					Indices: []models.IndexPerms{},
				},
			},
			expected: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
					Indices: []models.IndexPerms{},
				},
			},
		},
		{
			name: "index without AllowRestrictedIndices gets default false",
			input: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:      []string{"index1"},
							Privileges: []string{"read"},
						},
					},
				},
			},
			expected: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: schemautil.Pointer(false),
						},
					},
				},
			},
		},
		{
			name: "index with AllowRestrictedIndices true preserves value",
			input: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: schemautil.Pointer(true),
						},
					},
				},
			},
			expected: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: schemautil.Pointer(true),
						},
					},
				},
			},
		},
		{
			name: "index with AllowRestrictedIndices false preserves value",
			input: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: schemautil.Pointer(false),
						},
					},
				},
			},
			expected: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: schemautil.Pointer(false),
						},
					},
				},
			},
		},
		{
			name: "multiple indices with mixed AllowRestrictedIndices values",
			input: map[string]models.APIKeyRoleDescriptor{
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
							AllowRestrictedIndices: schemautil.Pointer(true),
						},
						{
							Names:                  []string{"index3"},
							Privileges:             []string{"read", "write"},
							AllowRestrictedIndices: schemautil.Pointer(false),
						},
					},
				},
			},
			expected: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"index1"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: schemautil.Pointer(false),
						},
						{
							Names:                  []string{"index2"},
							Privileges:             []string{"write"},
							AllowRestrictedIndices: schemautil.Pointer(true),
						},
						{
							Names:                  []string{"index3"},
							Privileges:             []string{"read", "write"},
							AllowRestrictedIndices: schemautil.Pointer(false),
						},
					},
				},
			},
		},
		{
			name: "multiple roles with mixed configurations",
			input: map[string]models.APIKeyRoleDescriptor{
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
							AllowRestrictedIndices: schemautil.Pointer(true),
						},
					},
				},
				"writer": {
					Cluster: []string{"monitor"},
					// No indices
				},
			},
			expected: map[string]models.APIKeyRoleDescriptor{
				"admin": {
					Cluster: []string{"monitor"},
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"admin-*"},
							Privileges:             []string{"all"},
							AllowRestrictedIndices: schemautil.Pointer(false),
						},
					},
				},
				"reader": {
					Indices: []models.IndexPerms{
						{
							Names:                  []string{"logs-*"},
							Privileges:             []string{"read"},
							AllowRestrictedIndices: schemautil.Pointer(true),
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
			input: map[string]models.APIKeyRoleDescriptor{
				"complex": {
					Cluster: []string{"monitor", "manage"},
					Indices: []models.IndexPerms{
						{
							Names:      []string{"sensitive-*"},
							Privileges: []string{"read", "view_index_metadata"},
							Query:      schemautil.Pointer(`{"term": {"public": true}}`),
							FieldSecurity: &models.FieldSecurity{
								Grant: []string{"public_*"},
							},
						},
					},
					Metadata: map[string]any{
						"version": 1,
						"tags":    []string{"production"},
					},
				},
			},
			expected: map[string]models.APIKeyRoleDescriptor{
				"complex": {
					Cluster: []string{"monitor", "manage"},
					Indices: []models.IndexPerms{
						{
							Names:      []string{"sensitive-*"},
							Privileges: []string{"read", "view_index_metadata"},
							Query:      schemautil.Pointer(`{"term": {"public": true}}`),
							FieldSecurity: &models.FieldSecurity{
								Grant: []string{"public_*"},
							},
							AllowRestrictedIndices: schemautil.Pointer(false),
						},
					},
					Metadata: map[string]any{
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
	var input map[string]models.APIKeyRoleDescriptor
	result := populateRoleDescriptorsDefaults(input)
	assert.Nil(t, result)
}
