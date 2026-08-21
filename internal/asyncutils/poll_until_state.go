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

package asyncutils

import "context"

// PollUntilState repeatedly fetches a resource's current state via fetch and
// waits until isDesired reports the desired state has been reached, the
// context is done, or fetch returns an error.
//
// It performs an immediate fetch+check before entering the poll loop so a
// resource that is already settled returns without incurring the minimum
// poll interval. The immediate check is skipped when ctx is already done, so
// that case is instead reported by the poll loop's context handling (and
// fetch, which may depend on a live context, is never invoked with an
// already-expired one).
//
// The most recently fetched non-nil value is returned alongside the error,
// even when the desired state was never reached (e.g. on timeout), so
// callers can fall back to the last-observed value.
func PollUntilState[T any](ctx context.Context, resourceType, resourceID string, fetch func(context.Context) (*T, error), isDesired func(*T) bool, opts ...Option) (*T, error) {
	var last *T
	check := func(ctx context.Context) (bool, error) {
		val, err := fetch(ctx)
		if err != nil {
			return false, err
		}
		if val != nil {
			last = val
		}
		return isDesired(val), nil
	}

	if ctx.Err() == nil {
		done, err := check(ctx)
		if err != nil || done {
			return last, err
		}
	}

	err := WaitForStateTransition(ctx, resourceType, resourceID, check, opts...)
	return last, err
}
