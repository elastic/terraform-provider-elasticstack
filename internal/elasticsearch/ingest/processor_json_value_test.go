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

package ingest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeProcessorJSON_CollapseSingleElementArrays(t *testing.T) {
	input := map[string]any{
		"remove": map[string]any{
			"field":          []any{"kubernetes.audit.responseObject.status"},
			"ignore_missing": true,
		},
	}
	want := map[string]any{
		"remove": map[string]any{
			"field":          "kubernetes.audit.responseObject.status",
			"ignore_missing": true,
		},
	}

	got := normalizeProcessorJSON(input)
	assert.Equal(t, want, got)
}

func TestNormalizeProcessorJSON_PreservesMultiElementArrays(t *testing.T) {
	input := map[string]any{
		"remove": map[string]any{
			"field": []any{"field_a", "field_b"},
		},
	}

	got := normalizeProcessorJSON(input)
	assert.Equal(t, input, got)
}

func TestNormalizeProcessorJSON_AppendValueCollapse(t *testing.T) {
	input := map[string]any{
		"append": map[string]any{
			"allow_duplicates": false,
			"field":            "tags",
			"value":            []any{"preserve_original_event"},
		},
	}
	want := map[string]any{
		"append": map[string]any{
			"allow_duplicates": false,
			"field":            "tags",
			"value":            "preserve_original_event",
		},
	}

	got := normalizeProcessorJSON(input)
	assert.Equal(t, want, got)
}

func TestNormalizeProcessorJSON_PreservesNestedOnFailureProcessors(t *testing.T) {
	input := map[string]any{
		"remove": map[string]any{
			"field":          []any{"my_field"},
			"ignore_missing": true,
			"on_failure": []any{
				map[string]any{
					"set": map[string]any{
						"field": "error.message",
						"value": []any{"failed to remove field"},
					},
				},
			},
		},
	}
	want := map[string]any{
		"remove": map[string]any{
			"field":          "my_field",
			"ignore_missing": true,
			"on_failure": []any{
				map[string]any{
					"set": map[string]any{
						"field": "error.message",
						"value": "failed to remove field",
					},
				},
			},
		},
	}

	got := normalizeProcessorJSON(input)
	assert.Equal(t, want, got, "on_failure array with one object element should be preserved, but scalar fields inside should collapse")
}

func TestProcessorJSONValue_SemanticEquality(t *testing.T) {
	ctx := context.Background()

	configJSON := `{"remove":{"field":"_audit_temp","ignore_missing":true}}`
	apiJSON := `{"remove":{"field":["_audit_temp"],"ignore_missing":true}}`

	configVal := NewProcessorJSONValue(configJSON)
	apiVal := NewProcessorJSONValue(apiJSON)

	eq, diags := configVal.StringSemanticEquals(ctx, apiVal)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.True(t, eq, "config and API processor JSON should be semantically equal")
}

func TestProcessorJSONValue_SemanticEquality_MultiElementDiffers(t *testing.T) {
	ctx := context.Background()

	a := NewProcessorJSONValue(`{"remove":{"field":"x"}}`)
	b := NewProcessorJSONValue(`{"remove":{"field":["x","y"]}}`)

	eq, diags := a.StringSemanticEquals(ctx, b)
	require.False(t, diags.HasError())
	assert.False(t, eq, "single string should not equal multi-element array")
}
