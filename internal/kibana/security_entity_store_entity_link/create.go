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

package security_entity_store_entity_link

import (
	"context"
	"net/http"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// createRetryPollInterval is the cadence at which the entity-link create call
// is retried while the entity store is still initializing (HTTP 500). The
// overall retry budget is bounded by the Create ctx deadline (from the resource
// timeouts block), not by this interval.
const createRetryPollInterval = 5 * time.Second

func createEntityLink(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[entityLinkModel]) (entitycore.KibanaWriteResult[entityLinkModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	entityIDs := typeutils.SetTypeAs[string](ctx, plan.EntityIDs, path.Root("entity_ids"), &diags)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[entityLinkModel]{}, diags
	}

	body := kbapi.PostSecurityEntityStoreResolutionLinkJSONRequestBody{
		TargetId:  plan.TargetID.ValueString(),
		EntityIds: entityIDs,
	}

	attempt := func(ctx context.Context) (int, []byte, error) {
		resp, err := client.GetKibanaOapiClient().API.PostSecurityEntityStoreResolutionLinkWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(req.SpaceID))
		if err != nil {
			return 0, nil, err
		}
		return resp.StatusCode(), resp.Body, nil
	}

	if d := retryCreateOnServerError(ctx, "security entity store entity link", plan.TargetID.ValueString(), attempt, createRetryPollInterval); d.HasError() {
		return entitycore.KibanaWriteResult[entityLinkModel]{}, d
	}

	return entitycore.KibanaWriteResult[entityLinkModel]{Model: plan}, diags
}

// createAttemptFunc performs a single create call and reports the raw HTTP
// status code and body (or a transport error).
type createAttemptFunc func(ctx context.Context) (statusCode int, body []byte, err error)

// retryCreateOnServerError performs an initial synchronous create attempt and,
// when the entity store returns HTTP 500 (still initializing), retries at
// createRetryPollInterval until the create succeeds (HTTP 2xx) or the Create
// ctx deadline is exceeded. Any non-500 non-2xx response fails fast. The retry
// is bounded solely by the ctx deadline; no separate wall-clock budget is used.
func retryCreateOnServerError(ctx context.Context, resourceType, resourceID string, attempt createAttemptFunc, pollInterval time.Duration) diag.Diagnostics {
	// Immediate first attempt so the happy path incurs no poll-interval delay.
	status, respBody, err := attempt(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	if fatal := classifyCreateStatus(status, respBody); fatal != nil {
		return fatal.diagnostics
	}
	if status != http.StatusInternalServerError {
		// 2xx success.
		return nil
	}

	// Still 500: enter the retry loop, capturing the last-observed response so
	// a deadline expiry can describe the final HTTP 500.
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
		if fatal := classifyCreateStatus(status, respBody); fatal != nil {
			return false, fatal.err
		}
		return true, nil
	}

	if waitErr := asyncutils.WaitForStateTransition(ctx, resourceType, resourceID, checker, asyncutils.WithPollInterval(pollInterval)); waitErr != nil {
		// The last-observed response is a non-2xx (500 on deadline expiry, or a
		// fatal status that stopped the loop); surface its status and body.
		if lastStatus != 0 && (lastStatus < 200 || lastStatus >= 300) {
			return diagutil.ReportUnknownHTTPError(lastStatus, lastBody)
		}
		return diagutil.FrameworkDiagFromError(waitErr)
	}
	return nil
}

// createStatusResult carries a fatal create outcome out of the classifier.
type createStatusResult struct {
	diagnostics diag.Diagnostics
	err         error
}

// classifyCreateStatus returns a non-nil result for a fatal (non-2xx, non-500)
// response and nil for a 2xx or a retriable 500. The returned result carries
// both the diagnostics (for the initial attempt) and an error (to stop the
// wait loop) describing the same failure.
func classifyCreateStatus(status int, body []byte) *createStatusResult {
	if status >= 200 && status < 300 {
		return nil
	}
	if status == http.StatusInternalServerError {
		return nil
	}
	diags := diagutil.ReportUnknownHTTPError(status, body)
	return &createStatusResult{
		diagnostics: diags,
		err:         &fatalCreateError{status: status},
	}
}

// fatalCreateError signals a non-retriable create failure so the retry loop
// stops immediately.
type fatalCreateError struct {
	status int
}

func (e *fatalCreateError) Error() string {
	return http.StatusText(e.status)
}
