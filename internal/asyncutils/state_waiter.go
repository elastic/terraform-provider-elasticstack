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
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// StateChecker is a function that checks if a resource is in the desired state.
// It should return true if the resource is in the desired state, false otherwise, and any error that occurred during the check.
type StateChecker func(ctx context.Context) (isDesiredState bool, err error)

// WaitForStateTransition waits for a resource to reach the desired state by polling its current state.
// It uses exponential backoff with a maximum interval to avoid overwhelming the API.
func WaitForStateTransition(ctx context.Context, resourceType, resourceID string, stateChecker StateChecker) error {
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

			tflog.Debug(ctx, fmt.Sprintf("Waiting for %s %s to reach desired state...", resourceType, resourceID))
		}
	}
}
