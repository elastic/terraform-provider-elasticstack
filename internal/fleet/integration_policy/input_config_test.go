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
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
		"kept":      "new",
		"nested":    map[string]any{"b": "new-b"},
		"nulled":    nil,
		"both_null": nil,
		"list":      []any{"new"},
	}
	src := map[string]any{
		"kept":         "old",
		"missing":      "old-missing",
		"missing_null": nil,
		"nested":       map[string]any{"a": "old-a", "b": "old-b", "c": nil},
		"nulled":       "old-nulled",
		"both_null":    nil,
		"list":         []any{"old"},
	}
	require.True(t, deepFill(dst, src))
	require.Equal(t, map[string]any{
		"kept":         "new",
		"missing":      "old-missing",
		"missing_null": nil,
		"nested":       map[string]any{"a": "old-a", "b": "new-b", "c": nil},
		"nulled":       "old-nulled",
		"both_null":    nil,
		"list":         []any{"new"},
	}, dst)

	require.False(t, deepFill(dst, src), "a second fill must be a no-op")
}

func TestDeepFill_NullOnlyChange(t *testing.T) {
	t.Run("a nested key with a null value removed under a kept parent is restored", func(t *testing.T) {
		dst := map[string]any{"parent": map[string]any{"kept": 1}}
		src := map[string]any{"parent": map[string]any{"kept": 1, "removed": nil}}
		require.True(t, deepFill(dst, src))
		require.Equal(t, map[string]any{"parent": map[string]any{"kept": 1, "removed": nil}}, dst)
	})

	t.Run("a key null on both sides is not a change", func(t *testing.T) {
		dst := map[string]any{"parent": map[string]any{"same": nil}}
		src := map[string]any{"parent": map[string]any{"same": nil}}
		require.False(t, deepFill(dst, src))
	})
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

	t.Run("restores nested keys with null values removed under a kept parent", func(t *testing.T) {
		before := `{
			"id": "p", "name": "p", "enabled": true, "revision": 1,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"type": "x", "policy_template": "x", "enabled": true,
				"config": {"settings": {"value": {"parent": {"kept": 1, "removed": null}}}},
				"streams": [{"enabled": true, "data_stream": {"type": "logs", "dataset": "x.log"},
					"config": {"settings": {"value": {"opt": null, "kept": true}}}}]}]
		}`
		after := `{
			"id": "p", "name": "p", "enabled": true, "revision": 2,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"type": "x", "policy_template": "x", "enabled": true,
				"config": {"settings": {"value": {"parent": {"kept": 1}}}},
				"streams": [{"enabled": true, "data_stream": {"type": "logs", "dataset": "x.log"},
					"config": {"settings": {"value": {"kept": true}}}}]}]
		}`
		snapshot, err := snapshotInputConfig(mustPolicy(t, before))
		require.NoError(t, err)

		inputs, changed, err := restoreInputConfig(snapshot, mustPolicy(t, after))
		require.NoError(t, err)
		require.True(t, changed)

		raw := requestInputsAsMaps(t, inputs)[0]
		parent := raw["config"].(map[string]any)["settings"].(map[string]any)["value"].(map[string]any)["parent"].(map[string]any)
		require.Equal(t, map[string]any{"kept": float64(1), "removed": nil}, parent)
		stream := raw["streams"].([]any)[0].(map[string]any)
		require.Equal(t, map[string]any{"opt": nil, "kept": true}, stream["config"].(map[string]any)["settings"].(map[string]any)["value"])
	})

	t.Run("restores a key that the update left null", func(t *testing.T) {
		before := `{
			"id": "p", "name": "p", "enabled": true, "revision": 1,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"type": "x", "policy_template": "x", "enabled": true, "streams": [],
				"config": {"settings": {"value": {"a": 1, "b": "kept"}}}}]
		}`
		after := `{
			"id": "p", "name": "p", "enabled": true, "revision": 2,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"type": "x", "policy_template": "x", "enabled": true, "streams": [],
				"config": {"settings": {"value": {"a": null, "b": "kept"}}}}]
		}`
		snapshot, err := snapshotInputConfig(mustPolicy(t, before))
		require.NoError(t, err)

		inputs, changed, err := restoreInputConfig(snapshot, mustPolicy(t, after))
		require.NoError(t, err)
		require.True(t, changed)
		raw := requestInputsAsMaps(t, inputs)[0]
		require.Equal(t, map[string]any{"a": float64(1), "b": "kept"}, raw["config"].(map[string]any)["settings"].(map[string]any)["value"])
	})

	t.Run("no change when a restored null cannot be carried by the typed request", func(t *testing.T) {
		// The typed request omits a null config entry `value`, so restoring it
		// would send a body identical to the current inputs.
		before := `{
			"id": "p", "name": "p", "enabled": true, "revision": 1,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"type": "x", "policy_template": "x", "enabled": true, "streams": [],
				"config": {"settings": {"type": "text", "value": null}}}]
		}`
		after := `{
			"id": "p", "name": "p", "enabled": true, "revision": 2,
			"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
			"inputs": [{"type": "x", "policy_template": "x", "enabled": true, "streams": [],
				"config": {"settings": {"type": "text"}}}]
		}`
		snapshot, err := snapshotInputConfig(mustPolicy(t, before))
		require.NoError(t, err)

		inputs, changed, err := restoreInputConfig(snapshot, mustPolicy(t, after))
		require.NoError(t, err)
		require.False(t, changed)
		require.Nil(t, inputs)
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

func TestRestorePreservedInputConfig_StaleSnapshot(t *testing.T) {
	// Another client wrote the policy between the snapshot (revision 2) and the
	// restoring read: the restoring read shows revision 4 instead of 3.
	after := mustPolicy(t, apmPolicyAfterSimplifiedUpdate)
	after.Revision = 4
	afterJSON, err := json.Marshal(map[string]any{"item": after})
	require.NoError(t, err)

	var methods []string
	client := newTestFleetClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(afterJSON)
	}))

	snapshot, err := snapshotInputConfig(mustPolicy(t, apmPolicyWithKeys))
	require.NoError(t, err)
	diags := restorePreservedInputConfig(context.Background(), client, "apm-pp", "", snapshot)
	require.True(t, diags.HasError(), "restoring from a snapshot older than the last write must be an error")
	require.Contains(t, diags.Errors()[0].Detail(), "apm-pp")
	require.Equal(t, []string{http.MethodGet}, methods, "no restoring update may be sent from a stale snapshot")
}

// osqueryPolicy is a typed package policy with the given revision and packs in
// its input config. No packs means no input config.
func osqueryPolicy(t *testing.T, revision int, packs ...string) map[string]any {
	t.Helper()
	input := map[string]any{"id": "osquery-pp", "type": "osquery", "policy_template": "osquery_manager", "enabled": true, "streams": []any{}}
	if len(packs) > 0 {
		packConfig := map[string]any{}
		for _, pack := range packs {
			packConfig[pack] = map[string]any{"queries": map[string]any{}}
		}
		input["config"] = map[string]any{"osquery": map[string]any{"value": map[string]any{"packs": packConfig}}}
	}
	return map[string]any{
		"id": "osquery-pp", "name": "osquery", "namespace": "default", "enabled": true,
		"revision": revision, "version": fmt.Sprintf("v%d", revision),
		"created_at": "", "created_by": "", "updated_at": "", "updated_by": "",
		"package": map[string]any{"name": "osquery_manager", "version": "1.0.0"},
		"inputs":  []any{input},
	}
}

// fakeFleet serves a single package policy. A simplified PUT drops the input
// config, as Kibana does, and bumps the revision; a typed PUT replaces the
// inputs. beforeSimplifiedPut, when set, runs first to model a concurrent write.
type fakeFleet struct {
	t                   *testing.T
	policy              map[string]any
	beforeSimplifiedPut func(f *fakeFleet)
	getStatus           int
	typedPutStatus      int
	requests            []string
	typedPutBody        map[string]any
}

func (f *fakeFleet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	simplified := r.URL.Query().Get("format") == "simplified"
	f.requests = append(f.requests, fmt.Sprintf("%s simplified=%t", r.Method, simplified))
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodGet && f.getStatus != 0:
		w.WriteHeader(f.getStatus)
		_, _ = fmt.Fprintf(w, `{"statusCode":%d,"error":"error","message":"boom"}`, f.getStatus)
		return
	case r.Method == http.MethodPut && simplified:
		if f.beforeSimplifiedPut != nil {
			f.beforeSimplifiedPut(f)
		}
		f.policy = osqueryPolicy(f.t, f.policy["revision"].(int)+1)
	case r.Method == http.MethodPut && f.typedPutStatus != 0:
		w.WriteHeader(f.typedPutStatus)
		_, _ = fmt.Fprintf(w, `{"statusCode":%d,"error":"error","message":"boom"}`, f.typedPutStatus)
		return
	case r.Method == http.MethodPut:
		if err := json.NewDecoder(r.Body).Decode(&f.typedPutBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next := osqueryPolicy(f.t, f.policy["revision"].(int)+1)
		next["inputs"] = f.typedPutBody["inputs"]
		f.policy = next
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"item": f.policy})
}

func osqueryUpdateBody() kbapi.PackagePolicyRequestMappedInputs {
	return kbapi.PackagePolicyRequestMappedInputs{
		Name:    "osquery",
		Package: kbapi.PackagePolicyRequestPackage{Name: "osquery_manager", Version: "1.0.0"},
	}
}

func TestUpdatePackagePolicyPreservingInputConfig(t *testing.T) {
	for name, status := range map[string]int{"fails": http.StatusInternalServerError, "does not find the policy": http.StatusNotFound} {
		t.Run("a snapshot read that "+name+" prevents the simplified update", func(t *testing.T) {
			fake := &fakeFleet{t: t, policy: osqueryPolicy(t, 1, "pack-1"), getStatus: status}
			client := newTestFleetClient(t, fake)

			policy, diags := updatePackagePolicyPreservingInputConfig(context.Background(), client, "osquery-pp", "", osqueryUpdateBody())
			require.True(t, diags.HasError(), "a snapshot read that %s must be an error", name)
			require.Nil(t, policy)
			require.Equal(t, []string{"GET simplified=false"}, fake.requests, "no update may be sent without a snapshot")
		})
	}

	t.Run("config removed by the update is restored", func(t *testing.T) {
		fake := &fakeFleet{t: t, policy: osqueryPolicy(t, 1, "pack-1", "pack-2")}
		client := newTestFleetClient(t, fake)

		policy, diags := updatePackagePolicyPreservingInputConfig(context.Background(), client, "osquery-pp", "", osqueryUpdateBody())
		require.False(t, diags.HasError(), "%v", diags)
		require.NotNil(t, policy)
		require.Equal(t, []string{"GET simplified=false", "PUT simplified=true", "GET simplified=false", "PUT simplified=false"}, fake.requests)

		inputs := fake.typedPutBody["inputs"].([]any)
		packs := inputs[0].(map[string]any)["config"].(map[string]any)["osquery"].(map[string]any)["value"].(map[string]any)["packs"].(map[string]any)
		require.Contains(t, packs, "pack-1")
		require.Contains(t, packs, "pack-2")
	})

	t.Run("a pack removed between the snapshot and the update is not restored", func(t *testing.T) {
		fake := &fakeFleet{
			t:      t,
			policy: osqueryPolicy(t, 1, "pack-1", "pack-2"),
			beforeSimplifiedPut: func(f *fakeFleet) {
				// Osquery deletes pack-2 after the snapshot was taken.
				f.policy = osqueryPolicy(f.t, f.policy["revision"].(int)+1, "pack-1")
			},
		}
		client := newTestFleetClient(t, fake)

		_, diags := updatePackagePolicyPreservingInputConfig(context.Background(), client, "osquery-pp", "", osqueryUpdateBody())
		require.True(t, diags.HasError(), "a concurrent write since the snapshot must be an error")
		require.Contains(t, diags.Errors()[0].Detail(), "osquery-pp")
		requireConfigNotRestoredError(t, diags)
		require.Equal(t, []string{"GET simplified=false", "PUT simplified=true", "GET simplified=false"}, fake.requests,
			"no restoring update may be sent from a stale snapshot")
		require.Nil(t, fake.typedPutBody)
	})
}

// requireConfigNotRestoredError asserts that the diagnostics tell the user the
// config removed by the update was not written back and that re-running apply
// does not bring it back.
func requireConfigNotRestoredError(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	for _, d := range diags.Errors() {
		if d.Summary() == "Integration policy input config not restored" {
			require.Contains(t, d.Detail(), "Re-running apply does not restore it")
			return
		}
	}
	require.Fail(t, "missing the input config not restored error", "%v", diags)
}

func TestUpdatePackagePolicyPreservingInputConfig_RestoreFailure(t *testing.T) {
	t.Run("restoring read fails", func(t *testing.T) {
		fake := &fakeFleet{t: t, policy: osqueryPolicy(t, 1, "pack-1")}
		fake.beforeSimplifiedPut = func(f *fakeFleet) { f.getStatus = http.StatusInternalServerError }
		client := newTestFleetClient(t, fake)

		policy, diags := updatePackagePolicyPreservingInputConfig(context.Background(), client, "osquery-pp", "", osqueryUpdateBody())
		require.True(t, diags.HasError())
		require.Nil(t, policy)
		requireConfigNotRestoredError(t, diags)
		require.Equal(t, []string{"GET simplified=false", "PUT simplified=true", "GET simplified=false"}, fake.requests)
	})

	t.Run("restoring update fails", func(t *testing.T) {
		fake := &fakeFleet{t: t, policy: osqueryPolicy(t, 1, "pack-1"), typedPutStatus: http.StatusInternalServerError}
		client := newTestFleetClient(t, fake)

		policy, diags := updatePackagePolicyPreservingInputConfig(context.Background(), client, "osquery-pp", "", osqueryUpdateBody())
		require.True(t, diags.HasError())
		require.Nil(t, policy)
		require.Len(t, diags.Errors(), 2, "the restoring update error and the config not restored error: %v", diags)
		requireConfigNotRestoredError(t, diags)
		require.Equal(t, []string{"GET simplified=false", "PUT simplified=true", "GET simplified=false", "PUT simplified=false"}, fake.requests)
	})
}
