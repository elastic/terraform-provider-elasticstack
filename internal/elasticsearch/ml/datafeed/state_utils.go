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

// GetDatafeedState returns the current state of a datafeed
func GetDatafeedState(ctx context.Context, client *clients.ApiClient, datafeedId string) (*string, diag.Diagnostics) {
	statsResponse, diags := elasticsearch.GetDatafeedStats(ctx, client, datafeedId)
	if diags.HasError() {
		return nil, diags
	}

	if statsResponse == nil {
		return nil, nil
	}

	return &statsResponse.State, nil
}

var terminalDatafeedStates = map[string]struct{}{
	"stopped": {},
	"started": {},
}

var errDatafeedInUndesiredState = errors.New("datafeed stuck in undesired state")

// WaitForDatafeedState waits for a datafeed to reach the desired state
func WaitForDatafeedState(ctx context.Context, client *clients.ApiClient, datafeedId, desiredState string) (bool, diag.Diagnostics) {
	stateChecker := func(ctx context.Context) (bool, error) {
		currentState, diags := GetDatafeedState(ctx, client, datafeedId)
		if diags.HasError() {
			return false, diagutil.FwDiagsAsError(diags)
		}

		if currentState == nil {
			return false, fmt.Errorf("datafeed %s not found", datafeedId)
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

	err := asyncutils.WaitForStateTransition(ctx, "datafeed", datafeedId, stateChecker)
	if errors.Is(err, errDatafeedInUndesiredState) {
		return false, nil
	}

	return err == nil, diagutil.FrameworkDiagFromError(err)
}
