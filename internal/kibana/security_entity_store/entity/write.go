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
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// createRetryPollInterval is the cadence at which the entity create call is
// retried while the entity store is still initializing (HTTP 500). The overall
// retry budget is bounded by the Create ctx deadline (from the resource
// timeouts block), not by this interval.
const createRetryPollInterval = 5 * time.Second

func writeEntity(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tfModel],
) (entitycore.KibanaWriteResult[tfModel], diag.Diagnostics) {
	plan := req.Plan

	if req.Prior == nil {
		if (plan.Entity.IsNull() || plan.Entity.IsUnknown()) && (plan.EntityJSON.IsNull() || plan.EntityJSON.IsUnknown()) {
			return entitycore.KibanaWriteResult[tfModel]{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Missing entity data", "Either entity or entity_json must be provided"),
			}
		}
	}

	spaceID := NormalizeSpaceID(plan.SpaceID)
	entityType := plan.EntityType.ValueString()
	entityID := plan.EntityID.ValueString()

	bodyMap, bodyDiags := modelToAPIBody(ctx, plan)
	var diags diag.Diagnostics
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	bodyBytes, marshalDiags := injectEntityIDAndMarshal(bodyMap, entityID)
	diags.Append(marshalDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	if req.Prior == nil {
		if d := createEntityWithRetry(ctx, client, spaceID, entityType, bodyBytes); d.HasError() {
			return entitycore.KibanaWriteResult[tfModel]{}, d
		}
	} else {
		force := false
		if !plan.Force.IsNull() && !plan.Force.IsUnknown() {
			force = plan.Force.ValueBool()
		}

		if d := kibanaoapi.UpdateSecurityEntityStoreEntity(ctx, client.GetKibanaOapiClient(), spaceID, entityType, bytes.NewReader(bodyBytes), force); d.HasError() {
			return entitycore.KibanaWriteResult[tfModel]{}, d
		}
	}

	return entitycore.KibanaWriteResult[tfModel]{Model: plan}, nil
}

// createEntityWithRetry performs an initial synchronous create and, when the
// entity store returns HTTP 500 (still initializing), retries until the create
// succeeds or the Create ctx deadline is exceeded, reusing the shared
// kibanaoapi retry helper. A fresh body reader is constructed on each attempt
// because the request consumes it.
func createEntityWithRetry(ctx context.Context, client *clients.KibanaScopedClient, spaceID, entityType string, bodyBytes []byte) diag.Diagnostics {
	attempt := func(ctx context.Context) (int, []byte, error) {
		return kibanaoapi.CreateSecurityEntityStoreEntityStatus(ctx, client.GetKibanaOapiClient(), spaceID, entityType, bytes.NewReader(bodyBytes))
	}
	return kibanaoapi.RetryCreateOnServerError(ctx, "security entity store entity", entityType, attempt, createRetryPollInterval)
}

// injectEntityIDAndMarshal sets entity.id in bodyMap and marshals it to JSON.
func injectEntityIDAndMarshal(bodyMap map[string]any, entityID string) ([]byte, diag.Diagnostics) {
	if entityMap, ok := bodyMap["entity"].(map[string]any); ok {
		entityMap["id"] = entityID
		bodyMap["entity"] = entityMap
	} else {
		bodyMap["entity"] = map[string]any{"id": entityID}
	}

	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("JSON marshal error", err.Error()),
		}
	}
	return bodyBytes, nil
}
