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

package entitycore

import "github.com/hashicorp/terraform-plugin-framework/types"

// PreserveStringFromPriorIfUnknown returns prior when out is Unknown in the plan
// and prior holds a known value. Used when SkipReadAfterWrite persists the write
// callback model without a read refresh (server-computed fields stay Unknown in
// the plan but must not be written as Unknown to state).
func PreserveStringFromPriorIfUnknown(out, prior types.String) types.String {
	if out.IsUnknown() && !prior.IsUnknown() {
		return prior
	}
	return out
}
