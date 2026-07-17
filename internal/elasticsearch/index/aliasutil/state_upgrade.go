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

package aliasutil

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
)

// NormalizeTemplateAliasesInV1State collapses SDK-style echoed index_routing/search_routing (same
// non-empty value, routing empty or equal) into the Plugin Framework routing-only shape so migrated
// state matches configuration and avoids spurious refresh plans after upgrade.
func NormalizeTemplateAliasesInV1State(tmpl map[string]any) {
	av, ok := tmpl["alias"]
	if !ok || av == nil {
		return
	}
	list, ok := av.([]any)
	if !ok {
		return
	}
	for i, el := range list {
		am, ok := el.(map[string]any)
		if !ok {
			continue
		}
		// Plugin Framework jsontypes.Normalized rejects ""; SDK state may store the literal empty string.
		if fv, ok := am["filter"]; ok {
			if s, ok := fv.(string); ok && s == "" {
				am["filter"] = nil
			}
		}
		ir := StringishJSONState(am["index_routing"])
		sr := StringishJSONState(am["search_routing"])
		rt := StringishJSONState(am["routing"])
		if ir != "" && ir == sr && (rt == "" || rt == ir) {
			am["routing"] = ir
			am["index_routing"] = ""
			am["search_routing"] = ""
			list[i] = am
		}
	}
	tmpl["alias"] = list
}

// StringishJSONState returns v as a string, handling nil and non-string JSON-decoded values.
func StringishJSONState(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// NormalizeTemplateObjectInV1State normalizes the "template" sub-object inside a V0→V1 state map.
// It ensures the base keys (alias, mappings, settings) plus any extraKeys are present, nullifies
// empty-string mappings/settings, and normalizes alias routing fields. The caller is responsible for
// collapsing the "template" list path (via stateutil.CollapseListPath) before calling this.
func NormalizeTemplateObjectInV1State(stateMap map[string]any, extraKeys ...string) {
	tmpl, ok := stateMap["template"].(map[string]any)
	if !ok {
		return
	}
	baseKeys := []string{"alias", "mappings", "settings"}
	stateutil.EnsureMapKeys(tmpl, append(baseKeys, extraKeys...)...)
	stateutil.NullifyEmptyString(tmpl, "mappings", "settings")
	NormalizeTemplateAliasesInV1State(tmpl)
}

// NormalizeVersionZero drops the "version" key when its JSON-decoded value is 0.
// SDKv2 may persist version = 0 when the field is omitted in HCL; Elasticsearch readback and
// the Plugin Framework schema treat that as unset (null).
func NormalizeVersionZero(stateMap map[string]any) {
	if v, ok := stateMap["version"]; ok && JSONNumberish(v) == 0 {
		delete(stateMap, "version")
	}
}

// JSONNumberish returns v as float64 when JSON-decoded state stores numbers (typically float64).
func JSONNumberish(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		f, _ := x.Float64()
		return f
	default:
		return 0
	}
}
