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
	"strings"
)

// StringSliceFromAny extracts a []string from a JSON-decoded any value
// representing a flat Elasticsearch setting, which may arrive as a scalar
// string, a JSON-encoded string array embedded in a string (for example
// index.query.default_field), a []string, or a []any of strings. ok is false
// when v cannot be represented as string elements, distinguishing that case
// from a legal empty slice.
func StringSliceFromAny(v any) ([]string, bool) {
	switch x := v.(type) {
	case []string:
		return x, true
	case []any:
		elems := make([]string, 0, len(x))
		for _, e := range x {
			s, ok := e.(string)
			if !ok {
				return nil, false
			}
			elems = append(elems, s)
		}
		return elems, true
	case string:
		trimmed := strings.TrimSpace(x)
		if strings.HasPrefix(trimmed, "[") {
			var arr []string
			if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
				return arr, true
			}
		}
		if trimmed == "" {
			return nil, false
		}
		return []string{trimmed}, true
	default:
		return nil, false
	}
}

// StringSliceFromRawJSON decodes raw JSON and extracts a []string from it via
// StringSliceFromAny. ok is false when raw is not valid JSON or cannot be
// represented as string elements.
func StringSliceFromRawJSON(raw json.RawMessage) ([]string, bool) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, false
	}
	return StringSliceFromAny(v)
}
