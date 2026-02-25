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
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newRegionMapPanelConfigConverter(t *testing.T) {
	converter := newRegionMapPanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, string(kbapi.RegionMapNoESQLTypeRegionMap), converter.visualizationType)
}

func Test_regionMapConfigModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name         string
		apiNoESQL    *kbapi.RegionMapNoESQL
		apiESQL      *kbapi.RegionMapESQL
		expectESQL   bool
		expectQuery  bool
		expectTitle  string
		expectSample float64
	}{
		{
			name: "noesql region map",
			apiNoESQL: func() *kbapi.RegionMapNoESQL {
				api := kbapi.RegionMapNoESQL{
					Type:                kbapi.RegionMapNoESQLTypeRegionMap,
					Title:               new("Region Map"),
					Description:         new("Region map description"),
					IgnoreGlobalFilters: new(true),
					Sampling:            new(float32(0.75)),
				}

				lang := kbapi.FilterSimpleSchemaLanguage("kuery")
				api.Query = kbapi.FilterSimpleSchema{
					Language: &lang,
					Query:    "*",
				}

				_ = json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.Dataset)
				_ = json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric)
				_ = json.Unmarshal([]byte(`{"operation":"filters","filters":[{"filter":{"query":"*","language":"kuery"},"label":"All"}]}`), &api.Region)

				filter := kbapi.SearchFilterSchema0{}
				var query kbapi.SearchFilterSchema_0_Query
				_ = query.FromSearchFilterSchema0Query0("status:active")
				filter.Query = query
				var filterSchema kbapi.SearchFilterSchema
				_ = filterSchema.FromSearchFilterSchema0(filter)
				api.Filters = &[]kbapi.SearchFilterSchema{filterSchema}

				return &api
			}(),
			expectESQL:   false,
			expectQuery:  true,
			expectTitle:  "Region Map",
			expectSample: 0.75,
		},
		{
			name: "esql region map",
			apiESQL: func() *kbapi.RegionMapESQL {
				api := kbapi.RegionMapESQL{
					Type:                kbapi.RegionMapESQLTypeRegionMap,
					Title:               new("ESQL Region Map"),
					Description:         new("ESQL description"),
					IgnoreGlobalFilters: new(false),
					Sampling:            new(float32(0.25)),
				}

				_ = json.Unmarshal([]byte(`{"type":"esql","query":"FROM metrics-* | LIMIT 10"}`), &api.Dataset)
				_ = json.Unmarshal([]byte(`{"operation":"value","column":"value","format":{"id":"number"}}`), &api.Metric)
				_ = json.Unmarshal([]byte(`{"operation":"value","column":"region","ems":{"boundaries":"world_countries","join":"name"}}`), &api.Region)

				filter := kbapi.SearchFilterSchema0{}
				var query kbapi.SearchFilterSchema_0_Query
				_ = query.FromSearchFilterSchema0Query0("region:US")
				filter.Query = query
				var filterSchema kbapi.SearchFilterSchema
				_ = filterSchema.FromSearchFilterSchema0(filter)
				api.Filters = &[]kbapi.SearchFilterSchema{filterSchema}

				return &api
			}(),
			expectESQL:   true,
			expectQuery:  false,
			expectTitle:  "ESQL Region Map",
			expectSample: 0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &regionMapConfigModel{}

			if tt.apiNoESQL != nil {
				diags := model.fromAPINoESQL(context.Background(), *tt.apiNoESQL)
				require.False(t, diags.HasError())
			} else if tt.apiESQL != nil {
				diags := model.fromAPIESQL(context.Background(), *tt.apiESQL)
				require.False(t, diags.HasError())
			}

			assert.Equal(t, types.StringValue(tt.expectTitle), model.Title)
			assert.False(t, model.DatasetJSON.IsNull())
			assert.False(t, model.MetricJSON.IsNull())
			assert.False(t, model.RegionJSON.IsNull())

			if tt.expectQuery {
				require.NotNil(t, model.Query)
				assert.Equal(t, types.StringValue("*"), model.Query.Query)
			} else {
				assert.Nil(t, model.Query)
			}

			apiSchema, diags := model.toAPI()
			require.False(t, diags.HasError())

			if tt.expectESQL {
				apiESQL, err := apiSchema.AsRegionMapESQL()
				require.NoError(t, err)
				require.NotNil(t, apiESQL.Title)
				assert.Equal(t, tt.expectTitle, *apiESQL.Title)
				require.NotNil(t, apiESQL.Sampling)
				assert.InDelta(t, tt.expectSample, *apiESQL.Sampling, 0.001)
			} else {
				apiNoESQL, err := apiSchema.AsRegionMapNoESQL()
				require.NoError(t, err)
				require.NotNil(t, apiNoESQL.Title)
				assert.Equal(t, tt.expectTitle, *apiNoESQL.Title)
				require.NotNil(t, apiNoESQL.Sampling)
				assert.InDelta(t, tt.expectSample, *apiNoESQL.Sampling, 0.001)
				assert.Equal(t, "*", apiNoESQL.Query.Query)
			}
		})
	}
}
