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

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// errGetterFailed is a sentinel returned internally by
// [WaitWithGetterDiagnostics]'s state checker to bail out of the poll loop
// while preserving the getter's original framework diagnostics in a
// closure-captured variable, distinguishing that case from a ctx-deadline or
// ctx-cancellation error surfaced by [WaitForStateTransition] itself.
var errGetterFailed = errors.New("asyncutils: getter failed")

// GetterFunc fetches the current value of a polled resource. It is called
// synchronously from the same goroutine as [WaitWithGetterDiagnostics] on
// every poll attempt, including the immediate first check.
type GetterFunc[T any] func(ctx context.Context) (T, diag.Diagnostics)

// TerminalFunc inspects a value returned by [GetterFunc] and reports whether
// the wait is over. Returning diagnostics with an error marks the wait as
// over due to a terminal error state (e.g. a job that reached a failed
// status) rather than a getter failure or timeout.
type TerminalFunc[T any] func(value T) (done bool, diags diag.Diagnostics)

// WaitErrorFunc builds the diagnostics to return when the poll loop itself
// errors — a ctx deadline or cancellation surfaced by
// [WaitForStateTransition] — as opposed to a getter failure (handled
// internally) or a terminal error state (handled by [TerminalFunc]). last is
// the most recently observed value, so callers can report it (e.g. the last
// known status) even though the wait did not succeed.
type WaitErrorFunc[T any] func(last T, err error) diag.Diagnostics

// WaitWithGetterDiagnostics polls a resource via [WaitForStateTransition],
// generalizing the "immediate check, stateChecker closure distinguishing
// ctx-cancellation vs getter-failure sentinel vs terminal-state diagnostics,
// then translate the wait error" skeleton shared by create/action wait loops
// across the provider (e.g. waiting for a CCR follower index to become
// active, or a connector sync job to complete).
//
// get is called on every poll attempt. If it returns diagnostics with an
// error, the checker distinguishes a ctx-cancellation (surfaced directly to
// the caller as onWaitError) from any other getter failure (surfaced as the
// getter's own diagnostics, bypassing onWaitError entirely). Otherwise
// isDone classifies the fetched value: done=true with no error diagnostics
// ends the wait successfully; done=true with error diagnostics ends the
// wait with those diagnostics (a terminal error state); done=false keeps
// polling. If the poll loop itself times out or is cancelled, onWaitError
// builds the final diagnostics from the last observed value and that error.
func WaitWithGetterDiagnostics[T any](
	ctx context.Context,
	resourceType, resourceID string,
	get GetterFunc[T],
	isDone TerminalFunc[T],
	onWaitError WaitErrorFunc[T],
	opts ...Option,
) (T, diag.Diagnostics) {
	var (
		last          T
		getDiags      diag.Diagnostics
		terminalDiags diag.Diagnostics
	)

	stateChecker := func(checkCtx context.Context) (bool, error) {
		value, diags := get(checkCtx)
		if diags.HasError() {
			if checkCtx.Err() != nil {
				return false, checkCtx.Err()
			}
			getDiags = diags
			return false, errGetterFailed
		}

		last = value
		done, diags := isDone(value)
		if diags.HasError() {
			terminalDiags = diags
			return true, nil
		}
		return done, nil
	}

	waitErr := WaitForStateTransition(ctx, resourceType, resourceID, stateChecker, opts...)

	switch {
	case terminalDiags.HasError():
		return last, terminalDiags
	case errors.Is(waitErr, errGetterFailed):
		return last, getDiags
	case waitErr != nil:
		return last, onWaitError(last, waitErr)
	default:
		return last, nil
	}
}
