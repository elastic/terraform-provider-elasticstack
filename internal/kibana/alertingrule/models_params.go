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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// normalizeRuleParamsForState strips API-returned params keys that the user
// never specified. Kibana injects server-side defaults (e.g. aggType, groupBy)
// into the response even when the user's config omitted them. Without this,
// Terraform sees the extra keys as drift and produces "inconsistent result
// after apply" errors. The approach is generic: any key present in the API
// response but absent from the user's prior state params is removed.
func normalizeRuleParamsForState(apiParams map[string]any, previousParams jsontypes.Normalized) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiParams == nil {
		return apiParams, diags
	}

	priorParams, d := parsePriorParams(previousParams)
	diags.Append(d...)
	if diags.HasError() {
		return apiParams, diags
	}
	if priorParams == nil {
		// No prior state to compare against (first create); keep everything.
		return apiParams, diags
	}

	normalized, ok := removeInjectedDefaultsRecursive(apiParams, priorParams).(map[string]any)
	if !ok {
		// Defensive fallback: params are expected to be a JSON object, but if we
		// can't reconcile shapes, keep everything to avoid accidental data loss.
		return apiParams, diags
	}

	return normalized, diags
}

// parsePriorParams returns the decoded params JSON from the previous Terraform
// state, or nil if there is no usable prior state.
func parsePriorParams(previousParams jsontypes.Normalized) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(previousParams) || previousParams.IsNull() {
		return nil, diags
	}

	var prior map[string]any
	diags.Append(previousParams.Unmarshal(&prior)...)
	if diags.HasError() {
		return nil, diags
	}

	return prior, diags
}

// removeInjectedDefaultsRecursive removes API-injected keys at any nesting level.
//
// It walks the API params structure and removes keys that are present in the API
// payload but absent from the corresponding object in the prior state params.
//
// NOTE: This function assumes the prior state params represent what the user
// configured (i.e. prior-state keys are a subset of configured keys). It only
// iterates keys present in the API payload; therefore, keys present in the prior
// state but missing from the API response will not be preserved in the output.
//
//   - For JSON objects (map[string]interface{}), keys not present in the prior
//     object are dropped, and shared keys are recursed into.
//   - For JSON arrays ([]interface{}), elements are matched by index when the prior
//     is also an array; element values are recursed into when there is a matching
//     prior element, otherwise the element is kept as-is (defensive fallback).
//   - For scalars, the API value is returned unchanged.
func removeInjectedDefaultsRecursive(api any, prior any) any {
	switch apiTyped := api.(type) {
	case map[string]any:
		priorMap, ok := prior.(map[string]any)
		if !ok {
			// Can't compare nested keys without a prior object; keep as-is.
			return apiTyped
		}

		out := make(map[string]any, len(apiTyped))
		for k, v := range apiTyped {
			priorV, exists := priorMap[k]
			if !exists {
				// API injected this key but user never had it — drop it.
				continue
			}
			out[k] = removeInjectedDefaultsRecursive(v, priorV)
		}
		return out

	case []any:
		priorSlice, ok := prior.([]any)
		if !ok {
			// Can't compare element structure without a prior array; keep as-is.
			return apiTyped
		}

		out := make([]any, len(apiTyped))
		for i, v := range apiTyped {
			if i < len(priorSlice) {
				out[i] = removeInjectedDefaultsRecursive(v, priorSlice[i])
			} else {
				// Defensive fallback: no matching prior element; keep.
				out[i] = v
			}
		}
		return out

	default:
		return api
	}
}
