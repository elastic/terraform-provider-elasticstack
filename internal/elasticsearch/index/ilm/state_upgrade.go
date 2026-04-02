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

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ilmPhaseBlockKeys are top-level ILM phase blocks (schema version 0 stored each as a singleton list).
var ilmPhaseBlockKeys = [...]string{ilmPhaseHot, ilmPhaseWarm, ilmPhaseCold, ilmPhaseFrozen, ilmPhaseDelete}

// migrateILMStateV0ToV1 unwraps list-wrapped nested blocks from schema version 0 into single objects.
func migrateILMStateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || req.RawState.JSON == nil {
		resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
		return
	}

	var stateMap map[string]any
	err := json.Unmarshal(req.RawState.JSON, &stateMap)
	if err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not unmarshal prior state: "+err.Error())
		return
	}

	for _, pk := range ilmPhaseBlockKeys {
		if raw, ok := stateMap[pk]; ok {
			u := unwrapSingletonListToMap(raw)
			if u == nil {
				delete(stateMap, pk)
			} else {
				stateMap[pk] = u
			}
		}
	}

	stateJSON, err := json.Marshal(stateMap)
	if err != nil {
		resp.Diagnostics.AddError("State upgrade error", "Could not marshal new state: "+err.Error())
		return
	}
	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: stateJSON,
	}
}

func unwrapSingletonListToMap(v any) any {
	list, ok := v.([]any)
	if !ok {
		return v
	}
	if len(list) == 0 {
		return nil
	}
	first := list[0]
	phaseObj, ok := first.(map[string]any)
	if !ok {
		return v
	}
	unwrapPhaseActionLists(phaseObj)
	return phaseObj
}

func unwrapPhaseActionLists(m map[string]any) {
	for k, v := range m {
		if k == "min_age" {
			continue
		}
		list, ok := v.([]any)
		if !ok {
			continue
		}
		if len(list) == 0 {
			delete(m, k)
			continue
		}
		if inner, ok := list[0].(map[string]any); ok {
			m[k] = inner
		}
	}
}
