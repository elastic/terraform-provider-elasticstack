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

package datafeed

import (
	"context"
	"errors"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type State string

const (
	StateStopped  State = "stopped"
	StateStarted  State = "started"
	StateStarting State = "starting"
)

// GetDatafeedState returns the current state of a datafeed
func GetDatafeedState(ctx context.Context, client *clients.APIClient, datafeedID string) (*State, diag.Diagnostics) {
	statsResponse, diags := elasticsearch.GetDatafeedStats(ctx, client, datafeedID)
	if diags.HasError() {
		return nil, diags
	}

	if statsResponse == nil {
		return nil, nil
	}

	state := State(statsResponse.State)
	return &state, nil
}

var terminalDatafeedStates = map[State]struct{}{
	StateStopped: {},
	StateStarted: {},
}

var errDatafeedInUndesiredState = errors.New("datafeed stuck in undesired state")

// WaitForDatafeedState waits for a datafeed to reach the desired state
func WaitForDatafeedState(ctx context.Context, client *clients.APIClient, datafeedID string, desiredState State) (bool, diag.Diagnostics) {
	stateChecker := func(ctx context.Context) (bool, error) {
		currentState, diags := GetDatafeedState(ctx, client, datafeedID)
		if diags.HasError() {
			return false, diagutil.FwDiagsAsError(diags)
		}

		if currentState == nil {
			return false, fmt.Errorf("datafeed %s not found", datafeedID)
		}

		if *currentState == desiredState {
			return true, nil
		}

		_, isInTerminalState := terminalDatafeedStates[*currentState]
		if isInTerminalState {
			return false, fmt.Errorf("%w: datafeed is in state [%s] but desired state is [%s]", errDatafeedInUndesiredState, *currentState, desiredState)
		}

		return false, nil
	}

	err := asyncutils.WaitForStateTransition(ctx, "datafeed", datafeedID, stateChecker)
	if errors.Is(err, errDatafeedInUndesiredState) {
		return false, nil
	}

	return err == nil, diagutil.FrameworkDiagFromError(err)
}
