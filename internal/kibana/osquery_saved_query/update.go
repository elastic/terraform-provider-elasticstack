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

package osquerysavedquery

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateOsquerySavedQuery(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[osquerySavedQueryModel],
) (entitycore.KibanaWriteResult[osquerySavedQueryModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	body, bodyDiags := plan.toAPIUpdateRequest(ctx, req.Prior)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[osquerySavedQueryModel]{}, diags
	}

	oapiClient := client.GetKibanaOapiClient()

	entity, updateDiags := updateOsquerySavedQueryWithBestID(ctx, oapiClient, req.SpaceID, req.WriteID, plan, req.Prior, body)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[osquerySavedQueryModel]{}, diags
	}

	diags.Append(prebuiltGuardDiagnostic(entity.Prebuilt)...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[osquerySavedQueryModel]{}, diags
	}

	diags.Append(plan.populateFromUpdateAPI(ctx, entity)...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[osquerySavedQueryModel]{}, diags
	}

	return entitycore.KibanaWriteResult[osquerySavedQueryModel]{Model: plan}, diags
}

func updateOsquerySavedQueryWithBestID(
	ctx context.Context,
	oapiClient *kibanaoapi.Client,
	spaceID string,
	savedQueryID string,
	plan osquerySavedQueryModel,
	prior *osquerySavedQueryModel,
	body kbapi.OsqueryUpdateSavedQueryJSONRequestBody,
) (*kibanaoapi.OsquerySavedQueryUpdateEntity, diag.Diagnostics) {
	if typeutils.IsKnown(plan.SavedObjectID) && plan.SavedObjectID.ValueString() != "" {
		return kibanaoapi.UpdateOsquerySavedQueryBySavedObjectID(ctx, oapiClient, spaceID, plan.SavedObjectID.ValueString(), body)
	}
	if prior != nil && typeutils.IsKnown(prior.SavedObjectID) && prior.SavedObjectID.ValueString() != "" {
		return kibanaoapi.UpdateOsquerySavedQueryBySavedObjectID(ctx, oapiClient, spaceID, prior.SavedObjectID.ValueString(), body)
	}

	return kibanaoapi.UpdateOsquerySavedQuery(ctx, oapiClient, spaceID, savedQueryID, body)
}
