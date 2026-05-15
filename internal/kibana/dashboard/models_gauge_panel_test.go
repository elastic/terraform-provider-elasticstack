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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_gaugeConfigModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name     string
		api      kbapi.GaugeNoESQL
		expected *models.GaugeConfigModel
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

				var shape kbapi.GaugeStyling_Shape
				err = json.Unmarshal([]byte(`{"type":"circle"}`), &shape)
				require.NoError(t, err)
				api.Styling.Shape = &shape

				var fItem kbapi.LensPanelFilters_Item
				err = json.Unmarshal([]byte(`{"type":"condition","condition":{"field":"host.name","operator":"is","value":"foo"}}`), &fItem)
				require.NoError(t, err)
				filters := []kbapi.LensPanelFilters_Item{fItem}
				api.Filters = filters

				return api
			}(),
			expected: &models.GaugeConfigModel{
				Title:               types.StringValue("Test Gauge"),
				Description:         types.StringValue("A test gauge description"),
				IgnoreGlobalFilters: types.BoolValue(true),
				Sampling:            types.Float64Value(0.5),
				Query: &models.FilterSimpleModel{
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
			expected: &models.GaugeConfigModel{
				Title:               types.StringNull(),
				Description:         types.StringNull(),
				IgnoreGlobalFilters: types.BoolNull(),
				Sampling:            types.Float64Null(),
				Query: &models.FilterSimpleModel{
					Language:   types.StringValue("kql"), // Language should default to "kql"
					Expression: types.StringValue("*"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &models.GaugeConfigModel{}
			diags := gaugeConfigFromAPI(context.Background(), model, nil, nil, tt.api)
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
				require.NotNil(t, model.Styling)
				assert.False(t, model.Styling.ShapeJSON.IsNull(), "Shape should not be null")
				assert.Len(t, model.Filters, 1, "Filters should be populated")
			}

			attrsResult, diags := gaugeConfigToAPI(model, nil)
			require.False(t, diags.HasError(), "toAPI should not return errors")
			apiResult, err := attrsResult.AsGaugeNoESQL()
			require.NoError(t, err)

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

func Test_gaugeConfig_lensChartPresentation_comprehensive(t *testing.T) {
	runGaugeNoESQLLensChartPresentationComprehensive(t)
}

func Test_gaugeConfigModel_fromAPIESQL_toAPIESQL_roundTrip(t *testing.T) {
	ctx := context.Background()

	api := kbapi.GaugeESQL{
		Type: kbapi.GaugeESQLTypeGauge,
	}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM metrics-* | STATS revenue = SUM(value)"}`), &api.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number"}`), &api.Metric.Format))
	api.Metric.Column = "revenue"
	label := "Revenue"
	api.Metric.Label = &label

	model := &models.GaugeConfigModel{}
	diags := gaugeConfigFromAPIESQL(ctx, model, nil, nil, api)
	require.False(t, diags.HasError(), "fromAPIESQL should not return errors: %v", diags)

	// Query should be nil for ES|QL mode.
	assert.Nil(t, model.Query, "Query should be nil for ES|QL")
	assert.True(t, gaugeConfigUsesESQL(model), "model should report ES|QL mode")
	// metric_json should be null in ES|QL mode; typed esql_metric populated.
	assert.True(t, model.MetricJSON.IsNull(), "MetricJSON should be null")
	require.NotNil(t, model.EsqlMetric)
	assert.Equal(t, "revenue", model.EsqlMetric.Column.ValueString())
	assert.Equal(t, "Revenue", model.EsqlMetric.Label.ValueString())
	assert.JSONEq(t, `{"type":"number"}`, model.EsqlMetric.FormatJSON.ValueString())

	// Round-trip via toAPI -> AsGaugeESQL.
	attrs, diags := gaugeConfigToAPI(model, nil)
	require.False(t, diags.HasError(), "toAPI should not return errors: %v", diags)
	out, err := attrs.AsGaugeESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.GaugeESQLTypeGauge, out.Type)
	assert.Equal(t, "revenue", out.Metric.Column)
	require.NotNil(t, out.Metric.Label)
	assert.Equal(t, "Revenue", *out.Metric.Label)
}

func Test_gaugeConfigModel_toAPIESQL_requiresEsqlMetric(t *testing.T) {
	m := &models.GaugeConfigModel{
		DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM metrics-*"}`),
	}
	_, diags := gaugeConfigToAPIESQL(m, nil)
	require.True(t, diags.HasError(), "expected error when esql_metric is missing")
}

func Test_gaugeConfigModel_fromAPIESQL_toAPIESQL_roundTrip_populatedOptionalMetricFields(t *testing.T) {
	ctx := context.Background()

	api := kbapi.GaugeESQL{Type: kbapi.GaugeESQLTypeGauge}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM metrics-* | STATS revenue = SUM(value)"}`), &api.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number","decimals":4}`), &api.Metric.Format))
	api.Metric.Column = "revenue"
	label := "Rev"
	api.Metric.Label = &label
	sub := "Subtitle"
	api.Metric.Subtitle = &sub
	require.NoError(t, json.Unmarshal([]byte(`{"type":"auto"}`), &api.Metric.Color))

	gl := "Goal label"
	api.Metric.Goal = &struct {
		Column string  `json:"column"`
		Label  *string `json:"label,omitempty"`
	}{Column: "goal_col", Label: &gl}

	ml := "Max label"
	api.Metric.Max = &struct {
		Column string  `json:"column"`
		Label  *string `json:"label,omitempty"`
	}{Column: "max_col", Label: &ml}

	minl := "Min label"
	api.Metric.Min = &struct {
		Column string  `json:"column"`
		Label  *string `json:"label,omitempty"`
	}{Column: "min_col", Label: &minl}

	mode := kbapi.GaugeESQLMetricTicksModeAuto
	tvis := false
	api.Metric.Ticks = &struct {
		Mode    *kbapi.GaugeESQLMetricTicksMode `json:"mode,omitempty"`
		Visible *bool                           `json:"visible,omitempty"`
	}{Mode: &mode, Visible: &tvis}

	tx := "Gauge metric title"
	titleVis := true
	api.Metric.Title = &struct {
		Text    *string `json:"text,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	}{Text: &tx, Visible: &titleVis}

	m := &models.GaugeConfigModel{}
	diags := gaugeConfigFromAPIESQL(ctx, m, nil, nil, api)
	require.False(t, diags.HasError(), "%v", diags)

	attrs, diags := gaugeConfigToAPI(m, nil)
	require.False(t, diags.HasError(), "%v", diags)
	apiOut, err := attrs.AsGaugeESQL()
	require.NoError(t, err)

	wantMetric, err := json.Marshal(api.Metric)
	require.NoError(t, err)
	gotMetric, err := json.Marshal(apiOut.Metric)
	require.NoError(t, err)
	assert.JSONEq(t, string(wantMetric), string(gotMetric))

	m2 := &models.GaugeConfigModel{}
	diags = gaugeConfigFromAPIESQL(ctx, m2, nil, nil, apiOut)
	require.False(t, diags.HasError(), "%v", diags)

	attrs2, diags := gaugeConfigToAPI(m2, nil)
	require.False(t, diags.HasError(), "%v", diags)
	apiOut2, err := attrs2.AsGaugeESQL()
	require.NoError(t, err)
	gotMetric2, err := json.Marshal(apiOut2.Metric)
	require.NoError(t, err)
	assert.JSONEq(t, string(wantMetric), string(gotMetric2))
}
