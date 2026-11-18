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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PreserveSloInstanceID computes slo_instance_id for SLO panel types (slo_burn_rate,
// slo_error_budget, slo_alerts, slo_overview) from the API's raw pointer value, applying
// REQ-009 null-preservation for the API's "*" (all-instances) sentinel.
//
// hasPrior reports whether there is any prior plan/state context to consult for this field at
// all. It is false only for a genuine absence: a real terraform import, or (for list elements,
// e.g. slo_alerts' per-SLO entries) a newly-added item with no corresponding prior item. In that
// case there is no null intent to preserve, so the API value is adopted as-is, except that "*" has
// no meaningful non-null Terraform representation and is normalized to null.
//
// When hasPrior is true, prior is the corresponding prior plan/state field:
//   - If prior is not known (the practitioner never configured this field), the result stays null
//     regardless of what the API returns — this rule intentionally does not special-case "*" here,
//     since REQ-009 null-preservation means an unconfigured field must never pick up drift from the
//     API, sentinel or not.
//   - If prior is known, the raw API value is adopted verbatim (including "*" if the practitioner
//     explicitly configured it and the API echoed it back, and null if the API omits the value),
//     matching how every other optional scalar field is refreshed from the API on a known prior.
func PreserveSloInstanceID(api *string, hasPrior bool, prior types.String) types.String {
	if !hasPrior {
		if api != nil && *api != "*" {
			return types.StringValue(*api)
		}
		return types.StringNull()
	}
	if !typeutils.IsKnown(prior) {
		return types.StringNull()
	}
	return types.StringPointerValue(api)
}
