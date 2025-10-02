package datafeed

import (
	"errors"
	"testing"
)

func TestGetDatafeedState_Success(t *testing.T) {
	tests := []struct {
		name          string
		datafeedId    string
		response      map[string]interface{}
		expectedState string
		expectError   bool
	}{
		{
			name:       "running datafeed",
			datafeedId: "test-datafeed",
			response: map[string]interface{}{
				"datafeeds": []interface{}{
					map[string]interface{}{
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
			datafeedId: "test-datafeed",
			response: map[string]interface{}{
				"datafeeds": []interface{}{
					map[string]interface{}{
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
			datafeedId:  "test-datafeed",
			response:    nil,
			expectError: true,
		},
		{
			name:       "empty datafeeds array",
			datafeedId: "test-datafeed",
			response: map[string]interface{}{
				"datafeeds": []interface{}{},
			},
			expectError: true,
		},
		{
			name:       "missing state field",
			datafeedId: "test-datafeed",
			response: map[string]interface{}{
				"datafeeds": []interface{}{
					map[string]interface{}{
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
func parseDatafeedStateFromResponse(statsResponse map[string]interface{}) (string, error) {
	if statsResponse == nil {
		return "", errors.New("datafeed not found")
	}

	// Parse the response to get the state
	datafeeds, ok := statsResponse["datafeeds"].([]interface{})
	if !ok {
		return "", errors.New("unexpected response format: missing datafeeds field")
	}

	if len(datafeeds) == 0 {
		return "", errors.New("no datafeed found in response")
	}

	datafeedMap, ok := datafeeds[0].(map[string]interface{})
	if !ok {
		return "", errors.New("unexpected datafeed format in response")
	}

	state, exists := datafeedMap["state"].(string)
	if !exists {
		return "", errors.New("missing state field in datafeed response")
	}

	return state, nil
}
