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

package output

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputResourceUpgradeState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		rawState      map[string]any
		expectedState map[string]any
		expectError   bool
		errorContains string
	}{
		{
			name: "successful upgrade - ssl list to object",
			rawState: map[string]any{
				"id":   "test-output",
				"name": "Test Output",
				"type": "elasticsearch",
				"ssl": []any{
					map[string]any{
						"certificate":             "cert-content",
						"key":                     "key-content",
						"certificate_authorities": []any{"ca1", "ca2"},
					},
				},
				"hosts": []any{"https://localhost:9200"},
			},
			expectedState: map[string]any{
				"id":   "test-output",
				"name": "Test Output",
				"type": "elasticsearch",
				"ssl": map[string]any{
					"certificate":             "cert-content",
					"key":                     "key-content",
					"certificate_authorities": []any{"ca1", "ca2"},
				},
				"hosts": []any{"https://localhost:9200"},
			},
			expectError: false,
		},
		{
			name: "no ssl field - no changes",
			rawState: map[string]any{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"hosts": []any{"https://localhost:9200"},
			},
			expectedState: map[string]any{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"hosts": []any{"https://localhost:9200"},
			},
			expectError: false,
		},
		{
			name: "empty ssl list - removes ssl field",
			rawState: map[string]any{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"ssl":   []any{},
				"hosts": []any{"https://localhost:9200"},
			},
			expectedState: map[string]any{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"hosts": []any{"https://localhost:9200"},
			},
			expectError: false,
		},
		{
			name: "ssl not an array - returns error",
			rawState: map[string]any{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"ssl":   "invalid-type",
				"hosts": []any{"https://localhost:9200"},
			},
			expectedState: nil,
			expectError:   true,
			errorContains: "Unexpected type for legacy ssl attribute",
		},
		{
			name: "multiple ssl items - takes first item",
			rawState: map[string]any{
				"id":   "test-output",
				"name": "Test Output",
				"type": "elasticsearch",
				"ssl": []any{
					map[string]any{"certificate": "cert1"},
					map[string]any{"certificate": "cert2"},
				},
				"hosts": []any{"https://localhost:9200"},
			},
			expectedState: map[string]any{
				"id":    "test-output",
				"name":  "Test Output",
				"type":  "elasticsearch",
				"ssl":   map[string]any{"certificate": "cert1"},
				"hosts": []any{"https://localhost:9200"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Marshal the raw state to JSON
			rawStateJSON, err := json.Marshal(tt.rawState)
			require.NoError(t, err)

			// Create the upgrade request
			req := resource.UpgradeStateRequest{
				RawState: &tfprotov6.RawState{
					JSON: rawStateJSON,
				},
			}

			// Create a response
			resp := &resource.UpgradeStateResponse{}

			// Create the resource and call UpgradeState
			r := &outputResource{}
			upgraders := r.UpgradeState(context.Background())
			upgrader := upgraders[0]
			upgrader.StateUpgrader(context.Background(), req, resp)

			if tt.expectError {
				require.True(t, resp.Diagnostics.HasError(), "Expected error but got none")
				if tt.errorContains != "" {
					var errorSummary strings.Builder
					for _, diag := range resp.Diagnostics.Errors() {
						errorSummary.WriteString(diag.Summary() + " " + diag.Detail())
					}
					assert.Contains(t, errorSummary.String(), tt.errorContains)
				}
				return
			}

			// Check no errors occurred
			require.False(t, resp.Diagnostics.HasError(), "Unexpected error: %v", resp.Diagnostics.Errors())

			// Check that a DynamicValue is always returned
			require.NotNil(t, resp.DynamicValue, "DynamicValue should always be returned")

			// Unmarshal the upgraded state to compare
			var actualState map[string]any
			err = json.Unmarshal(resp.DynamicValue.JSON, &actualState)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedState, actualState)
		})
	}
}
