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

import (
	"encoding/json"
	"reflect"
)

// WalkJSON recursively walks a decoded JSON value tree, applying leaf to every
// non-container node. Container nodes (map[string]any and []any) are always
// traversed. If leaf is nil, non-container values are returned unchanged.
func WalkJSON(v any, leaf func(any) any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, vv := range val {
			out[k] = WalkJSON(vv, leaf)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, vv := range val {
			out[i] = WalkJSON(vv, leaf)
		}
		return out
	default:
		if leaf != nil {
			return leaf(val)
		}
		return val
	}
}

// NormalizeJSONScalar recursively walks a decoded JSON value and converts
// string-encoded JSON booleans and null back to their native Go types.
// "true" → bool(true), "false" → bool(false), "null" → nil.
// All other values are returned unchanged.
func NormalizeJSONScalar(v any) any {
	return WalkJSON(v, func(leaf any) any {
		s, ok := leaf.(string)
		if !ok {
			return leaf
		}
		switch s {
		case "true":
			return true
		case "false":
			return false
		case "null":
			return nil
		}
		return s
	})
}

// JSONBytesEqual reports whether the JSON in two byte slices is semantically equivalent.
func JSONBytesEqual(a, b []byte) (bool, error) {
	var j, j2 any
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j, j2), nil
}
