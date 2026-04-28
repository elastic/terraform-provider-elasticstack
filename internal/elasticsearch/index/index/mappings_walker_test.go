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

func TestMappingsSemanticallyEqual_MatchingMappings(t *testing.T) {
	t.Parallel()

	user := map[string]any{
		"properties": map[string]any{
			"field1": map[string]any{"type": "keyword"},
		},
	}
	api := map[string]any{
		"properties": map[string]any{
			"field1": map[string]any{"type": "keyword"},
		},
	}

	assert.True(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_TemplateInjectedProperties(t *testing.T) {
	t.Parallel()

	user := map[string]any{
		"properties": map[string]any{
			"user_field": map[string]any{"type": "keyword"},
		},
	}
	api := map[string]any{
		"properties": map[string]any{
			"user_field":     map[string]any{"type": "keyword"},
			"template_field": map[string]any{"type": "text"},
		},
	}

	assert.True(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_TemplateInjectedDynamicTemplates(t *testing.T) {
	t.Parallel()

	user := map[string]any{
		"properties": map[string]any{
			"user_field": map[string]any{"type": "keyword"},
		},
	}
	api := map[string]any{
		"dynamic_templates": []any{
			map[string]any{
				"strings_as_ip": map[string]any{
					"match":   "ip*",
					"runtime": map[string]any{"type": "ip"},
				},
			},
		},
		"properties": map[string]any{
			"user_field": map[string]any{"type": "keyword"},
		},
	}

	assert.True(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_RetainedFields(t *testing.T) {
	t.Parallel()

	// User removed a field from config, but ES retains it in API
	user := map[string]any{
		"properties": map[string]any{
			"user_field": map[string]any{"type": "keyword"},
		},
	}
	api := map[string]any{
		"properties": map[string]any{
			"user_field":     map[string]any{"type": "keyword"},
			"retained_field": map[string]any{"type": "text"},
		},
	}

	assert.True(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_TypeChangeNotEqual(t *testing.T) {
	t.Parallel()

	user := map[string]any{
		"properties": map[string]any{
			"field1": map[string]any{"type": "keyword"},
		},
	}
	api := map[string]any{
		"properties": map[string]any{
			"field1": map[string]any{"type": "text"},
		},
	}

	assert.False(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_SemanticTextModelSettings(t *testing.T) {
	t.Parallel()

	user := map[string]any{
		"properties": map[string]any{
			"semantic_field": map[string]any{
				"type": "semantic_text",
			},
		},
	}
	api := map[string]any{
		"properties": map[string]any{
			"semantic_field": map[string]any{
				"type": "semantic_text",
				"model_settings": map[string]any{
					"task_type": "sparse_embedding",
				},
			},
		},
	}

	assert.True(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_ExplicitModelSettings(t *testing.T) {
	t.Parallel()

	user := map[string]any{
		"properties": map[string]any{
			"semantic_field": map[string]any{
				"type": "semantic_text",
				"model_settings": map[string]any{
					"task_type": "sparse_embedding",
				},
			},
		},
	}
	api := map[string]any{
		"properties": map[string]any{
			"semantic_field": map[string]any{
				"type": "semantic_text",
				"model_settings": map[string]any{
					"task_type": "dense_embedding",
				},
			},
		},
	}

	assert.False(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_MissingFieldInAPI(t *testing.T) {
	t.Parallel()

	user := map[string]any{
		"properties": map[string]any{
			"field1": map[string]any{"type": "keyword"},
		},
	}
	api := map[string]any{
		"properties": map[string]any{},
	}

	assert.False(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_EmptyMappings(t *testing.T) {
	t.Parallel()

	user := map[string]any{}
	api := map[string]any{}

	assert.True(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_TemplateOnlyMappings(t *testing.T) {
	t.Parallel()

	user := map[string]any{}
	api := map[string]any{
		"properties": map[string]any{
			"template_field": map[string]any{"type": "keyword"},
		},
		"dynamic_templates": []any{
			map[string]any{"strings_as_ip": map[string]any{"match": "ip*"}},
		},
	}

	assert.True(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_NestedPropertiesEqual(t *testing.T) {
	t.Parallel()

	user := map[string]any{
		"properties": map[string]any{
			"parent": map[string]any{
				"properties": map[string]any{
					"child": map[string]any{"type": "keyword"},
				},
			},
		},
	}
	api := map[string]any{
		"properties": map[string]any{
			"parent": map[string]any{
				"properties": map[string]any{
					"child":  map[string]any{"type": "keyword"},
					"child2": map[string]any{"type": "text"},
				},
			},
		},
	}

	assert.True(t, mappingsSemanticallyEqual(user, api))
}

func TestMappingsSemanticallyEqual_NestedTypeChangeNotEqual(t *testing.T) {
	t.Parallel()

	user := map[string]any{
		"properties": map[string]any{
			"parent": map[string]any{
				"properties": map[string]any{
					"child": map[string]any{"type": "keyword"},
				},
			},
		},
	}
	api := map[string]any{
		"properties": map[string]any{
			"parent": map[string]any{
				"properties": map[string]any{
					"child": map[string]any{"type": "text"},
				},
			},
		},
	}

	assert.False(t, mappingsSemanticallyEqual(user, api))
}
