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

package watch

import (
	"maps"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
)

const (
	watcherDefaultSearchType   = "query_then_fetch"
	watcherDefaultScriptLang   = "painless"
	watcherDefaultLoggingLevel = "info"
)

// populateWatcherJSONDefaults copy-on-write walks a Watcher JSON object and
// fills Elasticsearch-injected defaults used for semantic equality:
//
//   - On every object with a "search" key whose value is an object containing a
//     "request" object, absent rest_total_hits_as_int, search_type, and indices
//     are filled on that request only (never on http.request).
//   - On every object with a "script" key whose value is an object, absent lang
//     is filled with "painless".
//   - On every object with a "logging" key whose value is an object, absent
//     level is filled with "info".
//
// The input tree is never mutated. Unchanged subtrees are returned as-is.
func populateWatcherJSONDefaults(model map[string]any) map[string]any {
	if model == nil {
		return nil
	}
	out, _ := walkWatcherJSONMap(model)
	return out
}

func walkWatcherJSONValue(v any) (any, bool) {
	switch t := v.(type) {
	case map[string]any:
		return walkWatcherJSONMap(t)
	case []any:
		return walkWatcherJSONSlice(t)
	default:
		return v, false
	}
}

func walkWatcherJSONMap(m map[string]any) (map[string]any, bool) {
	if m == nil {
		return m, false
	}

	var result map[string]any
	changed := false

	current := func() map[string]any {
		if result != nil {
			return result
		}
		return m
	}
	ensure := func() map[string]any {
		if result == nil {
			result = maps.Clone(m)
		}
		return result
	}

	for k, v := range m {
		newV, childChanged := walkWatcherJSONValue(v)
		if childChanged {
			ensure()[k] = newV
			changed = true
		}
	}

	if search, ok := current()["search"].(map[string]any); ok {
		if request, ok := search["request"].(map[string]any); ok {
			newReq, reqChanged := fillSearchRequestDefaults(request)
			if reqChanged {
				newSearch := maps.Clone(search)
				newSearch["request"] = newReq
				ensure()["search"] = newSearch
				changed = true
			}
		}
	}

	if script, ok := current()["script"].(map[string]any); ok {
		newScript, scriptChanged := fillScriptLangDefault(script)
		if scriptChanged {
			ensure()["script"] = newScript
			changed = true
		}
	}

	if logging, ok := current()["logging"].(map[string]any); ok {
		newLogging, loggingChanged := fillLoggingLevelDefault(logging)
		if loggingChanged {
			ensure()["logging"] = newLogging
			changed = true
		}
	}

	if !changed {
		return m, false
	}
	return result, true
}

func walkWatcherJSONSlice(s []any) ([]any, bool) {
	if s == nil {
		return s, false
	}

	var result []any
	changed := false

	for i, v := range s {
		newV, childChanged := walkWatcherJSONValue(v)
		if childChanged {
			if result == nil {
				result = slices.Clone(s)
			}
			result[i] = newV
			changed = true
		}
	}

	if !changed {
		return s, false
	}
	return result, true
}

func fillSearchRequestDefaults(request map[string]any) (map[string]any, bool) {
	_, hasTotalHits := request["rest_total_hits_as_int"]
	_, hasSearchType := request["search_type"]
	_, hasIndices := request["indices"]
	if hasTotalHits && hasSearchType && hasIndices {
		return request, false
	}

	out := maps.Clone(request)
	if !hasTotalHits {
		out["rest_total_hits_as_int"] = true
	}
	if !hasSearchType {
		out["search_type"] = watcherDefaultSearchType
	}
	if !hasIndices {
		out["indices"] = []any{}
	}
	return out, true
}

func fillScriptLangDefault(script map[string]any) (map[string]any, bool) {
	if _, hasLang := script["lang"]; hasLang {
		return script, false
	}
	out := maps.Clone(script)
	out["lang"] = watcherDefaultScriptLang
	return out, true
}

func fillLoggingLevelDefault(logging map[string]any) (map[string]any, bool) {
	if _, hasLevel := logging["level"]; hasLevel {
		return logging, false
	}
	out := maps.Clone(logging)
	out["level"] = watcherDefaultLoggingLevel
	return out, true
}

func watcherJSONType() customtypes.JSONWithDefaultsType[map[string]any] {
	return customtypes.NewJSONWithDefaultsType(populateWatcherJSONDefaults)
}

func watcherJSONValue(s string) customtypes.JSONWithDefaultsValue[map[string]any] {
	return customtypes.NewJSONWithDefaultsValue(s, populateWatcherJSONDefaults)
}

func watcherJSONNull() customtypes.JSONWithDefaultsValue[map[string]any] {
	return customtypes.NewJSONWithDefaultsNull(populateWatcherJSONDefaults)
}
