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

package lenscommon

import (
	"encoding/json"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

// NormalizeXYPlanComparisonJSON normalizes nested JSON for semantic equality checks used when
// aligning Terraform plan vs state (Lens XY charts, gauge ES|QL blocks, tagcloud ES|QL dimensions, etc.).
func NormalizeXYPlanComparisonJSON(value any) any {
	switch t := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for key, value := range t {
			out[key] = NormalizeXYPlanComparisonJSON(value)
		}
		if formatValue, ok := out["format"]; ok {
			if formatMap, ok := formatValue.(map[string]any); ok {
				if formatBytes, err := json.Marshal(formatMap); err == nil {
					normalizedFormat := NormalizeKibanaLensNumberFormatJSONString(string(formatBytes))
					var formatAny any
					if json.Unmarshal([]byte(normalizedFormat), &formatAny) == nil {
						out["format"] = NormalizeXYPlanComparisonJSON(formatAny)
					}
				}
			}
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, elem := range t {
			out[i] = NormalizeXYPlanComparisonJSON(elem)
		}
		return out
	default:
		return value
	}
}

// PreserveNormalizedJSONSemanticEquality replaces state with plan when normalized structures match semantically.
func PreserveNormalizedJSONSemanticEquality(plan jsontypes.Normalized, state *jsontypes.Normalized) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	var planObj map[string]any
	if err := json.Unmarshal([]byte(plan.ValueString()), &planObj); err != nil {
		return
	}
	var stateObj map[string]any
	if err := json.Unmarshal([]byte(state.ValueString()), &stateObj); err != nil {
		return
	}

	if reflect.DeepEqual(NormalizeXYPlanComparisonJSON(planObj), NormalizeXYPlanComparisonJSON(stateObj)) {
		*state = plan
	}
}

// PreservePlanJSONIfStateAddsOptionalKeys treats optional keys Kibana adds on read as semantically absent when the practitioner omitted them.
func PreservePlanJSONIfStateAddsOptionalKeys(plan jsontypes.Normalized, state *jsontypes.Normalized, optionalKeys ...string) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	var planObj map[string]any
	if err := json.Unmarshal([]byte(plan.ValueString()), &planObj); err != nil {
		return
	}
	var stateObj map[string]any
	if err := json.Unmarshal([]byte(state.ValueString()), &stateObj); err != nil {
		return
	}

	for _, key := range optionalKeys {
		if _, hasPlan := planObj[key]; hasPlan {
			continue
		}
		delete(stateObj, key)
	}

	stateNormalized := NormalizeXYPlanComparisonJSON(stateObj)
	planNormalized := NormalizeXYPlanComparisonJSON(planObj)
	if reflect.DeepEqual(stateNormalized, planNormalized) {
		*state = plan
	}
}

// PreservePlanJSONIfStateOmitsOptionalKeys mirrors PreservePlanJSONIfStateAddsOptionalKeys when state drops optional keys.
func PreservePlanJSONIfStateOmitsOptionalKeys(plan jsontypes.Normalized, state *jsontypes.Normalized, optionalKeys ...string) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	var planObj map[string]any
	if err := json.Unmarshal([]byte(plan.ValueString()), &planObj); err != nil {
		return
	}
	var stateObj map[string]any
	if err := json.Unmarshal([]byte(state.ValueString()), &stateObj); err != nil {
		return
	}

	for _, key := range optionalKeys {
		if _, hasState := stateObj[key]; hasState {
			continue
		}
		delete(planObj, key)
	}

	stateNormalized := NormalizeXYPlanComparisonJSON(stateObj)
	planNormalized := NormalizeXYPlanComparisonJSON(planObj)
	if reflect.DeepEqual(stateNormalized, planNormalized) {
		*state = plan
	}
}
