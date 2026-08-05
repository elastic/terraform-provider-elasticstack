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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_filterSimpleModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name         string
		apiQueryJSON string
		expected     *models.FilterSimpleModel
	}{
		{
			name:         "all fields populated",
			apiQueryJSON: `{"expression":"test query","language":"kql"}`,
			expected: &models.FilterSimpleModel{
				Expression: types.StringValue("test query"),
				Language:   types.StringValue("kql"),
			},
		},
		{
			name:         "only required field",
			apiQueryJSON: `{"expression":"simple query"}`,
			expected: &models.FilterSimpleModel{
				Expression: types.StringValue("simple query"),
				Language:   types.StringValue("kql"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inputQuery kbapi.KibanaHTTPAPIsFilterSimple
			require.NoError(t, json.Unmarshal([]byte(tt.apiQueryJSON), &inputQuery))

			// Test fromAPI
			model := &models.FilterSimpleModel{}
			filterSimpleFromAPI(model, &inputQuery)

			assert.Equal(t, tt.expected.Expression, model.Expression)
			assert.Equal(t, tt.expected.Language, model.Language)

			// Test toAPI
			outputQuery := filterSimpleToAPI(model)
			assert.Equal(t, inputQuery.Expression, outputQuery.Expression)
			require.NotNil(t, outputQuery.Language)
			var language string
			require.NoError(t, json.Unmarshal(mustMarshalJSON(t, outputQuery.Language), &language))
			assert.Equal(t, tt.expected.Language.ValueString(), language)
		})
	}
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	b, err := json.Marshal(value)
	require.NoError(t, err)
	return b
}
