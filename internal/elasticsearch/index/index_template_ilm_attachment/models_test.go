package index_template_ilm_attachment

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestGetComponentTemplateName(t *testing.T) {
	model := tfModel{
		IndexTemplate: types.StringValue("logs-system.syslog"),
	}
	result := model.getComponentTemplateName()
	assert.Equal(t, "logs-system.syslog@custom", result)
}

func TestMergeILMSetting_EmptyExisting(t *testing.T) {
	existing := map[string]interface{}{}
	result := mergeILMSetting(existing, "my-policy")

	indexVal := result["index"].(map[string]interface{})
	lifecycleVal := indexVal["lifecycle"].(map[string]interface{})
	assert.Equal(t, "my-policy", lifecycleVal["name"])
	assert.Len(t, result, 1)
}

func TestMergeILMSetting_NilExisting(t *testing.T) {
	result := mergeILMSetting(nil, "my-policy")

	indexVal := result["index"].(map[string]interface{})
	lifecycleVal := indexVal["lifecycle"].(map[string]interface{})
	assert.Equal(t, "my-policy", lifecycleVal["name"])
	assert.Len(t, result, 1)
}

func TestMergeILMSetting_PreserveExisting(t *testing.T) {
	// Nested form as returned by Elasticsearch
	existing := map[string]interface{}{
		"index": map[string]interface{}{
			"number_of_replicas": 2,
			"refresh_interval":  "30s",
		},
	}
	result := mergeILMSetting(existing, "my-policy")

	indexVal := result["index"].(map[string]interface{})
	lifecycleVal := indexVal["lifecycle"].(map[string]interface{})
	assert.Equal(t, "my-policy", lifecycleVal["name"])
	assert.Equal(t, 2, indexVal["number_of_replicas"])
	assert.Equal(t, "30s", indexVal["refresh_interval"])
	assert.Len(t, indexVal, 3)
	assert.Len(t, result, 1)
}

func TestMergeILMSetting_OverwriteExistingILM(t *testing.T) {
	existing := map[string]interface{}{
		"index": map[string]interface{}{
			"lifecycle":         map[string]interface{}{"name": "old-policy"},
			"number_of_replicas": 2,
		},
	}
	result := mergeILMSetting(existing, "new-policy")

	indexVal := result["index"].(map[string]interface{})
	lifecycleVal := indexVal["lifecycle"].(map[string]interface{})
	assert.Equal(t, "new-policy", lifecycleVal["name"])
	assert.Equal(t, 2, indexVal["number_of_replicas"])
	assert.Len(t, indexVal, 2)
	assert.Len(t, result, 1)
}

func TestRemoveILMSetting_RemovesOnlyILM(t *testing.T) {
	// Nested structure as returned by Elasticsearch API
	settings := map[string]interface{}{
		"index": map[string]interface{}{
			"lifecycle":        map[string]interface{}{"name": "my-policy"},
			"number_of_shards": "1",
			"refresh_interval": "30s",
		},
	}
	result := removeILMSetting(settings)

	indexSettings, ok := result["index"].(map[string]interface{})
	assert.True(t, ok)
	_, hasLifecycle := indexSettings["lifecycle"]
	assert.False(t, hasLifecycle)
	assert.Equal(t, "1", indexSettings["number_of_shards"])
	assert.Equal(t, "30s", indexSettings["refresh_interval"])
	assert.Len(t, indexSettings, 2)
	assert.Len(t, result, 1)
}

func TestRemoveILMSetting_EmptyAfterRemoval(t *testing.T) {
	// Nested structure: only ILM present, so template is empty after removal
	settings := map[string]interface{}{
		"index": map[string]interface{}{
			"lifecycle": map[string]interface{}{"name": "my-policy"},
		},
	}
	result := removeILMSetting(settings)

	assert.Nil(t, result)
}

func TestRemoveILMSetting_NilSettings(t *testing.T) {
	result := removeILMSetting(nil)
	assert.Nil(t, result)
}

func TestIsComponentTemplateEmpty(t *testing.T) {
	tests := []struct {
		name     string
		template *models.Template
		expected bool
	}{
		{
			name:     "nil template",
			template: nil,
			expected: true,
		},
		{
			name:     "empty template",
			template: &models.Template{},
			expected: true,
		},
		{
			name: "empty settings map",
			template: &models.Template{
				Settings: map[string]interface{}{},
			},
			expected: true,
		},
		{
			name: "has settings",
			template: &models.Template{
				Settings: map[string]interface{}{
					"index.number_of_replicas": 2,
				},
			},
			expected: false,
		},
		{
			name: "has mappings",
			template: &models.Template{
				Mappings: map[string]interface{}{
					"properties": map[string]interface{}{},
				},
			},
			expected: false,
		},
		{
			name: "has aliases",
			template: &models.Template{
				Aliases: map[string]models.IndexAlias{
					"my-alias": {},
				},
			},
			expected: false,
		},
		{
			name: "has settings and mappings",
			template: &models.Template{
				Settings: map[string]interface{}{
					"index.number_of_replicas": 2,
				},
				Mappings: map[string]interface{}{
					"properties": map[string]interface{}{},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isComponentTemplateEmpty(tt.template)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractILMSetting(t *testing.T) {
	tests := []struct {
		name     string
		template *models.Template
		expected string
	}{
		{
			name:     "nil template",
			template: nil,
			expected: "",
		},
		{
			name:     "empty template",
			template: &models.Template{},
			expected: "",
		},
		{
			name: "nil settings",
			template: &models.Template{
				Settings: nil,
			},
			expected: "",
		},
		{
			name: "no ILM setting",
			template: &models.Template{
				Settings: map[string]interface{}{
					"index": map[string]interface{}{
						"number_of_replicas": 2,
					},
				},
			},
			expected: "",
		},
		{
			name: "has ILM setting",
			template: &models.Template{
				Settings: map[string]interface{}{
					"index": map[string]interface{}{
						"lifecycle": map[string]interface{}{
							"name": "my-policy",
						},
						"number_of_replicas": 2,
					},
				},
			},
			expected: "my-policy",
		},
		{
			name: "ILM setting is not a string",
			template: &models.Template{
				Settings: map[string]interface{}{
					"index": map[string]interface{}{
						"lifecycle": map[string]interface{}{
							"name": 123,
						},
					},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractILMSetting(tt.template)
			assert.Equal(t, tt.expected, result)
		})
	}
}
