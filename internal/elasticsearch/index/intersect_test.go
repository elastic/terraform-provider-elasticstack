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

func TestIntersectMappings_dropsUndeclaredProperties(t *testing.T) {
	api := map[string]any{
		"properties": map[string]any{
			"title": map[string]any{"type": "text"},
			"tags":  map[string]any{"type": "keyword"},
		},
	}
	state := map[string]any{
		"properties": map[string]any{
			"title": map[string]any{"type": "text"},
		},
	}

	got := IntersectMappings(api, state)
	props := got["properties"].(map[string]any)
	assert.Len(t, props, 1)
	assert.Contains(t, props, "title")
	assert.NotContains(t, props, "tags")
}

func TestIntersectMappings_retainsOtherTopLevelKeys(t *testing.T) {
	api := map[string]any{
		"dynamic": false,
		"properties": map[string]any{
			"title": map[string]any{"type": "text"},
		},
	}
	state := map[string]any{
		"dynamic": "strict",
		"properties": map[string]any{
			"title": map[string]any{"type": "text"},
		},
	}

	got := IntersectMappings(api, state)
	assert.Equal(t, false, got["dynamic"])
}

func TestIntersectMappings_retainsDeclaredKeyWhenAPIOmits(t *testing.T) {
	api := map[string]any{
		"properties": map[string]any{
			"title": map[string]any{"type": "text"},
		},
	}
	state := map[string]any{
		"_source": map[string]any{
			"enabled": true,
		},
		"properties": map[string]any{
			"title": map[string]any{"type": "text"},
		},
	}

	got := IntersectMappings(api, state)
	assert.Equal(t, true, got["_source"].(map[string]any)["enabled"])
}

func TestIntersectMappings_keepsDeclaredShapeWhenSemanticallyEqual(t *testing.T) {
	api := map[string]any{
		"runtime": map[string]any{
			"day_of_week": map[string]any{
				"type": "keyword",
				"script": map[string]any{
					"lang":   "painless",
					"source": "emit(1)",
				},
			},
		},
	}
	state := map[string]any{
		"runtime": map[string]any{
			"day_of_week": map[string]any{
				"type":   "keyword",
				"script": "emit(1)",
			},
		},
	}

	got := IntersectMappings(api, state)
	runtime := got["runtime"].(map[string]any)
	field := runtime["day_of_week"].(map[string]any)
	assert.Equal(t, "emit(1)", field["script"])
}

func TestIntersectProperties_nested(t *testing.T) {
	api := map[string]any{
		"author": map[string]any{
			"properties": map[string]any{
				"name":  map[string]any{"type": "text"},
				"email": map[string]any{"type": "keyword"},
			},
		},
	}
	state := map[string]any{
		"author": map[string]any{
			"properties": map[string]any{
				"name": map[string]any{"type": "text"},
			},
		},
	}

	got := intersectProperties(api, state)
	author := got["author"].(map[string]any)
	props := author["properties"].(map[string]any)
	assert.Len(t, props, 1)
	assert.Contains(t, props, "name")
}

func TestIntersectMappings_dynamicTemplates(t *testing.T) {
	alpha := map[string]any{
		"alpha": map[string]any{"mapping": map[string]any{"type": "text"}},
	}
	beta := map[string]any{
		"beta": map[string]any{"mapping": map[string]any{"type": "keyword"}},
	}
	extra := map[string]any{
		"template_default": map[string]any{"mapping": map[string]any{"type": "keyword"}},
	}

	tests := []struct {
		name            string
		api             map[string]any
		state           map[string]any
		wantNames       []string
		wantOmitKey     bool
		wantPassthrough bool
	}{
		{
			name: "drops index-template-injected extra",
			api: map[string]any{
				"dynamic_templates": []any{alpha, extra},
			},
			state: map[string]any{
				"dynamic_templates": []any{alpha},
			},
			wantNames: []string{"alpha"},
		},
		{
			name: "omits declared name absent from API",
			api: map[string]any{
				"dynamic_templates": []any{alpha},
			},
			state: map[string]any{
				"dynamic_templates": []any{alpha, beta},
			},
			wantNames: []string{"alpha"},
		},
		{
			name: "omits key when API omits dynamic_templates entirely",
			api:  map[string]any{},
			state: map[string]any{
				"dynamic_templates": []any{alpha},
			},
			wantOmitKey: true,
		},
		{
			name: "passthrough when API has duplicate template name",
			api: map[string]any{
				"dynamic_templates": []any{alpha, alpha},
			},
			state: map[string]any{
				"dynamic_templates": []any{alpha},
			},
			wantPassthrough: true,
		},
		{
			name: "passthrough when state has duplicate template name",
			api: map[string]any{
				"dynamic_templates": []any{alpha},
			},
			state: map[string]any{
				"dynamic_templates": []any{alpha, alpha},
			},
			wantPassthrough: true,
		},
		{
			name: "passthrough when API entry value is not an object",
			api: map[string]any{
				"dynamic_templates": []any{
					map[string]any{"alpha": "not-an-object"},
				},
			},
			state: map[string]any{
				"dynamic_templates": []any{alpha},
			},
			wantPassthrough: true,
		},
		{
			name: "passthrough when state entry value is not an object",
			api: map[string]any{
				"dynamic_templates": []any{alpha},
			},
			state: map[string]any{
				"dynamic_templates": []any{
					map[string]any{"alpha": "not-an-object"},
				},
			},
			wantPassthrough: true,
		},
		{
			name: "preserves state's declared order not API order",
			api: map[string]any{
				"dynamic_templates": []any{beta, alpha},
			},
			state: map[string]any{
				"dynamic_templates": []any{alpha, beta},
			},
			wantNames: []string{"alpha", "beta"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IntersectMappings(tc.api, tc.state)
			raw, present := got[dynamicTemplatesKey]
			if tc.wantOmitKey {
				assert.False(t, present, "expected dynamic_templates to be omitted from result")
				return
			}
			require.True(t, present, "expected dynamic_templates in result")
			gotArr, ok := raw.([]any)
			require.True(t, ok, "expected dynamic_templates to be an array")
			if tc.wantPassthrough {
				assert.Equal(t, tc.api[dynamicTemplatesKey], gotArr)
				return
			}
			assert.Equal(t, tc.wantNames, dynamicTemplateEntryNames(t, gotArr))
		})
	}
}

func dynamicTemplateEntryNames(t *testing.T, templates []any) []string {
	t.Helper()
	names := make([]string, 0, len(templates))
	for _, raw := range templates {
		entry, ok := raw.(map[string]any)
		require.True(t, ok)
		require.Len(t, entry, 1)
		for name := range entry {
			names = append(names, name)
		}
	}
	return names
}
