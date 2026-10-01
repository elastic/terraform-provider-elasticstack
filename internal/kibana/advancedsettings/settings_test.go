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
	"context"
	"testing"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/require"
)

func TestBuildChanges(t *testing.T) {
	t.Parallel()

	planned := map[string]jsontypes.Normalized{
		"dateFormat:tz":       jsontypes.NewNormalizedValue(`"Europe/Berlin"`),
		"discover:sampleSize": jsontypes.NewNormalizedValue(`500`),
		"theme:darkMode":      jsontypes.NewNormalizedValue(`true`),
		"custom:list":         jsontypes.NewNormalizedValue(`["a", "b"]`),
	}
	prior := map[string]jsontypes.Normalized{
		"dateFormat:tz":      jsontypes.NewNormalizedValue(`"UTC"`),
		"csv:separator":      jsontypes.NewNormalizedValue(`";"`),
		"discover:rowHeight": jsontypes.NewNormalizedValue(`3`),
	}

	changes, diags := buildChanges(planned, prior)
	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, map[string]any{
		"dateFormat:tz":       "Europe/Berlin",
		"discover:sampleSize": float64(500),
		"theme:darkMode":      true,
		"custom:list":         []any{"a", "b"},
		"csv:separator":       nil,
		"discover:rowHeight":  nil,
	}, changes)
}

func TestBuildChanges_invalidJSON(t *testing.T) {
	t.Parallel()

	_, diags := buildChanges(map[string]jsontypes.Normalized{
		"dateFormat:tz": jsontypes.NewNormalizedValue(`Europe/Berlin`),
	}, nil)
	require.True(t, diags.HasError())
}

func TestResetChanges(t *testing.T) {
	t.Parallel()

	changes := resetChanges(map[string]jsontypes.Normalized{
		"dateFormat:tz":  jsontypes.NewNormalizedValue(`"UTC"`),
		"theme:darkMode": jsontypes.NewNormalizedValue(`true`),
	})
	require.Equal(t, map[string]any{"dateFormat:tz": nil, "theme:darkMode": nil}, changes)
}

func TestTrackedSettingsFromAPI(t *testing.T) {
	t.Parallel()

	tracked := map[string]jsontypes.Normalized{
		"unchanged":  jsontypes.NewNormalizedValue(`{ "b": 2, "a": 1 }`),
		"drifted":    jsontypes.NewNormalizedValue(`"Europe/Berlin"`),
		"reset":      jsontypes.NewNormalizedValue(`true`),
		"html":       jsontypes.NewNormalizedValue(`"a<b"`),
		"number":     jsontypes.NewNormalizedValue(`500`),
		"overridden": jsontypes.NewNormalizedValue(`false`),
	}
	actual := map[string]kibanaoapi.AdvancedSetting{
		"unchanged":  {Value: map[string]any{"a": float64(1), "b": float64(2)}},
		"drifted":    {Value: "UTC"},
		"html":       {Value: "a<b&c"},
		"number":     {Value: float64(500)},
		"overridden": {Value: true, IsOverridden: true},
		"untracked":  {Value: "ignored"},
	}

	result, diags := trackedSettingsFromAPI(context.Background(), tracked, actual)
	require.False(t, diags.HasError(), "%v", diags)

	values := make(map[string]string, len(result))
	for key, value := range result {
		values[key] = value.ValueString()
	}
	require.Equal(t, map[string]string{
		// Semantically equal values keep the tracked formatting.
		"unchanged": `{ "b": 2, "a": 1 }`,
		"number":    `500`,
		// Values changed outside Terraform are reported as drift.
		"drifted":    `"UTC"`,
		"html":       `"a<b&c"`,
		"overridden": `true`,
	}, values)
}

func TestTrackedSettingsFromAPI_nilTracked(t *testing.T) {
	t.Parallel()

	result, diags := trackedSettingsFromAPI(context.Background(), nil, map[string]kibanaoapi.AdvancedSetting{
		"dateFormat:tz": {Value: "UTC"},
	})
	require.False(t, diags.HasError())
	require.Empty(t, result)
}

func TestSettingsModelRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := map[string]jsontypes.Normalized{
		"dateFormat:tz": jsontypes.NewNormalizedValue(`"UTC"`),
	}
	value, diags := settingsToModel(ctx, settings)
	require.False(t, diags.HasError())

	roundTripped, diags := settingsFromModel(ctx, value)
	require.False(t, diags.HasError())
	require.Equal(t, settings, roundTripped)
}
