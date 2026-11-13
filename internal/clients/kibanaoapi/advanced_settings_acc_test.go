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

package kibanaoapi_test

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccAdvancedSettings_NullChangeResetsKey(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	notFound, diags := kibanaoapi.ApplyAdvancedSettings(t.Context(), client, spaceID, map[string]any{
		"custom:reset":  "x",
		"custom:keep":   "y",
		"dateFormat:tz": "UTC",
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.False(t, notFound)

	notFound, diags = kibanaoapi.ApplyAdvancedSettings(t.Context(), client, spaceID, map[string]any{
		"custom:reset":  nil,
		"dateFormat:tz": nil,
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.False(t, notFound)

	settings, diags := kibanaoapi.GetAdvancedSettings(t.Context(), client, spaceID)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.NotContains(t, settings, "custom:reset")
	assert.NotContains(t, settings, "dateFormat:tz")
	assert.JSONEq(t, `"y"`, string(settings["custom:keep"]))
}

func TestAccAdvancedSettings_ValueEqualToKibanaDefaultIsRetained(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	notFound, diags := kibanaoapi.ApplyAdvancedSettings(t.Context(), client, spaceID, map[string]any{
		"csv:separator":      ",",
		"csv:quoteValues":    true,
		"truncate:maxHeight": 115,
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.False(t, notFound)

	settings, diags := kibanaoapi.GetAdvancedSettings(t.Context(), client, spaceID)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.JSONEq(t, `","`, string(settings["csv:separator"]))
	assert.JSONEq(t, `true`, string(settings["csv:quoteValues"]))
	assert.JSONEq(t, `115`, string(settings["truncate:maxHeight"]))
}

func TestAccAdvancedSettings_MissingSpaceIsNotReportedAsNotFound(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	missingSpaceID := "advset-missing-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	settings, diags := kibanaoapi.GetAdvancedSettings(t.Context(), client, missingSpaceID)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.NotNil(t, settings)

	notFound, diags := kibanaoapi.ApplyAdvancedSettings(t.Context(), client, missingSpaceID, map[string]any{"custom:a": nil})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.False(t, notFound)
}

func TestAccAdvancedSettings_DeletedSpaceKeepsServingItsSettings(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	_, diags := kibanaoapi.ApplyAdvancedSettings(t.Context(), client, spaceID, map[string]any{"custom:a": "x"})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	diags = kibanaoapi.DeleteSpace(t.Context(), client, spaceID)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	settings, diags := kibanaoapi.GetAdvancedSettings(t.Context(), client, spaceID)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.JSONEq(t, `"x"`, string(settings["custom:a"]))
}

// The stack pins observability:streamsEnableQueryStreams through uiSettings.overrides in kibana-9.4.yml.
func TestAccAdvancedSettings_SettingPinnedByKibanaConfigurationIsRejected(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	const pinnedKey = "observability:streamsEnableQueryStreams"

	for name, change := range map[string]any{"new value": false, "reset": nil} {
		t.Run(name, func(t *testing.T) {
			notFound, diags := kibanaoapi.ApplyAdvancedSettings(t.Context(), client, spaceID, map[string]any{pinnedKey: change})

			assert.False(t, notFound)
			require.Len(t, diags.Errors(), 1)
			assert.Contains(t, diags.Errors()[0].Detail(), pinnedKey)
			assert.Contains(t, diags.Errors()[0].Detail(), "overridden")
		})
	}

	settings, diags := kibanaoapi.GetAdvancedSettings(t.Context(), client, spaceID)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.JSONEq(t, `true`, string(settings[pinnedKey]))
}

func TestAccAdvancedSettings_MixedValueTypesRoundTrip(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	notFound, diags := kibanaoapi.ApplyAdvancedSettings(t.Context(), client, spaceID, map[string]any{
		"custom:string":  "x",
		"custom:boolean": true,
		"custom:number":  json.Number("42.5"),
		"custom:array":   json.RawMessage(`[1,"two",null]`),
		"custom:object":  json.RawMessage(`{"nested":{"k":1}}`),
	})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.False(t, notFound)

	settings, diags := kibanaoapi.GetAdvancedSettings(t.Context(), client, spaceID)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.JSONEq(t, `"x"`, string(settings["custom:string"]))
	assert.JSONEq(t, `true`, string(settings["custom:boolean"]))
	assert.JSONEq(t, `42.5`, string(settings["custom:number"]))
	assert.JSONEq(t, `[1,"two",null]`, string(settings["custom:array"]))
	assert.JSONEq(t, `{"nested":{"k":1}}`, string(settings["custom:object"]))
}

func TestAccAdvancedSettings_DefaultSpaceIsReadable(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)

	settings, diags := kibanaoapi.GetAdvancedSettings(t.Context(), client, "default")

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.NotNil(t, settings)
}

func TestAccAdvancedSettings_SpacesAreIndependent(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	opsSpaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	otherSpaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	for spaceID, timezone := range map[string]string{opsSpaceID: "UTC", otherSpaceID: "Europe/Rome"} {
		_, diags := kibanaoapi.ApplyAdvancedSettings(t.Context(), client, spaceID, map[string]any{"dateFormat:tz": timezone})
		require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	}
	_, diags := kibanaoapi.ApplyAdvancedSettings(t.Context(), client, opsSpaceID, map[string]any{"dateFormat:tz": nil})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	opsSettings, diags := kibanaoapi.GetAdvancedSettings(t.Context(), client, opsSpaceID)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	otherSettings, diags := kibanaoapi.GetAdvancedSettings(t.Context(), client, otherSpaceID)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.NotContains(t, opsSettings, "dateFormat:tz")
	assert.JSONEq(t, `"Europe/Rome"`, string(otherSettings["dateFormat:tz"]))
}
