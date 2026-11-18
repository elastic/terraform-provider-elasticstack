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

package datafeed

import (
	"context"
	"encoding/json"
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/expandwildcard"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromAPIModel_ExpandWildcards(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		apiValue []expandwildcard.ExpandWildcard
		want     []string
	}{
		{
			name:     "empty list maps to none",
			apiValue: []expandwildcard.ExpandWildcard{},
			want:     []string{"none"},
		},
		{
			name:     "nil list maps to none",
			apiValue: nil,
			want:     []string{"none"},
		},
		{
			name:     "open and closed stored as-is",
			apiValue: []expandwildcard.ExpandWildcard{expandwildcard.Open, expandwildcard.Closed},
			want:     []string{"open", "closed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			api := &elasticsearch.MLDatafeedResponse{
				MLDatafeed: &estypes.MLDatafeed{
					DatafeedId: "df",
					JobId:      "job",
					Indices:    []string{"test-index-*"},
					IndicesOptions: &estypes.IndicesOptions{
						ExpandWildcards: tt.apiValue,
					},
				},
				QueryRaw: json.RawMessage(`{"match_all":{}}`),
			}

			var model Datafeed
			diags := model.FromAPIModel(context.Background(), api)
			require.False(t, diags.HasError(), "unexpected diagnostics: %s", diags)

			got := expandWildcardsFromModel(t, model)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func expandWildcardsFromModel(t *testing.T, model Datafeed) []string {
	t.Helper()

	require.False(t, model.IndicesOptions.IsNull())
	var io IndicesOptions
	diags := model.IndicesOptions.As(context.Background(), &io, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), "unexpected diagnostics: %s", diags)
	require.False(t, io.ExpandWildcards.IsNull())

	var tokens []string
	diags = io.ExpandWildcards.ElementsAs(context.Background(), &tokens, false)
	require.False(t, diags.HasError(), "unexpected diagnostics: %s", diags)
	return tokens
}
