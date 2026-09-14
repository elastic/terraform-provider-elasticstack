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

// Package statetransition contains a shared start/stop-with-wait control
// flow used by ML resources that transition a resource between two states
// (e.g. an ML job between "opened"/"closed", or a datafeed between
// "started"/"stopped").
package statetransition

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Params describes a single start/stop-with-wait transition for a resource
// whose state is represented by the comparable type S.
type Params[S comparable] struct {
	// ResourceType is a human-readable label used in log messages and the
	// invalid-state error, e.g. "ML job" or "ML datafeed".
	ResourceType string
	// ResourceID identifies the specific resource instance being
	// transitioned, e.g. the job ID or datafeed ID.
	ResourceID string

	CurrentState S
	DesiredState S

	// StartState and StopState are the two valid values DesiredState may
	// take. ValidStatesDescription describes them for the invalid-state
	// error message, e.g. "'opened' and 'closed'".
	StartState             S
	StopState              S
	ValidStatesDescription string

	// Start is invoked when DesiredState == StartState.
	Start func(ctx context.Context) diag.Diagnostics
	// Stop is invoked when DesiredState == StopState.
	Stop func(ctx context.Context) diag.Diagnostics
	// Wait blocks until the resource reaches DesiredState, returning
	// whether it settled into that state.
	Wait func(ctx context.Context) (bool, diag.Diagnostics)
}

// Perform runs the shared start/stop-with-wait control flow: it is a no-op
// if the resource is already in the desired state, otherwise it invokes the
// appropriate Start/Stop function for the desired state and waits for the
// resource to settle into it.
func Perform[S comparable](ctx context.Context, p Params[S]) (bool, diag.Diagnostics) {
	if p.CurrentState == p.DesiredState {
		tflog.Debug(ctx, fmt.Sprintf("%s %s is already in desired state %v", p.ResourceType, p.ResourceID, p.DesiredState))
		return true, nil
	}

	switch p.DesiredState {
	case p.StartState:
		if diags := p.Start(ctx); diags.HasError() {
			return false, diags
		}
	case p.StopState:
		if diags := p.Stop(ctx); diags.HasError() {
			return false, diags
		}
	default:
		return false, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid state",
				fmt.Sprintf("Invalid state %v. Valid states are %s", p.DesiredState, p.ValidStatesDescription),
			),
		}
	}

	inDesiredState, diags := p.Wait(ctx)
	if diags.HasError() {
		return false, diags
	}

	tflog.Info(ctx, fmt.Sprintf("%s %s successfully transitioned to state %v", p.ResourceType, p.ResourceID, p.DesiredState))
	return inDesiredState, nil
}
