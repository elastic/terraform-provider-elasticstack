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

func TestDeleteAdvancedSettings_succeedsWithoutASettingsRequestWhenTheSpaceNoLongerExists(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	model := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"})}

	diags := deleteAdvancedSettings(t.Context(), client, "ops", "ops", model)

	require.False(t, diags.HasError())
	assert.Equal(t, []string{"GET /api/spaces/space/ops"}, fake.recordedRequests())
	assert.Empty(t, fake.appliedChanges())
}

func TestDeleteAdvancedSettings_aFailedSpaceCheckIsAnErrorAndSendsNothing(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t, "ops")
	fake.failSpaceLookups = true
	model := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"})}

	diags := deleteAdvancedSettings(t.Context(), client, "ops", "ops", model)

	require.True(t, diags.HasError())
	assert.Equal(t, []string{"GET /api/spaces/space/ops"}, fake.recordedRequests())
	assert.Empty(t, fake.appliedChanges())
}

func TestDeleteAdvancedSettings_sendsNothingWhenNoKeysAreManaged(t *testing.T) {
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
	for name, model := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fake, client := newFakeKibana(t)

			diags := deleteAdvancedSettings(t.Context(), client, "default", "default", model)

			require.False(t, diags.HasError())
			assert.Empty(t, fake.recordedRequests())
		})
	}
}

func TestDeleteAdvancedSettings_resetsEveryManagedKeyInOneRequest(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	model := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome", "custom:teamNote": "hello"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"csv:quoteValues": "false", "defaultColumns": `["@timestamp"]`}),
	}

	diags := deleteAdvancedSettings(t.Context(), client, "default", "default", model)

	require.False(t, diags.HasError())
	assert.Equal(t, []string{"POST /api/kibana/settings"}, fake.recordedRequests())
	assert.Equal(t, []map[string]json.RawMessage{{
		"dateFormat:tz":   json.RawMessage(`null`),
		"custom:teamNote": json.RawMessage(`null`),
		"csv:quoteValues": json.RawMessage(`null`),
		"defaultColumns":  json.RawMessage(`null`),
	}}, fake.appliedChanges())
}

func TestDeleteAdvancedSettings_checksTheSpaceThenResetsInIt(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t, "ops")
	model := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"})}

	diags := deleteAdvancedSettings(t.Context(), client, "ops", "ops", model)

	require.False(t, diags.HasError())
	assert.Equal(t, []string{"GET /api/spaces/space/ops", "POST /s/ops/api/kibana/settings"}, fake.recordedRequests())
}

func TestDeleteAdvancedSettings_surfacesKibanaRejectionOfTheReset(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	const kibanaMessage = `Unable to update "dateFormat:tz" because it is overridden`
	fake.rejection = kibanaMessage
	model := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"})}

	diags := deleteAdvancedSettings(t.Context(), client, "default", "default", model)

	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, kibanaMessage, diags.Errors()[0].Detail())
}

func TestDeleteAdvancedSettings_aMissingSettingsRouteIsAnErrorNamingTheRoute(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.missingSettingsRoute = true
	model := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"})}

	diags := deleteAdvancedSettings(t.Context(), client, "default", "default", model)

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), "POST /api/kibana/settings")
}
