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

package ml

import (
	"context"
	"errors"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// ErrResourceInUndesiredTerminalState is wrapped into the error returned by
// [WaitForResourceState] when the resource settles into a Terminal state
// (per [WaitForResourceStateConfig.Terminal]) other than the desired one, so
// continued polling would never succeed.
var ErrResourceInUndesiredTerminalState = errors.New("resource in undesired terminal state")

// WaitForResourceStateConfig configures [WaitForResourceState].
type WaitForResourceStateConfig[T comparable] struct {
	// Get fetches the resource's current state. A nil state means the
	// resource was not found.
	Get func(ctx context.Context) (*T, fwdiags.Diagnostics)
	// Desired is the state being waited for.
	Desired T
	// NotFoundIsDesired treats a nil state from Get as having reached the
	// desired state (e.g. a deleted job counts as "closed") instead of as an
	// error.
	NotFoundIsDesired bool
	// Terminal, if set, lists states the resource cannot transition out of.
	// When the current state is terminal but not Desired, polling stops
	// immediately instead of waiting out the full context deadline.
	Terminal map[T]struct{}
}

// WaitForResourceState polls cfg.Get until the resource reaches the desired
// state, centralizing the get-state/wait-for-state pattern shared by the ML
// sub-resource packages (datafeeds, jobs, anomaly detection jobs). An
// immediate check runs before entering the poll loop so a resource already
// in the desired state settles without incurring the minimum poll interval.
//
// It returns reached=true once the desired state is observed (including via
// NotFoundIsDesired). It returns reached=false with no diagnostics when
// polling stops early because the resource is stuck in an undesired Terminal
// state; any other failure (a Get error, or the resource never being found
// when NotFoundIsDesired is false) is returned as diagnostics.
func WaitForResourceState[T comparable](ctx context.Context, resourceType, resourceID string, cfg WaitForResourceStateConfig[T]) (reached bool, diags fwdiags.Diagnostics) {
	stateChecker := func(ctx context.Context) (bool, error) {
		currentState, diags := cfg.Get(ctx)
		if diags.HasError() {
			return false, diagutil.FwDiagsAsError(diags)
		}

		if currentState == nil {
			if cfg.NotFoundIsDesired {
				return true, nil
			}
			return false, fmt.Errorf("%s %s not found", resourceType, resourceID)
		}

		if *currentState == cfg.Desired {
			return true, nil
		}

		if _, isTerminal := cfg.Terminal[*currentState]; isTerminal {
			return false, fmt.Errorf("%w: %s %s is in state [%v] but desired state is [%v]", ErrResourceInUndesiredTerminalState, resourceType, resourceID, *currentState, cfg.Desired)
		}

		return false, nil
	}

	if done, err := stateChecker(ctx); err != nil || done {
		if errors.Is(err, ErrResourceInUndesiredTerminalState) {
			return false, nil
		}
		return done, diagutil.FrameworkDiagFromError(err)
	}

	err := asyncutils.WaitForStateTransition(ctx, resourceType, resourceID, stateChecker)
	if errors.Is(err, ErrResourceInUndesiredTerminalState) {
		return false, nil
	}

	return err == nil, diagutil.FrameworkDiagFromError(err)
}
