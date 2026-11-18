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

package statetransition

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noopDiags(context.Context) diag.Diagnostics { return nil }

func TestPerform_alreadyInDesiredState(t *testing.T) {
	startCalled := false
	stopCalled := false
	waitCalled := false

	inState, diags := Perform(context.Background(), Params[string]{
		ResourceType:           "ML job",
		ResourceID:             "job-1",
		CurrentState:           "opened",
		DesiredState:           "opened",
		StartState:             "opened",
		StopState:              "closed",
		ValidStatesDescription: "'opened' and 'closed'",
		Start: func(_ context.Context) diag.Diagnostics {
			startCalled = true
			return nil
		},
		Stop: func(_ context.Context) diag.Diagnostics {
			stopCalled = true
			return nil
		},
		Wait: func(_ context.Context) (bool, diag.Diagnostics) {
			waitCalled = true
			return true, nil
		},
	})

	require.False(t, diags.HasError())
	assert.True(t, inState)
	assert.False(t, startCalled, "start should not be called when already in the desired state")
	assert.False(t, stopCalled, "stop should not be called when already in the desired state")
	assert.False(t, waitCalled, "wait should not be called when already in the desired state")
}

func TestPerform_startTransition(t *testing.T) {
	startCalled := false
	stopCalled := false

	inState, diags := Perform(context.Background(), Params[string]{
		ResourceType:           "ML job",
		ResourceID:             "job-1",
		CurrentState:           "closed",
		DesiredState:           "opened",
		StartState:             "opened",
		StopState:              "closed",
		ValidStatesDescription: "'opened' and 'closed'",
		Start: func(_ context.Context) diag.Diagnostics {
			startCalled = true
			return nil
		},
		Stop: func(_ context.Context) diag.Diagnostics {
			stopCalled = true
			return nil
		},
		Wait: func(_ context.Context) (bool, diag.Diagnostics) {
			return true, nil
		},
	})

	require.False(t, diags.HasError())
	assert.True(t, inState)
	assert.True(t, startCalled)
	assert.False(t, stopCalled)
}

func TestPerform_stopTransition(t *testing.T) {
	startCalled := false
	stopCalled := false

	inState, diags := Perform(context.Background(), Params[string]{
		ResourceType:           "ML job",
		ResourceID:             "job-1",
		CurrentState:           "opened",
		DesiredState:           "closed",
		StartState:             "opened",
		StopState:              "closed",
		ValidStatesDescription: "'opened' and 'closed'",
		Start: func(_ context.Context) diag.Diagnostics {
			startCalled = true
			return nil
		},
		Stop: func(_ context.Context) diag.Diagnostics {
			stopCalled = true
			return nil
		},
		Wait: func(_ context.Context) (bool, diag.Diagnostics) {
			return true, nil
		},
	})

	require.False(t, diags.HasError())
	assert.True(t, inState)
	assert.False(t, startCalled)
	assert.True(t, stopCalled)
}

func TestPerform_invalidDesiredState(t *testing.T) {
	inState, diags := Perform(context.Background(), Params[string]{
		ResourceType:           "ML job",
		ResourceID:             "job-1",
		CurrentState:           "opened",
		DesiredState:           "bogus",
		StartState:             "opened",
		StopState:              "closed",
		ValidStatesDescription: "'opened' and 'closed'",
		Start:                  noopDiags,
		Stop:                   noopDiags,
		Wait: func(_ context.Context) (bool, diag.Diagnostics) {
			return true, nil
		},
	})

	require.True(t, diags.HasError())
	assert.False(t, inState)
	assert.Contains(t, diags[0].Summary(), "Invalid state")
	assert.Contains(t, diags[0].Detail(), "'opened' and 'closed'")
}

func TestPerform_startError(t *testing.T) {
	waitCalled := false

	inState, diags := Perform(context.Background(), Params[string]{
		ResourceType:           "ML job",
		ResourceID:             "job-1",
		CurrentState:           "closed",
		DesiredState:           "opened",
		StartState:             "opened",
		StopState:              "closed",
		ValidStatesDescription: "'opened' and 'closed'",
		Start: func(_ context.Context) diag.Diagnostics {
			return diag.Diagnostics{diag.NewErrorDiagnostic("boom", "start failed")}
		},
		Stop: noopDiags,
		Wait: func(_ context.Context) (bool, diag.Diagnostics) {
			waitCalled = true
			return true, nil
		},
	})

	require.True(t, diags.HasError())
	assert.False(t, inState)
	assert.False(t, waitCalled, "wait should not be called if start fails")
}

func TestPerform_waitError(t *testing.T) {
	inState, diags := Perform(context.Background(), Params[string]{
		ResourceType:           "ML job",
		ResourceID:             "job-1",
		CurrentState:           "closed",
		DesiredState:           "opened",
		StartState:             "opened",
		StopState:              "closed",
		ValidStatesDescription: "'opened' and 'closed'",
		Start:                  noopDiags,
		Stop:                   noopDiags,
		Wait: func(_ context.Context) (bool, diag.Diagnostics) {
			return false, diag.Diagnostics{diag.NewErrorDiagnostic("boom", "wait failed")}
		},
	})

	require.True(t, diags.HasError())
	assert.False(t, inState)
}

func TestPerform_waitSettlesFalseWithoutError(t *testing.T) {
	inState, diags := Perform(context.Background(), Params[string]{
		ResourceType:           "ML datafeed",
		ResourceID:             "datafeed-1",
		CurrentState:           "stopped",
		DesiredState:           "started",
		StartState:             "started",
		StopState:              "stopped",
		ValidStatesDescription: "'started' and 'stopped'",
		Start:                  noopDiags,
		Stop:                   noopDiags,
		Wait: func(_ context.Context) (bool, diag.Diagnostics) {
			return false, nil
		},
	})

	require.False(t, diags.HasError())
	assert.False(t, inState)
}

// customState exercises Perform with a non-string comparable type, mirroring
// how datafeed.State (a defined string type) is used by the datafeed_state
// package.
type customState string

func TestPerform_customComparableType(t *testing.T) {
	const (
		stateA customState = "a"
		stateB customState = "b"
	)

	inState, diags := Perform(context.Background(), Params[customState]{
		ResourceType:           "custom",
		ResourceID:             "id-1",
		CurrentState:           stateA,
		DesiredState:           stateB,
		StartState:             stateB,
		StopState:              stateA,
		ValidStatesDescription: "'a' and 'b'",
		Start: func(_ context.Context) diag.Diagnostics {
			return nil
		},
		Stop: noopDiags,
		Wait: func(_ context.Context) (bool, diag.Diagnostics) {
			return true, nil
		},
	})

	require.False(t, diags.HasError())
	assert.True(t, inState)
}
