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
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = newResource()
	_ resource.ResourceWithConfigure      = newResource()
	_ resource.ResourceWithImportState    = newResource()
	_ resource.ResourceWithValidateConfig = newResource()
)

type Resource struct {
	*entitycore.KibanaResource[advancedSettingsModel]
}

func newResource() *Resource {
	return &Resource{
		KibanaResource: entitycore.NewKibanaResource[advancedSettingsModel](
			entitycore.ComponentKibana,
			"advanced_settings",
			entitycore.KibanaResourceOptions[advancedSettingsModel]{
				Schema: getSchema,
				Read:   readAdvancedSettings,
				Delete: deleteAdvancedSettings,
				Create: createAdvancedSettings,
				Update: updateAdvancedSettings,
			},
		),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newResource()
}

// ValidateConfig rejects a space_id on the global scope and setting values
// that would reset a setting instead of managing it.
func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var global types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrGlobal), &global)...)
	var spaceID types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrSpaceID), &spaceID)...)
	var settingsValue types.Map
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrSettings), &settingsValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if global.ValueBool() && !spaceID.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root(attrSpaceID),
			"Conflicting attributes",
			"space_id cannot be set when global is true; global advanced settings apply to every space.",
		)
	}

	if settingsValue.IsNull() || settingsValue.IsUnknown() {
		return
	}
	var settings map[string]jsontypes.Normalized
	resp.Diagnostics.Append(settingsValue.ElementsAs(ctx, &settings, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for key, value := range settings {
		if value.IsUnknown() {
			continue
		}
		if value.IsNull() || isJSONNull(value.ValueString()) {
			resp.Diagnostics.AddAttributeError(
				path.Root(attrSettings).AtMapKey(key),
				"Invalid advanced setting value",
				fmt.Sprintf("The value of setting %q must not be null. Remove the setting from the map to reset it to its Kibana default.", key),
			)
		}
	}
}

func isJSONNull(value string) bool {
	var decoded any
	return json.Unmarshal([]byte(value), &decoded) == nil && decoded == nil
}

// ImportState imports the advanced settings scope. The import ID is either a
// space ID or "global". Settings are not imported; declare the settings to
// manage in the configuration.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"The import ID must be a Kibana space ID, or \"global\" for the global advanced settings.",
		)
		return
	}

	global := req.ID == globalScopeID
	spaceID := req.ID
	if global {
		spaceID = clients.DefaultSpaceID
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attrID), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attrSpaceID), spaceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attrGlobal), global)...)
}
