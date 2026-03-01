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

package dashboard

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_searchFilterModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name        string
		apiFilter   kbapi.SearchFilter
		expected    *searchFilterModel
		expectError bool
	}{
		{
			name: "valid filter with language",
			apiFilter: func() kbapi.SearchFilter {
				filter := kbapi.SearchFilter0{
					Language: func() *kbapi.SearchFilter0Language { l := kbapi.SearchFilter0Language("lucene"); return &l }(),
				}
				var query kbapi.SearchFilter_0_Query
				_ = query.FromSearchFilter0Query0("field:value")
				filter.Query = query

				var result kbapi.SearchFilter
				_ = result.FromSearchFilter0(filter)
				return result
			}(),
			expected: &searchFilterModel{
				Query:    types.StringValue("field:value"),
				Language: types.StringValue("lucene"),
			},
			expectError: false,
		},
		{
			name: "filter without language",
			apiFilter: func() kbapi.SearchFilter {
				filter := kbapi.SearchFilter0{}
				var query kbapi.SearchFilter_0_Query
				_ = query.FromSearchFilter0Query0("simple query")
				filter.Query = query

				var result kbapi.SearchFilter
				_ = result.FromSearchFilter0(filter)
				return result
			}(),
			expected: &searchFilterModel{
				Query:    types.StringValue("simple query"),
				Language: types.StringValue("kuery"), // Defaults to kuery when API doesn't return it
			},
			expectError: false,
		},
		{
			name: "filter with ES DSL object query (FilterQueryType fallback)",
			apiFilter: func() kbapi.SearchFilter {
				filter := kbapi.SearchFilter0{
					Language: func() *kbapi.SearchFilter0Language { l := kbapi.SearchFilter0Language("kuery"); return &l }(),
				}
				var query kbapi.SearchFilter_0_Query
				_ = query.FromFilterQueryType(kbapi.FilterQueryType{
					Match: map[string]any{"field": map[string]any{"query": "value"}},
				})
				filter.Query = query

				var result kbapi.SearchFilter
				_ = result.FromSearchFilter0(filter)
				return result
			}(),
			expected: &searchFilterModel{
				Query:    types.StringValue(`{"bool":null,"exists":null,"match":{"field":{"query":"value"}},"match_phrase":null,"prefix":null,"range":null,"terms":null,"wildcard":null}`),
				Language: types.StringValue("kuery"),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &searchFilterModel{}
			diags := model.fromAPI(tt.apiFilter)

			if tt.expectError {
				assert.True(t, diags.HasError())
				return
			}

			require.False(t, diags.HasError())
			assert.Equal(t, tt.expected.Query, model.Query)
			assert.Equal(t, tt.expected.Language, model.Language)

			// Test toAPI
			apiFilter, diags := model.toAPI()
			require.False(t, diags.HasError())
			assert.NotNil(t, apiFilter)
		})
	}
}
