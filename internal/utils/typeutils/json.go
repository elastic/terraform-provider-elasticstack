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
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

// IsEmptyJSONObject reports whether s is a semantically-empty JSON object —
// either whitespace-only, the literal `{}`, or any JSON object that unmarshals
// to a zero-length, non-nil map. It returns false for non-empty objects,
// arrays, scalars, the JSON literal `null`, and invalid JSON.
func IsEmptyJSONObject(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return true
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(trimmed), &m); err != nil {
		return false
	}
	if m == nil {
		return false
	}
	return len(m) == 0
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

// UnmarshalJSONDiag unmarshals a JSON string into T, returning a single-element
// Diagnostics slice on error instead of a raw error value.
func UnmarshalJSONDiag[T any](data string, errSummary string) (T, diag.Diagnostics) {
	var out T
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return out, diag.Diagnostics{
			diag.NewErrorDiagnostic(errSummary, err.Error()),
		}
	}
	return out, nil
}

// MarshalToNormalized marshals v to a jsontypes.Normalized value.
//
//   - If v is nil, a nil pointer/map/slice/channel/function stored inside an
//     interface{}, or marshals to JSON null, it returns jsontypes.NewNormalizedNull().
//   - On a marshal error it appends an attribute error at p to diags and
//     returns jsontypes.NewNormalizedNull().
func MarshalToNormalized(v any, p path.Path, diags *diag.Diagnostics) jsontypes.Normalized {
	if isNil(v) {
		return jsontypes.NewNormalizedNull()
	}
	b, err := json.Marshal(v)
	if err != nil {
		diags.AddAttributeError(p, "marshal failure", err.Error())
		return jsontypes.NewNormalizedNull()
	}
	if string(b) == "null" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(b))
}

// isNil reports whether v is nil or a nil interface containing a nil underlying
// pointer, slice, map, chan or func.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	}
	return false
}
