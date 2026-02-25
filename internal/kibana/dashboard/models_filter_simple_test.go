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
				Language: types.StringValue("kuery"),
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
