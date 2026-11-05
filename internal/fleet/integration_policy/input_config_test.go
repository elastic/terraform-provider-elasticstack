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

package integrationpolicy

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/require"
)

// APM package policy as returned by the typed GET after Kibana's APM plugin has
// written its API keys into the input config.
const apmPolicyWithKeys = `{
  "id": "apm-pp", "name": "apm", "namespace": "default", "enabled": true,
  "revision": 2, "version": "WzEsMV0=",
  "created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
  "package": {"name": "apm", "version": "9.4.2"},
  "inputs": [{
    "id": "apm-pp", "type": "apm", "policy_template": "apmserver", "enabled": true,
    "vars": {"host": {"type": "text", "value": "0.0.0.0:8200"}},
    "config": {"apm-server": {"value": {
      "rum": {"source_mapping": {"metadata": [], "elasticsearch": {"api_key": "sm-id:sm-secret"}}},
      "agent_config": [],
      "agent": {"config": {"elasticsearch": {"api_key": "ac-id:ac-secret"}}}
    }}},
    "streams": [],
    "compiled_input": {"apm-server": {"host": "0.0.0.0:8200"}}
  }]
}`

// The same policy after a simplified update: the input id changed, a var was
// updated, and the APM plugin rewrote the source map metadata but the API keys
// are gone.
const apmPolicyAfterSimplifiedUpdate = `{
  "id": "apm-pp", "name": "apm", "namespace": "default", "enabled": true,
  "revision": 3, "version": "WzIsMV0=",
  "created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
  "package": {"name": "apm", "version": "9.4.2"},
  "inputs": [{
    "id": "apm-apmserver-apm-pp", "type": "apm", "policy_template": "apmserver", "enabled": true,
    "vars": {"host": {"type": "text", "value": "0.0.0.0:8201"}},
    "config": {"apm-server": {"value": {
      "rum": {"source_mapping": {"metadata": [{"service": {"name": "web"}}]}},
      "agent_config": []
    }}},
    "streams": [],
    "compiled_input": {"apm-server": {"host": "0.0.0.0:8201"}}
  }]
}`

func mustPolicy(t *testing.T, data string) *kbapi.PackagePolicy {
	t.Helper()
	var p kbapi.PackagePolicy
	require.NoError(t, json.Unmarshal([]byte(data), &p))
	return &p
}

func requestInputsAsMaps(t *testing.T, inputs []kbapi.PackagePolicyRequestTypedInput) []map[string]any {
	t.Helper()
	data, err := json.Marshal(inputs)
	require.NoError(t, err)
	var out []map[string]any
	require.NoError(t, json.Unmarshal(data, &out))
	return out
}

func TestSnapshotInputConfig(t *testing.T) {
	t.Run("records input config by type and policy template", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(mustPolicy(t, apmPolicyWithKeys))
		require.NoError(t, err)
		require.False(t, snapshot.isEmpty())
		require.Contains(t, snapshot.inputs, "apm\x00apmserver")
		require.Empty(t, snapshot.streams)
	})

	t.Run("empty when no input or stream has config", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(mustPolicy(t, `{
			"id": "p", "name": "p", "enabled": true, "revision": 1,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"type": "tcp", "policy_template": "tcp", "enabled": true,
				"streams": [{"enabled": true, "data_stream": {"type": "logs", "dataset": "tcp.generic"}}]}]
		}`))
		require.NoError(t, err)
		require.True(t, snapshot.isEmpty())
	})

	t.Run("records stream config by dataset", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(mustPolicy(t, `{
			"id": "p", "name": "p", "enabled": true, "revision": 1,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"type": "tcp", "policy_template": "tcp", "enabled": true,
				"streams": [{"enabled": true, "data_stream": {"type": "logs", "dataset": "tcp.generic"},
					"config": {"extra": {"value": 1}}}]}]
		}`))
		require.NoError(t, err)
		require.Empty(t, snapshot.inputs)
		require.Equal(t, map[string]any{"extra": map[string]any{"value": float64(1)}}, snapshot.streams["tcp\x00tcp\x00tcp.generic"])
	})

	t.Run("ambiguous identities are skipped", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(mustPolicy(t, `{
			"id": "p", "name": "p", "enabled": true, "revision": 1,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [
				{"type": "x", "enabled": true, "streams": [], "config": {"a": {"value": 1}}},
				{"type": "x", "enabled": true, "streams": [], "config": {"b": {"value": 2}}},
				{"type": "x", "enabled": true, "streams": [], "config": {"c": {"value": 3}}}
			]
		}`))
		require.NoError(t, err)
		require.True(t, snapshot.isEmpty())
	})

	t.Run("nil policy", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(nil)
		require.NoError(t, err)
		require.True(t, snapshot.isEmpty())
	})
}

func TestDeepFill(t *testing.T) {
	dst := map[string]any{
		"kept":   "new",
		"nested": map[string]any{"b": "new-b"},
		"nulled": nil,
		"list":   []any{"new"},
	}
	src := map[string]any{
		"kept":    "old",
		"missing": "old-missing",
		"nested":  map[string]any{"a": "old-a", "b": "old-b"},
		"nulled":  "old-nulled",
		"list":    []any{"old"},
	}
	require.True(t, deepFill(dst, src))
	require.Equal(t, map[string]any{
		"kept":    "new",
		"missing": "old-missing",
		"nested":  map[string]any{"a": "old-a", "b": "new-b"},
		"nulled":  "old-nulled",
		"list":    []any{"new"},
	}, dst)

	require.False(t, deepFill(dst, src), "a second fill must be a no-op")
}

func TestRestoreInputConfig(t *testing.T) {
	t.Run("restores APM API keys and keeps values written by the update", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(mustPolicy(t, apmPolicyWithKeys))
		require.NoError(t, err)

		inputs, changed, err := restoreInputConfig(snapshot, mustPolicy(t, apmPolicyAfterSimplifiedUpdate))
		require.NoError(t, err)
		require.True(t, changed)
		require.Len(t, inputs, 1)

		raw := requestInputsAsMaps(t, inputs)[0]
		require.NotContains(t, raw, "compiled_input")
		require.Equal(t, "apm-apmserver-apm-pp", raw["id"])
		require.Equal(t, "0.0.0.0:8201", raw["vars"].(map[string]any)["host"].(map[string]any)["value"])

		value := raw["config"].(map[string]any)["apm-server"].(map[string]any)["value"].(map[string]any)
		sourceMapping := value["rum"].(map[string]any)["source_mapping"].(map[string]any)
		require.Equal(t, "sm-id:sm-secret", sourceMapping["elasticsearch"].(map[string]any)["api_key"])
		require.Equal(t, []any{map[string]any{"service": map[string]any{"name": "web"}}}, sourceMapping["metadata"])
		require.Equal(t, "ac-id:ac-secret", value["agent"].(map[string]any)["config"].(map[string]any)["elasticsearch"].(map[string]any)["api_key"])
	})

	t.Run("no change when config survived the update", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(mustPolicy(t, apmPolicyWithKeys))
		require.NoError(t, err)

		inputs, changed, err := restoreInputConfig(snapshot, mustPolicy(t, apmPolicyWithKeys))
		require.NoError(t, err)
		require.False(t, changed)
		require.Nil(t, inputs)
	})

	t.Run("restores config removed entirely and stream config", func(t *testing.T) {
		before := `{
			"id": "p", "name": "p", "enabled": true, "revision": 1,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"id": "p", "type": "osquery", "enabled": true,
				"config": {"osquery": {"value": {"packs": {"pack-1": {"queries": {}}}}}},
				"streams": [{"id": "s", "enabled": true, "data_stream": {"type": "logs", "dataset": "osquery_manager.result"},
					"config": {"extra": {"value": true}}}]}]
		}`
		after := `{
			"id": "p", "name": "p", "enabled": true, "revision": 2,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"id": "osquery-p", "type": "osquery", "enabled": true,
				"streams": [{"id": "s2", "enabled": true, "data_stream": {"type": "logs", "dataset": "osquery_manager.result"},
					"compiled_stream": {"x": 1}}]}]
		}`
		snapshot, err := snapshotInputConfig(mustPolicy(t, before))
		require.NoError(t, err)

		inputs, changed, err := restoreInputConfig(snapshot, mustPolicy(t, after))
		require.NoError(t, err)
		require.True(t, changed)

		raw := requestInputsAsMaps(t, inputs)[0]
		require.Equal(t, map[string]any{"osquery": map[string]any{"value": map[string]any{"packs": map[string]any{"pack-1": map[string]any{"queries": map[string]any{}}}}}}, raw["config"])
		stream := raw["streams"].([]any)[0].(map[string]any)
		require.NotContains(t, stream, "compiled_stream")
		require.Equal(t, map[string]any{"extra": map[string]any{"value": true}}, stream["config"])
	})

	t.Run("inputs that no longer exist are ignored", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(mustPolicy(t, apmPolicyWithKeys))
		require.NoError(t, err)

		inputs, changed, err := restoreInputConfig(snapshot, mustPolicy(t, `{
			"id": "p", "name": "p", "enabled": true, "revision": 2,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"type": "other", "enabled": true, "streams": []}]
		}`))
		require.NoError(t, err)
		require.False(t, changed)
		require.Nil(t, inputs)
	})

	t.Run("empty snapshot never changes anything", func(t *testing.T) {
		inputs, changed, err := restoreInputConfig(inputConfigSnapshot{}, mustPolicy(t, apmPolicyAfterSimplifiedUpdate))
		require.NoError(t, err)
		require.False(t, changed)
		require.Nil(t, inputs)
	})
}

func TestInputConfigAmbiguousIdentities(t *testing.T) {
	// Two inputs share type and policy template; only the first has config.
	before := `{
		"id": "p", "name": "p", "enabled": true, "revision": 1,
		"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
		"inputs": [
			{"id": "a", "type": "x", "policy_template": "t", "enabled": true, "streams": [], "config": {"keys": {"value": "secret"}}},
			{"id": "b", "type": "x", "policy_template": "t", "enabled": true, "streams": []}
		]
	}`

	t.Run("configured and unconfigured duplicates are not snapshotted", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(mustPolicy(t, before))
		require.NoError(t, err)
		require.True(t, snapshot.isEmpty())
	})

	t.Run("configured and unconfigured duplicate streams are not snapshotted", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(mustPolicy(t, `{
			"id": "p", "name": "p", "enabled": true, "revision": 1,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"type": "x", "enabled": true, "streams": [
				{"enabled": true, "data_stream": {"type": "logs", "dataset": "d"}, "config": {"k": {"value": 1}}},
				{"enabled": true, "data_stream": {"type": "logs", "dataset": "d"}}
			]}]
		}`))
		require.NoError(t, err)
		require.Empty(t, snapshot.streams)
	})

	t.Run("identity that becomes ambiguous after the update is not restored", func(t *testing.T) {
		snapshot, err := snapshotInputConfig(mustPolicy(t, `{
			"id": "p", "name": "p", "enabled": true, "revision": 1,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"id": "a", "type": "x", "policy_template": "t", "enabled": true, "streams": [], "config": {"keys": {"value": "secret"}}}]
		}`))
		require.NoError(t, err)
		require.False(t, snapshot.isEmpty())

		after := mustPolicy(t, `{
			"id": "p", "name": "p", "enabled": true, "revision": 2,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [
				{"id": "a2", "type": "x", "policy_template": "t", "enabled": true, "streams": []},
				{"id": "b2", "type": "x", "policy_template": "t", "enabled": true, "streams": []}
			]
		}`)
		inputs, changed, err := restoreInputConfig(snapshot, after)
		require.NoError(t, err)
		require.False(t, changed)
		require.Nil(t, inputs)
	})
}

func TestRestorePreservedInputConfig_PolicyNotFound(t *testing.T) {
	var methods []string
	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		w.WriteHeader(http.StatusNotFound)
	}))

	snapshot, err := snapshotInputConfig(mustPolicy(t, apmPolicyWithKeys))
	require.NoError(t, err)

	diags := restorePreservedInputConfig(context.Background(), client, "apm-pp", "", snapshot)
	require.True(t, diags.HasError(), "a policy deleted before the restoring read must be an error")
	require.Contains(t, diags.Errors()[0].Detail(), "apm-pp")
	require.Equal(t, []string{http.MethodGet}, methods, "only the restoring read may be sent when the policy is gone")
}

func TestRestorePreservedInputConfig_EmptySnapshotSendsNothing(t *testing.T) {
	client := newTestFleetClient(t, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	require.False(t, restorePreservedInputConfig(context.Background(), client, "p", "", inputConfigSnapshot{}).HasError())
}

func TestRestorePreservedInputConfig_RequestBody(t *testing.T) {
	after := mustPolicy(t, apmPolicyAfterSimplifiedUpdate)
	policyID := "agent-policy-1"
	after.PolicyId = &policyID
	after.PolicyIds = &[]string{policyID}
	afterJSON, err := json.Marshal(map[string]any{"item": after})
	require.NoError(t, err)

	var putBody map[string]any
	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			if err := json.NewDecoder(r.Body).Decode(&putBody); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		_, _ = w.Write(afterJSON)
	}))

	snapshot, err := snapshotInputConfig(mustPolicy(t, apmPolicyWithKeys))
	require.NoError(t, err)
	diags := restorePreservedInputConfig(context.Background(), client, "apm-pp", "", snapshot)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, putBody, "a restoring update must be sent")
	keys := make([]string, 0, len(putBody))
	for k := range putBody {
		keys = append(keys, k)
	}
	require.ElementsMatch(t, []string{"version", "inputs", "policy_id", "policy_ids"}, keys)
	require.Equal(t, "WzIsMV0=", putBody["version"])
	require.Equal(t, policyID, putBody["policy_id"])
	require.Equal(t, []any{policyID}, putBody["policy_ids"])
}
