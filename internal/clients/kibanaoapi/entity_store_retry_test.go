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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testPollInterval = 10 * time.Millisecond

func TestRetryCreateOnServerError_SucceedsImmediately(t *testing.T) {
	t.Parallel()

	calls := 0
	attempt := func(_ context.Context) (int, []byte, error) {
		calls++
		return http.StatusOK, nil, nil
	}

	diags := RetryCreateOnServerError(context.Background(), "test", "id", attempt, testPollInterval)
	require.False(t, diags.HasError())
	require.Equal(t, 1, calls, "happy path should not enter the retry loop")
}

func TestRetryCreateOnServerError_RetriesThenSucceeds(t *testing.T) {
	t.Parallel()

	calls := 0
	attempt := func(_ context.Context) (int, []byte, error) {
		calls++
		if calls < 3 {
			return http.StatusInternalServerError, []byte("still installing"), nil
		}
		return http.StatusOK, nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	diags := RetryCreateOnServerError(ctx, "test", "id", attempt, testPollInterval)
	require.False(t, diags.HasError())
	require.Equal(t, 3, calls)
}

func TestRetryCreateOnServerError_FailFastOnNon500(t *testing.T) {
	t.Parallel()

	calls := 0
	attempt := func(_ context.Context) (int, []byte, error) {
		calls++
		return http.StatusBadRequest, []byte(`{"error":"bad request"}`), nil
	}

	diags := RetryCreateOnServerError(context.Background(), "test", "id", attempt, testPollInterval)
	require.True(t, diags.HasError())
	require.Equal(t, 1, calls, "non-500 non-2xx must fail fast without retrying")
}

func TestRetryCreateOnServerError_FailFastOnNon500AfterInitial500(t *testing.T) {
	t.Parallel()

	calls := 0
	attempt := func(_ context.Context) (int, []byte, error) {
		calls++
		if calls == 1 {
			return http.StatusInternalServerError, []byte("still installing"), nil
		}
		return http.StatusForbidden, []byte(`{"error":"forbidden"}`), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	diags := RetryCreateOnServerError(ctx, "test", "id", attempt, testPollInterval)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "forbidden", "fatal status body during retry should be surfaced")
	require.Equal(t, 2, calls)
}

func TestRetryCreateOnServerError_DeadlineWhileStill500(t *testing.T) {
	t.Parallel()

	attempt := func(_ context.Context) (int, []byte, error) {
		return http.StatusInternalServerError, []byte("still installing"), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	diags := RetryCreateOnServerError(ctx, "test", "id", attempt, testPollInterval)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "still installing", "final 500 body should be surfaced")
}

func TestRetryCreateOnServerError_TransportErrorFailsFast(t *testing.T) {
	t.Parallel()

	calls := 0
	attempt := func(_ context.Context) (int, []byte, error) {
		calls++
		return 0, nil, context.Canceled
	}

	diags := RetryCreateOnServerError(context.Background(), "test", "id", attempt, testPollInterval)
	require.True(t, diags.HasError())
	require.Equal(t, 1, calls)
}
