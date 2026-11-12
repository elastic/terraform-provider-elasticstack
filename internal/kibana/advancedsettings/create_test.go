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
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRequest(spaceID string, plan advancedSettingsModel) entitycore.KibanaWriteRequest[advancedSettingsModel] {
	plan.SpaceID = types.StringValue(spaceID)
	return entitycore.KibanaWriteRequest[advancedSettingsModel]{Plan: plan, WriteID: spaceID, SpaceID: spaceID}
}

func TestCreateAdvancedSettings_appliesEveryDeclaredKeyInOneRequest(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	plan := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"csv:quoteValues": "false", "truncate:maxHeight": "0"}),
	}

	result, diags := createAdvancedSettings(t.Context(), client, createRequest("default", plan))

	require.False(t, diags.HasError())
	assert.Equal(t, []string{"POST /api/kibana/settings"}, fake.recordedRequests())
	assert.Equal(t, []map[string]json.RawMessage{{
		"dateFormat:tz":      json.RawMessage(`"Europe/Rome"`),
		"csv:quoteValues":    json.RawMessage(`false`),
		"truncate:maxHeight": json.RawMessage(`0`),
	}}, fake.appliedChanges())
	assert.Equal(t, "default", result.Model.ID.ValueString())
	assert.Equal(t, "default", result.Model.SpaceID.ValueString())
}

func TestCreateAdvancedSettings_sendsNothingForAnEmptyConfiguration(t *testing.T) {
	t.Parallel()
	tests := map[string]advancedSettingsModel{
		"omitted maps": {
			StringSettings: types.MapNull(types.StringType),
			JSONSettings:   types.MapNull(jsontypes.NormalizedType{}),
		},
		"empty maps": {
			StringSettings: stringSettingsValue(map[string]string{}),
			JSONSettings:   jsonSettingsValue(map[string]string{}),
		},
	}
	for name, plan := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fake, client := newFakeKibana(t)

			result, diags := createAdvancedSettings(t.Context(), client, createRequest("default", plan))

			require.False(t, diags.HasError())
			assert.Empty(t, fake.recordedRequests())
			assert.Equal(t, "default", result.Model.ID.ValueString())
		})
	}
}

func TestCreateAdvancedSettings_failsWithoutWritingWhenTheSpaceDoesNotExist(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	plan := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"custom:teamNote": "hello"})}

	_, diags := createAdvancedSettings(t.Context(), client, createRequest("ops", plan))

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), "must exist")
	assert.Equal(t, []string{"GET /api/spaces/space/ops"}, fake.recordedRequests())
	assert.Empty(t, fake.appliedChanges())
}

func TestCreateAdvancedSettings_checksTheSpaceThenWritesToIt(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t, "ops")
	plan := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"custom:teamNote": "hello"})}

	result, diags := createAdvancedSettings(t.Context(), client, createRequest("ops", plan))

	require.False(t, diags.HasError())
	assert.Equal(t, []string{"GET /api/spaces/space/ops", "POST /s/ops/api/kibana/settings"}, fake.recordedRequests())
	assert.Equal(t, "ops", result.Model.ID.ValueString())
}

func TestCreateAdvancedSettings_surfacesKibanaRejectionAndProducesNoModel(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.rejection = "[validation [csv:quoteValues]]: expected value of type [boolean]"
	plan := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"csv:quoteValues": "false"})}

	result, diags := createAdvancedSettings(t.Context(), client, createRequest("default", plan))

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), "expected value of type [boolean]")
	assert.True(t, result.Model.ID.IsNull())
}

func TestCreateAdvancedSettings_sendsNothingWhenAJSONValueIsUnparsable(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"malformed text":                 "not json",
		"a second value after a boolean": "true false",
		"garbage after an object":        `{"a":1}garbage`,
		"garbage after an array":         "[1,2] x",
	}
	for name, text := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fake, client := newFakeKibana(t)
			plan := advancedSettingsModel{
				StringSettings: stringSettingsValue(map[string]string{"custom:teamNote": "hello"}),
				JSONSettings:   jsonSettingsValue(map[string]string{"custom:valid": "true", "custom:broken": text}),
			}

			result, diags := createAdvancedSettings(t.Context(), client, createRequest("default", plan))

			require.True(t, diags.HasError())
			assert.Empty(t, fake.recordedRequests())
			assert.True(t, result.Model.ID.IsNull())
		})
	}
}

func TestCreateAdvancedSettings_advisesJSONSettingsWhenAStringSettingIsRejected(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	const kibanaMessage = "[validation [csv:quoteValues]]: expected value of type [boolean] but got [string]"
	fake.rejection = kibanaMessage
	plan := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"csv:quoteValues": "maybe"})}

	_, diags := createAdvancedSettings(t.Context(), client, createRequest("default", plan))

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), kibanaMessage)
	assert.Contains(t, diags.Errors()[0].Detail(), "non-string values")
	assert.Contains(t, diags.Errors()[0].Detail(), "json_settings")
}

func TestCreateAdvancedSettings_doesNotAdviseJSONSettingsWhenOnlyJSONSettingsWereSent(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	const kibanaMessage = "[validation [histogram:barTarget]]: expected value of type [number] but got [string]"
	fake.rejection = kibanaMessage
	plan := advancedSettingsModel{JSONSettings: jsonSettingsValue(map[string]string{"histogram:barTarget": `"tall"`})}

	_, diags := createAdvancedSettings(t.Context(), client, createRequest("default", plan))

	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, kibanaMessage, diags.Errors()[0].Detail())
}

func TestCreateAdvancedSettings_aMissingSettingsRouteIsAnErrorNamingTheRouteAndProducesNoModel(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.missingSettingsRoute = true
	plan := advancedSettingsModel{JSONSettings: jsonSettingsValue(map[string]string{"csv:quoteValues": "false"})}

	result, diags := createAdvancedSettings(t.Context(), client, createRequest("default", plan))

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), "POST /api/kibana/settings")
	assert.True(t, result.Model.ID.IsNull())
}

func TestCreateAdvancedSettings_doesNotAdviseJSONSettingsOnAMissingSettingsRoute(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.missingSettingsRoute = true
	plan := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"})}

	_, diags := createAdvancedSettings(t.Context(), client, createRequest("default", plan))

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), "POST /api/kibana/settings")
	assert.NotContains(t, diags.Errors()[0].Detail(), "json_settings")
}

func TestCreateAdvancedSettings_keepsKibanasMessageAndAdvisesJSONSettingsForEveryKindOfRejection(t *testing.T) {
	t.Parallel()
	const missingPrivilegeMessage = "API [POST /api/kibana/settings] is unauthorized for user, " +
		"this action is granted by the Kibana privileges [manage_advanced_settings]"
	tests := map[string]struct {
		status  int
		message string
	}{
		"invalid value for a registered setting": {http.StatusBadRequest, "[validation [csv:quoteValues]]: expected value of type [boolean] but got [string]"},
		"setting pinned by server configuration": {http.StatusBadRequest, `Unable to update "observability:streamsEnableQueryStreams" because it is overridden`},
		"missing privilege":                      {http.StatusForbidden, missingPrivilegeMessage},
		"server error":                           {http.StatusInternalServerError, "Internal Server Error"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fake, client := newFakeKibana(t)
			fake.rejection, fake.rejectionStatus = tt.message, tt.status
			plan := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"csv:quoteValues": "maybe"})}

			_, diags := createAdvancedSettings(t.Context(), client, createRequest("default", plan))

			require.Len(t, diags.Errors(), 1)
			assert.Contains(t, diags.Errors()[0].Detail(), tt.message)
			assert.Contains(t, diags.Errors()[0].Detail(), jsonSettingsHint)
		})
	}
}

func TestCreateAdvancedSettings_advisesJSONSettingsWhenTheRequestFailsInTransit(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.dropSettingsWrites = true
	plan := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"csv:quoteValues": "maybe"})}

	_, diags := createAdvancedSettings(t.Context(), client, createRequest("default", plan))

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), "/api/kibana/settings")
	assert.Contains(t, diags.Errors()[0].Detail(), jsonSettingsHint)
}
