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

package lensmosaic

import "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"

func populateMosaicLensAttributes(attrs map[string]any) map[string]any {
	lenscommon.PopulatePartitionLensAttributes(attrs)

	if groupBreakdownBy, ok := attrs["group_breakdown_by"].([]any); ok {
		groupBreakdownMaps := make([]map[string]any, 0, len(groupBreakdownBy))
		for _, g := range groupBreakdownBy {
			if m, ok := g.(map[string]any); ok {
				groupBreakdownMaps = append(groupBreakdownMaps, m)
			}
		}
		populated := lenscommon.PopulatePartitionGroupByDefaults(groupBreakdownMaps)
		for i := range groupBreakdownBy {
			if i < len(populated) {
				groupBreakdownBy[i] = populated[i]
			}
		}
	}
	return attrs
}
