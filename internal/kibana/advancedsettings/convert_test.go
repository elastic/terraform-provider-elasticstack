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
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stringSettingsValue(settings map[string]string) types.Map {
	elements := make(map[string]attr.Value, len(settings))
	for key, value := range settings {
		elements[key] = types.StringValue(value)
	}
	return types.MapValueMust(types.StringType, elements)
}

func jsonSettingsValue(settings map[string]string) types.Map {
	elements := make(map[string]attr.Value, len(settings))
	for key, value := range settings {
		elements[key] = jsontypes.NewNormalizedValue(value)
	}
	return types.MapValueMust(jsontypes.NormalizedType{}, elements)
}

func TestBuildChanges_stringSettingsAreSentAsStrings(t *testing.T) {
	t.Parallel()
	plan := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{
			"dateFormat:tz":   "Europe/Rome",
			"custom:jsonText": `["a","b"]`,
		}),
	}

	changes, diags := buildChanges(plan, nil)

	require.False(t, diags.HasError())
	assert.Equal(t, map[string]any{
		"dateFormat:tz":   "Europe/Rome",
		"custom:jsonText": `["a","b"]`,
	}, changes)
}

func TestBuildChanges_jsonSettingsKeepTheirJSONType(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		jsonText string
		want     any
	}{
		"boolean":                         {jsonText: "false", want: false},
		"array":                           {jsonText: `[ "@timestamp",  "message" ]`, want: []any{"@timestamp", "message"}},
		"object":                          {jsonText: `{"nested":{"enabled":true}}`, want: map[string]any{"nested": map[string]any{"enabled": true}}},
		"string":                          {jsonText: `"Europe/Rome"`, want: "Europe/Rome"},
		"object surrounded by whitespace": {jsonText: "\n  {\"enabled\": true}\t", want: map[string]any{"enabled": true}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			plan := advancedSettingsModel{JSONSettings: jsonSettingsValue(map[string]string{"custom:setting": tt.jsonText})}

			changes, diags := buildChanges(plan, nil)

			require.False(t, diags.HasError())
			assert.Equal(t, map[string]any{"custom:setting": tt.want}, changes)
		})
	}
}

func TestBuildChanges_jsonNumbersKeepTheirExactText(t *testing.T) {
	t.Parallel()
	plan := advancedSettingsModel{
		JSONSettings: jsonSettingsValue(map[string]string{
			"truncate:maxHeight":  "0",
			"custom:exactInteger": "9007199254740993",
		}),
	}

	changes, diags := buildChanges(plan, nil)

	require.False(t, diags.HasError())
	assert.Equal(t, map[string]any{
		"truncate:maxHeight":  json.Number("0"),
		"custom:exactInteger": json.Number("9007199254740993"),
	}, changes)
}

func TestBuildChanges_unparsableJSONIsReportedForItsKey(t *testing.T) {
	t.Parallel()

	tests := map[string]attr.Value{
		"malformed text":                 jsontypes.NewNormalizedValue("not json"),
		"a second value after a boolean": jsontypes.NewNormalizedValue("true false"),
		"garbage after an object":        jsontypes.NewNormalizedValue(`{"a":1}garbage`),
		"garbage after an array":         jsontypes.NewNormalizedValue("[1,2] x"),
		"a null entry":                   jsontypes.NewNormalizedNull(),
	}
	for name, element := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			plan := advancedSettingsModel{
				JSONSettings: types.MapValueMust(jsontypes.NormalizedType{}, map[string]attr.Value{
					"custom:valid":  jsontypes.NewNormalizedValue("true"),
					"custom:broken": element,
				}),
			}

			changes, diags := buildChanges(plan, nil)

			require.Len(t, diags.Errors(), 1)
			attributeError, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			assert.Equal(t, path.Root("json_settings").AtMapKey("custom:broken"), attributeError.Path())
			assert.NotContains(t, changes, "custom:broken")
		})
	}
}

func TestBuildChanges_theJSONValueNullIsRejectedAdvisingToRemoveTheKey(t *testing.T) {
	t.Parallel()
	plan := advancedSettingsModel{
		JSONSettings: jsonSettingsValue(map[string]string{
			"custom:valid": "true",
			"defaultRoute": "null",
		}),
	}

	changes, diags := buildChanges(plan, nil)

	require.Len(t, diags.Errors(), 1)
	assertErrorAtMapKey(t, diags.Errors()[0], "json_settings", "defaultRoute")
	assert.Contains(t, diags.Errors()[0].Detail(), "remove the key")
	assert.NotContains(t, changes, "defaultRoute")
}

func TestBuildChanges_aNullJSONEntryIsRejectedAdvisingToRemoveTheKey(t *testing.T) {
	t.Parallel()
	plan := advancedSettingsModel{
		JSONSettings: types.MapValueMust(jsontypes.NormalizedType{}, map[string]attr.Value{
			"custom:valid": jsontypes.NewNormalizedValue("true"),
			"defaultRoute": jsontypes.NewNormalizedNull(),
		}),
	}

	changes, diags := buildChanges(plan, nil)

	require.Len(t, diags.Errors(), 1)
	assertErrorAtMapKey(t, diags.Errors()[0], "json_settings", "defaultRoute")
	assert.Contains(t, diags.Errors()[0].Detail(), "remove the key")
	assert.NotContains(t, changes, "defaultRoute")
}

func TestBuildChanges_aNullStringEntryIsRejectedAdvisingToRemoveTheKey(t *testing.T) {
	t.Parallel()
	plan := advancedSettingsModel{
		StringSettings: types.MapValueMust(types.StringType, map[string]attr.Value{
			"custom:valid": types.StringValue("hello"),
			"defaultRoute": types.StringNull(),
		}),
	}

	changes, diags := buildChanges(plan, nil)

	require.Len(t, diags.Errors(), 1)
	assertErrorAtMapKey(t, diags.Errors()[0], "string_settings", "defaultRoute")
	assert.Contains(t, diags.Errors()[0].Detail(), "remove the key")
	assert.NotContains(t, changes, "defaultRoute")
}

func TestBuildChanges_unchangedStringSettingsAreOmitted(t *testing.T) {
	t.Parallel()
	prior := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{
			"dateFormat:tz": "Europe/Rome",
			"defaultRoute":  "/app/home",
		}),
	}
	plan := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{
			"dateFormat:tz": "Europe/Rome",
			"defaultRoute":  "/app/dashboards",
		}),
	}

	changes, diags := buildChanges(plan, &prior)

	require.False(t, diags.HasError())
	assert.Equal(t, map[string]any{"defaultRoute": "/app/dashboards"}, changes)
}

func TestBuildChanges_jsonSettingsWithTheSameMeaningAreOmitted(t *testing.T) {
	t.Parallel()
	prior := advancedSettingsModel{
		JSONSettings: jsonSettingsValue(map[string]string{
			"custom:object":      `{"a":1,"b":[1,2]}`,
			"truncate:maxHeight": "0",
			"csv:quoteValues":    "false",
		}),
	}
	plan := advancedSettingsModel{
		JSONSettings: jsonSettingsValue(map[string]string{
			"custom:object":      "{\n  \"b\": [1, 2],\n  \"a\": 1\n}",
			"truncate:maxHeight": "20",
			"csv:quoteValues":    "false",
		}),
	}

	changes, diags := buildChanges(plan, &prior)

	require.False(t, diags.HasError())
	assert.Equal(t, map[string]any{"truncate:maxHeight": json.Number("20")}, changes)
}

func TestBuildChanges_jsonNumbersWithTheSameValueInADifferentFormatAreOmitted(t *testing.T) {
	t.Parallel()
	prior := advancedSettingsModel{
		JSONSettings: jsonSettingsValue(map[string]string{
			"custom:ratio":  "1.50",
			"custom:whole":  "1.0",
			"custom:object": `{"weights":[1.0,2.50],"limit":1e2}`,
			"custom:big":    "9007199254740993",
		}),
	}
	plan := advancedSettingsModel{
		JSONSettings: jsonSettingsValue(map[string]string{
			"custom:ratio":  "1.5",
			"custom:whole":  "1",
			"custom:object": `{"limit":100,"weights":[1,2.5]}`,
			"custom:big":    "9007199254740992",
		}),
	}

	changes, diags := buildChanges(plan, &prior)

	require.False(t, diags.HasError())
	assert.Equal(t, map[string]any{"custom:big": json.Number("9007199254740992")}, changes)
}

func TestBuildChanges_keysNoLongerDeclaredAreReset(t *testing.T) {
	t.Parallel()
	prior := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{
			"dateFormat:tz": "Europe/Rome",
			"defaultRoute":  "/app/home",
		}),
		JSONSettings: jsonSettingsValue(map[string]string{
			"csv:quoteValues":    "false",
			"truncate:maxHeight": "0",
		}),
	}
	plan := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"truncate:maxHeight": "0"}),
	}

	changes, diags := buildChanges(plan, &prior)

	require.False(t, diags.HasError())
	assert.Equal(t, map[string]any{"defaultRoute": nil, "csv:quoteValues": nil}, changes)
}

func TestBuildChanges_aKeyMovedBetweenMapsIsSentWithItsNewTypeInsteadOfReset(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		prior advancedSettingsModel
		plan  advancedSettingsModel
		want  any
	}{
		"from string_settings to json_settings": {
			prior: advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"custom:teamNote": "5"})},
			plan:  advancedSettingsModel{JSONSettings: jsonSettingsValue(map[string]string{"custom:teamNote": "5"})},
			want:  json.Number("5"),
		},
		"from json_settings to string_settings": {
			prior: advancedSettingsModel{JSONSettings: jsonSettingsValue(map[string]string{"custom:teamNote": "5"})},
			plan:  advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"custom:teamNote": "5"})},
			want:  "5",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			changes, diags := buildChanges(tt.plan, &tt.prior)

			require.False(t, diags.HasError())
			assert.Equal(t, map[string]any{"custom:teamNote": tt.want}, changes)
		})
	}
}

func TestParseJSONKeepingNumbers_acceptsExactlyOneJSONValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		text    string
		want    any
		wantErr bool
	}{
		"malformed text":                         {text: "not json", wantErr: true},
		"a second value after a boolean":         {text: "true false", wantErr: true},
		"garbage after an object":                {text: `{"a":1}garbage`, wantErr: true},
		"garbage after an array":                 {text: "[1,2] x", wantErr: true},
		"a stray closing bracket after an array": {text: "[1,2]]", wantErr: true},
		"surrounding whitespace around an array": {text: "  [1, 2]\n", want: []any{json.Number("1"), json.Number("2")}},
		"surrounding whitespace around a number": {text: "\t12.50 \n", want: json.Number("12.50")},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			parsed, err := parseJSONKeepingNumbers(tt.text)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, parsed)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, parsed)
		})
	}
}
