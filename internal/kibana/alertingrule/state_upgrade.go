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
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	stateutil.SetDefaultState(req, resp)

	stateMap := stateutil.UnmarshalStateMap(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Normalize ID to composite format (space_id/rule_id)
	// SDKv2 may have stored just the rule_id in the id field
	if id, ok := stateMap["id"].(string); ok {
		// Check if ID is already in composite format
		if !strings.Contains(id, "/") {
			// Not composite - construct from space_id and rule_id
			spaceID := clients.DefaultSpaceID
			if sid, ok := stateMap["space_id"].(string); ok && sid != "" {
				spaceID = sid
			}
			ruleID := id
			if rid, ok := stateMap["rule_id"].(string); ok && rid != "" {
				ruleID = rid
			}
			stateMap["id"] = fmt.Sprintf("%s/%s", spaceID, ruleID)
		}
	}

	stateutil.NullifyEmptyString(stateMap, "notify_when", "throttle")
	stateutil.NullifyEmptyString(stateMap, "params")

	// Handle actions: convert frequency, alerts_filter, and timeframe from lists to objects
	if actions, ok := stateMap["actions"].([]any); ok {
		for _, actionAny := range actions {
			if action, ok := actionAny.(map[string]any); ok {
				stateutil.NullifyEmptyString(action, "params")

				resp.Diagnostics.Append(stateutil.CollapseListPath(action, "frequency", "actions.frequency")...)
				if resp.Diagnostics.HasError() {
					return
				}

				resp.Diagnostics.Append(stateutil.CollapseListPath(action, "alerts_filter", "actions.alerts_filter")...)
				if resp.Diagnostics.HasError() {
					return
				}

				if filterMap, ok := action["alerts_filter"].(map[string]any); ok {
					resp.Diagnostics.Append(stateutil.CollapseListPath(filterMap, "timeframe", "actions.alerts_filter.timeframe")...)
					if resp.Diagnostics.HasError() {
						return
					}
				}
			}
		}
	}

	stateutil.MarshalStateMap(stateMap, resp)
}
