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

	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ilmPhaseBlockKeys are top-level ILM phase blocks (schema version 0 stored each as a singleton list).
var ilmPhaseBlockKeys = [...]string{ilmPhaseHot, ilmPhaseWarm, ilmPhaseCold, ilmPhaseFrozen, ilmPhaseDelete}

// migrateILMStateV0ToV1 unwraps list-wrapped nested blocks from schema version 0 into single objects.
func migrateILMStateV0ToV1(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	stateMap := stateutil.UnmarshalStateMap(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, pk := range ilmPhaseBlockKeys {
		resp.Diagnostics.Append(stateutil.CollapseListPath(stateMap, pk, pk)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if phaseObj, ok := stateMap[pk].(map[string]any); ok {
			unwrapPhaseActionLists(phaseObj)
			if allocateObj, ok := phaseObj[ilmActionAllocate].(map[string]any); ok {
				stateutil.NullifyEmptyString(allocateObj, attrInclude, attrExclude, attrRequire)
			}
		}
	}

	// The SDK provider stored unset JSON string attributes as "" rather than
	// null. The Plugin Framework jsontypes.NormalizedType rejects empty strings,
	// so normalise them to nil before marshalling the upgraded state.
	stateutil.NullifyEmptyString(stateMap, attrMetadata)

	stateutil.MarshalStateMap(stateMap, resp)
}

func unwrapPhaseActionLists(m map[string]any) {
	for k, v := range m {
		if k == attrMinAge {
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
