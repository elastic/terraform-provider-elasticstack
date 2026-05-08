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
	"fmt"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// migrateStateV0ToV1 unwraps singleton-list nested blocks (source, destination,
// retention_policy, sync, and the inner time blocks within retention_policy and
// sync) into single objects. The schema previously modeled these as
// ListNestedBlock with SizeBetween(1,1) or SizeAtMost(1) and is now
// SingleNestedBlock. The aliases block remains a list and is left unchanged.
func migrateStateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	var stateMap map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &stateMap); err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not unmarshal prior state: "+err.Error())
		return
	}

	for _, key := range [...]string{"source", "destination", "retention_policy", "sync"} {
		resp.Diagnostics.Append(collapseSingletonList(stateMap, key, key)...)
	}
	for _, parent := range [...]string{"retention_policy", "sync"} {
		parentObj, ok := stateMap[parent].(map[string]any)
		if !ok {
			continue
		}
		resp.Diagnostics.Append(collapseSingletonList(parentObj, "time", parent+".time")...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// The SDK provider stored unset JSON string attributes as "" rather than
	// null. The Plugin Framework jsontypes.NormalizedType rejects empty strings,
	// so normalise them to nil before marshalling the upgraded state.
	normaliseEmptyJSONStrings(stateMap)

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not marshal new state: "+err.Error())
		return
	}
	resp.DynamicValue = &tfprotov6.DynamicValue{JSON: stateJSON}
}

// normaliseEmptyJSONStrings converts empty-string values stored by the SDK
// provider for optional JSON attributes into nil so the Plugin Framework
// jsontypes.NormalizedType can accept them.
func normaliseEmptyJSONStrings(state map[string]any) {
	for _, key := range []string{"metadata", "pivot", "latest"} {
		if v, ok := state[key].(string); ok && v == "" {
			state[key] = nil
		}
	}

	if src, ok := state["source"].(map[string]any); ok {
		for _, key := range []string{"query", "runtime_mappings"} {
			if v, ok := src[key].(string); ok && v == "" {
				src[key] = nil
			}
		}
	}

	if dst, ok := state["destination"].(map[string]any); ok {
		if v, ok := dst["pipeline"].(string); ok && v == "" {
			dst["pipeline"] = nil
		}
	}
}

// collapseSingletonList unwraps m[key] from a singleton list (the SDK shape for
// SizeBetween(1,1) / SizeAtMost(1) blocks) into a single object. An empty list
// or nil value drops the key entirely; a multi-element list — which the prior
// schema disallowed — is treated as corrupt state and surfaces a diagnostic.
func collapseSingletonList(m map[string]any, key, pathLabel string) fwdiag.Diagnostics {
	v, ok := m[key]
	if !ok {
		return nil
	}
	if v == nil {
		delete(m, key)
		return nil
	}
	list, ok := v.([]any)
	if !ok {
		return nil // already a single object; pass through
	}
	switch len(list) {
	case 0:
		delete(m, key)
		return nil
	case 1:
		obj, ok := list[0].(map[string]any)
		if !ok {
			return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(
				"State upgrade error",
				fmt.Sprintf("unexpected element type at path %q: want object, got %T", pathLabel, list[0]),
			)}
		}
		m[key] = obj
		return nil
	default:
		return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(
			"State upgrade error",
			fmt.Sprintf("unexpected multi-element array at path %q (expected at most one element)", pathLabel),
		)}
	}
}
