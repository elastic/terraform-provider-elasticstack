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

package index

import (
	"encoding/json"
	"fmt"
	"strings"
)

// stringSliceFromAny extracts a []string from a JSON-decoded any value
// (scalar string, []string, []any). Returns nil when v is nil or unrecognised.
func stringSliceFromAny(v any) []string {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		result := make([]string, len(x))
		for i, e := range x {
			result[i] = fmt.Sprint(e)
		}
		return result
	case string:
		trimmed := strings.TrimSpace(x)
		if strings.HasPrefix(trimmed, "[") {
			var arr []string
			if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
				return arr
			}
		}
		if trimmed != "" {
			return []string{trimmed}
		}
		return nil
	}
	return nil
}
