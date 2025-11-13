package asyncutils

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// StateChecker is a function that checks if a resource is in the desired state.
// It should return true if the resource is in the desired state, false otherwise, and any error that occurred during the check.
type StateChecker func(ctx context.Context) (isDesiredState bool, err error)

// WaitForStateTransition waits for a resource to reach the desired state by polling its current state.
// It uses exponential backoff with a maximum interval to avoid overwhelming the API.
func WaitForStateTransition(ctx context.Context, resourceType, resourceId string, stateChecker StateChecker) error {
	const pollInterval = 2 * time.Second
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			isInDesiredState, err := stateChecker(ctx)
			if err != nil {
				return fmt.Errorf("failed to check state during wait: %w", err)
			}
			if isInDesiredState {
				return nil
			}

			tflog.Debug(ctx, fmt.Sprintf("Waiting for %s %s to reach desired state...", resourceType, resourceId))
		}
	}
}
