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

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func updateRequest(spaceID string, plan, prior advancedSettingsModel) entitycore.KibanaWriteRequest[advancedSettingsModel] {
	prior.SpaceID = types.StringValue(spaceID)
	request := createRequest(spaceID, plan)
	request.Prior = &prior
	return request
}

func TestUpdateAdvancedSettings_sendsChangesAdditionsAndResetsInOneRequest(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	prior := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome", "defaultRoute": "/app/home"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"truncate:maxHeight": "0"}),
	}
	plan := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "America/New_York", "custom:teamNote": "hello"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"truncate:maxHeight": "0"}),
	}

	result, diags := updateAdvancedSettings(t.Context(), client, updateRequest("default", plan, prior))

	require.False(t, diags.HasError())
	assert.Equal(t, []string{"POST /api/kibana/settings"}, fake.recordedRequests())
	assert.Equal(t, []map[string]json.RawMessage{{
		"dateFormat:tz":   json.RawMessage(`"America/New_York"`),
		"custom:teamNote": json.RawMessage(`"hello"`),
		"defaultRoute":    json.RawMessage(`null`),
	}}, fake.appliedChanges())
	assert.Equal(t, plan.StringSettings, result.Model.StringSettings)
}

func TestUpdateAdvancedSettings_failsWithoutWritingWhenTheSpaceWasDeleted(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	prior := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"custom:teamNote": "hello", "defaultRoute": "/app/home"})}
	plan := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"custom:teamNote": "changed"})}

	result, diags := updateAdvancedSettings(t.Context(), client, updateRequest("ops", plan, prior))

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), `"ops" must exist`)
	assert.Equal(t, []string{"GET /api/spaces/space/ops"}, fake.recordedRequests())
	assert.Empty(t, fake.appliedChanges())
	assert.True(t, result.Model.ID.IsNull())
}

func TestUpdateAdvancedSettings_sendsNothingWhenNothingChanged(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	state := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"csv:quoteValues": "false"}),
	}

	result, diags := updateAdvancedSettings(t.Context(), client, updateRequest("default", state, state))

	require.False(t, diags.HasError())
	assert.Empty(t, fake.recordedRequests())
	assert.Equal(t, "default", result.Model.ID.ValueString())
}

func TestUpdateAdvancedSettings_doesNotAdviseJSONSettingsWhenAnUnchangedStringSettingWasNotSent(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	const kibanaMessage = "[validation [histogram:barTarget]]: expected value of type [number] but got [string]"
	fake.rejection = kibanaMessage
	prior := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"histogram:barTarget": "6"}),
	}
	plan := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"histogram:barTarget": `"tall"`}),
	}

	_, diags := updateAdvancedSettings(t.Context(), client, updateRequest("default", plan, prior))

	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, kibanaMessage, diags.Errors()[0].Detail())
}

func TestUpdateAdvancedSettings_doesNotAdviseJSONSettingsWhenAFormerStringSettingIsOnlyReset(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	const kibanaMessage = `Unable to update "observability:streamsEnableQueryStreams" because it is overridden`
	fake.rejection = kibanaMessage
	prior := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"observability:streamsEnableQueryStreams": "true"})}
	plan := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{})}

	_, diags := updateAdvancedSettings(t.Context(), client, updateRequest("default", plan, prior))

	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, kibanaMessage, diags.Errors()[0].Detail())
}

func TestUpdateAdvancedSettings_aMissingSettingsRouteIsAnErrorNamingTheRoute(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.missingSettingsRoute = true
	prior := advancedSettingsModel{JSONSettings: jsonSettingsValue(map[string]string{"csv:quoteValues": "false"})}
	plan := advancedSettingsModel{JSONSettings: jsonSettingsValue(map[string]string{"csv:quoteValues": "true"})}

	result, diags := updateAdvancedSettings(t.Context(), client, updateRequest("default", plan, prior))

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), "POST /api/kibana/settings")
	assert.True(t, result.Model.ID.IsNull())
}

func TestUpdateAdvancedSettings_doesNotAdviseJSONSettingsOnAMissingSettingsRoute(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.missingSettingsRoute = true
	prior := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"})}
	plan := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "America/New_York"})}

	_, diags := updateAdvancedSettings(t.Context(), client, updateRequest("default", plan, prior))

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), "POST /api/kibana/settings")
	assert.NotContains(t, diags.Errors()[0].Detail(), "json_settings")
}

func TestUpdateAdvancedSettings_advisesJSONSettingsWhenAChangedStringSettingIsRejected(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	const kibanaMessage = "[validation [csv:quoteValues]]: expected value of type [boolean] but got [string]"
	fake.rejection = kibanaMessage
	prior := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"csv:quoteValues": "true"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"truncate:maxHeight": "0"}),
	}
	plan := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"csv:quoteValues": "maybe"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"truncate:maxHeight": "0"}),
	}

	_, diags := updateAdvancedSettings(t.Context(), client, updateRequest("default", plan, prior))

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), kibanaMessage)
	assert.Contains(t, diags.Errors()[0].Detail(), jsonSettingsHint)
}

func TestUpdateAdvancedSettings_sendsNothingWhenAJSONValueIsOnlyReformatted(t *testing.T) {
	t.Parallel()
	tests := map[string]struct{ recorded, reformatted string }{
		"number formatting":      {`1.5`, `1.50`},
		"whitespace in an array": {`["a","b"]`, `[ "a",   "b" ]`},
		"object key order":       {`{"a":1,"b":{"c":2,"d":3}}`, `{"b":{"d":3,"c":2},"a":1}`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fake, client := newFakeKibana(t)
			prior := advancedSettingsModel{JSONSettings: jsonSettingsValue(map[string]string{"custom:value": tt.recorded})}
			plan := advancedSettingsModel{JSONSettings: jsonSettingsValue(map[string]string{"custom:value": tt.reformatted})}

			result, diags := updateAdvancedSettings(t.Context(), client, updateRequest("default", plan, prior))

			require.False(t, diags.HasError())
			assert.Empty(t, fake.recordedRequests())
			assert.Equal(t, plan.JSONSettings, result.Model.JSONSettings)
		})
	}
}
