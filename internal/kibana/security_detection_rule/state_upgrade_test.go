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

package securitydetectionrule

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func TestMigrateAlertsFilterV1ToV2(t *testing.T) {
	t.Parallel()

	r := &securityDetectionRuleResource{}
	upgraders := r.UpgradeState(context.Background())
	up, ok := upgraders[1]
	require.True(t, ok)

	tests := []struct {
		name    string
		raw     map[string]any
		wantErr bool
	}{
		{
			name: "no actions key",
			raw: map[string]any{
				"name": "rule",
			},
		},
		{
			name: "empty actions list",
			raw: map[string]any{
				"actions": []any{},
			},
		},
		{
			name: "action without alerts_filter",
			raw: map[string]any{
				"actions": []any{
					map[string]any{
						"action_type_id": ".slack",
						"id":             "connector-1",
						"params":         `{"message":"hi"}`,
					},
				},
			},
		},
		{
			name: "action with alerts_filter map removed",
			raw: map[string]any{
				"actions": []any{
					map[string]any{
						"action_type_id": ".slack",
						"id":             "connector-1",
						"params":         `{"message":"hi"}`,
						"alerts_filter": map[string]any{
							"query": "broken",
						},
					},
				},
			},
		},
		{
			name:    "malformed JSON",
			raw:     nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var rawJSON []byte
			var err error
			if tc.name == "malformed JSON" {
				rawJSON = []byte("{")
			} else {
				rawJSON, err = json.Marshal(tc.raw)
				require.NoError(t, err)
			}

			req := resource.UpgradeStateRequest{
				RawState: &tfprotov6.RawState{JSON: rawJSON},
			}
			resp := &resource.UpgradeStateResponse{}
			up.StateUpgrader(context.Background(), req, resp)

			if tc.wantErr {
				require.True(t, resp.Diagnostics.HasError())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)

			var got map[string]any
			require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

			actions, ok := got["actions"].([]any)
			if !ok || len(actions) == 0 {
				return
			}
			action, ok := actions[0].(map[string]any)
			require.True(t, ok)
			_, hasFilter := action["alerts_filter"]
			require.False(t, hasFilter, "alerts_filter should always be removed by the v1→v2 upgrader")
		})
	}
}
