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

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func (r *securityDetectionRuleResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// Version 0 stored actions[*].params as map(string).
		// Version 1 stores it as a JSON-normalized string.
		0: {StateUpgrader: migrateParamsV0ToV1},
		// Version 2 replaces the broken alerts_filter map with a structured block.
		1: {StateUpgrader: migrateAlertsFilterV1ToV2},
	}
}

// migrateParamsV0ToV1 converts each action's params from a JSON object
// (map(string) in the old schema) to a JSON-encoded string (jsontypes.Normalized).
func migrateParamsV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	resp.DynamicValue = &tfprotov6.DynamicValue{JSON: req.RawState.JSON}

	var stateMap map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &stateMap); err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal raw state", err.Error())
		return
	}

	actions, ok := stateMap["actions"].([]any)
	if !ok || len(actions) == 0 {
		return
	}

	modified := false
	for _, actionAny := range actions {
		action, ok := actionAny.(map[string]any)
		if !ok {
			continue
		}
		params, ok := action["params"]
		if !ok || params == nil {
			continue
		}
		// In v0 state, params is a JSON object (map[string]any).
		// In v1, params must be a JSON string.
		if _, alreadyString := params.(string); alreadyString {
			continue
		}
		jsonBytes, err := json.Marshal(params)
		if err != nil {
			resp.Diagnostics.AddError("Failed to marshal action params during state upgrade", err.Error())
			return
		}
		action["params"] = string(jsonBytes)
		modified = true
	}

	if !modified {
		return
	}

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal upgraded state", err.Error())
		return
	}
	resp.DynamicValue.JSON = stateJSON
}

// migrateAlertsFilterV1ToV2 removes any stored alerts_filter map values from actions.
// The prior MapAttribute implementation was non-functional and incompatible with the v2 object shape.
func migrateAlertsFilterV1ToV2(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	resp.DynamicValue = &tfprotov6.DynamicValue{JSON: req.RawState.JSON}

	var stateMap map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &stateMap); err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal raw state", err.Error())
		return
	}

	actions, ok := stateMap["actions"].([]any)
	if !ok || len(actions) == 0 {
		return
	}

	modified := false
	for _, actionAny := range actions {
		action, ok := actionAny.(map[string]any)
		if !ok {
			continue
		}
		if _, hasFilter := action["alerts_filter"]; hasFilter {
			delete(action, "alerts_filter")
			modified = true
		}
	}

	if !modified {
		return
	}

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal upgraded state", err.Error())
		return
	}
	resp.DynamicValue.JSON = stateJSON
}
