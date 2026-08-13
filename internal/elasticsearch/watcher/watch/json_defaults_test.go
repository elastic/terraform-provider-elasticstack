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

package watch

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func searchRequest(extras map[string]any) map[string]any {
	req := map[string]any{
		"body": map[string]any{
			"query": map[string]any{
				"match_all": map[string]any{},
			},
		},
	}
	for k, v := range extras {
		req[k] = v
	}
	return map[string]any{
		"search": map[string]any{
			"request": req,
		},
	}
}

func Test_populateWatcherJSONDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name:  "top-level search.request missing all three keys",
			input: searchRequest(nil),
			expected: searchRequest(map[string]any{
				"rest_total_hits_as_int": true,
				"search_type":            "query_then_fetch",
				"indices":                []any{},
			}),
		},
		{
			name: "missing only rest_total_hits_as_int",
			input: searchRequest(map[string]any{
				"search_type": "query_then_fetch",
				"indices":     []any{".monitoring-es-*"},
			}),
			expected: searchRequest(map[string]any{
				"rest_total_hits_as_int": true,
				"search_type":            "query_then_fetch",
				"indices":                []any{".monitoring-es-*"},
			}),
		},
		{
			name: "explicit non-default search-request keys are preserved",
			input: searchRequest(map[string]any{
				"rest_total_hits_as_int": false,
				"search_type":            "dfs_query_then_fetch",
				"indices":                []any{".monitoring-es-*"},
			}),
			expected: searchRequest(map[string]any{
				"rest_total_hits_as_int": false,
				"search_type":            "dfs_query_then_fetch",
				"indices":                []any{".monitoring-es-*"},
			}),
		},
		{
			name: "chain inputs array defaults each named search independently",
			input: map[string]any{
				"chain": map[string]any{
					"inputs": []any{
						map[string]any{
							"first": searchRequest(nil),
						},
						map[string]any{
							"second": searchRequest(map[string]any{
								"search_type": "query_then_fetch",
							}),
						},
					},
				},
			},
			expected: map[string]any{
				"chain": map[string]any{
					"inputs": []any{
						map[string]any{
							"first": searchRequest(map[string]any{
								"rest_total_hits_as_int": true,
								"search_type":            "query_then_fetch",
								"indices":                []any{},
							}),
						},
						map[string]any{
							"second": searchRequest(map[string]any{
								"rest_total_hits_as_int": true,
								"search_type":            "query_then_fetch",
								"indices":                []any{},
							}),
						},
					},
				},
			},
		},
		{
			name: "http input request is left untouched",
			input: map[string]any{
				"http": map[string]any{
					"request": map[string]any{
						"host": "api.example",
						"path": "/v1/data",
					},
				},
			},
			expected: map[string]any{
				"http": map[string]any{
					"request": map[string]any{
						"host": "api.example",
						"path": "/v1/data",
					},
				},
			},
		},
		{
			name:  "transform search.request missing defaults",
			input: searchRequest(nil),
			expected: searchRequest(map[string]any{
				"rest_total_hits_as_int": true,
				"search_type":            "query_then_fetch",
				"indices":                []any{},
			}),
		},
		{
			name: "actions nested search transform is defaulted",
			input: map[string]any{
				"log": map[string]any{
					"transform": searchRequest(nil),
					"logging": map[string]any{
						"text": "watch fired",
					},
				},
			},
			expected: map[string]any{
				"log": map[string]any{
					"transform": searchRequest(map[string]any{
						"rest_total_hits_as_int": true,
						"search_type":            "query_then_fetch",
						"indices":                []any{},
					}),
					"logging": map[string]any{
						"text": "watch fired",
					},
				},
			},
		},
		{
			name: "script object omitting lang gets painless",
			input: map[string]any{
				"script": map[string]any{
					"source": "return true",
				},
			},
			expected: map[string]any{
				"script": map[string]any{
					"source": "return true",
					"lang":   "painless",
				},
			},
		},
		{
			name: "object with source but no script key is not treated as a script",
			input: map[string]any{
				"source": "return true",
				"id":     "not-a-script",
			},
			expected: map[string]any{
				"source": "return true",
				"id":     "not-a-script",
			},
		},
		{
			name: "explicit non-painless lang is preserved",
			input: map[string]any{
				"script": map[string]any{
					"source": "return true",
					"lang":   "mustache",
				},
			},
			expected: map[string]any{
				"script": map[string]any{
					"source": "return true",
					"lang":   "mustache",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := populateWatcherJSONDefaults(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func Test_populateWatcherJSONDefaults_copyOnWrite(t *testing.T) {
	t.Parallel()

	request := map[string]any{
		"body": map[string]any{
			"query": map[string]any{"match_all": map[string]any{}},
		},
	}
	search := map[string]any{"request": request}
	input := map[string]any{"search": search}

	got := populateWatcherJSONDefaults(input)

	_, hasTotalHits := request["rest_total_hits_as_int"]
	assert.False(t, hasTotalHits, "original request must not be mutated")
	_, hasSearchType := request["search_type"]
	assert.False(t, hasSearchType)
	_, hasIndices := request["indices"]
	assert.False(t, hasIndices)
	assert.Equal(t, search, input["search"])
	assert.Equal(t, request, search["request"])

	got["marker"] = true
	_, leaked := input["marker"]
	assert.False(t, leaked, "result must be a different map than the input")

	gotSearch, ok := got["search"].(map[string]any)
	require.True(t, ok)
	gotSearch["marker"] = true
	_, leaked = search["marker"]
	assert.False(t, leaked, "result search object must be a copy")

	gotRequest, ok := gotSearch["request"].(map[string]any)
	require.True(t, ok)
	gotRequest["marker"] = true
	_, leaked = request["marker"]
	assert.False(t, leaked, "result request object must be a copy")
	assert.Equal(t, true, gotRequest["rest_total_hits_as_int"])
	assert.Equal(t, "query_then_fetch", gotRequest["search_type"])
	assert.Equal(t, []any{}, gotRequest["indices"])
}

func Test_populateWatcherJSONDefaults_unchangedTreeReturnedAsIs(t *testing.T) {
	t.Parallel()

	input := searchRequest(map[string]any{
		"rest_total_hits_as_int": true,
		"search_type":            "query_then_fetch",
		"indices":                []any{},
	})
	got := populateWatcherJSONDefaults(input)
	got["marker"] = true
	_, shared := input["marker"]
	assert.True(t, shared, "unchanged input must be returned as the same map")
}

func TestJSONWithDefaultsValue_StringSemanticEquals_watcherDefaults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := []struct {
		name  string
		prior string
		next  string
		equal bool
	}{
		{
			name:  "omitted search-request keys including indices equal ES injected defaults",
			prior: `{"search":{"request":{"body":{"query":{"match_all":{}}}}}}`,
			next:  `{"search":{"request":{"body":{"query":{"match_all":{}}},"indices":[],"rest_total_hits_as_int":true,"search_type":"query_then_fetch"}}}`,
			equal: true,
		},
		{
			name:  "genuinely different search_type is unequal",
			prior: `{"search":{"request":{"body":{"query":{"match_all":{}}}}}}`,
			next:  `{"search":{"request":{"body":{"query":{"match_all":{}}},"indices":[],"rest_total_hits_as_int":true,"search_type":"dfs_query_then_fetch"}}}`,
			equal: false,
		},
		{
			name:  "redacted sentinel vs concrete secret is unequal",
			prior: `{"http":{"request":{"auth":{"basic":{"password":"::es_redacted::"}}}}}`,
			next:  `{"http":{"request":{"auth":{"basic":{"password":"super-secret"}}}}}`,
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			prior := customtypes.NewJSONWithDefaultsValue(tt.prior, populateWatcherJSONDefaults)
			next := customtypes.NewJSONWithDefaultsValue(tt.next, populateWatcherJSONDefaults)
			equal, diags := prior.StringSemanticEquals(ctx, next)
			require.False(t, diags.HasError(), "diags: %v", diags)
			assert.Equal(t, tt.equal, equal)
		})
	}
}
