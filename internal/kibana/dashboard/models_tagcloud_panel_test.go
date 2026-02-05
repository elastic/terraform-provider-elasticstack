package dashboard

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newTagcloudPanelConfigConverter(t *testing.T) {
	converter := newTagcloudPanelConfigConverter()
	assert.NotNil(t, converter)
	assert.Equal(t, "tagcloud", converter.visualizationType)
}

func Test_tagcloudConfigModel_fromAPI_toAPI(t *testing.T) {
	tests := []struct {
		name     string
		api      kbapi.TagcloudNoESQL
		expected *tagcloudConfigModel
	}{
		{
			name: "full tagcloud config",
			api: func() kbapi.TagcloudNoESQL {
				api := kbapi.TagcloudNoESQL{
					Type:                "tagcloud",
					Title:               utils.Pointer("Test Tagcloud"),
					Description:         utils.Pointer("A test tagcloud description"),
					IgnoreGlobalFilters: utils.Pointer(true),
					Sampling:            utils.Pointer(float32(0.5)),
					Orientation:         utils.Pointer(kbapi.TagcloudNoESQLOrientation("horizontal")),
					FontSize: &struct {
						Max *float32 `json:"max,omitempty"`
						Min *float32 `json:"min,omitempty"`
					}{
						Min: utils.Pointer(float32(18)),
						Max: utils.Pointer(float32(72)),
					},
				}

				// Set dataset as JSON
				_ = json.Unmarshal([]byte(`{"index":"test-index"}`), &api.Dataset)
				// Set query as JSON
				_ = json.Unmarshal([]byte(`{"query":"status:active","language":"kuery"}`), &api.Query)
				// Set metric as JSON
				_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric)
				// Set tagBy as JSON
				_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"tags.keyword","size":10}`), &api.TagBy)

				return api
			}(),
			expected: &tagcloudConfigModel{
				Title:               types.StringValue("Test Tagcloud"),
				Description:         types.StringValue("A test tagcloud description"),
				IgnoreGlobalFilters: types.BoolValue(true),
				Sampling:            types.Float64Value(0.5),
				Query: &filterSimpleModel{
					Language: types.StringValue("kuery"),
					Query:    types.StringValue("status:active"),
				},
				Orientation: types.StringValue("horizontal"),
				FontSize: &fontSizeModel{
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
				_ = json.Unmarshal([]byte(`{"index":"test-index"}`), &api.Dataset)
				// Set query as JSON
				_ = json.Unmarshal([]byte(`{"query":"*"}`), &api.Query)
				// Set metric as JSON
				_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric)
				// Set tagBy as JSON
				_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"tags.keyword"}`), &api.TagBy)

				return api
			}(),
			expected: &tagcloudConfigModel{
				Title:               types.StringNull(),
				Description:         types.StringNull(),
				IgnoreGlobalFilters: types.BoolNull(),
				Sampling:            types.Float64Null(),
				Query: &filterSimpleModel{
					Language: types.StringNull(),
					Query:    types.StringValue("*"),
				},
				Orientation: types.StringNull(),
				FontSize:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test fromAPI
			model := &tagcloudConfigModel{}
			diags := model.fromAPI(context.Background(), tt.api)
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
				assert.Equal(t, tt.expected.Query.Query, model.Query.Query, "Query text should match")
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
			assert.False(t, model.Dataset.IsNull(), "Dataset should not be null")

			// Validate metric and tagBy exist when present in API
			if tt.name == "full tagcloud config" || tt.name == "minimal tagcloud config" {
				// These should have metric and tagBy JSON
				assert.False(t, model.Metric.IsNull(), "Metric should not be null")
				assert.False(t, model.TagBy.IsNull(), "TagBy should not be null")
			}

			// Test toAPI round-trip
			apiResult, diags := model.toAPI()
			require.False(t, diags.HasError(), "toAPI should not return errors")

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

			if tt.api.Orientation != nil {
				require.NotNil(t, apiResult.Orientation, "Round-trip Orientation should not be nil")
				assert.Equal(t, *tt.api.Orientation, *apiResult.Orientation, "Round-trip Orientation should match")
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
				Min: utils.Pointer(float32(10)),
				Max: utils.Pointer(float32(100)),
			},
		},
		{
			name: "only min",
			apiFont: &struct {
				Max *float32 `json:"max,omitempty"`
				Min *float32 `json:"min,omitempty"`
			}{
				Min: utils.Pointer(float32(15)),
			},
		},
		{
			name: "only max",
			apiFont: &struct {
				Max *float32 `json:"max,omitempty"`
				Min *float32 `json:"min,omitempty"`
			}{
				Max: utils.Pointer(float32(80)),
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
				Type:     "tagcloud",
				FontSize: tt.apiFont,
			}

			// Set dataset as JSON
			_ = json.Unmarshal([]byte(`{"index":"test-index"}`), &api.Dataset)
			// Set query as JSON
			_ = json.Unmarshal([]byte(`{"query":"*"}`), &api.Query)
			// Set metric as JSON
			_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"count"}}`), &api.Metric)
			// Set tagBy as JSON
			_ = json.Unmarshal([]byte(`{"operation":{"operation_type":"terms"},"field":"tags.keyword"}`), &api.TagBy)

			// Convert to model
			model := &tagcloudConfigModel{}
			diags := model.fromAPI(context.Background(), api)
			require.False(t, diags.HasError())

			// Convert back to API
			apiResult, diags := model.toAPI()
			require.False(t, diags.HasError())

			// Verify font size round-trip
			if tt.apiFont.Min != nil {
				require.NotNil(t, apiResult.FontSize)
				require.NotNil(t, apiResult.FontSize.Min)
				assert.InDelta(t, *tt.apiFont.Min, *apiResult.FontSize.Min, 0.001)
			}

			if tt.apiFont.Max != nil {
				require.NotNil(t, apiResult.FontSize)
				require.NotNil(t, apiResult.FontSize.Max)
				assert.InDelta(t, *tt.apiFont.Max, *apiResult.FontSize.Max, 0.001)
			}
		})
	}
}

func Test_tagcloudPanelConfigConverter_handlesTFPanelConfig(t *testing.T) {
	converter := newTagcloudPanelConfigConverter()

	tests := []struct {
		name     string
		panel    panelModel
		expected bool
	}{
		{
			name: "has tagcloud config",
			panel: panelModel{
				TagcloudConfig: &tagcloudConfigModel{},
			},
			expected: true,
		},
		{
			name: "no tagcloud config",
			panel: panelModel{
				MarkdownConfig: &markdownConfigModel{},
			},
			expected: false,
		},
		{
			name:     "empty panel",
			panel:    panelModel{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.handlesTFPanelConfig(tt.panel)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_tagcloudPanelConfigConverter_mapPanelToAPI(t *testing.T) {
	converter := newTagcloudPanelConfigConverter()

	// Create a minimal tagcloud config
	datasetJSON := `{"index":"test-index"}`
	metricJSON := `{"operation":{"operation_type":"count"}}`
	tagByJSON := `{"operation":{"operation_type":"terms"},"field":"tags.keyword"}`

	panel := panelModel{
		Type: types.StringValue("lens"),
		TagcloudConfig: &tagcloudConfigModel{
			Title:       types.StringValue("Test Tagcloud"),
			Description: types.StringValue("Test description"),
			Dataset:     jsontypes.NewNormalizedValue(datasetJSON),
			Query: &filterSimpleModel{
				Language: types.StringValue("kuery"),
				Query:    types.StringValue("*"),
			},
			Metric: customtypes.NewJSONWithDefaultsValue[map[string]any](metricJSON, populateTagcloudMetricDefaults),
			TagBy:  customtypes.NewJSONWithDefaultsValue[map[string]any](tagByJSON, populateTagcloudTagByDefaults),
		},
	}

	var apiConfig kbapi.DashboardPanelItem_Config
	diags := converter.mapPanelToAPI(panel, &apiConfig)
	require.False(t, diags.HasError())

	// Verify the config was created
	configMap, err := apiConfig.AsDashboardPanelItemConfig2()
	require.NoError(t, err)

	// Verify the attributes exist
	attrs, ok := configMap["attributes"]
	require.True(t, ok, "attributes should exist in config")

	attrsMap, ok := attrs.(map[string]interface{})
	require.True(t, ok, "attributes should be a map")

	// Verify the type field exists with tagcloud
	typeField, ok := attrsMap["type"]
	require.True(t, ok, "type should exist")
	assert.Equal(t, "tagcloud", typeField)

	// Verify title exists in attributes
	title, ok := attrsMap["title"]
	require.True(t, ok, "title should exist in attributes")
	assert.Equal(t, "Test Tagcloud", title)
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
			model := &tagcloudConfigModel{
				Title:       types.StringValue("Test"),
				Description: types.StringValue("Test description"),
				Dataset:     jsontypes.NewNormalizedValue(tt.datasetJSON),
				Query: &filterSimpleModel{
					Query: types.StringValue("*"),
				},
				Metric: customtypes.NewJSONWithDefaultsValue[map[string]any](tt.metricJSON, populateTagcloudMetricDefaults),
				TagBy:  customtypes.NewJSONWithDefaultsValue[map[string]any](tt.tagByJSON, populateTagcloudTagByDefaults),
			}

			_, diags := model.toAPI()
			if tt.wantError {
				assert.True(t, diags.HasError(), "Expected error for invalid JSON")
			} else {
				assert.False(t, diags.HasError(), "Expected no error for valid JSON")
			}
		})
	}
}
