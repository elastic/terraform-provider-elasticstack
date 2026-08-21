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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollUntilState_ImmediateSuccess(t *testing.T) {
	t.Parallel()

	callCount := 0
	fetch := func(_ context.Context) (*string, error) {
		callCount++
		state := "running"
		return &state, nil
	}
	isDesired := func(s *string) bool { return s != nil && *s == "running" }

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := PollUntilState(ctx, "test-resource", "test-id", fetch, isDesired)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "running", *result)
	assert.Equal(t, 1, callCount, "immediate check should avoid waiting for a poll tick")
}

func TestPollUntilState_TransitionAfterDelay(t *testing.T) {
	t.Parallel()

	states := []string{"starting", "starting", "running"}
	callCount := 0
	fetch := func(_ context.Context) (*string, error) {
		state := states[callCount]
		callCount++
		return &state, nil
	}
	isDesired := func(s *string) bool { return s != nil && *s == "running" }

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := PollUntilState(ctx, "test-resource", "test-id", fetch, isDesired, WithPollInterval(10*time.Millisecond))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "running", *result)
	assert.Equal(t, 3, callCount)
}

func TestPollUntilState_FetchError(t *testing.T) {
	t.Parallel()

	giveErr := errors.New("boom")
	fetch := func(_ context.Context) (*string, error) { return nil, giveErr }
	isDesired := func(_ *string) bool { return false }

	result, err := PollUntilState(context.Background(), "test-resource", "test-id", fetch, isDesired)
	require.ErrorIs(t, err, giveErr)
	assert.Nil(t, result)
}

func TestPollUntilState_ReturnsLastFetchedValueOnTimeout(t *testing.T) {
	t.Parallel()

	fetch := func(_ context.Context) (*string, error) {
		state := "starting"
		return &state, nil
	}
	isDesired := func(_ *string) bool { return false }

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	result, err := PollUntilState(ctx, "test-resource", "test-id", fetch, isDesired, WithPollInterval(5*time.Millisecond))
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.NotNil(t, result, "the last successfully fetched value should still be returned on timeout")
	assert.Equal(t, "starting", *result)
}

func TestPollUntilState_SkipsImmediateCheckWhenContextAlreadyDone(t *testing.T) {
	t.Parallel()

	called := false
	fetch := func(_ context.Context) (*string, error) {
		called = true
		state := "running"
		return &state, nil
	}
	isDesired := func(s *string) bool { return s != nil && *s == "running" }

	ctx, cancel := context.WithTimeout(context.Background(), -time.Millisecond)
	defer cancel()

	result, err := PollUntilState(ctx, "test-resource", "test-id", fetch, isDesired)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Nil(t, result)
	assert.False(t, called, "fetch must not be invoked when ctx is already done, so callers relying on a live context are not called with an expired one")
}

func TestPollUntilState_NilValueIsNotDesiredByDefault(t *testing.T) {
	t.Parallel()

	callCount := 0
	fetch := func(_ context.Context) (*string, error) {
		callCount++
		if callCount < 2 {
			return nil, nil
		}
		state := "closed"
		return &state, nil
	}
	isDesired := func(s *string) bool { return s == nil || *s == "closed" }

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := PollUntilState(ctx, "test-resource", "test-id", fetch, isDesired)
	require.NoError(t, err)
	assert.Nil(t, result, "a nil fetch result that satisfies isDesired should not be tracked as the last value")
}
