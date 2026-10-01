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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readAdvancedSettings(ctx context.Context, client *clients.KibanaScopedClient, _, _ string, model advancedSettingsModel) (advancedSettingsModel, bool, diag.Diagnostics) {
	oapiClient := client.GetKibanaOapiClient()

	// Kibana serves the settings of a space that does not exist (and creates
	// them on write), so check the space explicitly to detect its deletion.
	if !model.Global.ValueBool() {
		space, diags := kibanaoapi.GetSpace(ctx, oapiClient, model.SpaceID.ValueString())
		if diags.HasError() || space == nil {
			return model, false, diags
		}
	}

	actual, diags := kibanaoapi.GetAdvancedSettings(ctx, oapiClient, model.SpaceID.ValueString(), model.Global.ValueBool())
	if diags.HasError() {
		return model, false, diags
	}
	if actual == nil {
		return model, false, diags
	}

	model.ID = types.StringValue(model.scopeID())

	// After import only the scope is known; leave settings null until the
	// configuration declares which settings to manage.
	if model.Settings.IsNull() {
		return model, true, diags
	}

	tracked, trackedDiags := settingsFromModel(ctx, model.Settings)
	diags.Append(trackedDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	settings, settingsDiags := trackedSettingsFromAPI(ctx, tracked, actual)
	diags.Append(settingsDiags...)
	settingsValue, mapDiags := settingsToModel(ctx, settings)
	diags.Append(mapDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	model.Settings = settingsValue
	return model, true, diags
}
