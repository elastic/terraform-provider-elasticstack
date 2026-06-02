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

package entity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createEntity(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tfModel],
) (entitycore.KibanaWriteResult[tfModel], diag.Diagnostics) {
	plan := req.Plan
	spaceID := normalizeSpaceID(plan.SpaceID)
	entityType := plan.EntityType.ValueString()
	entityID := plan.EntityID.ValueString()

	if (plan.Entity.IsNull() || plan.Entity.IsUnknown()) && (plan.EntityJSON.IsNull() || plan.EntityJSON.IsUnknown()) {
		return entitycore.KibanaWriteResult[tfModel]{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Missing entity data", "Either entity or entity_json must be provided"),
		}
	}

	bodyMap, diags := modelToAPIBody(ctx, plan)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	if entityMap, ok := bodyMap["entity"].(map[string]any); ok {
		entityMap["id"] = entityID
		bodyMap["entity"] = entityMap
	} else {
		bodyMap["entity"] = map[string]any{"id": entityID}
	}

	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return entitycore.KibanaWriteResult[tfModel]{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("JSON marshal error", err.Error()),
		}
	}

	if d := kibanaoapi.CreateSecurityEntityStoreEntity(ctx, client.GetKibanaOapiClient(), spaceID, entityType, bytes.NewReader(bodyBytes)); d.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, d
	}

	// Authoritative read-after-write
	readModel, found, readDiags := readEntity(ctx, client, buildID(spaceID, entityID), spaceID, plan)
	if readDiags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, readDiags
	}
	if !found {
		return entitycore.KibanaWriteResult[tfModel]{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Entity not found after create", fmt.Sprintf("Entity %s was not found after creation", entityID)),
		}
	}

	readModel.ID = types.StringValue(buildID(spaceID, entityID))
	readModel.SpaceID = types.StringValue(spaceID)
	readModel.EntityType = types.StringValue(entityType)
	readModel.EntityID = types.StringValue(entityID)
	readModel.Force = plan.Force

	return entitycore.KibanaWriteResult[tfModel]{Model: readModel}, nil
}
