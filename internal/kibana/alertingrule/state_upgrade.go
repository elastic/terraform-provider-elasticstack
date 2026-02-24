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

package alertingrule

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// Version 0 is the SDKv2 state format - migrate to PFW format
		0: {
			StateUpgrader: migrateV0ToV1,
		},
	}
}

// migrateV0ToV1 handles migration from SDKv2 state format to Plugin Framework state format.
// The main differences are:
// - JSON fields may need normalization for jsontypes.Normalized
// - notify_when may be empty string instead of null
// - throttle may be empty string instead of null
// - frequency, alerts_filter, and timeframe change from lists to single objects
func migrateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	// Default to returning the original state if no changes are needed
	resp.DynamicValue = &tfprotov6.DynamicValue{JSON: req.RawState.JSON}

	var stateMap map[string]any
	if err := json.Unmarshal(req.RawState.JSON, &stateMap); err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal raw state", err.Error())
		return
	}

	modified := false

	// Normalize ID to composite format (space_id/rule_id)
	// SDKv2 may have stored just the rule_id in the id field
	if id, ok := stateMap["id"].(string); ok {
		// Check if ID is already in composite format
		if !strings.Contains(id, "/") {
			// Not composite - construct from space_id and rule_id
			spaceID := "default"
			if sid, ok := stateMap["space_id"].(string); ok && sid != "" {
				spaceID = sid
			}
			ruleID := id
			if rid, ok := stateMap["rule_id"].(string); ok && rid != "" {
				ruleID = rid
			}
			stateMap["id"] = fmt.Sprintf("%s/%s", spaceID, ruleID)
			modified = true
		}
	}

	// Handle notify_when: convert empty string to null for proper handling
	if notifyWhen, ok := stateMap["notify_when"]; ok {
		if notifyWhenStr, ok := notifyWhen.(string); ok && notifyWhenStr == "" {
			stateMap["notify_when"] = nil
			modified = true
		}
	}

	// Handle throttle: convert empty string to null
	if throttle, ok := stateMap["throttle"]; ok {
		if throttleStr, ok := throttle.(string); ok && throttleStr == "" {
			stateMap["throttle"] = nil
			modified = true
		}
	}

	// Handle actions: convert frequency, alerts_filter, and timeframe from lists to objects
	if actions, ok := stateMap["actions"].([]any); ok {
		for _, actionAny := range actions {
			if action, ok := actionAny.(map[string]any); ok {
				// Convert frequency from list to object
				if freq, ok := action["frequency"].([]any); ok {
					if len(freq) > 0 {
						action["frequency"] = freq[0]
					} else {
						action["frequency"] = nil
					}
					modified = true
				}

				// Convert alerts_filter from list to object
				if filter, ok := action["alerts_filter"].([]any); ok {
					if len(filter) > 0 {
						filterObj := filter[0]
						action["alerts_filter"] = filterObj

						// Also convert timeframe within alerts_filter
						if filterMap, ok := filterObj.(map[string]any); ok {
							if tf, ok := filterMap["timeframe"].([]any); ok {
								if len(tf) > 0 {
									filterMap["timeframe"] = tf[0]
								} else {
									filterMap["timeframe"] = nil
								}
							}
						}
					} else {
						action["alerts_filter"] = nil
					}
					modified = true
				}
			}
		}
	}

	// Only re-marshal if we made changes
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
