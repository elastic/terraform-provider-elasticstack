package datafeed

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
)

// getDatafeedState returns the current state of a datafeed
func (r *datafeedResource) getDatafeedState(ctx context.Context, datafeedId string) (string, error) {
	statsResponse, diags := elasticsearch.GetDatafeedStats(ctx, r.client, datafeedId)
	if diags.HasError() {
		return "", fmt.Errorf("failed to get datafeed stats: %v", diags)
	}

	if statsResponse == nil {
		return "", fmt.Errorf("datafeed %s not found", datafeedId)
	}

	return statsResponse.State, nil
}

// waitForDatafeedState waits for a datafeed to reach the desired state
func (r *datafeedResource) waitForDatafeedState(ctx context.Context, datafeedId, desiredState string) error {
	stateChecker := func(ctx context.Context) (bool, error) {
		currentState, err := r.getDatafeedState(ctx, datafeedId)
		if err != nil {
			return false, err
		}
		return currentState == desiredState, nil
	}

	return asyncutils.WaitForStateTransition(ctx, "datafeed", datafeedId, stateChecker)
}
