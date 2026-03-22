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
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func priorHasDeclaredToggle(ctx context.Context, prior types.List, toggle string) bool {
	if prior.IsNull() || prior.IsUnknown() {
		return false
	}
	var elems []types.Object
	if diags := prior.ElementsAs(ctx, &elems, false); diags.HasError() {
		return false
	}
	if len(elems) != 1 {
		return false
	}
	attrVals := elems[0].Attributes()
	v, ok := attrVals[toggle]
	if !ok || v.IsNull() || v.IsUnknown() {
		return false
	}
	listV, ok := v.(types.List)
	if !ok {
		return false
	}
	return len(listV.Elements()) > 0
}

func flattenPhase(ctx context.Context, phaseName string, p models.Phase, prior types.List) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	phase := make(map[string]any)
	// Single map shared across readonly/freeze/unfollow, matching SDK flatten behavior.
	enabled := make(map[string]any)

	for _, aCase := range []string{"readonly", "freeze", "unfollow"} {
		if priorHasDeclaredToggle(ctx, prior, aCase) {
			enabled["enabled"] = false
			phase[aCase] = []any{enabled}
		}
	}

	if p.MinAge != "" {
		phase["min_age"] = p.MinAge
	}
	for actionName, action := range p.Actions {
		switch actionName {
		case "readonly", "freeze", "unfollow":
			enabled["enabled"] = true
			phase[actionName] = []any{enabled}
		case "allocate":
			allocateAction := make(map[string]any)
			if v, ok := action["number_of_replicas"]; ok {
				allocateAction["number_of_replicas"] = v
			}
			if v, ok := action["total_shards_per_node"]; ok {
				allocateAction["total_shards_per_node"] = v
			} else {
				allocateAction["total_shards_per_node"] = int64(-1)
			}
			for _, f := range []string{"include", "require", "exclude"} {
				if v, ok := action[f]; ok {
					res, err := json.Marshal(v)
					if err != nil {
						diags.AddError("Failed to marshal allocate filter", err.Error())
						return types.ListUnknown(phaseObjectType(phaseName)), diags
					}
					s := string(res)
					// Omit empty objects so unset optional JSON attrs stay null (matches config).
					if s != "{}" {
						allocateAction[f] = s
					}
				}
			}
			phase[actionName] = []any{allocateAction}
		case "shrink":
			shrinkAction := make(map[string]any, len(action)+1)
			maps.Copy(shrinkAction, map[string]any(action))
			if _, ok := shrinkAction["allow_write_after_shrink"]; !ok {
				shrinkAction["allow_write_after_shrink"] = false
			}
			phase[actionName] = []any{shrinkAction}
		default:
			// models.Action is a named map type; type assertions expect map[string]any.
			phase[actionName] = []any{map[string]any(action)}
		}
	}

	return phaseMapToListValue(ctx, phaseName, phase)
}
