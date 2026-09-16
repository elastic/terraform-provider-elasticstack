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

package kibanautil

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ConflictMaxAttempts is the default number of attempts for ConflictRetry.
const ConflictMaxAttempts = 5

// conflictAttemptResult carries fn's result and diagnostics through
// asyncutils.PollWithBackoff, which only threads a single T plus a plain
// error between attempts.
type conflictAttemptResult[T any] struct {
	value T
	diags diag.Diagnostics
}

// ConflictRetry retries fn (up to maxAttempts) when the Kibana API returns
// HTTP 409 Conflict. Kibana holds a short exclusive write lock during
// mutations; a brief exponential backoff with jitter resolves the contention
// without user-visible errors on concurrent applies/destroys.
func ConflictRetry[T any](ctx context.Context, maxAttempts int, fn func() (T, int, diag.Diagnostics)) (T, diag.Diagnostics) {
	cfg := asyncutils.BackoffConfig{
		Initial:     500 * time.Millisecond,
		Max:         time.Hour,
		MaxAttempts: maxAttempts,
		Jitter:      0.5,
	}

	attempt, err := asyncutils.PollWithBackoff(ctx, cfg, func(ctx context.Context, attemptNum int) (conflictAttemptResult[T], bool, error) {
		value, statusCode, diags := fn()
		done := statusCode != http.StatusConflict
		if !done {
			tflog.Debug(ctx, fmt.Sprintf("HTTP 409 Conflict (attempt %d/%d)", attemptNum, maxAttempts))
		}
		return conflictAttemptResult[T]{value: value, diags: diags}, done, nil
	})

	diags := attempt.diags
	if err != nil {
		diags.AddError("retry aborted", err.Error())
	}
	return attempt.value, diags
}
