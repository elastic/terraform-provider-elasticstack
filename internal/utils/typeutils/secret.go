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

import "github.com/hashicorp/terraform-plugin-framework/types"

// PreserveRedactedSecret resolves a secret attribute read back from an API
// that redacts stored secrets by omitting or blanking them in responses. A
// non-empty apiValue is a real value and is used as-is. Otherwise, the API
// is treated as having redacted the value, so the previously configured
// value is kept when known, falling back to null (e.g. on import, where
// there is no prior value to fall back to).
func PreserveRedactedSecret(apiValue *string, fallback types.String) types.String {
	switch {
	case apiValue != nil && *apiValue != "":
		return types.StringPointerValue(apiValue)
	case IsKnown(fallback):
		return fallback
	default:
		return types.StringNull()
	}
}
