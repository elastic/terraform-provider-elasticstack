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

package transform

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

// TestTransformResourceUpgradeStateV0ToV1 verifies the v0→v1 upgrader unwraps
// the singleton-list shape produced by the SDK schema into single objects for
// every block that became SingleNestedBlock, while leaving the multi-element
// destination.aliases list and the elasticsearch_connection block untouched.
func TestTransformResourceUpgradeStateV0ToV1(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"id":   "cluster-uuid/transform-x",
		"name": "transform-x",
		"source": []any{
			map[string]any{
				"indices": []any{"src-1", "src-2"},
				"query":   `{"match_all":{}}`,
			},
		},
		"destination": []any{
			map[string]any{
				"index":    "dest-x",
				"pipeline": "ingest-1",
				"aliases": []any{
					map[string]any{"alias": "alias-1", "move_on_creation": true},
					map[string]any{"alias": "alias-2", "move_on_creation": false},
				},
			},
		},
		"sync": []any{
			map[string]any{
				"time": []any{
					map[string]any{"field": "@timestamp", "delay": "30s"},
				},
			},
		},
		"retention_policy": []any{
			map[string]any{
				"time": []any{
					map[string]any{"field": "@timestamp", "max_age": "30d"},
				},
			},
		},
		"elasticsearch_connection": []any{
			map[string]any{"username": "u"},
		},
	}
	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)

	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: rawJSON}}
	resp := &resource.UpgradeStateResponse{}
	migrateStateV0ToV1(context.Background(), req, resp)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)

	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

	src, ok := got["source"].(map[string]any)
	require.True(t, ok, "source should be a single object")
	require.Equal(t, []any{"src-1", "src-2"}, src["indices"])

	dst, ok := got["destination"].(map[string]any)
	require.True(t, ok, "destination should be a single object")
	require.Equal(t, "dest-x", dst["index"])
	aliases, ok := dst["aliases"].([]any)
	require.True(t, ok, "destination.aliases must remain a list")
	require.Len(t, aliases, 2)

	sync, ok := got["sync"].(map[string]any)
	require.True(t, ok, "sync should be a single object")
	syncTime, ok := sync["time"].(map[string]any)
	require.True(t, ok, "sync.time should be a single object")
	require.Equal(t, "@timestamp", syncTime["field"])

	rp, ok := got["retention_policy"].(map[string]any)
	require.True(t, ok, "retention_policy should be a single object")
	rpTime, ok := rp["time"].(map[string]any)
	require.True(t, ok, "retention_policy.time should be a single object")
	require.Equal(t, "30d", rpTime["max_age"])

	conn, ok := got["elasticsearch_connection"].([]any)
	require.True(t, ok, "elasticsearch_connection must remain a list")
	require.Len(t, conn, 1)
}

// TestTransformResourceUpgradeStateV0ToV1_MultiElementListErrors ensures the
// upgrader surfaces a clear diagnostic when a singleton-list path holds more
// than one element — that shape was disallowed by the prior schema, so finding
// it on disk indicates corruption rather than a recoverable upgrade.
func TestTransformResourceUpgradeStateV0ToV1_MultiElementListErrors(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"id":   "cluster-uuid/transform-z",
		"name": "transform-z",
		"source": []any{
			map[string]any{"indices": []any{"a"}},
			map[string]any{"indices": []any{"b"}},
		},
		"destination": []any{map[string]any{"index": "d"}},
	}
	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)

	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: rawJSON}}
	resp := &resource.UpgradeStateResponse{}
	migrateStateV0ToV1(context.Background(), req, resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics[0].Detail(), `"source"`)
}

// TestTransformResourceUpgradeStateV0ToV1_OptionalBlocksOmitted ensures the
// upgrader handles state where the optional retention_policy / sync blocks were
// not configured (stored as either nil or empty list).
func TestTransformResourceUpgradeStateV0ToV1_OptionalBlocksOmitted(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"id":   "cluster-uuid/transform-y",
		"name": "transform-y",
		"source": []any{
			map[string]any{"indices": []any{"src"}},
		},
		"destination": []any{
			map[string]any{"index": "dest"},
		},
		"sync":             []any{},
		"retention_policy": nil,
	}
	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)

	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: rawJSON}}
	resp := &resource.UpgradeStateResponse{}
	migrateStateV0ToV1(context.Background(), req, resp)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)

	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

	_, present := got["sync"]
	require.False(t, present, "empty sync list should be removed from state")

	_, present = got["retention_policy"]
	require.False(t, present, "nil retention_policy should be removed from state")
}
