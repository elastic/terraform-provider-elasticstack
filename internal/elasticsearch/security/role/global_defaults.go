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

package role

import "maps"

// populateGlobalPrivilegesDefaults removes server-injected empty `global`
// defaults that Elasticsearch adds but that are not meaningful for Terraform
// state, so state matches user intent rather than the raw API blob.
//
// Two classes of server-injected empty defaults are stripped:
//   - the `role` category when it is an empty object (`{}`)
//   - any category whose value is an empty array (`[]`), such as the
//     `data_source: []` entry Elasticsearch 9.5 injects. Empty arrays are
//     never a meaningful user configuration for a `global` category, so this
//     rule is generalized to future empty-array categories as well.
//
// Empty objects other than `role` (for example a user-configured
// `application: {}`) are intentionally preserved.
func populateGlobalPrivilegesDefaults(model map[string]any) map[string]any {
	if model == nil {
		return nil
	}
	out := maps.Clone(model)

	for key, val := range out {
		switch v := val.(type) {
		case map[string]any:
			if key == "role" && len(v) == 0 {
				delete(out, key)
			}
		case []any:
			if len(v) == 0 {
				delete(out, key)
			}
		}
	}

	return out
}
