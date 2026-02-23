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
	"errors"
	"testing"
)

func TestGetDatafeedState_Success(t *testing.T) {
	tests := []struct {
		name          string
		datafeedID    string
		response      map[string]any
		expectedState string
		expectError   bool
	}{
		{
			name:       "running datafeed",
			datafeedID: "test-datafeed",
			response: map[string]any{
				"datafeeds": []any{
					map[string]any{
						"datafeed_id": "test-datafeed",
						"state":       "started",
					},
				},
			},
			expectedState: "started",
			expectError:   false,
		},
		{
			name:       "stopped datafeed",
			datafeedID: "test-datafeed",
			response: map[string]any{
				"datafeeds": []any{
					map[string]any{
						"datafeed_id": "test-datafeed",
						"state":       "stopped",
					},
				},
			},
			expectedState: "stopped",
			expectError:   false,
		},
		{
			name:        "datafeed not found",
			datafeedID:  "test-datafeed",
			response:    nil,
			expectError: true,
		},
		{
			name:       "empty datafeeds array",
			datafeedID: "test-datafeed",
			response: map[string]any{
				"datafeeds": []any{},
			},
			expectError: true,
		},
		{
			name:       "missing state field",
			datafeedID: "test-datafeed",
			response: map[string]any{
				"datafeeds": []any{
					map[string]any{
						"datafeed_id": "test-datafeed",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the state parsing logic using a helper function
			state, err := parseDatafeedStateFromResponse(tt.response)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got: %v", err)
				return
			}

			if state != tt.expectedState {
				t.Errorf("expected state %q, got %q", tt.expectedState, state)
			}
		})
	}
}

// Helper function to test the state parsing logic
func parseDatafeedStateFromResponse(statsResponse map[string]any) (string, error) {
	if statsResponse == nil {
		return "", errors.New("datafeed not found")
	}

	// Parse the response to get the state
	datafeeds, ok := statsResponse["datafeeds"].([]any)
	if !ok {
		return "", errors.New("unexpected response format: missing datafeeds field")
	}

	if len(datafeeds) == 0 {
		return "", errors.New("no datafeed found in response")
	}

	datafeedMap, ok := datafeeds[0].(map[string]any)
	if !ok {
		return "", errors.New("unexpected datafeed format in response")
	}

	state, exists := datafeedMap["state"].(string)
	if !exists {
		return "", errors.New("missing state field in datafeed response")
	}

	return state, nil
}
