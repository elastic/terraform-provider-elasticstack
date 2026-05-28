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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PreserveNullStringIfStateEquals copies a null plan value back into state when the API
// read-back returned the supplied default. Use this for optional typed string attributes
// (e.g. `tagcloud.orientation`, `pie.label_position`) that Kibana auto-populates with a
// hard-coded default when the practitioner omitted the field. Without this, the
// inconsistent plan/state values would surface as "Provider produced inconsistent result
// after apply" diagnostics.
func PreserveNullStringIfStateEquals(plan types.String, state *types.String, expected string) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueString() == expected {
		*state = plan
	}
}

// PreserveNullBoolIfStateEquals mirrors PreserveNullStringIfStateEquals for bool attributes.
// See PreserveNullStringIfStateEquals.
func PreserveNullBoolIfStateEquals(plan types.Bool, state *types.Bool, expected bool) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueBool() == expected {
		*state = plan
	}
}

// PreserveNullInt64IfStateEquals mirrors PreserveNullStringIfStateEquals for int64 attributes.
// See PreserveNullStringIfStateEquals.
func PreserveNullInt64IfStateEquals(plan types.Int64, state *types.Int64, expected int64) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueInt64() == expected {
		*state = plan
	}
}

// PreserveNullFloat64IfStateEquals mirrors PreserveNullStringIfStateEquals for float64 attributes.
// See PreserveNullStringIfStateEquals.
func PreserveNullFloat64IfStateEquals(plan types.Float64, state *types.Float64, expected float64) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueFloat64() == expected {
		*state = plan
	}
}
