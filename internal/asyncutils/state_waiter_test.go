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

func TestWaitForStateTransition_Success(t *testing.T) {
	tests := []struct {
		name              string
		resourceType      string
		resourceID        string
		stateSequence     []bool
		expectedCallCount int
	}{
		{
			name:              "immediate success",
			resourceType:      "test-resource",
			resourceID:        "test-id",
			stateSequence:     []bool{true},
			expectedCallCount: 1,
		},
		{
			name:              "transition after delay",
			resourceType:      "test-resource",
			resourceID:        "test-id",
			stateSequence:     []bool{false, false, true},
			expectedCallCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			stateChecker := func(_ context.Context) (bool, error) {
				if callCount >= len(tt.stateSequence) {
					t.Errorf("unexpected call count: %d", callCount)
					return false, errors.New("unexpected call")
				}
				state := tt.stateSequence[callCount]
				callCount++
				return state, nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
			defer cancel()

			err := WaitForStateTransition(ctx, tt.resourceType, tt.resourceID, stateChecker)
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if callCount != tt.expectedCallCount {
				t.Errorf("expected %d calls, got %d", tt.expectedCallCount, callCount)
			}
		})
	}
}

func TestWaitForStateTransition_ContextTimeout(t *testing.T) {
	stateChecker := func(_ context.Context) (bool, error) {
		return false, nil // Always return false to indicate not in desired state
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := WaitForStateTransition(ctx, "test-resource", "test-id", stateChecker)
	if err == nil {
		t.Error("expected context timeout error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context deadline exceeded error, got: %v", err)
	}
}

func TestWaitForStateTransition_CheckerError(t *testing.T) {
	callCount := 0
	stateChecker := func(_ context.Context) (bool, error) {
		callCount++
		return false, assert.AnError
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	err := WaitForStateTransition(ctx, "test-resource", "test-id", stateChecker)
	require.ErrorIs(t, err, assert.AnError)
	require.Equal(t, 1, callCount)
}
