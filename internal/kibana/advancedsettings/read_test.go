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

func TestReadAdvancedSettings_refreshesManagedKeysFromKibana(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t, "ops")
	fake.settings["ops"] = map[string]json.RawMessage{
		"dateFormat:tz":   json.RawMessage(`"Europe/Rome"`),
		"csv:quoteValues": json.RawMessage(`false`),
		"defaultRoute":    json.RawMessage(`"/app/home"`),
	}
	model := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "UTC"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"csv:quoteValues": "true"}),
	}

	state, found, diags := readAdvancedSettings(t.Context(), client, "ops", "ops", model)

	require.False(t, diags.HasError())
	assert.True(t, found)
	assert.Equal(t, []string{"GET /api/spaces/space/ops", "GET /s/ops/api/kibana/settings"}, fake.recordedRequests())
	assert.Equal(t, stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"}), state.StringSettings)
	assert.Equal(t, jsonSettingsValue(map[string]string{"csv:quoteValues": "false"}), state.JSONSettings)
	assert.Equal(t, "ops", state.ID.ValueString())
	assert.Equal(t, "ops", state.SpaceID.ValueString())
}

func TestReadAdvancedSettings_reportsAFailedReadWithoutDroppingTheResource(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.failReads = true
	model := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "UTC"})}

	state, found, diags := readAdvancedSettings(t.Context(), client, "default", "default", model)

	require.True(t, diags.HasError())
	assert.False(t, found)
	assert.Equal(t, model.StringSettings, state.StringSettings)
}

func TestReadAdvancedSettings_removesTheResourceWhenTheSpaceNoLongerExists(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.settings["ops"] = map[string]json.RawMessage{"dateFormat:tz": json.RawMessage(`"Europe/Rome"`)}
	model := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"})}

	_, found, diags := readAdvancedSettings(t.Context(), client, "ops", "ops", model)

	require.False(t, diags.HasError())
	assert.False(t, found)
	assert.Equal(t, []string{"GET /api/spaces/space/ops"}, fake.recordedRequests())
}

func TestReadAdvancedSettings_keepsTheResourceForTheDefaultSpaceWithoutLookingItUp(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.settings["default"] = map[string]json.RawMessage{"dateFormat:tz": json.RawMessage(`"Europe/Rome"`)}
	model := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "UTC"})}

	state, found, diags := readAdvancedSettings(t.Context(), client, "default", "default", model)

	require.False(t, diags.HasError())
	assert.True(t, found)
	assert.Equal(t, []string{"GET /api/kibana/settings"}, fake.recordedRequests())
	assert.Equal(t, stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"}), state.StringSettings)
}

func TestReadAdvancedSettings_aFailedSpaceCheckIsAnErrorAndDoesNotRemoveTheResource(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t, "ops")
	fake.failSpaceLookups = true
	model := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "UTC"})}

	state, found, diags := readAdvancedSettings(t.Context(), client, "ops", "ops", model)

	require.True(t, diags.HasError())
	assert.False(t, found)
	assert.Equal(t, model.StringSettings, state.StringSettings)
	assert.Equal(t, []string{"GET /api/spaces/space/ops"}, fake.recordedRequests())
}

func TestReadAdvancedSettings_aMissingSettingsRouteIsAnErrorNamingTheRouteAndKeepsTheManagedKeys(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	fake.missingSettingsRoute = true
	model := advancedSettingsModel{StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"})}

	state, found, diags := readAdvancedSettings(t.Context(), client, "default", "default", model)

	require.Len(t, diags.Errors(), 1)
	assert.Contains(t, diags.Errors()[0].Detail(), "GET /api/kibana/settings")
	assert.False(t, found)
	assert.Equal(t, model.StringSettings, state.StringSettings)
}

func TestReadAdvancedSettings_dropsManagedKeysWhenKibanaHoldsNoUserSetSettings(t *testing.T) {
	t.Parallel()
	fake, client := newFakeKibana(t)
	model := advancedSettingsModel{
		StringSettings: stringSettingsValue(map[string]string{"dateFormat:tz": "Europe/Rome"}),
		JSONSettings:   jsonSettingsValue(map[string]string{"csv:quoteValues": "false"}),
	}

	state, found, diags := readAdvancedSettings(t.Context(), client, "default", "default", model)

	require.False(t, diags.HasError())
	assert.True(t, found)
	assert.Equal(t, []string{"GET /api/kibana/settings"}, fake.recordedRequests())
	assert.Equal(t, stringSettingsValue(map[string]string{}), state.StringSettings)
	assert.Equal(t, jsonSettingsValue(map[string]string{}), state.JSONSettings)
}

func TestReadAdvancedSettings_keepsOmittedMapsNullWhenKibanaHoldsNoUserSetSettings(t *testing.T) {
	t.Parallel()
	_, client := newFakeKibana(t)
	model := advancedSettingsModel{
		StringSettings: types.MapNull(types.StringType),
		JSONSettings:   types.MapNull(jsontypes.NormalizedType{}),
	}

	state, found, diags := readAdvancedSettings(t.Context(), client, "default", "default", model)

	require.False(t, diags.HasError())
	assert.True(t, found)
	assert.True(t, state.StringSettings.IsNull())
	assert.True(t, state.JSONSettings.IsNull())
}
