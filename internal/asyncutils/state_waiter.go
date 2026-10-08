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
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// StateChecker is a function that checks if a resource is in the desired state.
// It should return true if the resource is in the desired state, false otherwise, and any error that occurred during the check.
type StateChecker func(ctx context.Context) (isDesiredState bool, err error)

// defaultPollInterval is the polling cadence used when callers do not pass
// [WithPollInterval]. Two seconds keeps short transitions snappy without
// hammering the API.
const defaultPollInterval = 2 * time.Second

// Option customizes [WaitForStateTransition] behavior.
type Option func(*waitConfig)

type waitConfig struct {
	pollInterval time.Duration
}

// WithPollInterval overrides the default polling interval. Useful for
// long-running resources (e.g. connector sync jobs) where a slower cadence
// is more appropriate than the default snappy 2-second tick.
func WithPollInterval(d time.Duration) Option {
	return func(c *waitConfig) {
		if d > 0 {
			c.pollInterval = d
		}
	}
}

// WaitForStateTransition waits for a resource to reach the desired state by polling its current state.
// If ctx is already done, it returns ctx.Err() without calling the checker.
// Otherwise the first check runs immediately so transitions that complete in
// under one poll interval (for example a lookback-only ML datafeed that starts
// and stops in well under two seconds) are observed. Subsequent checks use the
// default poll interval of two seconds; pass [WithPollInterval] to customize it.
//
// It is a thin wrapper over [PollWithBackoff] with a flat delay (no
// exponential growth, no jitter, no attempt/elapsed bound): the loop only
// stops on success, checker error, or ctx cancellation.
func WaitForStateTransition(ctx context.Context, resourceType, resourceID string, stateChecker StateChecker, opts ...Option) error {
	cfg := waitConfig{pollInterval: defaultPollInterval}
	for _, opt := range opts {
		opt(&cfg)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	fn := func(ctx context.Context, attempt int) (struct{}, bool, error) {
		isInDesiredState, err := stateChecker(ctx)
		if err != nil {
			return struct{}{}, true, fmt.Errorf("failed to check state during wait: %w", err)
		}
		if isInDesiredState {
			return struct{}{}, true, nil
		}

		// The first check runs immediately with no prior wait, so only log on
		// the later, ticker-driven checks to match the original behavior.
		if attempt > 1 {
			tflog.Debug(ctx, fmt.Sprintf("Waiting for %s %s to reach desired state...", resourceType, resourceID))
		}
		return struct{}{}, false, nil
	}

	_, err := PollWithBackoff(ctx, BackoffConfig{Initial: cfg.pollInterval}, fn)
	return err
}

// ErrTerminalState is wrapped into the error [WaitForTerminalOrDesiredState]
// returns when the resource settles into a terminal state other than the
// desired one. Callers can check for it with [errors.Is] to distinguish a
// fast-fail terminal mismatch from a context cancellation or a state-lookup
// failure.
var ErrTerminalState = errors.New("resource settled into a terminal state other than the desired one")

// WaitForTerminalOrDesiredState polls getState, via [WaitForStateTransition],
// until it reports desiredState, a state in terminalStates other than
// desiredState, or ctx is done.
//
// Resources such as ML jobs and datafeeds only transition between states on
// their own up to a point: once they land in a terminal state (e.g. "opened",
// "closed", "failed" for a job) they stay there until something external acts
// on them again. Polling past that point would just wait out the context
// deadline instead of surfacing the mismatch, so callers that know a
// resource's terminal states can pass them here to fail fast with an error
// wrapping [ErrTerminalState].
func WaitForTerminalOrDesiredState[T comparable](
	ctx context.Context,
	resourceType, resourceID string,
	desiredState T,
	terminalStates map[T]struct{},
	getState func(ctx context.Context) (*T, error),
	opts ...Option,
) error {
	stateChecker := func(ctx context.Context) (bool, error) {
		currentState, err := getState(ctx)
		if err != nil {
			return false, err
		}

		if currentState == nil {
			return false, fmt.Errorf("%s %s not found", resourceType, resourceID)
		}

		if *currentState == desiredState {
			return true, nil
		}

		if _, isTerminal := terminalStates[*currentState]; isTerminal {
			return false, fmt.Errorf("%w: %s %s is in state [%v] but desired state is [%v]", ErrTerminalState, resourceType, resourceID, *currentState, desiredState)
		}

		return false, nil
	}

	return WaitForStateTransition(ctx, resourceType, resourceID, stateChecker, opts...)
}
