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

// PopulateAnySliceDefaults applies fn to each map[string]any element of items, writing each
// result back in place. Non-map elements are left untouched. Use this when the []any slice has
// already been extracted from its parent map (e.g. a nested field within a layer).
func PopulateAnySliceDefaults(items []any, fn func(map[string]any) map[string]any) {
	for i, item := range items {
		if m, ok := item.(map[string]any); ok {
			items[i] = fn(m)
		}
	}
}

// PopulateMapSliceDefaults applies fn to each map[string]any element of the []any-typed field
// attrs[key], writing results back in place. It is a no-op if attrs[key] is absent or not a
// []any. This centralizes the type-assert/filter/apply/write-back loop shared by every Lens
// panel's attribute defaulting.
func PopulateMapSliceDefaults(attrs map[string]any, key string, fn func(map[string]any) map[string]any) {
	if items, ok := attrs[key].([]any); ok {
		PopulateAnySliceDefaults(items, fn)
	}
}

// PopulateMapSliceDefaultsBatch applies a batch defaulting function (one that operates on the
// full []map[string]any at once, such as PopulatePartitionGroupByDefaults) to the []any-typed
// field attrs[key]. It filters attrs[key] to its map[string]any elements, runs fn over them, and
// writes the populated results back in place by index. It is a no-op if attrs[key] is absent or
// not a []any.
func PopulateMapSliceDefaultsBatch(attrs map[string]any, key string, fn func([]map[string]any) []map[string]any) {
	items, ok := attrs[key].([]any)
	if !ok {
		return
	}

	maps := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			maps = append(maps, m)
		}
	}

	populated := fn(maps)
	for i := range items {
		if i < len(populated) {
			items[i] = populated[i]
		}
	}
}
