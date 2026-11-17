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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshState_stringSettingHoldingNonStringIsRecordedAsJSONText(t *testing.T) {
	t.Parallel()
	model := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{
			"custom:flag":  "on",
			"custom:array": "old",
		}),
	}
	userProvided := map[string]json.RawMessage{
		"custom:flag":  json.RawMessage(`true`),
		"custom:array": json.RawMessage(`["a","b"]`),
	}

	refreshed, diags := refreshState(t.Context(), model, userProvided)

	require.False(t, diags.HasError())
	assert.Equal(t, stringSettingsValue(map[string]string{
		"custom:flag":  "true",
		"custom:array": `["a","b"]`,
	}), refreshed.StringSettings)
}

func TestRefreshState_stringSettingHoldingANumberIsRecordedAsItsJSONText(t *testing.T) {
	t.Parallel()
	model := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"custom:teamNote": "hello"}),
	}
	userProvided := map[string]json.RawMessage{"custom:teamNote": json.RawMessage(`5`)}

	refreshed, diags := refreshState(t.Context(), model, userProvided)

	require.False(t, diags.HasError())
	assert.Equal(t, stringSettingsValue(map[string]string{"custom:teamNote": "5"}), refreshed.StringSettings)
}

func TestRefreshState_stringSettingIsRecordedAsKibanaString(t *testing.T) {
	t.Parallel()
	model := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "UTC"}),
	}
	userProvided := map[string]json.RawMessage{"dateFormat:tz": json.RawMessage(`"Europe/Rome"`)}

	refreshed, diags := refreshState(t.Context(), model, userProvided)

	require.False(t, diags.HasError())
	assert.Equal(t, stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"}), refreshed.StringSettings)
}

func TestRefreshState_jsonSettingIsRecordedAsJSONTextAndSemanticallyEqualToConfiguration(t *testing.T) {
	t.Parallel()
	configured := `{ "b": [ 1, 2 ],   "a": true }`
	model := advancedSettingsModel{
		JSONSettings: jsonSettingsValue(map[string]string{"custom:object": configured}),
	}
	userProvided := map[string]json.RawMessage{"custom:object": json.RawMessage(`{"a":true,"b":[1,2]}`)}

	refreshed, diags := refreshState(t.Context(), model, userProvided)

	require.False(t, diags.HasError())
	recorded, ok := refreshed.JSONSettings.Elements()["custom:object"].(jsontypes.Normalized)
	require.True(t, ok)
	assert.JSONEq(t, `{"a":true,"b":[1,2]}`, recorded.ValueString())
	equal, semanticDiags := recorded.StringSemanticEquals(t.Context(), jsontypes.NewNormalizedValue(configured))
	require.False(t, semanticDiags.HasError())
	assert.True(t, equal)
}

func TestRefreshState_jsonNumberFormattedDifferentlyKeepsTheRecordedText(t *testing.T) {
	t.Parallel()
	model := advancedSettingsModel{
		JSONSettings: jsonSettingsValue(map[string]string{"custom:ratio": "1.50"}),
	}
	userProvided := map[string]json.RawMessage{"custom:ratio": json.RawMessage(`1.5`)}

	refreshed, diags := refreshState(t.Context(), model, userProvided)

	require.False(t, diags.HasError())
	assert.Equal(t, jsonSettingsValue(map[string]string{"custom:ratio": "1.50"}), refreshed.JSONSettings)
}

func TestRefreshState_jsonNumbersThatReallyDifferAreRecordedAsKibanasText(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configured string
		kibana     string
	}{
		"a different value":                         {configured: "1.50", kibana: "2"},
		"an integer beyond 2^53 that Kibana rounds": {configured: "9007199254740993", kibana: "9007199254740992"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			model := advancedSettingsModel{JSONSettings: jsonSettingsValue(map[string]string{"custom:number": tt.configured})}
			userProvided := map[string]json.RawMessage{"custom:number": json.RawMessage(tt.kibana)}

			refreshed, diags := refreshState(t.Context(), model, userProvided)

			require.False(t, diags.HasError())
			assert.Equal(t, jsonSettingsValue(map[string]string{"custom:number": tt.kibana}), refreshed.JSONSettings)
		})
	}
}

func TestRefreshState_jsonNumbersNestedInObjectsAndArraysAreComparedByValue(t *testing.T) {
	t.Parallel()
	configured := `{ "weights": [1.0, 2.50], "limit": 1e2 }`
	model := advancedSettingsModel{
		JSONSettings: jsonSettingsValue(map[string]string{"custom:object": configured}),
	}
	userProvided := map[string]json.RawMessage{"custom:object": json.RawMessage(`{"limit":100,"weights":[1,2.5]}`)}

	refreshed, diags := refreshState(t.Context(), model, userProvided)

	require.False(t, diags.HasError())
	assert.Equal(t, jsonSettingsValue(map[string]string{"custom:object": configured}), refreshed.JSONSettings)
}

func TestRefreshState_omittedMapsStayNull(t *testing.T) {
	t.Parallel()
	model := advancedSettingsModel{
		StringSettings: types.MapNull(types.StringType),
		JSONSettings:   types.MapNull(jsontypes.NormalizedType{}),
	}
	userProvided := map[string]json.RawMessage{"defaultRoute": json.RawMessage(`"/app/home"`)}

	refreshed, diags := refreshState(t.Context(), model, userProvided)

	require.False(t, diags.HasError())
	assert.True(t, refreshed.StringSettings.IsNull())
	assert.True(t, refreshed.JSONSettings.IsNull())
}

func TestRefreshState_managedKeysAbsentFromKibanaAreDropped(t *testing.T) {
	t.Parallel()
	model := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"custom:kept": "x", "custom:reset": "y"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"custom:reset": "1"}),
	}
	userProvided := map[string]json.RawMessage{"custom:kept": json.RawMessage(`"x"`)}

	refreshed, diags := refreshState(t.Context(), model, userProvided)

	require.False(t, diags.HasError())
	assert.Equal(t, stringSettingsValue(map[string]string{"custom:kept": "x"}), refreshed.StringSettings)
	assert.Equal(t, jsonSettingsValue(map[string]string{}), refreshed.JSONSettings)
}

func TestRefreshState_unmanagedKeysAreNeverRecorded(t *testing.T) {
	t.Parallel()
	model := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "UTC"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"csv:quoteValues": "true"}),
	}
	userProvided := map[string]json.RawMessage{
		"dateFormat:tz":   json.RawMessage(`"UTC"`),
		"csv:quoteValues": json.RawMessage(`true`),
		"defaultRoute":    json.RawMessage(`"/app/home"`),
	}

	refreshed, diags := refreshState(t.Context(), model, userProvided)

	require.False(t, diags.HasError())
	assert.Equal(t, stringSettingsValue(map[string]string{"dateFormat:tz": "UTC"}), refreshed.StringSettings)
	assert.Equal(t, jsonSettingsValue(map[string]string{"csv:quoteValues": "true"}), refreshed.JSONSettings)
}
