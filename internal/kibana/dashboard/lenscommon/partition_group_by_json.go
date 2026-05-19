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

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// StripTopLevelNullMapKeys removes keys whose value is nil so JSON state matches compact user configs.
func StripTopLevelNullMapKeys(m map[string]any) {
	if m == nil {
		return
	}
	for k, v := range m {
		if v == nil {
			delete(m, k)
		}
	}
}

// NewPartitionGroupByJSONFromAPI builds group_by / group_breakdown_by JSON for Terraform state from the API payload.
func NewPartitionGroupByJSONFromAPI(apiPayload any) (customtypes.JSONWithDefaultsValue[[]map[string]any], diag.Diagnostics) {
	var diags diag.Diagnostics
	raw, err := json.Marshal(apiPayload)
	if err != nil {
		diags.AddError("Failed to marshal group_by from API", err.Error())
		return customtypes.JSONWithDefaultsValue[[]map[string]any]{}, diags
	}
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		diags.AddError("Failed to unmarshal group_by from API", err.Error())
		return customtypes.JSONWithDefaultsValue[[]map[string]any]{}, diags
	}
	for i := range items {
		StripTopLevelNullMapKeys(items[i])
	}
	out, err := json.Marshal(items)
	if err != nil {
		diags.AddError("Failed to marshal normalized group_by", err.Error())
		return customtypes.JSONWithDefaultsValue[[]map[string]any]{}, diags
	}
	return customtypes.NewJSONWithDefaultsValue(string(out), PopulatePartitionGroupByDefaults), diags
}
