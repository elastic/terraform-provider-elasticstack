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

package ccr

import (
	"maps"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
)

// NormalizeFlatSettingsKeys converts flat dotted keys (e.g. index.refresh_interval)
// into nested maps suitable for unmarshalling into types.IndexSettings.
// The bool return indicates whether any normalization was performed.
func NormalizeFlatSettingsKeys(m map[string]any) (map[string]any, bool) {
	hasDotted := false
	for k := range m {
		if strings.Contains(k, ".") {
			hasDotted = true
			break
		}
	}
	if !hasDotted {
		return m, false
	}

	flat := make(map[string]any)
	root := make(map[string]any)
	for k, v := range m {
		if strings.Contains(k, ".") {
			flat[k] = v
		} else {
			root[k] = v
		}
	}

	unflattened := customtypes.UnflattenDottedMap(flat)
	return MergeSettingsMaps(root, unflattened), true
}

// MergeSettingsMaps deep-merges overlay into base, recursing into nested
// map[string]any values. Leaf values in overlay always win.
func MergeSettingsMaps(base, overlay map[string]any) map[string]any {
	if len(base) == 0 {
		return overlay
	}
	if len(overlay) == 0 {
		return base
	}
	out := make(map[string]any, len(base)+len(overlay))
	maps.Copy(out, base)
	for k, v := range overlay {
		existing, ok := out[k]
		if !ok {
			out[k] = v
			continue
		}
		baseMap, baseOK := existing.(map[string]any)
		overlayMap, overlayOK := v.(map[string]any)
		if baseOK && overlayOK {
			out[k] = MergeSettingsMaps(baseMap, overlayMap)
			continue
		}
		out[k] = v
	}
	return out
}
