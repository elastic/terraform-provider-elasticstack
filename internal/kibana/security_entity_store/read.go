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

package security_entity_store

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readEntityStore(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	_ string,
	spaceID string,
	model tfModel,
) (tfModel, bool, diag.Diagnostics) {
	status, rawBody, diags := waitForStarted(ctx, client, spaceID)
	if diags.HasError() {
		return model, true, diags
	}

	if status.Status == kbapi.SecurityEntityAnalyticsAPIStoreStatusNotInstalled {
		return model, false, diags
	}

	entityTypes, started, logExtraction, flattenDiags := flattenStatus(ctx, status.Engines)
	if flattenDiags.HasError() {
		return model, true, flattenDiags
	}

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: resourceID}).String())
	model.SpaceID = types.StringValue(spaceID)
	model.EntityTypes = entityTypes
	model.Started = types.BoolValue(started)
	model.LogExtraction = logExtraction
	model.StatusJSON = jsontypes.NewNormalizedValue(string(rawBody))
	return model, true, append(diags, flattenDiags...)
}
