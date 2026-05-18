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

package security_role

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// UpgradeState migrates Plugin SDKv2 state (version 0) into the Plugin
// Framework shape used by this resource (version 1). The schema change
// reshapes a few SDK-legacy block layouts:
//
//   - elasticsearch: SDK TypeList(MinItems:1,MaxItems:1) -> PF SingleNestedBlock.
//     Unwrap the single-element list into a bare object.
//   - elasticsearch.indices[*].field_security and
//     elasticsearch.remote_indices[*].field_security: SDK TypeList(MaxItems:1)
//     -> PF SingleNestedBlock. Unwrap (empty list -> null, one-element list ->
//     the element).
//
// In addition, optional sets that SDKv2 persisted as `[]` (cluster, run_as,
// base, field_security.grant, field_security.except) are normalised to null
// so the migrated state matches what the PF flatten functions produce. This
// avoids spurious set-element identity diffs on the first plan after upgrade.
func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: migrateV0ToV1,
		},
	}
}

func migrateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	var state map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &state); err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal raw state", err.Error())
		return
	}

	emptyStringToNull(state, "description")

	if esList, ok := state["elasticsearch"].([]any); ok {
		var es map[string]any
		if len(esList) > 0 {
			es, _ = esList[0].(map[string]any)
		}
		if es == nil {
			state["elasticsearch"] = nil
		} else {
			emptySetToNull(es, "cluster")
			emptySetToNull(es, "run_as")
			emptySetToNull(es, "remote_indices")
			unwrapIndexFieldSecurity(es, "indices")
			unwrapIndexFieldSecurity(es, "remote_indices")
			state["elasticsearch"] = es
		}
	}

	if kibList, ok := state["kibana"].([]any); ok {
		for _, raw := range kibList {
			kib, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			emptySetToNull(kib, "base")
			emptySetToNull(kib, "feature")
		}
	}

	out, err := json.Marshal(state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal upgraded state", err.Error())
		return
	}
	resp.DynamicValue = &tfprotov6.DynamicValue{JSON: out}
}

// emptySetToNull normalises a `key` whose value is an empty array to JSON
// null. SDKv2 stored omitted optional `TypeSet` attributes as `[]`, whereas
// PF flatten produces null; without this both refresh and plan would see a
// known-empty-vs-null mismatch on every run.
func emptySetToNull(m map[string]any, key string) {
	v, ok := m[key]
	if !ok {
		return
	}
	arr, ok := v.([]any)
	if !ok {
		return
	}
	if len(arr) == 0 {
		m[key] = nil
	}
}

// unwrapIndexFieldSecurity converts `field_security` from a list of (0 or 1)
// objects to a single object (or null) on every entry of the named indices
// list, and normalises grant/except empty arrays to null on the unwrapped
// object.
func unwrapIndexFieldSecurity(es map[string]any, indicesKey string) {
	raw, ok := es[indicesKey]
	if !ok || raw == nil {
		return
	}
	indices, ok := raw.([]any)
	if !ok {
		return
	}
	for _, entryRaw := range indices {
		entry, ok := entryRaw.(map[string]any)
		if !ok {
			continue
		}
		emptyStringToNull(entry, "query")
		fsRaw, ok := entry["field_security"]
		if !ok {
			continue
		}
		fsList, ok := fsRaw.([]any)
		if !ok {
			continue
		}
		if len(fsList) == 0 {
			entry["field_security"] = nil
			continue
		}
		// Leave grant/except as-is. The resource flatten normalises missing
		// API keys to known-empty sets, so v0 entries that stored `[]` (the
		// SDK default for omitted optional sets) already align with what PF
		// Read produces.
		entry["field_security"] = fsList[0]
	}
}

// emptyStringToNull normalises an empty-string `key` to null. SDKv2 stored
// omitted optional `TypeString` attributes as "" rather than null; PF flatten
// produces null, so without this the difference perturbs set element
// identity in nested SetNestedBlocks.
func emptyStringToNull(m map[string]any, key string) {
	v, ok := m[key]
	if !ok {
		return
	}
	s, ok := v.(string)
	if !ok {
		return
	}
	if s == "" {
		m[key] = nil
	}
}
