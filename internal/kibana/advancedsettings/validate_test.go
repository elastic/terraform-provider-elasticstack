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
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validateSettingsAttribute(t *testing.T, attributeName string, value types.Map) diag.Diagnostics {
	t.Helper()
	attribute, ok := getSchema(t.Context()).Attributes[attributeName].(schema.MapAttribute)
	require.True(t, ok)

	var diags diag.Diagnostics
	for _, attributeValidator := range attribute.Validators {
		resp := validator.MapResponse{}
		attributeValidator.ValidateMap(t.Context(), validator.MapRequest{
			Path:           path.Root(attributeName),
			PathExpression: path.MatchRoot(attributeName),
			ConfigValue:    value,
		}, &resp)
		diags.Append(resp.Diagnostics...)
	}
	return diags
}

func TestStringSettingsValidators_rejectAnEmptyKey(t *testing.T) {
	t.Parallel()

	diags := validateSettingsAttribute(t, "string_settings", stringSettingsValue(map[string]string{"": "value"}))

	require.True(t, diags.HasError())
}

func TestJSONSettingsValidators_rejectAKeyLongerThan1024Characters(t *testing.T) {
	t.Parallel()

	diags := validateSettingsAttribute(t, "json_settings", jsonSettingsValue(map[string]string{strings.Repeat("k", 1025): "true"}))

	require.True(t, diags.HasError())
}

func TestSettingsValidators_acceptKeysFromOneToOneThousandTwentyFourCharacters(t *testing.T) {
	t.Parallel()

	keys := map[string]string{
		"shortest":                        "k",
		"longest":                         strings.Repeat("k", 1024),
		"longest in multibyte characters": strings.Repeat("é", 1024),
	}
	for name, key := range keys {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stringDiags := validateSettingsAttribute(t, "string_settings", stringSettingsValue(map[string]string{key: "value"}))
			jsonDiags := validateSettingsAttribute(t, "json_settings", jsonSettingsValue(map[string]string{key: "true"}))

			assert.False(t, stringDiags.HasError())
			assert.False(t, jsonDiags.HasError())
		})
	}
}

func TestStringSettingsValidators_rejectANullEntryAdvisingToRemoveTheKey(t *testing.T) {
	t.Parallel()
	settings := types.MapValueMust(types.StringType, map[string]attr.Value{
		"defaultRoute":    types.StringNull(),
		"custom:teamNote": types.StringValue("hello"),
	})

	diags := validateSettingsAttribute(t, "string_settings", settings)

	require.Len(t, diags.Errors(), 1)
	assertErrorAtMapKey(t, diags.Errors()[0], "string_settings", "defaultRoute")
	assert.Contains(t, diags.Errors()[0].Detail(), "defaultRoute")
	assert.Contains(t, diags.Errors()[0].Detail(), "remove the key")
}

func assertErrorAtMapKey(t *testing.T, diagnostic diag.Diagnostic, attributeName, key string) {
	t.Helper()
	attributeError, ok := diagnostic.(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, path.Root(attributeName).AtMapKey(key), attributeError.Path())
}

func TestJSONSettingsValidators_rejectANullEntryAdvisingToRemoveTheKey(t *testing.T) {
	t.Parallel()
	settings := types.MapValueMust(jsontypes.NormalizedType{}, map[string]attr.Value{
		"defaultRoute":    jsontypes.NewNormalizedNull(),
		"custom:teamNote": jsontypes.NewNormalizedValue(`"hello"`),
	})

	diags := validateSettingsAttribute(t, "json_settings", settings)

	require.Len(t, diags.Errors(), 1)
	assertErrorAtMapKey(t, diags.Errors()[0], "json_settings", "defaultRoute")
	assert.Contains(t, diags.Errors()[0].Detail(), "defaultRoute")
	assert.Contains(t, diags.Errors()[0].Detail(), "remove the key")
}

func validateConfig(t *testing.T, stringSettings, jsonSettings types.Map) diag.Diagnostics {
	t.Helper()
	ctx := t.Context()
	configSchema := getSchema(ctx)
	stringSettingsRaw, err := stringSettings.ToTerraformValue(ctx)
	require.NoError(t, err)
	jsonSettingsRaw, err := jsonSettings.ToTerraformValue(ctx)
	require.NoError(t, err)
	configType, ok := configSchema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	req := resource.ValidateConfigRequest{Config: tfsdk.Config{
		Schema: configSchema,
		Raw: tftypes.NewValue(configType, map[string]tftypes.Value{
			"id":              tftypes.NewValue(tftypes.String, nil),
			"space_id":        tftypes.NewValue(tftypes.String, nil),
			"string_settings": stringSettingsRaw,
			"json_settings":   jsonSettingsRaw,
		}),
	}}
	resp := resource.ValidateConfigResponse{}
	newResource().ValidateConfig(ctx, req, &resp)
	return resp.Diagnostics
}

func TestValidateConfig_rejectsAKeyDeclaredInBothMaps(t *testing.T) {
	t.Parallel()

	diags := validateConfig(t,
		stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome", "custom:teamNote": "hello"}),
		jsonSettingsValue(map[string]string{"dateFormat:tz": `"Europe/Rome"`, "csv:quoteValues": "false"}),
	)

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), "dateFormat:tz")
}

func TestValidateConfig_rejectsInvalidJSONIdentifyingTheKey(t *testing.T) {
	t.Parallel()

	diags := validateConfig(t,
		types.MapNull(types.StringType),
		jsonSettingsValue(map[string]string{"truncate:maxHeight": "not json", "csv:quoteValues": "false"}),
	)

	require.Len(t, diags.Errors(), 1)
	assertErrorAtMapKey(t, diags.Errors()[0], "json_settings", "truncate:maxHeight")
	assert.Contains(t, diags.Errors()[0].Detail(), "truncate:maxHeight")
}

func TestValidateConfig_leavesUnknownAndNullJSONEntriesToOtherChecks(t *testing.T) {
	t.Parallel()

	tests := map[string]attr.Value{
		"unknown at plan time": jsontypes.NewNormalizedUnknown(),
		"null":                 jsontypes.NewNormalizedNull(),
	}
	for name, element := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			settings := types.MapValueMust(jsontypes.NormalizedType{}, map[string]attr.Value{"truncate:maxHeight": element})

			diags := validateConfig(t, types.MapNull(types.StringType), settings)

			assert.False(t, diags.HasError())
		})
	}
}

func TestValidateConfig_rejectsTheJSONValueNullAdvisingToRemoveTheKey(t *testing.T) {
	t.Parallel()

	diags := validateConfig(t,
		types.MapNull(types.StringType),
		jsonSettingsValue(map[string]string{"defaultRoute": "null", "csv:quoteValues": "false"}),
	)

	require.Len(t, diags.Errors(), 1)
	assertErrorAtMapKey(t, diags.Errors()[0], "json_settings", "defaultRoute")
	assert.Contains(t, diags.Errors()[0].Detail(), "defaultRoute")
	assert.Contains(t, diags.Errors()[0].Detail(), "remove the key")
}

func TestValidateConfig_acceptsJSONValuesOfEveryType(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"string":          `"/app/dashboards"`,
		"empty string":    `""`,
		"false":           "false",
		"zero":            "0",
		"array":           `["@timestamp","message"]`,
		"empty array":     "[]",
		"object":          `{"a":1}`,
		"padded document": "  true\n",
	}
	for name, text := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := validateConfig(t, types.MapNull(types.StringType), jsonSettingsValue(map[string]string{"defaultRoute": text}))

			assert.False(t, diags.HasError())
		})
	}
}
