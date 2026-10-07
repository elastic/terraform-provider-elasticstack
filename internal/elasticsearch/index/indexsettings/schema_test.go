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

package indexsettings

import (
	"context"
	"testing"

	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSchemaFactory_mergesDynamicSettingAttributes(t *testing.T) {
	attributes := getSchemaFactory(context.Background()).Attributes

	for _, key := range indexparent.DynamicSettingsKeys {
		attributeName := typeutils.ConvertSettingsKeyToTFFieldKey(key)
		assert.Contains(t, attributes, attributeName, "missing dynamic attribute for settings key %q", key)
	}

	assert.Contains(t, attributes, "id")
	assert.Contains(t, attributes, "index")
	assert.Contains(t, attributes, "settings_json")
	assert.NotContains(t, attributes, "elasticsearch_connection")
	assert.NotContains(t, attributes, "timeouts")
}

func TestGetSchemaFactory_id(t *testing.T) {
	id, ok := getSchemaFactory(context.Background()).Attributes["id"].(schema.StringAttribute)
	require.True(t, ok, "id attribute must be a string attribute")
	assert.True(t, id.Computed)
	assert.NotEmpty(t, planModifierDescriptions(id.PlanModifiers), "id must carry the UseStateForUnknown plan modifier")
}

func TestGetSchemaFactory_index(t *testing.T) {
	index, ok := getSchemaFactory(context.Background()).Attributes["index"].(schema.StringAttribute)
	require.True(t, ok, "index attribute must be a string attribute")
	assert.True(t, index.Required)
	assert.NotEmpty(t, planModifierDescriptions(index.PlanModifiers), "index must carry the RequiresReplace plan modifier")
}

func TestGetSchemaFactory_settingsJSON(t *testing.T) {
	attributes := getSchemaFactory(context.Background()).Attributes
	settingsJSON, ok := attributes["settings_json"].(schema.StringAttribute)
	require.True(t, ok, "settings_json attribute must be a string attribute")
	assert.True(t, settingsJSON.Optional)
	assert.NotEmpty(t, settingsJSON.Validators, "settings_json must carry the plan-time JSON-object validator")
}

func planModifierDescriptions(modifiers []planmodifier.String) []string {
	descriptions := make([]string, 0, len(modifiers))
	for _, m := range modifiers {
		descriptions = append(descriptions, m.Description(context.Background()))
	}
	return descriptions
}
