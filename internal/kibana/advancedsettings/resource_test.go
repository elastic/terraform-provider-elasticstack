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
	"maps"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func resourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	newResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	return resp
}

// buildConfig returns a configuration where every attribute is null except
// the given overrides.
func buildConfig(t *testing.T, overrides map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
	schemaResp := resourceSchema(t)
	objectType := schemaResp.Schema.Type().TerraformType(context.Background()).(tftypes.Object)

	values := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attrType := range objectType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	maps.Copy(values, overrides)

	return tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(objectType, values),
	}
}

func settingsValue(settings map[string]any) tftypes.Value {
	mapType := tftypes.Map{ElementType: tftypes.String}
	values := make(map[string]tftypes.Value, len(settings))
	for key, value := range settings {
		values[key] = tftypes.NewValue(tftypes.String, value)
	}
	return tftypes.NewValue(mapType, values)
}

func validateConfig(t *testing.T, overrides map[string]tftypes.Value) resource.ValidateConfigResponse {
	t.Helper()
	var resp resource.ValidateConfigResponse
	newResource().ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: buildConfig(t, overrides)}, &resp)
	return resp
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		overrides map[string]tftypes.Value
		wantError bool
	}{
		{
			name: "space settings",
			overrides: map[string]tftypes.Value{
				attrSpaceID:  tftypes.NewValue(tftypes.String, "team-a"),
				attrSettings: settingsValue(map[string]any{"dateFormat:tz": `"UTC"`}),
			},
		},
		{
			name: "global settings",
			overrides: map[string]tftypes.Value{
				attrGlobal:   tftypes.NewValue(tftypes.Bool, true),
				attrSettings: settingsValue(map[string]any{"hideAnnouncements": `true`}),
			},
		},
		{
			name: "global with space_id",
			overrides: map[string]tftypes.Value{
				attrGlobal:   tftypes.NewValue(tftypes.Bool, true),
				attrSpaceID:  tftypes.NewValue(tftypes.String, "default"),
				attrSettings: settingsValue(map[string]any{"hideAnnouncements": `true`}),
			},
			wantError: true,
		},
		{
			name: "JSON null value",
			overrides: map[string]tftypes.Value{
				attrSettings: settingsValue(map[string]any{"dateFormat:tz": `null`}),
			},
			wantError: true,
		},
		{
			name: "null map element",
			overrides: map[string]tftypes.Value{
				attrSettings: settingsValue(map[string]any{"dateFormat:tz": nil}),
			},
			wantError: true,
		},
		{
			name: "unknown settings",
			overrides: map[string]tftypes.Value{
				attrSettings: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
			},
		},
		{
			name: "unknown setting value",
			overrides: map[string]tftypes.Value{
				attrSettings: settingsValue(map[string]any{"dateFormat:tz": tftypes.UnknownValue}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := validateConfig(t, tt.overrides)
			require.Equal(t, tt.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}

func importState(t *testing.T, id string) (resource.ImportStateResponse, advancedSettingsModel) {
	t.Helper()
	schemaResp := resourceSchema(t)
	objectType := schemaResp.Schema.Type().TerraformType(context.Background())
	resp := resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(objectType, nil),
		},
	}
	newResource().ImportState(context.Background(), resource.ImportStateRequest{ID: id}, &resp)

	var model advancedSettingsModel
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Get(context.Background(), &model)...)
	}
	return resp, model
}

func TestImportState(t *testing.T) {
	t.Parallel()

	t.Run("space", func(t *testing.T) {
		t.Parallel()
		resp, model := importState(t, "team-a")
		require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		require.Equal(t, types.StringValue("team-a"), model.ID)
		require.Equal(t, types.StringValue("team-a"), model.SpaceID)
		require.Equal(t, types.BoolValue(false), model.Global)
		require.True(t, model.Settings.IsNull())
	})

	t.Run("global", func(t *testing.T) {
		t.Parallel()
		resp, model := importState(t, globalScopeID)
		require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		require.Equal(t, types.StringValue(globalScopeID), model.ID)
		require.Equal(t, types.StringValue(clients.DefaultSpaceID), model.SpaceID)
		require.Equal(t, types.BoolValue(true), model.Global)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		resp, _ := importState(t, "")
		require.True(t, resp.Diagnostics.HasError())
	})
}

func TestModelScope(t *testing.T) {
	t.Parallel()

	space := advancedSettingsModel{SpaceID: types.StringValue("team-a"), Global: types.BoolValue(false)}
	require.Equal(t, types.StringValue("team-a"), space.GetResourceID())
	require.False(t, space.IsUnscopedSpace())
	reqs, diags := space.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Empty(t, reqs)

	global := advancedSettingsModel{SpaceID: types.StringValue("default"), Global: types.BoolValue(true)}
	require.Equal(t, types.StringValue(globalScopeID), global.GetResourceID())
	require.True(t, global.IsUnscopedSpace())
	reqs, diags = global.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.Equal(t, *minGlobalSettingsVersion, reqs[0].MinVersion)
	require.Equal(t, path.Root(attrGlobal), *reqs[0].AttributePath)
}
