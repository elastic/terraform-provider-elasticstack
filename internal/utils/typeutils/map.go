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

package typeutils

import "github.com/hashicorp/terraform-plugin-framework/types"

// BoolFromMap extracts a bool from a map[string]any by key, returning types.BoolNull() when the map is nil, the key is absent, or the value is not a bool.
func BoolFromMap(m map[string]any, key string) types.Bool {
	if m == nil {
		return types.BoolNull()
	}
	if v, ok := m[key].(bool); ok {
		return types.BoolValue(v)
	}
	return types.BoolNull()
}

// StringFromMap extracts a string from a map[string]any by key, returning types.StringNull() when the map is nil, the key is absent, or the value is not a string.
func StringFromMap(m map[string]any, key string) types.String {
	if m == nil {
		return types.StringNull()
	}
	if v, ok := m[key].(string); ok {
		return types.StringValue(v)
	}
	return types.StringNull()
}

// Float64FromMap extracts a float64 from a map[string]any by key, returning types.Float64Null()
// when the map is nil, the key is absent, or the value is not a numeric type.
// It coerces int and int64 values to float64.
func Float64FromMap(m map[string]any, key string) types.Float64 {
	if m == nil {
		return types.Float64Null()
	}
	v, ok := m[key]
	if !ok {
		return types.Float64Null()
	}
	switch val := v.(type) {
	case float64:
		return types.Float64Value(val)
	case int:
		return types.Float64Value(float64(val))
	case int64:
		return types.Float64Value(float64(val))
	default:
		return types.Float64Null()
	}
}

// SetBoolInMap sets a bool field in the map if the value is known and non-null.
func SetBoolInMap(m map[string]any, key string, val types.Bool) {
	if !val.IsNull() && !val.IsUnknown() {
		m[key] = val.ValueBool()
	}
}

// SetStringInMap sets a string field in the map if the value is known and non-null.
func SetStringInMap(m map[string]any, key string, val types.String) {
	if !val.IsNull() && !val.IsUnknown() {
		m[key] = val.ValueString()
	}
}

// PointerInterfaceMapFromAnyMap converts a map[string]any to map[string]*any by
// taking pointers to each value. This is needed when constructing API request bodies
// that require pointer values.
func PointerInterfaceMapFromAnyMap(input map[string]any) map[string]*any {
	output := make(map[string]*any, len(input))
	for k, v := range input {
		value := v
		output[k] = &value
	}

	return output
}

// FlipMap returns a new map with keys and values swapped.
func FlipMap[K comparable, V comparable](m map[K]V) map[V]K {
	inv := make(map[V]K)
	for k, v := range m {
		inv[v] = k
	}
	return inv
}

// FlattenMap recursively flattens a nested map into a single-level map with dot-separated keys.
// For example, {"index": {"key": 1}} becomes {"index.key": 1}.
func FlattenMap(m map[string]any) map[string]any {
	out := make(map[string]any)
	var flattener func(string, map[string]any, map[string]any)
	flattener = func(k string, src, dst map[string]any) {
		if len(k) > 0 {
			k += "."
		}
		for key, v := range src {
			switch inner := v.(type) {
			case map[string]any:
				flattener(k+key, inner, dst)
			default:
				dst[k+key] = v
			}
		}
	}
	flattener("", m, out)
	return out
}
