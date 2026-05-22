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
	"sort"
	"testing"

	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchemaDrift_AccessAttributes_KeysMatch asserts that the per-flavor
// `access` attribute builders expose the same nested-attribute keys and
// descriptions. Future divergence should be intentional.
func TestSchemaDrift_AccessAttributes_KeysMatch(t *testing.T) {
	resourceAttrs := AccessAttributesResource()
	ephemeralAttrs := AccessAttributesEphemeral()

	assert.Equal(t, sortedKeys(resourceAttrs), sortedKeys(ephemeralAttrs), "access nested attribute keys should match")

	resourceSearch, ok := resourceAttrs["search"].(schema.ListNestedAttribute)
	require.True(t, ok)
	ephemeralSearch, ok := ephemeralAttrs["search"].(eschema.ListNestedAttribute)
	require.True(t, ok)
	assert.Equal(t, resourceSearch.Description, ephemeralSearch.Description, "search descriptions should match")
	assert.Equal(t, sortedKeys(resourceSearch.NestedObject.Attributes), sortedKeys(ephemeralSearch.NestedObject.Attributes), "search nested keys should match")

	resourceReplication, ok := resourceAttrs["replication"].(schema.ListNestedAttribute)
	require.True(t, ok)
	ephemeralReplication, ok := ephemeralAttrs["replication"].(eschema.ListNestedAttribute)
	require.True(t, ok)
	assert.Equal(t, resourceReplication.Description, ephemeralReplication.Description, "replication descriptions should match")
	assert.Equal(t, sortedKeys(resourceReplication.NestedObject.Attributes), sortedKeys(ephemeralReplication.NestedObject.Attributes), "replication nested keys should match")
}

// TestSchemaDrift_SharedValidators_Identical asserts that the validator
// counts on the shared validator-builder functions match expectations and
// don't silently diverge.
func TestSchemaDrift_SharedValidators_Identical(t *testing.T) {
	assert.Len(t, NameValidators(), 2, "name validators should include LengthBetween + RegexMatches")
	assert.Len(t, TypeValidators(), 1, "type validator should be OneOf only")
	assert.Len(t, RoleDescriptorsValidators(), 1, "role_descriptors validator should be RequiresType only")
	assert.Len(t, AccessValidators(), 1, "access validator should be RequiresType only")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
