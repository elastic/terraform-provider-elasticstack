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

// NormalizeJSONScalar recursively walks a decoded JSON value and converts
// string-encoded JSON booleans and null back to their native Go types.
// "true" → bool(true), "false" → bool(false), "null" → nil.
// All other values are returned unchanged.
func NormalizeJSONScalar(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, vv := range val {
			out[k] = NormalizeJSONScalar(vv)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, vv := range val {
			out[i] = NormalizeJSONScalar(vv)
		}
		return out
	case string:
		switch val {
		case "true":
			return true
		case "false":
			return false
		case "null":
			return nil
		}
		return val
	default:
		return v
	}
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
