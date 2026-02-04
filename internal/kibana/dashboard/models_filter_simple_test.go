package dashboard

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_filterSimpleModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name     string
		apiQuery kbapi.FilterSimpleSchema
		expected *filterSimpleModel
	}{
		{
			name: "all fields populated",
			apiQuery: kbapi.FilterSimpleSchema{
				Query:    "test query",
				Language: func() *kbapi.FilterSimpleSchemaLanguage { l := kbapi.FilterSimpleSchemaLanguage("kuery"); return &l }(),
			},
			expected: &filterSimpleModel{
				Query:    types.StringValue("test query"),
				Language: types.StringValue("kuery"),
			},
		},
		{
			name: "only required field",
			apiQuery: kbapi.FilterSimpleSchema{
				Query:    "simple query",
				Language: nil,
			},
			expected: &filterSimpleModel{
				Query:    types.StringValue("simple query"),
				Language: types.StringNull(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &filterSimpleModel{}
			model.fromAPI(tt.apiQuery)

			assert.Equal(t, tt.expected.Query, model.Query)
			assert.Equal(t, tt.expected.Language, model.Language)

			// Test toAPI
			apiQuery := model.toAPI()
			assert.Equal(t, tt.apiQuery.Query, apiQuery.Query)
		})
	}
}
