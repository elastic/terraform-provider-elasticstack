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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollWithBackoff_ImmediateSuccess(t *testing.T) {
	t.Parallel()

	calls := 0
	fn := func(_ context.Context, attempt int) (int, bool, error) {
		calls++
		return attempt, true, nil
	}

	result, err := PollWithBackoff(context.Background(), BackoffConfig{Initial: time.Second}, fn)
	require.NoError(t, err)
	require.Equal(t, 1, calls)
	require.Equal(t, 1, result)
}

func TestPollWithBackoff_RetriesUntilDone(t *testing.T) {
	t.Parallel()

	calls := 0
	fn := func(_ context.Context, attempt int) (int, bool, error) {
		calls++
		return attempt, attempt >= 3, nil
	}

	cfg := BackoffConfig{Initial: 5 * time.Millisecond, Max: 20 * time.Millisecond}
	result, err := PollWithBackoff(context.Background(), cfg, fn)
	require.NoError(t, err)
	require.Equal(t, 3, calls)
	require.Equal(t, 3, result)
}

func TestPollWithBackoff_StopsOnError(t *testing.T) {
	t.Parallel()

	calls := 0
	fn := func(_ context.Context, _ int) (int, bool, error) {
		calls++
		return 0, false, assert.AnError
	}

	cfg := BackoffConfig{Initial: 5 * time.Millisecond, Max: 20 * time.Millisecond}
	_, err := PollWithBackoff(context.Background(), cfg, fn)
	require.ErrorIs(t, err, assert.AnError)
	require.Equal(t, 1, calls)
}

func TestPollWithBackoff_MaxAttemptsBound(t *testing.T) {
	t.Parallel()

	calls := 0
	fn := func(_ context.Context, _ int) (int, bool, error) {
		calls++
		return calls, false, nil
	}

	cfg := BackoffConfig{Initial: 5 * time.Millisecond, MaxAttempts: 3}
	result, err := PollWithBackoff(context.Background(), cfg, fn)
	require.NoError(t, err)
	require.Equal(t, 3, calls)
	require.Equal(t, 3, result)
}

func TestPollWithBackoff_MaxElapsedBound(t *testing.T) {
	t.Parallel()

	calls := 0
	fn := func(_ context.Context, _ int) (int, bool, error) {
		calls++
		return calls, false, nil
	}

	cfg := BackoffConfig{Initial: 10 * time.Millisecond, Max: 10 * time.Millisecond, MaxElapsed: 35 * time.Millisecond}
	start := time.Now()
	_, err := PollWithBackoff(context.Background(), cfg, fn)
	require.NoError(t, err)
	require.Less(t, time.Since(start), time.Second)
	require.GreaterOrEqual(t, calls, 2)
}

func TestPollWithBackoff_ContextCancelledWhileWaiting(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	calls := 0
	fn := func(_ context.Context, _ int) (int, bool, error) {
		calls++
		return calls, false, nil
	}

	cfg := BackoffConfig{Initial: time.Second}
	_, err := PollWithBackoff(ctx, cfg, fn)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Equal(t, 1, calls)
}

func TestPollWithBackoff_FlatDelayWhenMaxUnset(t *testing.T) {
	t.Parallel()

	var gaps []time.Duration
	last := time.Now()
	fn := func(_ context.Context, attempt int) (int, bool, error) {
		now := time.Now()
		if attempt > 1 {
			gaps = append(gaps, now.Sub(last))
		}
		last = now
		return attempt, attempt >= 3, nil
	}

	cfg := BackoffConfig{Initial: 15 * time.Millisecond}
	_, err := PollWithBackoff(context.Background(), cfg, fn)
	require.NoError(t, err)
	require.Len(t, gaps, 2)
	for _, gap := range gaps {
		assert.Less(t, gap, 60*time.Millisecond, "delay should stay flat around Initial, not grow")
	}
}
