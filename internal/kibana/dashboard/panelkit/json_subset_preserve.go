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

package panelkit

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// PreservePriorPanelConfigJSON returns prior when it is a value-subset of current (Kibana may
// expand defaults on read) or semantically equal with defaults applied; otherwise current.
// Used for opaque panel-level config_json on vis panels and unknown-panel fallback reads.
func PreservePriorPanelConfigJSON[T any](ctx context.Context, prior, current customtypes.JSONWithDefaultsValue[T], diags *diag.Diagnostics) customtypes.JSONWithDefaultsValue[T] {
	if prior.IsNull() || prior.IsUnknown() || current.IsNull() || current.IsUnknown() {
		return current
	}

	embedded, err := PriorJSONEmbeddedInExpandedCurrent(prior.ValueString(), current.ValueString())
	if err == nil && embedded {
		return prior
	}

	return PreservePriorJSONWithDefaultsIfEquivalent(ctx, prior, current, diags)
}

// PriorJSONEmbeddedInExpandedCurrent reports whether every value path set in priorJSON is
// present with the same value in currentJSON. Kibana often returns a superset (extra defaults,
// reordered keys, expanded styling). prior must decode to an object with a non-empty string
// top-level "type" (inline chart discriminator).
func PriorJSONEmbeddedInExpandedCurrent(priorJSON, currentJSON string) (bool, error) {
	var priorObj map[string]any
	if err := json.Unmarshal([]byte(priorJSON), &priorObj); err != nil {
		return false, err
	}
	if !hasChartTypeAtRoot(priorObj) {
		return false, nil
	}
	var currentObj map[string]any
	if err := json.Unmarshal([]byte(currentJSON), &currentObj); err != nil {
		return false, err
	}
	return jsonValueSubsumedByCurrentObject(priorObj, currentObj, true), nil
}

func hasChartTypeAtRoot(m map[string]any) bool {
	if m == nil {
		return false
	}
	v, ok := m["type"]
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s != ""
}

func isEmptyJSONSlice(prior any) bool {
	if prior == nil {
		return true
	}
	if pArr, ok := prior.([]any); ok && len(pArr) == 0 {
		return true
	}
	return false
}

func isEmptyJSONMap(prior any) bool {
	if prior == nil {
		return true
	}
	if pMap, ok := prior.(map[string]any); ok && len(pMap) == 0 {
		return true
	}
	return false
}

func isOmissibleDefaultKqlQuery(m map[string]any) bool {
	if len(m) == 0 {
		return true
	}
	lang, hasLang := m["language"]
	expr, hasExpr := m["expression"]
	switch {
	case hasLang && lang == "kql" && !hasExpr && len(m) == 1:
		return true
	case hasLang && lang == "kql" && hasExpr && expr == "" && len(m) == 2:
		return true
	default:
		return false
	}
}

func jsonValueSubsumedByCurrentObject(prior, current map[string]any, isRoot bool) bool {
	for k, pv := range prior {
		if isRoot && k == "styling" {
			continue
		}
		cv, ok := current[k]
		if !ok {
			if isEmptyJSONSlice(pv) || isEmptyJSONMap(pv) {
				continue
			}
			if s, y := pv.(string); y && s == "" {
				continue
			}
			if k == "query" {
				if qm, y := pv.(map[string]any); y && isOmissibleDefaultKqlQuery(qm) {
					continue
				}
			}
			return false
		}
		if isEmptyJSONSlice(pv) {
			if isEmptyJSONSlice(cv) {
				continue
			}
			return false
		}
		if !jsonValueSubsumedByCurrentAny(pv, cv) {
			return false
		}
	}
	return true
}

func jsonValueSubsumedByCurrentAny(prior, current any) bool {
	switch p := prior.(type) {
	case nil:
		return current == nil
	case bool:
		c, ok := current.(bool)
		return ok && c == p
	case float64:
		c, ok := current.(float64)
		return ok && c == p
	case string:
		c, ok := current.(string)
		return ok && c == p
	case []any:
		if isEmptyJSONSlice(prior) && (current == nil) {
			return true
		}
		c, ok := current.([]any)
		if !ok {
			return false
		}
		if len(p) == 0 {
			return len(c) == 0
		}
		if len(p) > len(c) {
			return false
		}
		for i := range p {
			if !jsonValueSubsumedByCurrentAny(p[i], c[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		c, ok := current.(map[string]any)
		if !ok {
			return false
		}
		return jsonValueSubsumedByCurrentObject(p, c, false)
	default:
		return false
	}
}
