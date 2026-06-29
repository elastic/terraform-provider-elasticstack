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

package panelkit

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ResolvePanelAttrsShape detects whether attrs carries a flat-keyed representation (all flatKeys
// present at the top level) or a nested config object under configKey. Returns shaped=false when
// neither pattern matches.
func ResolvePanelAttrsShape(attrs map[string]attr.Value, configKey string, flatKeys ...string) (flat bool, nested types.Object, shaped bool) {
	if attrs == nil {
		return false, types.Object{}, false
	}
	if len(flatKeys) > 0 {
		if _, first := attrs[flatKeys[0]]; first {
			if len(flatKeys) == 1 {
				return true, types.Object{}, true
			}
			if _, second := attrs[flatKeys[1]]; second {
				return true, types.Object{}, true
			}
			return false, types.Object{}, false
		}
	}
	raw, ok := attrs[configKey]
	if !ok || raw == nil {
		return false, types.Object{}, false
	}
	obj, ok := raw.(types.Object)
	if !ok {
		return false, types.Object{}, false
	}
	return false, obj, true
}
