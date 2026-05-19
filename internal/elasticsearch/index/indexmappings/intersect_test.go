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

package indexmappings

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	got := intersectMappings(api, state)
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

	got := intersectMappings(api, state)
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

	got := intersectMappings(api, state)
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

	got := intersectMappings(api, state)
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
