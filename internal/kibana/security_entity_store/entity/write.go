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
	"net/http"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
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
// entity store returns HTTP 500 (still initializing), retries at
// createRetryPollInterval until the create succeeds (HTTP 2xx) or the Create
// ctx deadline is exceeded. Any non-500 non-2xx response fails fast. The retry
// is bounded solely by the ctx deadline; no separate wall-clock budget is used.
func createEntityWithRetry(ctx context.Context, client *clients.KibanaScopedClient, spaceID, entityType string, bodyBytes []byte) diag.Diagnostics {
	attempt := func(ctx context.Context) (int, []byte, error) {
		return kibanaoapi.CreateSecurityEntityStoreEntityStatus(ctx, client.GetKibanaOapiClient(), spaceID, entityType, bytes.NewReader(bodyBytes))
	}
	return retryCreateOnServerError(ctx, "security entity store entity", entityType, attempt, createRetryPollInterval)
}

// createAttemptFunc performs a single create call and reports the raw HTTP
// status code and body (or a transport error).
type createAttemptFunc func(ctx context.Context) (statusCode int, body []byte, err error)

// retryCreateOnServerError performs an initial synchronous create attempt and,
// when the store returns HTTP 500, retries at createRetryPollInterval until the
// create succeeds (HTTP 2xx) or the Create ctx deadline is exceeded. Any
// non-500 non-2xx response fails fast.
func retryCreateOnServerError(ctx context.Context, resourceType, resourceID string, attempt createAttemptFunc, pollInterval time.Duration) diag.Diagnostics {
	status, respBody, err := attempt(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	if status >= 200 && status < 300 {
		return nil
	}
	if status != http.StatusInternalServerError {
		return diagutil.ReportUnknownHTTPError(status, respBody)
	}

	lastStatus, lastBody := status, respBody
	checker := func(ctx context.Context) (bool, error) {
		status, respBody, err := attempt(ctx)
		if err != nil {
			return false, err
		}
		lastStatus, lastBody = status, respBody
		if status == http.StatusInternalServerError {
			return false, nil
		}
		if status < 200 || status >= 300 {
			return false, &fatalCreateError{status: status}
		}
		return true, nil
	}

	if waitErr := asyncutils.WaitForStateTransition(ctx, resourceType, resourceID, checker, asyncutils.WithPollInterval(pollInterval)); waitErr != nil {
		if lastStatus != 0 && (lastStatus < 200 || lastStatus >= 300) {
			return diagutil.ReportUnknownHTTPError(lastStatus, lastBody)
		}
		return diagutil.FrameworkDiagFromError(waitErr)
	}
	return nil
}

// fatalCreateError signals a non-retriable create failure so the retry loop
// stops immediately.
type fatalCreateError struct {
	status int
}

func (e *fatalCreateError) Error() string {
	return http.StatusText(e.status)
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
