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
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ConflictMaxAttempts is the default number of attempts for ConflictRetry.
const ConflictMaxAttempts = 5

// ConflictRetry retries fn (up to maxAttempts) when the Kibana API returns
// HTTP 409 Conflict. Kibana holds a short exclusive write lock during
// mutations; a brief exponential backoff with jitter resolves the contention
// without user-visible errors on concurrent applies/destroys.
func ConflictRetry[T any](ctx context.Context, maxAttempts int, fn func() (T, int, diag.Diagnostics)) (T, diag.Diagnostics) {
	backoff := 500 * time.Millisecond
	for attempt := 1; ; attempt++ {
		result, statusCode, diags := fn()
		if statusCode != http.StatusConflict || attempt >= maxAttempts {
			return result, diags
		}

		jitter := time.Duration(rand.Int64N(int64(backoff) / 2))
		wait := backoff + jitter

		tflog.Debug(ctx, fmt.Sprintf("HTTP 409 Conflict, retrying (attempt %d/%d, backoff %s)", attempt, maxAttempts, wait))

		select {
		case <-ctx.Done():
			diags.AddError("retry aborted", ctx.Err().Error())
			return result, diags
		case <-time.After(wait):
		}
		backoff *= 2
	}
}
