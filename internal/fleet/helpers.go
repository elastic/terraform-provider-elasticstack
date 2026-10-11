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

package fleet

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProxyIDForUpdate returns a pointer suitable for a generated update body's
// proxy_id field. A known non-empty plan value is sent as-is. When the plan
// unsets a previously set proxy_id, an empty string is sent rather than nil:
// the generated `json:"proxy_id,omitempty"` tag drops nil, and Fleet treats
// an omitted field as "leave unchanged", which would leave the prior
// proxy_id in place and produce an "inconsistent result after apply". When
// proxy_id was already unset, nil is returned so the field stays omitted.
func ProxyIDForUpdate(plan, prior types.String) *string {
	if typeutils.IsKnown(plan) && plan.ValueString() != "" {
		return plan.ValueStringPointer()
	}
	if typeutils.IsKnown(prior) && prior.ValueString() != "" {
		empty := ""
		return &empty
	}
	return nil
}
