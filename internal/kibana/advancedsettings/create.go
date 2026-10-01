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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createAdvancedSettings(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[advancedSettingsModel],
) (entitycore.KibanaWriteResult[advancedSettingsModel], diag.Diagnostics) {
	return writeAdvancedSettings(ctx, client, req)
}

// writeAdvancedSettings applies the planned settings for Create and Update.
//
// The final state is built from the update response rather than from a
// separate read. Kibana caches advanced settings per instance for a few
// seconds and only invalidates the cache of the instance that handled the
// write, so a follow-up GET routed to another Kibana instance can return the
// previous values. The update response is computed by the instance that
// performed the write after invalidating its cache, so it reflects the
// persisted settings.
func writeAdvancedSettings(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[advancedSettingsModel],
) (entitycore.KibanaWriteResult[advancedSettingsModel], diag.Diagnostics) {
	plan := req.Plan

	planned, diags := settingsFromModel(ctx, plan.Settings)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[advancedSettingsModel]{}, diags
	}

	var prior map[string]jsontypes.Normalized
	if req.Prior != nil {
		priorSettings, priorDiags := settingsFromModel(ctx, req.Prior.Settings)
		diags.Append(priorDiags...)
		prior = priorSettings
	}
	if diags.HasError() {
		return entitycore.KibanaWriteResult[advancedSettingsModel]{}, diags
	}

	changes, changeDiags := buildChanges(planned, prior)
	diags.Append(changeDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[advancedSettingsModel]{}, diags
	}

	actual, apiDiags := kibanaoapi.UpdateAdvancedSettings(ctx, client.GetKibanaOapiClient(), plan.SpaceID.ValueString(), plan.Global.ValueBool(), changes)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[advancedSettingsModel]{}, diags
	}

	settings, settingsDiags := trackedSettingsFromAPI(ctx, planned, actual)
	diags.Append(settingsDiags...)
	settingsValue, mapDiags := settingsToModel(ctx, settings)
	diags.Append(mapDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[advancedSettingsModel]{}, diags
	}

	plan.ID = types.StringValue(plan.scopeID())
	plan.Settings = settingsValue
	return entitycore.KibanaWriteResult[advancedSettingsModel]{Model: plan, SkipReadAfterWrite: true}, diags
}
