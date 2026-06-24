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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readOsquerySavedQuery(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model osquerySavedQueryModel) (osquerySavedQueryModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()

	entity, getDiags := getOsquerySavedQueryForRead(ctx, oapiClient, spaceID, resourceID, model)
	diags.Append(getDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	if entity == nil {
		return model, false, diags
	}

	model.SpaceID = types.StringValue(spaceID)

	diags.Append(model.populateFromGetAPI(ctx, entity)...)
	if diags.HasError() {
		return model, false, diags
	}

	diags.Append(prebuiltGuardDiagnostic(entity.Prebuilt)...)
	if diags.HasError() {
		return model, false, diags
	}

	return model, true, diags
}

func getOsquerySavedQueryForRead(
	ctx context.Context,
	oapiClient *kibanaoapi.Client,
	spaceID string,
	savedQueryID string,
	model osquerySavedQueryModel,
) (*kibanaoapi.OsquerySavedQueryGetEntity, diag.Diagnostics) {
	if typeutils.IsKnown(model.SavedObjectID) && model.SavedObjectID.ValueString() != "" {
		return kibanaoapi.GetOsquerySavedQueryBySavedObjectID(ctx, oapiClient, spaceID, model.SavedObjectID.ValueString())
	}

	return kibanaoapi.GetOsquerySavedQuery(ctx, oapiClient, spaceID, savedQueryID)
}
