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

	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
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

// upgradeActions decodes the raw state, applies mutate to each action map, and
// re-encodes only when at least one mutation occurred. mutate returns true when
// the action was modified; it may also add diagnostics on resp to abort the
// upgrade (the encoded state is left as the original input).
func upgradeActions(req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse, mutate func(action map[string]any) bool) {
	// Default to returning the original state if no changes are needed
	if req.RawState != nil && req.RawState.JSON != nil {
		resp.DynamicValue = &tfprotov6.DynamicValue{JSON: req.RawState.JSON}
	}

	stateMap := stateutil.UnmarshalStateMap(req, resp)
	if resp.Diagnostics.HasError() {
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
		if mutate(action) {
			modified = true
		}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !modified {
		return
	}

	stateutil.MarshalStateMap(stateMap, resp)
}

// migrateParamsV0ToV1 converts each action's params from a JSON object
// (map(string) in the old schema) to a JSON-encoded string (jsontypes.Normalized).
func migrateParamsV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeActions(req, resp, func(action map[string]any) bool {
		params, ok := action["params"]
		if !ok || params == nil {
			return false
		}
		if _, alreadyString := params.(string); alreadyString {
			return false
		}
		jsonBytes, err := json.Marshal(params)
		if err != nil {
			resp.Diagnostics.AddError("Failed to marshal action params during state upgrade", err.Error())
			return false
		}
		action["params"] = string(jsonBytes)
		return true
	})
}

// migrateAlertsFilterV1ToV2 removes any stored alerts_filter map values from actions.
// The prior MapAttribute implementation was non-functional and incompatible with the v2 object shape.
func migrateAlertsFilterV1ToV2(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	upgradeActions(req, resp, func(action map[string]any) bool {
		if _, hasFilter := action["alerts_filter"]; !hasFilter {
			return false
		}
		delete(action, "alerts_filter")
		return true
	})
}
