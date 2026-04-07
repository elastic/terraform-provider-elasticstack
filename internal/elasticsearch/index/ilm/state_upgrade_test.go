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

package ilm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func TestILMResourceUpgradeState(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"id":   "cluster-uuid/policy-x",
		"name": "policy-x",
		"hot": []any{
			map[string]any{
				"min_age": "1h",
				"set_priority": []any{
					map[string]any{"priority": float64(10)},
				},
				"rollover": []any{
					map[string]any{"max_age": "1d"},
				},
				"readonly": []any{
					map[string]any{"enabled": true},
				},
			},
		},
		"delete": []any{
			map[string]any{
				"min_age": "0ms",
				"delete": []any{
					map[string]any{"delete_searchable_snapshot": true},
				},
			},
		},
		"elasticsearch_connection": []any{
			map[string]any{"username": "u"},
		},
	}
	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)

	r := &Resource{}
	upgraders := r.UpgradeState(context.Background())
	up, ok := upgraders[0]
	require.True(t, ok)

	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: rawJSON},
	}
	resp := &resource.UpgradeStateResponse{}
	up.StateUpgrader(context.Background(), req, resp)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)

	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

	hot, ok := got["hot"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "1h", hot["min_age"])
	sp, ok := hot["set_priority"].(map[string]any)
	require.True(t, ok)
	priority, ok := sp["priority"].(float64)
	require.True(t, ok)
	require.InEpsilon(t, 10.0, priority, 0.0001)
	_, listSP := hot["set_priority"].([]any)
	require.False(t, listSP)

	del, ok := got["delete"].(map[string]any)
	require.True(t, ok)
	innerDel, ok := del["delete"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, innerDel["delete_searchable_snapshot"])

	conn, ok := got["elasticsearch_connection"].([]any)
	require.True(t, ok)
	require.Len(t, conn, 1)
}
