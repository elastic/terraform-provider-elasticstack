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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newGaugePanelConfigConverter(t *testing.T) {
	converter := newGaugePanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, string(kbapi.GaugeNoESQLTypeGauge), converter.visualizationType)
}

func Test_gaugeConfigModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name     string
		api      kbapi.GaugeNoESQL
		expected *gaugeConfigModel
	}{
		{
			name: "full gauge config",
			api: func() kbapi.GaugeNoESQL {
				api := kbapi.GaugeNoESQL{
					Type:                kbapi.GaugeNoESQLTypeGauge,
					Title:               new("Test Gauge"),
					Description:         new("A test gauge description"),
					IgnoreGlobalFilters: new(true),
					Sampling:            new(float32(0.5)),
				}

				err := json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.Dataset)
				require.NoError(t, err)
				err = json.Unmarshal([]byte(`{"query":"status:active","language":"kuery"}`), &api.Query)
				require.NoError(t, err)
				err = json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric)
				require.NoError(t, err)

				var shape kbapi.GaugeNoESQL_Shape
				err = json.Unmarshal([]byte(`{"type":"circle"}`), &shape)
				require.NoError(t, err)
				api.Shape = &shape

				filter := kbapi.SearchFilterSchema0{
					Language: func() *kbapi.SearchFilterSchema0Language { l := kbapi.SearchFilterSchema0Language("lucene"); return &l }(),
				}
				var query kbapi.SearchFilterSchema_0_Query
				err = query.FromSearchFilterSchema0Query0("host.name:foo")
				require.NoError(t, err)
				filter.Query = query

				var filterUnion kbapi.SearchFilterSchema
				err = filterUnion.FromSearchFilterSchema0(filter)
				require.NoError(t, err)
				filters := []kbapi.SearchFilterSchema{filterUnion}
				api.Filters = &filters

				return api
			}(),
			expected: &gaugeConfigModel{
				Title:               types.StringValue("Test Gauge"),
				Description:         types.StringValue("A test gauge description"),
				IgnoreGlobalFilters: types.BoolValue(true),
				Sampling:            types.Float64Value(0.5),
				Query: &filterSimpleModel{
					Language: types.StringValue("kuery"),
					Query:    types.StringValue("status:active"),
				},
			},
		},
		{
			name: "minimal gauge config",
			api: func() kbapi.GaugeNoESQL {
				api := kbapi.GaugeNoESQL{
					Type: kbapi.GaugeNoESQLTypeGauge,
				}

				err := json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.Dataset)
				require.NoError(t, err)
				err = json.Unmarshal([]byte(`{"query":"*"}`), &api.Query)
				require.NoError(t, err)
				err = json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric)
				require.NoError(t, err)

				return api
			}(),
			expected: &gaugeConfigModel{
				Title:               types.StringNull(),
				Description:         types.StringNull(),
				IgnoreGlobalFilters: types.BoolNull(),
				Sampling:            types.Float64Null(),
				Query: &filterSimpleModel{
					Language: types.StringValue("kuery"), // Language should default to "kuery"
					Query:    types.StringValue("*"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &gaugeConfigModel{}
			diags := model.fromAPI(context.Background(), tt.api)
			require.False(t, diags.HasError(), "fromAPI should not return errors")

			assert.Equal(t, tt.expected.Title, model.Title, "Title should match")
			assert.Equal(t, tt.expected.Description, model.Description, "Description should match")
			assert.Equal(t, tt.expected.IgnoreGlobalFilters, model.IgnoreGlobalFilters, "IgnoreGlobalFilters should match")
			assert.Equal(t, tt.expected.Sampling, model.Sampling, "Sampling should match")

			if tt.expected.Query != nil {
				require.NotNil(t, model.Query, "Query should not be nil")
				assert.Equal(t, tt.expected.Query.Language, model.Query.Language, "Query language should match")
				assert.Equal(t, tt.expected.Query.Query, model.Query.Query, "Query text should match")
			}

			assert.False(t, model.DatasetJSON.IsNull(), "Dataset should not be null")
			assert.False(t, model.MetricJSON.IsNull(), "Metric should not be null")

			if tt.name == "full gauge config" {
				assert.False(t, model.ShapeJSON.IsNull(), "Shape should not be null")
				assert.Len(t, model.Filters, 1, "Filters should be populated")
			}

			apiResult, diags := model.toAPI()
			require.False(t, diags.HasError(), "toAPI should not return errors")

			if tt.api.Title != nil {
				require.NotNil(t, apiResult.Title)
				assert.Equal(t, *tt.api.Title, *apiResult.Title)
			}

			if tt.api.Description != nil {
				require.NotNil(t, apiResult.Description)
				assert.Equal(t, *tt.api.Description, *apiResult.Description)
			}

			if tt.api.IgnoreGlobalFilters != nil {
				require.NotNil(t, apiResult.IgnoreGlobalFilters)
				assert.Equal(t, *tt.api.IgnoreGlobalFilters, *apiResult.IgnoreGlobalFilters)
			}

			if tt.api.Sampling != nil {
				require.NotNil(t, apiResult.Sampling)
				assert.InDelta(t, *tt.api.Sampling, *apiResult.Sampling, 0.001)
			}
		})
	}
}

func Test_gaugePanelConfigConverter_roundTrip(t *testing.T) {
	converter := newGaugePanelConfigConverter()
	ctx := context.Background()

	panel := panelModel{
		Type: types.StringValue("lens"),
		GaugeConfig: &gaugeConfigModel{
			Title:       types.StringValue("Round Trip Gauge"),
			Description: types.StringValue("Round-trip test"),
			DatasetJSON: jsontypes.NewNormalizedValue(`{"type":"dataView","id":"metrics-*"}`),
			Query: &filterSimpleModel{
				Language: types.StringValue("kuery"),
				Query:    types.StringValue("status:active"),
			},
			MetricJSON: customtypes.NewJSONWithDefaultsValue(`{"operation":"count"}`, populateGaugeMetricDefaults),
			ShapeJSON:  jsontypes.NewNormalizedValue(`{"type":"circle"}`),
		},
	}

	var apiConfig kbapi.DashboardPanelItem_Config
	diags := converter.mapPanelToAPI(panel, &apiConfig)
	require.False(t, diags.HasError())

	newPanel := panelModel{Type: types.StringValue("lens")}
	diags = converter.populateFromAPIPanel(ctx, &newPanel, apiConfig)
	require.False(t, diags.HasError())
	require.NotNil(t, newPanel.GaugeConfig)
	assert.Equal(t, types.StringValue("Round Trip Gauge"), newPanel.GaugeConfig.Title)
	assert.Equal(t, types.StringValue("Round-trip test"), newPanel.GaugeConfig.Description)
	assert.False(t, newPanel.GaugeConfig.DatasetJSON.IsNull())
	assert.False(t, newPanel.GaugeConfig.MetricJSON.IsNull())
	assert.False(t, newPanel.GaugeConfig.ShapeJSON.IsNull())
	require.NotNil(t, newPanel.GaugeConfig.Query)
	assert.Equal(t, types.StringValue("kuery"), newPanel.GaugeConfig.Query.Language)
	assert.Equal(t, types.StringValue("status:active"), newPanel.GaugeConfig.Query.Query)
}
