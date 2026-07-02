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

	// Still 500: enter the retry loop, capturing the last-observed response so a
	// deadline expiry (or a fatal status that stops the loop) can surface the
	// final status and body.
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
