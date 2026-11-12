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
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
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

func writeAdvancedSettings(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[advancedSettingsModel],
) (entitycore.KibanaWriteResult[advancedSettingsModel], diag.Diagnostics) {
	plan := req.Plan

	diags := ensureSpaceExists(ctx, req.SpaceID, spacesAPIGetter(client))
	if diags.HasError() {
		return entitycore.KibanaWriteResult[advancedSettingsModel]{}, diags
	}

	changes, changeDiags := buildChanges(plan, req.Prior)
	diags.Append(changeDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[advancedSettingsModel]{}, diags
	}

	diags.Append(applyChanges(ctx, client, req.SpaceID, changes, sendsStringSetting(plan, changes))...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[advancedSettingsModel]{}, diags
	}

	plan.ID = types.StringValue(req.SpaceID)
	plan.SpaceID = types.StringValue(req.SpaceID)

	return entitycore.KibanaWriteResult[advancedSettingsModel]{Model: plan}, diags
}

func applyChanges(ctx context.Context, client *clients.KibanaScopedClient, spaceID string, changes map[string]any, adviseJSONSettings bool) diag.Diagnostics {
	if len(changes) == 0 {
		return nil
	}
	routeMissing, diags := kibanaoapi.ApplyAdvancedSettings(ctx, client.GetKibanaOapiClient(), spaceID, changes)
	if adviseJSONSettings {
		diags = withJSONSettingsHint(diags)
	}
	if routeMissing {
		diags.Append(missingSettingsRouteDiagnostic(http.MethodPost))
	}
	return diags
}
