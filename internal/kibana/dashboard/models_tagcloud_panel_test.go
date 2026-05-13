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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_tagcloudPanelConfigConverter_populateFromAttributes_buildAttributes_roundTrip(t *testing.T) {
	ctx := context.Background()

	api := kbapi.TagcloudNoESQL{
		Type:        "tagcloud",
		Title:       new("Round-Trip Tagcloud"),
		Description: new("Converter round-trip test"),
	}
	_ = json.Unmarshal([]byte(`{"index":"test-index"}`), &api.DataSource)
	_ = json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query)
	_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric)
	_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"tags.keyword"}`), &api.TagBy)

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromTagcloudNoESQL(api))

	converter := newTagcloudPanelConfigConverter()
	visBv := models.VisByValueModel{}
	diags := converter.populateFromAttributes(ctx, nil, nil, &visBv.LensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, visBv.TagcloudConfig)

	attrs2, diags := converter.buildAttributes(&visBv.LensByValueChartBlocks, nil)
	require.False(t, diags.HasError())

	tagcloudNoESQL2, err := attrs2.AsTagcloudNoESQL()
	require.NoError(t, err)
	assert.Equal(t, "Round-Trip Tagcloud", *tagcloudNoESQL2.Title)
	assert.Equal(t, "Converter round-trip test", *tagcloudNoESQL2.Description)
}

func Test_newTagcloudPanelConfigConverter(t *testing.T) {
	converter := newTagcloudPanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, "tag_cloud", converter.visualizationType)
}

func Test_tagcloudConfigModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name     string
		api      kbapi.TagcloudNoESQL
		expected *models.TagcloudConfigModel
	}{
		{
			name: "full tagcloud config",
			api: func() kbapi.TagcloudNoESQL {
				api := kbapi.TagcloudNoESQL{
					Type:                "tagcloud",
					Title:               new("Test Tagcloud"),
					Description:         new("A test tagcloud description"),
					IgnoreGlobalFilters: new(true),
					Sampling:            new(float32(0.5)),
					Styling: kbapi.TagcloudStyling{
						Orientation: kbapi.VisApiOrientation("horizontal"),
						FontSize: &struct {
							Max *float32 `json:"max,omitempty"`
							Min *float32 `json:"min,omitempty"`
						}{
							Min: new(float32(18)),
							Max: new(float32(72)),
						},
					},
				}

				// Set dataset as JSON
				_ = json.Unmarshal([]byte(`{"index":"test-index"}`), &api.DataSource)
				// Set query as JSON
				_ = json.Unmarshal([]byte(`{"expression":"status:active","language":"kql"}`), &api.Query)
				// Set metric as JSON
				_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric)
				// Set tagBy as JSON
				_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"tags.keyword","size":10}`), &api.TagBy)

				return api
			}(),
			expected: &models.TagcloudConfigModel{
				Title:               types.StringValue("Test Tagcloud"),
				Description:         types.StringValue("A test tagcloud description"),
				IgnoreGlobalFilters: types.BoolValue(true),
				Sampling:            types.Float64Value(0.5),
				Query: &models.FilterSimpleModel{
					Language:   types.StringValue("kql"),
					Expression: types.StringValue("status:active"),
				},
				Orientation: types.StringValue("horizontal"),
				FontSize: &models.FontSizeModel{
					Min: types.Float64Value(18),
					Max: types.Float64Value(72),
				},
			},
		},
		{
			name: "minimal tagcloud config",
			api: func() kbapi.TagcloudNoESQL {
				api := kbapi.TagcloudNoESQL{
					Type: "tagcloud",
				}

				// Set dataset as JSON
				_ = json.Unmarshal([]byte(`{"index":"test-index"}`), &api.DataSource)
				// Set query as JSON
				_ = json.Unmarshal([]byte(`{"expression":"*"}`), &api.Query)
				// Set metric as JSON
				_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric)
				// Set tagBy as JSON
				_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"tags.keyword"}`), &api.TagBy)

				return api
			}(),
			expected: &models.TagcloudConfigModel{
				Title:               types.StringNull(),
				Description:         types.StringNull(),
				IgnoreGlobalFilters: types.BoolNull(),
				Sampling:            types.Float64Null(),
				Query: &models.FilterSimpleModel{
					Language:   types.StringValue("kql"),
					Expression: types.StringValue("*"),
				},
				Orientation: types.StringNull(),
				FontSize:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &models.TagcloudConfigModel{}
			diags := tagcloudConfigFromAPI(context.Background(), model, nil, nil, tt.api)
			require.False(t, diags.HasError(), "fromAPI should not return errors")

			// Validate expected fields
			assert.Equal(t, tt.expected.Title, model.Title, "Title should match")
			assert.Equal(t, tt.expected.Description, model.Description, "Description should match")
			assert.Equal(t, tt.expected.IgnoreGlobalFilters, model.IgnoreGlobalFilters, "IgnoreGlobalFilters should match")
			assert.Equal(t, tt.expected.Sampling, model.Sampling, "Sampling should match")
			assert.Equal(t, tt.expected.Orientation, model.Orientation, "Orientation should match")

			// Validate query
			if tt.expected.Query != nil {
				require.NotNil(t, model.Query, "Query should not be nil")
				assert.Equal(t, tt.expected.Query.Language, model.Query.Language, "Query language should match")
				assert.Equal(t, tt.expected.Query.Expression, model.Query.Expression, "Query text should match")
			}

			// Validate font size
			if tt.expected.FontSize != nil {
				require.NotNil(t, model.FontSize, "FontSize should not be nil")
				assert.Equal(t, tt.expected.FontSize.Min, model.FontSize.Min, "FontSize.Min should match")
				assert.Equal(t, tt.expected.FontSize.Max, model.FontSize.Max, "FontSize.Max should match")
			} else {
				assert.Nil(t, model.FontSize, "FontSize should be nil")
			}

			// Validate dataset is not null
			assert.False(t, model.DataSourceJSON.IsNull(), "Dataset should not be null")

			// Validate metric and tagBy exist when present in API
			if tt.name == "full tagcloud config" || tt.name == "minimal tagcloud config" {
				// These should have metric and tagBy JSON
				assert.False(t, model.MetricJSON.IsNull(), "Metric should not be null")
				assert.False(t, model.TagByJSON.IsNull(), "TagBy should not be null")
			}

			// Test toAPI round-trip
			attrsResult, diags := tagcloudConfigToAPI(model, nil)
			require.False(t, diags.HasError(), "toAPI should not return errors")
			apiResult, err := attrsResult.AsTagcloudNoESQL()
			require.NoError(t, err)

			// Validate round-trip for basic fields
			if tt.api.Title != nil {
				require.NotNil(t, apiResult.Title, "Round-trip Title should not be nil")
				assert.Equal(t, *tt.api.Title, *apiResult.Title, "Round-trip Title should match")
			}

			if tt.api.Description != nil {
				require.NotNil(t, apiResult.Description, "Round-trip Description should not be nil")
				assert.Equal(t, *tt.api.Description, *apiResult.Description, "Round-trip Description should match")
			}

			if tt.api.IgnoreGlobalFilters != nil {
				require.NotNil(t, apiResult.IgnoreGlobalFilters, "Round-trip IgnoreGlobalFilters should not be nil")
				assert.Equal(t, *tt.api.IgnoreGlobalFilters, *apiResult.IgnoreGlobalFilters, "Round-trip IgnoreGlobalFilters should match")
			}

			if tt.api.Sampling != nil {
				require.NotNil(t, apiResult.Sampling, "Round-trip Sampling should not be nil")
				assert.InDelta(t, *tt.api.Sampling, *apiResult.Sampling, 0.001, "Round-trip Sampling should match")
			}

			if tt.api.Styling.Orientation != "" {
				assert.Equal(t, tt.api.Styling.Orientation, apiResult.Styling.Orientation, "Round-trip Orientation should match")
			}
		})
	}
}

func Test_fontSizeModel_roundTrip(t *testing.T) {
	tests := []struct {
		name    string
		apiFont *struct {
			Max *float32 `json:"max,omitempty"`
			Min *float32 `json:"min,omitempty"`
		}
	}{
		{
			name: "both min and max",
			apiFont: &struct {
				Max *float32 `json:"max,omitempty"`
				Min *float32 `json:"min,omitempty"`
			}{
				Min: new(float32(10)),
				Max: new(float32(100)),
			},
		},
		{
			name: "only min",
			apiFont: &struct {
				Max *float32 `json:"max,omitempty"`
				Min *float32 `json:"min,omitempty"`
			}{
				Min: new(float32(15)),
			},
		},
		{
			name: "only max",
			apiFont: &struct {
				Max *float32 `json:"max,omitempty"`
				Min *float32 `json:"min,omitempty"`
			}{
				Max: new(float32(80)),
			},
		},
		{
			name: "empty font size",
			apiFont: &struct {
				Max *float32 `json:"max,omitempty"`
				Min *float32 `json:"min,omitempty"`
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a full tagcloud schema with the font size
			api := kbapi.TagcloudNoESQL{
				Type: "tagcloud",
				Styling: kbapi.TagcloudStyling{
					FontSize: tt.apiFont,
				},
			}

			// Set dataset as JSON
			_ = json.Unmarshal([]byte(`{"index":"test-index"}`), &api.DataSource)
			// Set query as JSON
			_ = json.Unmarshal([]byte(`{"expression":"*"}`), &api.Query)
			// Set metric as JSON
			_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric)
			// Set tagBy as JSON
			_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"tags.keyword"}`), &api.TagBy)

			// Convert to model
			model := &models.TagcloudConfigModel{}
			diags := tagcloudConfigFromAPI(context.Background(), model, nil, nil, api)
			require.False(t, diags.HasError())

			// Convert back to API
			attrsResult, diags := tagcloudConfigToAPI(model, nil)
			require.False(t, diags.HasError())
			apiResult, err := attrsResult.AsTagcloudNoESQL()
			require.NoError(t, err)

			// Verify font size round-trip
			if tt.apiFont.Min != nil {
				require.NotNil(t, apiResult.Styling.FontSize)
				require.NotNil(t, apiResult.Styling.FontSize.Min)
				assert.InDelta(t, *tt.apiFont.Min, *apiResult.Styling.FontSize.Min, 0.001)
			}

			if tt.apiFont.Max != nil {
				require.NotNil(t, apiResult.Styling.FontSize)
				require.NotNil(t, apiResult.Styling.FontSize.Max)
				assert.InDelta(t, *tt.apiFont.Max, *apiResult.Styling.FontSize.Max, 0.001)
			}
		})
	}
}

func Test_tagcloudConfig_JSONFields(t *testing.T) {
	tests := []struct {
		name        string
		datasetJSON string
		metricJSON  string
		tagByJSON   string
		wantError   bool
	}{
		{
			name:        "valid JSON fields",
			datasetJSON: `{"index":"logs-*"}`,
			metricJSON:  `{"operation":{"operation_type":"count"}}`,
			tagByJSON:   `{"operation":{"operation_type":"terms"},"field":"user.keyword","size":20}`,
			wantError:   false,
		},
		{
			name:        "invalid dataset JSON",
			datasetJSON: `{invalid json}`,
			metricJSON:  `{"operation":{"operation_type":"count"}}`,
			tagByJSON:   `{"operation":{"operation_type":"terms"},"field":"user.keyword"}`,
			wantError:   true,
		},
		{
			name:        "invalid metric JSON",
			datasetJSON: `{"index":"logs-*"}`,
			metricJSON:  `{invalid json}`,
			tagByJSON:   `{"operation":{"operation_type":"terms"},"field":"user.keyword"}`,
			wantError:   true,
		},
		{
			name:        "invalid tagBy JSON",
			datasetJSON: `{"index":"logs-*"}`,
			metricJSON:  `{"operation":{"operation_type":"count"}}`,
			tagByJSON:   `{invalid json}`,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &models.TagcloudConfigModel{
				Title:          types.StringValue("Test"),
				Description:    types.StringValue("Test description"),
				DataSourceJSON: jsontypes.NewNormalizedValue(tt.datasetJSON),
				Query: &models.FilterSimpleModel{
					Expression: types.StringValue("*"),
				},
				MetricJSON: customtypes.NewJSONWithDefaultsValue[map[string]any](tt.metricJSON, populateTagcloudMetricDefaults),
				TagByJSON:  customtypes.NewJSONWithDefaultsValue[map[string]any](tt.tagByJSON, populateTagcloudTagByDefaults),
			}

			_, diags := tagcloudConfigToAPI(model, nil)
			if tt.wantError {
				assert.True(t, diags.HasError(), "Expected error for invalid JSON")
			} else {
				assert.False(t, diags.HasError(), "Expected no error for valid JSON")
			}
		})
	}
}

func Test_tagcloudConfigModel_fromAPIESQL_toAPIESQL_roundTrip(t *testing.T) {
	ctx := context.Background()

	api := kbapi.TagcloudESQL{Type: kbapi.TagcloudESQLTypeTagCloud}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM logs-* | STATS count = COUNT() BY host"}`), &api.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number"}`), &api.Metric.Format))
	api.Metric.Column = "count"
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number"}`), &api.TagBy.Format))
	require.NoError(t, json.Unmarshal([]byte(`{"mode":"categorical","palette":"default","mapping":[]}`), &api.TagBy.Color))
	api.TagBy.Column = "host"

	model := &models.TagcloudConfigModel{}
	diags := tagcloudConfigFromAPIESQL(ctx, model, nil, nil, api)
	require.False(t, diags.HasError(), "fromAPIESQL should not return errors: %v", diags)

	assert.Nil(t, model.Query)
	assert.True(t, tagcloudConfigUsesESQL(model))
	assert.True(t, model.MetricJSON.IsNull())
	assert.True(t, model.TagByJSON.IsNull())
	require.NotNil(t, model.EsqlMetric)
	require.NotNil(t, model.EsqlTagBy)
	assert.Equal(t, "count", model.EsqlMetric.Column.ValueString())
	assert.Equal(t, "host", model.EsqlTagBy.Column.ValueString())
	assert.JSONEq(t, `{"type":"number"}`, model.EsqlMetric.FormatJSON.ValueString())
	assert.JSONEq(t, `{"type":"number"}`, model.EsqlTagBy.FormatJSON.ValueString())
	assert.JSONEq(t, `{"mode":"categorical","palette":"default","mapping":[]}`, model.EsqlTagBy.ColorJSON.ValueString())

	attrs, diags := tagcloudConfigToAPI(model, nil)
	require.False(t, diags.HasError(), "toAPI should not return errors: %v", diags)
	out, err := attrs.AsTagcloudESQL()
	require.NoError(t, err)
	assert.Equal(t, kbapi.TagcloudESQLTypeTagCloud, out.Type)
	assert.Equal(t, "count", out.Metric.Column)
	assert.Equal(t, "host", out.TagBy.Column)
}

func Test_tagcloudConfigModel_toAPIESQL_requiresEsqlBlocks(t *testing.T) {
	t.Run("missing_esql_metric", func(t *testing.T) {
		m := &models.TagcloudConfigModel{
			DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM logs-*"}`),
		}
		_, diags := tagcloudConfigToAPIESQL(m, nil)
		require.True(t, diags.HasError(), "expected error when esql_metric is missing")
	})
	t.Run("missing_esql_tag_by", func(t *testing.T) {
		m := &models.TagcloudConfigModel{
			DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"esql","query":"FROM logs-*"}`),
			EsqlMetric: &models.TagcloudEsqlMetric{
				Column:     types.StringValue("c"),
				FormatJSON: jsontypes.NewNormalizedValue(`{"type":"number"}`),
			},
		}
		_, diags := tagcloudConfigToAPIESQL(m, nil)
		require.True(t, diags.HasError(), "expected error when esql_tag_by is missing")
		found := false
		for _, d := range diags {
			if d.Summary() == "Missing esql_tag_by" {
				found = true
				break
			}
		}
		require.True(t, found, "expected Missing esql_tag_by diagnostic, got %#v", diags)
	})
}

func Test_tagcloudPanelConfigConverter_routesESQL(t *testing.T) {
	ctx := context.Background()

	api := kbapi.TagcloudESQL{Type: kbapi.TagcloudESQLTypeTagCloud}
	require.NoError(t, json.Unmarshal([]byte(`{"type":"esql","query":"FROM logs-* | STATS c = COUNT() BY h"}`), &api.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number"}`), &api.Metric.Format))
	api.Metric.Column = "c"
	require.NoError(t, json.Unmarshal([]byte(`{"type":"number"}`), &api.TagBy.Format))
	require.NoError(t, json.Unmarshal([]byte(`{"mode":"categorical","palette":"default","mapping":[]}`), &api.TagBy.Color))
	api.TagBy.Column = "h"

	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	require.NoError(t, attrs.FromTagcloudESQL(api))

	converter := newTagcloudPanelConfigConverter()
	visBv := models.VisByValueModel{}
	diags := converter.populateFromAttributes(ctx, nil, nil, &visBv.LensByValueChartBlocks, attrs)
	require.False(t, diags.HasError())
	require.NotNil(t, visBv.TagcloudConfig)
	require.NotNil(t, visBv.TagcloudConfig.EsqlMetric)
	require.NotNil(t, visBv.TagcloudConfig.EsqlTagBy)
	assert.Nil(t, visBv.TagcloudConfig.Query)
}

func Test_tagcloudConfig_lensChartPresentation_hideTitleRoundTrip(t *testing.T) {
	ctx := context.Background()
	dash := lensPresentationTestDashboard()

	api := kbapi.TagcloudNoESQL{Type: "tagcloud"}
	require.NoError(t, json.Unmarshal([]byte(`{"index":"test-index"}`), &api.DataSource))
	require.NoError(t, json.Unmarshal([]byte(`{"expression":"*","language":"kql"}`), &api.Query))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric))
	require.NoError(t, json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"tags.keyword"}`), &api.TagBy))

	base := &models.TagcloudConfigModel{}
	require.False(t, tagcloudConfigFromAPI(ctx, base, nil, nil, api).HasError())

	m := *base
	m.HideTitle = types.BoolValue(true)

	attrs, diags := tagcloudConfigToAPI(&m, dash)
	require.False(t, diags.HasError())
	out, err := attrs.AsTagcloudNoESQL()
	require.NoError(t, err)

	got := &models.TagcloudConfigModel{}
	require.False(t, tagcloudConfigFromAPI(ctx, got, dash, &m, out).HasError())
	assert.Equal(t, types.BoolValue(true), got.HideTitle)
}
