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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func priorHasDeclaredToggle(_ context.Context, prior types.Object, toggle string) bool {
	if prior.IsNull() || prior.IsUnknown() {
		return false
	}
	attrVals := prior.Attributes()
	v, ok := attrVals[toggle]
	if !ok || v.IsNull() || v.IsUnknown() {
		return false
	}
	objV, ok := v.(types.Object)
	if !ok {
		return false
	}
	return !objV.IsNull() && !objV.IsUnknown()
}

func flattenPhase(ctx context.Context, phaseName string, minAge string, actions map[string]map[string]any, prior types.Object) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	phase := make(map[string]any)

	for _, aCase := range []string{ilmActionReadonly, ilmActionFreeze, ilmActionUnfollow} {
		if priorHasDeclaredToggle(ctx, prior, aCase) {
			phase[aCase] = []any{map[string]any{attrEnabled: false}}
		}
	}

	if minAge != "" {
		phase[attrMinAge] = minAge
	}
	for actionName, action := range actions {
		switch actionName {
		case ilmActionReadonly, ilmActionFreeze, ilmActionUnfollow:
			phase[actionName] = []any{map[string]any{attrEnabled: true}}
		case ilmActionAllocate:
			allocateAction := make(map[string]any)
			if v, ok := action[attrNumberOfReplicas]; ok {
				allocateAction[attrNumberOfReplicas] = v
			}
			if v, ok := action[attrTotalShardsPerNode]; ok {
				allocateAction[attrTotalShardsPerNode] = v
			}
			for _, f := range []string{attrInclude, attrRequire, attrExclude} {
				if v, ok := action[f]; ok {
					res, err := json.Marshal(v)
					if err != nil {
						diags.AddError("Failed to marshal allocate filter", err.Error())
						return types.ObjectUnknown(phaseObjectType(phaseName).AttrTypes), diags
					}
					s := string(res)
					// Omit empty objects so unset optional JSON attrs stay null (matches config).
					if s != "{}" {
						allocateAction[f] = s
					}
				}
			}
			phase[actionName] = []any{allocateAction}
		case ilmActionShrink:
			shrinkAction := make(map[string]any, len(action))
			maps.Copy(shrinkAction, action)
			if _, ok := shrinkAction[attrAllowWriteAfterShrink]; !ok {
				shrinkAction[attrAllowWriteAfterShrink] = false
			}
			phase[actionName] = []any{shrinkAction}
		default:
			phase[actionName] = []any{action}
		}
	}

	return phaseMapToObjectValue(ctx, phaseName, phase)
}
