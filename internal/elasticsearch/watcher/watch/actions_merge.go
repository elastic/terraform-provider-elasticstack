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

// elasticsearchWatcherRedactedSecret is the placeholder Elasticsearch returns
// for redacted secret string leaves in Watcher documents (including actions).
const elasticsearchWatcherRedactedSecret = "::es_redacted::"

// mergeActionsPreservingRedactedLeaves returns a deep copy of apiActions where
// each string leaf equal to elasticsearchWatcherRedactedSecret is replaced by
// the value at the same JSON path in priorActions when that prior value is
// non-nil and is not itself the redacted sentinel. The substituted prior value
// may be of any JSON type (string, object, array, number, bool); this matters
// because Elasticsearch returns the sentinel as a string even when the
// pre-redaction value was, for example, a stored-script reference object such
// as `{"id": "<script-id>"}` or an inline-script object. All non-redacted
// values come from the API document.
func mergeActionsPreservingRedactedLeaves(apiActions map[string]any, priorActions any) map[string]any {
	priorRoot, _ := priorActions.(map[string]any)
	out := make(map[string]any, len(apiActions))
	for k, v := range apiActions {
		var priorChild any
		if priorRoot != nil {
			priorChild = priorRoot[k]
		}
		out[k] = mergePreserveRedactedLeaves(v, priorChild)
	}
	return out
}

// isRedactedOrAbsent reports whether priorVal carries no usable replacement
// for a redacted leaf: nil, or the same redacted sentinel string.
func isRedactedOrAbsent(priorVal any) bool {
	if priorVal == nil {
		return true
	}
	s, ok := priorVal.(string)
	return ok && s == elasticsearchWatcherRedactedSecret
}

func mergePreserveRedactedLeaves(apiVal, priorVal any) any {
	if apiVal == nil {
		return nil
	}
	if s, ok := apiVal.(string); ok {
		if s == elasticsearchWatcherRedactedSecret && !isRedactedOrAbsent(priorVal) {
			return priorVal
		}
		return apiVal
	}
	switch av := apiVal.(type) {
	case map[string]any:
		priorMap, _ := priorVal.(map[string]any)
		out := make(map[string]any, len(av))
		for k, v := range av {
			var pv any
			if priorMap != nil {
				pv = priorMap[k]
			}
			out[k] = mergePreserveRedactedLeaves(v, pv)
		}
		return out
	case []any:
		priorArr, _ := priorVal.([]any)
		out := make([]any, len(av))
		for i, v := range av {
			var pv any
			if i < len(priorArr) {
				pv = priorArr[i]
			}
			out[i] = mergePreserveRedactedLeaves(v, pv)
		}
		return out
	default:
		return apiVal
	}
}
