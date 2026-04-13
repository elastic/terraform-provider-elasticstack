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

				err := json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.DataSource)
				require.NoError(t, err)
				err = json.Unmarshal([]byte(`{"expression":"status:active","language":"kql"}`), &api.Query)
				require.NoError(t, err)
				err = json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric)
				require.NoError(t, err)

				var shape kbapi.GaugeNoESQL_Shape
				err = json.Unmarshal([]byte(`{"type":"circle"}`), &shape)
				require.NoError(t, err)
				api.Shape = &shape

				var fItem kbapi.LensPanelFilters_Item
				err = json.Unmarshal([]byte(`{"type":"condition","condition":{"field":"host.name","operator":"is","value":"foo"}}`), &fItem)
				require.NoError(t, err)
				filters := []kbapi.LensPanelFilters_Item{fItem}
				api.Filters = filters

				return api
			}(),
			expected: &gaugeConfigModel{
				Title:               types.StringValue("Test Gauge"),
				Description:         types.StringValue("A test gauge description"),
				IgnoreGlobalFilters: types.BoolValue(true),
				Sampling:            types.Float64Value(0.5),
				Query: &filterSimpleModel{
					Language:   types.StringValue("kql"),
					Expression: types.StringValue("status:active"),
				},
			},
		},
		{
			name: "minimal gauge config",
			api: func() kbapi.GaugeNoESQL {
				api := kbapi.GaugeNoESQL{
					Type: kbapi.GaugeNoESQLTypeGauge,
				}

				err := json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.DataSource)
				require.NoError(t, err)
				err = json.Unmarshal([]byte(`{"expression":"*"}`), &api.Query)
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
					Language:   types.StringValue("kql"), // Language should default to "kql"
					Expression: types.StringValue("*"),
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
				assert.Equal(t, tt.expected.Query.Expression, model.Query.Expression, "Query text should match")
			}

			assert.False(t, model.DataSourceJSON.IsNull(), "Dataset should not be null")
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

func Test_gaugePanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip(t *testing.T) {
	ctx := context.Background()

	api := kbapi.GaugeNoESQL{
		Type:                kbapi.GaugeNoESQLTypeGauge,
		Title:               new("Round-Trip Gauge"),
		Description:         new("Converter round-trip test"),
		IgnoreGlobalFilters: new(true),
		Sampling:            new(float32(0.5)),
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"dataView","id":"metrics-*"}`), &api.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"expression":"status:active","language":"kql"}`), &api.Query))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count"}`), &api.Metric))

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromGaugeNoESQL(api))

	converter := newGaugePanelConfigConverter()
	pm := &panelModel{}
	diags := converter.populateFromAttributes(ctx, pm, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.GaugeConfig)

	attrs2, diags := converter.buildAttributes(*pm)
	require.False(t, diags.HasError())

	gaugeNoESQL2, err := attrs2.AsGaugeNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "Round-Trip Gauge", *gaugeNoESQL2.Title)
	assert.Equal(t, "Converter round-trip test", *gaugeNoESQL2.Description)
	assert.True(t, *gaugeNoESQL2.IgnoreGlobalFilters)
	assert.InDelta(t, 0.5, *gaugeNoESQL2.Sampling, 0.001)
}
