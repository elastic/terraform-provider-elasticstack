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

package stateutil

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// CollapseListPath applies the singleton-list collapse rule for Terraform state upgrades at m[key].
// It handles the SDK-to-Plugin-Framework migration where MaxItems:1 / SizeBetween(1,1) list blocks
// become SingleNestedBlock.
//
// Semantics:
//   - key absent: no-op
//   - value nil or non-list: no-op (pass-through)
//   - empty list: m[key] = nil
//   - singleton list: m[key] = list[0]
//   - multi-element list: returns an error diagnostic (corrupt state)
func CollapseListPath(m map[string]any, key, pathLabel string) diag.Diagnostics {
	v, ok := m[key]
	if !ok {
		return nil
	}
	if v == nil {
		return nil
	}
	list, ok := v.([]any)
	if !ok {
		return nil
	}
	switch len(list) {
	case 0:
		m[key] = nil
	case 1:
		m[key] = list[0]
	default:
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"State upgrade error",
				fmt.Sprintf(`unexpected multi-element array at path %q`, pathLabel),
			),
		}
	}
	return nil
}
