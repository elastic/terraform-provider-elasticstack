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
		apiFilter   kbapi.SearchFilterSchema
		expected    *searchFilterModel
		expectError bool
	}{
		{
			name: "valid filter with language",
			apiFilter: func() kbapi.SearchFilterSchema {
				filter := kbapi.SearchFilterSchema0{
					Language: func() *kbapi.SearchFilterSchema0Language { l := kbapi.SearchFilterSchema0Language("lucene"); return &l }(),
				}
				var query kbapi.SearchFilterSchema_0_Query
				_ = query.FromSearchFilterSchema0Query0("field:value")
				filter.Query = query

				var result kbapi.SearchFilterSchema
				_ = result.FromSearchFilterSchema0(filter)
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
			apiFilter: func() kbapi.SearchFilterSchema {
				filter := kbapi.SearchFilterSchema0{}
				var query kbapi.SearchFilterSchema_0_Query
				_ = query.FromSearchFilterSchema0Query0("simple query")
				filter.Query = query

				var result kbapi.SearchFilterSchema
				_ = result.FromSearchFilterSchema0(filter)
				return result
			}(),
			expected: &searchFilterModel{
				Query:    types.StringValue("simple query"),
				Language: types.StringNull(),
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
