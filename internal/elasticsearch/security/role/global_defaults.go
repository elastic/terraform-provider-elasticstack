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

// populateGlobalPrivilegesDefaults removes API default fields that Elasticsearch
// adds to global privileges but that are not meaningful for Terraform state.
func populateGlobalPrivilegesDefaults(model map[string]any) map[string]any {
	if model == nil {
		return nil
	}
	out := maps.Clone(model)
	if roleVal, ok := out["role"]; ok {
		roleMap, ok := roleVal.(map[string]any)
		if ok && len(roleMap) == 0 {
			delete(out, "role")
		}
	}
	return out
}
