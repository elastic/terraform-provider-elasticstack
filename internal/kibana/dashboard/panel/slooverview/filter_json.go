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

package slooverview

import (
	"encoding/json"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func populateFilterJSONFromMarshaled(item any, out *jsontypes.Normalized) diag.Diagnostics {
	var diags diag.Diagnostics
	b, err := json.Marshal(item)
	if err != nil {
		diags.AddError("Failed to marshal filter from API", err.Error())
		return diags
	}
	var root any
	if err := json.Unmarshal(b, &root); err != nil {
		diags.AddError("Failed to unmarshal filter for canonical JSON", err.Error())
		return diags
	}
	canon, err := json.Marshal(sortJSONMapKeysRecursive(root))
	if err != nil {
		diags.AddError("Failed to marshal canonical filter JSON", err.Error())
		return diags
	}
	*out = jsontypes.NewNormalizedValue(string(canon))
	return diags
}

func sortJSONMapKeysRecursive(v any) any {
	switch x := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]any, len(x))
		for _, k := range keys {
			out[k] = sortJSONMapKeysRecursive(x[k])
		}
		return out
	case []any:
		for i := range x {
			x[i] = sortJSONMapKeysRecursive(x[i])
		}
		return x
	default:
		return v
	}
}
