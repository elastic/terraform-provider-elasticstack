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

package advancedsettings

import (
	"maps"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchema_idIsComputedAndKeptAcrossUpdates(t *testing.T) {
	t.Parallel()

	attr, ok := getSchema(t.Context()).Attributes["id"].(schema.StringAttribute)

	require.True(t, ok)
	assert.True(t, attr.Computed)
	assert.False(t, attr.Optional)
	assert.False(t, attr.Required)
	require.Len(t, attr.PlanModifiers, 1)
	assert.Equal(t, stringplanmodifier.UseStateForUnknown().Description(t.Context()), attr.PlanModifiers[0].Description(t.Context()))
}

func TestSchema_spaceIDDefaultsToDefaultAndForcesReplacement(t *testing.T) {
	t.Parallel()

	attr, ok := getSchema(t.Context()).Attributes["space_id"].(schema.StringAttribute)

	require.True(t, ok)
	assert.True(t, attr.Optional)
	assert.True(t, attr.Computed)
	require.NotNil(t, attr.Default)
	defaultResp := defaults.StringResponse{}
	attr.Default.DefaultString(t.Context(), defaults.StringRequest{}, &defaultResp)
	assert.Equal(t, "default", defaultResp.PlanValue.ValueString())
	require.Len(t, attr.PlanModifiers, 1)
	assert.Equal(t, stringplanmodifier.RequiresReplace().Description(t.Context()), attr.PlanModifiers[0].Description(t.Context()))
}

func TestSchema_stringSettingsIsAnOptionalMapOfStrings(t *testing.T) {
	t.Parallel()

	attr, ok := getSchema(t.Context()).Attributes["string_settings"].(schema.MapAttribute)

	require.True(t, ok)
	assert.True(t, attr.Optional)
	assert.False(t, attr.Computed)
	assert.Equal(t, types.StringType, attr.ElementType)
}

func TestSchema_jsonSettingsIsAnOptionalMapOfNormalizedJSON(t *testing.T) {
	t.Parallel()

	attr, ok := getSchema(t.Context()).Attributes["json_settings"].(schema.MapAttribute)

	require.True(t, ok)
	assert.True(t, attr.Optional)
	assert.False(t, attr.Computed)
	assert.Equal(t, jsontypes.NormalizedType{}, attr.ElementType)
}

func TestSchema_envelopeSchemaHasOnlyTheDeclaredAttributesAndNoVersion(t *testing.T) {
	t.Parallel()
	resp := resource.SchemaResponse{}

	NewResource().Schema(t.Context(), resource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError())
	assert.Equal(t,
		[]string{"id", "json_settings", "space_id", "string_settings", "timeouts"},
		slices.Sorted(maps.Keys(resp.Schema.Attributes)),
	)
	assert.Equal(t, []string{"kibana_connection"}, slices.Sorted(maps.Keys(resp.Schema.Blocks)))
	assert.Zero(t, resp.Schema.Version)
}

func TestSchema_descriptionsExplainHowSettingsAreManaged(t *testing.T) {
	t.Parallel()
	resourceSchema := getSchema(t.Context())
	stringSettings := resourceSchema.Attributes["string_settings"].GetMarkdownDescription()
	jsonSettings := resourceSchema.Attributes["json_settings"].GetMarkdownDescription()

	tests := map[string]struct {
		description string
		wantPhrases []string
	}{
		"only declared keys are managed and others are untouched": {
			resourceSchema.MarkdownDescription,
			[]string{"Only the settings you declare", "left untouched"},
		},
		"settings are scoped to a space that must exist": {
			resourceSchema.MarkdownDescription,
			[]string{"`space_id`", "must already exist"},
		},
		"a space deleted outside Terraform removes the resource from state": {
			resourceSchema.MarkdownDescription,
			[]string{"created or updated", "deleted outside Terraform", "removes the resource from state"},
		},
		"values are stored in state and must not hold secrets": {
			resourceSchema.MarkdownDescription,
			[]string{"stored in the Terraform state", "must not contain secrets"},
		},
		"two resources in one space overwrite each other": {
			resourceSchema.MarkdownDescription,
			[]string{"overwrite each other", "`defaultIndex`", "`elasticstack_kibana_default_data_view`"},
		},
		"string versus JSON typing with examples": {
			resourceSchema.MarkdownDescription,
			[]string{"`string_settings`", "`json_settings`", "jsonencode(", "timepicker:timeDefaults"},
		},
		"removing a key resets the setting to its Kibana default": {
			resourceSchema.MarkdownDescription,
			[]string{"Removing a key", "resets that setting to its Kibana default", "moving a key"},
		},
		"drift detection cannot tell a string true from a boolean true": {
			resourceSchema.MarkdownDescription,
			[]string{"Drift detection", "string `\"true\"`", "boolean `true`", "indistinguishable"},
		},
		"a key resets only when it is declared in neither map": {
			resourceSchema.MarkdownDescription,
			[]string{"Removing a key from the configuration", "declared in neither map"},
		},
		"writing needs the manage_advanced_settings privilege and the credentials must be able to read the space": {
			resourceSchema.MarkdownDescription,
			[]string{
				"Permissions", "`manage_advanced_settings`", "Reading settings does not require the privilege",
				"spaces API", "must be able to read that space", "instead of being treated as a missing space",
			},
		},
		"a rejected operation fails with Kibana's message and records nothing": {
			resourceSchema.MarkdownDescription,
			[]string{"includes Kibana's message", "failed create leaves the resource out of the Terraform state", "failed update keeps the values recorded before the update"},
		},
		"a failed request that sent string settings advises json_settings": {
			resourceSchema.MarkdownDescription,
			[]string{
				"advises declaring settings with non-string values in `json_settings`", "only new or changed",
				"only `json_settings` entries", "unrelated to types", "not added to the error for a missing settings route",
				"accepted the string `\"false\"`",
			},
		},
		"a missing settings route is an error naming the route": {
			resourceSchema.MarkdownDescription,
			[]string{"`GET /api/kibana/settings`", "`POST /api/kibana/settings`", "naming the route", "not dropped from the state by mistake"},
		},
		"the settings routes are internal with no public contract": {
			resourceSchema.MarkdownDescription,
			[]string{"internal Kibana routes", "no public API contract", "`x-elastic-internal-origin: Kibana`", "change or remove them"},
		},
		"settings can outlive a deleted space": {
			resourceSchema.MarkdownDescription,
			[]string{"Settings can outlive a deleted space", "recreating a space with the same ID", "earlier settings"},
		},
		"integers beyond the JavaScript safe range may show as drift": {
			resourceSchema.MarkdownDescription,
			[]string{"JavaScript safe range", "2^53", "rounded", "may show as drift"},
		},
		"verified on current stateful Kibana only": {
			resourceSchema.MarkdownDescription,
			[]string{"verified against current stateful Kibana (9.5.6)", "Older Kibana versions", "Serverless", "not verified"},
		},
		"JSON formatting differences returned by Kibana are not drift but reformatting the configuration plans an update": {
			resourceSchema.MarkdownDescription,
			[]string{
				"does not show as drift",
				"in your own configuration can show as an in-place update that changes nothing in Kibana",
			},
		},
		"json settings explain that reformatting the configuration plans an update": {
			jsonSettings,
			[]string{
				"does not show as drift",
				"in your own configuration can show as an in-place update that changes nothing in Kibana",
			},
		},
		"string settings explain plain strings and JSON stored as a string": {
			stringSettings,
			[]string{"plain string", "JSON text"},
		},
		"json settings explain the accepted JSON types": {
			jsonSettings,
			[]string{"boolean", "number", "array", "object"},
		},
		"string settings explain the key rules and that null is not accepted": {
			stringSettings,
			[]string{"between 1 and 1024 characters", "only one of `string_settings` and `json_settings`", "null", "remove the key"},
		},
		"json settings explain the key rules and that null is not accepted": {
			jsonSettings,
			[]string{"between 1 and 1024 characters", "only one of `string_settings` and `json_settings`", "`null`", "remove the key"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, phrase := range tt.wantPhrases {
				assert.Contains(t, tt.description, phrase)
			}
		})
	}
}
