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

package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareMappingsForPlan_MatchingMappings(t *testing.T) {
	t.Parallel()

	state := map[string]any{
		"properties": map[string]any{
			"field1": map[string]any{"type": "keyword"},
		},
	}
	cfg := map[string]any{
		"properties": map[string]any{
			"field1": map[string]any{"type": "keyword"},
		},
	}

	result := compareMappingsForPlan(state, cfg)
	assert.False(t, result.RequiresReplace)
	assert.Empty(t, result.RemovedFields)
	assert.Empty(t, result.Diags)
}

func TestCompareMappingsForPlan_TemplateInjectedProperties(t *testing.T) {
	t.Parallel()

	state := map[string]any{
		"properties": map[string]any{
			"user_field":     map[string]any{"type": "keyword"},
			"template_field": map[string]any{"type": "text"},
		},
	}
	cfg := map[string]any{
		"properties": map[string]any{
			"user_field": map[string]any{"type": "keyword"},
		},
	}

	result := compareMappingsForPlan(state, cfg)
	assert.False(t, result.RequiresReplace)
	require.Len(t, result.RemovedFields, 1)
	assert.Contains(t, result.RemovedFields, `mappings["properties"]["template_field"]`)
	require.Len(t, result.Diags, 1)
}

func TestCompareMappingsForPlan_TypeChangeRequiresReplace(t *testing.T) {
	t.Parallel()

	state := map[string]any{
		"properties": map[string]any{
			"field1": map[string]any{"type": "keyword"},
		},
	}
	cfg := map[string]any{
		"properties": map[string]any{
			"field1": map[string]any{"type": "text"},
		},
	}

	result := compareMappingsForPlan(state, cfg)
	assert.True(t, result.RequiresReplace)
}

func TestCompareMappingsForPlan_PropertiesRemovedEntirely(t *testing.T) {
	t.Parallel()

	state := map[string]any{
		"properties": map[string]any{
			"field1": map[string]any{"type": "keyword"},
		},
	}
	cfg := map[string]any{}

	result := compareMappingsForPlan(state, cfg)
	assert.True(t, result.RequiresReplace)
}

func TestCompareMappingsForPlan_SemanticTextModelSettings(t *testing.T) {
	t.Parallel()

	state := map[string]any{
		"properties": map[string]any{
			"semantic_field": map[string]any{
				"type": "semantic_text",
				"model_settings": map[string]any{
					"task_type": "sparse_embedding",
				},
			},
		},
	}
	cfg := map[string]any{
		"properties": map[string]any{
			"semantic_field": map[string]any{
				"type": "semantic_text",
			},
		},
	}

	result := compareMappingsForPlan(state, cfg)
	assert.False(t, result.RequiresReplace)
}

func TestCompareMappingsForPlan_SemanticTextExplicitModelSettings(t *testing.T) {
	t.Parallel()

	state := map[string]any{
		"properties": map[string]any{
			"semantic_field": map[string]any{
				"type": "semantic_text",
				"model_settings": map[string]any{
					"task_type": "sparse_embedding",
				},
			},
		},
	}
	cfg := map[string]any{
		"properties": map[string]any{
			"semantic_field": map[string]any{
				"type": "semantic_text",
				"model_settings": map[string]any{
					"task_type": "dense_embedding",
				},
			},
		},
	}

	result := compareMappingsForPlan(state, cfg)
	assert.False(t, result.RequiresReplace)
}

func TestCompareMappingsForPlan_NestedPropertiesRemoved(t *testing.T) {
	t.Parallel()

	state := map[string]any{
		"properties": map[string]any{
			"parent": map[string]any{
				"properties": map[string]any{
					"child": map[string]any{"type": "keyword"},
				},
			},
		},
	}
	cfg := map[string]any{
		"properties": map[string]any{
			"parent": map[string]any{
				"properties": map[string]any{},
			},
		},
	}

	result := compareMappingsForPlan(state, cfg)
	assert.False(t, result.RequiresReplace)
	require.Len(t, result.RemovedFields, 1)
	assert.Contains(t, result.RemovedFields, `mappings["properties"]["parent"]["properties"]["child"]`)
	require.Len(t, result.Diags, 1)
}
