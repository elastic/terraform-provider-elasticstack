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

import (
	"context"
	"time"
)

// BackoffConfig bounds the retry policy used by [PollWithBackoff]. It
// generalizes the wall-clock-bound, attempt-count-bound, and flat-delay
// policies that call sites previously hand-rolled independently.
type BackoffConfig struct {
	// Initial is the delay before the second attempt. The first attempt
	// always runs immediately.
	Initial time.Duration
	// Max caps the backoff delay after it doubles on each retry. Leave zero
	// to keep the delay flat at Initial (no exponential growth).
	Max time.Duration
	// MaxAttempts bounds the number of attempts. Zero means unbounded (rely
	// on MaxElapsed and/or ctx cancellation to stop retrying).
	MaxAttempts int
	// MaxElapsed bounds the total wall-clock time spent retrying, measured
	// from the first attempt. Zero means unbounded (rely on MaxAttempts
	// and/or ctx cancellation to stop retrying).
	MaxElapsed time.Duration
}

// PollFunc performs one poll attempt, 1-indexed by attempt. It returns the
// current result, whether polling should stop (done), and any error. A
// non-nil error stops polling immediately, same as done=true.
type PollFunc[T any] func(ctx context.Context, attempt int) (result T, done bool, err error)

// PollWithBackoff calls fn immediately, then retries with exponential
// backoff (per cfg) until fn reports done, fn returns an error, ctx is
// cancelled, or a configured bound (MaxAttempts / MaxElapsed) is reached.
//
// On ctx cancellation while waiting between attempts, PollWithBackoff
// returns the last result along with ctx.Err(). When a bound is reached
// without fn ever reporting done, it returns the last result with a nil
// error — callers that need retries-exhausted to be an error should check
// the result themselves (e.g. a not-found sentinel), mirroring how the
// original per-call-site loops behaved.
func PollWithBackoff[T any](ctx context.Context, cfg BackoffConfig, fn PollFunc[T]) (T, error) {
	backoff := cfg.Initial
	start := time.Now()

	var result T
	for attempt := 1; ; attempt++ {
		var done bool
		var err error
		result, done, err = fn(ctx, attempt)
		if done || err != nil {
			return result, err
		}

		if cfg.MaxAttempts > 0 && attempt >= cfg.MaxAttempts {
			return result, nil
		}
		if cfg.MaxElapsed > 0 && time.Since(start) >= cfg.MaxElapsed {
			return result, nil
		}

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(backoff):
		}

		if cfg.Max > 0 {
			backoff *= 2
			if backoff > cfg.Max {
				backoff = cfg.Max
			}
		}
	}
}
