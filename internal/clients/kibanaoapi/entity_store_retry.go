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

package kibanaoapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateAttemptFunc performs a single create call against the entity store API
// and reports the raw HTTP status code and body, or a transport error.
type CreateAttemptFunc func(ctx context.Context) (statusCode int, body []byte, err error)

// RetryCreateOnServerError performs an initial synchronous create attempt and,
// when the entity store returns HTTP 500 (still initializing), retries at
// pollInterval until the create succeeds (HTTP 2xx) or the ctx deadline is
// exceeded. Any non-500 non-2xx response fails fast. The retry is bounded
// solely by the ctx deadline (derived from the resource's Create timeouts
// block); pollInterval only controls cadence.
//
// It is shared by the entity and entity-link create paths, which have identical
// retry semantics, reusing the asyncutils.WaitForStateTransition primitive
// rather than introducing a separate retry package.
func RetryCreateOnServerError(ctx context.Context, resourceType, resourceID string, attempt CreateAttemptFunc, pollInterval time.Duration) diag.Diagnostics {
	// Immediate first attempt so the happy path incurs no poll-interval delay.
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

	// Still 500: enter the retry loop, capturing the last-observed 500 response so
	// a deadline expiry can describe the final HTTP 500 (REQ-WAIT-003).
	lastStatus, lastBody := status, respBody
	checker := func(ctx context.Context) (bool, error) {
		status, respBody, err := attempt(ctx)
		if err != nil {
			// Transport error: surface it directly rather than masking it with
			// the last HTTP 500.
			return false, err
		}
		if status == http.StatusInternalServerError {
			lastStatus, lastBody = status, respBody
			return false, nil
		}
		if status < 200 || status >= 300 {
			return false, &fatalCreateError{status: status, body: respBody}
		}
		return true, nil
	}

	waitErr := asyncutils.WaitForStateTransition(ctx, resourceType, resourceID, checker, asyncutils.WithPollInterval(pollInterval))
	if waitErr == nil {
		return nil
	}

	// A non-500 non-2xx response stopped the loop: report that HTTP failure.
	if fatal, ok := errors.AsType[*fatalCreateError](waitErr); ok {
		return diagutil.ReportUnknownHTTPError(fatal.status, fatal.body)
	}

	// The Create deadline bounded the retries while the store kept returning
	// HTTP 500: describe the final 500 (REQ-WAIT-003).
	if errors.Is(waitErr, context.DeadlineExceeded) {
		return diagutil.ReportUnknownHTTPError(lastStatus, lastBody)
	}

	// Context cancellation or a transport error encountered during retries:
	// surface it directly rather than reporting a stale HTTP 500.
	return diagutil.FrameworkDiagFromError(waitErr)
}

// fatalCreateError signals a non-retriable create failure so the retry loop
// stops immediately, carrying the HTTP status and response body so the caller
// can build an accurate diagnostic.
type fatalCreateError struct {
	status int
	body   []byte
}

func (e *fatalCreateError) Error() string {
	return http.StatusText(e.status)
}
